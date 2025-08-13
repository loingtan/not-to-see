package queue

import (
	"context"
	"fmt"
	"sync"

	"cobra-template/internal/service"

	"github.com/google/uuid"
)

// SimpleQueue implements QueueService using Go channels
type SimpleQueue struct {
	registrationQueue chan service.RegistrationJob
	waitlistQueue     chan uuid.UUID
	mu                sync.RWMutex
}

// NewInMemoryQueue creates a new simple in-memory queue
func NewInMemoryQueue(bufferSize, workers int) service.QueueService {
	return &SimpleQueue{
		registrationQueue: make(chan service.RegistrationJob, bufferSize),
		waitlistQueue:     make(chan uuid.UUID, bufferSize),
	}
}

// EnqueueRegistration adds a registration job to the queue
func (q *SimpleQueue) EnqueueRegistration(ctx context.Context, job service.RegistrationJob) error {
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
func (q *SimpleQueue) DequeueRegistration(ctx context.Context) (*service.RegistrationJob, error) {
	select {
	case job := <-q.registrationQueue:
		return &job, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// EnqueueWaitlistProcessing adds a section to waitlist processing queue
func (q *SimpleQueue) EnqueueWaitlistProcessing(ctx context.Context, sectionID uuid.UUID) error {
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
func (q *SimpleQueue) EnqueueFailedJob(ctx context.Context, job service.RegistrationJob, reason string) error {
	// For simplicity, just log the failed job
	fmt.Printf("Failed job: StudentID=%s, SectionID=%s, Reason=%s\n",
		job.StudentID, job.SectionID, reason)
	return nil
}

// GetQueueLength returns the current length of the specified queue
func (q *SimpleQueue) GetQueueLength(ctx context.Context, queueName string) (int, error) {
	switch queueName {
	case "registration":
		return len(q.registrationQueue), nil
	case "waitlist":
		return len(q.waitlistQueue), nil
	default:
		return 0, fmt.Errorf("unknown queue: %s", queueName)
	}
}

// Compile-time check to ensure SimpleQueue implements QueueService
var _ service.QueueService = (*SimpleQueue)(nil)
