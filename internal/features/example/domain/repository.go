// STUB FEATURE — delete internal/features/example to start your project.

package domain

import (
	"context"

	"github.com/google/uuid"
)

// Repository is the outbound port for Item persistence.
//
//go:generate go tool mockgen -source=repository.go -destination=../repository/mock/repository_mock.go -package=mock
type Repository interface {
	Create(ctx context.Context, item *Item) error
	GetByID(ctx context.Context, id uuid.UUID) (*Item, error)
	List(ctx context.Context, limit, offset int32) ([]Item, error)
}
