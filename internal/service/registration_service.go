package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	domain "cobra-template/internal/domain/registration"
	"cobra-template/pkg/logger"

	"github.com/google/uuid"
)

// RegistrationService implements the registration business logic
type RegistrationService struct {
	studentRepo      domain.StudentRepository
	sectionRepo      domain.SectionRepository
	registrationRepo domain.RegistrationRepository
	waitlistRepo     domain.WaitlistRepository
	cacheService     CacheService
	queueService     QueueService
}

// NewRegistrationService creates a new registration service
func NewRegistrationService(
	studentRepo domain.StudentRepository,
	sectionRepo domain.SectionRepository,
	registrationRepo domain.RegistrationRepository,
	waitlistRepo domain.WaitlistRepository,
	cacheService CacheService,
	queueService QueueService,
) *RegistrationService {
	return &RegistrationService{
		studentRepo:      studentRepo,
		sectionRepo:      sectionRepo,
		registrationRepo: registrationRepo,
		waitlistRepo:     waitlistRepo,
		cacheService:     cacheService,
		queueService:     queueService,
	}
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	StudentID  uuid.UUID   `json:"student_id" validate:"required"`
	SectionIDs []uuid.UUID `json:"section_ids" validate:"required,min=1"`
}

// RegisterResponse represents the response for registration
type RegisterResponse struct {
	Results []RegistrationResult `json:"results"`
}

