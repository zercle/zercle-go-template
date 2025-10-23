package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	h "github.com/zercle/zercle-go-template/internal/adapter/handler/http"
	"github.com/zercle/zercle-go-template/pkg/dto"
	"github.com/zercle/zercle-go-template/test/mocks"
	"go.uber.org/mock/gomock"
)

func TestUserHandler_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockUserService(ctrl)
	handler := h.NewUserHandler(mockSvc)

	app := fiber.New()
	app.Post("/register", handler.Register)

	reqBody := dto.RegisterRequest{
		Email:    "test@example.com",
		Password: "password",
		Name:     "Test",
	}
	body, _ := json.Marshal(reqBody)

	resDTO := &dto.UserResponse{
		ID:    "uuid",
		Email: reqBody.Email,
		Name:  reqBody.Name,
	}

	mockSvc.EXPECT().Register(gomock.Any(), gomock.Any()).Return(resDTO, nil)

	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	if resp != nil {
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}
}

func TestUserHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockUserService(ctrl)
	handler := h.NewUserHandler(mockSvc)

	app := fiber.New()
	app.Post("/login", handler.Login)

	reqBody := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password",
	}
	body, _ := json.Marshal(reqBody)

	mockSvc.EXPECT().Login(gomock.Any(), gomock.Any()).Return("jwt-token", nil)

	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	if resp != nil {
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
}

func TestUserHandler_GetProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockUserService(ctrl)
	handler := h.NewUserHandler(mockSvc)

	app := fiber.New()
	userID := uuid.New()

	// Middleware mock
	app.Get("/me", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID.String())
		return c.Next()
	}, handler.GetProfile)

	resDTO := &dto.UserResponse{
		ID:    userID.String(),
		Email: "test@example.com",
		Name:  "Test",
	}

	mockSvc.EXPECT().GetProfile(gomock.Any(), userID).Return(resDTO, nil)

	req := httptest.NewRequest("GET", "/me", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	if resp != nil {
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
}

func TestUserHandler_GetProfile_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockUserService(ctrl)
	handler := h.NewUserHandler(mockSvc)

	app := fiber.New()
	// No middleware setting user_id
	app.Get("/me", handler.GetProfile)

	req := httptest.NewRequest("GET", "/me", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	if resp != nil {
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestUserHandler_GetProfile_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockUserService(ctrl)
	handler := h.NewUserHandler(mockSvc)

	app := fiber.New()
	userID := uuid.New()

	app.Get("/me", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID.String())
		return c.Next()
	}, handler.GetProfile)

	mockSvc.EXPECT().GetProfile(gomock.Any(), userID).Return(nil, errors.New("service error"))

	req := httptest.NewRequest("GET", "/me", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	if resp != nil {
		defer resp.Body.Close()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	}
}
