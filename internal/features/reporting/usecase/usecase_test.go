//go:build unit

package usecase_test

import (
	"context"
	"errors"
	"testing"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/zercle/zercle-go-template/internal/features/reporting/contract"
	"github.com/zercle/zercle-go-template/internal/features/reporting/domain"
	"github.com/zercle/zercle-go-template/internal/features/reporting/repository/mock"
	"github.com/zercle/zercle-go-template/internal/features/reporting/usecase"
)

const (
	testDefaultTop int32 = 5
	testMaxTop     int32 = 20
)

func TestService_Summary_NilRequestUsesDefault(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	repo.EXPECT().GetOverview(ctx).Return(domain.Overview{ProductCount: 1}, nil)
	repo.EXPECT().GetTopMachines(ctx, testDefaultTop).Return([]domain.MachineSales{}, nil)

	svc := usecase.NewUsecase(repo, testDefaultTop, testMaxTop)
	resp, err := svc.Summary(ctx, nil)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.EqualValues(t, 1, resp.Catalog.ProductCount)
	require.NotNil(t, resp.TopMachines)
	assert.Empty(t, resp.TopMachines)
}

func TestService_Summary_NonPositiveTopUsesDefault(t *testing.T) {
	t.Parallel()

	for _, top := range []int32{0, -1} {
		t.Run("top", func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			repo := mock.NewMockRepository(gomock.NewController(t))
			repo.EXPECT().GetOverview(ctx).Return(domain.Overview{}, nil)
			repo.EXPECT().GetTopMachines(ctx, testDefaultTop).Return([]domain.MachineSales{}, nil)

			svc := usecase.NewUsecase(repo, testDefaultTop, testMaxTop)
			resp, err := svc.Summary(ctx, &contract.SummaryRequest{TopMachines: top})

			require.NoError(t, err)
			require.NotNil(t, resp)
		})
	}
}

func TestService_Summary_TopAboveMaxRejected(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	svc := usecase.NewUsecase(repo, testDefaultTop, testMaxTop)
	resp, err := svc.Summary(ctx, &contract.SummaryRequest{TopMachines: testMaxTop + 1})

	require.ErrorIs(t, err, domain.ErrInvalidTopMachines)
	assert.Nil(t, resp)
}

func TestService_Summary_RepositoryErrorPropagates(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	sentinel := errors.New("database unavailable")

	repo.EXPECT().GetOverview(ctx).Return(domain.Overview{}, sentinel)

	svc := usecase.NewUsecase(repo, testDefaultTop, testMaxTop)
	resp, err := svc.Summary(ctx, &contract.SummaryRequest{})

	require.ErrorIs(t, err, sentinel)
	assert.Nil(t, resp)
}

func TestService_Summary_MapsDomainToContract(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	firstID := uuid.MustParse("12345678-1234-1234-1234-123456789abc")
	secondID := uuid.MustParse("22345678-1234-1234-1234-123456789abc")

	repo.EXPECT().GetOverview(ctx).Return(domain.Overview{
		ProductCount:   4,
		TotalStock:     12,
		MachineCount:   2,
		TotalBankCents: 350,
		PurchaseCount:  7,
		RevenueCents:   900,
	}, nil)
	repo.EXPECT().GetTopMachines(ctx, int32(3)).Return([]domain.MachineSales{
		{MachineID: firstID, Label: "lobby", PurchaseCount: 5, RevenueCents: 700},
		{MachineID: secondID, Label: "cafe", PurchaseCount: 2, RevenueCents: 200},
	}, nil)

	svc := usecase.NewUsecase(repo, testDefaultTop, testMaxTop)
	resp, err := svc.Summary(ctx, &contract.SummaryRequest{TopMachines: 3})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.EqualValues(t, 4, resp.Catalog.ProductCount)
	assert.EqualValues(t, 12, resp.Catalog.TotalStock)
	assert.EqualValues(t, 2, resp.Machines.MachineCount)
	assert.EqualValues(t, 350, resp.Machines.TotalCoinBankCents)
	assert.EqualValues(t, 7, resp.Sales.PurchaseCount)
	assert.EqualValues(t, 900, resp.Sales.RevenueCents)

	require.Len(t, resp.TopMachines, 2)
	assert.Equal(t, firstID.String(), resp.TopMachines[0].MachineID)
	assert.Equal(t, "lobby", resp.TopMachines[0].Label)
	assert.EqualValues(t, 5, resp.TopMachines[0].PurchaseCount)
	assert.EqualValues(t, 700, resp.TopMachines[0].RevenueCents)
	assert.Equal(t, secondID.String(), resp.TopMachines[1].MachineID)
}

func TestService_Summary_ZeroValueUsecaseUsesFallbackDefaults(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	// Zero-valued constructor args must fall back to 5/20.
	repo.EXPECT().GetOverview(ctx).Return(domain.Overview{}, nil)
	repo.EXPECT().GetTopMachines(ctx, int32(5)).Return([]domain.MachineSales{}, nil)

	svc := usecase.NewUsecase(repo, 0, 0)
	resp, err := svc.Summary(ctx, &contract.SummaryRequest{})

	require.NoError(t, err)
	require.NotNil(t, resp)
}
