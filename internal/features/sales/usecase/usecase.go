package usecase

import (
	"context"
	"fmt"
	"time"
	"uuid"

	"github.com/zercle/zercle-go-template/internal/features/sales/contract"
	"github.com/zercle/zercle-go-template/internal/features/sales/domain"
	"github.com/zercle/zercle-go-template/internal/features/sales/repository"
)

const timeFormat = time.RFC3339

// Usecase implements the Service inbound use-case interface.
type Usecase struct {
	repo repository.Repository
}

// NewUsecase returns a Usecase backed by the provided repository.
func NewUsecase(repo repository.Repository) *Usecase {
	return &Usecase{repo: repo}
}

// Purchase validates a coin payment, composes change from the machine bank,
// records the sale atomically, and returns its wire form.
//
// Steps run cheapest-first and fail fast: the two ids are parsed before any
// repository call, and the coins are validated against the domain before a
// read, so a malformed request never reaches the database. domain.Purchase
// then decides insufficiency, stock, and exact-change against the prices and
// bank actually read, and CommitPurchase re-checks stock under a lock so a
// concurrent sale that drained the stock loses the race with ErrOutOfStock.
func (u *Usecase) Purchase(ctx context.Context, req *contract.PurchaseRequest) (*contract.PurchaseResponse, error) {
	if req == nil {
		return nil, domain.ErrInvalidID
	}
	machineID, err := uuid.Parse(req.MachineID)
	if err != nil {
		return nil, domain.ErrInvalidID
	}
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return nil, domain.ErrInvalidID
	}
	if err := domain.ValidateCoins(req.Coins); err != nil {
		return nil, fmt.Errorf("validate coins: %w", err)
	}

	product, err := u.repo.GetProduct(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("get product: %w", err)
	}
	bank, err := u.repo.GetMachineBank(ctx, machineID)
	if err != nil {
		return nil, fmt.Errorf("get machine bank: %w", err)
	}

	change, remaining, err := domain.Purchase(product.PriceCents, product.Stock, bank, req.Coins)
	if err != nil {
		return nil, fmt.Errorf("purchase: %w", err)
	}

	totalInserted := domain.SumCoins(req.Coins)
	record := &domain.PurchaseRecord{
		ID:                 uuid.New(),
		MachineID:          machineID,
		ProductID:          productID,
		PriceCents:         product.PriceCents,
		TotalInsertedCents: totalInserted,
		ChangeCents:        domain.SumCoins(change),
		ChangeCoins:        change,
		PurchasedAt:        time.Now().UTC(),
	}

	if err := u.repo.CommitPurchase(ctx, machineID, productID, record, remaining); err != nil {
		return nil, fmt.Errorf("commit purchase: %w", err)
	}

	resp := newPurchaseResponse(record)
	return &resp, nil
}

func newPurchaseResponse(record *domain.PurchaseRecord) contract.PurchaseResponse {
	if record == nil {
		return contract.PurchaseResponse{}
	}
	return contract.PurchaseResponse{
		ID:                 record.ID.String(),
		MachineID:          record.MachineID.String(),
		ProductID:          record.ProductID.String(),
		PriceCents:         record.PriceCents,
		TotalInsertedCents: record.TotalInsertedCents,
		ChangeCents:        record.ChangeCents,
		ChangeCoins:        record.ChangeCoins,
		PurchasedAt:        record.PurchasedAt.Format(timeFormat),
	}
}
