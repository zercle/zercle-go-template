// STUB FEATURE — delete internal/features/example to start your project.

package domain

import (
	"context"

	"github.com/google/uuid"
)

// Service is the inbound use-case port for Items.
//
//go:generate go run go.uber.org/mock/mockgen -source=service.go -destination=../service/mock/service_mock.go -package=mock
type Service interface {
	Create(ctx context.Context, name string) (*Item, error)
	Get(ctx context.Context, id uuid.UUID) (*Item, error)
	List(ctx context.Context, limit, offset int32) ([]Item, error)
}
