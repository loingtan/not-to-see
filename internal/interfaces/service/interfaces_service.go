package service

import (
	domain "cobra-template/internal/domain/registration"
	infrastructure "cobra-template/internal/interfaces/infrastructure"
	"context"

	"github.com/google/uuid"
)

// Request/Response types for Registration Service
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
type RegistrationService interface {
	Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error)
	DropCourse(ctx context.Context, studentID, sectionID uuid.UUID) error
	GetStudentRegistrations(ctx context.Context, studentID uuid.UUID) ([]*domain.Registration, error)
	GetStudentWaitlistStatus(ctx context.Context, studentID uuid.UUID) ([]*domain.WaitlistEntry, error)
	GetAvailableSections(ctx context.Context, semesterID uuid.UUID, courseID *uuid.UUID) ([]*domain.Section, error)
	ProcessDatabaseSyncJob(ctx context.Context, job infrastructure.DatabaseSyncJob) error
	ProcessWaitlistJob(ctx context.Context, job infrastructure.WaitlistJob) error
	ProcessWaitlist(ctx context.Context, sectionID uuid.UUID) error
}
