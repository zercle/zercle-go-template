//go:build unit

// STUB FEATURE — delete internal/features/example to start your project.

package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"uuid"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/zercle/zercle-go-template/internal/features/example/contract"
	"github.com/zercle/zercle-go-template/internal/features/example/domain"
	"github.com/zercle/zercle-go-template/internal/features/example/handler"
	"github.com/zercle/zercle-go-template/internal/features/example/usecase/mock"
	apperrors "github.com/zercle/zercle-go-template/internal/infrastructure/errors"
)

// registerSentinelsOnce registers the example feature's domain sentinels exactly
// once per package. setupTest runs for every parallel test; the registry itself
// is mutex-protected, but the repeated calls were redundant, so guard them.
var registerSentinelsOnce sync.Once

func setupTest(t *testing.T) (*echo.Echo, *mock.MockService) {
	t.Helper()

	registerSentinelsOnce.Do(func() {
		apperrors.RegisterSentinel(domain.ErrItemNotFound, apperrors.ErrNotFound)
		apperrors.RegisterSentinel(domain.ErrInvalidName, apperrors.ErrInvalidInput)
		apperrors.RegisterSentinel(domain.ErrInvalidID, apperrors.ErrInvalidInput)
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

	svc.EXPECT().Create(ctx, &contract.CreateItemRequest{Name: "stub"}).
		Return(&contract.ItemResponse{ID: id.String(), Name: "stub"}, nil)

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

	svc.EXPECT().Get(ctx, id.String()).Return(&contract.ItemResponse{ID: id.String(), Name: "found"}, nil)

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

	svc.EXPECT().Get(ctx, id.String()).Return(nil, domain.ErrItemNotFound)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/items/"+id.String(), nil)

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "NOT_FOUND", body["error"])
}

func TestHandler_Get_InvalidID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)

	svc.EXPECT().Get(ctx, "not-a-uuid").Return(nil, domain.ErrInvalidID)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/items/not-a-uuid", nil)

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "INVALID_INPUT", body["error"])
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

	svc.EXPECT().Create(ctx, &contract.CreateItemRequest{Name: "stub"}).Return(nil, errors.New("boom"))

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

	svc.EXPECT().List(ctx, &contract.ListItemsRequest{}).
		Return(&contract.ListItemsResponse{Items: []contract.ItemResponse{{ID: uuid.New().String(), Name: "default"}}}, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/items", nil)

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

// TestHandler_Create_NameAtConfiguredLimit proves the wire layer does not cap
// the name at a hardcoded length: a name permitted by the configured maximum
// must reach the application service instead of being rejected by a stale
// validate:"max=255" tag.
func TestHandler_Create_NameAtConfiguredLimit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)

	name := strings.Repeat("a", 1000) // > old hardcoded 255, < configured cap 4096
	svc.EXPECT().Create(ctx, &contract.CreateItemRequest{Name: name}).
		Return(&contract.ItemResponse{ID: uuid.New().String(), Name: name}, nil)

	body, err := json.Marshal(map[string]string{"name": name})
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/items", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code, "configured limit must govern, not a hardcoded tag")
}

// TestHandler_List_LimitAboveOldHardcodedCap proves the query limit is not
// rejected by a hardcoded max=100 tag; clamping happens in the usecase.
func TestHandler_List_LimitAboveOldHardcodedCap(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	e, svc := setupTest(t)

	svc.EXPECT().List(ctx, &contract.ListItemsRequest{Limit: 500, Offset: 0}).
		Return(&contract.ListItemsResponse{}, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/items?limit=500", nil)

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, "configured limit must govern, not a hardcoded tag")
}
