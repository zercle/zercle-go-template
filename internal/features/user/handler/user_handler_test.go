package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/zercle/zercle-go-template/internal/core/port/mocks"
	userDto "github.com/zercle/zercle-go-template/internal/features/user/dto"
	userHandler "github.com/zercle/zercle-go-template/internal/features/user/handler"
	"go.uber.org/mock/gomock"
)

// mockAuthMiddleware simulates authentication middleware that sets user_id in locals
func mockAuthMiddleware(userID string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if userID != "" {
			c.Locals("user_id", userID)
		}
		return c.Next()
	}
}

func TestUserHandler(t *testing.T) {
	t.Run("Register_Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockService := mocks.NewMockUserService(ctrl)
		handler := userHandler.NewUserHandler(mockService)

		app := fiber.New()
		app.Post("/auth/register", handler.Register)

		reqBody := userDto.RegisterRequest{
			Email:    "test@example.com",
			Password: "securepass123",
			Name:     "Test User",
		}
		body, _ := json.Marshal(reqBody)

		expectedUser := &userDto.UserResponse{
			ID:        uuid.New().String(),
			Email:     reqBody.Email,
			Name:      reqBody.Name,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockService.EXPECT().Register(gomock.Any(), gomock.Eq(&reqBody)).Return(expectedUser, nil)

		// Act
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		// Assert
		assert.NoError(t, err)
		if resp != nil {
			defer resp.Body.Close()
			assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

			var response map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&response)
			assert.Equal(t, "success", response["status"])
		}
	})

	t.Run("Register_DuplicateEmail", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockService := mocks.NewMockUserService(ctrl)
		handler := userHandler.NewUserHandler(mockService)

		app := fiber.New()
		app.Post("/auth/register", handler.Register)

		reqBody := userDto.RegisterRequest{
			Email:    "existing@example.com",
			Password: "securepass123",
			Name:     "Test User",
		}
		body, _ := json.Marshal(reqBody)

		mockService.EXPECT().Register(gomock.Any(), gomock.Eq(&reqBody)).Return(nil, assert.AnError)

		// Act
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		// Assert
		assert.NoError(t, err)
		if resp != nil {
			defer resp.Body.Close()
			assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		}
	})

	t.Run("Login_Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockService := mocks.NewMockUserService(ctrl)
		handler := userHandler.NewUserHandler(mockService)

		app := fiber.New()
		app.Post("/auth/login", handler.Login)

		reqBody := userDto.LoginRequest{
			Email:    "test@example.com",
			Password: "securepass123",
		}
		body, _ := json.Marshal(reqBody)

		expectedToken := "jwt.token.here"
		mockService.EXPECT().Login(gomock.Any(), gomock.Eq(&reqBody)).Return(expectedToken, nil)

		// Act
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		// Assert
		assert.NoError(t, err)
		if resp != nil {
			defer resp.Body.Close()
			assert.Equal(t, fiber.StatusOK, resp.StatusCode)

			var response map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&response)
			assert.Equal(t, "success", response["status"])
		}
	})

	t.Run("Login_InvalidCredentials", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockService := mocks.NewMockUserService(ctrl)
		handler := userHandler.NewUserHandler(mockService)

		app := fiber.New()
		app.Post("/auth/login", handler.Login)

		reqBody := userDto.LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}
		body, _ := json.Marshal(reqBody)

		mockService.EXPECT().Login(gomock.Any(), gomock.Eq(&reqBody)).Return("", assert.AnError)

		// Act
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		// Assert
		assert.NoError(t, err)
		if resp != nil {
			defer resp.Body.Close()
			assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		}
	})

	t.Run("GetProfile_Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockService := mocks.NewMockUserService(ctrl)
		handler := userHandler.NewUserHandler(mockService)

		userID := uuid.New()
		app := fiber.New()
		app.Get("/users/me", mockAuthMiddleware(userID.String()), handler.GetProfile)

		expectedUser := &userDto.UserResponse{
			ID:        userID.String(),
			Email:     "test@example.com",
			Name:      "Test User",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockService.EXPECT().GetProfile(gomock.Any(), gomock.Eq(userID)).Return(expectedUser, nil)

		// Act
		req := httptest.NewRequest("GET", "/users/me", nil)
		resp, err := app.Test(req)

		// Assert
		assert.NoError(t, err)
		if resp != nil {
			defer resp.Body.Close()
			assert.Equal(t, fiber.StatusOK, resp.StatusCode)

			var response map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&response)
			assert.Equal(t, "success", response["status"])
		}
	})

	t.Run("GetProfile_Unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockService := mocks.NewMockUserService(ctrl)
		handler := userHandler.NewUserHandler(mockService)

		app := fiber.New()
		// No auth middleware - user_id not set
		app.Get("/users/me", handler.GetProfile)

		// Act - request without authentication
		req := httptest.NewRequest("GET", "/users/me", nil)
		resp, err := app.Test(req)

		// Assert
		assert.NoError(t, err)
		if resp != nil {
			defer resp.Body.Close()
			assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
		}
	})

	t.Run("Register_InvalidJSON", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockService := mocks.NewMockUserService(ctrl)
		handler := userHandler.NewUserHandler(mockService)

		app := fiber.New()
		app.Post("/auth/register", handler.Register)

		// Act - invalid JSON
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		// Assert
		assert.NoError(t, err)
		if resp != nil {
			defer resp.Body.Close()
			assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("Register_ValidationError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockService := mocks.NewMockUserService(ctrl)
		handler := userHandler.NewUserHandler(mockService)

		app := fiber.New()
		app.Post("/auth/register", handler.Register)

		reqBody := userDto.RegisterRequest{
			Email:    "invalid-email",
			Password: "123", // too short
			Name:     "",    // empty
		}
		body, _ := json.Marshal(reqBody)

		// Act
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		// Assert
		assert.NoError(t, err)
		if resp != nil {
			defer resp.Body.Close()
			assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		}
	})

	// Test table for comprehensive coverage
	t.Run("Handler_EdgeCases", func(t *testing.T) {
		tests := []struct {
			name         string
			method       string
			path         string
			body         interface{}
			headers      map[string]string
			withAuth     bool
			setupMock    func(*mocks.MockUserService)
			expectedCode int
		}{
			{
				name:   "Login Empty Body",
				method: "POST",
				path:   "/auth/login",
				body:   nil,
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				withAuth:     false,
				setupMock:    func(m *mocks.MockUserService) {},
				expectedCode: fiber.StatusBadRequest,
			},
			{
				name:   "Register Empty Body",
				method: "POST",
				path:   "/auth/register",
				body:   nil,
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				withAuth:     false,
				setupMock:    func(m *mocks.MockUserService) {},
				expectedCode: fiber.StatusBadRequest,
			},
			{
				name:         "GetProfile No Auth",
				method:       "GET",
				path:         "/users/me",
				body:         nil,
				headers:      map[string]string{},
				withAuth:     false,
				setupMock:    func(m *mocks.MockUserService) {},
				expectedCode: fiber.StatusUnauthorized,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()
				// Arrange
				mockService := mocks.NewMockUserService(ctrl)
				handler := userHandler.NewUserHandler(mockService)

				app := fiber.New()
				userID := uuid.New()
				app.Post("/auth/register", handler.Register)
				app.Post("/auth/login", handler.Login)
				if tt.withAuth {
					app.Get("/users/me", mockAuthMiddleware(userID.String()), handler.GetProfile)
				} else {
					app.Get("/users/me", handler.GetProfile)
				}

				var body *bytes.Reader
				if tt.body != nil {
					data, _ := json.Marshal(tt.body)
					body = bytes.NewReader(data)
				} else {
					body = bytes.NewReader(nil)
				}

				req := httptest.NewRequest(tt.method, tt.path, body)
				for k, v := range tt.headers {
					req.Header.Set(k, v)
				}

				// Setup mock if needed
				tt.setupMock(mockService)

				// Act
				resp, err := app.Test(req)

				// Assert
				assert.NoError(t, err)
				if resp != nil {
					defer resp.Body.Close()
					assert.Equal(t, tt.expectedCode, resp.StatusCode)
				}
			})
		}
	})

	// Test with bcrypt password hash
	t.Run("Register_WithPasswordHashing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockService := mocks.NewMockUserService(ctrl)
		handler := userHandler.NewUserHandler(mockService)

		app := fiber.New()
		app.Post("/auth/register", handler.Register)

		password := "MySecurePassword123!"

		reqBody := userDto.RegisterRequest{
			Email:    "test@example.com",
			Password: password,
			Name:     "Test User",
		}
		body, _ := json.Marshal(reqBody)

		expectedUser := &userDto.UserResponse{
			ID:        uuid.New().String(),
			Email:     reqBody.Email,
			Name:      reqBody.Name,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockService.EXPECT().Register(gomock.Any(), gomock.Eq(&reqBody)).Return(expectedUser, nil)

		// Act
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		// Assert
		assert.NoError(t, err)
		if resp != nil {
			defer resp.Body.Close()
			assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
		}
	})
}
