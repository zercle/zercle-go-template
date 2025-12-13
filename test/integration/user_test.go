package integration

import (
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	userDto "github.com/zercle/zercle-go-template/internal/features/user/dto"
	userHandler "github.com/zercle/zercle-go-template/internal/features/user/handler"
	"github.com/zercle/zercle-go-template/internal/features/user/service"
	"go.uber.org/mock/gomock"
)

func TestUserEndpoints_Integration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock repository
	mockRepo := SetupMockUserRepo(ctrl)

	// Create service with mock repository
	jwtSecret := "test-secret-key-for-integration"
	jwtExpiry := time.Hour
	svc := service.NewUserService(mockRepo, jwtSecret, jwtExpiry)

	// Create handler with service
	handler := userHandler.NewUserHandler(svc)

	// Create Fiber app
	app := fiber.New()
	auth := app.Group("/api/v1/auth")
	auth.Post("/register", handler.Register)
	auth.Post("/login", handler.Login)

	users := app.Group("/api/v1/users")
	users.Get("/me", handler.GetProfile)

	t.Run("POST /api/v1/auth/register - Success", func(t *testing.T) {
		// Setup mock expectations
		mockRepo.EXPECT().GetByEmail(gomock.Any(), "newuser@example.com").Return(nil, ErrNotFound)
		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

		req := NewRequest("POST", "/api/v1/auth/register", ValidRegisterRequest, false, uuid.Nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		jsend := ValidateSuccessResponse(t, resp, fiber.StatusCreated)
		assert.NotNil(t, jsend.Data)
	})

	t.Run("POST /api/v1/auth/register - Duplicate Email", func(t *testing.T) {
		// Setup mock expectations
		mockRepo.EXPECT().GetByEmail(gomock.Any(), "existing@example.com").Return(MockUser1, nil)

		req := NewRequest("POST", "/api/v1/auth/register", userDto.RegisterRequest{
			Email:    "existing@example.com",
			Password: "password123",
			Name:     "Existing User",
		}, false, uuid.Nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		jsend := ValidateErrorResponse(t, resp, fiber.StatusConflict)
		assert.NotNil(t, jsend.Message)
	})

	t.Run("POST /api/v1/auth/register - Invalid Email", func(t *testing.T) {
		req := NewRequest("POST", "/api/v1/auth/register", InvalidRegisterRequest_InvalidEmail, false, uuid.Nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		jsend := ValidateFailResponse(t, resp, fiber.StatusBadRequest)
		assert.NotNil(t, jsend.Data)
	})

	t.Run("POST /api/v1/auth/register - Empty Email", func(t *testing.T) {
		req := NewRequest("POST", "/api/v1/auth/register", InvalidRegisterRequest_EmptyEmail, false, uuid.Nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		jsend := ValidateFailResponse(t, resp, fiber.StatusBadRequest)
		assert.NotNil(t, jsend.Data)
	})

	t.Run("POST /api/v1/auth/register - Short Password", func(t *testing.T) {
		req := NewRequest("POST", "/api/v1/auth/register", InvalidRegisterRequest_ShortPassword, false, uuid.Nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		jsend := ValidateFailResponse(t, resp, fiber.StatusBadRequest)
		assert.NotNil(t, jsend.Data)
	})

	t.Run("POST /api/v1/auth/register - Short Name", func(t *testing.T) {
		req := NewRequest("POST", "/api/v1/auth/register", InvalidRegisterRequest_ShortName, false, uuid.Nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		jsend := ValidateFailResponse(t, resp, fiber.StatusBadRequest)
		assert.NotNil(t, jsend.Data)
	})

	t.Run("POST /api/v1/auth/login - Success", func(t *testing.T) {
		// Setup mock expectations
		mockRepo.EXPECT().GetByEmail(gomock.Any(), "test1@example.com").Return(MockUser1, nil)

		req := NewRequest("POST", "/api/v1/auth/login", ValidLoginRequest, false, uuid.Nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		jsend := ValidateSuccessResponse(t, resp, fiber.StatusOK)
		assert.NotNil(t, jsend.Data)
		// Verify token is present
		dataMap, ok := jsend.Data.(map[string]interface{})
		assert.True(t, ok, "Expected data to be a map")
		assert.NotNil(t, dataMap["token"])
	})

	t.Run("POST /api/v1/auth/login - Invalid Email", func(t *testing.T) {
		// Setup mock expectations
		mockRepo.EXPECT().GetByEmail(gomock.Any(), "nonexistent@example.com").Return(nil, ErrNotFound)

		req := NewRequest("POST", "/api/v1/auth/login", userDto.LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}, false, uuid.Nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		jsend := ValidateErrorResponse(t, resp, fiber.StatusUnauthorized)
		assert.NotNil(t, jsend.Message)
	})

	t.Run("POST /api/v1/auth/login - Invalid Password", func(t *testing.T) {
		// Setup mock expectations
		mockRepo.EXPECT().GetByEmail(gomock.Any(), "test1@example.com").Return(MockUser1, nil)

		req := NewRequest("POST", "/api/v1/auth/login", userDto.LoginRequest{
			Email:    "test1@example.com",
			Password: "wrongpassword",
		}, false, uuid.Nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		jsend := ValidateErrorResponse(t, resp, fiber.StatusUnauthorized)
		assert.NotNil(t, jsend.Message)
	})

	t.Run("GET /api/v1/users/me - Success", func(t *testing.T) {
		// Setup mock expectations
		mockRepo.EXPECT().GetByID(gomock.Any(), MockUserID1).Return(MockUser1, nil)

		// Create request with auth
		req := NewRequest("GET", "/api/v1/users/me", nil, true, MockUserID1)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		jsend := ValidateSuccessResponse(t, resp, fiber.StatusOK)
		assert.NotNil(t, jsend.Data)
	})

	t.Run("GET /api/v1/users/me - User Not Found", func(t *testing.T) {
		// Setup mock expectations
		mockRepo.EXPECT().GetByID(gomock.Any(), MockUserID3).Return(nil, ErrNotFound)

		// Create request with auth
		req := NewRequest("GET", "/api/v1/users/me", nil, true, MockUserID3)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		jsend := ValidateErrorResponse(t, resp, fiber.StatusNotFound)
		assert.NotNil(t, jsend.Message)
	})
}
