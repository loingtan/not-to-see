package repository

import (
	"context"

	domain "cobra-template/internal/domain/registration"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CourseRepository implements CourseRepository using GORM
type CourseRepository struct {
	db *gorm.DB
}

// NewCourseRepository creates a new GORM course repository
func NewCourseRepository(db *gorm.DB) domain.CourseRepository {
	return &CourseRepository{
		db: db,
	}
}

// Create creates a new course
func (r *CourseRepository) Create(ctx context.Context, course *domain.Course) error {
	return r.db.WithContext(ctx).Create(course).Error
}

// GetByID retrieves a course by ID
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

// GetByCode retrieves a course by course code
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
