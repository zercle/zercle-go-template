package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
	"uuid"

	"github.com/zercle/zercle-go-template/internal/features/catalog/domain"
)

// Create validates the name and price, persists a new product, and returns it
// as a domain value. The name is trimmed first so surrounding whitespace never
// counts toward the configured length limit.
func (u *Usecase) Create(ctx context.Context, name string, priceCents, stock int32) (*domain.Product, error) {
	name = strings.TrimSpace(name)
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

	return product, nil
}
