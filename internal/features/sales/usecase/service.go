// Package usecase holds the sales feature's inbound use-case service and its
// implementation. The HTTP handler consumes the Service interface; the
// implementation orchestrates the pure purchase domain rules and the outbound
// repository that commits the sale.
package usecase

import (
	"context"

	"github.com/zercle/zercle-go-template/internal/features/sales/contract"
)

// Service is the inbound use-case interface for sales. It speaks the feature's
// contract types at the boundary so the handler binds responses directly and
// never maps to or from domain entities.
//
//go:generate go tool mockgen -source=service.go -destination=mock/service_mock.go -package=mock
type Service interface {
	Purchase(ctx context.Context, req *contract.PurchaseRequest) (*contract.PurchaseResponse, error)
}
