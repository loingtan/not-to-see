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
type QueueService interface {
	EnqueueRegistration(ctx context.Context, job RegistrationJob) error
	DequeueRegistration(ctx context.Context) (*RegistrationJob, error)

	EnqueueWaitlistProcessing(ctx context.Context, sectionID uuid.UUID) error

	EnqueueFailedJob(ctx context.Context, job RegistrationJob, reason string) error

	GetQueueLength(ctx context.Context, queueName string) (int, error)
}
