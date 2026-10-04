package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
	"uuid"

	"github.com/zercle/zercle-go-template/internal/features/machines/contract"
	"github.com/zercle/zercle-go-template/internal/features/machines/domain"
	"github.com/zercle/zercle-go-template/internal/features/machines/repository"
)

const (
	defaultPageSizeFallback int32 = 20
	maxPageSizeFallback     int32 = 100
	maxLabelLengthFallback  int32 = 255

	timeFormat = time.RFC3339
)

// Usecase implements the Service inbound use-case interface.
type Usecase struct {
	repo            repository.Repository
	defaultPageSize int32
	maxPageSize     int32
	maxLabelLength  int32
}

// NewUsecase returns a Usecase backed by the provided repository. The limit
// arguments override the package fallback defaults; pass <= 0 to use the
// built-in defaults (20/100/255).
func NewUsecase(repo repository.Repository, defaultPageSize, maxPageSize, maxLabelLength int32) *Usecase {
	if defaultPageSize <= 0 {
		defaultPageSize = defaultPageSizeFallback
	}
	if maxPageSize <= 0 {
		maxPageSize = maxPageSizeFallback
	}
	if maxLabelLength <= 0 {
		maxLabelLength = maxLabelLengthFallback
	}
	return &Usecase{
		repo:            repo,
		defaultPageSize: defaultPageSize,
		maxPageSize:     maxPageSize,
		maxLabelLength:  maxLabelLength,
	}
}

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

func newMachineResponse(machine *domain.Machine) contract.MachineResponse {
	if machine == nil {
		return contract.MachineResponse{}
	}
	bank := machine.CoinBank
	if bank == nil {
		bank = domain.CoinBank{}
	}
	return contract.MachineResponse{
		ID:        machine.ID.String(),
		Label:     machine.Label,
		CoinBank:  bank,
		CreatedAt: machine.CreatedAt.Format(timeFormat),
		UpdatedAt: machine.UpdatedAt.Format(timeFormat),
	}
}
