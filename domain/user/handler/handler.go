package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/zercle/zercle-go-template/domain/user"
	"github.com/zercle/zercle-go-template/domain/user/repository"
	"github.com/zercle/zercle-go-template/domain/user/request"
	"github.com/zercle/zercle-go-template/domain/user/usecase"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
	"github.com/zercle/zercle-go-template/pkg/middleware"
	"github.com/zercle/zercle-go-template/pkg/response"
)

type userHandler struct {
	usecase user.Usecase
	log     *logger.Logger
}

// NewUserHandler creates a new user HTTP handler
func NewUserHandler(usecase user.Usecase, log *logger.Logger) user.Handler {
	return &userHandler{
		usecase: usecase,
		log:     log,
	}
}

// Register handles user registration
// @Summary      Register a new user
// @Description  Create a new user account with email, password, and name
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body request.RegisterUser true "Registration details"
// @Success      201  {object}  map[string]interface{} "User registered successfully"
// @Failure      400  {object} map[string]interface{} "Validation error"
// @Failure      409  {object} map[string]interface{} "Email already exists"
// @Failure      500  {object} map[string]interface{} "Internal server error"
// @Router       /auth/register [post]
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

// Login handles user login
// @Summary      User login
// @Description  Authenticate user with email and password, returns JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body request.LoginUser true "Login credentials"
// @Success      200  {object}  map[string]interface{} "Login successful"
// @Failure      400  {object} map[string]interface{} "Validation error"
// @Failure      401  {object} map[string]interface{} "Invalid credentials"
// @Failure      500  {object} map[string]interface{} "Internal server error"
// @Router       /auth/login [post]
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

// GetProfile handles get user profile
// @Summary      Get user profile
// @Description  Get the authenticated user's profile information
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {object}  map[string]interface{} "Profile retrieved"
// @Failure      401  {object} map[string]interface{} "Unauthorized"
// @Failure      404  {object} map[string]interface{} "User not found"
// @Failure      500  {object} map[string]interface{} "Internal server error"
// @Router       /users/profile [get]
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
		if errors.Is(err, repository.ErrUserNotFound) {
			return response.NotFound(c, "User not found")
		}
		return response.InternalError(c, "Failed to get user profile")
	}

	return response.OK(c, result)
}

// UpdateProfile handles update user profile
// @Summary      Update user profile
// @Description  Update the authenticated user's profile information
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body request.UpdateUser true "Profile update details"
// @Success      200  {object}  map[string]interface{} "Profile updated"
// @Failure      400  {object} map[string]interface{} "Validation error"
// @Failure      401  {object} map[string]interface{} "Unauthorized"
// @Failure      404  {object} map[string]interface{} "User not found"
// @Failure      409  {object} map[string]interface{} "Email already exists"
// @Failure      500  {object} map[string]interface{} "Internal server error"
// @Router       /users/profile [put]
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
		if errors.Is(err, repository.ErrUserNotFound) {
			return response.NotFound(c, "User not found")
		}
		return response.InternalError(c, "Failed to update user profile")
	}

	return response.OK(c, result)
}

// DeleteAccount handles delete user account
// @Summary      Delete user account
// @Description  Delete the authenticated user's account
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      204  "Account deleted successfully"
// @Failure      401  {object} map[string]interface{} "Unauthorized"
// @Failure      404  {object} map[string]interface{} "User not found"
// @Failure      500  {object} map[string]interface{} "Internal server error"
// @Router       /users/profile [delete]
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
		if errors.Is(err, repository.ErrUserNotFound) {
			return response.NotFound(c, "User not found")
		}
		return response.InternalError(c, "Failed to delete user account")
	}

	return response.NoContent(c)
}

// ListUsers handles list users
// @Summary      List users
// @Description  Get a paginated list of users
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        limit   query     int     false  "Number of items to return (max 100)"  default(20)
// @Param        offset  query     int     false  "Number of items to skip"              default(0)
// @Success      200     {object}  map[string]interface{} "Users retrieved"
// @Failure      401     {object} map[string]interface{} "Unauthorized"
// @Failure      500     {object} map[string]interface{} "Internal server error"
// @Router       /users [get]
func (h *userHandler) ListUsers(c echo.Context) error {
	// Parse query parameters
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
