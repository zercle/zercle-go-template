package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zercle/zercle-go-template/internal/feature/auth/domain"
	"github.com/zercle/zercle-go-template/internal/feature/auth/ports"
)

// UserRepository implements ports.UserRepository for PostgreSQL.
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new PostgreSQL user repository.
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user in the database.
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, display_name, avatar_url, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.Password,
		user.DisplayName,
		user.AvatarURL,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
	)
	return err
}

// FindByID finds a user by ID.
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, avatar_url, status, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`
	var user domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.DisplayName,
		&user.AvatarURL,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	return &user, err
}

// FindByEmail finds a user by email.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, avatar_url, status, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`
	var user domain.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.DisplayName,
		&user.AvatarURL,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	return &user, err
}

// FindByUsername finds a user by username.
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, avatar_url, status, created_at, updated_at, deleted_at
		FROM users
		WHERE username = $1 AND deleted_at IS NULL
	`
	var user domain.User
	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.DisplayName,
		&user.AvatarURL,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	return &user, err
}

// Update updates an existing user.
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET username = $2, email = $3, password_hash = $4, display_name = $5, avatar_url = $6, status = $7, updated_at = $8
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.Password,
		user.DisplayName,
		user.AvatarURL,
		user.Status,
		user.UpdatedAt,
	)
	return err
}

// Delete soft-deletes a user.
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET deleted_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

var _ ports.UserRepository = (*UserRepository)(nil)
