// Package usecase holds the reporting feature's inbound use-case service and
// its implementation. The HTTP handler consumes the Service interface; the
// implementation orchestrates the outbound repository that reads the shared
// catalog, machines, and sales tables.
package usecase

import (
	"context"

	apiv1 "github.com/zercle/zercle-go-template/pkg/api/v1"
)

// Service is the inbound use-case interface for reporting. It speaks the
// feature's contract types at the boundary so the handler binds responses
// directly and never maps to or from domain entities.
//
//go:generate go tool mockgen -source=service.go -destination=mock/service_mock.go -package=mock
type Service interface {
	Summary(ctx context.Context, req *apiv1.SummaryRequest) (*apiv1.SummaryResponse, error)
}
