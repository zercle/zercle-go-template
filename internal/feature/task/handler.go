package task

import (
	"errors"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"

	"github.com/zercle/zercle-go-template/internal/feature/user"
	http_transport "github.com/zercle/zercle-go-template/internal/transport/http"
)

// Handler handles HTTP requests for the task feature.
type Handler struct {
	service *Service
	logger  zerolog.Logger
}

// NewHandler creates a new task HTTP handler.
func NewHandler(service *Service, logger zerolog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

// RegisterPublicRoutes registers public task routes.
func (h *Handler) RegisterPublicRoutes(g *echo.Group) {}

// RegisterProtectedRoutes registers protected task routes.
func (h *Handler) RegisterProtectedRoutes(g *echo.Group) {
	g.POST("", h.Create)
	g.GET("", h.List)
	g.GET("/:id", h.GetByID)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
}

// Create handles task creation requests.
func (h *Handler) Create(c *echo.Context) error {
	var req CreateTaskInput

	if err := c.Bind(&req); err != nil {
		h.logger.Warn().Err(err).Msg("failed to parse create task request")
		return http_transport.JSONBadRequest(c, "Invalid request body", err)
	}

	if err := validateCreateRequest(&req); err != nil {
		h.logger.Warn().Err(err).Msg("validation failed for create task request")
		return http_transport.JSONValidationError(c, err)
	}

	result, err := h.service.Create(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to create task")
		if errors.Is(err, user.ErrUserNotFound) {
			return http_transport.JSONNotFound(c, "User")
		}
		return http_transport.JSONInternalError(c, "An unexpected error occurred")
	}

	h.logger.Info().Str("task_id", result.ID).Msg("task created successfully")
	return http_transport.JSONCreated(c, result)
}

// GetByID handles retrieving a single task by ID.
func (h *Handler) GetByID(c *echo.Context) error {
	id := c.Param("id")
	if id == "" {
		h.logger.Warn().Msg("missing task id parameter")
		return http_transport.JSONBadRequest(c, "Task ID is required", nil)
	}

	result, err := h.service.Get(c.Request().Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Str("task_id", id).Msg("failed to get task")
		if errors.Is(err, ErrTaskNotFound) {
			return http_transport.JSONNotFound(c, "Task")
		}
		return http_transport.JSONInternalError(c, "An unexpected error occurred")
	}

	h.logger.Info().Str("task_id", result.ID).Msg("task retrieved successfully")
	return http_transport.JSONSuccess(c, result)
}

// List handles listing tasks with optional filtering and pagination.
func (h *Handler) List(c *echo.Context) error {
	params := &ListQuery{}

	limit, err := parseLimitParam(c, 20)
	if err != nil {
		return http_transport.JSONBadRequest(c, "Invalid limit: must be between 1 and 100", nil)
	}
	params.Limit = limit

	offset, err := parseOffsetParam(c)
	if err != nil {
		return http_transport.JSONBadRequest(c, "Invalid offset: must be non-negative", nil)
	}
	params.Offset = offset

	params.UserID = c.QueryParam("user_id")

	status, err := parseStatusParam(c)
	if err != nil {
		return http_transport.JSONBadRequest(c, "Invalid status: must be pending, in_progress, completed, or cancelled", nil)
	}
	params.Status = status

	priority, err := parsePriorityParam(c)
	if err != nil {
		return http_transport.JSONBadRequest(c, "Invalid priority: must be low, medium, or high", nil)
	}
	params.Priority = priority

	result, err := h.service.List(c.Request().Context(), params)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to list tasks")
		return http_transport.JSONInternalError(c, "An unexpected error occurred")
	}

	h.logger.Info().Int64("total", result.Total).Msg("tasks listed successfully")
	return http_transport.JSONSuccess(c, result)
}

func parseLimitParam(c *echo.Context, defaultLimit int32) (int32, error) {
	limitStr := c.QueryParam("limit")
	if limitStr == "" {
		return defaultLimit, nil
	}
	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil || limit < 1 || limit > 100 {
		return 0, errors.New("invalid limit")
	}
	return int32(limit), nil
}

func parseOffsetParam(c *echo.Context) (int32, error) {
	offsetStr := c.QueryParam("offset")
	if offsetStr == "" {
		return 0, nil
	}
	offset, err := strconv.ParseInt(offsetStr, 10, 32)
	if err != nil || offset < 0 {
		return 0, errors.New("invalid offset")
	}
	return int32(offset), nil
}

