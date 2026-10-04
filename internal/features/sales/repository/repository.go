// Package repository declares the sales feature's outbound (driven) interface.
// The usecase layer consumes it; the gorm implementation under
// repository/postgres satisfies it structurally.
package repository

import (
	"context"
	"uuid"

	"github.com/zercle/zercle-go-template/internal/features/sales/domain"
)

// Repository is the outbound interface for sales orchestration. It reads the
// catalog and machines tables through its own port (see models.refs) and writes
// the purchase and its effects in one transaction.
//
//go:generate go tool mockgen -source=repository.go -destination=mock/repository_mock.go -package=mock
type Repository interface {
	// GetProduct reads the catalog projection for a product. A missing product
	// maps to domain.ErrProductNotFound.
	GetProduct(ctx context.Context, productID uuid.UUID) (domain.SaleProduct, error)
	// GetMachineBank reads a machine's current coin bank. A missing machine maps
	// to domain.ErrMachineNotFound.
	GetMachineBank(ctx context.Context, machineID uuid.UUID) (domain.CoinBank, error)
	// CommitPurchase records a completed sale atomically: it locks and
	// decrements the product's stock (guarding the read/commit race), replaces
	// the machine's coin bank with bankAfter, and inserts the purchase record.
	// A missing product maps to domain.ErrProductNotFound, losing the stock race
	// to domain.ErrOutOfStock, and a missing machine to domain.ErrMachineNotFound.
	CommitPurchase(ctx context.Context, machineID, productID uuid.UUID, record *domain.PurchaseRecord, bankAfter domain.CoinBank) error
}
