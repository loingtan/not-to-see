package service

import (
	domain "cobra-template/internal/domain/registration"
	interfaces "cobra-template/internal/interfaces/infrastructure"
	"cobra-template/pkg/logger"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	DefaultIdempotencyTTL = 24 * time.Hour
)

type IdempotencyService struct {
	idempotencyRepo interfaces.IdempotencyRepository
}

func NewIdempotencyService(idempotencyRepo interfaces.IdempotencyRepository) *IdempotencyService {
	return &IdempotencyService{
		idempotencyRepo: idempotencyRepo,
	}
}

func (s *IdempotencyService) CheckDuplicateRequest(ctx context.Context, key string, studentID uuid.UUID, requestData any) (*domain.IdempotencyKey, bool, error) {
	if key == "" {

		return nil, false, nil
	}

	existingKey, err := s.idempotencyRepo.GetByKey(ctx, key)
	if err != nil && err != gorm.ErrRecordNotFound {
		logger.Error("Failed to check idempotency key: %v", err)
		return nil, false, fmt.Errorf("failed to check idempotency key: %w", err)
	}

	if existingKey != nil {

		if existingKey.IsExpired() {

			if err := s.idempotencyRepo.Delete(ctx, key); err != nil {
				logger.Warn("Failed to delete expired idempotency key %s: %v", key, err)
			}
			return nil, false, nil
		}

		requestHash := s.generateRequestHash(studentID, requestData)
		if existingKey.RequestHash == requestHash {

			logger.Info("Duplicate request detected for idempotency key: %s", key)
			return existingKey, true, nil
		} else {

			logger.Warn("Idempotency key %s used with different request data", key)
			return nil, false, fmt.Errorf("idempotency key already used with different request data")
		}
	}

	return nil, false, nil
}

func (s *IdempotencyService) StoreProcessedRequest(ctx context.Context, key string, studentID uuid.UUID, requestData any, responseData any, statusCode int) error {
	if key == "" {

		return nil
	}

	requestHash := s.generateRequestHash(studentID, requestData)

	responseJSON, err := json.Marshal(responseData)
	if err != nil {
		logger.Error("Failed to marshal response data for idempotency key %s: %v", key, err)
		return fmt.Errorf("failed to marshal response data: %w", err)
	}

	idempotencyKey := &domain.IdempotencyKey{
		Key:          key,
		StudentID:    studentID,
		RequestHash:  requestHash,
		ResponseData: string(responseJSON),
		StatusCode:   statusCode,
		ProcessedAt:  time.Now(),
		ExpiresAt:    time.Now().Add(DefaultIdempotencyTTL),
		CreatedAt:    time.Now(),
	}

	if err := s.idempotencyRepo.Create(ctx, idempotencyKey); err != nil {
		logger.Error("Failed to store idempotency key %s: %v", key, err)
		return fmt.Errorf("failed to store idempotency key: %w", err)
	}

	logger.Info("Stored idempotency key: %s", key)
	return nil
}

func (s *IdempotencyService) CleanupExpiredKeys(ctx context.Context) error {
	if err := s.idempotencyRepo.DeleteExpired(ctx); err != nil {
		logger.Error("Failed to cleanup expired idempotency keys: %v", err)
		return fmt.Errorf("failed to cleanup expired keys: %w", err)
	}
	return nil
}

func (s *IdempotencyService) generateRequestHash(studentID uuid.UUID, requestData any) string {
	data := map[string]any{
		"student_id":   studentID.String(),
		"request_data": requestData,
	}

	jsonData, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])
}
