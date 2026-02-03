package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/internal/features/auth/domain"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

type UserReader interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
}

type UserWriter interface {
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type UserRepository interface {
	UserReader
	UserWriter
}

type SessionReader interface {
	FindByToken(ctx context.Context, token string) (*domain.Session, error)
}

type SessionWriter interface {
	Create(ctx context.Context, session *domain.Session) error
	Delete(ctx context.Context, token string) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

type SessionRepository interface {
	SessionReader
	SessionWriter
}
