package cache

import (
	"cobra-template/internal/config"
	interfaces "cobra-template/internal/interfaces/infrastructure"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type RedisCache struct {
	client redis.UniversalClient
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

// NewRedisCacheWithConfig creates a new Redis cache instance using configuration
func NewRedisCacheWithConfig(cfg *config.CacheConfig) *RedisCache {
	

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
		// Check if the error is due to key not existing
		if strings.Contains(err.Error(), "Key does not exist") {
			return fmt.Errorf("seat key not found for section %s", sectionID.String())
		}
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
		// Check if the error is due to key not existing
		if strings.Contains(err.Error(), "Key does not exist") {
			return -1, fmt.Errorf("seat key not found for section %s", sectionID.String())
		}
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

	rm := json.RawMessage([]byte(val))
	return rm, nil
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

	rm := json.RawMessage([]byte(val))
	return rm, nil
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

func (r *RedisCache) GetStudentDetails(ctx context.Context, studentID uuid.UUID) (interface{}, error) {
	key := fmt.Sprintf("student:details:%s", studentID.String())

	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("student details not cached")
		}
		return nil, fmt.Errorf("failed to get student details: %w", err)
	}

	rm := json.RawMessage([]byte(val))
	return rm, nil
}

func (r *RedisCache) SetStudentDetails(ctx context.Context, studentID uuid.UUID, data interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("student:details:%s", studentID.String())

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal student details: %w", err)
	}

	err = r.client.Set(ctx, key, jsonData, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set student details: %w", err)
	}

	return nil
}

func (r *RedisCache) GetStudentRegistrations(ctx context.Context, studentID uuid.UUID) (interface{}, error) {
	key := fmt.Sprintf("student:registrations:%s", studentID.String())

	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("student registrations not cached")
		}
		return nil, fmt.Errorf("failed to get student registrations: %w", err)
	}

	rm := json.RawMessage([]byte(val))
	return rm, nil
}

func (r *RedisCache) SetStudentRegistrations(ctx context.Context, studentID uuid.UUID, data interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("student:registrations:%s", studentID.String())

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal student registrations: %w", err)
	}

	err = r.client.Set(ctx, key, jsonData, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set student registrations: %w", err)
	}

	return nil
}

func (r *RedisCache) GetStudentWaitlistStatus(ctx context.Context, studentID uuid.UUID) (interface{}, error) {
	key := fmt.Sprintf("student:waitlist:%s", studentID.String())

	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("student waitlist status not cached")
		}
		return nil, fmt.Errorf("failed to get student waitlist status: %w", err)
	}

	rm := json.RawMessage([]byte(val))
	return rm, nil
}

func (r *RedisCache) SetStudentWaitlistStatus(ctx context.Context, studentID uuid.UUID, data interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("student:waitlist:%s", studentID.String())

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal student waitlist: %w", err)
	}

	err = r.client.Set(ctx, key, jsonData, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set student waitlist: %w", err)
	}

	return nil
}

func (r *RedisCache) GetAvailableSections(ctx context.Context, semesterID uuid.UUID) (interface{}, error) {
	key := fmt.Sprintf("sections:available:%s", semesterID.String())

	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("available sections not cached")
		}
		return nil, fmt.Errorf("failed to get available sections: %w", err)
	}

	rm := json.RawMessage([]byte(val))
	return rm, nil
}

func (r *RedisCache) SetAvailableSections(ctx context.Context, semesterID uuid.UUID, data interface{}, ttl time.Duration) error {

	key := fmt.Sprintf("sections:available:%s", semesterID.String())

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal available sections: %w", err)
	}

	err = r.client.Set(ctx, key, jsonData, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set available sections: %w", err)
	}

	return nil
}

// Cache invalidation methods
func (r *RedisCache) InvalidateStudentCache(ctx context.Context, studentID uuid.UUID) error {
	pattern := fmt.Sprintf("student:*:%s", studentID.String())
	return r.Clear(ctx, pattern)
}

