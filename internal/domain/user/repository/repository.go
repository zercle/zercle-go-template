package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zercle/zercle-go-template/internal/domain/user"
	"github.com/zercle/zercle-go-template/internal/domain/user/entity"
	"github.com/zercle/zercle-go-template/internal/infrastructure/logger"
	"github.com/zercle/zercle-go-template/internal/infrastructure/sqlc/db"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrEmailExists  = errors.New("email already exists")
)

type userRepository struct {
	sqlc *db.Queries
	log  *logger.Logger
}

func NewUserRepository(sqlc *db.Queries, log *logger.Logger) user.UserRepository {
	return &userRepository{
		sqlc: sqlc,
		log:  log,
	}
}

func toUUID(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: u, Valid: true}
}

func fromUUID(u pgtype.UUID) uuid.UUID {
	return u.Bytes
}

func toText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

func fromText(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

func toTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func fromTimestamptz(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

func toInt32Safe(i int) int32 {
	if i < math.MinInt32 || i > math.MaxInt32 {
		panic(fmt.Sprintf("value %d overflows int32", i))
	}
	return int32(i)
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	now := time.Now()
	// Generate a new UUID for the user
	newID := uuid.New()

	params := db.CreateUserParams{
		ID:           toUUID(newID),
		Email:        user.Email,
		PasswordHash: user.Password,
		FullName:     toText(user.FullName),
		Phone:        toText(user.Phone),
		CreatedAt:    toTimestamptz(now),
		UpdatedAt:    toTimestamptz(now),
	}

	row, err := r.sqlc.CreateUser(ctx, params)
	if err != nil {
		r.log.Error("Failed to create user", "error", err, "email", user.Email)
		return nil, err
	}

	return &entity.User{
		ID:        fromUUID(row.ID),
		Email:     row.Email,
		Password:  row.PasswordHash,
		FullName:  fromText(row.FullName),
		Phone:     fromText(row.Phone),
		CreatedAt: fromTimestamptz(row.CreatedAt),
		UpdatedAt: fromTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	row, err := r.sqlc.GetUser(ctx, toUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		r.log.Error("Failed to get user by ID", "error", err, "user_id", id)
		return nil, err
	}

	return &entity.User{
		ID:        fromUUID(row.ID),
		Email:     row.Email,
		Password:  row.PasswordHash,
		FullName:  fromText(row.FullName),
		Phone:     fromText(row.Phone),
		CreatedAt: fromTimestamptz(row.CreatedAt),
		UpdatedAt: fromTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	row, err := r.sqlc.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		r.log.Error("Failed to get user by email", "error", err, "email", email)
		return nil, err
	}

	return &entity.User{
		ID:        fromUUID(row.ID),
		Email:     row.Email,
		Password:  row.PasswordHash,
		FullName:  fromText(row.FullName),
		Phone:     fromText(row.Phone),
		CreatedAt: fromTimestamptz(row.CreatedAt),
		UpdatedAt: fromTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	params := db.UpdateUserParams{
		ID:        toUUID(user.ID),
		UpdatedAt: toTimestamptz(time.Now()),
		FullName:  toText(user.FullName),
		Phone:     toText(user.Phone),
	}

	row, err := r.sqlc.UpdateUser(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		r.log.Error("Failed to update user", "error", err, "user_id", user.ID)
		return nil, err
	}

	return &entity.User{
		ID:        fromUUID(row.ID),
		Email:     row.Email,
		Password:  row.PasswordHash,
		FullName:  fromText(row.FullName),
		Phone:     fromText(row.Phone),
		CreatedAt: fromTimestamptz(row.CreatedAt),
		UpdatedAt: fromTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.sqlc.DeleteUser(ctx, toUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		r.log.Error("Failed to delete user", "error", err, "user_id", id)
		return err
	}
	return nil
}

func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*entity.User, int, error) {
	params := db.ListUsersParams{
		Limit:  toInt32Safe(limit),
		Offset: toInt32Safe(offset),
	}

	rows, err := r.sqlc.ListUsers(ctx, params)
	if err != nil {
		r.log.Error("Failed to list users", "error", err)
		return nil, 0, err
	}

	users := make([]*entity.User, len(rows))
	for i, row := range rows {
		users[i] = &entity.User{
			ID:        fromUUID(row.ID),
			Email:     row.Email,
			Password:  row.PasswordHash,
			FullName:  fromText(row.FullName),
			Phone:     fromText(row.Phone),
			CreatedAt: fromTimestamptz(row.CreatedAt),
			UpdatedAt: fromTimestamptz(row.UpdatedAt),
		}
	}

	return users, len(users), nil
}
