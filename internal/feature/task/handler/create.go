package handler

import (
	"errors"

	"github.com/labstack/echo/v5"

	task_entity "github.com/zercle/zercle-go-template/internal/feature/task"
	"github.com/zercle/zercle-go-template/internal/transport/http/response"
)

// Create handles POST /tasks - creates a new task.
// It parses and validates the request body, calls the usecase to create the task,
// and returns 201 Created with the task response on success.
// Error responses:
//   - 400 Bad Request if validation fails
//   - 409 Conflict if task already exists
//   - 500 Internal Server Error for unexpected errors
func (h *handler) Create(c *echo.Context) error {
	var req task_entity.CreateTaskInput

	// Parse request body
	if err := c.Bind(&req); err != nil {
		h.getLogger(c).Warn().Err(err).Msg("failed to parse create task request")
		return response.BadRequest(c, "Invalid request body", err)
	}

	// Validate request
	if err := validateCreateRequest(&req); err != nil {
		h.getLogger(c).Warn().Err(err).Msg("validation failed for create task request")
		return response.ValidationError(c, err)
	}

	// Call usecase
	result, err := h.usecase.Create(c.Request().Context(), &req)
	if err != nil {
		h.getLogger(c).Error().Err(err).Msg("failed to create task")

		// Map domain errors to HTTP responses
		if errors.Is(err, task_entity.ErrTaskAlreadyExists) {
			return response.Conflict(c, "Task already exists")
		}
		if errors.Is(err, task_entity.ErrInvalidTaskPriority) {
			return response.BadRequest(c, "Invalid task priority", err)
		}

		// Default to internal error
		return response.InternalError(c, "An unexpected error occurred")
	}

	h.getLogger(c).Info().Str("task_id", result.ID).Msg("task created successfully")
	return response.Created(c, result)
}

// validateCreateRequest performs custom validation on CreateTaskInput.
// Returns an error if validation fails.
func validateCreateRequest(req *task_entity.CreateTaskInput) error {
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
	if req.Priority != "low" && req.Priority != "medium" && req.Priority != "high" {
		return errors.New("priority must be low, medium, or high")
	}
	if req.UserID == "" {
		return errors.New("user_id is required")
	}
	return nil
}
