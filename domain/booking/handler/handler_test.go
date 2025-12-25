package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	bookingmock "github.com/zercle/zercle-go-template/domain/booking/mock"
	"github.com/zercle/zercle-go-template/domain/booking/model"
	bookingResponse "github.com/zercle/zercle-go-template/domain/booking/response"
	"github.com/zercle/zercle-go-template/domain/booking/usecase"
	"github.com/zercle/zercle-go-template/infrastructure/config"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
	"github.com/zercle/zercle-go-template/pkg/middleware"
	"github.com/zercle/zercle-go-template/pkg/response"
)

const testUserID = "550e8400-e29b-41d4-a716-446655440000"

func setupTestHandler(t *testing.T) (*bookingHandler, *bookingmock.MockUsecase, *echo.Echo) {
	ctrl := gomock.NewController(t)
	mockUsecase := bookingmock.NewMockUsecase(ctrl)

	logConfig := &config.LoggingConfig{Level: "debug", Format: "console"}
	log := logger.NewLogger(logConfig)
	h := &bookingHandler{
		usecase: mockUsecase,
		log:     log,
	}

	e := echo.New()
	e.Validator = middleware.NewCustomValidator()

	return h, mockUsecase, e
}

func TestBookingHandler_CreateBooking(t *testing.T) {
	h, mockUsecase, e := setupTestHandler(t)
	testUserUUID := uuid.MustParse(testUserID)
	testServiceID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	futureTime := time.Now().Add(1 * time.Hour).Format(time.RFC3339)

	tests := []struct {
		setupMock       func()
		name            string
		userID          string
		requestBody     string
		wantStatusField response.Status
		wantStatus      int
	}{
		{
			name: "successful booking creation",
			setupMock: func() {
				mockUsecase.EXPECT().CreateBooking(gomock.Any(), testUserUUID, gomock.Any()).Return(&bookingResponse.BookingResponse{
					ID:        uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
					UserID:    testUserUUID,
					ServiceID: testServiceID,
					Status:    model.BookingStatusPending,
				}, nil)
			},
			userID:          testUserID,
			requestBody:     `{"service_id":"550e8400-e29b-41d4-a716-446655440001","start_time":"` + futureTime + `"}`,
			wantStatus:      http.StatusCreated,
			wantStatusField: response.StatusSuccess,
		},
		{
			name:            "validation error - missing service_id",
			setupMock:       func() {},
			userID:          testUserID,
			requestBody:     `{"start_time":"` + futureTime + `"}`,
			wantStatus:      http.StatusBadRequest,
			wantStatusField: response.StatusFail,
		},
		{
			name:            "validation error - missing start_time",
			setupMock:       func() {},
			userID:          testUserID,
			requestBody:     `{"service_id":"550e8400-e29b-41d4-a716-446655440001"}`,
			wantStatus:      http.StatusBadRequest,
			wantStatusField: response.StatusFail,
		},
		{
			name:            "unauthorized - missing user ID in context",
			setupMock:       func() {},
			userID:          "",
			requestBody:     `{"service_id":"550e8400-e29b-41d4-a716-446655440001","start_time":"` + futureTime + `"}`,
			wantStatus:      http.StatusUnauthorized,
			wantStatusField: response.StatusError,
		},
		{
			name: "booking time in past",
			setupMock: func() {
				mockUsecase.EXPECT().CreateBooking(gomock.Any(), testUserUUID, gomock.Any()).Return(nil, usecase.ErrBookingTimeInPast)
			},
			userID:          testUserID,
			requestBody:     `{"service_id":"550e8400-e29b-41d4-a716-446655440001","start_time":"` + futureTime + `"}`,
			wantStatus:      http.StatusBadRequest,
			wantStatusField: response.StatusFail,
		},
		{
			name: "service not found",
			setupMock: func() {
				mockUsecase.EXPECT().CreateBooking(gomock.Any(), testUserUUID, gomock.Any()).Return(nil, usecase.ErrServiceNotFound)
			},
			userID:          testUserID,
			requestBody:     `{"service_id":"550e8400-e29b-41d4-a716-446655440001","start_time":"` + futureTime + `"}`,
			wantStatus:      http.StatusNotFound,
			wantStatusField: response.StatusError,
		},
		{
			name: "booking conflict",
			setupMock: func() {
				mockUsecase.EXPECT().CreateBooking(gomock.Any(), testUserUUID, gomock.Any()).Return(nil, usecase.ErrBookingConflict)
			},
			userID:          testUserID,
			requestBody:     `{"service_id":"550e8400-e29b-41d4-a716-446655440001","start_time":"` + futureTime + `"}`,
			wantStatus:      http.StatusBadRequest,
			wantStatusField: response.StatusFail,
		},
		{
			name:            "malformed JSON",
			setupMock:       func() {},
			userID:          testUserID,
			requestBody:     `{invalid json}`,
			wantStatus:      http.StatusBadRequest,
			wantStatusField: response.StatusFail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.userID != "" {
				c.Set("user_id", tt.userID)
			}

			_ = h.CreateBooking(c)

			assert.Equal(t, tt.wantStatus, rec.Code)

			var resp response.JSend
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatusField, resp.Status)
		})
	}
}

