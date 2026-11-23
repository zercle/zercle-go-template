//go:build unit

package usecase_test

import (
	"context"
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/zercle/zercle-go-template/internal/features/sales/domain"
	"github.com/zercle/zercle-go-template/internal/features/sales/port/mock"
	"github.com/zercle/zercle-go-template/internal/features/sales/usecase"
)

func TestService_Purchase_HappyWithChange(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	machineID, productID := uuid.New(), uuid.New()

	product := domain.SaleProduct{ProductID: productID, PriceCents: 75, Stock: 5}
	bank := knownBank()

	wantChange, wantRemaining, err := domain.Purchase(product.PriceCents, product.Stock, bank, []int32{100})
	require.NoError(t, err)

	repo.EXPECT().GetProduct(ctx, productID).Return(product, nil)
	repo.EXPECT().GetMachineBank(ctx, machineID).Return(bank, nil)

	var gotRecord *domain.PurchaseRecord
	var gotBank domain.CoinBank
	repo.EXPECT().CommitPurchase(ctx, machineID, productID, gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _, _ uuid.UUID, record *domain.PurchaseRecord, bankAfter domain.CoinBank) error {
			gotRecord = record
			gotBank = bankAfter
			return nil
		})

	svc := usecase.NewUsecase(repo)
	record, err := svc.Purchase(ctx, machineID.String(), productID.String(), []int32{100})

	require.NoError(t, err)
	require.NotNil(t, gotRecord)
	require.Equal(t, int32(75), gotRecord.PriceCents)
	require.Equal(t, int32(100), gotRecord.TotalInsertedCents)
	require.Equal(t, int32(25), gotRecord.ChangeCents)
	require.Equal(t, wantChange, gotRecord.ChangeCoins)
	require.Equal(t, wantRemaining, gotBank, "bankAfter must be the domain-remaining bank")

	require.Equal(t, machineID, record.MachineID)
	require.Equal(t, productID, record.ProductID)
	require.Equal(t, int32(75), record.PriceCents)
	require.Equal(t, int32(25), record.ChangeCents)
	require.Equal(t, []int32{25}, record.ChangeCoins)
	require.False(t, record.PurchasedAt.IsZero())
}

func TestService_Purchase_ExactPayment(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	machineID, productID := uuid.New(), uuid.New()

	product := domain.SaleProduct{ProductID: productID, PriceCents: 50, Stock: 2}
	bank := knownBank()

	repo.EXPECT().GetProduct(ctx, productID).Return(product, nil)
	repo.EXPECT().GetMachineBank(ctx, machineID).Return(bank, nil)

	var gotBank domain.CoinBank
	repo.EXPECT().CommitPurchase(ctx, machineID, productID, gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _, _ uuid.UUID, _ *domain.PurchaseRecord, bankAfter domain.CoinBank) error {
			gotBank = bankAfter
			return nil
		})

	svc := usecase.NewUsecase(repo)
	record, err := svc.Purchase(ctx, machineID.String(), productID.String(), []int32{50})

	require.NoError(t, err)
	require.Equal(t, int32(0), record.ChangeCents)
	require.NotNil(t, record.ChangeCoins)
	require.Empty(t, record.ChangeCoins)
	require.Equal(t, bank, gotBank, "no change means the bank is unchanged")
}

func TestService_Purchase_InsufficientPayment(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	machineID, productID := uuid.New(), uuid.New()

	product := domain.SaleProduct{ProductID: productID, PriceCents: 100, Stock: 5}
	repo.EXPECT().GetProduct(ctx, productID).Return(product, nil)
	repo.EXPECT().GetMachineBank(ctx, machineID).Return(knownBank(), nil)

	svc := usecase.NewUsecase(repo)
	record, err := svc.Purchase(ctx, machineID.String(), productID.String(), []int32{25})

	require.ErrorIs(t, err, domain.ErrInsufficientPayment)
	require.Nil(t, record)
}

func TestService_Purchase_UnsupportedCoinRejected(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := usecase.NewUsecase(repo)

	record, err := svc.Purchase(ctx, uuid.New().String(), uuid.New().String(), []int32{7})

	require.ErrorIs(t, err, domain.ErrUnsupportedCoin)
	require.Nil(t, record)
}

func TestService_Purchase_ProductMissing(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	machineID, productID := uuid.New(), uuid.New()

	repo.EXPECT().GetProduct(ctx, productID).Return(domain.SaleProduct{}, domain.ErrProductNotFound)

	svc := usecase.NewUsecase(repo)
	record, err := svc.Purchase(ctx, machineID.String(), productID.String(), []int32{50})

	require.ErrorIs(t, err, domain.ErrProductNotFound)
	require.Nil(t, record)
}

func TestService_Purchase_MachineMissing(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	machineID, productID := uuid.New(), uuid.New()

	product := domain.SaleProduct{ProductID: productID, PriceCents: 50, Stock: 5}
	repo.EXPECT().GetProduct(ctx, productID).Return(product, nil)
	repo.EXPECT().GetMachineBank(ctx, machineID).Return(nil, domain.ErrMachineNotFound)

	svc := usecase.NewUsecase(repo)
	record, err := svc.Purchase(ctx, machineID.String(), productID.String(), []int32{50})

	require.ErrorIs(t, err, domain.ErrMachineNotFound)
	require.Nil(t, record)
}

func TestService_Purchase_OutOfStock(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	machineID, productID := uuid.New(), uuid.New()

	product := domain.SaleProduct{ProductID: productID, PriceCents: 50, Stock: 0}
	repo.EXPECT().GetProduct(ctx, productID).Return(product, nil)
	repo.EXPECT().GetMachineBank(ctx, machineID).Return(knownBank(), nil)

	svc := usecase.NewUsecase(repo)
	record, err := svc.Purchase(ctx, machineID.String(), productID.String(), []int32{50})

	require.ErrorIs(t, err, domain.ErrOutOfStock)
	require.Nil(t, record)
}

func TestService_Purchase_CommitOutOfStockRacePropagates(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	machineID, productID := uuid.New(), uuid.New()

	product := domain.SaleProduct{ProductID: productID, PriceCents: 50, Stock: 1}
	repo.EXPECT().GetProduct(ctx, productID).Return(product, nil)
	repo.EXPECT().GetMachineBank(ctx, machineID).Return(knownBank(), nil)
	repo.EXPECT().CommitPurchase(ctx, machineID, productID, gomock.Any(), gomock.Any()).Return(domain.ErrOutOfStock)

	svc := usecase.NewUsecase(repo)
	record, err := svc.Purchase(ctx, machineID.String(), productID.String(), []int32{50})

	require.ErrorIs(t, err, domain.ErrOutOfStock)
	require.Nil(t, record)
}

func TestService_Purchase_InvalidMachineID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := usecase.NewUsecase(repo)

	record, err := svc.Purchase(ctx, "not-a-uuid", uuid.New().String(), []int32{50})

	require.ErrorIs(t, err, domain.ErrInvalidID)
	require.Nil(t, record)
}

func TestService_Purchase_InvalidProductID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := usecase.NewUsecase(repo)

	record, err := svc.Purchase(ctx, uuid.New().String(), "not-a-uuid", []int32{50})

	require.ErrorIs(t, err, domain.ErrInvalidID)
	require.Nil(t, record)
}

func knownBank() domain.CoinBank {
	return domain.CoinBank{5: 10, 10: 10, 25: 10, 50: 10, 100: 10}
}
