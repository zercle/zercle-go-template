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

	paymentmock "github.com/zercle/zercle-go-template/domain/payment/mock"
	"github.com/zercle/zercle-go-template/domain/payment/model"
	paymentResponse "github.com/zercle/zercle-go-template/domain/payment/response"
	"github.com/zercle/zercle-go-template/domain/payment/usecase"
	"github.com/zercle/zercle-go-template/infrastructure/config"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
	"github.com/zercle/zercle-go-template/pkg/middleware"
	"github.com/zercle/zercle-go-template/pkg/response"
)

func setupTestPaymentHandler(t *testing.T) (*paymentHandler, *paymentmock.MockUsecase, *echo.Echo) {
	ctrl := gomock.NewController(t)
	mockUsecase := paymentmock.NewMockUsecase(ctrl)

	logConfig := &config.LoggingConfig{Level: "debug", Format: "console"}
	log := logger.NewLogger(logConfig)
	h := &paymentHandler{
		usecase: mockUsecase,
		log:     log,
	}

	e := echo.New()
	e.Validator = middleware.NewCustomValidator()

	return h, mockUsecase, e
}

func TestPaymentHandler_CreatePayment(t *testing.T) {
	h, mockUsecase, e := setupTestPaymentHandler(t)
	testPaymentID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	testBookingID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")

	tests := []struct {
		setupMock       func()
		name            string
		requestBody     string
		wantStatusField response.Status
		wantStatus      int
	}{
		{
			name: "successful payment creation",
			setupMock: func() {
				mockUsecase.EXPECT().CreatePayment(gomock.Any(), gomock.Any()).Return(&paymentResponse.PaymentResponse{
					ID:            testPaymentID,
					BookingID:     testBookingID,
					Status:        model.PaymentStatusPending,
					PaymentMethod: "credit_card",
					TransactionID: "txn_123",
				}, nil)
			},
			requestBody:     `{"booking_id":"550e8400-e29b-41d4-a716-446655440001","payment_method":"credit_card","transaction_id":"txn_123"}`,
			wantStatus:      http.StatusCreated,
			wantStatusField: response.StatusSuccess,
		},
		{
			name:            "validation error - missing booking_id",
			setupMock:       func() {},
			requestBody:     `{"payment_method":"credit_card"}`,
			wantStatus:      http.StatusBadRequest,
			wantStatusField: response.StatusFail,
		},
		{
			name:            "validation error - missing payment_method",
			setupMock:       func() {},
			requestBody:     `{"booking_id":"550e8400-e29b-41d4-a716-446655440001"}`,
			wantStatus:      http.StatusBadRequest,
			wantStatusField: response.StatusFail,
		},
		{
			name: "duplicate transaction ID",
			setupMock: func() {
				mockUsecase.EXPECT().CreatePayment(gomock.Any(), gomock.Any()).Return(nil, usecase.ErrDuplicateTransactionID)
			},
			requestBody:     `{"booking_id":"550e8400-e29b-41d4-a716-446655440001","payment_method":"credit_card","transaction_id":"existing_txn"}`,
			wantStatus:      http.StatusBadRequest,
			wantStatusField: response.StatusFail,
		},
		{
			name: "booking not found",
			setupMock: func() {
				mockUsecase.EXPECT().CreatePayment(gomock.Any(), gomock.Any()).Return(nil, usecase.ErrBookingNotFound)
			},
			requestBody:     `{"booking_id":"550e8400-e29b-41d4-a716-446655440001","payment_method":"credit_card"}`,
			wantStatus:      http.StatusNotFound,
			wantStatusField: response.StatusError,
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

			req := httptest.NewRequest(http.MethodPost, "/payments", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			_ = h.CreatePayment(c)

			assert.Equal(t, tt.wantStatus, rec.Code)

			var resp response.JSend
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatusField, resp.Status)
		})
	}
}

