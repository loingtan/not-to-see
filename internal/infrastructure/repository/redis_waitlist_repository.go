package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	domain "cobra-template/internal/domain/registration"
	interfaces "cobra-template/internal/interfaces/infrastructure"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type RedisWaitlistRepository struct {
	client redis.UniversalClient
}

func NewRedisWaitlistRepository(client redis.UniversalClient) interfaces.WaitlistRepository {
	return &RedisWaitlistRepository{
		client: client,
	}
}

func (r *RedisWaitlistRepository) Create(ctx context.Context, entry *domain.WaitlistEntry) error {

	waitlistKey := fmt.Sprintf("waitlist:section:%s", entry.SectionID.String())

	entryKey := fmt.Sprintf("waitlist:entry:%s", entry.WaitlistID.String())

	studentWaitlistKey := fmt.Sprintf("waitlist:student:%s", entry.StudentID.String())

	entryData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal waitlist entry: %w", err)
	}

	pipe := r.client.Pipeline()

	pipe.ZAdd(ctx, waitlistKey, &redis.Z{
		Score:  float64(entry.Position),
		Member: entry.WaitlistID.String(),
	})

	pipe.Set(ctx, entryKey, entryData, 24*time.Hour)

	pipe.SAdd(ctx, studentWaitlistKey, entry.SectionID.String())
	pipe.Expire(ctx, studentWaitlistKey, 24*time.Hour)

	studentSectionKey := fmt.Sprintf("waitlist:mapping:%s:%s", entry.StudentID.String(), entry.SectionID.String())
	pipe.Set(ctx, studentSectionKey, entry.WaitlistID.String(), 24*time.Hour)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create waitlist entry in Redis: %w", err)
	}

	return nil
}

func (r *RedisWaitlistRepository) GetByStudentAndSection(ctx context.Context, studentID, sectionID uuid.UUID) (*domain.WaitlistEntry, error) {
	studentSectionKey := fmt.Sprintf("waitlist:mapping:%s:%s", studentID.String(), sectionID.String())

	waitlistID, err := r.client.Get(ctx, studentSectionKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get waitlist mapping: %w", err)
	}

	entryKey := fmt.Sprintf("waitlist:entry:%s", waitlistID)
	entryData, err := r.client.Get(ctx, entryKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get waitlist entry: %w", err)
	}

	var entry domain.WaitlistEntry
	if err := json.Unmarshal([]byte(entryData), &entry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal waitlist entry: %w", err)
	}

	return &entry, nil
}

func (r *RedisWaitlistRepository) GetNextInLine(ctx context.Context, sectionID uuid.UUID) (*domain.WaitlistEntry, error) {
	waitlistKey := fmt.Sprintf("waitlist:section:%s", sectionID.String())

	result, err := r.client.ZRangeWithScores(ctx, waitlistKey, 0, 0).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get next in line: %w", err)
	}

	if len(result) == 0 {
		return nil, nil
	}

	waitlistID := result[0].Member.(string)
	entryKey := fmt.Sprintf("waitlist:entry:%s", waitlistID)

	entryData, err := r.client.Get(ctx, entryKey).Result()
	if err != nil {
		if err == redis.Nil {

			r.client.ZRem(ctx, waitlistKey, waitlistID)
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get waitlist entry: %w", err)
	}

	var entry domain.WaitlistEntry
	if err := json.Unmarshal([]byte(entryData), &entry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal waitlist entry: %w", err)
	}

	return &entry, nil
}

func (r *RedisWaitlistRepository) GetNextPosition(ctx context.Context, sectionID uuid.UUID) (int, error) {
	waitlistKey := fmt.Sprintf("waitlist:section:%s", sectionID.String())

	count, err := r.client.ZCard(ctx, waitlistKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get waitlist count: %w", err)
	}

	return int(count) + 1, nil
}

func (r *RedisWaitlistRepository) Delete(ctx context.Context, id uuid.UUID) error {
	entryKey := fmt.Sprintf("waitlist:entry:%s", id.String())

	entryData, err := r.client.Get(ctx, entryKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return fmt.Errorf("failed to get waitlist entry for deletion: %w", err)
	}

	var entry domain.WaitlistEntry
	if err := json.Unmarshal([]byte(entryData), &entry); err != nil {
		return fmt.Errorf("failed to unmarshal waitlist entry for deletion: %w", err)
	}

	pipe := r.client.Pipeline()

	waitlistKey := fmt.Sprintf("waitlist:section:%s", entry.SectionID.String())
	pipe.ZRem(ctx, waitlistKey, id.String())

	pipe.Del(ctx, entryKey)

	studentWaitlistKey := fmt.Sprintf("waitlist:student:%s", entry.StudentID.String())
	pipe.SRem(ctx, studentWaitlistKey, entry.SectionID.String())

	studentSectionKey := fmt.Sprintf("waitlist:mapping:%s:%s", entry.StudentID.String(), entry.SectionID.String())
	pipe.Del(ctx, studentSectionKey)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete waitlist entry from Redis: %w", err)
	}

	return nil
}