func TestBookingHandler_GetBooking(t *testing.T) {
	h, mockUsecase, e := setupTestHandler(t)
	testBookingID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		setupMock       func()
		name            string
		bookingID       string
		wantStatusField response.Status
		wantStatus      int
	}{
		{
			name: "successful get booking",
			setupMock: func() {
				mockUsecase.EXPECT().GetBooking(gomock.Any(), testBookingID).Return(&bookingResponse.BookingResponse{
					ID:     testBookingID,
					Status: model.BookingStatusPending,
				}, nil)
			},
			bookingID:       testBookingID.String(),
			wantStatus:      http.StatusOK,
			wantStatusField: response.StatusSuccess,
		},
		{
			name:            "invalid booking ID format",
			setupMock:       func() {},
			bookingID:       "invalid-uuid",
			wantStatus:      http.StatusBadRequest,
			wantStatusField: response.StatusFail,
		},
		{
			name: "booking not found",
			setupMock: func() {
				mockUsecase.EXPECT().GetBooking(gomock.Any(), testBookingID).Return(nil, usecase.ErrBookingNotFound)
			},
			bookingID:       testBookingID.String(),
			wantStatus:      http.StatusNotFound,
			wantStatusField: response.StatusError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest(http.MethodGet, "/api/v1/bookings/"+tt.bookingID, http.NoBody)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.bookingID)

			_ = h.GetBooking(c)

			assert.Equal(t, tt.wantStatus, rec.Code)

			var resp response.JSend
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatusField, resp.Status)
		})
	}
}

