package user

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	sqlcuser "github.com/zercle/zercle-go-template/internal/infrastructure/database/sqlc/user"
)

// Repository defines the interface for user storage.
type Repository interface {
	Create(ctx context.Context, user *User) (*User, error)
	GetByID(ctx context.Context, id ID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) (*User, error)
	Delete(ctx context.Context, id ID) error
	List(ctx context.Context, params *ListParams) (*ListResult, error)
	Exists(ctx context.Context, email string) (bool, error)
	ExistsByID(ctx context.Context, id ID) (bool, error)
}

type postgresRepository struct {
	queries *sqlcuser.Queries
}

// NewPostgresRepository creates a new Postgres user repository.
func NewPostgresRepository(db sqlcuser.DBTX) Repository {
	return &postgresRepository{
		queries: sqlcuser.New(db),
	}
}

func mapPgError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrDuplicateEmail
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrUserNotFound
	}

	return err
}

func (r *postgresRepository) Create(ctx context.Context, user *User) (*User, error) {
	params := &sqlcuser.CreateUserParams{
		ID:        uuid.MustParse(string(user.ID)),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Status:    string(user.Status),
		CreatedAt: toPgTimestamptz(user.CreatedAt),
		UpdatedAt: toPgTimestamptz(user.UpdatedAt),
	}
	result, err := r.queries.CreateUser(ctx, params)
	if err != nil {
		return nil, mapPgError(err)
	}
	return sqlcUserToDomain(result)
}

func (r *postgresRepository) GetByID(ctx context.Context, id ID) (*User, error) {
	uid, err := uuid.Parse(string(id))
	if err != nil {
		return nil, err
	}
	result, err := r.queries.GetUserByID(ctx, uid)
	if err != nil {
		return nil, mapPgError(err)
	}
	return sqlcUserToDomain(result)
}

func (r *postgresRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	result, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, mapPgError(err)
	}
	return sqlcUserToDomain(result)
}

func (r *postgresRepository) Update(ctx context.Context, user *User) (*User, error) {
	params := &sqlcuser.UpdateUserParams{
		ID:        uuid.MustParse(string(user.ID)),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Status:    string(user.Status),
		UpdatedAt: toPgTimestamptz(user.UpdatedAt),
	}
	result, err := r.queries.UpdateUser(ctx, params)
	if err != nil {
		return nil, mapPgError(err)
	}
	return sqlcUserToDomain(result)
}

func (r *postgresRepository) Delete(ctx context.Context, id ID) error {
	uid, err := uuid.Parse(string(id))
	if err != nil {
		return err
	}
	err = r.queries.DeleteUser(ctx, uid)
	return mapPgError(err)
}

func (r *postgresRepository) List(ctx context.Context, params *ListParams) (*ListResult, error) {
	listParams := &sqlcuser.ListUsersParams{
		Column1: "",
		Column2: "",
		Limit:   params.Limit,
		Offset:  params.Offset,
	}

	if params.Email != "" {
		listParams.Column1 = params.Email
	}
	if params.Status != nil {
		listParams.Column2 = *params.Status
	}

	users, err := r.queries.ListUsers(ctx, listParams)
	if err != nil {
		return nil, mapPgError(err)
	}

	countParams := &sqlcuser.CountUsersParams{
		Column1: listParams.Column1,
		Column2: listParams.Column2,
	}
	total, err := r.queries.CountUsers(ctx, countParams)
	if err != nil {
		return nil, mapPgError(err)
	}

	domainUsers := make([]*User, 0, len(users))
	for _, u := range users {
		domainUser, err := sqlcUserToDomain(u)
		if err != nil {
			return nil, err
		}
		domainUsers = append(domainUsers, domainUser)
	}

	return &ListResult{
		Users:  domainUsers,
		Total:  total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}, nil
}

func (r *postgresRepository) Exists(ctx context.Context, email string) (bool, error) {
	return r.queries.ExistsUserByEmail(ctx, email)
}

func (r *postgresRepository) ExistsByID(ctx context.Context, id ID) (bool, error) {
	uid, err := uuid.Parse(string(id))
	if err != nil {
		return false, err
	}
	return r.queries.ExistsUserByID(ctx, uid)
}

func sqlcUserToDomain(m *sqlcuser.User) (*User, error) {
	createdAt := time.Time{}
	if m.CreatedAt.Valid {
		createdAt = m.CreatedAt.Time
	}
	updatedAt := time.Time{}
	if m.UpdatedAt.Valid {
		updatedAt = m.UpdatedAt.Time
	}

	return &User{
		ID:        ID(m.ID.String()),
		Email:     m.Email,
		FirstName: m.FirstName,
		LastName:  m.LastName,
		Status:    Status(m.Status),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func toPgTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  t,
		Valid: true,
	}
}
