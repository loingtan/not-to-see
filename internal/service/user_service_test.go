package service

import (
	"testing"

	"cobra-template/internal/domain"
	"cobra-template/internal/infrastructure/repository"
)

func TestUserService_CreateUser(t *testing.T) {
	// Initialize dependencies
	userRepo := repository.NewMockUserRepository()
	userService := NewUserService(userRepo)

	// Test data
	req := &domain.CreateUserRequest{
		Username:  "testuser",
		Email:     "testuser@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	// Create user
	user, err := userService.CreateUser(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if user == nil {
		t.Fatal("Expected user to be created, got nil")
	}

	// Verify user data
	if user.Username != req.Username {
		t.Errorf("Expected username %s, got %s", req.Username, user.Username)
	}

	if user.Email != req.Email {
		t.Errorf("Expected email %s, got %s", req.Email, user.Email)
	}

	if user.FirstName != req.FirstName {
		t.Errorf("Expected first name %s, got %s", req.FirstName, user.FirstName)
	}

	if user.LastName != req.LastName {
		t.Errorf("Expected last name %s, got %s", req.LastName, user.LastName)
	}

	if !user.Active {
		t.Error("Expected user to be active")
	}
}

func TestUserService_CreateUser_DuplicateEmail(t *testing.T) {
	// Initialize dependencies
	userRepo := repository.NewMockUserRepository()
	userService := NewUserService(userRepo)

	// Test data with existing email
	req := &domain.CreateUserRequest{
		Username:  "newuser",
		Email:     "john.doe@example.com", // This email already exists in mock data
		FirstName: "New",
		LastName:  "User",
	}

	// Try to create user with duplicate email
	user, err := userService.CreateUser(req)
	if err == nil {
		t.Fatal("Expected error for duplicate email, got nil")
	}

	if user != nil {
		t.Fatal("Expected nil user for duplicate email, got user")
	}

	expectedError := "user with this email already exists"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

func TestUserService_GetUser(t *testing.T) {
	// Initialize dependencies
	userRepo := repository.NewMockUserRepository()
	userService := NewUserService(userRepo)

	// Get user by email first to get the ID (since we don't know the mock IDs)
	user, err := userService.GetUserByEmail("john.doe@example.com")
	if err != nil {
		t.Fatalf("Failed to get user by email: %v", err)
	}

	// Now get user by ID
	foundUser, err := userService.GetUser(user.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if foundUser == nil {
		t.Fatal("Expected user to be found, got nil")
	}

	if foundUser.ID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, foundUser.ID)
	}
}

func TestUserService_ListUsers(t *testing.T) {
	// Initialize dependencies
	userRepo := repository.NewMockUserRepository()
	userService := NewUserService(userRepo)

	// List users
	users, err := userService.ListUsers(10, 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if users == nil {
		t.Fatal("Expected users slice, got nil")
	}

	// Should have at least the seeded users
	if len(users) < 3 {
		t.Errorf("Expected at least 3 users from seed data, got %d", len(users))
	}
}
