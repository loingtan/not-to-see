package service

import (
	"context"

	"github.com/google/uuid"
)

type NotificationService interface {
	SendRegistrationSuccess(ctx context.Context, studentID, sectionID uuid.UUID) error
	SendRegistrationFailure(ctx context.Context, studentID, sectionID uuid.UUID, reason string) error
	SendWaitlistNotification(ctx context.Context, studentID, sectionID uuid.UUID, position int) error
	SendSeatAvailable(ctx context.Context, studentID, sectionID uuid.UUID) error
}
