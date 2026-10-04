//go:build unit

package usecase_test

import (
	"context"
	"errors"
	"maps"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/zercle/zercle-go-template/internal/features/machines/contract"
	"github.com/zercle/zercle-go-template/internal/features/machines/domain"
	"github.com/zercle/zercle-go-template/internal/features/machines/repository/mock"
	"github.com/zercle/zercle-go-template/internal/features/machines/usecase"
)

func TestService_Create_HappyWithCoins(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	repo.EXPECT().Create(ctx, matchMachine("front", domain.CoinBank{5: 1, 25: 2})).Return(nil)

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	resp, err := svc.Create(ctx, &contract.CreateMachineRequest{Label: "  front  ", InitialCoins: []int32{5, 25, 25}})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "front", resp.Label)
	require.Equal(t, map[int32]int32{5: 1, 25: 2}, resp.CoinBank)
	parsed, err := uuid.Parse(resp.ID)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil(), parsed)
	require.NotEmpty(t, resp.CreatedAt)
	require.NotEmpty(t, resp.UpdatedAt)
}

func TestService_Create_HappyWithoutCoins(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	repo.EXPECT().Create(ctx, matchMachine("front", domain.CoinBank{})).Return(nil)

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	resp, err := svc.Create(ctx, &contract.CreateMachineRequest{Label: "front"})

	require.NoError(t, err)
	require.NotNil(t, resp.CoinBank)
	require.Empty(t, resp.CoinBank)
}

func TestService_Create_EmptyLabel(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := usecase.NewUsecase(repo, 0, 0, 0)

	resp, err := svc.Create(ctx, &contract.CreateMachineRequest{Label: "   "})

	require.ErrorIs(t, err, domain.ErrInvalidMachineLabel)
	require.Nil(t, resp)
}

func TestService_Create_OverLengthLabel(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := usecase.NewUsecase(repo, 0, 0, 3)

	resp, err := svc.Create(ctx, &contract.CreateMachineRequest{Label: "abcd"})

	require.ErrorIs(t, err, domain.ErrInvalidMachineLabel)
	require.Nil(t, resp)
}

func TestService_Create_UnsupportedCoinRejected(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := usecase.NewUsecase(repo, 0, 0, 0)

	resp, err := svc.Create(ctx, &contract.CreateMachineRequest{Label: "front", InitialCoins: []int32{7}})

	require.ErrorIs(t, err, domain.ErrUnsupportedCoin)
	require.Nil(t, resp)
}

func TestService_Create_RepositoryError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	repo.EXPECT().Create(ctx, matchMachine("front", domain.CoinBank{})).Return(errors.New("boom"))

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	resp, err := svc.Create(ctx, &contract.CreateMachineRequest{Label: "front"})

	require.Error(t, err)
	require.Nil(t, resp)
}

func TestService_Get_Happy(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	id := uuid.New()

	now := time.Now().UTC().Truncate(time.Second)
	expected := &domain.Machine{ID: id, Label: "found", CoinBank: domain.CoinBank{25: 3}, CreatedAt: now, UpdatedAt: now}
	repo.EXPECT().GetByID(ctx, id).Return(expected, nil)

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	resp, err := svc.Get(ctx, id.String())

	require.NoError(t, err)
	require.Equal(t, &contract.MachineResponse{
		ID:        id.String(),
		Label:     "found",
		CoinBank:  map[int32]int32{25: 3},
		CreatedAt: now.Format(time.RFC3339),
		UpdatedAt: now.Format(time.RFC3339),
	}, resp)
}

func TestService_Get_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	id := uuid.New()

	repo.EXPECT().GetByID(ctx, id).Return(nil, domain.ErrMachineNotFound)

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	resp, err := svc.Get(ctx, id.String())

	require.ErrorIs(t, err, domain.ErrMachineNotFound)
	require.Nil(t, resp)
}

func TestService_Get_InvalidID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := usecase.NewUsecase(repo, 0, 0, 0)

	resp, err := svc.Get(ctx, "not-a-uuid")

	require.ErrorIs(t, err, domain.ErrInvalidID)
	require.Nil(t, resp)
}

