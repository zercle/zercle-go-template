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

	usermock "github.com/zercle/zercle-go-template/internal/domain/user/mock"
	userResponse "github.com/zercle/zercle-go-template/internal/domain/user/response"
	"github.com/zercle/zercle-go-template/internal/domain/user/usecase"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/logger"
	"github.com/zercle/zercle-go-template/pkg/middleware"
	"github.com/zercle/zercle-go-template/pkg/response"
)

func setupTestUserHandler(t *testing.T) (*userHandler, *usermock.MockUserService, *echo.Echo) {
	ctrl := gomock.NewController(t)
	mockUsecase := usermock.NewMockUserService(ctrl)

	logConfig := &config.LoggingConfig{Level: "debug", Format: "console"}
	log := logger.NewLogger(logConfig)
	h := &userHandler{
		usecase: mockUsecase,
		log:     log,
	}

	e := echo.New()
	e.Validator = middleware.NewCustomValidator()

	return h, mockUsecase, e
}

func TestUserHandler_Register(t *testing.T) {
	h, mockUsecase, e := setupTestUserHandler(t)
	testUserID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		setupMock       func()
		name            string
		requestBody     string
		wantStatusField response.Status
		wantStatus      int
	}{
		{
			name: "successful registration",
			setupMock: func() {
				mockUsecase.EXPECT().Register(gomock.Any(), gomock.Any()).Return(&userResponse.LoginResponse{
					Token: "test-token",
					User: userResponse.UserResponse{
						ID:       testUserID,
						Email:    "test@example.com",
						FullName: "Test User",
					},
				}, nil)
			},
			requestBody:     `{"email":"test@example.com","password":"password123","full_name":"Test User","phone":"1234567890"}`,
			wantStatus:      http.StatusCreated,
			wantStatusField: response.StatusSuccess,
		},
		{
			name:            "validation error - missing email",
			setupMock:       func() {},
			requestBody:     `{"password":"password123","full_name":"Test User"}`,
			wantStatus:      http.StatusBadRequest,
			wantStatusField: response.StatusFail,
		},
		{
			name:            "validation error - missing password",
			setupMock:       func() {},
			requestBody:     `{"email":"test@example.com","full_name":"Test User"}`,
			wantStatus:      http.StatusBadRequest,
			wantStatusField: response.StatusFail,
		},
		{
			name:            "validation error - invalid email format",
			setupMock:       func() {},
			requestBody:     `{"email":"invalid-email","password":"password123","full_name":"Test User"}`,
			wantStatus:      http.StatusBadRequest,
			wantStatusField: response.StatusFail,
		},
		{
			name: "user already exists",
			setupMock: func() {
				mockUsecase.EXPECT().Register(gomock.Any(), gomock.Any()).Return(nil, usecase.ErrUserAlreadyExists)
			},
			requestBody:     `{"email":"existing@example.com","password":"password123","full_name":"Test User"}`,
			wantStatus:      http.StatusConflict,
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

			req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			_ = h.Register(c)

			assert.Equal(t, tt.wantStatus, rec.Code)

			var resp response.JSend
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatusField, resp.Status)
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	h, mockUsecase, e := setupTestUserHandler(t)

	tests := []struct {
		setupMock       func()
		name            string
		requestBody     string
		wantStatusField response.Status
		wantStatus      int
	}{
		{
			name: "successful login",
			setupMock: func() {
				mockUsecase.EXPECT().Login(gomock.Any(), gomock.Any()).Return(&userResponse.LoginResponse{
					Token: "test-token",
					User: userResponse.UserResponse{
						ID:       uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
						Email:    "test@example.com",
						FullName: "Test User",
					},
				}, nil)
			},
			requestBody:     `{"email":"test@example.com","password":"password123"}`,
			wantStatus:      http.StatusOK,
			wantStatusField: response.StatusSuccess,
		},
		{
			name:            "validation error - missing email",
			setupMock:       func() {},
			requestBody:     `{"password":"password123"}`,
			wantStatus:      http.StatusBadRequest,
			wantStatusField: response.StatusFail,
		},
		{
			name:            "validation error - missing password",
			setupMock:       func() {},
			requestBody:     `{"email":"test@example.com"}`,
			wantStatus:      http.StatusBadRequest,
			wantStatusField: response.StatusFail,
		},
		{
			name: "invalid credentials",
			setupMock: func() {
				mockUsecase.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil, usecase.ErrInvalidCredentials)
			},
			requestBody:     `{"email":"test@example.com","password":"wrongpassword"}`,
			wantStatus:      http.StatusUnauthorized,
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

			req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			_ = h.Login(c)

			assert.Equal(t, tt.wantStatus, rec.Code)

			var resp response.JSend
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatusField, resp.Status)
		})
	}
}

