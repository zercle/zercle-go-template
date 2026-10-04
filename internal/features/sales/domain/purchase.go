package domain

import (
	"fmt"
	"time"
	"uuid"
)

// PurchaseRecord is the durable outcome of a purchase.
type PurchaseRecord struct {
	ID                 uuid.UUID
	MachineID          uuid.UUID
	ProductID          uuid.UUID
	PriceCents         int32
	TotalInsertedCents int32
	ChangeCents        int32
	ChangeCoins        []int32
	PurchasedAt        time.Time
}

// SaleProduct is the catalog projection the sales feature needs to price and
// guard a sale. It carries no catalog behavior, only the fields a purchase
// decision reads.
type SaleProduct struct {
	ProductID  uuid.UUID
	PriceCents int32
	Stock      int32
}

// Purchase validates a coin payment against a product price and composes change
// from the machine bank. It is pure: neither the price/stock inputs nor the
// bank are mutated; the caller is responsible for persisting the sale and its
// coin effects. priceCents is guaranteed positive by the catalog schema's
// CHECK (price_cents > 0) constraint, so it is not re-validated here.
//
// Coins are validated before totals so an unsupported coin is rejected even
// when the payment would otherwise be sufficient.
func Purchase(priceCents, stock int32, bank CoinBank, coins []int32) (change []int32, remaining CoinBank, err error) {
	if err := ValidateCoins(coins); err != nil {
		return nil, nil, err
	}
	total := SumCoins(coins)
	if total < priceCents {
		return nil, nil, fmt.Errorf("%w: inserted %d cents, price %d cents", ErrInsufficientPayment, total, priceCents)
	}
	if stock <= 0 {
		return nil, nil, ErrOutOfStock
	}
	changeAmt := total - priceCents
	if changeAmt == 0 {
		return []int32{}, bank.Clone(), nil
	}
	return MakeChange(bank, changeAmt)
}
