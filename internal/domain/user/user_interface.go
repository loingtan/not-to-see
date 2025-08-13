package user

import "github.com/google/uuid"

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(user *User) error
	GetByID(id uuid.UUID) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByUsername(username string) (*User, error)
	Update(user *User) error
	Delete(id uuid.UUID) error
	List(limit, offset int) ([]*User, error)
}

// UserService defines the interface for user business logic
type UserService interface {
	CreateUser(req *CreateUserRequest) (*User, error)
	GetUser(id uuid.UUID) (*User, error)
	GetUserByEmail(email string) (*User, error)
	GetUserByUsername(username string) (*User, error)
	UpdateUser(id uuid.UUID, req *UpdateUserRequest) (*User, error)
	DeleteUser(id uuid.UUID) error
	ListUsers(limit, offset int) ([]*User, error)
}
