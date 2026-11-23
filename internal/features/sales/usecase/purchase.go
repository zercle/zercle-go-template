package usecase

import (
	"context"
	"fmt"
	"time"
	"uuid"

	"github.com/zercle/zercle-go-template/internal/features/sales/domain"
)

// Purchase validates a coin payment, composes change from the machine bank,
// records the sale atomically, and returns the durable purchase record.
//
// Steps run cheapest-first and fail fast: the two ids are parsed before any
// repository call, and the coins are validated against the domain before a
// read, so a malformed request never reaches the database. domain.Purchase
// then decides insufficiency, stock, and exact-change against the prices and
// bank actually read, and CommitPurchase re-checks stock under a lock so a
// concurrent sale that drained the stock loses the race with ErrOutOfStock.
func (u *Usecase) Purchase(ctx context.Context, machineID, productID string, coins []int32) (*domain.PurchaseRecord, error) {
	machine, err := uuid.Parse(machineID)
	if err != nil {
		return nil, domain.ErrInvalidID
	}
	product, err := uuid.Parse(productID)
	if err != nil {
		return nil, domain.ErrInvalidID
	}
	if err := domain.ValidateCoins(coins); err != nil {
		return nil, fmt.Errorf("validate coins: %w", err)
	}

	saleProduct, err := u.repo.GetProduct(ctx, product)
	if err != nil {
		return nil, fmt.Errorf("get product: %w", err)
	}
	bank, err := u.repo.GetMachineBank(ctx, machine)
	if err != nil {
		return nil, fmt.Errorf("get machine bank: %w", err)
	}

	change, remaining, err := domain.Purchase(saleProduct.PriceCents, saleProduct.Stock, bank, coins)
	if err != nil {
		return nil, fmt.Errorf("purchase: %w", err)
	}

	record := &domain.PurchaseRecord{
		ID:                 uuid.New(),
		MachineID:          machine,
		ProductID:          product,
		PriceCents:         saleProduct.PriceCents,
		TotalInsertedCents: domain.SumCoins(coins),
		ChangeCents:        domain.SumCoins(change),
		ChangeCoins:        change,
		PurchasedAt:        time.Now().UTC(),
	}

	if err := u.repo.CommitPurchase(ctx, machine, product, record, remaining); err != nil {
		return nil, fmt.Errorf("commit purchase: %w", err)
	}

	return record, nil
}
