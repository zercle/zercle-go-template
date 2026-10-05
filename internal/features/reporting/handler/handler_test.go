//go:build unit

package handler_test

import (
	"context"
	"net/http"
	"sync"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/zercle/zercle-go-template/internal/features/reporting/contract"
	"github.com/zercle/zercle-go-template/internal/features/reporting/domain"
	"github.com/zercle/zercle-go-template/internal/features/reporting/handler"
	"github.com/zercle/zercle-go-template/internal/features/reporting/usecase/mock"
	apperrors "github.com/zercle/zercle-go-template/internal/infrastructure/errors"
	"github.com/zercle/zercle-go-template/internal/testutil"
)

// registerSentinelsOnce registers the reporting feature's domain sentinels
// exactly once per test binary.
var registerSentinelsOnce sync.Once

func setupTest(t *testing.T) (*echo.Echo, *mock.MockService) {
	t.Helper()

	registerSentinelsOnce.Do(func() {
		apperrors.RegisterSentinel(domain.ErrInvalidTopMachines, apperrors.ErrInvalidInput)
	})

	e := echo.New()
	e.Validator = &validatorAdapter{v: validator.New()}
	svc := mock.NewMockService(gomock.NewController(t))
	h := handler.New(svc)

	h.Register(e.Group("/api/v1"))

	return e, svc
}

type validatorAdapter struct {
	v *validator.Validate
}

func (v *validatorAdapter) Validate(i any) error {
	return v.v.Struct(i)
}

func TestHandler_Summary(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)

	svc.EXPECT().Summary(ctx, &contract.SummaryRequest{TopMachines: 3}).
		Return(&contract.SummaryResponse{
			Catalog:  contract.CatalogStats{ProductCount: 2, TotalStock: 8},
			Machines: contract.MachineStats{MachineCount: 1, TotalCoinBankCents: 350},
			Sales:    contract.SalesStats{PurchaseCount: 3, RevenueCents: 450},
			TopMachines: []contract.MachineSales{
				{MachineID: "m1", Label: "lobby", PurchaseCount: 3, RevenueCents: 450},
			},
		}, nil)

	rec := testutil.DoJSON(t, e, http.MethodGet, "/api/v1/reports/summary?top=3", nil)

	require.Equal(t, http.StatusOK, rec.Code)

	var body contract.SummaryResponse
	testutil.DecodeJSON(t, rec, &body)
	require.EqualValues(t, 2, body.Catalog.ProductCount)
	require.Len(t, body.TopMachines, 1)
	require.Equal(t, "lobby", body.TopMachines[0].Label)
}

func TestHandler_Summary_DefaultTop(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)

	svc.EXPECT().Summary(ctx, &contract.SummaryRequest{}).
		Return(&contract.SummaryResponse{TopMachines: []contract.MachineSales{}}, nil)

	rec := testutil.DoJSON(t, e, http.MethodGet, "/api/v1/reports/summary", nil)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_Summary_InvalidQuery(t *testing.T) {
	t.Parallel()

	e, _ := setupTest(t)

	rec := testutil.DoJSON(t, e, http.MethodGet, "/api/v1/reports/summary?top=-1", nil)

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var body map[string]any
	testutil.DecodeJSON(t, rec, &body)
	require.Equal(t, "INVALID_INPUT", body["error"])
}

func TestHandler_Summary_ServiceError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)

	svc.EXPECT().Summary(ctx, &contract.SummaryRequest{TopMachines: 99}).
		Return(nil, domain.ErrInvalidTopMachines)

	rec := testutil.DoJSON(t, e, http.MethodGet, "/api/v1/reports/summary?top=99", nil)

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var body map[string]any
	testutil.DecodeJSON(t, rec, &body)
	require.Equal(t, "INVALID_INPUT", body["error"])
}
