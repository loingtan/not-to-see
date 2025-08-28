package repository

import (
	"context"
	"fmt"

	domain "cobra-template/internal/domain/registration"
	interfaces "cobra-template/internal/interfaces/infrastructure"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SectionRepository struct {
	db *gorm.DB
}

func NewSectionRepository(db *gorm.DB) interfaces.SectionRepository {
	return &SectionRepository{
		db: db,
	}
}
func (r *SectionRepository) Create(ctx context.Context, section *domain.Section) error {
	return r.db.WithContext(ctx).Create(section).Error
}

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
func (r *SectionRepository) UpdateWithOptimisticLock(ctx context.Context, section *domain.Section) error {
	result := r.db.WithContext(ctx).Model(section).
		Where("section_id = ? AND version = ?", section.SectionID, section.Version-1).
		Updates(map[string]any{
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
func (r *SectionRepository) GetBySemester(ctx context.Context, semesterID uuid.UUID) ([]*domain.Section, error) {
	var sections []*domain.Section
	err := r.db.WithContext(ctx).
		Preload("Course").
		Preload("Semester").
		Where("semester_id = ?", semesterID).
		Find(&sections).Error
	if err != nil {
		return nil, err
	}
	return sections, nil
}
func (r *SectionRepository) GetAllActive(ctx context.Context) ([]*domain.Section, error) {
	var sections []*domain.Section
	err := r.db.WithContext(ctx).
		Preload("Course").
		Preload("Semester").
		Where("available_seats > 0").
		Find(&sections).Error
	if err != nil {
		return nil, err
	}
	return sections, nil
}
