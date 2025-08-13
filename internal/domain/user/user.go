package user

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	Username  string `json:"username" validate:"required,min=3,max=50"`
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"first_name" validate:"required,min=1,max=50"`
	LastName  string `json:"last_name" validate:"required,min=1,max=50"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Username  *string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	Email     *string `json:"email,omitempty" validate:"omitempty,email"`
	FirstName *string `json:"first_name,omitempty" validate:"omitempty,min=1,max=50"`
	LastName  *string `json:"last_name,omitempty" validate:"omitempty,min=1,max=50"`
	Active    *bool   `json:"active,omitempty"`
}

// NewUser creates a new user with generated ID and timestamps
func NewUser(username, email, firstName, lastName string) *User {
	now := time.Now()
	return &User{
		ID:        uuid.New(),
		Username:  username,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// FullName returns the full name of the user
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// IsValid validates the user data
func (u *User) IsValid() bool {
	return u.ID != uuid.Nil &&
		u.Username != "" &&
		u.Email != "" &&
		u.FirstName != "" &&
		u.LastName != ""
}
