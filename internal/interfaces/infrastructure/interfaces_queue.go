package interfaces

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type JobType string

const (
	JobTypeCreateRegistration JobType = "create_registration"
	JobTypeUpdateSeats        JobType = "update_seats"
	JobTypeDropRegistration   JobType = "drop_registration"
)

type Status string

const (
	StatusEnrolled   Status = "enrolled"
	StatusFailed     Status = "failed"
	StatusDropped    Status = "dropped"
	StatusWaitlisted Status = "waitlisted"
)

type RegistrationJob struct {
	StudentID uuid.UUID `json:"student_id"`
	SectionID uuid.UUID `json:"section_id"`
	Timestamp time.Time `json:"timestamp"`
	Attempts  int       `json:"attempts"`
}

type DatabaseSyncJob struct {
	JobType   JobType   `json:"job_type"` // "create_registration", "update_seats", "drop_registration"
	Status    Status    `json:"status"`   // "enrolled", "failed", "dropped", "waitlisted"
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
	SetRegistrationService(service interface{})
	StartWorkers()
	StopWorkers()
}
