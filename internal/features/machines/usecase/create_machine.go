package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
	"uuid"

	"github.com/zercle/zercle-go-template/internal/features/machines/contract"
	"github.com/zercle/zercle-go-template/internal/features/machines/domain"
)

// Create validates the label and the initial coins, persists a new machine,
// and returns its wire form. The label is trimmed first so surrounding
// whitespace never counts toward the configured length limit.
func (u *Usecase) Create(ctx context.Context, req *contract.CreateMachineRequest) (*contract.MachineResponse, error) {
	var label string
	var initialCoins []int32
	if req != nil {
		label = strings.TrimSpace(req.Label)
		initialCoins = req.InitialCoins
	}
	if label == "" || utf8.RuneCountInString(label) > int(u.maxLabelLength) {
		return nil, fmt.Errorf("%w: must be 1..%d characters", domain.ErrInvalidMachineLabel, u.maxLabelLength)
	}
	if len(initialCoins) > 0 {
		if err := domain.ValidateCoins(initialCoins); err != nil {
			return nil, fmt.Errorf("validate initial coins: %w", err)
		}
	}

	now := time.Now().UTC()
	machine := &domain.Machine{
		ID:        uuid.New(),
		Label:     label,
		CoinBank:  domain.AddCoins(domain.CoinBank{}, initialCoins),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := u.repo.Create(ctx, machine); err != nil {
		return nil, fmt.Errorf("create machine: %w", err)
	}

	resp := newMachineResponse(machine)
	return &resp, nil
}
