package integration

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	healthHandler "github.com/zercle/zercle-go-template/internal/features/health/handler"
	"github.com/zercle/zercle-go-template/internal/features/health/service"
	"go.uber.org/mock/gomock"
)

func TestHealthEndpoints_Integration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock repository
	mockRepo := SetupMockHealthRepo(ctrl)

	// Create service with mock repository
	svc := service.NewHealthService(mockRepo)

	// Create handler with service
	handler := healthHandler.NewHealthHandler(svc)

	// Create Fiber app
	app := fiber.New()
	app.Get("/health", handler.HealthCheck)
	app.Get("/health/live", handler.Liveness)

	t.Run("GET /health - Success", func(t *testing.T) {
		// This endpoint doesn't use the repository
		req := httptest.NewRequest("GET", "/health", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		if resp != nil {
			defer resp.Body.Close()
		}
		jsend := ValidateSuccessResponse(t, resp, fiber.StatusOK)
		assert.NotNil(t, jsend.Data)
	})

	t.Run("GET /health/live - Database Connected", func(t *testing.T) {
		// Setup mock expectation
		mockRepo.EXPECT().CheckDatabase(gomock.Any()).Return("connected", nil)

		req := httptest.NewRequest("GET", "/health/live", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		if resp != nil {
			defer resp.Body.Close()
		}
		jsend := ValidateSuccessResponse(t, resp, fiber.StatusOK)
		assert.NotNil(t, jsend.Data)
	})

	t.Run("GET /health/live - Database Connection Error", func(t *testing.T) {
		// Setup mock expectation
		mockRepo.EXPECT().CheckDatabase(gomock.Any()).Return("unreachable", assert.AnError)

		req := httptest.NewRequest("GET", "/health/live", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		if resp != nil {
			defer resp.Body.Close()
		}
		jsend := ValidateFailResponse(t, resp, fiber.StatusServiceUnavailable)
		assert.NotNil(t, jsend.Data)
	})

	t.Run("GET /health/live - Database Unreachable", func(t *testing.T) {
		// Setup mock expectation
		mockRepo.EXPECT().CheckDatabase(gomock.Any()).Return("unreachable", nil)

		req := httptest.NewRequest("GET", "/health/live", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		if resp != nil {
			defer resp.Body.Close()
		}
		jsend := ValidateFailResponse(t, resp, fiber.StatusServiceUnavailable)
		assert.NotNil(t, jsend.Data)
	})
}
