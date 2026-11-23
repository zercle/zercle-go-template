package handler

import (
	"time"

	"github.com/zercle/zercle-go-template/internal/features/catalog/domain"
	apiv1 "github.com/zercle/zercle-go-template/pkg/api/v1"
)

const timeFormat = time.RFC3339

func toProductResponse(product *domain.Product) apiv1.ProductResponse {
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

func toProductResponseList(products []domain.Product) apiv1.ListProductsResponse {
	resp := apiv1.ListProductsResponse{Products: make([]apiv1.ProductResponse, len(products))}
	for i := range products {
		resp.Products[i] = toProductResponse(&products[i])
	}
	return resp
}
