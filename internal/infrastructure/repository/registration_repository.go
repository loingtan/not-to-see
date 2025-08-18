package repository

import (
	domain "cobra-template/internal/domain/registration"
	interfaces "cobra-template/internal/interfaces/infrastructure"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RegistrationRepository struct {
	db *gorm.DB
}

func NewRegistrationRepository(db *gorm.DB) interfaces.RegistrationRepository {
	return &RegistrationRepository{
		db: db,
	}
}

func (r *RegistrationRepository) Create(ctx context.Context, registration *domain.Registration) error {
	return r.db.WithContext(ctx).Create(registration).Error
}

func (r *RegistrationRepository) GetByStudentAndSection(ctx context.Context, studentID, sectionID uuid.UUID) (*domain.Registration, error) {
	var registration domain.Registration
	err := r.db.WithContext(ctx).
		Preload("Student").
		Preload("Section").
		Where("student_id = ? AND section_id = ?", studentID, sectionID).
		First(&registration).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &registration, nil
}

func (r *RegistrationRepository) Update(ctx context.Context, registration *domain.Registration) error {
	return r.db.WithContext(ctx).Save(registration).Error
}

func (r *RegistrationRepository) GetByStudentID(ctx context.Context, studentID uuid.UUID) ([]*domain.Registration, error) {
	var registrations []*domain.Registration
	err := r.db.WithContext(ctx).
		Preload("Student").
		Preload("Section").
		Where("student_id = ?", studentID).
		Find(&registrations).Error
	if err != nil {
		return nil, err
	}
	return registrations, nil
}

func (r *RegistrationRepository) GetBySectionID(ctx context.Context, sectionID uuid.UUID) ([]*domain.Registration, error) {
	var registrations []*domain.Registration
	err := r.db.WithContext(ctx).
		Preload("Student").
		Preload("Section").
		Where("section_id = ?", sectionID).
		Find(&registrations).Error
	if err != nil {
		return nil, err
	}
	return registrations, nil
}
