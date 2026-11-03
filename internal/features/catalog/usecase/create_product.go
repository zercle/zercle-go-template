package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
	"uuid"

	"github.com/zercle/zercle-go-template/internal/features/catalog/contract"
	"github.com/zercle/zercle-go-template/internal/features/catalog/domain"
)

// Create validates the name and price, persists a new product, and returns its
// wire form. The name is trimmed first so surrounding whitespace never counts
// toward the configured length limit.
func (u *Usecase) Create(ctx context.Context, req *contract.CreateProductRequest) (*contract.ProductResponse, error) {
	var name string
	var priceCents, stock int32
	if req != nil {
		name = strings.TrimSpace(req.Name)
		priceCents = req.PriceCents
		stock = req.Stock
	}
	if name == "" || utf8.RuneCountInString(name) > int(u.maxNameLength) {
		return nil, fmt.Errorf("%w: must be 1..%d characters", domain.ErrInvalidProductName, u.maxNameLength)
	}
	if priceCents <= 0 {
		return nil, domain.ErrInvalidPrice
	}

	now := time.Now().UTC()
	product := &domain.Product{
		ID:         uuid.New(),
		Name:       name,
		PriceCents: priceCents,
		Stock:      stock,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := u.repo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("create product: %w", err)
	}

	resp := newProductResponse(product)
	return &resp, nil
}
