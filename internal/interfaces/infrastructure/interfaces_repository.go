package interfaces

import (
	domain "cobra-template/internal/domain/registration"
	"context"

	"github.com/google/uuid"
)

type StudentRepository interface {
	Create(ctx context.Context, student *domain.Student) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Student, error)
	GetByStudentNumber(ctx context.Context, studentNumber string) (*domain.Student, error)
}

type CourseRepository interface {
	Create(ctx context.Context, course *domain.Course) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Course, error)
	GetByCode(ctx context.Context, courseCode string) (*domain.Course, error)
}

type SemesterRepository interface {
	Create(ctx context.Context, semester *domain.Semester) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Semester, error)
	GetCurrent(ctx context.Context) (*domain.Semester, error)
}

type SectionRepository interface {
	Create(ctx context.Context, section *domain.Section) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Section, error)
	UpdateWithOptimisticLock(ctx context.Context, section *domain.Section) error
	GetByCourseAndSemester(ctx context.Context, courseID, semesterID uuid.UUID) ([]*domain.Section, error)
	GetBySemester(ctx context.Context, semesterID uuid.UUID) ([]*domain.Section, error)
	GetAllActive(ctx context.Context) ([]*domain.Section, error)
}

type RegistrationRepository interface {
	Create(ctx context.Context, registration *domain.Registration) error
	GetByStudentAndSection(ctx context.Context, studentID, sectionID uuid.UUID) (*domain.Registration, error)
	Update(ctx context.Context, registration *domain.Registration) error
	GetByStudentID(ctx context.Context, studentID uuid.UUID) ([]*domain.Registration, error)
	GetBySectionID(ctx context.Context, sectionID uuid.UUID) ([]*domain.Registration, error)
}

type WaitlistRepository interface {
	Create(ctx context.Context, entry *domain.WaitlistEntry) error
	GetByStudentAndSection(ctx context.Context, studentID, sectionID uuid.UUID) (*domain.WaitlistEntry, error)
	GetNextInLine(ctx context.Context, sectionID uuid.UUID) (*domain.WaitlistEntry, error)
	GetNextPosition(ctx context.Context, sectionID uuid.UUID) (int, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetBySectionID(ctx context.Context, sectionID uuid.UUID) ([]*domain.WaitlistEntry, error)
	GetByStudentID(ctx context.Context, studentID uuid.UUID) ([]*domain.WaitlistEntry, error)
}
