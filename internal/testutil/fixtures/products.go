// Package fixtures provides sample domain objects for tests.
package fixtures

import (
	"time"
	"uuid"

	"github.com/zercle/zercle-go-template/internal/features/catalog/domain"
)

// NewProduct returns a sample catalog Product with the given name, price, and
// stock. It uses a deterministic generated UUID for the ID and fixed
// timestamps so tests can assert against known values.
func NewProduct(name string, priceCents, stock int32) domain.Product {
	return domain.Product{
		ID:         uuid.MustParse("22345678-1234-1234-1234-123456789abc"),
		Name:       name,
		PriceCents: priceCents,
		Stock:      stock,
		CreatedAt:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}
