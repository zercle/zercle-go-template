package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/internal/core/domain"
)

// UserRepository defines the output port for User persistence.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// PostRepository defines the output port for Post persistence.
type PostRepository interface {
	Create(ctx context.Context, post *domain.Post) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Post, error)
	GetAll(ctx context.Context) ([]*domain.Post, error)
	Update(ctx context.Context, post *domain.Post) error
	Delete(ctx context.Context, id uuid.UUID) error
}
