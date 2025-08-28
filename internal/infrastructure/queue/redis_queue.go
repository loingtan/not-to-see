package queue

import (
	"cobra-template/internal/config"
	interfaces "cobra-template/internal/interfaces/infrastructure"
	serviceInterfaces "cobra-template/internal/interfaces/service"
	"cobra-template/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

const (
	DatabaseSyncQueueKey  = "queue:database_sync"
	WaitlistQueueKey      = "queue:waitlist"
	WaitlistEntryQueueKey = "queue:waitlist_entry"
	DefaultDequeueTimeout = 2 * time.Second // Reasonable timeout for polling
	DefaultJobTimeout     = 30 * time.Second
	WorkerSleepDuration   = 50 * time.Millisecond // Sleep when no work available
)

type RedisQueue struct {
	client redis.UniversalClient

	workers int
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	started bool
	mu      sync.RWMutex

	registrationService serviceInterfaces.RegistrationService
}

// NewRedisQueue creates a new Redis-based queue service
func NewRedisQueue(cfg *config.CacheConfig, workers int) interfaces.QueueService {
	ctx, cancel := context.WithCancel(context.Background())

	rdb := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:       cfg.Sentinel.MasterName,
		SentinelAddrs:    cfg.Sentinel.SentinelAddrs,
		SentinelPassword: cfg.Sentinel.SentinelPassword,
		Password:         cfg.Password,
		DB:               cfg.DB,
		MaxRetries:       cfg.MaxRetries,
		PoolSize:         cfg.PoolSize,
		PoolTimeout:      time.Duration(cfg.PoolTimeout) * time.Second,
		IdleTimeout:      time.Duration(cfg.IdleTimeout) * time.Second,
	})

	queue := &RedisQueue{
		client:  rdb,
		workers: workers,
		ctx:     ctx,
		cancel:  cancel,
		started: false,
	}

	return queue
}

func (rq *RedisQueue) SetRegistrationService(service interface{}) {
	rq.mu.Lock()
	defer rq.mu.Unlock()

	if regService, ok := service.(serviceInterfaces.RegistrationService); ok {
		rq.registrationService = regService
	} else {
		logger.Error("Invalid service type provided to SetRegistrationService")
	}
}

func (rq *RedisQueue) StartWorkers() {
	rq.mu.Lock()
	defer rq.mu.Unlock()

	if rq.started {
		return
	}

	if rq.registrationService == nil {
		logger.Warn("Registration service not set, workers cannot process jobs")
		return
	}

	logger.Info("Starting %d Redis queue workers", rq.workers)

	// Start database sync workers
	for i := 0; i < rq.workers; i++ {
		rq.wg.Add(1)
		go rq.databaseSyncWorker(i)
	}

	// Start waitlist processing workers
	for i := 0; i < rq.workers; i++ {
		rq.wg.Add(1)
		go rq.waitlistProcessingWorker(i)
	}

	// Start waitlist entry workers
	for i := 0; i < rq.workers; i++ {
		rq.wg.Add(1)
		go rq.waitlistEntryWorker(i)
	}

	rq.started = true
	logger.Info("Redis queue workers started successfully")
}

func (rq *RedisQueue) StopWorkers() {
	rq.mu.Lock()
	defer rq.mu.Unlock()

	if !rq.started {
		return
	}

	logger.Info("Stopping Redis queue workers...")
	rq.cancel()
	rq.wg.Wait()
	rq.started = false
	logger.Info("Redis queue workers stopped")
}

// EnqueueDatabaseSync adds a database sync job to the Redis queue
func (rq *RedisQueue) EnqueueDatabaseSync(ctx context.Context, job interfaces.DatabaseSyncJob) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal database sync job: %w", err)
	}

	err = rq.client.LPush(ctx, DatabaseSyncQueueKey, data).Err()
	if err != nil {
		return fmt.Errorf("failed to enqueue database sync job: %w", err)
	}

	logger.Debug("Enqueued database sync job: %s for student %s, section %s",
		job.JobType, job.StudentID, job.SectionID)
	return nil
}

