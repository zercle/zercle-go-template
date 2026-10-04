package domain

import (
	"time"
	"uuid"
)

// Product is a sellable item held in the global catalog pool. Prices are
// integer cents so the coin arithmetic on the payment path is exact; Stock is
// the number of units still available across every machine.
type Product struct {
	ID         uuid.UUID
	Name       string
	PriceCents int32
	Stock      int32
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
