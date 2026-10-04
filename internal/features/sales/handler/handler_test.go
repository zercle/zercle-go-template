//go:build unit

package handler_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"uuid"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/zercle/zercle-go-template/internal/features/sales/contract"
	"github.com/zercle/zercle-go-template/internal/features/sales/domain"
	"github.com/zercle/zercle-go-template/internal/features/sales/handler"
	"github.com/zercle/zercle-go-template/internal/features/sales/usecase/mock"
	apperrors "github.com/zercle/zercle-go-template/internal/infrastructure/errors"
	"github.com/zercle/zercle-go-template/internal/testutil"
)

// registerSentinelsOnce registers the sales feature's domain sentinels exactly
// once per test binary.
var registerSentinelsOnce sync.Once

func setupTest(t *testing.T) (*echo.Echo, *mock.MockService) {
	t.Helper()

	registerSentinelsOnce.Do(func() {
		apperrors.RegisterSentinel(domain.ErrProductNotFound, apperrors.ErrNotFound)
		apperrors.RegisterSentinel(domain.ErrMachineNotFound, apperrors.ErrNotFound)
		apperrors.RegisterSentinel(domain.ErrOutOfStock, apperrors.ErrConflict)
		apperrors.RegisterSentinel(domain.ErrInvalidID, apperrors.ErrInvalidInput)
		apperrors.RegisterSentinel(domain.ErrUnsupportedCoin, apperrors.ErrInvalidInput)
		apperrors.RegisterSentinel(domain.ErrInsufficientPayment, apperrors.ErrInvalidInput)
		apperrors.RegisterSentinel(domain.ErrExactChangeRequired, apperrors.ErrInvalidInput)
	})

	e := echo.New()
	e.Validator = newValidator(t)
	svc := mock.NewMockService(gomock.NewController(t))
	h := handler.New(svc)

	h.Register(e.Group("/api/v1"))

	return e, svc
}

func newValidator(t *testing.T) echo.Validator {
	t.Helper()
	return &validatorAdapter{v: validator.New()}
}

type validatorAdapter struct {
	v *validator.Validate
}

func (v *validatorAdapter) Validate(i any) error {
	return v.v.Struct(i)
}

// purchaseBody builds a valid purchase payload with the given ids and coins.
func purchaseBody(machineID, productID string, coins []int32) map[string]any {
	return map[string]any{"machine_id": machineID, "product_id": productID, "coins": coins}
}

func TestHandler_Purchase(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)
	machineID := uuid.New().String()
	productID := uuid.New().String()

	svc.EXPECT().Purchase(ctx, &contract.PurchaseRequest{MachineID: machineID, ProductID: productID, Coins: []int32{50, 25}}).
		Return(&contract.PurchaseResponse{
			ID: uuid.New().String(), MachineID: machineID, ProductID: productID,
			PriceCents: 50, TotalInsertedCents: 75, ChangeCents: 25, ChangeCoins: []int32{25},
		}, nil)

	rec := testutil.DoJSON(t, e, http.MethodPost, "/api/v1/purchases", purchaseBody(machineID, productID, []int32{50, 25}))

	require.Equal(t, http.StatusCreated, rec.Code)

	var body contract.PurchaseResponse
	testutil.DecodeJSON(t, rec, &body)
	require.Equal(t, []int32{25}, body.ChangeCoins)
}

func TestHandler_Purchase_InvalidJSON(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, _ := setupTest(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/purchases", bytes.NewReader([]byte(`{`)))
	req.Header.Set("Content-Type", "application/json")

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var body map[string]any
	testutil.DecodeJSON(t, rec, &body)
	require.Equal(t, "INVALID_INPUT", body["error"])
}

func TestHandler_Purchase_ValidationFailure(t *testing.T) {
	t.Parallel()

	e, _ := setupTest(t)

	rec := testutil.DoJSON(t, e, http.MethodPost, "/api/v1/purchases",
		map[string]any{"machine_id": "not-a-uuid", "product_id": uuid.New().String(), "coins": []int32{50}})

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var body map[string]any
	testutil.DecodeJSON(t, rec, &body)
	require.Equal(t, "INVALID_INPUT", body["error"])
}

func TestHandler_Purchase_InsufficientPayment(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)
	machineID := uuid.New().String()
	productID := uuid.New().String()

	svc.EXPECT().Purchase(ctx, &contract.PurchaseRequest{MachineID: machineID, ProductID: productID, Coins: []int32{5}}).
		Return(nil, domain.ErrInsufficientPayment)

	rec := testutil.DoJSON(t, e, http.MethodPost, "/api/v1/purchases", purchaseBody(machineID, productID, []int32{5}))

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var body map[string]any
	testutil.DecodeJSON(t, rec, &body)
	require.Equal(t, "INVALID_INPUT", body["error"])
}

func TestHandler_Purchase_OutOfStock(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)
	machineID := uuid.New().String()
	productID := uuid.New().String()

	svc.EXPECT().Purchase(ctx, &contract.PurchaseRequest{MachineID: machineID, ProductID: productID, Coins: []int32{50}}).
		Return(nil, domain.ErrOutOfStock)

	rec := testutil.DoJSON(t, e, http.MethodPost, "/api/v1/purchases", purchaseBody(machineID, productID, []int32{50}))

	require.Equal(t, http.StatusConflict, rec.Code)

	var body map[string]any
	testutil.DecodeJSON(t, rec, &body)
	require.Equal(t, "CONFLICT", body["error"])
}

func TestHandler_Purchase_ProductNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)
	machineID := uuid.New().String()
	productID := uuid.New().String()

	svc.EXPECT().Purchase(ctx, &contract.PurchaseRequest{MachineID: machineID, ProductID: productID, Coins: []int32{50}}).
		Return(nil, domain.ErrProductNotFound)

	rec := testutil.DoJSON(t, e, http.MethodPost, "/api/v1/purchases", purchaseBody(machineID, productID, []int32{50}))

	require.Equal(t, http.StatusNotFound, rec.Code)

	var body map[string]any
	testutil.DecodeJSON(t, rec, &body)
	require.Equal(t, "NOT_FOUND", body["error"])
}