func (r *RedisCache) InvalidateSectionCache(ctx context.Context, sectionID uuid.UUID) error {
	// Clear section-specific cache
	sectionPattern := fmt.Sprintf("section:*:%s", sectionID.String())
	if err := r.Clear(ctx, sectionPattern); err != nil {
		return err
	}

	// Clear available sections cache (since it includes this section)
	availableSectionsPattern := "sections:available:*"
	return r.Clear(ctx, availableSectionsPattern)
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

// Generic cache operations for HTTP responses and other data
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key not found")
		}
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}
	return val, nil
}

func (r *RedisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	err := r.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}
	return nil
}

func (r *RedisCache) GetWithMetadata(ctx context.Context, key string) (string, map[string]string, error) {
	// Use Redis HMGET to get both value and metadata
	dataKey := key + ":data"
	metaKey := key + ":meta"

	pipe := r.client.Pipeline()
	dataCmd := pipe.Get(ctx, dataKey)
	metaCmd := pipe.HGetAll(ctx, metaKey)

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return "", nil, fmt.Errorf("failed to get key with metadata %s: %w", key, err)
	}

	data, err := dataCmd.Result()
	if err == redis.Nil {
		return "", nil, fmt.Errorf("key not found")
	} else if err != nil {
		return "", nil, fmt.Errorf("failed to get data for key %s: %w", key, err)
	}

	metadata, err := metaCmd.Result()
	if err != nil && err != redis.Nil {
		return "", nil, fmt.Errorf("failed to get metadata for key %s: %w", key, err)
	}

	return data, metadata, nil
}

func (r *RedisCache) SetWithMetadata(ctx context.Context, key string, value string, metadata map[string]string, ttl time.Duration) error {
	dataKey := key + ":data"
	metaKey := key + ":meta"

	pipe := r.client.Pipeline()
	pipe.Set(ctx, dataKey, value, ttl)

	if len(metadata) > 0 {
		pipe.HMSet(ctx, metaKey, metadata)
		pipe.Expire(ctx, metaKey, ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to set key with metadata %s: %w", key, err)
	}

	return nil
}

// Waitlist management using Redis sorted sets
func (r *RedisCache) AddToWaitlist(ctx context.Context, sectionID, studentID uuid.UUID, position int, entry interface{}) error {
	waitlistKey := fmt.Sprintf("waitlist:section:%s", sectionID.String())
	entryKey := fmt.Sprintf("waitlist:entry:%s:%s", sectionID.String(), studentID.String())
	studentWaitlistKey := fmt.Sprintf("waitlist:student:%s", studentID.String())

	// Serialize entry data
	entryData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal waitlist entry: %w", err)
	}

	// Use pipeline for atomic operations
	pipe := r.client.Pipeline()

	// Add to section waitlist sorted set (score = position for ordering)
	pipe.ZAdd(ctx, waitlistKey, &redis.Z{
		Score:  float64(position),
		Member: studentID.String(),
	})

	// Store detailed entry information
	pipe.Set(ctx, entryKey, entryData, 24*time.Hour)

	// Add to student's waitlist set
	pipe.SAdd(ctx, studentWaitlistKey, sectionID.String())
	pipe.Expire(ctx, studentWaitlistKey, 24*time.Hour)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to add to waitlist: %w", err)
	}

	return nil
}

func (r *RedisCache) RemoveFromWaitlist(ctx context.Context, sectionID, studentID uuid.UUID) error {
	waitlistKey := fmt.Sprintf("waitlist:section:%s", sectionID.String())
	entryKey := fmt.Sprintf("waitlist:entry:%s:%s", sectionID.String(), studentID.String())
	studentWaitlistKey := fmt.Sprintf("waitlist:student:%s", studentID.String())

	// Use pipeline for atomic operations
	pipe := r.client.Pipeline()

	// Remove from section waitlist sorted set
	pipe.ZRem(ctx, waitlistKey, studentID.String())

	// Remove detailed entry
	pipe.Del(ctx, entryKey)

	// Remove from student's waitlist set
	pipe.SRem(ctx, studentWaitlistKey, sectionID.String())

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to remove from waitlist: %w", err)
	}

	return nil
}

