package queue

import (
	interfaces "cobra-template/internal/interfaces/infrastructure"
	serviceInterfaces "cobra-template/internal/interfaces/service"
	"cobra-template/pkg/logger"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Queue struct {
	databaseSyncQueue  chan interfaces.DatabaseSyncJob
	waitlistQueue      chan uuid.UUID
	waitlistEntryQueue chan interfaces.WaitlistJob

	workers int
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	started bool
	mu      sync.RWMutex

	registrationService serviceInterfaces.RegistrationService
}

func NewInMemoryQueue(bufferSize, workers int) interfaces.QueueService {
	ctx, cancel := context.WithCancel(context.Background())

	queue := &Queue{
		databaseSyncQueue:  make(chan interfaces.DatabaseSyncJob, bufferSize),
		waitlistQueue:      make(chan uuid.UUID, bufferSize),
		waitlistEntryQueue: make(chan interfaces.WaitlistJob, bufferSize),
		workers:            workers,
		ctx:                ctx,
		cancel:             cancel,
		started:            false,
	}

	return queue
}

func (q *Queue) SetRegistrationService(service interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if regService, ok := service.(serviceInterfaces.RegistrationService); ok {
		q.registrationService = regService
	} else {
		logger.Error("Invalid service type provided to SetRegistrationService")
	}
}

func (q *Queue) StartWorkers() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.started {
		return
	}

	if q.registrationService == nil {
		logger.Warn("Registration service not set, workers cannot process jobs")
		return
	}

	logger.Info("Starting %d queue workers", q.workers)

	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.databaseSyncWorker(i)
	}

	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.waitlistProcessingWorker(i)
	}

	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.waitlistEntryWorker(i)
	}

	q.started = true
	logger.Info("Queue workers started successfully")
}

func (q *Queue) StopWorkers() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.started {
		return
	}

	logger.Info("Stopping queue workers...")
	q.cancel()
	q.wg.Wait()
	q.started = false
	logger.Info("Queue workers stopped")
}

func (q *Queue) EnqueueDatabaseSync(ctx context.Context, job interfaces.DatabaseSyncJob) error {
	select {
	case q.databaseSyncQueue <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("database sync queue is full")
	}
}

func (q *Queue) DequeueDatabaseSync(ctx context.Context) (*interfaces.DatabaseSyncJob, error) {
	select {
	case job := <-q.databaseSyncQueue:
		return &job, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

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
func (q *Queue) DequeueWaitlistProcessing(ctx context.Context) (uuid.UUID, error) {
	select {
	case id := <-q.waitlistQueue:
		return id, nil
	case <-ctx.Done():
		return uuid.UUID{}, ctx.Err()
	}
}

func (q *Queue) EnqueueWaitlistEntry(ctx context.Context, job interfaces.WaitlistJob) error {
	select {
	case q.waitlistEntryQueue <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("waitlist entry queue is full")
	}
}

func (q *Queue) DequeueWaitlistEntry(ctx context.Context) (*interfaces.WaitlistJob, error) {
	select {
	case job := <-q.waitlistEntryQueue:
		return &job, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (q *Queue) databaseSyncWorker(workerID int) {
	defer q.wg.Done()

	logger.Info("Database sync worker %d started", workerID)

	for {
		select {
		case <-q.ctx.Done():
			logger.Info("Database sync worker %d stopped", workerID)
			return
		default:

			ctx, cancel := context.WithTimeout(q.ctx, 5*time.Second)
			job, err := q.DequeueDatabaseSync(ctx)
			cancel()

			if err != nil {
				if err == context.DeadlineExceeded {
					continue
				}
				logger.Error("Database sync worker %d error: %v", workerID, err)
				continue
			}

			if job != nil {
				q.processDatabaseSyncJob(workerID, job)
			}
		}
	}
}

func (q *Queue) waitlistProcessingWorker(workerID int) {
	defer q.wg.Done()

	logger.Info("Waitlist processing worker %d started", workerID)

	for {
		select {
		case <-q.ctx.Done():
			logger.Info("Waitlist processing worker %d stopped", workerID)
			return
		default:

			ctx, cancel := context.WithTimeout(q.ctx, 5*time.Second)
			sectionID, err := q.DequeueWaitlistProcessing(ctx)
			cancel()

			if err != nil {
				if err == context.DeadlineExceeded {
					continue
				}
				logger.Error("Waitlist processing worker %d error: %v", workerID, err)
				continue
			}

			q.processWaitlistProcessing(workerID, sectionID)
		}
	}
}

func (q *Queue) waitlistEntryWorker(workerID int) {
	defer q.wg.Done()

	logger.Info("Waitlist entry worker %d started", workerID)

	for {
		select {
		case <-q.ctx.Done():
			logger.Info("Waitlist entry worker %d stopped", workerID)
			return
		default:

			ctx, cancel := context.WithTimeout(q.ctx, 5*time.Second)
			job, err := q.DequeueWaitlistEntry(ctx)
			cancel()

			if err != nil {
				if err == context.DeadlineExceeded {
					continue
				}
				logger.Error("Waitlist entry worker %d error: %v", workerID, err)
				continue
			}

			if job != nil {
				q.processWaitlistEntryJob(workerID, job)
			}
		}
	}
}

func (q *Queue) processDatabaseSyncJob(workerID int, job *interfaces.DatabaseSyncJob) {
	logger.Info("Worker %d processing database sync job: %s for student %s, section %s",
		workerID, job.JobType, job.StudentID, job.SectionID)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := q.registrationService.ProcessDatabaseSyncJob(ctx, *job); err != nil {
		logger.Error("Worker %d failed to process database sync job: %v", workerID, err)
	} else {
		logger.Info("Worker %d successfully processed database sync job", workerID)
	}
}

func (q *Queue) processWaitlistProcessing(workerID int, sectionID uuid.UUID) {
	logger.Info("Worker %d processing waitlist for section %s", workerID, sectionID)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := q.registrationService.ProcessWaitlist(ctx, sectionID); err != nil {
		logger.Error("Worker %d failed to process waitlist for section %s: %v", workerID, sectionID, err)

	} else {
		logger.Info("Worker %d successfully processed waitlist for section %s", workerID, sectionID)
	}
}

func (q *Queue) processWaitlistEntryJob(workerID int, job *interfaces.WaitlistJob) {
	logger.Info("Worker %d processing waitlist entry for student %s, section %s, position %d",
		workerID, job.StudentID, job.SectionID, job.Position)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := q.registrationService.ProcessWaitlistJob(ctx, *job); err != nil {
		logger.Error("Worker %d failed to process waitlist entry: %v", workerID, err)

	} else {
		logger.Info("Worker %d successfully processed waitlist entry", workerID)
	}
}

var _ interfaces.QueueService = (*Queue)(nil)