func TestService_List_Happy(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	now := time.Now().UTC().Truncate(time.Second)
	expected := []domain.Machine{{ID: uuid.New(), Label: "one", CoinBank: domain.CoinBank{5: 2}, CreatedAt: now, UpdatedAt: now}}
	repo.EXPECT().List(ctx, int32(10), int32(5)).Return(expected, nil)

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	resp, err := svc.List(ctx, &contract.ListMachinesRequest{Limit: 10, Offset: 5})

	require.NoError(t, err)
	require.Len(t, resp.Machines, 1)
	require.Equal(t, expected[0].ID.String(), resp.Machines[0].ID)
	require.Equal(t, "one", resp.Machines[0].Label)
	require.Equal(t, map[int32]int32{5: 2}, resp.Machines[0].CoinBank)
}

func TestService_List_AppliesDefaultLimit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	repo.EXPECT().List(ctx, int32(20), int32(0)).Return(nil, nil)

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	resp, err := svc.List(ctx, &contract.ListMachinesRequest{Limit: 0, Offset: 0})

	require.NoError(t, err)
	require.NotNil(t, resp.Machines)
	require.Empty(t, resp.Machines)
}

func TestService_List_ClampsOverMaxLimit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	repo.EXPECT().List(ctx, int32(50), int32(0)).Return(nil, nil)

	svc := usecase.NewUsecase(repo, 10, 50, 255)
	resp, err := svc.List(ctx, &contract.ListMachinesRequest{Limit: 999})

	require.NoError(t, err)
	require.Empty(t, resp.Machines)
}

func TestService_RestockBank_Happy(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	id := uuid.New()

	now := time.Now().UTC().Truncate(time.Second)
	updated := &domain.Machine{ID: id, Label: "front", CoinBank: domain.CoinBank{5: 1, 25: 3}, CreatedAt: now, UpdatedAt: now}
	repo.EXPECT().RestockBank(ctx, id, []int32{25}).Return(nil)
	repo.EXPECT().GetByID(ctx, id).Return(updated, nil)

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	resp, err := svc.RestockBank(ctx, id.String(), &contract.RestockBankRequest{Coins: []int32{25}})

	require.NoError(t, err)
	require.Equal(t, id.String(), resp.ID)
	require.Equal(t, map[int32]int32{5: 1, 25: 3}, resp.CoinBank)
}

func TestService_RestockBank_InvalidID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := usecase.NewUsecase(repo, 0, 0, 0)

	resp, err := svc.RestockBank(ctx, "not-a-uuid", &contract.RestockBankRequest{Coins: []int32{25}})

	require.ErrorIs(t, err, domain.ErrInvalidID)
	require.Nil(t, resp)
}

func TestService_RestockBank_UnsupportedCoinRejected(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := usecase.NewUsecase(repo, 0, 0, 0)

	resp, err := svc.RestockBank(ctx, uuid.New().String(), &contract.RestockBankRequest{Coins: []int32{7}})

	require.ErrorIs(t, err, domain.ErrUnsupportedCoin)
	require.Nil(t, resp)
}

func TestService_RestockBank_MachineMissing(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	id := uuid.New()

	repo.EXPECT().RestockBank(ctx, id, []int32{25}).Return(domain.ErrMachineNotFound)

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	resp, err := svc.RestockBank(ctx, id.String(), &contract.RestockBankRequest{Coins: []int32{25}})

	require.ErrorIs(t, err, domain.ErrMachineNotFound)
	require.Nil(t, resp)
}

func TestService_RestockBank_ReadBackMissing(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	id := uuid.New()

	repo.EXPECT().RestockBank(ctx, id, []int32{25}).Return(nil)
	repo.EXPECT().GetByID(ctx, id).Return(nil, domain.ErrMachineNotFound)

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	resp, err := svc.RestockBank(ctx, id.String(), &contract.RestockBankRequest{Coins: []int32{25}})

	require.ErrorIs(t, err, domain.ErrMachineNotFound)
	require.Nil(t, resp)
}

func matchMachine(label string, bank domain.CoinBank) any {
	return matchMachineFields{label: label, bank: bank}
}

type matchMachineFields struct {
	label string
	bank  domain.CoinBank
}

func (m matchMachineFields) Matches(x any) bool {
	machine, ok := x.(*domain.Machine)
	return ok && machine.Label == m.label && maps.Equal(machine.CoinBank, m.bank)
}

func (m matchMachineFields) String() string {
	return "is machine " + m.label
}