func TestPaymentHandler_GetPayment(t *testing.T) {
	h, mockUsecase, e := setupTestPaymentHandler(t)
	testPaymentID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		setupMock       func()
		name            string
		paymentID       string
		wantStatusField response.Status
		wantStatus      int
	}{
		{
			name: "successful get payment",
			setupMock: func() {
				mockUsecase.EXPECT().GetPayment(gomock.Any(), testPaymentID).Return(&paymentResponse.PaymentResponse{
					ID:        testPaymentID,
					BookingID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
					Status:    model.PaymentStatusPending,
				}, nil)
			},
			paymentID:       testPaymentID.String(),
			wantStatus:      http.StatusOK,
			wantStatusField: response.StatusSuccess,
		},
		{
			name:            "invalid payment ID format",
			setupMock:       func() {},
			paymentID:       "invalid-uuid",
			wantStatus:      http.StatusBadRequest,
			wantStatusField: response.StatusFail,
		},
		{
			name: "payment not found",
			setupMock: func() {
				mockUsecase.EXPECT().GetPayment(gomock.Any(), testPaymentID).Return(nil, usecase.ErrPaymentNotFound)
			},
			paymentID:       testPaymentID.String(),
			wantStatus:      http.StatusNotFound,
			wantStatusField: response.StatusError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest(http.MethodGet, "/payments/"+tt.paymentID, http.NoBody)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.paymentID)

			_ = h.GetPayment(c)

			assert.Equal(t, tt.wantStatus, rec.Code)

			var resp response.JSend
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatusField, resp.Status)
		})
	}
}

func TestPaymentHandler_ConfirmPayment(t *testing.T) {
	h, mockUsecase, e := setupTestPaymentHandler(t)
	testPaymentID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		setupMock       func()
		name            string
		paymentID       string
		requestBody     string
		wantStatusField response.Status
		wantStatus      int
	}{
		{
			name: "successful payment confirmation",
			setupMock: func() {
				mockUsecase.EXPECT().ConfirmPayment(gomock.Any(), testPaymentID).Return(&paymentResponse.PaymentResponse{
					ID:     testPaymentID,
					Status: model.PaymentStatusCompleted,
				}, nil)
			},
			paymentID:       testPaymentID.String(),
			requestBody:     `{}`,
			wantStatus:      http.StatusOK,
			wantStatusField: response.StatusSuccess,
		},
		{
			name:            "invalid payment ID format",
			setupMock:       func() {},
			paymentID:       "invalid-uuid",
			requestBody:     `{}`,
			wantStatus:      http.StatusBadRequest,
			wantStatusField: response.StatusFail,
		},
		{
			name: "payment not found",
			setupMock: func() {
				mockUsecase.EXPECT().ConfirmPayment(gomock.Any(), testPaymentID).Return(nil, usecase.ErrPaymentNotFound)
			},
			paymentID:       testPaymentID.String(),
			requestBody:     `{}`,
			wantStatus:      http.StatusNotFound,
			wantStatusField: response.StatusError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest(http.MethodPut, "/payments/"+tt.paymentID+"/confirm", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.paymentID)

			_ = h.ConfirmPayment(c)

			assert.Equal(t, tt.wantStatus, rec.Code)

			var resp response.JSend
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatusField, resp.Status)
		})
	}
}

func TestPaymentHandler_RefundPayment(t *testing.T) {
	h, mockUsecase, e := setupTestPaymentHandler(t)
	testPaymentID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		setupMock       func()
		name            string
		paymentID       string
		requestBody     string
		wantStatusField response.Status
		wantStatus      int
	}{
		{
			name: "successful payment refund",
			setupMock: func() {
				mockUsecase.EXPECT().RefundPayment(gomock.Any(), testPaymentID, gomock.Any()).Return(&paymentResponse.PaymentResponse{
					ID:     testPaymentID,
					Status: model.PaymentStatusRefunded,
				}, nil)
			},
			paymentID:       testPaymentID.String(),
			requestBody:     `{}`,
			wantStatus:      http.StatusOK,
			wantStatusField: response.StatusSuccess,
		},
		{
			name:            "invalid payment ID format",
			setupMock:       func() {},
			paymentID:       "invalid-uuid",
			requestBody:     `{}`,
			wantStatus:      http.StatusBadRequest,
			wantStatusField: response.StatusFail,
		},
		{
			name: "payment not found",
			setupMock: func() {
				mockUsecase.EXPECT().RefundPayment(gomock.Any(), testPaymentID, gomock.Any()).Return(nil, usecase.ErrPaymentNotFound)
			},
			paymentID:       testPaymentID.String(),
			requestBody:     `{}`,
			wantStatus:      http.StatusNotFound,
			wantStatusField: response.StatusError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest(http.MethodPut, "/payments/"+tt.paymentID+"/refund", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.paymentID)

			_ = h.RefundPayment(c)

			assert.Equal(t, tt.wantStatus, rec.Code)

			var resp response.JSend
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatusField, resp.Status)
		})
	}
}
