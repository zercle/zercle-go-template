// Package handler provides HTTP handlers for the user feature.
// It handles request parsing, validation, and response formatting.
package handler

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"

	appErr "zercle-go-template/internal/errors"
	authusecase "zercle-go-template/internal/feature/auth/usecase"
	"zercle-go-template/internal/feature/user/dto"
	"zercle-go-template/internal/feature/user/usecase"
	"zercle-go-template/internal/logger"
)

// Package-level validator instance with lazy initialization.
// validator.New() is safe for concurrent use, so we can share one instance.
var (
	validateInstance *validator.Validate
	validateOnce     sync.Once
)

// errorMapPool is a sync.Pool for reusing validation error maps to reduce GC pressure.
// This reduces allocations during request validation by reusing map structures.
var errorMapPool = sync.Pool{
	New: func() any {
		return make(map[string]string)
	},
}

// getErrorMap retrieves a map[string]string from the pool.
func getErrorMap() map[string]string {
	return errorMapPool.Get().(map[string]string)
}

// putErrorMap returns a map to the pool after clearing it.
func putErrorMap(m map[string]string) {
	// Clear the map to prevent data leakage between uses
	for k := range m {
		delete(m, k)
	}
	errorMapPool.Put(m)
}

// responseBufferPool is a sync.Pool for reusing byte slices for JSON marshaling.
// This reduces allocations during response encoding.
var responseBufferPool = sync.Pool{
	New: func() any {
		buf := make([]byte, 0, 512) // Pre-allocate 512 bytes as typical response size
		return &buf
	},
}

// getResponseBuffer retrieves a byte slice from the pool.
func getResponseBuffer() *[]byte {
	return responseBufferPool.Get().(*[]byte)
}

// putResponseBuffer returns a byte slice to the pool after clearing it.
func putResponseBuffer(buf *[]byte) {
	// Reset the slice to empty but keep capacity
	*buf = (*buf)[:0]
	responseBufferPool.Put(buf)
}

// getValidator returns the singleton validator instance.
// Uses sync.Once for thread-safe lazy initialization.
func getValidator() *validator.Validate {
	validateOnce.Do(func() {
		validateInstance = validator.New()
	})
	return validateInstance
}

// UserHandler handles HTTP requests related to users.
type UserHandler struct {
	userUsecase usecase.UserUsecase
	jwtUsecase  authusecase.JWTUsecase
	logger      logger.Logger
}

// NewUserHandler creates a new user handler.
func NewUserHandler(userUc usecase.UserUsecase, jwtUc authusecase.JWTUsecase, log logger.Logger) *UserHandler {
	return &UserHandler{
		userUsecase: userUc,
		jwtUsecase:  jwtUc,
		logger:      log,
	}
}

// RegisterRoutes registers all user-related routes.
func (h *UserHandler) RegisterRoutes(router *echo.Group) {
	// Auth routes
	auth := router.Group("/auth")
	auth.POST("/login", h.Login)

	// User routes
	users := router.Group("/users")
	users.POST("", h.CreateUser)
	users.GET("", h.ListUsers)
	users.GET("/:id", h.GetUser)
	users.PUT("/:id", h.UpdateUser)
	users.DELETE("/:id", h.DeleteUser)
	users.PUT("/:id/password", h.UpdatePassword)
}

// Login handles POST /auth/login - authenticates a user and returns a JWT token.
//
//	@Summary		User login
//	@Description	Authenticate a user with email and password
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.UserLoginRequest	true	"Login credentials"
//	@Success		200		{object}	Response{data=dto.UserLoginResponse}
//	@Failure		400		{object}	Response
//	@Failure		401		{object}	Response
//	@Failure		500		{object}	Response
//	@Router			/auth/login [post]
func (h *UserHandler) Login(c *echo.Context) error {
	var req dto.UserLoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    string(appErr.ErrCodeValidation),
				Message: "Invalid request body",
			},
		})
	}

	// Validate request
	if validationErrors := validateStruct(&req); validationErrors != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    string(appErr.ErrCodeValidation),
				Message: "validation failed",
				Details: formatValidationErrors(validationErrors),
			},
		})
	}

	// Authenticate user
	user, err := h.userUsecase.Authenticate(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		status := appErr.GetStatusCode(err)
		return c.JSON(status, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    string(appErr.ErrCodeUnauthorized),
				Message: "Invalid credentials",
			},
		})
	}

	// Generate token pair
	tokenPair, err := h.jwtUsecase.GenerateTokenPair(user)
	if err != nil {
		h.logger.Error("failed to generate token pair", logger.Error(err))
		return c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    string(appErr.ErrCodeInternal),
				Message: "Failed to generate authentication token",
			},
		})
	}

	return c.JSON(http.StatusOK, Response{
		Success: true,
		Data: dto.UserLoginResponse{
			User:  dto.ToUserResponse(user),
			Token: tokenPair.AccessToken,
		},
	})
}

