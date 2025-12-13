package port

import (
	"context"

	"github.com/google/uuid"
	userDomain "github.com/zercle/zercle-go-template/internal/features/user/domain"
)

//go:generate mockgen -destination=./mocks/$GOFILE -package=mocks -source=$GOFILE

// UserRepository defines the output port for User persistence.
type UserRepository interface {
	Create(ctx context.Context, user *userDomain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*userDomain.User, error)
	GetByEmail(ctx context.Context, email string) (*userDomain.User, error)
	Update(ctx context.Context, user *userDomain.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}
