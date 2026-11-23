// Package usecase holds the reporting feature's inbound use-case service and
// its implementation. The HTTP handler consumes the Service interface; the
// implementation orchestrates the outbound port that reads the shared
// catalog, machines, and sales tables.
package usecase

import (
	"context"

	"github.com/zercle/zercle-go-template/internal/features/reporting/domain"
)

// Service is the inbound use-case interface for reporting. It speaks domain
// types at the boundary; the handler owns the wire shapes and maps the domain
// result to the pkg/api/v1 response.
//
//go:generate go tool mockgen -source=service.go -destination=mock/service_mock.go -package=mock
type Service interface {
	Summary(ctx context.Context, topMachines int32) (*domain.Summary, error)
}
