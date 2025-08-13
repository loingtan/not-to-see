package repository

import (
	"cobra-template/internal/domain/user"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// mockUserRepository is an in-memory implementation of UserRepository for testing/demo purposes
type mockUserRepository struct {
	users map[uuid.UUID]*user.User
	mutex sync.RWMutex
}

// NewMockUserRepository creates a new mock user repository
func NewMockUserRepository() user.UserRepository {
	repo := &mockUserRepository{
		users: make(map[uuid.UUID]*user.User),
		mutex: sync.RWMutex{},
	}

	// Add some sample data
	repo.seedData()
	return repo
}

// Create creates a new user
func (r *mockUserRepository) Create(user *user.User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if user already exists
	if _, exists := r.users[user.ID]; exists {
		return errors.New("user already exists")
	}

	// Check for duplicate email
	for _, existingUser := range r.users {
		if existingUser.Email == user.Email {
			return errors.New("email already exists")
		}
	}

	// Check for duplicate username
	for _, existingUser := range r.users {
		if existingUser.Username == user.Username {
			return errors.New("username already exists")
		}
	}

	r.users[user.ID] = user
	return nil
}

// GetByID retrieves a user by ID
func (r *mockUserRepository) GetByID(id uuid.UUID) (*user.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *mockUserRepository) GetByEmail(email string) (*user.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}

	return nil, errors.New("user not found")
}

// GetByUsername retrieves a user by username
func (r *mockUserRepository) GetByUsername(username string) (*user.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, user := range r.users {
		if user.Username == username {
			return user, nil
		}
	}

	return nil, errors.New("user not found")
}

// Update updates an existing user
func (r *mockUserRepository) Update(user *user.User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if user exists
	if _, exists := r.users[user.ID]; !exists {
		return errors.New("user not found")
	}

	// Check for duplicate email (excluding current user)
	for id, existingUser := range r.users {
		if id != user.ID && existingUser.Email == user.Email {
			return errors.New("email already exists")
		}
	}

	// Check for duplicate username (excluding current user)
	for id, existingUser := range r.users {
		if id != user.ID && existingUser.Username == user.Username {
			return errors.New("username already exists")
		}
	}

	user.UpdatedAt = time.Now()
	r.users[user.ID] = user
	return nil
}

// Delete deletes a user
func (r *mockUserRepository) Delete(id uuid.UUID) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if user exists
	if _, exists := r.users[id]; !exists {
		return errors.New("user not found")
	}

	delete(r.users, id)
	return nil
}

// List retrieves a list of users with pagination
func (r *mockUserRepository) List(limit, offset int) ([]*user.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var users []*user.User
	count := 0

	for _, user := range r.users {
		if count >= offset {
			users = append(users, user)
			if len(users) >= limit {
				break
			}
		}
		count++
	}

	return users, nil
}

// seedData adds some sample users for demonstration
func (r *mockUserRepository) seedData() {
	sampleUsers := []*user.User{
		{
			ID:        uuid.New(),
			Username:  "john_doe",
			Email:     "john.doe@example.com",
			FirstName: "John",
			LastName:  "Doe",
			Active:    true,
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now().Add(-24 * time.Hour),
		},
		{
			ID:        uuid.New(),
			Username:  "jane_smith",
			Email:     "jane.smith@example.com",
			FirstName: "Jane",
			LastName:  "Smith",
			Active:    true,
			CreatedAt: time.Now().Add(-12 * time.Hour),
			UpdatedAt: time.Now().Add(-12 * time.Hour),
		},
		{
			ID:        uuid.New(),
			Username:  "bob_wilson",
			Email:     "bob.wilson@example.com",
			FirstName: "Bob",
			LastName:  "Wilson",
			Active:    false,
			CreatedAt: time.Now().Add(-6 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		},
	}

	for _, user := range sampleUsers {
		r.users[user.ID] = user
	}
}
