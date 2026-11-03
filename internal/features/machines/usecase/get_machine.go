package usecase

import (
	"context"
	"errors"
	"fmt"
	"uuid"

	"github.com/zercle/zercle-go-template/internal/features/machines/contract"
	"github.com/zercle/zercle-go-template/internal/features/machines/domain"
)

// Get retrieves a machine by ID, passing through domain.ErrMachineNotFound. The
// wire id string is parsed here so both driving adapters share one validation
// path. A syntactically valid but absent id (including the nil UUID) is a
// not-found, which the repository decides.
func (u *Usecase) Get(ctx context.Context, id string) (*contract.MachineResponse, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, domain.ErrInvalidID
	}
	machine, err := u.repo.GetByID(ctx, parsed)
	if err != nil {
		if errors.Is(err, domain.ErrMachineNotFound) {
			return nil, domain.ErrMachineNotFound
		}
		return nil, fmt.Errorf("get machine: %w", err)
	}

	resp := newMachineResponse(machine)
	return &resp, nil
}
