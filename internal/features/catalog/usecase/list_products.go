package usecase

import (
	"context"
	"fmt"

	"github.com/zercle/zercle-go-template/internal/features/catalog/contract"
)

// List returns a paginated list of products. It enforces safe defaults so a
// zero-value limit (e.g. no query parameter) never produces LIMIT 0 and a
// negative offset never reaches the query.
//
// Pagination is offset-based, so a concurrent insert between two page requests
// can shift rows across a page boundary (a page may repeat or skip a product).
// Keyset/cursor pagination on (created_at, id) would be stable but changes the
// published request contract, so it is left as a documented limitation here.
func (u *Usecase) List(ctx context.Context, req *contract.ListProductsRequest) (*contract.ListProductsResponse, error) {
	var limit, offset int32
	if req != nil {
		limit, offset = req.Limit, req.Offset
	}
	if limit <= 0 {
		limit = u.defaultPageSize
	}
	if limit > u.maxPageSize {
		limit = u.maxPageSize
	}
	if offset < 0 {
		offset = 0
	}

	products, err := u.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}

	resp := &contract.ListProductsResponse{Products: make([]contract.ProductResponse, len(products))}
	for i := range products {
		resp.Products[i] = newProductResponse(&products[i])
	}
	return resp, nil
}
