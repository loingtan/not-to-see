package repository

import (
	"context"
	"time"

	domain "cobra-template/internal/domain/registration"
	interfaces "cobra-template/internal/interfaces/infrastructure"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SemesterRepository struct {
	db *gorm.DB
}

func NewSemesterRepository(db *gorm.DB) interfaces.SemesterRepository {
	return &SemesterRepository{
		db: db,
	}
}

func (r *SemesterRepository) Create(ctx context.Context, semester *domain.Semester) error {
	return r.db.WithContext(ctx).Create(semester).Error
}

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

func (r *SemesterRepository) GetAllActive(ctx context.Context) ([]*domain.Semester, error) {
	var semesters []*domain.Semester
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Find(&semesters).Error
	if err != nil {
		return nil, err
	}
	return semesters, nil
}
