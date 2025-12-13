package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/internal/core/port"
	userDomain "github.com/zercle/zercle-go-template/internal/features/user/domain"
	db "github.com/zercle/zercle-go-template/internal/infrastructure/sqlc"
	sharederrors "github.com/zercle/zercle-go-template/internal/shared/errors"
)

type userRepository struct {
	q  *db.Queries
	db *sql.DB
}

// NewUserRepository creates a new MySQL repository for Users.
func NewUserRepository(d *sql.DB) port.UserRepository {
	return &userRepository{
		q:  db.New(d),
		db: d,
	}
}

func (r *userRepository) Create(ctx context.Context, u *userDomain.User) error {
	err := r.q.CreateUser(ctx, db.CreateUserParams{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.Password,
		FirstName:    sql.NullString{String: u.Name, Valid: u.Name != ""},
		IsActive:     sql.NullBool{Bool: true, Valid: true},
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*userDomain.User, error) {
	u, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharederrors.ErrNotFound
		}
		return nil, err
	}
	return mapUserToDomain(u), nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*userDomain.User, error) {
	u, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharederrors.ErrNotFound
		}
		return nil, err
	}
	return mapUserToDomain(u), nil
}

func (r *userRepository) Update(ctx context.Context, u *userDomain.User) error {
	// Updating only fields present in User domain that are also in DB and mutable
	return r.q.UpdateUser(ctx, db.UpdateUserParams{
		Email:        u.Email,
		PasswordHash: u.Password,
		FirstName:    sql.NullString{String: u.Name, Valid: u.Name != ""},
		LastName:     sql.NullString{}, // Not mapped yet
		IsActive:     sql.NullBool{Bool: true, Valid: true},
		ID:           u.ID,
	})
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteUser(ctx, id)
}

func mapUserToDomain(u db.User) *userDomain.User {
	return &userDomain.User{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.FirstName.String,
		Password:  u.PasswordHash,
		CreatedAt: u.CreatedAt.Time,
		UpdatedAt: u.UpdatedAt.Time,
	}
}
