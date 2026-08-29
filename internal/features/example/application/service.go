// STUB FEATURE — delete internal/features/example to start your project.

// Package application holds the example feature's inbound use-case port and
// its implementation. The driving adapters under adapter/in consume the
// Service interface; the implementation orchestrates the domain and the
// outbound ports.
package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/zercle/zercle-go-template/internal/features/example/domain"
)

// Service is the inbound use-case port for Items.
//
//go:generate go tool mockgen -source=service.go -destination=mock/service_mock.go -package=mock
type Service interface {
	Create(ctx context.Context, name string) (*domain.Item, error)
	Get(ctx context.Context, id uuid.UUID) (*domain.Item, error)
	List(ctx context.Context, limit, offset int32) ([]domain.Item, error)
}
