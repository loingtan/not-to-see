package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"cobra-template/internal/service"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// RedisCache implements the CacheService interface using Redis
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(addr, password string, db int) *RedisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisCache{
		client: rdb,
	}
}

// GetAvailableSeats gets the available seats for a section from cache
func (r *RedisCache) GetAvailableSeats(ctx context.Context, sectionID uuid.UUID) (int, error) {
	key := fmt.Sprintf("section:seats:%s", sectionID.String())
	
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return -1, fmt.Errorf("section seats not cached")
		}
		return -1, fmt.Errorf("failed to get seats from cache: %w", err)
	}

	seats, err := strconv.Atoi(val)
	if err != nil {
		return -1, fmt.Errorf("invalid seats value in cache: %w", err)
	}

	return seats, nil
}

// SetAvailableSeats sets the available seats for a section in cache
func (r *RedisCache) SetAvailableSeats(ctx context.Context, sectionID uuid.UUID, seats int, ttl time.Duration) error {
	key := fmt.Sprintf("section:seats:%s", sectionID.String())
	
	err := r.client.Set(ctx, key, seats, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set seats in cache: %w", err)
	}

	return nil
}

// DecrementAvailableSeats atomically decrements available seats
func (r *RedisCache) DecrementAvailableSeats(ctx context.Context, sectionID uuid.UUID) error {
	key := fmt.Sprintf("section:seats:%s", sectionID.String())
	
	// Use Lua script for atomic decrement with bounds checking
	luaScript := `
		local key = KEYS[1]
		local current = redis.call("GET", key)
		if current == false then
			return redis.error_reply("Key does not exist")
		end
		local value = tonumber(current)
		if value <= 0 then
			return redis.error_reply("No seats available")
		end
		return redis.call("DECR", key)
	`
	
	err := r.client.Eval(ctx, luaScript, []string{key}).Err()
	if err != nil {
		return fmt.Errorf("failed to decrement seats: %w", err)
	}

	return nil
}

// IncrementAvailableSeats atomically increments available seats
func (r *RedisCache) IncrementAvailableSeats(ctx context.Context, sectionID uuid.UUID) error {
	key := fmt.Sprintf("section:seats:%s", sectionID.String())
	
	err := r.client.Incr(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to increment seats: %w", err)
	}

	return nil
}

// GetSectionDetails gets section details from cache
func (r *RedisCache) GetSectionDetails(ctx context.Context, sectionID uuid.UUID) (interface{}, error) {
	key := fmt.Sprintf("section:details:%s", sectionID.String())
	
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("section details not cached")
		}
		return nil, fmt.Errorf("failed to get section details: %w", err)
	}

	var details interface{}
	if err := json.Unmarshal([]byte(val), &details); err != nil {
		return nil, fmt.Errorf("failed to unmarshal section details: %w", err)
	}

	return details, nil
}

// SetSectionDetails sets section details in cache
func (r *RedisCache) SetSectionDetails(ctx context.Context, sectionID uuid.UUID, data interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("section:details:%s", sectionID.String())
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal section details: %w", err)
	}

	err = r.client.Set(ctx, key, jsonData, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set section details: %w", err)
	}

	return nil
}

// GetCourseDetails gets course details from cache
func (r *RedisCache) GetCourseDetails(ctx context.Context, courseID uuid.UUID) (interface{}, error) {
	key := fmt.Sprintf("course:details:%s", courseID.String())
	
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("course details not cached")
		}
		return nil, fmt.Errorf("failed to get course details: %w", err)
	}

	var details interface{}
	if err := json.Unmarshal([]byte(val), &details); err != nil {
		return nil, fmt.Errorf("failed to unmarshal course details: %w", err)
	}

	return details, nil
}

// SetCourseDetails sets course details in cache
func (r *RedisCache) SetCourseDetails(ctx context.Context, courseID uuid.UUID, data interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("course:details:%s", courseID.String())
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal course details: %w", err)
	}

	err = r.client.Set(ctx, key, jsonData, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set course details: %w", err)
	}

	return nil
}

// Delete removes a key from cache
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}

	return nil
}

// Clear removes all keys matching a pattern
func (r *RedisCache) Clear(ctx context.Context, pattern string) error {
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys for pattern %s: %w", pattern, err)
	}

	if len(keys) > 0 {
		err = r.client.Del(ctx, keys...).Err()
		if err != nil {
			return fmt.Errorf("failed to delete keys: %w", err)
		}
	}

	return nil
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// Health checks Redis connectivity
func (r *RedisCache) Health(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Compile-time check to ensure RedisCache implements CacheService
var _ service.CacheService = (*RedisCache)(nil)