func TestBookingHandler_CancelBooking(t *testing.T) {
	h, mockUsecase, e := setupTestHandler(t)
	testUserUUID := uuid.MustParse(testUserID)
	testBookingID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")

	tests := []struct {
		setupMock       func()
		name            string
		userID          string
		bookingID       string
		wantStatusField response.Status
		wantStatus      int
	}{
		{
			name: "successful cancellation",
			setupMock: func() {
				mockUsecase.EXPECT().CancelBooking(gomock.Any(), testBookingID, testUserUUID).Return(&bookingResponse.BookingResponse{
					ID:     testBookingID,
					UserID: testUserUUID,
					Status: model.BookingStatusCancelled,
				}, nil)
			},
			userID:          testUserID,
			bookingID:       testBookingID.String(),
			wantStatus:      http.StatusOK,
			wantStatusField: response.StatusSuccess,
		},
		{
			name:            "unauthorized - missing user ID",
			setupMock:       func() {},
			userID:          "",
			bookingID:       testBookingID.String(),
			wantStatus:      http.StatusUnauthorized,
			wantStatusField: response.StatusError,
		},
		{
			name: "booking not found",
			setupMock: func() {
				mockUsecase.EXPECT().CancelBooking(gomock.Any(), testBookingID, testUserUUID).Return(nil, usecase.ErrBookingNotFound)
			},
			userID:          testUserID,
			bookingID:       testBookingID.String(),
			wantStatus:      http.StatusNotFound,
			wantStatusField: response.StatusError,
		},
		{
			name: "unauthorized access - different user",
			setupMock: func() {
				mockUsecase.EXPECT().CancelBooking(gomock.Any(), testBookingID, testUserUUID).Return(nil, usecase.ErrUnauthorizedAccess)
			},
			userID:          testUserID,
			bookingID:       testBookingID.String(),
			wantStatus:      http.StatusForbidden,
			wantStatusField: response.StatusError,
		},
		{
			name: "cannot cancel completed booking",
			setupMock: func() {
				mockUsecase.EXPECT().CancelBooking(gomock.Any(), testBookingID, testUserUUID).Return(nil, usecase.ErrCannotCancelComplete)
			},
			userID:          testUserID,
			bookingID:       testBookingID.String(),
			wantStatus:      http.StatusBadRequest,
			wantStatusField: response.StatusFail,
		},
		{
			name:            "invalid booking ID format",
			setupMock:       func() {},
			userID:          testUserID,
			bookingID:       "invalid-uuid",
			wantStatus:      http.StatusBadRequest,
			wantStatusField: response.StatusFail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest(http.MethodPut, "/api/v1/bookings/"+tt.bookingID+"/cancel", http.NoBody)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.bookingID)

			if tt.userID != "" {
				c.Set("user_id", tt.userID)
			}

			_ = h.CancelBooking(c)

			assert.Equal(t, tt.wantStatus, rec.Code)

			var resp response.JSend
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatusField, resp.Status)
		})
	}
}

func TestBookingHandler_ListBookingsByUser(t *testing.T) {
	h, mockUsecase, e := setupTestHandler(t)
	testUserUUID := uuid.MustParse(testUserID)

	tests := []struct {
		setupMock       func()
		name            string
		userID          string
		queryParams     string
		wantStatusField response.Status
		wantStatus      int
	}{
		{
			name: "successful list with default pagination",
			setupMock: func() {
				mockUsecase.EXPECT().ListBookingsByUser(gomock.Any(), testUserUUID, 20, 0).Return(&bookingResponse.ListBookingsResponse{
					Bookings: []bookingResponse.BookingResponse{
						{ID: uuid.New(), UserID: testUserUUID, Status: model.BookingStatusPending},
						{ID: uuid.New(), UserID: testUserUUID, Status: model.BookingStatusConfirmed},
					},
					Total: 2,
				}, nil)
			},
			userID:          testUserID,
			queryParams:     "",
			wantStatus:      http.StatusOK,
			wantStatusField: response.StatusSuccess,
		},
		{
			name: "successful list with custom pagination",
			setupMock: func() {
				mockUsecase.EXPECT().ListBookingsByUser(gomock.Any(), testUserUUID, 10, 5).Return(&bookingResponse.ListBookingsResponse{
					Bookings: []bookingResponse.BookingResponse{
						{ID: uuid.New(), UserID: testUserUUID, Status: model.BookingStatusPending},
					},
					Total: 1,
				}, nil)
			},
			userID:          testUserID,
			queryParams:     "?limit=10&offset=5",
			wantStatus:      http.StatusOK,
			wantStatusField: response.StatusSuccess,
		},
		{
			name:            "unauthorized - missing user ID",
			setupMock:       func() {},
			userID:          "",
			queryParams:     "",
			wantStatus:      http.StatusUnauthorized,
			wantStatusField: response.StatusError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest(http.MethodGet, "/api/v1/bookings/user"+tt.queryParams, http.NoBody)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.userID != "" {
				c.Set("user_id", tt.userID)
			}

			_ = h.ListBookingsByUser(c)

			assert.Equal(t, tt.wantStatus, rec.Code)

			var resp response.JSend
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatusField, resp.Status)
		})
	}
}
