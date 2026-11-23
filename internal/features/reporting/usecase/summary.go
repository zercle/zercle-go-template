package usecase

import (
	"context"
	"fmt"

	"github.com/zercle/zercle-go-template/internal/features/reporting/domain"
	apiv1 "github.com/zercle/zercle-go-template/pkg/api/v1"
)

// Summary aggregates the cross-feature totals and the top machines into one
// report. A nil request or a non-positive top-machines value uses the
// configured default; a value above the configured maximum is rejected before
// any repository call so an unbounded leaderboard never reaches the database.
func (u *Usecase) Summary(ctx context.Context, req *apiv1.SummaryRequest) (*apiv1.SummaryResponse, error) {
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

func newSummaryResponse(overview domain.Overview, machines []domain.MachineSales) *apiv1.SummaryResponse {
	resp := &apiv1.SummaryResponse{
		Catalog: apiv1.CatalogStats{
			ProductCount: overview.ProductCount,
			TotalStock:   overview.TotalStock,
		},
		Machines: apiv1.MachineStats{
			MachineCount:       overview.MachineCount,
			TotalCoinBankCents: overview.TotalBankCents,
		},
		Sales: apiv1.SalesStats{
			PurchaseCount: overview.PurchaseCount,
			RevenueCents:  overview.RevenueCents,
		},
		TopMachines: make([]apiv1.MachineSales, len(machines)),
	}
	for i := range machines {
		resp.TopMachines[i] = apiv1.MachineSales{
			MachineID:     machines[i].MachineID.String(),
			Label:         machines[i].Label,
			PurchaseCount: machines[i].PurchaseCount,
			RevenueCents:  machines[i].RevenueCents,
		}
	}
	return resp
}
