package service

import (
	domain "cobra-template/internal/domain/registration"
	interfaces "cobra-template/internal/interfaces/infrastructure"
	"cobra-template/pkg/logger"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type RegistrationService struct {
	studentRepo      interfaces.StudentRepository
	sectionRepo      interfaces.SectionRepository
	registrationRepo interfaces.RegistrationRepository
	waitlistRepo     interfaces.WaitlistRepository
	cacheService     interfaces.CacheService
	queueService     interfaces.QueueService
}

func NewRegistrationService(
	studentRepo interfaces.StudentRepository,
	sectionRepo interfaces.SectionRepository,
	registrationRepo interfaces.RegistrationRepository,
	waitlistRepo interfaces.WaitlistRepository,
	cacheService interfaces.CacheService,
	queueService interfaces.QueueService,
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

type RegisterRequest struct {
	StudentID  uuid.UUID   `json:"student_id" validate:"required"`
	SectionIDs []uuid.UUID `json:"section_ids" validate:"required,min=1"`
}

type RegisterResponse struct {
	Results []RegistrationResult `json:"results"`
}

type RegistrationResult struct {
	SectionID uuid.UUID `json:"section_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Position  *int      `json:"waitlist_position,omitempty"`
}

func (s *RegistrationService) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	logger.Info("Processing registration for student %s with %d sections", req.StudentID, len(req.SectionIDs))

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

	for _, sectionID := range req.SectionIDs {
		result := s.registerForSection(ctx, req.StudentID, sectionID)
		response.Results = append(response.Results, result)
	}

	return response, nil
}

func (s *RegistrationService) registerForSection(ctx context.Context, studentID, sectionID uuid.UUID) RegistrationResult {

	existing, err := s.registrationRepo.GetByStudentAndSection(ctx, studentID, sectionID)
	if err == nil && existing != nil {
		return RegistrationResult{
			SectionID: sectionID,
			Status:    "already_registered",
			Message:   fmt.Sprintf("Already registered with status: %s", existing.Status),
		}
	}

	available, err := s.cacheService.GetAvailableSeats(ctx, sectionID)
	if err == nil && available <= 0 {

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

	registrationJob := interfaces.RegistrationJob{
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

func (s *RegistrationService) ProcessRegistrationJob(ctx context.Context, job interfaces.RegistrationJob) error {
	logger.Info("Processing registration job for student %s and section %s", job.StudentID, job.SectionID)

	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		success, err := s.attemptRegistration(ctx, job.StudentID, job.SectionID)
		if err != nil {
			logger.Error("Registration attempt %d failed: %v", attempt, err)
			if attempt == maxRetries {

				_, waitlistErr := s.addToWaitlist(ctx, job.StudentID, job.SectionID)
				if waitlistErr != nil {
					logger.Error("Failed to add to waitlist after failed registration: %v", waitlistErr)
				}
				return fmt.Errorf("registration failed after %d attempts: %w", maxRetries, err)
			}

			time.Sleep(time.Duration(attempt*100) * time.Millisecond)
			continue
		}

		if success {
			logger.Info("Registration successful for student %s in section %s", job.StudentID, job.SectionID)

			if err := s.cacheService.DecrementAvailableSeats(ctx, job.SectionID); err != nil {
				logger.Warn("Failed to update cache after successful registration: %v", err)
			}
			return nil
		}

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

func (s *RegistrationService) attemptRegistration(ctx context.Context, studentID, sectionID uuid.UUID) (bool, error) {

	section, err := s.sectionRepo.GetByID(ctx, sectionID)
	if err != nil {
		return false, fmt.Errorf("failed to get section: %w", err)
	}

	if section.AvailableSeats <= 0 {
		return false, nil
	}

	section.AvailableSeats--
	section.Version++

	err = s.sectionRepo.UpdateWithOptimisticLock(ctx, section)
	if err != nil {

		return false, err
	}

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

		section.AvailableSeats++
		if rollbackErr := s.sectionRepo.UpdateWithOptimisticLock(ctx, section); rollbackErr != nil {
			logger.Error("Failed to rollback seat count: %v", rollbackErr)
		}
		return false, fmt.Errorf("failed to create registration: %w", err)
	}

	return true, nil
}

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

func (s *RegistrationService) DropCourse(ctx context.Context, studentID, sectionID uuid.UUID) error {
	logger.Info("Processing course drop for student %s and section %s", studentID, sectionID)

	registration, err := s.registrationRepo.GetByStudentAndSection(ctx, studentID, sectionID)
	if err != nil {
		return fmt.Errorf("registration not found: %w", err)
	}

	if registration.Status != domain.StatusEnrolled {
		return errors.New("can only drop enrolled courses")
	}

	registration.Status = domain.StatusDropped
	registration.UpdatedAt = time.Now()

	if err := s.registrationRepo.Update(ctx, registration); err != nil {
		return fmt.Errorf("failed to update registration: %w", err)
	}

	section, err := s.sectionRepo.GetByID(ctx, sectionID)
	if err != nil {
		return fmt.Errorf("failed to get section: %w", err)
	}

	section.AvailableSeats++
	if err := s.sectionRepo.UpdateWithOptimisticLock(ctx, section); err != nil {
		logger.Error("Failed to update seat count after drop: %v", err)
	}

	if err := s.cacheService.IncrementAvailableSeats(ctx, sectionID); err != nil {
		logger.Warn("Failed to update cache after course drop: %v", err)
	}

	if err := s.processWaitlist(ctx, sectionID); err != nil {
		logger.Error("Failed to process waitlist after course drop: %v", err)
	}

	logger.Info("Course drop completed for student %s and section %s", studentID, sectionID)
	return nil
}

func (s *RegistrationService) processWaitlist(ctx context.Context, sectionID uuid.UUID) error {

	nextEntry, err := s.waitlistRepo.GetNextInLine(ctx, sectionID)
	if err != nil || nextEntry == nil {
		return nil
	}

	if err := s.waitlistRepo.Delete(ctx, nextEntry.WaitlistID); err != nil {
		return fmt.Errorf("failed to remove from waitlist: %w", err)
	}

	registrationJob := interfaces.RegistrationJob{
		StudentID: nextEntry.StudentID,
		SectionID: sectionID,
		Timestamp: time.Now(),
	}

	return s.queueService.EnqueueRegistration(ctx, registrationJob)
}
