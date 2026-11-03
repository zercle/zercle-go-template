package usecase

import (
	"context"
	"errors"
	"fmt"
	"uuid"

	"github.com/zercle/zercle-go-template/internal/features/catalog/contract"
	"github.com/zercle/zercle-go-template/internal/features/catalog/domain"
)

// Get retrieves a product by ID, passing through domain.ErrProductNotFound. The
// wire id string is parsed here so both driving adapters share one validation
// path. A syntactically valid but absent id (including the nil UUID) is a
// not-found, which the repository decides.
func (u *Usecase) Get(ctx context.Context, id string) (*contract.ProductResponse, error) {
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

	resp := newProductResponse(product)
	return &resp, nil
}
