package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	paymentmock "github.com/zercle/zercle-go-template/domain/payment/mock"
	"github.com/zercle/zercle-go-template/domain/payment/model"
	"github.com/zercle/zercle-go-template/domain/payment/repository"
	"github.com/zercle/zercle-go-template/domain/payment/request"
	"github.com/zercle/zercle-go-template/infrastructure/config"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
)

func setupTestPaymentUseCase(t *testing.T) (*paymentUseCase, *paymentmock.MockRepository) {
	ctrl := gomock.NewController(t)
	mockRepo := paymentmock.NewMockRepository(ctrl)

	logConfig := &config.LoggingConfig{Level: "debug", Format: "console"}
	log := logger.NewLogger(logConfig)

	uc := &paymentUseCase{
		repo: mockRepo,
		log:  log,
	}

	return uc, mockRepo
}

func TestPaymentUseCase_CreatePayment(t *testing.T) {
	uc, mockRepo := setupTestPaymentUseCase(t)
	testPaymentID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	testBookingID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")

	tests := []struct {
		setup   func()
		request request.CreatePayment
		name    string
		errMsg  string
		wantErr bool
	}{
		{
			name: "successful payment creation",
			setup: func() {
				mockRepo.EXPECT().GetByTransactionID(gomock.Any(), "txn_123").Return(nil, repository.ErrPaymentNotFound)
				mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&model.Payment{
					ID:            testPaymentID,
					BookingID:     testBookingID,
					Status:        model.PaymentStatusPending,
					PaymentMethod: "credit_card",
					TransactionID: "txn_123",
				}, nil)
			},
			request: request.CreatePayment{
				BookingID:     testBookingID,
				PaymentMethod: "credit_card",
				TransactionID: "txn_123",
			},
			wantErr: false,
		},
		{
			name: "duplicate transaction ID",
			setup: func() {
				mockRepo.EXPECT().GetByTransactionID(gomock.Any(), "existing_txn").Return(&model.Payment{
					ID: testPaymentID,
				}, nil)
			},
			request: request.CreatePayment{
				BookingID:     testBookingID,
				PaymentMethod: "credit_card",
				TransactionID: "existing_txn",
			},
			wantErr: true,
			errMsg:  ErrDuplicateTransactionID.Error(),
		},
		{
			name: "repository error on duplicate check",
			setup: func() {
				mockRepo.EXPECT().GetByTransactionID(gomock.Any(), "txn_123").Return(nil, assert.AnError)
			},
			request: request.CreatePayment{
				BookingID:     testBookingID,
				PaymentMethod: "credit_card",
				TransactionID: "txn_123",
			},
			wantErr: true,
			errMsg:  "assert.AnError",
		},
		{
			name: "repository error on create",
			setup: func() {
				mockRepo.EXPECT().GetByTransactionID(gomock.Any(), "txn_123").Return(nil, repository.ErrPaymentNotFound)
				mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, assert.AnError)
			},
			request: request.CreatePayment{
				BookingID:     testBookingID,
				PaymentMethod: "credit_card",
				TransactionID: "txn_123",
			},
			wantErr: true,
			errMsg:  "assert.AnError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			result, err := uc.CreatePayment(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.request.BookingID, result.BookingID)
			}
		})
	}
}

func TestPaymentUseCase_GetPayment(t *testing.T) {
	uc, mockRepo := setupTestPaymentUseCase(t)
	testPaymentID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		setup   func()
		name    string
		errMsg  string
		id      uuid.UUID
		wantErr bool
	}{
		{
			name: "successful get payment",
			setup: func() {
				mockRepo.EXPECT().GetByID(gomock.Any(), testPaymentID).Return(&model.Payment{
					ID:     testPaymentID,
					Status: model.PaymentStatusPending,
				}, nil)
			},
			id:      testPaymentID,
			wantErr: false,
		},
		{
			name: "payment not found",
			setup: func() {
				mockRepo.EXPECT().GetByID(gomock.Any(), testPaymentID).Return(nil, repository.ErrPaymentNotFound)
			},
			id:      testPaymentID,
			wantErr: true,
			errMsg:  ErrPaymentNotFound.Error(),
		},
		{
			name: "repository error",
			setup: func() {
				mockRepo.EXPECT().GetByID(gomock.Any(), testPaymentID).Return(nil, assert.AnError)
			},
			id:      testPaymentID,
			wantErr: true,
			errMsg:  "assert.AnError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			result, err := uc.GetPayment(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.id, result.ID)
			}
		})
	}
}

