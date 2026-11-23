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

	"github.com/zercle/zercle-go-template/internal/features/machines/domain"
	"github.com/zercle/zercle-go-template/internal/features/machines/port/mock"
	"github.com/zercle/zercle-go-template/internal/features/machines/usecase"
)

func TestService_Create_HappyWithCoins(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	repo.EXPECT().Create(ctx, matchMachine("front", domain.CoinBank{5: 1, 25: 2})).Return(nil)

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	machine, err := svc.Create(ctx, "  front  ", []int32{5, 25, 25})

	require.NoError(t, err)
	require.NotNil(t, machine)
	require.Equal(t, "front", machine.Label)
	require.Equal(t, domain.CoinBank{5: 1, 25: 2}, machine.CoinBank)
	require.NotEqual(t, uuid.Nil(), machine.ID)
	require.False(t, machine.CreatedAt.IsZero())
	require.False(t, machine.UpdatedAt.IsZero())
}

func TestService_Create_HappyWithoutCoins(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	repo.EXPECT().Create(ctx, matchMachine("front", domain.CoinBank{})).Return(nil)

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	machine, err := svc.Create(ctx, "front", nil)

	require.NoError(t, err)
	require.NotNil(t, machine.CoinBank)
	require.Empty(t, machine.CoinBank)
}

func TestService_Create_EmptyLabel(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := usecase.NewUsecase(repo, 0, 0, 0)

	machine, err := svc.Create(ctx, "   ", nil)

	require.ErrorIs(t, err, domain.ErrInvalidMachineLabel)
	require.Nil(t, machine)
}

func TestService_Create_OverLengthLabel(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := usecase.NewUsecase(repo, 0, 0, 3)

	machine, err := svc.Create(ctx, "abcd", nil)

	require.ErrorIs(t, err, domain.ErrInvalidMachineLabel)
	require.Nil(t, machine)
}

func TestService_Create_UnsupportedCoinRejected(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := usecase.NewUsecase(repo, 0, 0, 0)

	machine, err := svc.Create(ctx, "front", []int32{7})

	require.ErrorIs(t, err, domain.ErrUnsupportedCoin)
	require.Nil(t, machine)
}

func TestService_Create_RepositoryError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	repo.EXPECT().Create(ctx, matchMachine("front", domain.CoinBank{})).Return(errors.New("boom"))

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	machine, err := svc.Create(ctx, "front", nil)

	require.Error(t, err)
	require.Nil(t, machine)
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
	machine, err := svc.Get(ctx, id.String())

	require.NoError(t, err)
	require.Equal(t, expected, machine)
}

func TestService_Get_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	id := uuid.New()

	repo.EXPECT().GetByID(ctx, id).Return(nil, domain.ErrMachineNotFound)

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	machine, err := svc.Get(ctx, id.String())

	require.ErrorIs(t, err, domain.ErrMachineNotFound)
	require.Nil(t, machine)
}

func TestService_Get_InvalidID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := usecase.NewUsecase(repo, 0, 0, 0)

	machine, err := svc.Get(ctx, "not-a-uuid")

	require.ErrorIs(t, err, domain.ErrInvalidID)
	require.Nil(t, machine)
}

func TestService_List_Happy(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	now := time.Now().UTC().Truncate(time.Second)
	expected := []domain.Machine{{ID: uuid.New(), Label: "one", CoinBank: domain.CoinBank{5: 2}, CreatedAt: now, UpdatedAt: now}}
	repo.EXPECT().List(ctx, int32(10), int32(5)).Return(expected, nil)

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	machines, err := svc.List(ctx, 10, 5)

	require.NoError(t, err)
	require.Len(t, machines, 1)
	require.Equal(t, expected[0].ID, machines[0].ID)
	require.Equal(t, "one", machines[0].Label)
	require.Equal(t, domain.CoinBank{5: 2}, machines[0].CoinBank)
}

func TestService_List_AppliesDefaultLimit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	repo.EXPECT().List(ctx, int32(20), int32(0)).Return(nil, nil)

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	machines, err := svc.List(ctx, 0, 0)

	require.NoError(t, err)
	require.Empty(t, machines)
}

func TestService_List_ClampsOverMaxLimit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	repo.EXPECT().List(ctx, int32(50), int32(0)).Return(nil, nil)

	svc := usecase.NewUsecase(repo, 10, 50, 255)
	machines, err := svc.List(ctx, 999, 0)

	require.NoError(t, err)
	require.Empty(t, machines)
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
	machine, err := svc.RestockBank(ctx, id.String(), []int32{25})

	require.NoError(t, err)
	require.Equal(t, updated, machine)
}

func TestService_RestockBank_InvalidID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := usecase.NewUsecase(repo, 0, 0, 0)

	machine, err := svc.RestockBank(ctx, "not-a-uuid", []int32{25})

	require.ErrorIs(t, err, domain.ErrInvalidID)
	require.Nil(t, machine)
}

func TestService_RestockBank_UnsupportedCoinRejected(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := usecase.NewUsecase(repo, 0, 0, 0)

	machine, err := svc.RestockBank(ctx, uuid.New().String(), []int32{7})

	require.ErrorIs(t, err, domain.ErrUnsupportedCoin)
	require.Nil(t, machine)
}

func TestService_RestockBank_MachineMissing(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	id := uuid.New()

	repo.EXPECT().RestockBank(ctx, id, []int32{25}).Return(domain.ErrMachineNotFound)

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	machine, err := svc.RestockBank(ctx, id.String(), []int32{25})

	require.ErrorIs(t, err, domain.ErrMachineNotFound)
	require.Nil(t, machine)
}

func TestService_RestockBank_ReadBackMissing(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	id := uuid.New()

	repo.EXPECT().RestockBank(ctx, id, []int32{25}).Return(nil)
	repo.EXPECT().GetByID(ctx, id).Return(nil, domain.ErrMachineNotFound)

	svc := usecase.NewUsecase(repo, 0, 0, 0)
	machine, err := svc.RestockBank(ctx, id.String(), []int32{25})

	require.ErrorIs(t, err, domain.ErrMachineNotFound)
	require.Nil(t, machine)
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
