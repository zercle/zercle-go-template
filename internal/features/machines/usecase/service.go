// Package usecase holds the machines feature's inbound use-case service and
// its implementation. The HTTP handler consumes the Service interface; the
// implementation orchestrates the domain and the outbound repository.
package usecase

import (
	"context"

	"github.com/zercle/zercle-go-template/internal/features/machines/domain"
)

// Service is the inbound use-case interface for machines. It speaks the
// feature's domain types at the boundary; the handler owns the wire mapping.
//
//go:generate go tool mockgen -source=service.go -destination=mock/service_mock.go -package=mock
type Service interface {
	Create(ctx context.Context, label string, initialCoins []int32) (*domain.Machine, error)
	Get(ctx context.Context, id string) (*domain.Machine, error)
	List(ctx context.Context, limit, offset int32) ([]domain.Machine, error)
	RestockBank(ctx context.Context, id string, coins []int32) (*domain.Machine, error)
}
