package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	bookingmock "github.com/zercle/zercle-go-template/domain/booking/mock"
	"github.com/zercle/zercle-go-template/domain/booking/model"
	"github.com/zercle/zercle-go-template/domain/booking/repository"
	"github.com/zercle/zercle-go-template/domain/booking/request"
	servicemock "github.com/zercle/zercle-go-template/domain/service/mock"
	serviceModel "github.com/zercle/zercle-go-template/domain/service/model"
	serviceRepository "github.com/zercle/zercle-go-template/domain/service/repository"
	"github.com/zercle/zercle-go-template/infrastructure/config"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
)

func setupTestUseCase(t *testing.T) (*bookingUseCase, *bookingmock.MockRepository, *servicemock.MockRepository) {
	ctrl := gomock.NewController(t)
	bookingRepo := bookingmock.NewMockRepository(ctrl)
	serviceRepo := servicemock.NewMockRepository(ctrl)

	logConfig := &config.LoggingConfig{Level: "debug", Format: "console"}
	log := logger.NewLogger(logConfig)

	uc := &bookingUseCase{
		repo:        bookingRepo,
		serviceRepo: serviceRepo,
		log:         log,
	}

	return uc, bookingRepo, serviceRepo
}

func TestBookingUseCase_CreateBooking(t *testing.T) {
	uc, bookingRepo, serviceRepo := setupTestUseCase(t)
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	serviceID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	futureTime := time.Now().Add(1 * time.Hour)

	tests := []struct {
		setupMock func()
		request   request.CreateBooking
		name      string
		errMsg    string
		wantErr   bool
	}{
		{
			name: "successful booking creation",
			setupMock: func() {
				serviceRepo.EXPECT().GetByID(gomock.Any(), serviceID).Return(&serviceModel.Service{
					ID:              serviceID,
					Name:            "Test Service",
					DurationMinutes: 60,
					Price:           100.0,
					IsActive:        true,
				}, nil)
				bookingRepo.EXPECT().CheckConflict(gomock.Any(), serviceID, gomock.Any(), gomock.Any()).Return([]*model.Booking{}, nil)
				bookingRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&model.Booking{
					ID:        uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
					UserID:    userID,
					ServiceID: serviceID,
					StartTime: futureTime,
					Status:    model.BookingStatusPending,
				}, nil)
			},
			request: request.CreateBooking{
				ServiceID: serviceID,
				StartTime: futureTime,
			},
			wantErr: false,
		},
		{
			name:      "booking time in past",
			setupMock: func() {},
			request: request.CreateBooking{
				ServiceID: serviceID,
				StartTime: time.Now().Add(-1 * time.Hour),
			},
			wantErr: true,
			errMsg:  ErrBookingTimeInPast.Error(),
		},
		{
			name: "service not found",
			setupMock: func() {
				serviceRepo.EXPECT().GetByID(gomock.Any(), serviceID).Return(nil, serviceRepository.ErrServiceNotFound)
			},
			request: request.CreateBooking{
				ServiceID: serviceID,
				StartTime: futureTime,
			},
			wantErr: true,
			errMsg:  ErrServiceNotFound.Error(),
		},
		{
			name: "service not active",
			setupMock: func() {
				serviceRepo.EXPECT().GetByID(gomock.Any(), serviceID).Return(&serviceModel.Service{
					ID:       serviceID,
					Name:     "Inactive Service",
					IsActive: false,
				}, nil)
			},
			request: request.CreateBooking{
				ServiceID: serviceID,
				StartTime: futureTime,
			},
			wantErr: true,
			errMsg:  "not available for booking",
		},
		{
			name: "booking conflict",
			setupMock: func() {
				serviceRepo.EXPECT().GetByID(gomock.Any(), serviceID).Return(&serviceModel.Service{
					ID:              serviceID,
					Name:            "Test Service",
					DurationMinutes: 60,
					Price:           100.0,
					IsActive:        true,
				}, nil)
				bookingRepo.EXPECT().CheckConflict(gomock.Any(), serviceID, gomock.Any(), gomock.Any()).Return([]*model.Booking{{ID: uuid.New()}}, nil)
			},
			request: request.CreateBooking{
				ServiceID: serviceID,
				StartTime: futureTime,
			},
			wantErr: true,
			errMsg:  ErrBookingConflict.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			result, err := uc.CreateBooking(context.Background(), userID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestBookingUseCase_GetBooking(t *testing.T) {
	uc, bookingRepo, _ := setupTestUseCase(t)
	bookingID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		setupMock func()
		name      string
		errMsg    string
		id        uuid.UUID
		wantErr   bool
	}{
		{
			name: "successful get booking",
			setupMock: func() {
				bookingRepo.EXPECT().GetByID(gomock.Any(), bookingID).Return(&model.Booking{
					ID:     bookingID,
					Status: model.BookingStatusPending,
				}, nil)
			},
			id:      bookingID,
			wantErr: false,
		},
		{
			name: "booking not found",
			setupMock: func() {
				bookingRepo.EXPECT().GetByID(gomock.Any(), bookingID).Return(nil, repository.ErrBookingNotFound)
			},
			id:      bookingID,
			wantErr: true,
			errMsg:  ErrBookingNotFound.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			result, err := uc.GetBooking(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestBookingUseCase_CancelBooking(t *testing.T) {
	uc, bookingRepo, _ := setupTestUseCase(t)
	bookingID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")

	tests := []struct {
		setupMock func()
		name      string
		errMsg    string
		wantErr   bool
	}{
		{
			name: "successful cancellation",
			setupMock: func() {
				bookingRepo.EXPECT().GetByID(gomock.Any(), bookingID).Return(&model.Booking{
					ID:     bookingID,
					UserID: userID,
					Status: model.BookingStatusPending,
				}, nil)
				bookingRepo.EXPECT().Cancel(gomock.Any(), bookingID).Return(&model.Booking{
					ID:     bookingID,
					UserID: userID,
					Status: model.BookingStatusCancelled,
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "unauthorized access",
			setupMock: func() {
				bookingRepo.EXPECT().GetByID(gomock.Any(), bookingID).Return(&model.Booking{
					ID:     bookingID,
					UserID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
					Status: model.BookingStatusPending,
				}, nil)
			},
			wantErr: true,
			errMsg:  ErrUnauthorizedAccess.Error(),
		},
		{
			name: "cannot cancel completed booking",
			setupMock: func() {
				bookingRepo.EXPECT().GetByID(gomock.Any(), bookingID).Return(&model.Booking{
					ID:     bookingID,
					UserID: userID,
					Status: model.BookingStatusCompleted,
				}, nil)
			},
			wantErr: true,
			errMsg:  ErrCannotCancelComplete.Error(),
		},
		{
			name: "already canceled",
			setupMock: func() {
				bookingRepo.EXPECT().GetByID(gomock.Any(), bookingID).Return(&model.Booking{
					ID:     bookingID,
					UserID: userID,
					Status: model.BookingStatusCancelled,
				}, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			result, err := uc.CancelBooking(context.Background(), bookingID, userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestBookingUseCase_UpdateBookingStatus(t *testing.T) {
	uc, bookingRepo, _ := setupTestUseCase(t)
	bookingID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		setupMock func()
		name      string
		request   request.UpdateBookingStatus
		errMsg    string
		wantErr   bool
	}{
		{
			name: "successful status update - pending to confirmed",
			setupMock: func() {
				bookingRepo.EXPECT().GetByID(gomock.Any(), bookingID).Return(&model.Booking{
					ID:     bookingID,
					Status: model.BookingStatusPending,
				}, nil)
				bookingRepo.EXPECT().UpdateStatus(gomock.Any(), bookingID, model.BookingStatusConfirmed).Return(&model.Booking{
					ID:     bookingID,
					Status: model.BookingStatusConfirmed,
				}, nil)
			},
			request: request.UpdateBookingStatus{
				Status: model.BookingStatusConfirmed,
			},
			wantErr: false,
		},
		{
			name: "invalid status transition - completed to pending",
			setupMock: func() {
				bookingRepo.EXPECT().GetByID(gomock.Any(), bookingID).Return(&model.Booking{
					ID:     bookingID,
					Status: model.BookingStatusCompleted,
				}, nil)
			},
			request: request.UpdateBookingStatus{
				Status: model.BookingStatusPending,
			},
			wantErr: true,
			errMsg:  ErrInvalidStatus.Error(),
		},
		{
			name: "booking not found",
			setupMock: func() {
				bookingRepo.EXPECT().GetByID(gomock.Any(), bookingID).Return(nil, repository.ErrBookingNotFound)
			},
			request: request.UpdateBookingStatus{
				Status: model.BookingStatusConfirmed,
			},
			wantErr: true,
			errMsg:  ErrBookingNotFound.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			result, err := uc.UpdateBookingStatus(context.Background(), bookingID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestIsValidStatusTransition(t *testing.T) {
	tests := []struct {
		name     string
		current  model.BookingStatus
		new      model.BookingStatus
		expected bool
	}{
		{"pending to confirmed", model.BookingStatusPending, model.BookingStatusConfirmed, true},
		{"pending to canceled", model.BookingStatusPending, model.BookingStatusCancelled, true},
		{"confirmed to completed", model.BookingStatusConfirmed, model.BookingStatusCompleted, true},
		{"confirmed to canceled", model.BookingStatusConfirmed, model.BookingStatusCancelled, true},
		{"completed to pending", model.BookingStatusCompleted, model.BookingStatusPending, false},
		{"canceled to confirmed", model.BookingStatusCancelled, model.BookingStatusConfirmed, false},
		{"pending to completed", model.BookingStatusPending, model.BookingStatusCompleted, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidStatusTransition(tt.current, tt.new)
			assert.Equal(t, tt.expected, result)
		})
	}
}
