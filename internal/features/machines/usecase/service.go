// Package usecase holds the machines feature's inbound use-case service and
// its implementation. The HTTP handler consumes the Service interface; the
// implementation orchestrates the domain and the outbound repository.
package usecase

import (
	"context"

	"github.com/zercle/zercle-go-template/internal/features/machines/contract"
)

// Service is the inbound use-case interface for machines. It speaks the
// feature's contract types at the boundary so the handler binds responses directly
// and never maps to or from domain entities.
//
//go:generate go tool mockgen -source=service.go -destination=mock/service_mock.go -package=mock
type Service interface {
	Create(ctx context.Context, req *contract.CreateMachineRequest) (*contract.MachineResponse, error)
	Get(ctx context.Context, id string) (*contract.MachineResponse, error)
	List(ctx context.Context, req *contract.ListMachinesRequest) (*contract.ListMachinesResponse, error)
	RestockBank(ctx context.Context, id string, req *contract.RestockBankRequest) (*contract.MachineResponse, error)
}