// RegistrationResult represents the result of a single section registration
type RegistrationResult struct {
	SectionID uuid.UUID `json:"section_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Position  *int      `json:"waitlist_position,omitempty"`
}

// Register handles course registration with optimistic locking and caching
func (s *RegistrationService) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	logger.Info("Processing registration for student %s with %d sections", req.StudentID, len(req.SectionIDs))

	// Validate student exists
	student, err := s.studentRepo.GetByID(ctx, req.StudentID)
	if err != nil {
		return nil, fmt.Errorf("student not found: %w", err)
	}

	if student.EnrollmentStatus != "active" {
		return nil, errors.New("student is not in active status")
	}

	response := &RegisterResponse{
		Results: make([]RegistrationResult, 0, len(req.SectionIDs)),
	}

	// Process each section registration
	for _, sectionID := range req.SectionIDs {
		result := s.registerForSection(ctx, req.StudentID, sectionID)
		response.Results = append(response.Results, result)
	}

	return response, nil
}

// registerForSection handles registration for a single section
func (s *RegistrationService) registerForSection(ctx context.Context, studentID, sectionID uuid.UUID) RegistrationResult {
	// Check if already registered
	existing, err := s.registrationRepo.GetByStudentAndSection(ctx, studentID, sectionID)
	if err == nil && existing != nil {
		return RegistrationResult{
			SectionID: sectionID,
			Status:    "already_registered",
			Message:   fmt.Sprintf("Already registered with status: %s", existing.Status),
		}
	}

	// Check cache for seat availability
	available, err := s.cacheService.GetAvailableSeats(ctx, sectionID)
	if err == nil && available <= 0 {
		// No seats available, add to waitlist
		position, err := s.addToWaitlist(ctx, studentID, sectionID)
		if err != nil {
			logger.Error("Failed to add to waitlist: %v", err)
			return RegistrationResult{
				SectionID: sectionID,
				Status:    "failed",
				Message:   "Failed to add to waitlist",
			}
		}
		return RegistrationResult{
			SectionID: sectionID,
			Status:    "waitlisted",
			Message:   "Added to waitlist",
			Position:  &position,
		}
	}

	// Enqueue registration request for async processing
	registrationJob := RegistrationJob{
		StudentID: studentID,
		SectionID: sectionID,
		Timestamp: time.Now(),
	}

	if err := s.queueService.EnqueueRegistration(ctx, registrationJob); err != nil {
		logger.Error("Failed to enqueue registration: %v", err)
		return RegistrationResult{
			SectionID: sectionID,
			Status:    "failed",
			Message:   "Failed to process registration",
		}
	}

	return RegistrationResult{
		SectionID: sectionID,
		Status:    "processing",
		Message:   "Registration request submitted for processing",
	}
}

// ProcessRegistrationJob processes a registration job from the queue
func (s *RegistrationService) ProcessRegistrationJob(ctx context.Context, job RegistrationJob) error {
	logger.Info("Processing registration job for student %s and section %s", job.StudentID, job.SectionID)

	// Use optimistic locking for seat allocation
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		success, err := s.attemptRegistration(ctx, job.StudentID, job.SectionID)
		if err != nil {
			logger.Error("Registration attempt %d failed: %v", attempt, err)
			if attempt == maxRetries {
				// Final attempt failed, add to waitlist
				_, waitlistErr := s.addToWaitlist(ctx, job.StudentID, job.SectionID)
				if waitlistErr != nil {
					logger.Error("Failed to add to waitlist after failed registration: %v", waitlistErr)
				}
				return fmt.Errorf("registration failed after %d attempts: %w", maxRetries, err)
			}
			// Brief delay before retry
			time.Sleep(time.Duration(attempt*100) * time.Millisecond)
			continue
		}

		if success {
			logger.Info("Registration successful for student %s in section %s", job.StudentID, job.SectionID)
			// Update cache
			if err := s.cacheService.DecrementAvailableSeats(ctx, job.SectionID); err != nil {
				logger.Warn("Failed to update cache after successful registration: %v", err)
			}
			return nil
		}

		// No seats available, add to waitlist
		position, err := s.addToWaitlist(ctx, job.StudentID, job.SectionID)
		if err != nil {
			logger.Error("Failed to add to waitlist: %v", err)
			return err
		}
		logger.Info("Student %s added to waitlist position %d for section %s", job.StudentID, position, job.SectionID)
		return nil
	}

	return errors.New("unexpected end of registration processing")
}

// attemptRegistration attempts to register using optimistic locking
func (s *RegistrationService) attemptRegistration(ctx context.Context, studentID, sectionID uuid.UUID) (bool, error) {
	// Get current section state
	section, err := s.sectionRepo.GetByID(ctx, sectionID)
	if err != nil {
		return false, fmt.Errorf("failed to get section: %w", err)
	}

	if section.AvailableSeats <= 0 {
		return false, nil // No seats available
	}

	// Attempt to decrement available seats with optimistic locking
	section.AvailableSeats--
	section.Version++

	err = s.sectionRepo.UpdateWithOptimisticLock(ctx, section)
	if err != nil {
		// Optimistic lock failure - retry
		return false, err
	}

	// Create registration record
	registration := &domain.Registration{
		RegistrationID:   uuid.New(),
		StudentID:        studentID,
		SectionID:        sectionID,
		Status:           domain.StatusEnrolled,
		RegistrationDate: time.Now(),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Version:          1,
	}

	if err := s.registrationRepo.Create(ctx, registration); err != nil {
		// Registration failed, need to rollback seat decrement
		section.AvailableSeats++
		if rollbackErr := s.sectionRepo.UpdateWithOptimisticLock(ctx, section); rollbackErr != nil {
			logger.Error("Failed to rollback seat count: %v", rollbackErr)
		}
		return false, fmt.Errorf("failed to create registration: %w", err)
	}

	return true, nil
}

// addToWaitlist adds a student to the waitlist
func (s *RegistrationService) addToWaitlist(ctx context.Context, studentID, sectionID uuid.UUID) (int, error) {
	position, err := s.waitlistRepo.GetNextPosition(ctx, sectionID)
	if err != nil {
		return 0, fmt.Errorf("failed to get waitlist position: %w", err)
	}

	waitlistEntry := &domain.WaitlistEntry{
		WaitlistID: uuid.New(),
		StudentID:  studentID,
		SectionID:  sectionID,
		Position:   position,
		Timestamp:  time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.waitlistRepo.Create(ctx, waitlistEntry); err != nil {
		return 0, fmt.Errorf("failed to create waitlist entry: %w", err)
	}

	return position, nil
}

// DropCourse handles course dropping and waitlist processing
func (s *RegistrationService) DropCourse(ctx context.Context, studentID, sectionID uuid.UUID) error {
	logger.Info("Processing course drop for student %s and section %s", studentID, sectionID)

	// Get current registration
	registration, err := s.registrationRepo.GetByStudentAndSection(ctx, studentID, sectionID)
	if err != nil {
		return fmt.Errorf("registration not found: %w", err)
	}

	if registration.Status != domain.StatusEnrolled {
		return errors.New("can only drop enrolled courses")
	}

	// Update registration status
	registration.Status = domain.StatusDropped
	registration.UpdatedAt = time.Now()

	if err := s.registrationRepo.Update(ctx, registration); err != nil {
		return fmt.Errorf("failed to update registration: %w", err)
	}

	// Increment available seats
	section, err := s.sectionRepo.GetByID(ctx, sectionID)
	if err != nil {
		return fmt.Errorf("failed to get section: %w", err)
	}

	section.AvailableSeats++
	if err := s.sectionRepo.UpdateWithOptimisticLock(ctx, section); err != nil {
		logger.Error("Failed to update seat count after drop: %v", err)
	}

	// Update cache
	if err := s.cacheService.IncrementAvailableSeats(ctx, sectionID); err != nil {
		logger.Warn("Failed to update cache after course drop: %v", err)
	}

	// Process waitlist
	if err := s.processWaitlist(ctx, sectionID); err != nil {
		logger.Error("Failed to process waitlist after course drop: %v", err)
	}

	logger.Info("Course drop completed for student %s and section %s", studentID, sectionID)
	return nil
}

// processWaitlist processes the waitlist when seats become available
func (s *RegistrationService) processWaitlist(ctx context.Context, sectionID uuid.UUID) error {
	// Get next student from waitlist
	nextEntry, err := s.waitlistRepo.GetNextInLine(ctx, sectionID)
	if err != nil || nextEntry == nil {
		return nil // No one in waitlist
	}

	// Remove from waitlist
	if err := s.waitlistRepo.Delete(ctx, nextEntry.WaitlistID); err != nil {
		return fmt.Errorf("failed to remove from waitlist: %w", err)
	}

	// Enqueue registration job
	registrationJob := RegistrationJob{
		StudentID: nextEntry.StudentID,
		SectionID: sectionID,
		Timestamp: time.Now(),
	}

	return s.queueService.EnqueueRegistration(ctx, registrationJob)
}
