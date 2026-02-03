package user

import (
	"errors"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"

	http_transport "github.com/zercle/zercle-go-template/internal/transport/http"
)

// Handler handles HTTP requests for the user feature.
type Handler struct {
	service *Service
	logger  zerolog.Logger
}

// NewHandler creates a new user HTTP handler.
func NewHandler(service *Service, logger zerolog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

// RegisterPublicRoutes registers public user routes.
func (h *Handler) RegisterPublicRoutes(g *echo.Group) {
	g.GET("", h.List)
	g.GET("/:id", h.GetByID)
}

// RegisterProtectedRoutes registers protected user routes.
func (h *Handler) RegisterProtectedRoutes(g *echo.Group) {
	g.POST("", h.Create)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
}

// Create godoc
// @Summary Create a new user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateUserInput true "User data"
// @Success 201 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 409 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /users [post]
func (h *Handler) Create(c *echo.Context) error {
	var req CreateUserInput
	if err := c.Bind(&req); err != nil {
		h.logger.Warn().Err(err).Msg("failed to parse create user request")
		return http_transport.JSONBadRequest(c, "Invalid request body", err)
	}
	if err := validateCreateRequest(&req); err != nil {
		h.logger.Warn().Err(err).Msg("validation failed for create user request")
		return http_transport.JSONValidationError(c, err)
	}
	result, err := h.service.Create(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to create user")
		if errors.Is(err, ErrDuplicateEmail) {
			return http_transport.JSONConflict(c, "Email already exists")
		}
		return http_transport.JSONInternalError(c, "An unexpected error occurred")
	}
	h.logger.Info().Str("user_id", result.ID).Msg("user created successfully")
	return http_transport.JSONCreated(c, result)
}

// GetByID handles retrieving a user by ID.
func (h *Handler) GetByID(c *echo.Context) error {
	id := c.Param("id")
	if id == "" {
		h.logger.Warn().Msg("missing user id parameter")
		return http_transport.JSONBadRequest(c, "User ID is required", nil)
	}
	result, err := h.service.GetByID(c.Request().Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Str("user_id", id).Msg("failed to get user")
		if errors.Is(err, ErrUserNotFound) {
			return http_transport.JSONNotFound(c, "User")
		}
		return http_transport.JSONInternalError(c, "An unexpected error occurred")
	}
	h.logger.Info().Str("user_id", result.ID).Msg("user retrieved successfully")
	return http_transport.JSONSuccess(c, result)
}

// List handles listing users.
func (h *Handler) List(c *echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	params := &ListQuery{
		Email:  c.QueryParam("email"),
		Limit:  int32(limit),
		Offset: int32(offset),
	}
	if status := c.QueryParam("status"); status != "" {
		params.Status = &status
	}
	result, err := h.service.List(c.Request().Context(), params)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to list users")
		return http_transport.JSONInternalError(c, "An unexpected error occurred")
	}
	return http_transport.JSONSuccess(c, result)
}

// Update handles user update requests.
func (h *Handler) Update(c *echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return http_transport.JSONBadRequest(c, "User ID is required", nil)
	}
	var req UpdateUserInput
	if err := c.Bind(&req); err != nil {
		return http_transport.JSONBadRequest(c, "Invalid request body", err)
	}
	result, err := h.service.Update(c.Request().Context(), id, &req)
	if err != nil {
		h.logger.Error().Err(err).Str("user_id", id).Msg("failed to update user")
		if errors.Is(err, ErrUserNotFound) {
			return http_transport.JSONNotFound(c, "User")
		}
		if errors.Is(err, ErrDuplicateEmail) {
			return http_transport.JSONConflict(c, "Email already exists")
		}
		return http_transport.JSONInternalError(c, "An unexpected error occurred")
	}
	return http_transport.JSONSuccess(c, result)
}

// Delete handles user deletion requests.
func (h *Handler) Delete(c *echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return http_transport.JSONBadRequest(c, "User ID is required", nil)
	}
	if err := h.service.Delete(c.Request().Context(), id); err != nil {
		h.logger.Error().Err(err).Str("user_id", id).Msg("failed to delete user")
		if errors.Is(err, ErrUserNotFound) {
			return http_transport.JSONNotFound(c, "User")
		}
		return http_transport.JSONInternalError(c, "An unexpected error occurred")
	}
	return http_transport.JSONSuccess(c, map[string]string{"message": "User deleted successfully"})
}

func validateCreateRequest(req *CreateUserInput) error {
	if req.Email == "" {
		return errors.New("email is required")
	}
	if req.Password == "" {
		return errors.New("password is required")
	}
	if len(req.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if req.FirstName == "" {
		return errors.New("first_name is required")
	}
	if req.LastName == "" {
		return errors.New("last_name is required")
	}
	return nil
}
