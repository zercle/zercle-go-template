package usecase

import (
	"context"
	"fmt"

	"github.com/zercle/zercle-go-template/internal/features/reporting/domain"
)

// Summary aggregates the cross-feature totals and the top machines into one
// report. A non-positive top-machines value uses the configured default; a
// value above the configured maximum is rejected before any repository call
// so an unbounded leaderboard never reaches the database.
func (u *Usecase) Summary(ctx context.Context, topMachines int32) (*domain.Summary, error) {
	top := topMachines
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

	return &domain.Summary{Overview: overview, TopMachines: machines}, nil
}
