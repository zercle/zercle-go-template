// STUB FEATURE — delete internal/features/example to start your project.

// Package usecase holds the example feature's inbound use-case service and
// its implementation. The HTTP handler consumes the Service interface; the
// implementation orchestrates the domain and the outbound repository.
package usecase

import (
	"context"

	"github.com/zercle/zercle-go-template/internal/features/example/contract"
)

// Service is the inbound use-case interface for Items. It speaks the
// feature's contract types at the boundary so the handler binds responses directly
// and never map to or from domain entities.
//
//go:generate go tool mockgen -source=service.go -destination=mock/service_mock.go -package=mock
type Service interface {
	Create(ctx context.Context, req *contract.CreateItemRequest) (*contract.ItemResponse, error)
	Get(ctx context.Context, id string) (*contract.ItemResponse, error)
	List(ctx context.Context, req *contract.ListItemsRequest) (*contract.ListItemsResponse, error)
}
