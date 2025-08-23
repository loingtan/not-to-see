package interfaces

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type RegistrationJob struct {
	StudentID uuid.UUID `json:"student_id"`
	SectionID uuid.UUID `json:"section_id"`
	Timestamp time.Time `json:"timestamp"`
	Attempts  int       `json:"attempts"`
}

type DatabaseSyncJob struct {
	JobType   string    `json:"job_type"` // "create_registration", "update_seats"
	StudentID uuid.UUID `json:"student_id"`
	SectionID uuid.UUID `json:"section_id"`
	Timestamp time.Time `json:"timestamp"`
}

type WaitlistJob struct {
	StudentID uuid.UUID `json:"student_id"`
	SectionID uuid.UUID `json:"section_id"`
	Position  int       `json:"position"`
	Timestamp time.Time `json:"timestamp"`
}

type QueueService interface {
	EnqueueDatabaseSync(ctx context.Context, job DatabaseSyncJob) error
	DequeueDatabaseSync(ctx context.Context) (*DatabaseSyncJob, error)
	EnqueueWaitlistProcessing(ctx context.Context, sectionID uuid.UUID) error
	DequeueWaitlistProcessing(ctx context.Context) (uuid.UUID, error)
	EnqueueWaitlistEntry(ctx context.Context, job WaitlistJob) error
	DequeueWaitlistEntry(ctx context.Context) (*WaitlistJob, error)

	// Worker management methods
	SetRegistrationService(service interface{}) // Using interface{} to avoid circular dependency
	StartWorkers()
	StopWorkers()
}
