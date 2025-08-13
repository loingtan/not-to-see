package repository

import (
	"context"
	"time"

	domain "cobra-template/internal/domain/registration"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SemesterRepository implements SemesterRepository using GORM
type SemesterRepository struct {
	db *gorm.DB
}

// NewSemesterRepository creates a new GORM semester repository
func NewSemesterRepository(db *gorm.DB) domain.SemesterRepository {
	return &SemesterRepository{
		db: db,
	}
}

// Create creates a new semester
func (r *SemesterRepository) Create(ctx context.Context, semester *domain.Semester) error {
	return r.db.WithContext(ctx).Create(semester).Error
}

// GetByID retrieves a semester by ID
func (r *SemesterRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Semester, error) {
	var semester domain.Semester
	err := r.db.WithContext(ctx).First(&semester, "semester_id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &semester, nil
}

// GetCurrent retrieves the current active semester
func (r *SemesterRepository) GetCurrent(ctx context.Context) (*domain.Semester, error) {
	var semester domain.Semester
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("start_date <= ? AND end_date >= ?", now, now).
		First(&semester).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &semester, nil
}
