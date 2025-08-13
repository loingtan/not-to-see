package domain

import (
	"context"

	"github.com/google/uuid"
)

// StudentRepository defines the interface for student data access
type StudentRepository interface {
	Create(ctx context.Context, student *Student) error
	GetByID(ctx context.Context, id uuid.UUID) (*Student, error)
	GetByStudentNumber(ctx context.Context, studentNumber string) (*Student, error)
}

// CourseRepository defines the interface for course data access
type CourseRepository interface {
	Create(ctx context.Context, course *Course) error
	GetByID(ctx context.Context, id uuid.UUID) (*Course, error)
	GetByCode(ctx context.Context, courseCode string) (*Course, error)
}

// SemesterRepository defines the interface for semester data access
type SemesterRepository interface {
	Create(ctx context.Context, semester *Semester) error
	GetByID(ctx context.Context, id uuid.UUID) (*Semester, error)
	GetCurrent(ctx context.Context) (*Semester, error)
}

// SectionRepository defines the interface for section data access
type SectionRepository interface {
	Create(ctx context.Context, section *Section) error
	GetByID(ctx context.Context, id uuid.UUID) (*Section, error)
	UpdateWithOptimisticLock(ctx context.Context, section *Section) error
	GetByCourseAndSemester(ctx context.Context, courseID, semesterID uuid.UUID) ([]*Section, error)
}

// RegistrationRepository defines the interface for registration data access
type RegistrationRepository interface {
	Create(ctx context.Context, registration *Registration) error
	GetByStudentAndSection(ctx context.Context, studentID, sectionID uuid.UUID) (*Registration, error)
	Update(ctx context.Context, registration *Registration) error
	GetByStudentID(ctx context.Context, studentID uuid.UUID) ([]*Registration, error)
	GetBySectionID(ctx context.Context, sectionID uuid.UUID) ([]*Registration, error)
}

// WaitlistRepository defines the interface for waitlist data access
type WaitlistRepository interface {
	Create(ctx context.Context, entry *WaitlistEntry) error
	GetByStudentAndSection(ctx context.Context, studentID, sectionID uuid.UUID) (*WaitlistEntry, error)
	GetNextInLine(ctx context.Context, sectionID uuid.UUID) (*WaitlistEntry, error)
	GetNextPosition(ctx context.Context, sectionID uuid.UUID) (int, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetBySectionID(ctx context.Context, sectionID uuid.UUID) ([]*WaitlistEntry, error)
	GetByStudentID(ctx context.Context, studentID uuid.UUID) ([]*WaitlistEntry, error)
}
