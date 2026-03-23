package handler

import (
	"errors"

	"github.com/labstack/echo/v5"

	task_entity "github.com/zercle/zercle-go-template/internal/feature/task"
	"github.com/zercle/zercle-go-template/internal/transport/http/response"
)

// Update handles PUT /tasks/:id - updates an existing task.
// It parses the ID from URL parameter, parses and validates the request body,
// calls the usecase to update the task, and returns 200 OK with the updated task response.
// Error responses:
//   - 400 Bad Request if ID is invalid or validation fails
//   - 404 Not Found if task does not exist
//   - 500 Internal Server Error for unexpected errors
func (h *handler) Update(c *echo.Context) error {
	id := c.Param("id")
	if id == "" {
		h.getLogger(c).Warn().Msg("missing task id parameter")
		return response.BadRequest(c, "Task ID is required", nil)
	}

	var req task_entity.UpdateTaskInput

	// Parse request body
	if err := c.Bind(&req); err != nil {
		h.getLogger(c).Warn().Err(err).Msg("failed to parse update task request")
		return response.BadRequest(c, "Invalid request body", err)
	}

	// Validate request
	if err := validateUpdateRequest(&req); err != nil {
		h.getLogger(c).Warn().Err(err).Msg("validation failed for update task request")
		return response.ValidationError(c, err)
	}

	// Call usecase
	result, err := h.usecase.Update(c.Request().Context(), id, &req)
	if err != nil {
		h.getLogger(c).Error().Err(err).Str("task_id", id).Msg("failed to update task")

		// Map domain errors to HTTP responses
		if errors.Is(err, task_entity.ErrTaskNotFound) {
			return response.NotFound(c, "Task")
		}
		if errors.Is(err, task_entity.ErrInvalidTaskStatus) {
			return response.BadRequest(c, "Invalid task status", err)
		}
		if errors.Is(err, task_entity.ErrInvalidTaskPriority) {
			return response.BadRequest(c, "Invalid task priority", err)
		}

		// Default to internal error
		return response.InternalError(c, "An unexpected error occurred")
	}

	h.getLogger(c).Info().Str("task_id", result.ID).Msg("task updated successfully")
	return response.Success(c, result)
}

// validateUpdateRequest performs custom validation on UpdateTaskInput.
// Returns an error if validation fails.
func validateUpdateRequest(req *task_entity.UpdateTaskInput) error {
	// Validate title if provided
	if req.Title != nil {
		if *req.Title == "" {
			return errors.New("title cannot be empty")
		}
		if len(*req.Title) > 255 {
			return errors.New("title must not exceed 255 characters")
		}
	}

	// Validate description if provided
	if req.Description != nil {
		if len(*req.Description) > 2000 {
			return errors.New("description must not exceed 2000 characters")
		}
	}

	// Validate status if provided
	if req.Status != nil {
		if *req.Status != "pending" && *req.Status != "in_progress" && *req.Status != "completed" && *req.Status != "cancelled" {
			return errors.New("status must be pending, in_progress, completed, or cancelled")
		}
	}

	// Validate priority if provided
	if req.Priority != nil {
		if *req.Priority != "low" && *req.Priority != "medium" && *req.Priority != "high" {
			return errors.New("priority must be low, medium, or high")
		}
	}

	return nil
}
