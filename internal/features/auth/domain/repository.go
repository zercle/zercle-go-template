package domain

import (
	"context"

	"github.com/google/uuid"
)

type UserReader interface {
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
}

type UserWriter interface {
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type UserRepository interface {
	UserReader
	UserWriter
}

type SessionReader interface {
	FindByToken(ctx context.Context, token string) (*Session, error)
}

type SessionWriter interface {
	Create(ctx context.Context, session *Session) error
	Delete(ctx context.Context, token string) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

type SessionRepository interface {
	SessionReader
	SessionWriter
}
