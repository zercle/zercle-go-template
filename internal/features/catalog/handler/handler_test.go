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

	"github.com/zercle/zercle-go-template/internal/features/catalog/contract"
	"github.com/zercle/zercle-go-template/internal/features/catalog/domain"
	"github.com/zercle/zercle-go-template/internal/features/catalog/handler"
	"github.com/zercle/zercle-go-template/internal/features/catalog/usecase/mock"
	apperrors "github.com/zercle/zercle-go-template/internal/infrastructure/errors"
	"github.com/zercle/zercle-go-template/internal/testutil"
)

// registerSentinelsOnce registers the catalog feature's domain sentinels
// exactly once per test binary.
var registerSentinelsOnce sync.Once

func setupTest(t *testing.T) (*echo.Echo, *mock.MockService) {
	t.Helper()

	registerSentinelsOnce.Do(func() {
		apperrors.RegisterSentinel(domain.ErrProductNotFound, apperrors.ErrNotFound)
		apperrors.RegisterSentinel(domain.ErrInvalidID, apperrors.ErrInvalidInput)
		apperrors.RegisterSentinel(domain.ErrInvalidProductName, apperrors.ErrInvalidInput)
		apperrors.RegisterSentinel(domain.ErrInvalidPrice, apperrors.ErrInvalidInput)
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

	svc.EXPECT().Create(ctx, &contract.CreateProductRequest{Name: "widget", PriceCents: 100, Stock: 5}).
		Return(&contract.ProductResponse{ID: id.String(), Name: "widget", PriceCents: 100, Stock: 5}, nil)

	rec := testutil.DoJSON(t, e, http.MethodPost, "/api/v1/products",
		map[string]any{"name": "widget", "price_cents": 100, "stock": 5})

	require.Equal(t, http.StatusCreated, rec.Code)
	require.Contains(t, rec.Body.String(), "widget")
}

func TestHandler_Create_InvalidJSON(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, _ := setupTest(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/products", bytes.NewReader([]byte(`{`)))
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

	rec := testutil.DoJSON(t, e, http.MethodPost, "/api/v1/products",
		map[string]any{"name": "", "price_cents": 100, "stock": 0})

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
		Return(&contract.ProductResponse{ID: id.String(), Name: "found"}, nil)

	rec := testutil.DoJSON(t, e, http.MethodGet, "/api/v1/products/"+id.String(), nil)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_Get_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)
	id := uuid.New()

	svc.EXPECT().Get(ctx, id.String()).Return(nil, domain.ErrProductNotFound)

	rec := testutil.DoJSON(t, e, http.MethodGet, "/api/v1/products/"+id.String(), nil)

	require.Equal(t, http.StatusNotFound, rec.Code)

	var body map[string]any
	testutil.DecodeJSON(t, rec, &body)
	require.Equal(t, "NOT_FOUND", body["error"])
}

func TestHandler_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)

	svc.EXPECT().List(ctx, &contract.ListProductsRequest{Limit: 10}).
		Return(&contract.ListProductsResponse{Products: []contract.ProductResponse{{ID: uuid.New().String(), Name: "listed"}}}, nil)

	rec := testutil.DoJSON(t, e, http.MethodGet, "/api/v1/products?limit=10", nil)

	require.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	testutil.DecodeJSON(t, rec, &body)
	require.Contains(t, rec.Body.String(), "listed")
}
