package service

import (
	"errors"
	"fmt"

	"cobra-template/internal/domain/user"
	"cobra-template/pkg/logger"

	"github.com/google/uuid"
)

// userService implements the UserService interface
type userService struct {
	userRepo user.UserRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo user.UserRepository) user.UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// CreateUser creates a new user
func (s *userService) CreateUser(req *user.CreateUserRequest) (*user.User, error) {
	logger.Info("Creating user with username: %s", req.Username)

	// Check if user already exists by email
	existingUser, err := s.userRepo.GetByEmail(req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Check if user already exists by username
	existingUser, err = s.userRepo.GetByUsername(req.Username)
	if err == nil && existingUser != nil {
		return nil, errors.New("user with this username already exists")
	}

	// Create new user
	user := user.NewUser(req.Username, req.Email, req.FirstName, req.LastName)

	// Save user
	if err := s.userRepo.Create(user); err != nil {
		logger.Error("Failed to create user: %v", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	logger.Info("User created successfully with ID: %s", user.ID)
	return user, nil
}

// GetUser retrieves a user by ID
func (s *userService) GetUser(id uuid.UUID) (*user.User, error) {
	logger.Debug("Getting user with ID: %s", id)

	user, err := s.userRepo.GetByID(id)
	if err != nil {
		logger.Error("Failed to get user: %v", err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *userService) GetUserByEmail(email string) (*user.User, error) {
	logger.Debug("Getting user with email: %s", email)

	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		logger.Error("Failed to get user by email: %v", err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username
func (s *userService) GetUserByUsername(username string) (*user.User, error) {
	logger.Debug("Getting user with username: %s", username)

	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		logger.Error("Failed to get user by username: %v", err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// UpdateUser updates an existing user
func (s *userService) UpdateUser(id uuid.UUID, req *user.UpdateUserRequest) (*user.User, error) {
	logger.Info("Updating user with ID: %s", id)

	// Get existing user
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		logger.Error("Failed to get user for update: %v", err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	// Update fields if provided
	if req.Username != nil {
		// Check if username is already taken by another user
		existingUser, err := s.userRepo.GetByUsername(*req.Username)
		if err == nil && existingUser != nil && existingUser.ID != id {
			return nil, errors.New("username already taken")
		}
		user.Username = *req.Username
	}

	if req.Email != nil {
		// Check if email is already taken by another user
		existingUser, err := s.userRepo.GetByEmail(*req.Email)
		if err == nil && existingUser != nil && existingUser.ID != id {
			return nil, errors.New("email already taken")
		}
		user.Email = *req.Email
	}

	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}

	if req.LastName != nil {
		user.LastName = *req.LastName
	}

	if req.Active != nil {
		user.Active = *req.Active
	}

	// Save updated user
	if err := s.userRepo.Update(user); err != nil {
		logger.Error("Failed to update user: %v", err)
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	logger.Info("User updated successfully with ID: %s", user.ID)
	return user, nil
}

// DeleteUser deletes a user
func (s *userService) DeleteUser(id uuid.UUID) error {
	logger.Info("Deleting user with ID: %s", id)

	// Check if user exists
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		logger.Error("Failed to get user for deletion: %v", err)
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return errors.New("user not found")
	}

	// Delete user
	if err := s.userRepo.Delete(id); err != nil {
		logger.Error("Failed to delete user: %v", err)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	logger.Info("User deleted successfully with ID: %s", id)
	return nil
}

// ListUsers retrieves a list of users
func (s *userService) ListUsers(limit, offset int) ([]*user.User, error) {
	logger.Debug("Listing users with limit: %d, offset: %d", limit, offset)

	users, err := s.userRepo.List(limit, offset)
	if err != nil {
		logger.Error("Failed to list users: %v", err)
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}
