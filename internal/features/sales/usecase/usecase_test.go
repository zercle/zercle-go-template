//go:build unit

package usecase_test

import (
	"context"
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/zercle/zercle-go-template/internal/features/sales/contract"
	"github.com/zercle/zercle-go-template/internal/features/sales/domain"
	"github.com/zercle/zercle-go-template/internal/features/sales/repository/mock"
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
	resp, err := svc.Purchase(ctx, &contract.PurchaseRequest{
		MachineID: machineID.String(),
		ProductID: productID.String(),
		Coins:     []int32{100},
	})

	require.NoError(t, err)
	require.NotNil(t, gotRecord)
	require.Equal(t, int32(75), gotRecord.PriceCents)
	require.Equal(t, int32(100), gotRecord.TotalInsertedCents)
	require.Equal(t, int32(25), gotRecord.ChangeCents)
	require.Equal(t, wantChange, gotRecord.ChangeCoins)
	require.Equal(t, wantRemaining, gotBank, "bankAfter must be the domain-remaining bank")

	require.Equal(t, machineID.String(), resp.MachineID)
	require.Equal(t, productID.String(), resp.ProductID)
	require.Equal(t, int32(75), resp.PriceCents)
	require.Equal(t, int32(25), resp.ChangeCents)
	require.Equal(t, []int32{25}, resp.ChangeCoins)
	require.NotEmpty(t, resp.PurchasedAt)
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
	resp, err := svc.Purchase(ctx, &contract.PurchaseRequest{
		MachineID: machineID.String(),
		ProductID: productID.String(),
		Coins:     []int32{50},
	})

	require.NoError(t, err)
	require.Equal(t, int32(0), resp.ChangeCents)
	require.NotNil(t, resp.ChangeCoins)
	require.Empty(t, resp.ChangeCoins)
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
	resp, err := svc.Purchase(ctx, &contract.PurchaseRequest{
		MachineID: machineID.String(),
		ProductID: productID.String(),
		Coins:     []int32{25},
	})

	require.ErrorIs(t, err, domain.ErrInsufficientPayment)
	require.Nil(t, resp)
}

func TestService_Purchase_UnsupportedCoinRejected(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := usecase.NewUsecase(repo)

	resp, err := svc.Purchase(ctx, &contract.PurchaseRequest{
		MachineID: uuid.New().String(),
		ProductID: uuid.New().String(),
		Coins:     []int32{7},
	})

	require.ErrorIs(t, err, domain.ErrUnsupportedCoin)
	require.Nil(t, resp)
}

func TestService_Purchase_ProductMissing(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	machineID, productID := uuid.New(), uuid.New()

	repo.EXPECT().GetProduct(ctx, productID).Return(domain.SaleProduct{}, domain.ErrProductNotFound)

	svc := usecase.NewUsecase(repo)
	resp, err := svc.Purchase(ctx, &contract.PurchaseRequest{
		MachineID: machineID.String(),
		ProductID: productID.String(),
		Coins:     []int32{50},
	})

	require.ErrorIs(t, err, domain.ErrProductNotFound)
	require.Nil(t, resp)
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
	resp, err := svc.Purchase(ctx, &contract.PurchaseRequest{
		MachineID: machineID.String(),
		ProductID: productID.String(),
		Coins:     []int32{50},
	})

	require.ErrorIs(t, err, domain.ErrMachineNotFound)
	require.Nil(t, resp)
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
	resp, err := svc.Purchase(ctx, &contract.PurchaseRequest{
		MachineID: machineID.String(),
		ProductID: productID.String(),
		Coins:     []int32{50},
	})

	require.ErrorIs(t, err, domain.ErrOutOfStock)
	require.Nil(t, resp)
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
	resp, err := svc.Purchase(ctx, &contract.PurchaseRequest{
		MachineID: machineID.String(),
		ProductID: productID.String(),
		Coins:     []int32{50},
	})

	require.ErrorIs(t, err, domain.ErrOutOfStock)
	require.Nil(t, resp)
}

func TestService_Purchase_InvalidMachineID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := usecase.NewUsecase(repo)

	resp, err := svc.Purchase(ctx, &contract.PurchaseRequest{
		MachineID: "not-a-uuid",
		ProductID: uuid.New().String(),
		Coins:     []int32{50},
	})

	require.ErrorIs(t, err, domain.ErrInvalidID)
	require.Nil(t, resp)
}

func TestService_Purchase_InvalidProductID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := mock.NewMockRepository(gomock.NewController(t))
	svc := usecase.NewUsecase(repo)

	resp, err := svc.Purchase(ctx, &contract.PurchaseRequest{
		MachineID: uuid.New().String(),
		ProductID: "not-a-uuid",
		Coins:     []int32{50},
	})

	require.ErrorIs(t, err, domain.ErrInvalidID)
	require.Nil(t, resp)
}

func knownBank() domain.CoinBank {
	return domain.CoinBank{5: 10, 10: 10, 25: 10, 50: 10, 100: 10}
}
