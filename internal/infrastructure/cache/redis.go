package cache

import (
	interfaces "cobra-template/internal/interfaces/infrastructure"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type RedisCache struct {
	client *redis.Client
}

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

func (r *RedisCache) SetAvailableSeats(ctx context.Context, sectionID uuid.UUID, seats int, ttl time.Duration) error {
	key := fmt.Sprintf("section:seats:%s", sectionID.String())

	err := r.client.Set(ctx, key, seats, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set seats in cache: %w", err)
	}

	return nil
}

func (r *RedisCache) DecrementAvailableSeats(ctx context.Context, sectionID uuid.UUID) error {
	key := fmt.Sprintf("section:seats:%s", sectionID.String())

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

func (r *RedisCache) DecrementAndGetAvailableSeats(ctx context.Context, sectionID uuid.UUID) (int, error) {
	key := fmt.Sprintf("section:seats:%s", sectionID.String())

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

	result, err := r.client.Eval(ctx, luaScript, []string{key}).Result()
	if err != nil {
		return -1, fmt.Errorf("failed to decrement seats: %w", err)
	}

	newValue, ok := result.(int64)
	if !ok {
		return -1, fmt.Errorf("unexpected result type from Redis")
	}

	return int(newValue), nil
}

func (r *RedisCache) IncrementAndGetAvailableSeats(ctx context.Context, sectionID uuid.UUID) (int, error) {
	key := fmt.Sprintf("section:seats:%s", sectionID.String())

	result, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return -1, fmt.Errorf("failed to increment seats: %w", err)
	}

	return int(result), nil
}

func (r *RedisCache) IncrementAvailableSeats(ctx context.Context, sectionID uuid.UUID) error {
	key := fmt.Sprintf("section:seats:%s", sectionID.String())

	err := r.client.Incr(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to increment seats: %w", err)
	}

	return nil
}

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

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}

	return nil
}

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

func (r *RedisCache) Close() error {
	return r.client.Close()
}

func (r *RedisCache) Health(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

var _ interfaces.CacheService = (*RedisCache)(nil)