// DequeueDatabaseSync retrieves a database sync job from the Redis queue
func (rq *RedisQueue) DequeueDatabaseSync(ctx context.Context) (*interfaces.DatabaseSyncJob, error) {
	result, err := rq.client.BRPop(ctx, DefaultDequeueTimeout, DatabaseSyncQueueKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No items available, return nil job
		}
		if err == context.DeadlineExceeded {
			return nil, nil // Timeout is expected when no jobs, return nil job
		}
		return nil, fmt.Errorf("failed to dequeue database sync job: %w", err)
	}

	if len(result) != 2 {
		return nil, fmt.Errorf("unexpected Redis BRPOP result format")
	}

	var job interfaces.DatabaseSyncJob
	err = json.Unmarshal([]byte(result[1]), &job)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal database sync job: %w", err)
	}

	return &job, nil
}

// EnqueueWaitlistProcessing adds a section ID for waitlist processing to the Redis queue
func (rq *RedisQueue) EnqueueWaitlistProcessing(ctx context.Context, sectionID uuid.UUID) error {
	err := rq.client.LPush(ctx, WaitlistQueueKey, sectionID.String()).Err()
	if err != nil {
		return fmt.Errorf("failed to enqueue waitlist processing for section %s: %w", sectionID, err)
	}

	logger.Debug("Enqueued waitlist processing for section: %s", sectionID)
	return nil
}

// DequeueWaitlistProcessing retrieves a section ID for waitlist processing from the Redis queue
func (rq *RedisQueue) DequeueWaitlistProcessing(ctx context.Context) (uuid.UUID, error) {
	result, err := rq.client.BRPop(ctx, DefaultDequeueTimeout, WaitlistQueueKey).Result()
	if err != nil {
		if err == redis.Nil {
			return uuid.UUID{}, nil // No items available, return empty UUID
		}
		if err == context.DeadlineExceeded {
			return uuid.UUID{}, nil // Timeout is expected when no jobs, return empty UUID
		}
		return uuid.UUID{}, fmt.Errorf("failed to dequeue waitlist processing: %w", err)
	}

	if len(result) != 2 {
		return uuid.UUID{}, fmt.Errorf("unexpected Redis BRPOP result format")
	}

	sectionID, err := uuid.Parse(result[1])
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to parse section ID: %w", err)
	}

	return sectionID, nil
}

// EnqueueWaitlistEntry adds a waitlist entry job to the Redis queue
func (rq *RedisQueue) EnqueueWaitlistEntry(ctx context.Context, job interfaces.WaitlistJob) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal waitlist entry job: %w", err)
	}

	err = rq.client.LPush(ctx, WaitlistEntryQueueKey, data).Err()
	if err != nil {
		return fmt.Errorf("failed to enqueue waitlist entry job: %w", err)
	}

	logger.Debug("Enqueued waitlist entry job for student %s, section %s, position %d",
		job.StudentID, job.SectionID, job.Position)
	return nil
}

// DequeueWaitlistEntry retrieves a waitlist entry job from the Redis queue
func (rq *RedisQueue) DequeueWaitlistEntry(ctx context.Context) (*interfaces.WaitlistJob, error) {
	result, err := rq.client.BRPop(ctx, DefaultDequeueTimeout, WaitlistEntryQueueKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No items available, return nil job
		}
		if err == context.DeadlineExceeded {
			return nil, nil // Timeout is expected when no jobs, return nil job
		}
		return nil, fmt.Errorf("failed to dequeue waitlist entry job: %w", err)
	}

	if len(result) != 2 {
		return nil, fmt.Errorf("unexpected Redis BRPOP result format")
	}

	var job interfaces.WaitlistJob
	err = json.Unmarshal([]byte(result[1]), &job)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal waitlist entry job: %w", err)
	}

	return &job, nil
}

// Worker methods
func (rq *RedisQueue) databaseSyncWorker(workerID int) {
	defer rq.wg.Done()

	logger.Info("Redis database sync worker %d started", workerID)

	for {
		select {
		case <-rq.ctx.Done():
			logger.Info("Redis database sync worker %d stopped", workerID)
			return
		default:
			// Create a timeout context for each dequeue operation
			ctx, cancel := context.WithTimeout(context.Background(), DefaultDequeueTimeout)
			job, err := rq.DequeueDatabaseSync(ctx)
			cancel()

			if err != nil {
				logger.Error("Redis database sync worker %d error: failed to dequeue database sync job: %v", workerID, err)
				time.Sleep(WorkerSleepDuration) // Brief pause on error
				continue
			}

			if job != nil {
				rq.processDatabaseSyncJob(workerID, job)
			} else {
				// No jobs available, sleep briefly to avoid busy polling
				time.Sleep(WorkerSleepDuration)
			}
		}
	}
}

