package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	user_entity "github.com/zercle/zercle-go-template/internal/feature/user"
	user_sqlc "github.com/zercle/zercle-go-template/internal/infrastructure/database/sqlc/user"
)

// postgresRepository implements user.Repository using PostgreSQL.
type postgresRepository struct {
	queries *user_sqlc.Queries
}

// NewPostgresRepository creates a new PostgreSQL implementation of user.Repository.
func NewPostgresRepository(db user_sqlc.DBTX) user_entity.Repository {
	return &postgresRepository{
		queries: user_sqlc.New(db),
	}
}

// mapPgError maps PostgreSQL errors to domain errors.
func mapPgError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			if pgErr.ConstraintName == "users_email_key" || pgErr.ConstraintName == "users_email_idx" {
				return user_entity.ErrDuplicateEmail
			}
		case "23503": // foreign_key_violation
			// Could map to a specific domain error if needed
		}
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return user_entity.ErrUserNotFound
	}

	return err
}

// Create inserts a new user into the data store.
func (r *postgresRepository) Create(ctx context.Context, u *user_entity.User) (*user_entity.User, error) {
	params := domainUserToCreateParams(u)

	result, err := r.queries.CreateUser(ctx, &params)
	if err != nil {
		return nil, mapPgError(err)
	}

	return sqlcUserToDomain(result)
}

// GetByID retrieves a user by their ID.
func (r *postgresRepository) GetByID(ctx context.Context, id user_entity.UserID) (*user_entity.User, error) {
	result, err := r.queries.GetUserByID(ctx, uuid.MustParse(string(id)))
	if err != nil {
		return nil, mapPgError(err)
	}

	return sqlcUserToDomain(result)
}

// GetByEmail retrieves a user by their email address.
func (r *postgresRepository) GetByEmail(ctx context.Context, email string) (*user_entity.User, error) {
	result, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, mapPgError(err)
	}

	return sqlcUserToDomain(result)
}

// Update modifies an existing user in the data store.
func (r *postgresRepository) Update(ctx context.Context, u *user_entity.User) (*user_entity.User, error) {
	params := domainUserToUpdateParams(u)

	result, err := r.queries.UpdateUser(ctx, &params)
	if err != nil {
		return nil, mapPgError(err)
	}

	return sqlcUserToDomain(result)
}

// Delete removes a user from the data store.
func (r *postgresRepository) Delete(ctx context.Context, id user_entity.UserID) error {
	err := r.queries.DeleteUser(ctx, uuid.MustParse(string(id)))
	return mapPgError(err)
}

// List returns a paginated list of users matching the given parameters.
func (r *postgresRepository) List(ctx context.Context, params *user_entity.ListParams) (*user_entity.ListResult, error) {
	emailFilter := params.Email
	if emailFilter == "" {
		emailFilter = ""
	}

	var statusFilter string
	if params.Status != nil {
		statusFilter = string(*params.Status)
	}

	listParams := &user_sqlc.ListUsersParams{
		Column1: emailFilter,
		Column2: statusFilter,
		Limit:   params.Limit,
		Offset:  params.Offset,
	}

	users, err := r.queries.ListUsers(ctx, listParams)
	if err != nil {
		return nil, mapPgError(err)
	}

	countParams := &user_sqlc.CountUsersParams{
		Column1: emailFilter,
		Column2: statusFilter,
	}

	total, err := r.queries.CountUsers(ctx, countParams)
	if err != nil {
		return nil, mapPgError(err)
	}

	domainUsers := make([]*user_entity.User, 0, len(users))
	for _, u := range users {
		domainUser, err := sqlcUserToDomain(u)
		if err != nil {
			return nil, err
		}
		domainUsers = append(domainUsers, domainUser)
	}

	return &user_entity.ListResult{
		Users:  domainUsers,
		Total:  total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}, nil
}

// Exists checks if a user with the given email exists.
func (r *postgresRepository) Exists(ctx context.Context, email string) (bool, error) {
	exists, err := r.queries.ExistsUserByEmail(ctx, email)
	return exists, mapPgError(err)
}

// ExistsByID checks if a user with the given ID exists.
func (r *postgresRepository) ExistsByID(ctx context.Context, id user_entity.UserID) (bool, error) {
	exists, err := r.queries.ExistsUserByID(ctx, uuid.MustParse(string(id)))
	return exists, mapPgError(err)
}