func parseStatusParam(c *echo.Context) (string, error) {
	status := c.QueryParam("status")
	if status == "" {
		return "", nil
	}
	if status != string(StatusPending) && status != string(StatusInProgress) && status != string(StatusCompleted) && status != string(StatusCancelled) {
		return "", errors.New("invalid status")
	}
	return status, nil
}

func parsePriorityParam(c *echo.Context) (string, error) {
	priority := c.QueryParam("priority")
	if priority == "" {
		return "", nil
	}
	if priority != string(PriorityLow) && priority != string(PriorityMedium) && priority != string(PriorityHigh) {
		return "", errors.New("invalid priority")
	}
	return priority, nil
}

// Update handles task update requests.
func (h *Handler) Update(c *echo.Context) error {
	id := c.Param("id")
	if id == "" {
		h.logger.Warn().Msg("missing task id parameter")
		return http_transport.JSONBadRequest(c, "Task ID is required", nil)
	}

	var req UpdateTaskInput

	if err := c.Bind(&req); err != nil {
		h.logger.Warn().Err(err).Msg("failed to parse update task request")
		return http_transport.JSONBadRequest(c, "Invalid request body", err)
	}

	if err := validateUpdateRequest(&req); err != nil {
		h.logger.Warn().Err(err).Msg("validation failed for update task request")
		return http_transport.JSONValidationError(c, err)
	}

	result, err := h.service.Update(c.Request().Context(), id, &req)
	if err != nil {
		h.logger.Error().Err(err).Str("task_id", id).Msg("failed to update task")
		if errors.Is(err, ErrTaskNotFound) {
			return http_transport.JSONNotFound(c, "Task")
		}
		return http_transport.JSONInternalError(c, "An unexpected error occurred")
	}

	h.logger.Info().Str("task_id", result.ID).Msg("task updated successfully")
	return http_transport.JSONSuccess(c, result)
}

// Delete handles task deletion requests.
func (h *Handler) Delete(c *echo.Context) error {
	id := c.Param("id")
	if id == "" {
		h.logger.Warn().Msg("missing task id parameter")
		return http_transport.JSONBadRequest(c, "Task ID is required", nil)
	}

	err := h.service.Delete(c.Request().Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Str("task_id", id).Msg("failed to delete task")
		if errors.Is(err, ErrTaskNotFound) {
			return http_transport.JSONNotFound(c, "Task")
		}
		return http_transport.JSONInternalError(c, "An unexpected error occurred")
	}

	h.logger.Info().Str("task_id", id).Msg("task deleted successfully")
	return http_transport.JSONSuccess(c, map[string]string{"message": "Task deleted successfully"})
}

func validateCreateRequest(req *CreateTaskInput) error {
	if req.Title == "" {
		return errors.New("title is required")
	}
	if len(req.Title) > 255 {
		return errors.New("title must not exceed 255 characters")
	}
	if len(req.Description) > 2000 {
		return errors.New("description must not exceed 2000 characters")
	}
	if req.Priority == "" {
		return errors.New("priority is required")
	}
	if req.Priority != string(PriorityLow) && req.Priority != string(PriorityMedium) && req.Priority != string(PriorityHigh) {
		return errors.New("priority must be low, medium, or high")
	}
	if req.UserID == "" {
		return errors.New("user_id is required")
	}
	return nil
}

func validateUpdateRequest(req *UpdateTaskInput) error {
	if req.Title != nil {
		if *req.Title == "" {
			return errors.New("title cannot be empty")
		}
		if len(*req.Title) > 255 {
			return errors.New("title must not exceed 255 characters")
		}
	}

	if req.Description != nil {
		if len(*req.Description) > 2000 {
			return errors.New("description must not exceed 2000 characters")
		}
	}

	if req.Status != nil {
		if *req.Status != string(StatusPending) && *req.Status != string(StatusInProgress) && *req.Status != string(StatusCompleted) && *req.Status != string(StatusCancelled) {
			return errors.New("status must be pending, in_progress, completed, or cancelled")
		}
	}

	if req.Priority != nil {
		if *req.Priority != string(PriorityLow) && *req.Priority != string(PriorityMedium) && *req.Priority != string(PriorityHigh) {
			return errors.New("priority must be low, medium, or high")
		}
	}

	return nil
}
