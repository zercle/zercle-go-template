// Package usecase holds the reporting feature's inbound use-case service and
// its implementation. The HTTP handler consumes the Service interface; the
// implementation orchestrates the outbound repository that reads the shared
// catalog, machines, and sales tables.
package usecase

import (
	"context"

	"github.com/zercle/zercle-go-template/internal/features/reporting/contract"
)

// Service is the inbound use-case interface for reporting. It speaks the
// feature's contract types at the boundary so the handler binds responses
// directly and never maps to or from domain entities.
//
//go:generate go tool mockgen -source=service.go -destination=mock/service_mock.go -package=mock
type Service interface {
	Summary(ctx context.Context, req *contract.SummaryRequest) (*contract.SummaryResponse, error)
}
