package usecase

import (
	"context"
	"fmt"

	"github.com/zercle/zercle-go-template/internal/features/reporting/contract"
	"github.com/zercle/zercle-go-template/internal/features/reporting/domain"
	"github.com/zercle/zercle-go-template/internal/features/reporting/repository"
)

const (
	defaultTopMachinesFallback int32 = 5
	maxTopMachinesFallback     int32 = 20
)

// Usecase implements the Service inbound use-case interface.
type Usecase struct {
	repo               repository.Repository
	defaultTopMachines int32
	maxTopMachines     int32
}

// NewUsecase returns a Usecase backed by the provided repository. The limit
// arguments override the package fallback defaults; pass <= 0 to use the
// built-in defaults (5/20).
func NewUsecase(repo repository.Repository, defaultTopMachines, maxTopMachines int32) *Usecase {
	if defaultTopMachines <= 0 {
		defaultTopMachines = defaultTopMachinesFallback
	}
	if maxTopMachines <= 0 {
		maxTopMachines = maxTopMachinesFallback
	}
	return &Usecase{
		repo:               repo,
		defaultTopMachines: defaultTopMachines,
		maxTopMachines:     maxTopMachines,
	}
}

// Summary aggregates the cross-feature totals and the top machines into one
// report. A nil request or a non-positive top-machines value uses the
// configured default; a value above the configured maximum is rejected before
// any repository call so an unbounded leaderboard never reaches the database.
func (u *Usecase) Summary(ctx context.Context, req *contract.SummaryRequest) (*contract.SummaryResponse, error) {
	var top int32
	if req != nil {
		top = req.TopMachines
	}
	if top <= 0 {
		top = u.defaultTopMachines
	}
	if top > u.maxTopMachines {
		return nil, fmt.Errorf("%w: must be 1..%d", domain.ErrInvalidTopMachines, u.maxTopMachines)
	}

	overview, err := u.repo.GetOverview(ctx)
	if err != nil {
		return nil, fmt.Errorf("get overview: %w", err)
	}

	machines, err := u.repo.GetTopMachines(ctx, top)
	if err != nil {
		return nil, fmt.Errorf("get top machines: %w", err)
	}

	return newSummaryResponse(overview, machines), nil
}

func newSummaryResponse(overview domain.Overview, machines []domain.MachineSales) *contract.SummaryResponse {
	resp := &contract.SummaryResponse{
		Catalog: contract.CatalogStats{
			ProductCount: overview.ProductCount,
			TotalStock:   overview.TotalStock,
		},
		Machines: contract.MachineStats{
			MachineCount:       overview.MachineCount,
			TotalCoinBankCents: overview.TotalBankCents,
		},
		Sales: contract.SalesStats{
			PurchaseCount: overview.PurchaseCount,
			RevenueCents:  overview.RevenueCents,
		},
		TopMachines: make([]contract.MachineSales, len(machines)),
	}
	for i := range machines {
		resp.TopMachines[i] = contract.MachineSales{
			MachineID:     machines[i].MachineID.String(),
			Label:         machines[i].Label,
			PurchaseCount: machines[i].PurchaseCount,
			RevenueCents:  machines[i].RevenueCents,
		}
	}
	return resp
}