func TestPaymentUseCase_ConfirmPayment(t *testing.T) {
	uc, mockRepo := setupTestPaymentUseCase(t)
	testPaymentID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		setup   func()
		name    string
		errMsg  string
		id      uuid.UUID
		wantErr bool
	}{
		{
			name: "successful payment confirmation",
			setup: func() {
				mockRepo.EXPECT().GetByID(gomock.Any(), testPaymentID).Return(&model.Payment{
					ID:     testPaymentID,
					Status: model.PaymentStatusPending,
				}, nil)
				mockRepo.EXPECT().Confirm(gomock.Any(), testPaymentID).Return(&model.Payment{
					ID:     testPaymentID,
					Status: model.PaymentStatusCompleted,
				}, nil)
			},
			id:      testPaymentID,
			wantErr: false,
		},
		{
			name: "payment not found",
			setup: func() {
				mockRepo.EXPECT().GetByID(gomock.Any(), testPaymentID).Return(nil, repository.ErrPaymentNotFound)
			},
			id:      testPaymentID,
			wantErr: true,
			errMsg:  ErrPaymentNotFound.Error(),
		},
		{
			name: "already completed payment returns success",
			setup: func() {
				mockRepo.EXPECT().GetByID(gomock.Any(), testPaymentID).Return(&model.Payment{
					ID:     testPaymentID,
					Status: model.PaymentStatusCompleted,
				}, nil)
			},
			id:      testPaymentID,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			result, err := uc.ConfirmPayment(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, model.PaymentStatusCompleted, result.Status)
			}
		})
	}
}

func TestPaymentUseCase_RefundPayment(t *testing.T) {
	uc, mockRepo := setupTestPaymentUseCase(t)
	testPaymentID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		setup   func()
		name    string
		req     request.RefundPayment
		errMsg  string
		id      uuid.UUID
		wantErr bool
	}{
		{
			name: "successful payment refund",
			setup: func() {
				mockRepo.EXPECT().GetByID(gomock.Any(), testPaymentID).Return(&model.Payment{
					ID:     testPaymentID,
					Status: model.PaymentStatusCompleted,
				}, nil)
				mockRepo.EXPECT().Refund(gomock.Any(), testPaymentID).Return(&model.Payment{
					ID:     testPaymentID,
					Status: model.PaymentStatusRefunded,
				}, nil)
			},
			id:      testPaymentID,
			req:     request.RefundPayment{},
			wantErr: false,
		},
		{
			name: "payment not found",
			setup: func() {
				mockRepo.EXPECT().GetByID(gomock.Any(), testPaymentID).Return(nil, repository.ErrPaymentNotFound)
			},
			id:      testPaymentID,
			req:     request.RefundPayment{},
			wantErr: true,
			errMsg:  ErrPaymentNotFound.Error(),
		},
		{
			name: "cannot refund pending payment",
			setup: func() {
				mockRepo.EXPECT().GetByID(gomock.Any(), testPaymentID).Return(&model.Payment{
					ID:     testPaymentID,
					Status: model.PaymentStatusPending,
				}, nil)
			},
			id:      testPaymentID,
			req:     request.RefundPayment{},
			wantErr: true,
			errMsg:  ErrCannotRefundPending.Error(),
		},
		{
			name: "cannot refund already refunded payment",
			setup: func() {
				mockRepo.EXPECT().GetByID(gomock.Any(), testPaymentID).Return(&model.Payment{
					ID:     testPaymentID,
					Status: model.PaymentStatusRefunded,
				}, nil)
			},
			id:      testPaymentID,
			req:     request.RefundPayment{},
			wantErr: true,
			errMsg:  ErrCannotRefundRefunded.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			result, err := uc.RefundPayment(context.Background(), tt.id, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, model.PaymentStatusRefunded, result.Status)
			}
		})
	}
}
