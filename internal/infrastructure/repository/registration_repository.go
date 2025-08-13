package repository

import (
	"context"

	domain "cobra-template/internal/domain/registration"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RegistrationRepository implements RegistrationRepository using GORM
type RegistrationRepository struct {
	db *gorm.DB
}

// NewRegistrationRepository creates a new GORM registration repository
func NewRegistrationRepository(db *gorm.DB) domain.RegistrationRepository {
	return &RegistrationRepository{
		db: db,
	}
}

// Create creates a new registration
func (r *RegistrationRepository) Create(ctx context.Context, registration *domain.Registration) error {
	return r.db.WithContext(ctx).Create(registration).Error
}

// GetByStudentAndSection retrieves a registration by student and section
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

// Update updates an existing registration
func (r *RegistrationRepository) Update(ctx context.Context, registration *domain.Registration) error {
	return r.db.WithContext(ctx).Save(registration).Error
}

// GetByStudentID retrieves all registrations for a student
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

// GetBySectionID retrieves all registrations for a section
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
