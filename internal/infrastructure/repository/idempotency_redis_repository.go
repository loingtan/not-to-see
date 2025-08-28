package repository

import (
	domain "cobra-template/internal/domain/registration"
	interfaces "cobra-template/internal/interfaces/infrastructure"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

var ErrIdempotencyKeyNotFound = errors.New("idempotency key not found")

var _ interfaces.IdempotencyRepository = (*RedisIdempotencyRepository)(nil)

type RedisIdempotencyRepository struct {
	client redis.UniversalClient
	prefix string
	ttl    time.Duration
}

func NewRedisIdempotencyRepository(client redis.UniversalClient) *RedisIdempotencyRepository {
	return &RedisIdempotencyRepository{
		client: client,
		prefix: "idempotency_key:",
		ttl:    24 * time.Hour,
	}
}

func (r *RedisIdempotencyRepository) Create(ctx context.Context, key *domain.IdempotencyKey) error {
	redisKey := r.getRedisKey(key.Key)

	data, err := json.Marshal(key)
	if err != nil {
		return fmt.Errorf("failed to marshal idempotency key: %w", err)
	}

	err = r.client.Set(ctx, redisKey, string(data), r.ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to store idempotency key in Redis: %w", err)
	}

	return nil
}

func (r *RedisIdempotencyRepository) GetByKey(ctx context.Context, key string) (*domain.IdempotencyKey, error) {
	redisKey := r.getRedisKey(key)

	val, err := r.client.Get(ctx, redisKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrIdempotencyKeyNotFound
		}
		return nil, fmt.Errorf("failed to get idempotency key from Redis: %w", err)
	}

	var idempotencyKey domain.IdempotencyKey
	err = json.Unmarshal([]byte(val), &idempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal idempotency key: %w", err)
	}

	return &idempotencyKey, nil
}

func (r *RedisIdempotencyRepository) DeleteExpired(ctx context.Context) error {

	return nil
}

func (r *RedisIdempotencyRepository) Delete(ctx context.Context, key string) error {
	redisKey := r.getRedisKey(key)

	err := r.client.Del(ctx, redisKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete idempotency key from Redis: %w", err)
	}

	return nil
}

func (r *RedisIdempotencyRepository) SetWithTTL(ctx context.Context, key *domain.IdempotencyKey, ttl time.Duration) error {
	redisKey := r.getRedisKey(key.Key)

	data, err := json.Marshal(key)
	if err != nil {
		return fmt.Errorf("failed to marshal idempotency key: %w", err)
	}

	err = r.client.Set(ctx, redisKey, string(data), ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to store idempotency key in Redis: %w", err)
	}

	return nil
}

func (r *RedisIdempotencyRepository) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	redisKey := r.getRedisKey(key)

	ttl, err := r.client.TTL(ctx, redisKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL for idempotency key: %w", err)
	}

	return ttl, nil
}

func (r *RedisIdempotencyRepository) Exists(ctx context.Context, key string) (bool, error) {
	redisKey := r.getRedisKey(key)

	exists, err := r.client.Exists(ctx, redisKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check if idempotency key exists: %w", err)
	}

	return exists > 0, nil
}

func (r *RedisIdempotencyRepository) GetKeysByStudentID(ctx context.Context, studentID uuid.UUID) ([]*domain.IdempotencyKey, error) {

	pattern := r.prefix + "*"

	var cursor uint64
	var keys []string

	for {
		var err error
		keys, cursor, err = r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan Redis keys: %w", err)
		}

		if cursor == 0 {
			break
		}
	}

	var result []*domain.IdempotencyKey

	if len(keys) > 0 {
		values, err := r.client.MGet(ctx, keys...).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to get multiple keys from Redis: %w", err)
		}

		for _, val := range values {
			if val == nil {
				continue
			}

			var idempotencyKey domain.IdempotencyKey
			err = json.Unmarshal([]byte(val.(string)), &idempotencyKey)
			if err != nil {
				continue
			}

			if idempotencyKey.StudentID == studentID {
				result = append(result, &idempotencyKey)
			}
		}
	}

	return result, nil
}

func (r *RedisIdempotencyRepository) getRedisKey(key string) string {
	return r.prefix + key
}
