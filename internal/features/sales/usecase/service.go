// Package usecase holds the sales feature's inbound use-case service and its
// implementation. The HTTP handler consumes the Service interface; the
// implementation orchestrates the pure purchase domain rules and the outbound
// port that commits the sale.
package usecase

import (
	"context"

	"github.com/zercle/zercle-go-template/internal/features/sales/domain"
)

// Service is the inbound use-case interface for sales. It speaks domain types
// at the boundary; the handler owns the wire mapping.
//
//go:generate go tool mockgen -source=service.go -destination=mock/service_mock.go -package=mock
type Service interface {
	Purchase(ctx context.Context, machineID, productID string, coins []int32) (*domain.PurchaseRecord, error)
}
