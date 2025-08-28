package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Registration struct {
	RegistrationID   uuid.UUID          `json:"registration_id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	StudentID        uuid.UUID          `json:"student_id" gorm:"type:uuid;not null;constraint:OnDelete:CASCADE"`
	SectionID        uuid.UUID          `json:"section_id" gorm:"type:uuid;not null;constraint:OnDelete:CASCADE"`
	Status           RegistrationStatus `json:"status" gorm:"type:registration_status;default:enrolled"`
	RegistrationDate time.Time          `json:"registration_date" gorm:"type:timestamptz;default:now()"`
	CreatedAt        time.Time          `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time          `json:"updated_at" gorm:"autoUpdateTime"`
	Version          int                `json:"version" gorm:"default:1"`
	Student          Student            `json:"student,omitempty" gorm:"foreignKey:StudentID;references:StudentID"`
	Section          Section            `json:"section,omitempty" gorm:"foreignKey:SectionID;references:SectionID"`
}

func (Registration) TableName() string {
	return "registrations"
}

func (r *Registration) BeforeCreate(db *gorm.DB) error {
	return nil
}

type Student struct {
	StudentID        uuid.UUID `json:"student_id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	StudentNumber    string    `json:"student_number" gorm:"type:text;unique;not null"`
	FirstName        string    `json:"first_name" gorm:"type:varchar(100);not null"`
	LastName         string    `json:"last_name" gorm:"type:varchar(100);not null"`
	EnrollmentStatus string    `json:"enrollment_status" gorm:"type:varchar(20);default:'active'"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	Version          int       `json:"version" gorm:"default:1"`
}

func (Student) TableName() string {
	return "students"
}

type Course struct {
	CourseID   uuid.UUID `json:"course_id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CourseCode string    `json:"course_code" gorm:"type:text;unique;not null"`
	CourseName string    `json:"course_name" gorm:"type:text;not null"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	Version    int       `json:"version" gorm:"default:1"`
}

func (Course) TableName() string {
	return "courses"
}

type Semester struct {
	SemesterID        uuid.UUID `json:"semester_id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	SemesterCode      string    `json:"semester_code" gorm:"type:text;unique;not null"`
	SemesterName      string    `json:"semester_name" gorm:"type:text;not null"`
	StartDate         time.Time `json:"start_date" gorm:"type:date;not null"`
	EndDate           time.Time `json:"end_date" gorm:"type:date;not null"`
	RegistrationStart time.Time `json:"registration_start" gorm:"type:timestamptz;not null"`
	RegistrationEnd   time.Time `json:"registration_end" gorm:"type:timestamptz;not null"`
	IsActive          bool      `json:"is_active" gorm:"default:true"`
	CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Semester) TableName() string {
	return "semesters"
}

type Section struct {
	SectionID      uuid.UUID `json:"section_id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CourseID       uuid.UUID `json:"course_id" gorm:"type:uuid;not null;constraint:OnDelete:CASCADE"`
	SemesterID     uuid.UUID `json:"semester_id" gorm:"type:uuid;not null;constraint:OnDelete:CASCADE"`
	SectionNumber  string    `json:"section_number" gorm:"type:varchar(10);not null"`
	TotalSeats     int       `json:"total_seats" gorm:"not null;check:total_seats > 0"`
	AvailableSeats int       `json:"available_seats" gorm:"not null;check:available_seats >= 0;default:0"`
	IsActive       bool      `json:"is_active" gorm:"default:true"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	Version        int       `json:"version" gorm:"default:1"`
	Course         Course    `json:"course,omitempty" gorm:"foreignKey:CourseID;references:CourseID"`
	Semester       Semester  `json:"semester,omitempty" gorm:"foreignKey:SemesterID;references:SemesterID"`
}

func (Section) TableName() string {
	return "sections"
}

func (s *Section) BeforeCreate(db *gorm.DB) error {
	return nil
}

type RegistrationStatus string

const (
	StatusEnrolled   RegistrationStatus = "enrolled"
	StatusWaitlisted RegistrationStatus = "waitlisted"
	StatusDropped    RegistrationStatus = "dropped"
	StatusFailed     RegistrationStatus = "failed"
)

type WaitlistEntry struct {
	WaitlistID uuid.UUID  `json:"waitlist_id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	StudentID  uuid.UUID  `json:"student_id" gorm:"type:uuid;not null;constraint:OnDelete:CASCADE"`
	SectionID  uuid.UUID  `json:"section_id" gorm:"type:uuid;not null;constraint:OnDelete:CASCADE"`
	Position   int        `json:"position" gorm:"not null"`
	Timestamp  time.Time  `json:"timestamp" gorm:"type:timestamptz;default:now()"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty" gorm:"type:timestamptz"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	Student    Student    `json:"student,omitempty" gorm:"foreignKey:StudentID;references:StudentID"`
	Section    Section    `json:"section,omitempty" gorm:"foreignKey:SectionID;references:SectionID"`
}

func (WaitlistEntry) TableName() string {
	return "waitlist"
}

func (w *WaitlistEntry) BeforeCreate(db *gorm.DB) error {
	return nil
}

type RegisterRequest struct {
	StudentID  uuid.UUID   `json:"student_id" validate:"required"`
	SectionIDs []uuid.UUID `json:"section_ids" validate:"required,min=1"`
}

type RegisterResponse struct {
	Results []RegistrationResult `json:"results"`
}

type RegistrationResult struct {
	SectionID uuid.UUID `json:"section_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Position  *int      `json:"waitlist_position,omitempty"`
}

type DropCourseRequest struct {
	StudentID uuid.UUID `json:"student_id" validate:"required"`
	SectionID uuid.UUID `json:"section_id" validate:"required"`
}

func GetUniqueConstraints() []string {
	return []string{
		"ALTER TABLE sections ADD CONSTRAINT unique_course_semester_section UNIQUE (course_id, semester_id, section_number);",
		"ALTER TABLE registrations ADD CONSTRAINT unique_student_section UNIQUE (student_id, section_id);",
		"ALTER TABLE waitlist ADD CONSTRAINT unique_waitlist_student_section UNIQUE (student_id, section_id);",
	}
}

func GetCheckConstraints() []string {
	return []string{
		"ALTER TABLE sections ADD CONSTRAINT check_total_seats_positive CHECK (total_seats > 0);",
	}
}

type IdempotencyKey struct {
	Key          string    `json:"key"`
	StudentID    uuid.UUID `json:"student_id"`
	RequestHash  string    `json:"request_hash"`
	ResponseData string    `json:"response_data"`
	StatusCode   int       `json:"status_code"`
	ProcessedAt  time.Time `json:"processed_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

func (i *IdempotencyKey) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}
