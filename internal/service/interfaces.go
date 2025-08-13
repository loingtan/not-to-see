package service

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// CacheService defines the interface for caching operations
type CacheService interface {
	// Seat availability operations
	GetAvailableSeats(ctx context.Context, sectionID uuid.UUID) (int, error)
	SetAvailableSeats(ctx context.Context, sectionID uuid.UUID, seats int, ttl time.Duration) error
	DecrementAvailableSeats(ctx context.Context, sectionID uuid.UUID) error
	IncrementAvailableSeats(ctx context.Context, sectionID uuid.UUID) error
	
	// Section details caching
	GetSectionDetails(ctx context.Context, sectionID uuid.UUID) (interface{}, error)
	SetSectionDetails(ctx context.Context, sectionID uuid.UUID, data interface{}, ttl time.Duration) error
	
	// Course details caching
	GetCourseDetails(ctx context.Context, courseID uuid.UUID) (interface{}, error)
	SetCourseDetails(ctx context.Context, courseID uuid.UUID, data interface{}, ttl time.Duration) error
	
	// Generic operations
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context, pattern string) error
}

// QueueService defines the interface for queue operations
type QueueService interface {
	// Registration queue operations
	EnqueueRegistration(ctx context.Context, job RegistrationJob) error
	DequeueRegistration(ctx context.Context) (*RegistrationJob, error)
	
	// Waitlist processing queue
	EnqueueWaitlistProcessing(ctx context.Context, sectionID uuid.UUID) error
	
	// Dead letter queue for failed jobs
	EnqueueFailedJob(ctx context.Context, job RegistrationJob, reason string) error
	
	// Queue health and monitoring
	GetQueueLength(ctx context.Context, queueName string) (int, error)
}

// RegistrationJob represents a registration job in the queue
type RegistrationJob struct {
	StudentID uuid.UUID `json:"student_id"`
	SectionID uuid.UUID `json:"section_id"`
	Timestamp time.Time `json:"timestamp"`
	Attempts  int       `json:"attempts"`
}

// NotificationService defines the interface for student notifications
type NotificationService interface {
	SendRegistrationSuccess(ctx context.Context, studentID, sectionID uuid.UUID) error
	SendRegistrationFailure(ctx context.Context, studentID, sectionID uuid.UUID, reason string) error
	SendWaitlistNotification(ctx context.Context, studentID, sectionID uuid.UUID, position int) error
	SendSeatAvailable(ctx context.Context, studentID, sectionID uuid.UUID) error
}
