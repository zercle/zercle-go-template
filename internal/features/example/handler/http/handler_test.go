//go:build unit

// STUB FEATURE — delete internal/features/example to start your project.

package httphandler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/zercle/zercle-go-template/internal/features/example/domain"
	httphandler "github.com/zercle/zercle-go-template/internal/features/example/handler/http"
	"github.com/zercle/zercle-go-template/internal/features/example/service/mock"
	sharederrors "github.com/zercle/zercle-go-template/internal/shared/errors"
)

func setupTest(t *testing.T) (*echo.Echo, *mock.MockService) {
	t.Helper()

	sharederrors.RegisterSentinel(domain.ErrItemNotFound, sharederrors.ErrNotFound)
	sharederrors.RegisterSentinel(domain.ErrInvalidName, sharederrors.ErrInvalidInput)
	sharederrors.RegisterSentinel(domain.ErrInvalidID, sharederrors.ErrInvalidInput)

	e := echo.New()
	e.Validator = newValidator(t)
	svc := mock.NewMockService(gomock.NewController(t))
	h := httphandler.New(svc)

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

	svc.EXPECT().Create(ctx, "stub").Return(&domain.Item{ID: id, Name: "stub"}, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/items", bytes.NewReader([]byte(`{"name":"stub"}`)))
	req.Header.Set("Content-Type", "application/json")

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	require.Contains(t, rec.Body.String(), "stub")
}

func TestHandler_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)
	id := uuid.New()

	svc.EXPECT().Get(ctx, id).Return(&domain.Item{ID: id, Name: "found"}, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/items/"+id.String(), nil)

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_Get_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)
	id := uuid.New()

	svc.EXPECT().Get(ctx, id).Return(nil, domain.ErrItemNotFound)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/items/"+id.String(), nil)

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "NOT_FOUND", body["error"])
}

func TestHandler_Create_EmptyName(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, _ := setupTest(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/items", bytes.NewReader([]byte(`{"name":""}`)))
	req.Header.Set("Content-Type", "application/json")

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "INVALID_INPUT", body["error"])
}

func TestHandler_Create_ServiceError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)

	svc.EXPECT().Create(ctx, "stub").Return(nil, errors.New("boom"))

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/items", bytes.NewReader([]byte(`{"name":"stub"}`)))
	req.Header.Set("Content-Type", "application/json")

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHandler_List_NoQueryParams(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)

	svc.EXPECT().List(ctx, int32(0), int32(0)).Return([]domain.Item{{ID: uuid.New(), Name: "default"}}, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/items", nil)

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}
