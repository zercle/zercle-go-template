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

	"github.com/zercle/zercle-go-template/internal/features/reporting/domain"
	"github.com/zercle/zercle-go-template/internal/features/reporting/port/mock"
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
	summary, err := svc.Summary(ctx, 0)

	require.NoError(t, err)
	require.NotNil(t, summary)
	assert.EqualValues(t, 1, summary.Overview.ProductCount)
	assert.Empty(t, summary.TopMachines)
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
			summary, err := svc.Summary(ctx, top)

			require.NoError(t, err)
			require.NotNil(t, summary)
		})
	}
}

func TestService_Summary_TopAboveMaxRejected(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	svc := usecase.NewUsecase(repo, testDefaultTop, testMaxTop)
	summary, err := svc.Summary(ctx, testMaxTop+1)

	require.ErrorIs(t, err, domain.ErrInvalidTopMachines)
	assert.Nil(t, summary)
}

func TestService_Summary_RepositoryErrorPropagates(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	sentinel := errors.New("database unavailable")

	repo.EXPECT().GetOverview(ctx).Return(domain.Overview{}, sentinel)

	svc := usecase.NewUsecase(repo, testDefaultTop, testMaxTop)
	summary, err := svc.Summary(ctx, 0)

	require.ErrorIs(t, err, sentinel)
	assert.Nil(t, summary)
}

func TestService_Summary_MapsDomainToResult(t *testing.T) {
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
	summary, err := svc.Summary(ctx, 3)

	require.NoError(t, err)
	require.NotNil(t, summary)
	assert.EqualValues(t, 4, summary.Overview.ProductCount)
	assert.EqualValues(t, 12, summary.Overview.TotalStock)
	assert.EqualValues(t, 2, summary.Overview.MachineCount)
	assert.EqualValues(t, 350, summary.Overview.TotalBankCents)
	assert.EqualValues(t, 7, summary.Overview.PurchaseCount)
	assert.EqualValues(t, 900, summary.Overview.RevenueCents)

	require.Len(t, summary.TopMachines, 2)
	assert.Equal(t, firstID, summary.TopMachines[0].MachineID)
	assert.Equal(t, "lobby", summary.TopMachines[0].Label)
	assert.EqualValues(t, 5, summary.TopMachines[0].PurchaseCount)
	assert.EqualValues(t, 700, summary.TopMachines[0].RevenueCents)
	assert.Equal(t, secondID, summary.TopMachines[1].MachineID)
}

func TestService_Summary_ZeroValueUsecaseUsesFallbackDefaults(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))

	// Zero-valued constructor args must fall back to 5/20.
	repo.EXPECT().GetOverview(ctx).Return(domain.Overview{}, nil)
	repo.EXPECT().GetTopMachines(ctx, int32(5)).Return([]domain.MachineSales{}, nil)

	svc := usecase.NewUsecase(repo, 0, 0)
	summary, err := svc.Summary(ctx, 0)

	require.NoError(t, err)
	require.NotNil(t, summary)
}
