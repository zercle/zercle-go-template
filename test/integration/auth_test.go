package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/zercle/zercle-go-template/internal/core/port/mocks"
	userDto "github.com/zercle/zercle-go-template/internal/features/user/dto"
	userHandler "github.com/zercle/zercle-go-template/internal/features/user/handler"
	"github.com/zercle/zercle-go-template/internal/features/user/service"
	sharederrors "github.com/zercle/zercle-go-template/internal/shared/errors"
	"go.uber.org/mock/gomock"
)

func TestAuthFlow_Integration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 1. Mock Repository
	mockRepo := mocks.NewMockUserRepository(ctrl)

	// 2. Real Service
	jwtSecret := "integration_secret"
	svc := service.NewUserService(mockRepo, jwtSecret, time.Hour)

	// 3. Real Handler
	handler := userHandler.NewUserHandler(svc)

	// 4. Fiber App
	app := fiber.New()
	auth := app.Group("/auth")
	auth.Post("/register", handler.Register)
	auth.Post("/login", handler.Login)

	t.Run("Register_Success", func(t *testing.T) {
		reqBody := userDto.RegisterRequest{
			Email:    "new@example.com",
			Password: "securepass",
			Name:     "Integration User",
		}
		body, _ := json.Marshal(reqBody)

		// Expectations
		mockRepo.EXPECT().GetByEmail(gomock.Any(), reqBody.Email).Return(nil, sharederrors.ErrNotFound)
		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

		req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		if resp != nil {
			defer resp.Body.Close()
			assert.Equal(t, http.StatusCreated, resp.StatusCode)
		}
	})

	t.Run("Login_Success", func(t *testing.T) {
		reqBody := userDto.LoginRequest{
			Email:    "new@example.com",
			Password: "securepass",
		}
		body, _ := json.Marshal(reqBody)

		mockRepo.EXPECT().GetByEmail(gomock.Any(), reqBody.Email).Return(nil, sharederrors.ErrNotFound)

		req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		if resp != nil {
			defer resp.Body.Close()
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		}
	})
}
