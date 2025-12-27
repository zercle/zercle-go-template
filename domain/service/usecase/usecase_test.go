package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	servicemock "github.com/zercle/zercle-go-template/domain/service/mock"
	"github.com/zercle/zercle-go-template/domain/service/model"
	"github.com/zercle/zercle-go-template/domain/service/repository"
	"github.com/zercle/zercle-go-template/domain/service/request"
	"github.com/zercle/zercle-go-template/infrastructure/config"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
)

func setupTestServiceUseCase(t *testing.T) (*serviceUseCase, *servicemock.MockRepository) {
	ctrl := gomock.NewController(t)
	mockRepo := servicemock.NewMockRepository(ctrl)

	logConfig := &config.LoggingConfig{Level: "debug", Format: "console"}
	log := logger.NewLogger(logConfig)

	uc := &serviceUseCase{
		repo: mockRepo,
		log:  log,
	}

	return uc, mockRepo
}

func TestServiceUseCase_CreateService(t *testing.T) {
	uc, mockRepo := setupTestServiceUseCase(t)

	tests := []struct {
		setup   func()
		name    string
		errMsg  string
		request request.CreateService
		wantErr bool
	}{
		{
			name: "successful service creation",
			setup: func() {
				mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&model.Service{
					ID:              uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
					Name:            "Test Service",
					Description:     "Test Description",
					DurationMinutes: 60,
					Price:           100.0,
					MaxCapacity:     10,
					IsActive:        true,
				}, nil)
			},
			request: request.CreateService{
				Name:            "Test Service",
				Description:     "Test Description",
				DurationMinutes: 60,
				Price:           100.0,
				MaxCapacity:     10,
				IsActive:        true,
			},
			wantErr: false,
		},
		{
			name:  "invalid price - zero",
			setup: func() {},
			request: request.CreateService{
				Name:            "Test Service",
				DurationMinutes: 60,
				Price:           0,
				MaxCapacity:     10,
			},
			wantErr: true,
			errMsg:  ErrInvalidServicePrice.Error(),
		},
		{
			name:  "invalid price - negative",
			setup: func() {},
			request: request.CreateService{
				Name:            "Test Service",
				DurationMinutes: 60,
				Price:           -50,
				MaxCapacity:     10,
			},
			wantErr: true,
			errMsg:  ErrInvalidServicePrice.Error(),
		},
		{
			name:  "invalid duration - too short",
			setup: func() {},
			request: request.CreateService{
				Name:            "Test Service",
				DurationMinutes: 0,
				Price:           100,
				MaxCapacity:     10,
			},
			wantErr: true,
			errMsg:  ErrInvalidDuration.Error(),
		},
		{
			name:  "invalid duration - too long",
			setup: func() {},
			request: request.CreateService{
				Name:            "Test Service",
				DurationMinutes: 500,
				Price:           100,
				MaxCapacity:     10,
			},
			wantErr: true,
			errMsg:  ErrInvalidDuration.Error(),
		},
		{
			name:  "invalid capacity - too low",
			setup: func() {},
			request: request.CreateService{
				Name:            "Test Service",
				DurationMinutes: 60,
				Price:           100,
				MaxCapacity:     0,
			},
			wantErr: true,
			errMsg:  ErrInvalidCapacity.Error(),
		},
		{
			name:  "invalid capacity - too high",
			setup: func() {},
			request: request.CreateService{
				Name:            "Test Service",
				DurationMinutes: 60,
				Price:           100,
				MaxCapacity:     100,
			},
			wantErr: true,
			errMsg:  ErrInvalidCapacity.Error(),
		},
		{
			name: "repository error",
			setup: func() {
				mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, assert.AnError)
			},
			request: request.CreateService{
				Name:            "Test Service",
				DurationMinutes: 60,
				Price:           100,
				MaxCapacity:     10,
			},
			wantErr: true,
			errMsg:  "assert.AnError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			result, err := uc.CreateService(context.Background(), tt.request)

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

func TestServiceUseCase_GetService(t *testing.T) {
	uc, mockRepo := setupTestServiceUseCase(t)
	testServiceID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		setup   func()
		name    string
		errMsg  string
		id      uuid.UUID
		wantErr bool
	}{
		{
			name: "successful get service",
			setup: func() {
				mockRepo.EXPECT().GetByID(gomock.Any(), testServiceID).Return(&model.Service{
					ID:              testServiceID,
					Name:            "Test Service",
					DurationMinutes: 60,
					Price:           100.0,
				}, nil)
			},
			id:      testServiceID,
			wantErr: false,
		},
		{
			name: "service not found",
			setup: func() {
				mockRepo.EXPECT().GetByID(gomock.Any(), testServiceID).Return(nil, repository.ErrServiceNotFound)
			},
			id:      testServiceID,
			wantErr: true,
			errMsg:  ErrServiceNotFound.Error(),
		},
		{
			name: "repository error",
			setup: func() {
				mockRepo.EXPECT().GetByID(gomock.Any(), testServiceID).Return(nil, assert.AnError)
			},
			id:      testServiceID,
			wantErr: true,
			errMsg:  "assert.AnError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			result, err := uc.GetService(context.Background(), tt.id)

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

func TestServiceUseCase_ListServices(t *testing.T) {
	uc, mockRepo := setupTestServiceUseCase(t)

	tests := []struct {
		setup   func()
		name    string
		request request.ListServices
		wantLen int
		wantErr bool
	}{
		{
			name: "successful list",
			setup: func() {
				id1, _ := uuid.NewV7()
				id2, _ := uuid.NewV7()
				mockRepo.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]*model.Service{
					{ID: id1, Name: "Service 1"},
					{ID: id2, Name: "Service 2"},
				}, nil)
			},
			request: request.ListServices{
				Limit:  20,
				Offset: 0,
			},
			wantErr: false,
			wantLen: 2,
		},
		{
			name: "empty list",
			setup: func() {
				mockRepo.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]*model.Service{}, nil)
			},
			request: request.ListServices{
				Limit:  20,
				Offset: 0,
			},
			wantErr: false,
			wantLen: 0,
		},
		{
			name: "repository error",
			setup: func() {
				mockRepo.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, assert.AnError)
			},
			request: request.ListServices{
				Limit:  20,
				Offset: 0,
			},
			wantErr: true,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			result, err := uc.ListServices(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result.Services, tt.wantLen)
			}
		})
	}
}
