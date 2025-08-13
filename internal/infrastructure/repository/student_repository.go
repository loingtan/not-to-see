package repository

import (
	"context"

	domain "cobra-template/internal/domain/registration"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StudentRepository implements StudentRepository using GORM
type StudentRepository struct {
	db *gorm.DB
}

// NewStudentRepository creates a new GORM student repository
func NewStudentRepository(db *gorm.DB) domain.StudentRepository {
	return &StudentRepository{
		db: db,
	}
}

// Create creates a new student
func (r *StudentRepository) Create(ctx context.Context, student *domain.Student) error {
	return r.db.WithContext(ctx).Create(student).Error
}

// GetByID retrieves a student by ID
func (r *StudentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Student, error) {
	var student domain.Student
	err := r.db.WithContext(ctx).First(&student, "student_id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &student, nil
}

// GetByStudentNumber retrieves a student by student number
func (r *StudentRepository) GetByStudentNumber(ctx context.Context, studentNumber string) (*domain.Student, error) {
	var student domain.Student
	err := r.db.WithContext(ctx).First(&student, "student_number = ?", studentNumber).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &student, nil
}