// CreateUser handles POST /users - creates a new user.
//
//	@Summary		Create user
//	@Description	Create a new user account
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.CreateUserRequest	true	"User creation data"
//	@Success		201		{object}	Response{data=dto.UserResponse}
//	@Failure		400		{object}	Response
//	@Failure		409		{object}	Response
//	@Failure		500		{object}	Response
//	@Router			/users [post]
func (h *UserHandler) CreateUser(c *echo.Context) error {
	var req dto.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    string(appErr.ErrCodeValidation),
				Message: "Invalid request body",
			},
		})
	}

	// Validate request
	if validationErrors := validateStruct(&req); validationErrors != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    string(appErr.ErrCodeValidation),
				Message: "validation failed",
				Details: formatValidationErrors(validationErrors),
			},
		})
	}

	user, err := h.userUsecase.CreateUser(c.Request().Context(), req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    dto.ToUserResponse(user),
	})
}

// GetUser handles GET /users/:id - retrieves a user by ID.
//
//	@Summary		Get user
//	@Description	Get a user by ID
//	@Tags			users
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	Response{data=dto.UserResponse}
//	@Failure		400	{object}	Response
//	@Failure		404	{object}	Response
//	@Failure		500	{object}	Response
//	@Router			/users/{id} [get]
func (h *UserHandler) GetUser(c *echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    string(appErr.ErrCodeValidation),
				Message: "user ID is required",
			},
		})
	}

	user, err := h.userUsecase.GetUser(c.Request().Context(), id)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    dto.ToUserResponse(user),
	})
}

// ListUsers handles GET /users - lists all users with pagination.
//
//	@Summary		List users
//	@Description	Get a paginated list of users
//	@Tags			users
//	@Produce		json
//	@Param			page	query		int	false	"Page number"	    	default(1)
//	@Param			limit	query		int	false	"Items per page"	default(10)
//	@Success		200		{object}	Response{data=dto.ListUsersResponse,meta=MetaInfo}
//	@Failure		500		{object}	Response
//	@Router			/users [get]
func (h *UserHandler) ListUsers(c *echo.Context) error {
	// Parse pagination parameters with defaults
	pageStr := c.QueryParam("page")
	if pageStr == "" {
		pageStr = "1"
	}
	page, _ := strconv.Atoi(pageStr)

	limitStr := c.QueryParam("limit")
	if limitStr == "" {
		limitStr = "10"
	}
	limit, _ := strconv.Atoi(limitStr)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	users, total, err := h.userUsecase.ListUsers(c.Request().Context(), page, limit)
	if err != nil {
		return h.handleError(c, err)
	}

	totalPages := total / limit
	if total%limit > 0 {
		totalPages++
	}

	return c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    dto.ToUserListResponse(users, total, page, limit),
		Meta: &MetaInfo{
			Page:       page,
			PerPage:    limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// UpdateUser handles PUT /users/:id - updates an existing user.
//
//	@Summary		Update user
//	@Description	Update an existing user's information
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"User ID"
//	@Param			request	body		dto.UpdateUserRequest	true	"User update data"
//	@Success		200		{object}	Response{data=dto.UserResponse}
//	@Failure		400		{object}	Response
//	@Failure		404		{object}	Response
//	@Failure		500		{object}	Response
//	@Router			/users/{id} [put]
func (h *UserHandler) UpdateUser(c *echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    string(appErr.ErrCodeValidation),
				Message: "user ID is required",
			},
		})
	}

	var req dto.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    string(appErr.ErrCodeValidation),
				Message: "Invalid request body",
			},
		})
	}

	// Validate request
	if validationErrors := validateStruct(&req); validationErrors != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    string(appErr.ErrCodeValidation),
				Message: "validation failed",
				Details: formatValidationErrors(validationErrors),
			},
		})
	}

	user, err := h.userUsecase.UpdateUser(c.Request().Context(), id, req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    dto.ToUserResponse(user),
	})
}

