package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	servicemock "github.com/zercle/zercle-go-template/domain/service/mock"
	serviceResponse "github.com/zercle/zercle-go-template/domain/service/response"
	"github.com/zercle/zercle-go-template/domain/service/usecase"
	"github.com/zercle/zercle-go-template/infrastructure/config"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
	"github.com/zercle/zercle-go-template/pkg/middleware"
	"github.com/zercle/zercle-go-template/pkg/response"
)

func setupTestServiceHandler(t *testing.T) (*serviceHandler, *servicemock.MockUsecase, *echo.Echo) {
	ctrl := gomock.NewController(t)
	mockUsecase := servicemock.NewMockUsecase(ctrl)

	logConfig := &config.LoggingConfig{Level: "debug", Format: "console"}
	log := logger.NewLogger(logConfig)
	h := &serviceHandler{
		usecase: mockUsecase,
		log:     log,
	}

	e := echo.New()
	e.Validator = middleware.NewCustomValidator()

	return h, mockUsecase, e
}

func TestServiceHandler_CreateService(t *testing.T) {
	h, mockUsecase, e := setupTestServiceHandler(t)
	testServiceID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		setupMock       func()
		name            string
		requestBody     string
		wantStatusField response.Status
		wantStatus      int
	}{
		{
			name: "successful service creation",
			setupMock: func() {
				mockUsecase.EXPECT().CreateService(gomock.Any(), gomock.Any()).Return(&serviceResponse.ServiceResponse{
					ID:              testServiceID,
					Name:            "Test Service",
					Description:     "Test Description",
					DurationMinutes: 60,
					Price:           100.0,
					MaxCapacity:     10,
					IsActive:        true,
				}, nil)
			},
			requestBody:     `{"name":"Test Service","description":"Test Description","duration_minutes":60,"price":100.0,"max_capacity":10,"is_active":true}`,
			wantStatus:      http.StatusCreated,
			wantStatusField: response.StatusSuccess,
		},
		{
			name:            "validation error - missing name",
			setupMock:       func() {},
			requestBody:     `{"description":"Test Description","duration_minutes":60,"price":100.0,"max_capacity":10}`,
			wantStatus:      http.StatusBadRequest,
			wantStatusField: response.StatusFail,
		},
		{
			name:            "validation error - missing price",
			setupMock:       func() {},
			requestBody:     `{"name":"Test Service","description":"Test Description","duration_minutes":60,"max_capacity":10}`,
			wantStatus:      http.StatusBadRequest,
			wantStatusField: response.StatusFail,
		},
		{
			name:            "validation error - invalid price (zero fails validation)",
			setupMock:       func() {},
			requestBody:     `{"name":"Test Service","description":"Test Description","duration_minutes":60,"price":0,"max_capacity":10}`,
			wantStatus:      http.StatusBadRequest,
			wantStatusField: response.StatusFail,
		},
		{
			name:            "validation error - invalid duration (zero fails validation)",
			setupMock:       func() {},
			requestBody:     `{"name":"Test Service","description":"Test Description","duration_minutes":0,"price":100,"max_capacity":10}`,
			wantStatus:      http.StatusBadRequest,
			wantStatusField: response.StatusFail,
		},
		{
			name:            "malformed JSON",
			setupMock:       func() {},
			requestBody:     `{invalid json}`,
			wantStatus:      http.StatusBadRequest,
			wantStatusField: response.StatusFail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest(http.MethodPost, "/services", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			_ = h.CreateService(c)

			assert.Equal(t, tt.wantStatus, rec.Code)

			var resp response.JSend
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatusField, resp.Status)
		})
	}
}

func TestServiceHandler_GetService(t *testing.T) {
	h, mockUsecase, e := setupTestServiceHandler(t)
	testServiceID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		setupMock       func()
		name            string
		serviceID       string
		wantStatusField response.Status
		wantStatus      int
	}{
		{
			name: "successful get service",
			setupMock: func() {
				mockUsecase.EXPECT().GetService(gomock.Any(), testServiceID).Return(&serviceResponse.ServiceResponse{
					ID:              testServiceID,
					Name:            "Test Service",
					DurationMinutes: 60,
					Price:           100.0,
				}, nil)
			},
			serviceID:       testServiceID.String(),
			wantStatus:      http.StatusOK,
			wantStatusField: response.StatusSuccess,
		},
		{
			name:            "invalid service ID format",
			setupMock:       func() {},
			serviceID:       "invalid-uuid",
			wantStatus:      http.StatusBadRequest,
			wantStatusField: response.StatusFail,
		},
		{
			name: "service not found",
			setupMock: func() {
				mockUsecase.EXPECT().GetService(gomock.Any(), testServiceID).Return(nil, usecase.ErrServiceNotFound)
			},
			serviceID:       testServiceID.String(),
			wantStatus:      http.StatusNotFound,
			wantStatusField: response.StatusError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest(http.MethodGet, "/services/"+tt.serviceID, http.NoBody)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.serviceID)

			_ = h.GetService(c)

			assert.Equal(t, tt.wantStatus, rec.Code)

			var resp response.JSend
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatusField, resp.Status)
		})
	}
}

func TestServiceHandler_ListServices(t *testing.T) {
	h, mockUsecase, e := setupTestServiceHandler(t)

	tests := []struct {
		setupMock       func()
		name            string
		queryParams     string
		wantStatusField response.Status
		wantStatus      int
	}{
		{
			name: "successful list with default pagination",
			setupMock: func() {
				mockUsecase.EXPECT().ListServices(gomock.Any(), gomock.Any()).Return(&serviceResponse.ListServicesResponse{
					Services: []serviceResponse.ServiceResponse{
						{ID: uuid.New(), Name: "Service 1", Price: 100.0},
						{ID: uuid.New(), Name: "Service 2", Price: 200.0},
					},
					Total: 2,
				}, nil)
			},
			queryParams:     "",
			wantStatus:      http.StatusOK,
			wantStatusField: response.StatusSuccess,
		},
		{
			name: "successful list with custom pagination",
			setupMock: func() {
				mockUsecase.EXPECT().ListServices(gomock.Any(), gomock.Any()).Return(&serviceResponse.ListServicesResponse{
					Services: []serviceResponse.ServiceResponse{
						{ID: uuid.New(), Name: "Service 1", Price: 100.0},
					},
					Total: 1,
				}, nil)
			},
			queryParams:     "?limit=10&offset=5",
			wantStatus:      http.StatusOK,
			wantStatusField: response.StatusSuccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest(http.MethodGet, "/services"+tt.queryParams, http.NoBody)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			_ = h.ListServices(c)

			assert.Equal(t, tt.wantStatus, rec.Code)

			var resp response.JSend
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatusField, resp.Status)
		})
	}
}
