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

type NotificationService interface {
	SendRegistrationSuccess(ctx context.Context, studentID, sectionID uuid.UUID) error
	SendRegistrationFailure(ctx context.Context, studentID, sectionID uuid.UUID, reason string) error
	SendWaitlistNotification(ctx context.Context, studentID, sectionID uuid.UUID, position int) error
	SendSeatAvailable(ctx context.Context, studentID, sectionID uuid.UUID) error
}
type RegistrationService interface {
	// Core registration operations
	Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error)
	DropCourse(ctx context.Context, studentID, sectionID uuid.UUID) error

	// Student query operations
	GetStudentRegistrations(ctx context.Context, studentID uuid.UUID) ([]*domain.Registration, error)
	GetStudentWaitlistStatus(ctx context.Context, studentID uuid.UUID) ([]*domain.WaitlistEntry, error)

	// Section query operations
	GetAvailableSections(ctx context.Context, semesterID uuid.UUID, courseID *uuid.UUID) ([]*domain.Section, error)

	// Internal processing (called by queue workers)
	ProcessRegistrationJob(ctx context.Context, job infrastructure.RegistrationJob) error
}
