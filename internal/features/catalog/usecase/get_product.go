package usecase

import (
	"context"
	"errors"
	"fmt"
	"uuid"

	"github.com/zercle/zercle-go-template/internal/features/catalog/domain"
)

// Get retrieves a product by ID, passing through domain.ErrProductNotFound. The
// id string is parsed here so both driving adapters share one validation path. A
// syntactically valid but absent id (including the nil UUID) is a not-found,
// which the port decides.
func (u *Usecase) Get(ctx context.Context, id string) (*domain.Product, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, domain.ErrInvalidID
	}
	product, err := u.repo.GetByID(ctx, parsed)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("get product: %w", err)
	}

	return product, nil
}
