// Package usecase holds the catalog feature's inbound use-case service and
// its implementation. The HTTP handler consumes the Service interface; the
// implementation orchestrates the domain and the outbound repository.
package usecase

import (
	"context"

	apiv1 "github.com/zercle/zercle-go-template/pkg/api/v1"
)

// Service is the inbound use-case interface for products. It speaks the
// feature's contract types at the boundary so the handler binds responses directly
// and never maps to or from domain entities.
//
//go:generate go tool mockgen -source=service.go -destination=mock/service_mock.go -package=mock
type Service interface {
	Create(ctx context.Context, req *apiv1.CreateProductRequest) (*apiv1.ProductResponse, error)
	Get(ctx context.Context, id string) (*apiv1.ProductResponse, error)
	List(ctx context.Context, req *apiv1.ListProductsRequest) (*apiv1.ListProductsResponse, error)
}
