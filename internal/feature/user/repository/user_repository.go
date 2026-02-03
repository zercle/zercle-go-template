// Package repository provides data access layer implementations for the user feature.
// It defines interfaces for data persistence and concrete implementations.
package repository

//go:generate mockgen -source=$GOFILE -destination=./mocks/$GOFILE -package=mocks

import (
	"context"

	"zercle-go-template/internal/feature/user/domain"
)

// UserRepository defines the interface for user data access.
// Implementations can be in-memory, database, or external API.
type UserRepository interface {
	// Create creates a new user in the repository.
	Create(ctx context.Context, user *domain.User) error

	// GetByID retrieves a user by their unique ID.
	GetByID(ctx context.Context, id string) (*domain.User, error)

	// GetByEmail retrieves a user by their email address.
	GetByEmail(ctx context.Context, email string) (*domain.User, error)

	// GetAll retrieves all users with pagination support.
	GetAll(ctx context.Context, offset, limit int) ([]*domain.User, error)

	// Count returns the total number of users.
	Count(ctx context.Context) (int, error)

	// Update updates an existing user.
	Update(ctx context.Context, user *domain.User) error

	// Delete removes a user by their ID.
	Delete(ctx context.Context, id string) error

	// Exists checks if a user with the given email exists.
	Exists(ctx context.Context, email string) (bool, error)
}
