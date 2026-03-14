//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -destination=../mocks/repository.mock.go -package=mocks

package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/zercle/zercle-go-template/internal/feature/auth/domain"
)

// UserReader defines methods for reading user data.
type UserReader interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
}

// UserWriter defines methods for writing user data.
type UserWriter interface {
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// UserRepository combines user reading and writing operations.
type UserRepository interface {
	UserReader
	UserWriter
}

// SessionReader defines methods for reading session data.
type SessionReader interface {
	FindByToken(ctx context.Context, token string) (*domain.Session, error)
}

// SessionWriter defines methods for writing session data.
type SessionWriter interface {
	Create(ctx context.Context, session *domain.Session) error
	Delete(ctx context.Context, token string) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

// SessionRepository combines session reading and writing operations.
type SessionRepository interface {
	SessionReader
	SessionWriter
}
