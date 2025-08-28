package repository

import (
	"context"

	domain "cobra-template/internal/domain/registration"
	interfaces "cobra-template/internal/interfaces/infrastructure"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CourseRepository struct {
	db *gorm.DB
}

func NewCourseRepository(db *gorm.DB) interfaces.CourseRepository {
	return &CourseRepository{
		db: db,
	}
}

func (r *CourseRepository) Create(ctx context.Context, course *domain.Course) error {
	return r.db.WithContext(ctx).Create(course).Error
}

func (r *CourseRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Course, error) {
	var course domain.Course
	err := r.db.WithContext(ctx).First(&course, "course_id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &course, nil
}

func (r *CourseRepository) GetByCode(ctx context.Context, courseCode string) (*domain.Course, error) {
	var course domain.Course
	err := r.db.WithContext(ctx).First(&course, "course_code = ?", courseCode).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &course, nil
}

func (r *CourseRepository) GetAllActive(ctx context.Context) ([]*domain.Course, error) {
	var courses []*domain.Course

	// Get all active courses - you might want to add additional criteria
	// like checking if the course has active sections in current semester
	err := r.db.WithContext(ctx).
		Where("active = ?", true). // Assuming there's an active field
		Find(&courses).Error

	if err != nil {
		return nil, err
	}

	return courses, nil
}