func (r *RedisWaitlistRepository) GetBySectionID(ctx context.Context, sectionID uuid.UUID) ([]*domain.WaitlistEntry, error) {
	waitlistKey := fmt.Sprintf("waitlist:section:%s", sectionID.String())

	result, err := r.client.ZRangeWithScores(ctx, waitlistKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get section waitlist: %w", err)
	}

	entries := make([]*domain.WaitlistEntry, 0, len(result))

	pipe := r.client.Pipeline()
	entryCommands := make([]*redis.StringCmd, len(result))

	for i, z := range result {
		waitlistID := z.Member.(string)
		entryKey := fmt.Sprintf("waitlist:entry:%s", waitlistID)
		entryCommands[i] = pipe.Get(ctx, entryKey)
	}

	_, err = pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get waitlist entry details: %w", err)
	}

	for i, cmd := range entryCommands {
		entryData, err := cmd.Result()
		if err != nil {
			if err == redis.Nil {

				waitlistID := result[i].Member.(string)
				r.client.ZRem(ctx, waitlistKey, waitlistID)
				continue
			}
			return nil, fmt.Errorf("failed to get waitlist entry data: %w", err)
		}

		var entry domain.WaitlistEntry
		if err := json.Unmarshal([]byte(entryData), &entry); err != nil {
			return nil, fmt.Errorf("failed to unmarshal waitlist entry: %w", err)
		}

		entries = append(entries, &entry)
	}

	return entries, nil
}

func (r *RedisWaitlistRepository) GetByStudentID(ctx context.Context, studentID uuid.UUID) ([]*domain.WaitlistEntry, error) {
	studentWaitlistKey := fmt.Sprintf("waitlist:student:%s", studentID.String())

	sectionIDs, err := r.client.SMembers(ctx, studentWaitlistKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get student waitlist sections: %w", err)
	}

	if len(sectionIDs) == 0 {
		return []*domain.WaitlistEntry{}, nil
	}

	entries := make([]*domain.WaitlistEntry, 0, len(sectionIDs))

	pipe := r.client.Pipeline()
	mappingCommands := make([]*redis.StringCmd, len(sectionIDs))

	for i, sectionID := range sectionIDs {
		studentSectionKey := fmt.Sprintf("waitlist:mapping:%s:%s", studentID.String(), sectionID)
		mappingCommands[i] = pipe.Get(ctx, studentSectionKey)
	}

	_, err = pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get waitlist mappings: %w", err)
	}

	pipe = r.client.Pipeline()
	entryCommands := make([]*redis.StringCmd, 0, len(sectionIDs))
	validIndices := make([]int, 0, len(sectionIDs))

	for i, cmd := range mappingCommands {
		waitlistID, err := cmd.Result()
		if err != nil {
			if err == redis.Nil {

				r.client.SRem(ctx, studentWaitlistKey, sectionIDs[i])
				continue
			}
			return nil, fmt.Errorf("failed to get waitlist mapping: %w", err)
		}

		entryKey := fmt.Sprintf("waitlist:entry:%s", waitlistID)
		entryCommands = append(entryCommands, pipe.Get(ctx, entryKey))
		validIndices = append(validIndices, i)
	}

	_, err = pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get waitlist entries: %w", err)
	}

	for i, cmd := range entryCommands {
		entryData, err := cmd.Result()
		if err != nil {
			if err == redis.Nil {

				sectionIdx := validIndices[i]
				r.client.SRem(ctx, studentWaitlistKey, sectionIDs[sectionIdx])
				continue
			}
			return nil, fmt.Errorf("failed to get waitlist entry data: %w", err)
		}

		var entry domain.WaitlistEntry
		if err := json.Unmarshal([]byte(entryData), &entry); err != nil {
			return nil, fmt.Errorf("failed to unmarshal waitlist entry: %w", err)
		}

		entries = append(entries, &entry)
	}

	return entries, nil
}

func (r *RedisWaitlistRepository) GetWaitlistSize(ctx context.Context, sectionID uuid.UUID) (int, error) {
	waitlistKey := fmt.Sprintf("waitlist:section:%s", sectionID.String())

	count, err := r.client.ZCard(ctx, waitlistKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get waitlist size: %w", err)
	}

	return int(count), nil
}

func (r *RedisWaitlistRepository) GetWaitlistPosition(ctx context.Context, studentID, sectionID uuid.UUID) (int, error) {
	studentSectionKey := fmt.Sprintf("waitlist:mapping:%s:%s", studentID.String(), sectionID.String())

	waitlistID, err := r.client.Get(ctx, studentSectionKey).Result()
	if err != nil {
		if err == redis.Nil {
			return -1, nil
		}
		return -1, fmt.Errorf("failed to get waitlist mapping: %w", err)
	}

	waitlistKey := fmt.Sprintf("waitlist:section:%s", sectionID.String())
	rank, err := r.client.ZRank(ctx, waitlistKey, waitlistID).Result()
	if err != nil {
		if err == redis.Nil {
			return -1, nil
		}
		return -1, fmt.Errorf("failed to get waitlist rank: %w", err)
	}

	return int(rank) + 1, nil
}

func (r *RedisWaitlistRepository) CleanupExpiredEntries(ctx context.Context) error {

	return nil
}
