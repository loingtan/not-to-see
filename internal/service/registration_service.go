package service

import (
	domain "cobra-template/internal/domain/registration"
	interfaces "cobra-template/internal/interfaces/infrastructure"
	serviceInterfaces "cobra-template/internal/interfaces/service"
	"cobra-template/pkg/logger"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	StudentRegistrationsTTL = 20 * time.Minute
	StudentWaitlistTTL      = 15 * time.Minute
	AvailableSectionsTTL    = 8 * time.Minute
	StudentDetailsTTL       = 8 * time.Hour
	CourseDetailsTTL        = 8 * time.Hour
	SectionDetailsTTL       = 45 * time.Minute

	HTTPResponseTTL   = 5 * time.Minute
	ShortTermCacheTTL = 2 * time.Minute
	LongTermCacheTTL  = 2 * time.Hour
)

var _ serviceInterfaces.RegistrationService = (*RegistrationService)(nil)

type RegistrationService struct {
	studentRepo             interfaces.StudentRepository
	sectionRepo             interfaces.SectionRepository
	registrationRepo        interfaces.RegistrationRepository
	waitlistRepo            interfaces.WaitlistRepository
	cacheService            interfaces.CacheService
	queueService            interfaces.QueueService
	idempotencyRepo         interfaces.IdempotencyRepository
	waitlistFallbackEnabled bool
}

func NewRegistrationService(
	studentRepo interfaces.StudentRepository,
	sectionRepo interfaces.SectionRepository,
	registrationRepo interfaces.RegistrationRepository,
	waitlistRepo interfaces.WaitlistRepository,
	cacheService interfaces.CacheService,
	queueService interfaces.QueueService,
	idempotencyRepo interfaces.IdempotencyRepository,
	waitlistFallbackEnabled bool,
) *RegistrationService {
	return &RegistrationService{
		studentRepo:             studentRepo,
		sectionRepo:             sectionRepo,
		registrationRepo:        registrationRepo,
		waitlistRepo:            waitlistRepo,
		cacheService:            cacheService,
		queueService:            queueService,
		idempotencyRepo:         idempotencyRepo,
		waitlistFallbackEnabled: waitlistFallbackEnabled,
	}
}

type RegisterRequest = serviceInterfaces.RegisterRequest
type RegisterResponse = serviceInterfaces.RegisterResponse
type RegistrationResult = serviceInterfaces.RegistrationResult

