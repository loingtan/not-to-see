package service

import (
	domain "cobra-template/internal/domain/registration"
	interfaces "cobra-template/internal/interfaces/infrastructure"
	serviceInterfaces "cobra-template/internal/interfaces/service"
	"cobra-template/pkg/logger"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var _ serviceInterfaces.RegistrationService = (*RegistrationService)(nil)

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

type RegisterRequest = serviceInterfaces.RegisterRequest
type RegisterResponse = serviceInterfaces.RegisterResponse
type RegistrationResult = serviceInterfaces.RegistrationResult

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
	newSeatCount, err := s.cacheService.DecrementAndGetAvailableSeats(ctx, sectionID)
	if err != nil {
		available, getErr := s.cacheService.GetAvailableSeats(ctx, sectionID)
		if getErr == nil && available <= 0 {
			position, waitlistErr := s.addToWaitlist(ctx, studentID, sectionID)
			if waitlistErr != nil {
				logger.Error("Failed to add to waitlist: %v", waitlistErr)
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

		logger.Error("Failed to decrement seats in cache: %v", err)
		return RegistrationResult{
			SectionID: sectionID,
			Status:    "failed",
			Message:   "Failed to process registration",
		}
	}
	logger.Info("Successfully reserved seat for student %s in section %s, remaining seats: %d", studentID, sectionID, newSeatCount)

	dbSyncJob := interfaces.DatabaseSyncJob{
		JobType:   "create_registration",
		StudentID: studentID,
		SectionID: sectionID,
		Timestamp: time.Now(),
	}

	if err := s.queueService.EnqueueDatabaseSync(ctx, dbSyncJob); err != nil {
		logger.Error("Failed to enqueue database sync job, rolling back cache: %v", err)
		if rollbackErr := s.cacheService.IncrementAvailableSeats(ctx, sectionID); rollbackErr != nil {
			logger.Error("Failed to rollback cache after sync job failure: %v", rollbackErr)
		}
		return RegistrationResult{
			SectionID: sectionID,
			Status:    "failed",
			Message:   "Failed to process registration",
		}
	}

	return RegistrationResult{
		SectionID: sectionID,
		Status:    "enrolled",
		Message:   "Registration completed successfully",
	}
}

func (s *RegistrationService) ProcessDatabaseSyncJob(ctx context.Context, job interfaces.DatabaseSyncJob) error {
	logger.Info("Processing database sync job: %s for student %s and section %s", job.JobType, job.StudentID, job.SectionID)

	switch job.JobType {
	case "create_registration":
		return s.createRegistrationRecord(ctx, job.StudentID, job.SectionID)
	case "update_seats":
		return s.updateSectionSeats(ctx, job.SectionID)
	default:
		return fmt.Errorf("unknown job type: %s", job.JobType)
	}
}

func (s *RegistrationService) createRegistrationRecord(ctx context.Context, studentID, sectionID uuid.UUID) error {
	existing, err := s.registrationRepo.GetByStudentAndSection(ctx, studentID, sectionID)
	if err == nil && existing != nil {
		logger.Info("Registration already exists for student %s and section %s", studentID, sectionID)
		return nil
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
		logger.Error("Failed to create registration record: %v", err)
		return fmt.Errorf("failed to create registration: %w", err)
	}

	logger.Info("Successfully created registration record for student %s in section %s", studentID, sectionID)
	return nil
}

func (s *RegistrationService) updateSectionSeats(ctx context.Context, sectionID uuid.UUID) error {
	cachedSeats, err := s.cacheService.GetAvailableSeats(ctx, sectionID)
	if err != nil {
		logger.Error("Failed to get cached seat count for section %s: %v", sectionID, err)
		return fmt.Errorf("failed to get cached seat count: %w", err)
	}

	section, err := s.sectionRepo.GetByID(ctx, sectionID)
	if err != nil {
		return fmt.Errorf("failed to get section: %w", err)
	}

	section.AvailableSeats = cachedSeats
	section.Version++

	if err := s.sectionRepo.UpdateWithOptimisticLock(ctx, section); err != nil {
		logger.Error("Failed to update section seat count: %v", err)
		return fmt.Errorf("failed to update section: %w", err)
	}

	logger.Info("Successfully synchronized seat count for section %s to %d", sectionID, cachedSeats)
	return nil
}

func (s *RegistrationService) addToWaitlist(ctx context.Context, studentID, sectionID uuid.UUID) (int, error) {
	position, err := s.waitlistRepo.GetNextPosition(ctx, sectionID)
	if err != nil {
		return 0, fmt.Errorf("failed to get waitlist position: %w", err)
	}

	waitlistJob := interfaces.WaitlistJob{
		StudentID: studentID,
		SectionID: sectionID,
		Position:  position,
		Timestamp: time.Now(),
	}

	if err := s.queueService.EnqueueWaitlistEntry(ctx, waitlistJob); err != nil {
		return 0, fmt.Errorf("failed to enqueue waitlist entry: %w", err)
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

	newSeatCount, err := s.cacheService.IncrementAndGetAvailableSeats(ctx, sectionID)
	if err != nil {
		logger.Error("Failed to increment seats in cache: %v", err)
		return fmt.Errorf("failed to update seat availability: %w", err)
	}

	logger.Info("Successfully freed seat for section %s, new seat count: %d", sectionID, newSeatCount)
	registration.Status = domain.StatusDropped
	registration.UpdatedAt = time.Now()

	if err := s.registrationRepo.Update(ctx, registration); err != nil {
		logger.Error("Failed to update registration, rolling back cache: %v", err)
		if rollbackErr := s.cacheService.DecrementAvailableSeats(ctx, sectionID); rollbackErr != nil {
			logger.Error("Failed to rollback cache after DB failure: %v", rollbackErr)
		}
		return fmt.Errorf("failed to update registration: %w", err)
	}

	dbSyncJob := interfaces.DatabaseSyncJob{
		JobType:   "update_seats",
		StudentID: studentID,
		SectionID: sectionID,
		Timestamp: time.Now(),
	}

	if err := s.queueService.EnqueueDatabaseSync(ctx, dbSyncJob); err != nil {
		logger.Warn("Failed to enqueue database sync job for seat update: %v", err)
	}

	if err := s.queueService.EnqueueWaitlistProcessing(ctx, sectionID); err != nil {
		logger.Error("Failed to process waitlist after course drop: %v", err)
	}

	logger.Info("Course drop completed for student %s and section %s", studentID, sectionID)
	return nil
}

func (s *RegistrationService) ProcessWaitlistJob(ctx context.Context, job interfaces.WaitlistJob) error {
	logger.Info("Processing waitlist job for student %s and section %s at position %d", job.StudentID, job.SectionID, job.Position)

	waitlistEntry := &domain.WaitlistEntry{
		WaitlistID: uuid.New(),
		StudentID:  job.StudentID,
		SectionID:  job.SectionID,
		Position:   job.Position,
		Timestamp:  job.Timestamp,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.waitlistRepo.Create(ctx, waitlistEntry); err != nil {
		return fmt.Errorf("failed to create waitlist entry: %w", err)
	}

	logger.Info("Successfully created waitlist entry for student %s in section %s at position %d", job.StudentID, job.SectionID, job.Position)
	return nil
}

func (s *RegistrationService) ProcessWaitlist(ctx context.Context, sectionID uuid.UUID) error {
	return s.processWaitlist(ctx, sectionID)
}

func (s *RegistrationService) processWaitlist(ctx context.Context, sectionID uuid.UUID) error {

	nextEntry, err := s.waitlistRepo.GetNextInLine(ctx, sectionID)
	if err != nil || nextEntry == nil {
		return nil
	}

	available, err := s.cacheService.GetAvailableSeats(ctx, sectionID)
	if err != nil || available <= 0 {
		return nil
	}
	newSeatCount, err := s.cacheService.DecrementAndGetAvailableSeats(ctx, sectionID)
	if err != nil {
		return nil
	}
	if err := s.waitlistRepo.Delete(ctx, nextEntry.WaitlistID); err != nil {

		if rollbackErr := s.cacheService.IncrementAvailableSeats(ctx, sectionID); rollbackErr != nil {
			logger.Error("Failed to rollback cache after waitlist removal failure: %v", rollbackErr)
		}
		return fmt.Errorf("failed to remove from waitlist: %w", err)
	}

	dbSyncJob := interfaces.DatabaseSyncJob{
		JobType:   "create_registration",
		StudentID: nextEntry.StudentID,
		SectionID: sectionID,
		Timestamp: time.Now(),
	}

	if err := s.queueService.EnqueueDatabaseSync(ctx, dbSyncJob); err != nil {
		logger.Error("Failed to enqueue database sync job for waitlisted student: %v", err)

	}

	logger.Info("Successfully processed waitlist entry for student %s in section %s, remaining seats: %d",
		nextEntry.StudentID, sectionID, newSeatCount)

	return nil
}

func (s *RegistrationService) GetStudentRegistrations(ctx context.Context, studentID uuid.UUID) ([]*domain.Registration, error) {
	logger.Info("Getting registrations for student %s", studentID)

	registrations, err := s.registrationRepo.GetByStudentID(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get student registrations: %w", err)
	}

	return registrations, nil
}

func (s *RegistrationService) GetStudentWaitlistStatus(ctx context.Context, studentID uuid.UUID) ([]*domain.WaitlistEntry, error) {
	logger.Info("Getting waitlist status for student %s", studentID)

	waitlistEntries, err := s.waitlistRepo.GetByStudentID(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get student waitlist status: %w", err)
	}

	return waitlistEntries, nil
}
func (s *RegistrationService) GetAvailableSections(ctx context.Context, semesterID uuid.UUID, courseID *uuid.UUID) ([]*domain.Section, error) {
	logger.Info("Getting available sections for semester %s", semesterID)

	var sections []*domain.Section
	var err error

	if courseID != nil {
		sections, err = s.sectionRepo.GetByCourseAndSemester(ctx, *courseID, semesterID)
	} else {
		sections, err = s.sectionRepo.GetBySemester(ctx, semesterID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get available sections: %w", err)
	}
	availableSections := make([]*domain.Section, 0)
	for _, section := range sections {
		if section.AvailableSeats > 0 {
			availableSections = append(availableSections, section)
		}
	}

	return availableSections, nil
}
