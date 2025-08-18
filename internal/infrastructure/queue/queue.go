package queue

import (
	interfaces "cobra-template/internal/interfaces/infrastructure"
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type Queue struct {
	registrationQueue chan interfaces.RegistrationJob
	waitlistQueue     chan uuid.UUID
	mu                sync.RWMutex
}

// NewInMemoryQueue creates a new  in-memory queue
func NewInMemoryQueue(bufferSize, workers int) interfaces.QueueService {
	return &Queue{
		registrationQueue: make(chan interfaces.RegistrationJob, bufferSize),
		waitlistQueue:     make(chan uuid.UUID, bufferSize),
	}
}

// EnqueueRegistration adds a registration job to the queue
func (q *Queue) EnqueueRegistration(ctx context.Context, job interfaces.RegistrationJob) error {
	select {
	case q.registrationQueue <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("registration queue is full")
	}
}

// DequeueRegistration gets a registration job from the queue
func (q *Queue) DequeueRegistration(ctx context.Context) (*interfaces.RegistrationJob, error) {
	select {
	case job := <-q.registrationQueue:
		return &job, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// EnqueueWaitlistProcessing adds a section to waitlist processing queue
func (q *Queue) EnqueueWaitlistProcessing(ctx context.Context, sectionID uuid.UUID) error {
	select {
	case q.waitlistQueue <- sectionID:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("waitlist queue is full")
	}
}

// EnqueueFailedJob adds a failed job to the dead letter queue (simplified)
func (q *Queue) EnqueueFailedJob(ctx context.Context, job interfaces.RegistrationJob, reason string) error {
	// For simplicity, just log the failed job
	fmt.Printf("Failed job: StudentID=%s, SectionID=%s, Reason=%s\n",
		job.StudentID, job.SectionID, reason)
	return nil
}
func (q *Queue) GetQueueLength(ctx context.Context, queueName string) (int, error) {
	switch queueName {
	case "registration":
		return len(q.registrationQueue), nil
	case "waitlist":
		return len(q.waitlistQueue), nil
	default:
		return 0, fmt.Errorf("unknown queue: %s", queueName)
	}
}

var _ interfaces.QueueService = (*Queue)(nil)
