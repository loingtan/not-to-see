package repository

import (
	domain "cobra-template/internal/domain/registration"
	interfaces "cobra-template/internal/interfaces/infrastructure"
	"context"
	"time"

	"gorm.io/gorm"
)

var _ interfaces.IdempotencyRepository = (*IdempotencyRepository)(nil)

type IdempotencyRepository struct {
	db *gorm.DB
}

func NewIdempotencyRepository(db *gorm.DB) *IdempotencyRepository {
	return &IdempotencyRepository{db: db}
}

func (r *IdempotencyRepository) Create(ctx context.Context, key *domain.IdempotencyKey) error {
	return r.db.WithContext(ctx).Create(key).Error
}

func (r *IdempotencyRepository) GetByKey(ctx context.Context, key string) (*domain.IdempotencyKey, error) {
	var idempotencyKey domain.IdempotencyKey
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&idempotencyKey).Error
	if err != nil {
		return nil, err
	}
	return &idempotencyKey, nil
}

func (r *IdempotencyRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&domain.IdempotencyKey{}).Error
}

func (r *IdempotencyRepository) Delete(ctx context.Context, key string) error {
	return r.db.WithContext(ctx).Where("key = ?", key).Delete(&domain.IdempotencyKey{}).Error
}
