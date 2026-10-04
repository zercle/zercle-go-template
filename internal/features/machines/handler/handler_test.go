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

	"github.com/zercle/zercle-go-template/internal/features/machines/contract"
	"github.com/zercle/zercle-go-template/internal/features/machines/domain"
	"github.com/zercle/zercle-go-template/internal/features/machines/handler"
	"github.com/zercle/zercle-go-template/internal/features/machines/usecase/mock"
	apperrors "github.com/zercle/zercle-go-template/internal/infrastructure/errors"
	"github.com/zercle/zercle-go-template/internal/testutil"
)

// registerSentinelsOnce registers the machines feature's domain sentinels
// exactly once per test binary.
var registerSentinelsOnce sync.Once

func setupTest(t *testing.T) (*echo.Echo, *mock.MockService) {
	t.Helper()

	registerSentinelsOnce.Do(func() {
		apperrors.RegisterSentinel(domain.ErrMachineNotFound, apperrors.ErrNotFound)
		apperrors.RegisterSentinel(domain.ErrInvalidID, apperrors.ErrInvalidInput)
		apperrors.RegisterSentinel(domain.ErrInvalidMachineLabel, apperrors.ErrInvalidInput)
		apperrors.RegisterSentinel(domain.ErrUnsupportedCoin, apperrors.ErrInvalidInput)
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

func TestHandler_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)
	id := uuid.New()

	svc.EXPECT().Create(ctx, &contract.CreateMachineRequest{Label: "m1", InitialCoins: []int32{25, 25}}).
		Return(&contract.MachineResponse{ID: id.String(), Label: "m1", CoinBank: map[int32]int32{25: 2}}, nil)

	rec := testutil.DoJSON(t, e, http.MethodPost, "/api/v1/machines",
		map[string]any{"label": "m1", "initial_coins": []int32{25, 25}})

	require.Equal(t, http.StatusCreated, rec.Code)
	require.Contains(t, rec.Body.String(), "m1")
}

func TestHandler_Create_InvalidJSON(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, _ := setupTest(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/machines", bytes.NewReader([]byte(`{`)))
	req.Header.Set("Content-Type", "application/json")

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var body map[string]any
	testutil.DecodeJSON(t, rec, &body)
	require.Equal(t, "INVALID_INPUT", body["error"])
}

func TestHandler_Create_ValidationFailure(t *testing.T) {
	t.Parallel()

	e, _ := setupTest(t)

	rec := testutil.DoJSON(t, e, http.MethodPost, "/api/v1/machines",
		map[string]any{"label": ""})

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var body map[string]any
	testutil.DecodeJSON(t, rec, &body)
	require.Equal(t, "INVALID_INPUT", body["error"])
}

func TestHandler_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)
	id := uuid.New()

	svc.EXPECT().Get(ctx, id.String()).
		Return(&contract.MachineResponse{ID: id.String(), Label: "found"}, nil)

	rec := testutil.DoJSON(t, e, http.MethodGet, "/api/v1/machines/"+id.String(), nil)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_Get_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)
	id := uuid.New()

	svc.EXPECT().Get(ctx, id.String()).Return(nil, domain.ErrMachineNotFound)

	rec := testutil.DoJSON(t, e, http.MethodGet, "/api/v1/machines/"+id.String(), nil)

	require.Equal(t, http.StatusNotFound, rec.Code)

	var body map[string]any
	testutil.DecodeJSON(t, rec, &body)
	require.Equal(t, "NOT_FOUND", body["error"])
}

func TestHandler_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)

	svc.EXPECT().List(ctx, &contract.ListMachinesRequest{Limit: 10}).
		Return(&contract.ListMachinesResponse{Machines: []contract.MachineResponse{{ID: uuid.New().String(), Label: "listed"}}}, nil)

	rec := testutil.DoJSON(t, e, http.MethodGet, "/api/v1/machines?limit=10", nil)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "listed")
}

func TestHandler_RestockBank(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)
	id := uuid.New()

	svc.EXPECT().RestockBank(ctx, id.String(), &contract.RestockBankRequest{Coins: []int32{25, 25}}).
		Return(&contract.MachineResponse{ID: id.String(), Label: "m1", CoinBank: map[int32]int32{25: 2}}, nil)

	rec := testutil.DoJSON(t, e, http.MethodPost, "/api/v1/machines/"+id.String()+"/bank",
		map[string]any{"coins": []int32{25, 25}})

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_RestockBank_UnsupportedCoin(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)
	id := uuid.New()

	svc.EXPECT().RestockBank(ctx, id.String(), &contract.RestockBankRequest{Coins: []int32{7}}).
		Return(nil, domain.ErrUnsupportedCoin)

	rec := testutil.DoJSON(t, e, http.MethodPost, "/api/v1/machines/"+id.String()+"/bank",
		map[string]any{"coins": []int32{7}})

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var body map[string]any
	testutil.DecodeJSON(t, rec, &body)
	require.Equal(t, "INVALID_INPUT", body["error"])
}

func TestHandler_RestockBank_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)
	id := uuid.New()

	svc.EXPECT().RestockBank(ctx, id.String(), &contract.RestockBankRequest{Coins: []int32{25}}).
		Return(nil, domain.ErrMachineNotFound)

	rec := testutil.DoJSON(t, e, http.MethodPost, "/api/v1/machines/"+id.String()+"/bank",
		map[string]any{"coins": []int32{25}})

	require.Equal(t, http.StatusNotFound, rec.Code)
}
