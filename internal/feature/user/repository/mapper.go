package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	user_entity "github.com/zercle/zercle-go-template/internal/feature/user"
	user_sqlc "github.com/zercle/zercle-go-template/internal/infrastructure/database/sqlc/user"
)

// sqlcUserToDomain converts a SQLC User model to a domain User entity.
func sqlcUserToDomain(m *user_sqlc.User) (*user_entity.User, error) {
	createdAt := time.Time{}
	if m.CreatedAt.Valid {
		createdAt = m.CreatedAt.Time
	}
	updatedAt := time.Time{}
	if m.UpdatedAt.Valid {
		updatedAt = m.UpdatedAt.Time
	}

	return user_entity.NewUserWithID(
		user_entity.UserID(m.ID.String()),
		m.Email,
		m.PasswordHash,
		m.FirstName,
		m.LastName,
		user_entity.UserStatus(m.Status),
		createdAt,
		updatedAt,
	)
}

// domainUserToCreateParams converts a domain User to SQLC CreateUserParams.
func domainUserToCreateParams(u *user_entity.User) user_sqlc.CreateUserParams {
	return user_sqlc.CreateUserParams{
		ID:           uuid.MustParse(string(u.ID)),
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Status:       string(u.Status),
		CreatedAt:    pgtypeTimestamptz(u.CreatedAt),
		UpdatedAt:    pgtypeTimestamptz(u.UpdatedAt),
	}
}

// domainUserToUpdateParams converts a domain User to SQLC UpdateUserParams.
func domainUserToUpdateParams(u *user_entity.User) user_sqlc.UpdateUserParams {
	return user_sqlc.UpdateUserParams{
		ID:           uuid.MustParse(string(u.ID)),
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Status:       string(u.Status),
		UpdatedAt:    pgtypeTimestamptz(u.UpdatedAt),
	}
}

// pgtypeTimestamptz converts a time.Time to pgtype.Timestamptz.
func pgtypeTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  t,
		Valid: true,
	}
}
