package handler_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/zercle/zercle-go-template/internal/core/port/mocks"
	healthDto "github.com/zercle/zercle-go-template/internal/features/health/dto"
	healthHandler "github.com/zercle/zercle-go-template/internal/features/health/handler"
	"go.uber.org/mock/gomock"
)

func TestHealthHandler(t *testing.T) {
	t.Run("HealthCheck_Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockService := mocks.NewMockHealthService(ctrl)
		handler := healthHandler.NewHealthHandler(mockService)

		app := fiber.New()
		app.Get("/health", handler.HealthCheck)

		// Expectations
		mockService.EXPECT().HealthCheck(gomock.Any()).Return(&healthDto.HealthResponse{
			Status:    "OK",
			Timestamp: time.Now(),
		}, nil)

		// Act
		req := httptest.NewRequest("GET", "/health", nil)
		resp, err := app.Test(req)

		// Assert
		assert.NoError(t, err)
		if resp != nil {
			defer resp.Body.Close()
			assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		}
	})

	t.Run("LivenessCheck_Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockService := mocks.NewMockHealthService(ctrl)
		handler := healthHandler.NewHealthHandler(mockService)

		app := fiber.New()
		app.Get("/health/live", handler.Liveness)

		// Expectations
		mockService.EXPECT().LivenessCheck(gomock.Any()).Return(&healthDto.LivenessResponse{
			Status:    "alive",
			Timestamp: time.Now(),
			Database:  "PostgreSQL 16.0",
		}, nil)

		// Act
		req := httptest.NewRequest("GET", "/health/live", nil)
		resp, err := app.Test(req)

		// Assert
		assert.NoError(t, err)
		if resp != nil {
			defer resp.Body.Close()
			assert.Equal(t, fiber.StatusOK, resp.StatusCode)

			// Verify response structure
			var response map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&response)
			assert.Equal(t, "success", response["status"])
		}
	})

	t.Run("LivenessCheck_DatabaseDown", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockService := mocks.NewMockHealthService(ctrl)
		handler := healthHandler.NewHealthHandler(mockService)

		app := fiber.New()
		app.Get("/health/live", handler.Liveness)

		// Expectations
		mockService.EXPECT().LivenessCheck(gomock.Any()).Return(&healthDto.LivenessResponse{
			Status:    "down",
			Timestamp: time.Now(),
			Database:  "unreachable",
			Error:     "connection timeout",
		}, nil)

		// Act
		req := httptest.NewRequest("GET", "/health/live", nil)
		resp, err := app.Test(req)

		// Assert
		assert.NoError(t, err)
		if resp != nil {
			defer resp.Body.Close()
			assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

			// Verify response structure
			var response map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&response)
			assert.Equal(t, "fail", response["status"])
		}
	})

	t.Run("LivenessCheck_ServiceError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockService := mocks.NewMockHealthService(ctrl)
		handler := healthHandler.NewHealthHandler(mockService)

		app := fiber.New()
		app.Get("/health/live", handler.Liveness)

		// Expectations
		mockService.EXPECT().LivenessCheck(gomock.Any()).Return(nil, assert.AnError)

		// Act
		req := httptest.NewRequest("GET", "/health/live", nil)
		resp, err := app.Test(req)

		// Assert
		assert.NoError(t, err)
		if resp != nil {
			defer resp.Body.Close()
			assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
		}
	})
}
