package port

import (
	"context"

	"github.com/google/uuid"
	postDomain "github.com/zercle/zercle-go-template/internal/features/post/domain"
)

//go:generate mockgen -destination=./mocks/$GOFILE -package=mocks -source=$GOFILE

// PostRepository defines the output port for Post persistence.
type PostRepository interface {
	Create(ctx context.Context, post *postDomain.Post) error
	GetByID(ctx context.Context, id uuid.UUID) (*postDomain.Post, error)
	GetAll(ctx context.Context) ([]*postDomain.Post, error)
	Update(ctx context.Context, post *postDomain.Post) error
	Delete(ctx context.Context, id uuid.UUID) error
}