func (rq *RedisQueue) waitlistProcessingWorker(workerID int) {
	defer rq.wg.Done()

	logger.Info("Redis waitlist processing worker %d started", workerID)

	for {
		select {
		case <-rq.ctx.Done():
			logger.Info("Redis waitlist processing worker %d stopped", workerID)
			return
		default:
			// Create a timeout context for each dequeue operation
			ctx, cancel := context.WithTimeout(context.Background(), DefaultDequeueTimeout)
			sectionID, err := rq.DequeueWaitlistProcessing(ctx)
			cancel()

			if err != nil {
				logger.Error("Redis waitlist processing worker %d error: failed to dequeue waitlist processing: %v", workerID, err)
				time.Sleep(WorkerSleepDuration) // Brief pause on error
				continue
			}

			// Check if we got a valid UUID (not empty)
			if sectionID != (uuid.UUID{}) {
				rq.processWaitlistProcessing(workerID, sectionID)
			} else {
				// No jobs available, sleep briefly to avoid busy polling
				time.Sleep(WorkerSleepDuration)
			}
		}
	}
}

func (rq *RedisQueue) waitlistEntryWorker(workerID int) {
	defer rq.wg.Done()

	logger.Info("Redis waitlist entry worker %d started", workerID)

	for {
		select {
		case <-rq.ctx.Done():
			logger.Info("Redis waitlist entry worker %d stopped", workerID)
			return
		default:
			// Create a timeout context for each dequeue operation
			ctx, cancel := context.WithTimeout(context.Background(), DefaultDequeueTimeout)
			job, err := rq.DequeueWaitlistEntry(ctx)
			cancel()

			if err != nil {
				logger.Error("Redis waitlist entry worker %d error: failed to dequeue waitlist entry job: %v", workerID, err)
				time.Sleep(WorkerSleepDuration) // Brief pause on error
				continue
			}

			if job != nil {
				rq.processWaitlistEntryJob(workerID, job)
			} else {
				// No jobs available, sleep briefly to avoid busy polling
				time.Sleep(WorkerSleepDuration)
			}
		}
	}
}

// Job processing methods
func (rq *RedisQueue) processDatabaseSyncJob(workerID int, job *interfaces.DatabaseSyncJob) {
	logger.Info("Redis worker %d processing database sync job: %s for student %s, section %s",
		workerID, job.JobType, job.StudentID, job.SectionID)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultJobTimeout)
	defer cancel()

	if err := rq.registrationService.ProcessDatabaseSyncJob(ctx, *job); err != nil {
		logger.Error("Redis worker %d failed to process database sync job: %v", workerID, err)
	} else {
		logger.Info("Redis worker %d successfully processed database sync job", workerID)
	}
}

func (rq *RedisQueue) processWaitlistProcessing(workerID int, sectionID uuid.UUID) {
	logger.Info("Redis worker %d processing waitlist for section %s", workerID, sectionID)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultJobTimeout)
	defer cancel()

	if err := rq.registrationService.ProcessWaitlist(ctx, sectionID); err != nil {
		logger.Error("Redis worker %d failed to process waitlist for section %s: %v", workerID, sectionID, err)
	} else {
		logger.Info("Redis worker %d successfully processed waitlist for section %s", workerID, sectionID)
	}
}

func (rq *RedisQueue) processWaitlistEntryJob(workerID int, job *interfaces.WaitlistJob) {
	logger.Info("Redis worker %d processing waitlist entry for student %s, section %s, position %d",
		workerID, job.StudentID, job.SectionID, job.Position)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultJobTimeout)
	defer cancel()

	if err := rq.registrationService.ProcessWaitlistJob(ctx, *job); err != nil {
		logger.Error("Redis worker %d failed to process waitlist entry: %v", workerID, err)
	} else {
		logger.Info("Redis worker %d successfully processed waitlist entry", workerID)
	}
}

// Ensure RedisQueue implements QueueService interface
var _ interfaces.QueueService = (*RedisQueue)(nil)
