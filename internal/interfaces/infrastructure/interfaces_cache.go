package interfaces

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CacheService interface {
	GetAvailableSeats(ctx context.Context, sectionID uuid.UUID) (int, error)
	SetAvailableSeats(ctx context.Context, sectionID uuid.UUID, seats int, ttl time.Duration) error
	DecrementAvailableSeats(ctx context.Context, sectionID uuid.UUID) error
	IncrementAvailableSeats(ctx context.Context, sectionID uuid.UUID) error

	// Atomic operations that return the new value
	DecrementAndGetAvailableSeats(ctx context.Context, sectionID uuid.UUID) (int, error)
	IncrementAndGetAvailableSeats(ctx context.Context, sectionID uuid.UUID) (int, error)

	GetSectionDetails(ctx context.Context, sectionID uuid.UUID) (interface{}, error)
	SetSectionDetails(ctx context.Context, sectionID uuid.UUID, data interface{}, ttl time.Duration) error

	GetCourseDetails(ctx context.Context, courseID uuid.UUID) (interface{}, error)
	SetCourseDetails(ctx context.Context, courseID uuid.UUID, data interface{}, ttl time.Duration) error

	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context, pattern string) error
}
