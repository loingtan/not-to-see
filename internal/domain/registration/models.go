package domain

import (
	"time"

	"github.com/google/uuid"
)

// Student represents a student in the system
type Student struct {
	StudentID        uuid.UUID `json:"student_id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	StudentNumber    string    `json:"student_number" gorm:"unique;not null"`
	EnrollmentStatus string    `json:"enrollment_status" gorm:"not null;default:active"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Course represents a course in the system
type Course struct {
	CourseID    uuid.UUID `json:"course_id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CourseCode  string    `json:"course_code" gorm:"unique;not null"`
	CourseName  string    `json:"course_name" gorm:"not null"`
	Description *string   `json:"description"`
	Credits     int       `json:"credits" gorm:"not null;check:credits > 0"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Semester represents an academic semester
type Semester struct {
	SemesterID   uuid.UUID `json:"semester_id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	SemesterCode string    `json:"semester_code" gorm:"unique;not null"`
	SemesterName string    `json:"semester_name" gorm:"not null"`
	StartDate    time.Time `json:"start_date" gorm:"not null"`
	EndDate      time.Time `json:"end_date" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Section represents a course section offered in a specific semester
type Section struct {
	SectionID      uuid.UUID `json:"section_id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CourseID       uuid.UUID `json:"course_id" gorm:"type:uuid;not null"`
	SemesterID     uuid.UUID `json:"semester_id" gorm:"type:uuid;not null"`
	SectionNumber  string    `json:"section_number" gorm:"not null"`
	TotalSeats     int       `json:"total_seats" gorm:"not null;check:total_seats > 0"`
	AvailableSeats int       `json:"available_seats" gorm:"not null;check:available_seats >= 0"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	Version        int       `json:"version" gorm:"default:1"`
	Course         Course    `json:"course,omitempty" gorm:"foreignKey:CourseID"`
	Semester       Semester  `json:"semester,omitempty" gorm:"foreignKey:SemesterID"`
}

// RegistrationStatus represents the status of a registration
type RegistrationStatus string

const (
	StatusEnrolled   RegistrationStatus = "enrolled"
	StatusWaitlisted RegistrationStatus = "waitlisted"
	StatusDropped    RegistrationStatus = "dropped"
	StatusFailed     RegistrationStatus = "failed"
)

// Registration represents a student's registration for a course section
type Registration struct {
	RegistrationID   uuid.UUID          `json:"registration_id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	StudentID        uuid.UUID          `json:"student_id" gorm:"type:uuid;not null"`
	SectionID        uuid.UUID          `json:"section_id" gorm:"type:uuid;not null"`
	Status           RegistrationStatus `json:"status" gorm:"type:text;default:enrolled"`
	RegistrationDate time.Time          `json:"registration_date" gorm:"not null"`
	CreatedAt        time.Time          `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time          `json:"updated_at" gorm:"autoUpdateTime"`
	Version          int                `json:"version" gorm:"default:1"`
	Student          Student            `json:"student,omitempty" gorm:"foreignKey:StudentID"`
	Section          Section            `json:"section,omitempty" gorm:"foreignKey:SectionID"`
}

// WaitlistEntry represents a student's position on a waitlist
type WaitlistEntry struct {
	WaitlistID uuid.UUID `json:"waitlist_id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	StudentID  uuid.UUID `json:"student_id" gorm:"type:uuid;not null"`
	SectionID  uuid.UUID `json:"section_id" gorm:"type:uuid;not null"`
	Position   int       `json:"position" gorm:"not null"`
	Timestamp  time.Time `json:"timestamp" gorm:"not null"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	Student    Student   `json:"student,omitempty" gorm:"foreignKey:StudentID"`
	Section    Section   `json:"section,omitempty" gorm:"foreignKey:SectionID"`
}

// Request DTOs

// RegisterRequest represents a registration request
type RegisterRequest struct {
	StudentID  uuid.UUID   `json:"student_id" validate:"required"`
	SectionIDs []uuid.UUID `json:"section_ids" validate:"required,min=1"`
}

// RegisterResponse represents the response for registration
type RegisterResponse struct {
	Results []RegistrationResult `json:"results"`
}

// RegistrationResult represents the result of a single section registration
type RegistrationResult struct {
	SectionID uuid.UUID `json:"section_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Position  *int      `json:"waitlist_position,omitempty"`
}
