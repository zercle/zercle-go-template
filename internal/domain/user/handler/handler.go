package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/zercle/zercle-go-template/internal/domain/user"
	"github.com/zercle/zercle-go-template/internal/domain/user/request"
	"github.com/zercle/zercle-go-template/internal/domain/user/usecase"
	"github.com/zercle/zercle-go-template/internal/infrastructure/logger"
	"github.com/zercle/zercle-go-template/pkg/middleware"
	"github.com/zercle/zercle-go-template/pkg/response"
)

type userHandler struct {
	usecase user.UserService
	log     *logger.Logger
}

func NewUserHandler(usecase user.UserService, log *logger.Logger) user.UserHandler {
	return &userHandler{
		usecase: usecase,
		log:     log,
	}
}

func (h *userHandler) Register(c echo.Context) error {
	var req request.RegisterUser
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, "Validation failed", middleware.ValidationErrors(err))
	}

	result, err := h.usecase.Register(c.Request().Context(), req)
	if err != nil {
		h.log.Error("Failed to register user", "error", err, "request_id", middleware.GetRequestID(c))
		if errors.Is(err, usecase.ErrUserAlreadyExists) {
			return response.Error(c, http.StatusConflict, "Email already exists")
		}
		return response.InternalError(c, "Failed to register user")
	}

	return response.Created(c, result)
}

func (h *userHandler) Login(c echo.Context) error {
	var req request.LoginUser
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, "Validation failed", middleware.ValidationErrors(err))
	}

	result, err := h.usecase.Login(c.Request().Context(), req)
	if err != nil {
		h.log.Error("Failed to login user", "error", err, "request_id", middleware.GetRequestID(c), "email", req.Email)
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			return response.Unauthorized(c, "Invalid email or password")
		}
		return response.InternalError(c, "Failed to login user")
	}

	return response.OK(c, result)
}

func (h *userHandler) GetProfile(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "Invalid token")
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		return response.BadRequest(c, "Invalid user ID", nil)
	}

	result, err := h.usecase.GetProfile(c.Request().Context(), id)
	if err != nil {
		h.log.Error("Failed to get user profile", "error", err, "request_id", middleware.GetRequestID(c))
		if errors.Is(err, usecase.ErrUserNotFound) {
			return response.NotFound(c, "User not found")
		}
		return response.InternalError(c, "Failed to get user profile")
	}

	return response.OK(c, result)
}

func (h *userHandler) UpdateProfile(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "Invalid token")
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		return response.BadRequest(c, "Invalid user ID", nil)
	}

	var req request.UpdateUser
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	if err := c.Validate(&req); err != nil {
		return response.BadRequest(c, "Validation failed", middleware.ValidationErrors(err))
	}

	result, err := h.usecase.UpdateProfile(c.Request().Context(), id, req)
	if err != nil {
		h.log.Error("Failed to update user profile", "error", err, "request_id", middleware.GetRequestID(c))
		if errors.Is(err, usecase.ErrUserNotFound) {
			return response.NotFound(c, "User not found")
		}
		return response.InternalError(c, "Failed to update user profile")
	}

	return response.OK(c, result)
}

func (h *userHandler) DeleteAccount(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "Invalid token")
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		return response.BadRequest(c, "Invalid user ID", nil)
	}

	if err := h.usecase.DeleteAccount(c.Request().Context(), id); err != nil {
		h.log.Error("Failed to delete user account", "error", err, "request_id", middleware.GetRequestID(c))
		if errors.Is(err, usecase.ErrUserNotFound) {
			return response.NotFound(c, "User not found")
		}
		return response.InternalError(c, "Failed to delete user account")
	}

	return response.NoContent(c)
}

func (h *userHandler) ListUsers(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	result, err := h.usecase.ListUsers(c.Request().Context(), limit, offset)
	if err != nil {
		h.log.Error("Failed to list users", "error", err, "request_id", middleware.GetRequestID(c))
		return response.InternalError(c, "Failed to list users")
	}

	return response.OK(c, result)
}

func (h *userHandler) RegisterRoutes(api, protected *echo.Group) {
	api.POST("/auth/register", h.Register)
	api.POST("/auth/login", h.Login)

	protected.GET("/users/profile", h.GetProfile)
	protected.PUT("/users/profile", h.UpdateProfile)
	protected.DELETE("/users/profile", h.DeleteAccount)
	protected.GET("/users", h.ListUsers)
}