// DeleteUser handles DELETE /users/:id - deletes a user.
//
//	@Summary		Delete user
//	@Description	Delete a user by ID
//	@Tags			users
//	@Param			id	path		string	true	"User ID"
//	@Success		204	"No Content"
//	@Failure		400	{object}	Response
//	@Failure		404	{object}	Response
//	@Failure		500	{object}	Response
//	@Router			/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    string(appErr.ErrCodeValidation),
				Message: "user ID is required",
			},
		})
	}

	if err := h.userUsecase.DeleteUser(c.Request().Context(), id); err != nil {
		return h.handleError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// UpdatePassword handles PUT /users/:id/password - updates user password.
//
//	@Summary		Update password
//	@Description	Update a user's password
//	@Tags			users
//	@Accept			json
//	@Param			id		path		string						true	"User ID"
//	@Param			request	body		dto.UpdatePasswordRequest	true	"Password update data"
//	@Success		204		"No Content"
//	@Failure		400		{object}	Response
//	@Failure		401		{object}	Response
//	@Failure		404		{object}	Response
//	@Failure		500		{object}	Response
//	@Router			/users/{id}/password [put]
func (h *UserHandler) UpdatePassword(c *echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    string(appErr.ErrCodeValidation),
				Message: "user ID is required",
			},
		})
	}

	var req dto.UpdatePasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    string(appErr.ErrCodeValidation),
				Message: "Invalid request body",
			},
		})
	}

	// Validate request
	if validationErrors := validateStruct(&req); validationErrors != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    string(appErr.ErrCodeValidation),
				Message: "validation failed",
				Details: formatValidationErrors(validationErrors),
			},
		})
	}

	if err := h.userUsecase.UpdatePassword(c.Request().Context(), id, req); err != nil {
		return h.handleError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// handleError converts an error to an appropriate HTTP response.
func (h *UserHandler) handleError(c *echo.Context, err error) error {
	status := appErr.GetStatusCode(err)

	var errorInfo *ErrorInfo
	if appError, ok := err.(*appErr.AppError); ok {
		errorInfo = &ErrorInfo{
			Code:    string(appError.Code),
			Message: appError.Message,
			Details: appError.Details,
		}
	} else {
		errorInfo = &ErrorInfo{
			Code:    string(appErr.ErrCodeInternal),
			Message: "An unexpected error occurred",
			Details: err.Error(),
		}
	}

	return c.JSON(status, Response{
		Success: false,
		Error:   errorInfo,
	})
}

// validateStruct validates a struct using the shared validator instance.
func validateStruct(obj any) map[string]string {
	if err := getValidator().Struct(obj); err != nil {
		result := getErrorMap()
		defer putErrorMap(result)
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, e := range validationErrors {
				field := e.Field()
				result[field] = getErrorMessage(e)
			}
		}
		return result
	}
	return nil
}

// getErrorMessage returns a human-readable error message for a validation error.
func getErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Value is too short, minimum length is " + e.Param()
	case "max":
		return "Value is too long, maximum length is " + e.Param()
	case "len":
		return "Value must be exactly " + e.Param() + " characters"
	case "alphanumspace":
		return "Value can only contain letters, numbers, and spaces"
	default:
		return "Invalid value"
	}
}

// formatValidationErrors converts validation errors to a string.
func formatValidationErrors(errors map[string]string) string {
	result := ""
	for field, msg := range errors {
		if result != "" {
			result += "; "
		}
		result += field + ": " + msg
	}
	return result
}

// Response represents the standard API response structure.
type Response struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Error   *ErrorInfo `json:"error,omitempty"`
	Meta    *MetaInfo  `json:"meta,omitempty"`
}

// ErrorInfo contains error details.
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// MetaInfo contains pagination and metadata.
type MetaInfo struct {
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}