func TestUserHandler_GetProfile(t *testing.T) {
	h, mockUsecase, e := setupTestUserHandler(t)
	testUserID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		setupMock       func()
		name            string
		userID          string
		wantStatusField response.Status
		wantStatus      int
	}{
		{
			name: "successful get profile",
			setupMock: func() {
				mockUsecase.EXPECT().GetProfile(gomock.Any(), testUserID).Return(&userResponse.UserResponse{
					ID:       testUserID,
					Email:    "test@example.com",
					FullName: "Test User",
					Phone:    "1234567890",
				}, nil)
			},
			userID:          testUserID.String(),
			wantStatus:      http.StatusOK,
			wantStatusField: response.StatusSuccess,
		},
		{
			name:            "unauthorized - missing user ID",
			setupMock:       func() {},
			userID:          "",
			wantStatus:      http.StatusUnauthorized,
			wantStatusField: response.StatusError,
		},
		{
			name: "user not found",
			setupMock: func() {
				mockUsecase.EXPECT().GetProfile(gomock.Any(), testUserID).Return(nil, usecase.ErrUserNotFound)
			},
			userID:          testUserID.String(),
			wantStatus:      http.StatusNotFound,
			wantStatusField: response.StatusError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest(http.MethodGet, "/users/profile", http.NoBody)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.userID != "" {
				c.Set("user_id", tt.userID)
			}

			_ = h.GetProfile(c)

			assert.Equal(t, tt.wantStatus, rec.Code)

			var resp response.JSend
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatusField, resp.Status)
		})
	}
}

func TestUserHandler_UpdateProfile(t *testing.T) {
	h, mockUsecase, e := setupTestUserHandler(t)
	testUserID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		setupMock       func()
		name            string
		userID          string
		requestBody     string
		wantStatusField response.Status
		wantStatus      int
	}{
		{
			name: "successful update profile",
			setupMock: func() {
				mockUsecase.EXPECT().UpdateProfile(gomock.Any(), testUserID, gomock.Any()).Return(&userResponse.UserResponse{
					ID:       testUserID,
					Email:    "test@example.com",
					FullName: "Updated Name",
				}, nil)
			},
			userID:          testUserID.String(),
			requestBody:     `{"full_name":"Updated Name"}`,
			wantStatus:      http.StatusOK,
			wantStatusField: response.StatusSuccess,
		},
		{
			name:            "unauthorized - missing user ID",
			setupMock:       func() {},
			userID:          "",
			requestBody:     `{"full_name":"Updated Name"}`,
			wantStatus:      http.StatusUnauthorized,
			wantStatusField: response.StatusError,
		},
		{
			name: "user not found",
			setupMock: func() {
				mockUsecase.EXPECT().UpdateProfile(gomock.Any(), testUserID, gomock.Any()).Return(nil, usecase.ErrUserNotFound)
			},
			userID:          testUserID.String(),
			requestBody:     `{"full_name":"Updated Name"}`,
			wantStatus:      http.StatusNotFound,
			wantStatusField: response.StatusError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest(http.MethodPut, "/users/profile", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.userID != "" {
				c.Set("user_id", tt.userID)
			}

			_ = h.UpdateProfile(c)

			assert.Equal(t, tt.wantStatus, rec.Code)

			var resp response.JSend
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatusField, resp.Status)
		})
	}
}
