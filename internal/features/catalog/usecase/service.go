// Package usecase holds the catalog feature's inbound use-case service and
// its implementation. It speaks the feature's domain types at the boundary:
// the handler owns the wire shapes and maps domain values to the published
// contract. The implementation orchestrates the domain and the outbound port.
package usecase

import (
	"context"

	"github.com/zercle/zercle-go-template/internal/features/catalog/domain"
)

// Service is the inbound use-case interface for products. It takes and
// returns the feature's domain types so the usecase never depends on the
// published wire contract; the driving handler translates to and from it.
//
//go:generate go tool mockgen -source=service.go -destination=mock/service_mock.go -package=mock
type Service interface {
	Create(ctx context.Context, name string, priceCents, stock int32) (*domain.Product, error)
	Get(ctx context.Context, id string) (*domain.Product, error)
	List(ctx context.Context, limit, offset int32) ([]domain.Product, error)
}