func (s *RegistrationService) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	logger.Info("Processing registration for student %s with %d sections", req.StudentID, len(req.SectionIDs))

	if req.IdempotencyKey != "" {
		existingKey, isDuplicate, err := s.checkIdempotency(ctx, req.IdempotencyKey, req.StudentID, req)
		if err != nil {
			return nil, fmt.Errorf("idempotency check failed: %w", err)
		}
		if isDuplicate {
			var cachedResponse RegisterResponse
			if err := json.Unmarshal([]byte(existingKey.ResponseData), &cachedResponse); err == nil {
				logger.Info("Returning cached response for idempotency key: %s", req.IdempotencyKey)
				return &cachedResponse, nil
			}
		}
	}

	student, err := s.GetStudentDetails(ctx, req.StudentID)
	if err != nil {
		return nil, fmt.Errorf("student not found: %w", err)
	}
	if student == nil {
		return nil, errors.New("student not found")
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

	if req.IdempotencyKey != "" {
		if err := s.storeIdempotencyResult(ctx, req.IdempotencyKey, req.StudentID, req, response, 200); err != nil {
			logger.Warn("Failed to store idempotency result: %v", err)
		}
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
		// If seat key not found, try to initialize it from database
		if strings.Contains(err.Error(), "seat key not found") {
			logger.Info("Seat key not found for section %s, initializing from database", sectionID)

			// Get section from database to get current seat count
			section, dbErr := s.sectionRepo.GetByID(ctx, sectionID)
			if dbErr != nil {
				logger.Error("Failed to get section from database: %v", dbErr)
				return RegistrationResult{
					SectionID: sectionID,
					Status:    "failed",
					Message:   "Failed to process registration",
				}
			}

			if section == nil {
				return RegistrationResult{
					SectionID: sectionID,
					Status:    "failed",
					Message:   "Section not found",
				}
			}

			// Initialize cache with current database value
			if setErr := s.cacheService.SetAvailableSeats(ctx, sectionID, section.AvailableSeats, 24*time.Hour); setErr != nil {
				logger.Error("Failed to initialize seat cache for section %s: %v", sectionID, setErr)
				return RegistrationResult{
					SectionID: sectionID,
					Status:    "failed",
					Message:   "Failed to process registration",
				}
			}

			// Try to decrement again
			newSeatCount, err = s.cacheService.DecrementAndGetAvailableSeats(ctx, sectionID)
			if err != nil {
				logger.Error("Failed to decrement seats after cache initialization: %v", err)
				return RegistrationResult{
					SectionID: sectionID,
					Status:    "failed",
					Message:   "Failed to process registration",
				}
			}
		} else {
			// Handle other types of errors (no seats available, etc.)
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
	}

	logger.Info("Successfully reserved seat for student %s in section %s, remaining seats: %d", studentID, sectionID, newSeatCount)

	dbSyncJob := interfaces.DatabaseSyncJob{
		JobType:   interfaces.JobTypeCreateRegistration,
		Status:    interfaces.StatusEnrolled,
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

	seatUpdateJob := interfaces.DatabaseSyncJob{
		JobType:   interfaces.JobTypeUpdateSeats,
		SectionID: sectionID,
		Timestamp: time.Now(),
	}
	if err := s.queueService.EnqueueDatabaseSync(ctx, seatUpdateJob); err != nil {
		logger.Warn("Failed to enqueue seat update job: %v", err)
	}
	s.updateStudentRegistrationCache(ctx, studentID, sectionID, domain.StatusEnrolled)
	s.updateAvailableSectionsCacheForSection(ctx, sectionID, newSeatCount)

	return RegistrationResult{
		SectionID: sectionID,
		Status:    "enrolled",
		Message:   "Registration completed successfully",
	}
}

func (s *RegistrationService) ProcessDatabaseSyncJob(ctx context.Context, job interfaces.DatabaseSyncJob) error {
	logger.Info("Processing database sync job: %s for student %s and section %s", job.JobType, job.StudentID, job.SectionID)

	switch job.JobType {
	case interfaces.JobTypeCreateRegistration:
		return s.createRegistrationRecord(ctx, job.StudentID, job.SectionID)
	case interfaces.JobTypeUpdateSeats:
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
	if section == nil {
		return fmt.Errorf("section not found")
	}

	section.AvailableSeats = cachedSeats
	section.Version++
	if err := s.sectionRepo.UpdateWithOptimisticLock(ctx, section); err != nil {
		logger.Error("Failed to update section seat count: %v", err)
		return fmt.Errorf("failed to update section: %w", err)
	}

	// Update section details cache with new seat count
	// Removed: s.updateSectionDetailsCache(ctx, section)

	// Update available sections cache for the specific semester this section belongs to
	s.updateAvailableSectionsCacheForSection(ctx, sectionID, cachedSeats)

	logger.Info("Successfully synchronized seat count for section %s to %d", sectionID, cachedSeats)
	return nil
}

func (s *RegistrationService) addToWaitlist(ctx context.Context, studentID, sectionID uuid.UUID) (int, error) {
	position, err := s.cacheService.GetWaitlistSize(ctx, sectionID)
	if err != nil {
		position, err = s.waitlistRepo.GetNextPosition(ctx, sectionID)
		if err != nil {
			return 0, fmt.Errorf("failed to get waitlist position: %w", err)
		}
	} else {
		position++
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

	if err := s.cacheService.AddToWaitlist(ctx, sectionID, studentID, position, waitlistEntry); err != nil {
		if s.waitlistFallbackEnabled {
			logger.Warn("Failed to add to Redis waitlist, falling back to database queue: %v", err)
			waitlistJob := interfaces.WaitlistJob{
				StudentID: studentID,
				SectionID: sectionID,
				Position:  position,
				Timestamp: time.Now(),
			}

			if err := s.queueService.EnqueueWaitlistEntry(ctx, waitlistJob); err != nil {
				return 0, fmt.Errorf("failed to enqueue waitlist entry: %w", err)
			}

			// Update student waitlist cache
			s.updateStudentWaitlistCache(ctx, studentID, waitlistEntry, "add")

			return position, nil
		} else {
			return 0, fmt.Errorf("failed to add to Redis waitlist and fallback is disabled: %w", err)
		}
	}

	waitlistJob := interfaces.WaitlistJob{
		StudentID: studentID,
		SectionID: sectionID,
		Position:  position,
		Timestamp: time.Now(),
	}

	if err := s.queueService.EnqueueWaitlistEntry(ctx, waitlistJob); err != nil {
		logger.Warn("Failed to enqueue waitlist entry for database persistence: %v", err)
	}

	// Update student waitlist cache
	s.updateStudentWaitlistCache(ctx, studentID, waitlistEntry, "add")

	logger.Info("Successfully added student %s to waitlist for section %s at position %d", studentID, sectionID, position)
	return position, nil
}

func (s *RegistrationService) DropCourse(ctx context.Context, studentID, sectionID uuid.UUID) error {
	logger.Info("Processing course drop for student %s and section %s", studentID.String(), sectionID.String())

	registration, err := s.registrationRepo.GetByStudentAndSection(ctx, studentID, sectionID)
	if err != nil {
		return fmt.Errorf("registration not found: %w", err)
	}
	if registration == nil {
		return errors.New("registration not found")
	}

	if registration.Status != domain.StatusEnrolled {
		return errors.New("can only drop enrolled courses")
	}

	newSeatCount, err := s.cacheService.IncrementAndGetAvailableSeats(ctx, sectionID)
	if err != nil {
		logger.Error("Failed to increment seats in cache: %v", err)
		return fmt.Errorf("failed to update seat availability: %w", err)
	}

	logger.Info("Successfully freed seat for section %s, new seat count: %d", sectionID.String(), newSeatCount)

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
		JobType:   interfaces.JobTypeUpdateSeats,
		StudentID: studentID,
		SectionID: sectionID,
		Timestamp: time.Now(),
	}
	if err := s.queueService.EnqueueDatabaseSync(ctx, dbSyncJob); err != nil {
		logger.Warn("Failed to enqueue database sync job for seat update: %v", err)
	}

	// Update student registration cache instead of deleting
	s.updateStudentRegistrationCache(ctx, studentID, sectionID, domain.StatusDropped)

	// Update available sections cache for all semesters this section belongs to
	s.updateAvailableSectionsCacheForSection(ctx, sectionID, newSeatCount)

	if err := s.queueService.EnqueueWaitlistProcessing(ctx, sectionID); err != nil {
		logger.Error("Failed to process waitlist after course drop: %v", err)
	}

	logger.Info("Course drop completed for student %s and section %s", studentID.String(), sectionID.String())
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
	nextEntryData, err := s.cacheService.GetNextInWaitlist(ctx, sectionID)
	if err != nil || nextEntryData == nil {
		if s.waitlistFallbackEnabled {
			nextEntry, err := s.waitlistRepo.GetNextInLine(ctx, sectionID)
			if err != nil || nextEntry == nil {
				return nil
			}
			return s.processWaitlistFromDB(ctx, sectionID, nextEntry)
		} else {
			if err != nil {
				return fmt.Errorf("failed to get next in waitlist from Redis and fallback is disabled: %w", err)
			}
			return nil
		}
	}

	var nextEntry domain.WaitlistEntry
	entryBytes, err := json.Marshal(nextEntryData)
	if err != nil {
		logger.Error("Failed to marshal waitlist entry from Redis: %v", err)
		if s.waitlistFallbackEnabled {
			dbEntry, err := s.waitlistRepo.GetNextInLine(ctx, sectionID)
			if err != nil || dbEntry == nil {
				return nil
			}
			return s.processWaitlistFromDB(ctx, sectionID, dbEntry)
		} else {
			return fmt.Errorf("failed to process waitlist entry from Redis and fallback is disabled: %w", err)
		}
	}

	if err := json.Unmarshal(entryBytes, &nextEntry); err != nil {
		logger.Error("Failed to unmarshal waitlist entry from Redis: %v", err)
		if s.waitlistFallbackEnabled {
			dbEntry, err := s.waitlistRepo.GetNextInLine(ctx, sectionID)
			if err != nil || dbEntry == nil {
				return nil
			}
			return s.processWaitlistFromDB(ctx, sectionID, dbEntry)
		} else {
			return fmt.Errorf("failed to unmarshal waitlist entry from Redis and fallback is disabled: %w", err)
		}
	}

	return s.processWaitlistFromRedis(ctx, sectionID, &nextEntry)
}

func (s *RegistrationService) processWaitlistFromRedis(ctx context.Context, sectionID uuid.UUID, nextEntry *domain.WaitlistEntry) error {
	available, err := s.cacheService.GetAvailableSeats(ctx, sectionID)
	if err != nil || available <= 0 {
		return nil
	}

	newSeatCount, err := s.cacheService.DecrementAndGetAvailableSeats(ctx, sectionID)
	if err != nil {
		return nil
	}

	if err := s.cacheService.RemoveFromWaitlist(ctx, sectionID, nextEntry.StudentID); err != nil {
		logger.Error("Failed to remove from Redis waitlist: %v", err)
		if rollbackErr := s.cacheService.IncrementAvailableSeats(ctx, sectionID); rollbackErr != nil {
			logger.Error("Failed to rollback cache after Redis waitlist removal failure: %v", rollbackErr)
		}
		return fmt.Errorf("failed to remove from Redis waitlist: %w", err)
	}

	if err := s.waitlistRepo.Delete(ctx, nextEntry.WaitlistID); err != nil {
		logger.Warn("Failed to remove waitlist entry from database: %v", err)
	}

	dbSyncJob := interfaces.DatabaseSyncJob{
		JobType:   interfaces.JobTypeCreateRegistration,
		StudentID: nextEntry.StudentID,
		SectionID: sectionID,
		Timestamp: time.Now(),
	}
	if err := s.queueService.EnqueueDatabaseSync(ctx, dbSyncJob); err != nil {
		logger.Error("Failed to enqueue database sync job for waitlisted student: %v", err)
	}

	// Update caches efficiently instead of invalidating
	s.updateStudentRegistrationCache(ctx, nextEntry.StudentID, sectionID, domain.StatusEnrolled)
	s.updateStudentWaitlistCache(ctx, nextEntry.StudentID, nextEntry, "remove")
	s.updateAvailableSectionsCacheForSection(ctx, sectionID, newSeatCount)

	logger.Info("Successfully processed waitlist entry from Redis for student %s in section %s, remaining seats: %d",
		nextEntry.StudentID, sectionID, newSeatCount)

	return nil
}

func (s *RegistrationService) processWaitlistFromDB(ctx context.Context, sectionID uuid.UUID, nextEntry *domain.WaitlistEntry) error {
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

	if err := s.cacheService.RemoveFromWaitlist(ctx, sectionID, nextEntry.StudentID); err != nil {
		logger.Warn("Failed to remove from Redis waitlist (continuing): %v", err)
	}

	dbSyncJob := interfaces.DatabaseSyncJob{
		JobType:   interfaces.JobTypeCreateRegistration,
		StudentID: nextEntry.StudentID,
		SectionID: sectionID,
		Timestamp: time.Now(),
	}
	if err := s.queueService.EnqueueDatabaseSync(ctx, dbSyncJob); err != nil {
		logger.Error("Failed to enqueue database sync job for waitlisted student: %v", err)
	}

	// Update caches efficiently instead of invalidating
	s.updateStudentRegistrationCache(ctx, nextEntry.StudentID, sectionID, domain.StatusEnrolled)
	s.updateStudentWaitlistCache(ctx, nextEntry.StudentID, nextEntry, "remove")
	s.updateAvailableSectionsCacheForSection(ctx, sectionID, newSeatCount)

	logger.Info("Successfully processed waitlist entry from database for student %s in section %s, remaining seats: %d",
		nextEntry.StudentID, sectionID, newSeatCount)

	return nil
}

// Smart cache update methods

func (s *RegistrationService) updateStudentRegistrationCache(ctx context.Context, studentID, sectionID uuid.UUID, status domain.RegistrationStatus) {
	// Get current cached registrations
	cached, err := s.cacheService.GetStudentRegistrations(ctx, studentID)
	if err != nil {
		// If no cache exists, we'll let it be populated on next read
		return
	}

	var registrations []*domain.Registration
	if rawJSON, ok := cached.(json.RawMessage); ok {
		if err := json.Unmarshal(rawJSON, &registrations); err != nil {
			logger.Warn("Failed to unmarshal cached registrations for student %s: %v", studentID, err)
			return
		}
	} else {
		logger.Warn("Failed to cast cached registrations for student %s to json.RawMessage", studentID)
		return
	}

	// Find and update the specific registration
	found := false
	for _, reg := range registrations {
		if reg.SectionID == sectionID {
			reg.Status = status
			reg.UpdatedAt = time.Now()
			found = true
			break
		}
	}

	// If not found and status is enrolled, add new registration
	if !found && status == domain.StatusEnrolled {
		newReg := &domain.Registration{
			RegistrationID:   uuid.New(),
			StudentID:        studentID,
			SectionID:        sectionID,
			Status:           status,
			RegistrationDate: time.Now(),
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
			Version:          1,
		}
		registrations = append(registrations, newReg)
	}

	// Update cache with modified data
	if err := s.cacheService.SetStudentRegistrations(ctx, studentID, registrations, StudentRegistrationsTTL); err != nil {
		logger.Warn("Failed to update student registrations cache for %s: %v", studentID, err)
	}
}

func (s *RegistrationService) updateStudentWaitlistCache(ctx context.Context, studentID uuid.UUID, entry *domain.WaitlistEntry, action string) {
	// Get current cached waitlist
	cached, err := s.cacheService.GetStudentWaitlistStatus(ctx, studentID)
	if err != nil {
		return
	}

	var waitlistEntries []*domain.WaitlistEntry
	if rawJSON, ok := cached.(json.RawMessage); ok {
		if err := json.Unmarshal(rawJSON, &waitlistEntries); err != nil {
			logger.Warn("Failed to unmarshal cached waitlist status for student %s: %v", studentID, err)
			return
		}
	} else {
		logger.Warn("Failed to cast cached waitlist status for student %s to json.RawMessage", studentID)
		return
	}

	switch action {
	case "add":
		waitlistEntries = append(waitlistEntries, entry)
	case "remove":
		// Remove the specific entry
		for i, we := range waitlistEntries {
			if we.SectionID == entry.SectionID && we.StudentID == entry.StudentID {
				waitlistEntries = append(waitlistEntries[:i], waitlistEntries[i+1:]...)
				break
			}
		}
	}

	// Update cache with modified data
	if err := s.cacheService.SetStudentWaitlistStatus(ctx, studentID, waitlistEntries, StudentWaitlistTTL); err != nil {
		logger.Warn("Failed to update student waitlist cache for %s: %v", studentID, err)
	}
}

func (s *RegistrationService) updateAvailableSectionsCacheForSection(ctx context.Context, sectionID uuid.UUID, newSeatCount int) {
	semesterID := uuid.MustParse("e093bb58-78e2-4985-bb7f-7a9b36c9102d")
	cached, err := s.cacheService.GetAvailableSections(ctx, semesterID)
	if err != nil {
	
		return
	}

	var sections []*domain.Section
	if rawJSON, ok := cached.(json.RawMessage); ok {
		if err := json.Unmarshal(rawJSON, &sections); err != nil {
			logger.Warn("Failed to unmarshal cached available sections for semester %s: %v", semesterID, err)
			return
		}
	} else {
		logger.Warn("Failed to cast cached available sections for semester %s to json.RawMessage", semesterID)
		return
	}

	// Find and update the specific section's available seats
	sectionFound := false
	for _, sec := range sections {
		if sec.SectionID == sectionID {
			sec.AvailableSeats = newSeatCount
			sectionFound = true
			break
		}
	}

	// If section not found in cache and has available seats, add it
	if !sectionFound && newSeatCount > 0 {
		section, err := s.sectionRepo.GetByID(ctx, sectionID)
		if err == nil && section != nil && section.SemesterID == semesterID {
			section.AvailableSeats = newSeatCount
			sections = append(sections, section)
		}
	}

	// Filter sections that still have available seats
	availableSections := make([]*domain.Section, 0)
	for _, sec := range sections {
		if sec.AvailableSeats > 0 {
			availableSections = append(availableSections, sec)
		}
	}

	// Update cache with modified data
	if err := s.cacheService.SetAvailableSections(ctx, semesterID, availableSections, AvailableSectionsTTL); err != nil {
		logger.Warn("Failed to update available sections cache for semester %s: %v", semesterID, err)
	}
}

func (s *RegistrationService) GetStudentRegistrations(ctx context.Context, studentID uuid.UUID) ([]*domain.Registration, error) {
	logger.Info("Getting registrations for student %s", studentID)

	cached, err := s.cacheService.GetStudentRegistrations(ctx, studentID)
	if err == nil {
		logger.Info("Found cached registrations for student %s", studentID)
		if rawJSON, ok := cached.(json.RawMessage); ok {
			var registrations []*domain.Registration
			if err := json.Unmarshal(rawJSON, &registrations); err == nil {
				return registrations, nil
			}
			logger.Warn("Failed to unmarshal cached registrations for student %s: %v", studentID, err)
		} else {
			logger.Warn("Failed to cast cached registrations for student %s to json.RawMessage", studentID)
		}
	}

	registrations, err := s.registrationRepo.GetByStudentID(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get student registrations: %w", err)
	}

	if err := s.cacheService.SetStudentRegistrations(ctx, studentID, registrations, StudentRegistrationsTTL); err != nil {
		logger.Warn("Failed to cache student registrations for %s: %v", studentID, err)
	}

	return registrations, nil
}

func (s *RegistrationService) GetStudentWaitlistStatus(ctx context.Context, studentID uuid.UUID) ([]*domain.WaitlistEntry, error) {
	logger.Info("Getting waitlist status for student %s", studentID)

	waitlistData, err := s.cacheService.GetStudentWaitlists(ctx, studentID)
	if err == nil && len(waitlistData) > 0 {
		logger.Info("Found Redis waitlist status for student %s", studentID)

		waitlistEntries := make([]*domain.WaitlistEntry, 0, len(waitlistData))
		for _, data := range waitlistData {
			entryBytes, err := json.Marshal(data)
			if err != nil {
				logger.Warn("Failed to marshal waitlist entry from Redis: %v", err)
				continue
			}

			var entry domain.WaitlistEntry
			if err := json.Unmarshal(entryBytes, &entry); err != nil {
				logger.Warn("Failed to unmarshal waitlist entry from Redis: %v", err)
				continue
			}

			waitlistEntries = append(waitlistEntries, &entry)
		}

		if len(waitlistEntries) > 0 {
			if err := s.cacheService.SetStudentWaitlistStatus(ctx, studentID, waitlistEntries, StudentWaitlistTTL); err != nil {
				logger.Warn("Failed to cache student waitlist status backup for %s: %v", studentID, err)
			}
			return waitlistEntries, nil
		}
	}

	cached, err := s.cacheService.GetStudentWaitlistStatus(ctx, studentID)
	if err == nil {
		logger.Info("Found cached waitlist status for student %s", studentID)
		if rawJSON, ok := cached.(json.RawMessage); ok {
			var waitlistEntries []*domain.WaitlistEntry
			if err := json.Unmarshal(rawJSON, &waitlistEntries); err == nil {
				return waitlistEntries, nil
			}
			logger.Warn("Failed to unmarshal cached waitlist status for %s: %v", studentID, err)
		} else {
			logger.Warn("Failed to cast cached waitlist status for student %s to json.RawMessage", studentID)
		}
	}

	if s.waitlistFallbackEnabled {
		logger.Info("Fetching waitlist status from database for student %s", studentID)
		waitlistEntries, err := s.waitlistRepo.GetByStudentID(ctx, studentID)
		if err != nil {
			return nil, fmt.Errorf("failed to get student waitlist status: %w", err)
		}

		if err := s.cacheService.SetStudentWaitlistStatus(ctx, studentID, waitlistEntries, StudentWaitlistTTL); err != nil {
			logger.Warn("Failed to cache student waitlist status for %s: %v", studentID, err)
		}

		for _, entry := range waitlistEntries {
			if err := s.cacheService.AddToWaitlist(ctx, entry.SectionID, entry.StudentID, entry.Position, entry); err != nil {
				logger.Warn("Failed to populate Redis waitlist for student %s, section %s: %v", studentID, entry.SectionID, err)
			}
		}

		return waitlistEntries, nil
	}

	logger.Info("No waitlist data found for student %s and database fallback is disabled", studentID)
	return []*domain.WaitlistEntry{}, nil
}

func (s *RegistrationService) GetAvailableSections(ctx context.Context, semesterID uuid.UUID) ([]*domain.Section, error) {
	logger.Info("Getting available sections for semester %s", semesterID)

	cached, err := s.cacheService.GetAvailableSections(ctx, semesterID)
	if err == nil {
		logger.Info("Found cached available sections for semester %s", semesterID)
		if rawJSON, ok := cached.(json.RawMessage); ok {
			var sections []*domain.Section
			if err := json.Unmarshal(rawJSON, &sections); err == nil {
				// Update with real-time seat counts from cache
				updatedSections := make([]*domain.Section, 0, len(sections))
				for _, section := range sections {
					if cachedSeats, cacheErr := s.cacheService.GetAvailableSeats(ctx, section.SectionID); cacheErr == nil {
						// Create a copy to avoid modifying the cached object
						updatedSection := *section
						updatedSection.AvailableSeats = cachedSeats
						if updatedSection.AvailableSeats > 0 {
							updatedSections = append(updatedSections, &updatedSection)
						}
					} else if section.AvailableSeats > 0 {
						updatedSections = append(updatedSections, section)
					}
				}
				return updatedSections, nil
			}
			logger.Warn("Failed to unmarshal cached available sections for semester %s: %v", semesterID, err)
		} else {
			logger.Warn("Failed to cast cached available sections for semester %s to json.RawMessage", semesterID)
		}
	}

	sections, err := s.sectionRepo.GetBySemester(ctx, semesterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get available sections: %w", err)
	}

	availableSections := make([]*domain.Section, 0)
	for _, section := range sections {

		if cachedSeats, cacheErr := s.cacheService.GetAvailableSeats(ctx, section.SectionID); cacheErr == nil {
			section.AvailableSeats = cachedSeats
		}

		if section.AvailableSeats > 0 {
			availableSections = append(availableSections, section)
		}
	}

	if err := s.cacheService.SetAvailableSections(ctx, semesterID, availableSections, AvailableSectionsTTL); err != nil {
		logger.Warn("Failed to cache available sections for semester %s: %v", semesterID, err)
	}

	return availableSections, nil
}

func (s *RegistrationService) GetStudentDetails(ctx context.Context, studentID uuid.UUID) (*domain.Student, error) {
	logger.Info("Getting student details for %s", studentID)

	cached, err := s.cacheService.GetStudentDetails(ctx, studentID)
	if err == nil {
		logger.Info("Found cached student details for %s", studentID)
		if rawJSON, ok := cached.(json.RawMessage); ok {
			var student domain.Student
			if err := json.Unmarshal(rawJSON, &student); err == nil {
				return &student, nil
			}
			logger.Warn("Failed to unmarshal cached student details for %s: %v", studentID, err)
		} else {
			logger.Warn("Failed to cast cached student details for %s to json.RawMessage", studentID)
		}
	}

	student, err := s.studentRepo.GetByID(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get student details: %w", err)
	}

	if student == nil {
		return nil, fmt.Errorf("student not found")
	}

	if err := s.cacheService.SetStudentDetails(ctx, studentID, student, StudentDetailsTTL); err != nil {
		logger.Warn("Failed to cache student details for %s: %v", studentID, err)
	}

	return student, nil
}

func (s *RegistrationService) GetCourseDetails(ctx context.Context, courseID uuid.UUID) (*domain.Course, error) {
	logger.Info("Getting course details for %s", courseID)

	cached, err := s.cacheService.GetCourseDetails(ctx, courseID)
	if err == nil {
		logger.Info("Found cached course details for %s", courseID)
		if course, ok := cached.(*domain.Course); ok {
			return course, nil
		}
		logger.Warn("Failed to cast cached course details for %s", courseID)
	}

	return nil, fmt.Errorf("course repository not available in registration service")
}

func (s *RegistrationService) GetSectionDetails(ctx context.Context, sectionID uuid.UUID) (*domain.Section, error) {
	logger.Info("Getting section details for %s", sectionID)

	section, err := s.sectionRepo.GetByID(ctx, sectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get section details: %w", err)
	}

	if section == nil {
		return nil, fmt.Errorf("section not found")
	}

	// Update with real-time seat count from cache
	if cachedSeats, cacheErr := s.cacheService.GetAvailableSeats(ctx, sectionID); cacheErr == nil {
		section.AvailableSeats = cachedSeats
	}

	return section, nil
}

func (s *RegistrationService) RefreshSectionCache(ctx context.Context, sectionID uuid.UUID) error {
	// Get fresh section data from database
	section, err := s.sectionRepo.GetByID(ctx, sectionID)
	if err != nil {
		return fmt.Errorf("failed to get section from database: %w", err)
	}

	if section == nil {
		return fmt.Errorf("section not found")
	}

	// Update with current cached seat count
	if cachedSeats, cacheErr := s.cacheService.GetAvailableSeats(ctx, sectionID); cacheErr == nil {
		section.AvailableSeats = cachedSeats
	}

	// Update available sections cache for the semester
	s.updateAvailableSectionsCacheForSection(ctx, sectionID, section.AvailableSeats)

	logger.Info("Successfully refreshed cache for section %s", sectionID)
	return nil
}

func (s *RegistrationService) RefreshAllSectionCaches(ctx context.Context) error {
	logger.Info("Starting bulk refresh of all section seat caches")

	// Get all active sections from database
	sections, err := s.sectionRepo.GetAllActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active sections: %w", err)
	}

	cached := 0
	failed := 0

	for _, section := range sections {
		if err := s.cacheService.SetAvailableSeats(ctx, section.SectionID, section.AvailableSeats, 24*time.Hour); err != nil {
			logger.Warn("Failed to cache seats for section %s: %v", section.SectionID, err)
			failed++
			continue
		}
		cached++
	}

	logger.Info("Bulk cache refresh completed: %d sections cached, %d failed", cached, failed)

	if failed > 0 {
		return fmt.Errorf("failed to cache %d sections", failed)
	}

	return nil
}

func (s *RegistrationService) InvalidateStudentCaches(ctx context.Context, studentID uuid.UUID) {
	// Only use this when we need to force a cache refresh
	keys := []string{
		fmt.Sprintf("student:registrations:%s", studentID.String()),
		fmt.Sprintf("student:waitlist:%s", studentID.String()),
		fmt.Sprintf("student:details:%s", studentID.String()),
	}

	for _, key := range keys {
		if err := s.cacheService.Delete(ctx, key); err != nil {
			logger.Warn("Failed to delete cache key %s: %v", key, err)
		}
	}

	logger.Info("Invalidated caches for student %s", studentID)
}

func (s *RegistrationService) WarmupCaches(ctx context.Context, studentID uuid.UUID) error {
	// Pre-populate caches with fresh data
	go func() {
		if _, err := s.GetStudentDetails(ctx, studentID); err != nil {
			logger.Warn("Failed to warmup student details cache: %v", err)
		}

		if _, err := s.GetStudentRegistrations(ctx, studentID); err != nil {
			logger.Warn("Failed to warmup student registrations cache: %v", err)
		}

		if _, err := s.GetStudentWaitlistStatus(ctx, studentID); err != nil {
			logger.Warn("Failed to warmup student waitlist cache: %v", err)
		}
	}()

	return nil
}

func (s *RegistrationService) checkIdempotency(ctx context.Context, key string, studentID uuid.UUID, requestData interface{}) (*domain.IdempotencyKey, bool, error) {
	if key == "" {
		return nil, false, nil
	}

	existingKey, err := s.idempotencyRepo.GetByKey(ctx, key)
	if err != nil {
		if err.Error() == "idempotency key not found" {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("failed to check idempotency key: %w", err)
	}

	if existingKey != nil {
		if existingKey.IsExpired() {
			if err := s.idempotencyRepo.Delete(ctx, key); err != nil {
				logger.Warn("Failed to delete expired idempotency key %s: %v", key, err)
			}
			return nil, false, nil
		}

		requestHash := s.generateRequestHash(studentID, requestData)
		if existingKey.RequestHash == requestHash {
			return existingKey, true, nil
		} else {
			return nil, false, fmt.Errorf("idempotency key already used with different request data")
		}
	}

	return nil, false, nil
}

func (s *RegistrationService) storeIdempotencyResult(ctx context.Context, key string, studentID uuid.UUID, requestData interface{}, responseData interface{}, statusCode int) error {
	if key == "" {
		return nil
	}

	requestHash := s.generateRequestHash(studentID, requestData)

	responseJSON, err := json.Marshal(responseData)
	if err != nil {
		return fmt.Errorf("failed to marshal response data: %w", err)
	}

	idempotencyKey := &domain.IdempotencyKey{
		Key:          key,
		StudentID:    studentID,
		RequestHash:  requestHash,
		ResponseData: string(responseJSON),
		StatusCode:   statusCode,
		ProcessedAt:  time.Now(),
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Now(),
	}

	return s.idempotencyRepo.Create(ctx, idempotencyKey)
}

func (s *RegistrationService) generateRequestHash(studentID uuid.UUID, requestData interface{}) string {
	data := map[string]any{
		"student_id":   studentID.String(),
		"request_data": requestData,
	}

	jsonData, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])
}

// ensureSeatCacheInitialized ensures that the seat count for a section is cached in Redis
// If not cached, it fetches from database and initializes the cache
func (s *RegistrationService) ensureSeatCacheInitialized(ctx context.Context, sectionID uuid.UUID) error {
	// Check if seat count is already cached
	_, err := s.cacheService.GetAvailableSeats(ctx, sectionID)
	if err == nil {
		// Already cached, nothing to do
		return nil
	}

	// Not cached, fetch from database and initialize cache
	logger.Info("Initializing seat cache for section %s from database", sectionID)

	section, dbErr := s.sectionRepo.GetByID(ctx, sectionID)
	if dbErr != nil {
		return fmt.Errorf("failed to get section from database: %w", dbErr)
	}

	if section == nil {
		return fmt.Errorf("section not found")
	}

	// Initialize cache with current database value and set 24-hour TTL
	if setErr := s.cacheService.SetAvailableSeats(ctx, sectionID, section.AvailableSeats, 24*time.Hour); setErr != nil {
		return fmt.Errorf("failed to initialize seat cache: %w", setErr)
	}

	logger.Info("Successfully initialized seat cache for section %s with %d seats", sectionID, section.AvailableSeats)
	return nil
}