func (r *RedisCache) GetNextInWaitlist(ctx context.Context, sectionID uuid.UUID) (interface{}, error) {
	waitlistKey := fmt.Sprintf("waitlist:section:%s", sectionID.String())

	// Get the member with the lowest score (first in line)
	result, err := r.client.ZRangeWithScores(ctx, waitlistKey, 0, 0).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get next in waitlist: %w", err)
	}

	if len(result) == 0 {
		return nil, nil // No one in waitlist
	}

	studentID := result[0].Member.(string)
	entryKey := fmt.Sprintf("waitlist:entry:%s:%s", sectionID.String(), studentID)

	entryData, err := r.client.Get(ctx, entryKey).Result()
	if err != nil {
		if err == redis.Nil {
			// Entry expired, clean up the sorted set
			r.client.ZRem(ctx, waitlistKey, studentID)
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get waitlist entry: %w", err)
	}

	var entry interface{}
	if err := json.Unmarshal([]byte(entryData), &entry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal waitlist entry: %w", err)
	}

	return entry, nil
}

func (r *RedisCache) GetWaitlistPosition(ctx context.Context, sectionID, studentID uuid.UUID) (int, error) {
	waitlistKey := fmt.Sprintf("waitlist:section:%s", sectionID.String())

	rank, err := r.client.ZRank(ctx, waitlistKey, studentID.String()).Result()
	if err != nil {
		if err == redis.Nil {
			return -1, nil // Not in waitlist
		}
		return -1, fmt.Errorf("failed to get waitlist rank: %w", err)
	}

	return int(rank) + 1, nil // Convert 0-based rank to 1-based position
}

func (r *RedisCache) GetWaitlistSize(ctx context.Context, sectionID uuid.UUID) (int, error) {
	waitlistKey := fmt.Sprintf("waitlist:section:%s", sectionID.String())

	count, err := r.client.ZCard(ctx, waitlistKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get waitlist size: %w", err)
	}

	return int(count), nil
}

func (r *RedisCache) GetStudentWaitlists(ctx context.Context, studentID uuid.UUID) ([]interface{}, error) {
	studentWaitlistKey := fmt.Sprintf("waitlist:student:%s", studentID.String())

	// Get all section IDs this student is waitlisted for
	sectionIDs, err := r.client.SMembers(ctx, studentWaitlistKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get student waitlist sections: %w", err)
	}

	if len(sectionIDs) == 0 {
		return []interface{}{}, nil
	}

	waitlists := make([]interface{}, 0, len(sectionIDs))

	// Get waitlist entry for each section
	pipe := r.client.Pipeline()
	entryCommands := make([]*redis.StringCmd, len(sectionIDs))

	for i, sectionID := range sectionIDs {
		entryKey := fmt.Sprintf("waitlist:entry:%s:%s", sectionID, studentID.String())
		entryCommands[i] = pipe.Get(ctx, entryKey)
	}

	_, err = pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get waitlist entries: %w", err)
	}

	for i, cmd := range entryCommands {
		entryData, err := cmd.Result()
		if err != nil {
			if err == redis.Nil {
				// Entry expired, clean up student waitlist set
				r.client.SRem(ctx, studentWaitlistKey, sectionIDs[i])
				continue
			}
			return nil, fmt.Errorf("failed to get waitlist entry data: %w", err)
		}

		var entry interface{}
		if err := json.Unmarshal([]byte(entryData), &entry); err != nil {
			return nil, fmt.Errorf("failed to unmarshal waitlist entry: %w", err)
		}

		waitlists = append(waitlists, entry)
	}

	return waitlists, nil
}

// GetCacheStats returns cache statistics
func (r *RedisCache) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	info, err := r.client.Info(ctx, "stats").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache stats: %w", err)
	}

	memory, err := r.client.Info(ctx, "memory").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory stats: %w", err)
	}

	// Parse key statistics
	dbSize, err := r.client.DBSize(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get db size: %w", err)
	}

	stats := map[string]interface{}{
		"total_keys":       dbSize,
		"stats_info":       info,
		"memory_info":      memory,
		"connection_count": r.client.PoolStats().TotalConns,
		"hit_rate":         "calculated_by_middleware", // Will be calculated by middleware
	}

	return stats, nil
}

var _ interfaces.CacheService = (*RedisCache)(nil)

// GetClient returns the underlying Redis client for advanced operations
func (r *RedisCache) GetClient() redis.UniversalClient {
	return r.client
}
