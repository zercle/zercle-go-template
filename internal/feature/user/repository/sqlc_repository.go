// Package repository provides data access layer implementations for the user feature.
// This file contains the sqlc-based repository implementation.
package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	appErr "zercle-go-template/internal/errors"
	"zercle-go-template/internal/feature/user/domain"
	"zercle-go-template/internal/infrastructure/db/sqlc"
)

// SqlcUserRepository is a PostgreSQL-based implementation of UserRepository using sqlc.
// It provides type-safe database access for user operations.
type SqlcUserRepository struct {
	querier sqlc.Querier
}

// NewSqlcUserRepository creates a new sqlc-based user repository.
func NewSqlcUserRepository(querier sqlc.Querier) *SqlcUserRepository {
	return &SqlcUserRepository{
		querier: querier,
	}
}

// toDomainUser converts a sqlc.User to a domain.User
func toDomainUser(u sqlc.User) *domain.User {
	return &domain.User{
		ID:           u.ID.String(),
		Email:        u.Email,
		Name:         u.Name,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

// Create creates a new user in the repository.
func (r *SqlcUserRepository) Create(ctx context.Context, user *domain.User) error {
	id, err := uuid.Parse(user.ID)
	if err != nil {
		return appErr.ValidationError("invalid user ID format")
	}

	params := sqlc.CreateUserParams{
		ID:           id,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Name:         user.Name,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}

	created, err := r.querier.CreateUser(ctx, params)
	if err != nil {
		// Handle unique constraint violation
		if isDuplicateError(err) {
			return appErr.ConflictError("user with this email already exists")
		}
		return appErr.InternalError("failed to create user").WithCause(err)
	}

	// Update the user with the created values
	*user = *toDomainUser(created)
	return nil
}

// GetByID retrieves a user by their unique ID.
func (r *SqlcUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, appErr.ValidationError("invalid user ID format")
	}

	user, err := r.querier.GetUserByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, appErr.NotFoundError("user")
		}
		return nil, appErr.InternalError("failed to get user by ID").WithCause(err)
	}

	return toDomainUser(user), nil
}

// GetByEmail retrieves a user by their email address.
func (r *SqlcUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := r.querier.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, appErr.NotFoundError("user")
		}
		return nil, appErr.InternalError("failed to get user by email").WithCause(err)
	}

	return toDomainUser(user), nil
}

// GetAll retrieves all users with pagination support.
func (r *SqlcUserRepository) GetAll(ctx context.Context, offset, limit int) ([]*domain.User, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	params := sqlc.ListUsersParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	users, err := r.querier.ListUsers(ctx, params)
	if err != nil {
		return nil, appErr.InternalError("failed to list users").WithCause(err)
	}

	result := make([]*domain.User, len(users))
	for i, u := range users {
		user := toDomainUser(u)
		result[i] = user
	}

	return result, nil
}

// Count returns the total number of users.
func (r *SqlcUserRepository) Count(ctx context.Context) (int, error) {
	count, err := r.querier.CountUsers(ctx)
	if err != nil {
		return 0, appErr.InternalError("failed to count users").WithCause(err)
	}
	return int(count), nil
}

// Update updates an existing user.
func (r *SqlcUserRepository) Update(ctx context.Context, user *domain.User) error {
	id, err := uuid.Parse(user.ID)
	if err != nil {
		return appErr.ValidationError("invalid user ID format")
	}

	// Set updated_at to current time before updating
	params := sqlc.UpdateUserParams{
		ID:           id,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Name:         user.Name,
		UpdatedAt:    time.Now(),
	}

	updated, err := r.querier.UpdateUser(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return appErr.NotFoundError("user")
		}
		// Handle unique constraint violation
		if isDuplicateError(err) {
			return appErr.ConflictError("user with this email already exists")
		}
		return appErr.InternalError("failed to update user").WithCause(err)
	}

	// Update the user with the updated values
	*user = *toDomainUser(updated)
	return nil
}

// Delete removes a user by their ID.
func (r *SqlcUserRepository) Delete(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return appErr.ValidationError("invalid user ID format")
	}

	// Check if user exists before deleting
	_, err = r.querier.GetUserByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return appErr.NotFoundError("user")
		}
		return appErr.InternalError("failed to check user existence").WithCause(err)
	}

	// User exists, proceed with deletion
	err = r.querier.DeleteUser(ctx, uid)
	if err != nil {
		return appErr.InternalError("failed to delete user").WithCause(err)
	}

	return nil
}

// Exists checks if a user with the given email exists.
func (r *SqlcUserRepository) Exists(ctx context.Context, email string) (bool, error) {
	exists, err := r.querier.CheckUserExists(ctx, email)
	if err != nil {
		return false, appErr.InternalError("failed to check user existence").WithCause(err)
	}
	return exists, nil
}

// isDuplicateError checks if the error is a duplicate/unique constraint violation
func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	// Check for PostgreSQL unique violation error code (23505)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	// Fallback to string matching for other error types
	errStr := err.Error()
	return strings.Contains(errStr, "duplicate key") ||
		strings.Contains(errStr, "unique constraint") ||
		strings.Contains(errStr, "23505")
}
