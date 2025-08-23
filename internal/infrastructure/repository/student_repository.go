package repository

import (
	"context"

	domain "cobra-template/internal/domain/registration"
	interfaces "cobra-template/internal/interfaces/infrastructure"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StudentRepository struct {
	db *gorm.DB
}

func NewStudentRepository(db *gorm.DB) interfaces.StudentRepository {
	return &StudentRepository{
		db: db,
	}
}

func (r *StudentRepository) Create(ctx context.Context, student *domain.Student) error {
	return r.db.WithContext(ctx).Create(student).Error
}

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
