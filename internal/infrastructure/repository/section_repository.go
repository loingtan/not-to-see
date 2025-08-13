package repository

import (
	"context"
	"fmt"

	domain "cobra-template/internal/domain/registration"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SectionRepository implements SectionRepository using GORM
type SectionRepository struct {
	db *gorm.DB
}

// NewSectionRepository creates a new GORM section repository
func NewSectionRepository(db *gorm.DB) domain.SectionRepository {
	return &SectionRepository{
		db: db,
	}
}

// Create creates a new section
func (r *SectionRepository) Create(ctx context.Context, section *domain.Section) error {
	return r.db.WithContext(ctx).Create(section).Error
}

// GetByID retrieves a section by ID
func (r *SectionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Section, error) {
	var section domain.Section
	err := r.db.WithContext(ctx).Preload("Course").Preload("Semester").First(&section, "section_id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &section, nil
}

// UpdateWithOptimisticLock updates a section using optimistic locking
func (r *SectionRepository) UpdateWithOptimisticLock(ctx context.Context, section *domain.Section) error {
	result := r.db.WithContext(ctx).Model(section).
		Where("section_id = ? AND version = ?", section.SectionID, section.Version-1).
		Updates(map[string]interface{}{
			"available_seats": section.AvailableSeats,
			"version":         section.Version,
			"updated_at":      section.UpdatedAt,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("optimistic lock failure: section has been modified by another process")
	}

	return nil
}

// GetByCourseAndSemester retrieves sections by course and semester
func (r *SectionRepository) GetByCourseAndSemester(ctx context.Context, courseID, semesterID uuid.UUID) ([]*domain.Section, error) {
	var sections []*domain.Section
	err := r.db.WithContext(ctx).
		Preload("Course").
		Preload("Semester").
		Where("course_id = ? AND semester_id = ?", courseID, semesterID).
		Find(&sections).Error
	if err != nil {
		return nil, err
	}
	return sections, nil
}
