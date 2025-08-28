package interfaces

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CacheService interface {
	// Seat management
	GetAvailableSeats(ctx context.Context, sectionID uuid.UUID) (int, error)
	SetAvailableSeats(ctx context.Context, sectionID uuid.UUID, seats int, ttl time.Duration) error
	DecrementAvailableSeats(ctx context.Context, sectionID uuid.UUID) error
	IncrementAvailableSeats(ctx context.Context, sectionID uuid.UUID) error
	DecrementAndGetAvailableSeats(ctx context.Context, sectionID uuid.UUID) (int, error)
	IncrementAndGetAvailableSeats(ctx context.Context, sectionID uuid.UUID) (int, error)

	// Section details
	GetSectionDetails(ctx context.Context, sectionID uuid.UUID) (interface{}, error)
	SetSectionDetails(ctx context.Context, sectionID uuid.UUID, data interface{}, ttl time.Duration) error
	GetAvailableSections(ctx context.Context, semesterID uuid.UUID) (interface{}, error)
	SetAvailableSections(ctx context.Context, semesterID uuid.UUID, data interface{}, ttl time.Duration) error

	// Course details
	GetCourseDetails(ctx context.Context, courseID uuid.UUID) (interface{}, error)
	SetCourseDetails(ctx context.Context, courseID uuid.UUID, data interface{}, ttl time.Duration) error

	// Student data
	GetStudentDetails(ctx context.Context, studentID uuid.UUID) (interface{}, error)
	SetStudentDetails(ctx context.Context, studentID uuid.UUID, data interface{}, ttl time.Duration) error
	GetStudentRegistrations(ctx context.Context, studentID uuid.UUID) (interface{}, error)
	SetStudentRegistrations(ctx context.Context, studentID uuid.UUID, data interface{}, ttl time.Duration) error
	GetStudentWaitlistStatus(ctx context.Context, studentID uuid.UUID) (interface{}, error)
	SetStudentWaitlistStatus(ctx context.Context, studentID uuid.UUID, data interface{}, ttl time.Duration) error

	// Generic cache operations for HTTP responses and other data
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	GetWithMetadata(ctx context.Context, key string) (string, map[string]string, error)
	SetWithMetadata(ctx context.Context, key string, value string, metadata map[string]string, ttl time.Duration) error

	// General cache operations
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context, pattern string) error
	InvalidateStudentCache(ctx context.Context, studentID uuid.UUID) error
	InvalidateSectionCache(ctx context.Context, sectionID uuid.UUID) error

	// Waitlist management using Redis sorted sets
	AddToWaitlist(ctx context.Context, sectionID, studentID uuid.UUID, position int, entry interface{}) error
	RemoveFromWaitlist(ctx context.Context, sectionID, studentID uuid.UUID) error
	GetNextInWaitlist(ctx context.Context, sectionID uuid.UUID) (interface{}, error)
	GetWaitlistPosition(ctx context.Context, sectionID, studentID uuid.UUID) (int, error)
	GetWaitlistSize(ctx context.Context, sectionID uuid.UUID) (int, error)
	GetStudentWaitlists(ctx context.Context, studentID uuid.UUID) ([]interface{}, error)

	// Cache statistics and monitoring
	GetCacheStats(ctx context.Context) (map[string]interface{}, error)

	// Health and connection management
	Health(ctx context.Context) error
	Close() error
}
