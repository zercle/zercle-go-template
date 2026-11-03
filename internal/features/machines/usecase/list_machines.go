package usecase

import (
	"context"
	"fmt"

	"github.com/zercle/zercle-go-template/internal/features/machines/contract"
)

// List returns a paginated list of machines. It enforces safe defaults so a
// zero-value limit (e.g. no query parameter) never produces LIMIT 0 and a
// negative offset never reaches the query.
func (u *Usecase) List(ctx context.Context, req *contract.ListMachinesRequest) (*contract.ListMachinesResponse, error) {
	var limit, offset int32
	if req != nil {
		limit, offset = req.Limit, req.Offset
	}
	if limit <= 0 {
		limit = u.defaultPageSize
	}
	if limit > u.maxPageSize {
		limit = u.maxPageSize
	}
	if offset < 0 {
		offset = 0
	}

	machines, err := u.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list machines: %w", err)
	}

	resp := &contract.ListMachinesResponse{Machines: make([]contract.MachineResponse, len(machines))}
	for i := range machines {
		resp.Machines[i] = newMachineResponse(&machines[i])
	}
	return resp, nil
}
