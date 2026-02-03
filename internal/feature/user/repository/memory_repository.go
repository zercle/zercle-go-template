package repository

import (
	"context"
	"sync"
	"time"

	appErr "zercle-go-template/internal/errors"
	"zercle-go-template/internal/feature/user/domain"
)

// MemoryUserRepository is an in-memory implementation of UserRepository.
// It is safe for concurrent use and intended for testing and development.
type MemoryUserRepository struct {
	mu     sync.RWMutex
	users  map[string]*domain.User
	emails map[string]string // email -> userID mapping
}

// NewMemoryUserRepository creates a new in-memory user repository.
func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		users:  make(map[string]*domain.User),
		emails: make(map[string]string),
	}
}

// Create implements UserRepository.Create.
func (r *MemoryUserRepository) Create(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.emails[user.Email]; exists {
		return appErr.ConflictError("user with this email already exists")
	}

	// Clone the user to avoid external modifications
	userCopy := *user
	userCopy.CreatedAt = time.Now()
	userCopy.UpdatedAt = time.Now()

	r.users[user.ID] = &userCopy
	r.emails[user.Email] = user.ID

	return nil
}

// GetByID implements UserRepository.GetByID.
func (r *MemoryUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, appErr.NotFoundError("user")
	}

	// Return a copy to prevent external modification
	userCopy := *user
	return &userCopy, nil
}

// GetByEmail implements UserRepository.GetByEmail.
func (r *MemoryUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userID, exists := r.emails[email]
	if !exists {
		return nil, appErr.NotFoundError("user")
	}

	user, exists := r.users[userID]
	if !exists {
		return nil, appErr.NotFoundError("user")
	}

	// Return a copy to prevent external modification
	userCopy := *user
	return &userCopy, nil
}

// GetAll implements UserRepository.GetAll.
func (r *MemoryUserRepository) GetAll(ctx context.Context, offset, limit int) ([]*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// Convert map to slice
	allUsers := make([]*domain.User, 0, len(r.users))
	for _, user := range r.users {
		userCopy := *user
		allUsers = append(allUsers, &userCopy)
	}

	// Apply pagination
	if offset >= len(allUsers) {
		return []*domain.User{}, nil
	}

	end := offset + limit
	if end > len(allUsers) {
		end = len(allUsers)
	}

	return allUsers[offset:end], nil
}

// Count implements UserRepository.Count.
func (r *MemoryUserRepository) Count(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.users), nil
}

// Update implements UserRepository.Update.
func (r *MemoryUserRepository) Update(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existingUser, exists := r.users[user.ID]
	if !exists {
		return appErr.NotFoundError("user")
	}

	// If email is changing, update the email index
	if existingUser.Email != user.Email {
		// Check if new email already exists
		if _, emailExists := r.emails[user.Email]; emailExists {
			return appErr.ConflictError("user with this email already exists")
		}
		// Remove old email mapping
		delete(r.emails, existingUser.Email)
		// Add new email mapping
		r.emails[user.Email] = user.ID
	}

	// Clone and update the user
	userCopy := *user
	userCopy.UpdatedAt = time.Now()
	r.users[user.ID] = &userCopy

	return nil
}

// Delete implements UserRepository.Delete.
func (r *MemoryUserRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[id]
	if !exists {
		return appErr.NotFoundError("user")
	}

	// Remove from both maps
	delete(r.users, id)
	delete(r.emails, user.Email)

	return nil
}

// Exists implements UserRepository.Exists.
func (r *MemoryUserRepository) Exists(ctx context.Context, email string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.emails[email]
	return exists, nil
}
