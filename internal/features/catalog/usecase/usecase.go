// Package usecase holds the catalog feature's inbound use-case service and
// its implementation. The HTTP handler consumes the Service interface; the
// implementation orchestrates the domain and the outbound port.
//
// Each operation lives in its own file: create_product.go, get_product.go,
// list_products.go. This file holds the shared implementation type, its
// constructor, and the domain-to-wire mapping used by every operation.
package usecase

import (
	"time"

	"github.com/zercle/zercle-go-template/internal/features/catalog/domain"
	"github.com/zercle/zercle-go-template/internal/features/catalog/port"
	apiv1 "github.com/zercle/zercle-go-template/pkg/api/v1"
)

const (
	defaultPageSizeFallback int32 = 20
	maxPageSizeFallback     int32 = 100
	maxNameLengthFallback   int32 = 255

	timeFormat = time.RFC3339
)

// Usecase implements the Service inbound use-case interface.
type Usecase struct {
	repo            port.Repository
	defaultPageSize int32
	maxPageSize     int32
	maxNameLength   int32
}

// NewUsecase returns a Usecase backed by the provided port. The limit
// arguments override the package fallback defaults; pass <= 0 to use the
// built-in defaults (20/100/255).
func NewUsecase(repo port.Repository, defaultPageSize, maxPageSize, maxNameLength int32) *Usecase {
	if defaultPageSize <= 0 {
		defaultPageSize = defaultPageSizeFallback
	}
	if maxPageSize <= 0 {
		maxPageSize = maxPageSizeFallback
	}
	if maxNameLength <= 0 {
		maxNameLength = maxNameLengthFallback
	}
	return &Usecase{
		repo:            repo,
		defaultPageSize: defaultPageSize,
		maxPageSize:     maxPageSize,
		maxNameLength:   maxNameLength,
	}
}

func newProductResponse(product *domain.Product) apiv1.ProductResponse {
	if product == nil {
		return apiv1.ProductResponse{}
	}
	return apiv1.ProductResponse{
		ID:         product.ID.String(),
		Name:       product.Name,
		PriceCents: product.PriceCents,
		Stock:      product.Stock,
		CreatedAt:  product.CreatedAt.Format(timeFormat),
		UpdatedAt:  product.UpdatedAt.Format(timeFormat),
	}
}
