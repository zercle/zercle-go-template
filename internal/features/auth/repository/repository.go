package domain

import (
	"context"

	"github.com/google/uuid"

	"github.com/zercle/zercle-go-template/internal/features/auth/domain"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

// UserReader defines read operations for User entities.
type UserReader interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
}

// UserWriter defines write operations for User entities.
type UserWriter interface {
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// UserRepository combines read and write operations for User entities.
type UserRepository interface {
	UserReader
	UserWriter
}

// SessionReader defines read operations for Session entities.
type SessionReader interface {
	FindByToken(ctx context.Context, token string) (*domain.Session, error)
}

// SessionWriter defines write operations for Session entities.
type SessionWriter interface {
	Create(ctx context.Context, session *domain.Session) error
	Delete(ctx context.Context, token string) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

// SessionRepository combines read and write operations for Session entities.
type SessionRepository interface {
	SessionReader
	SessionWriter
}
