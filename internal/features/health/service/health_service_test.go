package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zercle/zercle-go-template/internal/core/port/mocks"
	healthService "github.com/zercle/zercle-go-template/internal/features/health/service"
	"go.uber.org/mock/gomock"
)

func TestHealthService(t *testing.T) {
	t.Run("HealthCheck_Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockHealthRepository(ctrl)
		svc := healthService.NewHealthService(mockRepo)
		ctx := context.Background()

		// Expectations
		mockRepo.EXPECT().CheckDatabase(gomock.Any()).Times(0) // HealthCheck doesn't call repo

		// Act
		result, err := svc.HealthCheck(ctx)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "OK", result.Status)
		assert.WithinDuration(t, time.Now(), result.Timestamp, 1*time.Second)
	})

	t.Run("LivenessCheck_DatabaseHealthy", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockHealthRepository(ctrl)
		svc := healthService.NewHealthService(mockRepo)
		ctx := context.Background()

		expectedDBVersion := "PostgreSQL 16.0 on x86_64-pc-linux-gnu"
		mockRepo.EXPECT().CheckDatabase(gomock.Any()).Return(expectedDBVersion, nil)

		// Act
		result, err := svc.LivenessCheck(ctx)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "alive", result.Status)
		assert.Equal(t, expectedDBVersion, result.Database)
		assert.Empty(t, result.Error)
		assert.WithinDuration(t, time.Now(), result.Timestamp, 1*time.Second)
	})

	t.Run("LivenessCheck_DatabaseUnreachable", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockHealthRepository(ctrl)
		svc := healthService.NewHealthService(mockRepo)
		ctx := context.Background()

		expectedError := assert.AnError
		mockRepo.EXPECT().CheckDatabase(gomock.Any()).Return("unreachable", expectedError)

		// Act
		result, err := svc.LivenessCheck(ctx)

		// Assert
		assert.Error(t, err) // Service returns error when database is unreachable
		assert.NotNil(t, result)
		assert.Equal(t, "down", result.Status)
		assert.Equal(t, "unreachable", result.Database)
		assert.Equal(t, expectedError.Error(), result.Error)
	})

	t.Run("LivenessCheck_ContextTimeout", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockHealthRepository(ctrl)
		svc := healthService.NewHealthService(mockRepo)

		// Create context with very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// Expectations
		mockRepo.EXPECT().CheckDatabase(gomock.Any()).Return("", assert.AnError)

		// Act
		result, err := svc.LivenessCheck(ctx)

		// Assert
		assert.Error(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "down", result.Status)
	})

	t.Run("HealthCheck_ContextCancellation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockHealthRepository(ctrl)
		svc := healthService.NewHealthService(mockRepo)

		// Create context that's already cancelled
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Expectations
		mockRepo.EXPECT().CheckDatabase(gomock.Any()).Times(0)

		// Act
		result, err := svc.HealthCheck(ctx)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "OK", result.Status)
	})

	// Test table for LivenessCheck with different database responses
	t.Run("LivenessCheck_DatabaseVersionVariations", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		tests := []struct {
			name        string
			dbVersion   string
			dbError     error
			expectedDB  string
			expectedErr string
		}{
			{
				name:        "PostgreSQL 14",
				dbVersion:   "PostgreSQL 14.5",
				dbError:     nil,
				expectedDB:  "PostgreSQL 14.5",
				expectedErr: "",
			},
			{
				name:        "PostgreSQL 16",
				dbVersion:   "PostgreSQL 16.0",
				dbError:     nil,
				expectedDB:  "PostgreSQL 16.0",
				expectedErr: "",
			},
			{
				name:        "Connection Error",
				dbVersion:   "",
				dbError:     assert.AnError,
				expectedDB:  "unreachable",
				expectedErr: assert.AnError.Error(),
			},
			{
				name:        "Timeout Error",
				dbVersion:   "",
				dbError:     context.DeadlineExceeded,
				expectedDB:  "unreachable",
				expectedErr: context.DeadlineExceeded.Error(),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()
				// Arrange
				mockRepo := mocks.NewMockHealthRepository(ctrl)
				svc := healthService.NewHealthService(mockRepo)
				ctx := context.Background()

				mockRepo.EXPECT().CheckDatabase(gomock.Any()).Return(tt.dbVersion, tt.dbError)

				// Act
				result, err := svc.LivenessCheck(ctx)

				// Assert
				assert.NotNil(t, result)
				if tt.dbError == nil {
					assert.NoError(t, err)
					assert.Equal(t, "alive", result.Status)
					assert.Equal(t, tt.expectedDB, result.Database)
					assert.Empty(t, result.Error)
				} else {
					assert.Error(t, err)
					assert.Equal(t, "down", result.Status)
					assert.Equal(t, tt.expectedDB, result.Database)
					assert.Contains(t, result.Error, tt.expectedErr)
				}
			})
		}
	})
}
