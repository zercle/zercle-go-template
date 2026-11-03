package usecase

import (
	"context"
	"errors"
	"fmt"
	"uuid"

	"github.com/zercle/zercle-go-template/internal/features/machines/contract"
	"github.com/zercle/zercle-go-template/internal/features/machines/domain"
)

// RestockBank validates the coins, applies them to the machine's bank through
// the repository, and re-reads the machine to return its authoritative bank.
//
// The re-read is deliberate: the repository port adds the coins inside one
// locked transaction so concurrent restocks cannot lose an update, and the
// resulting bank is only knowable from the datastore. Returning the bank from
// this fresh read keeps the response consistent with what a later read returns
// rather than echoing a locally composed map that a concurrent writer may have
// already superseded.
func (u *Usecase) RestockBank(ctx context.Context, id string, req *contract.RestockBankRequest) (*contract.MachineResponse, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, domain.ErrInvalidID
	}
	var coins []int32
	if req != nil {
		coins = req.Coins
	}
	if err := domain.ValidateCoins(coins); err != nil {
		return nil, fmt.Errorf("validate coins: %w", err)
	}

	if err := u.repo.RestockBank(ctx, parsed, coins); err != nil {
		if errors.Is(err, domain.ErrMachineNotFound) {
			return nil, domain.ErrMachineNotFound
		}
		return nil, fmt.Errorf("restock machine bank: %w", err)
	}

	machine, err := u.repo.GetByID(ctx, parsed)
	if err != nil {
		if errors.Is(err, domain.ErrMachineNotFound) {
			return nil, domain.ErrMachineNotFound
		}
		return nil, fmt.Errorf("get machine after restock: %w", err)
	}

	resp := newMachineResponse(machine)
	return &resp, nil
}
