package handler

import (
	"errors"

	"github.com/labstack/echo/v5"

	task_entity "github.com/zercle/zercle-go-template/internal/feature/task"
	"github.com/zercle/zercle-go-template/internal/transport/http/response"
)

// GetByID handles GET /tasks/:id - retrieves a task by ID.
// It parses the ID from the URL parameter, calls the usecase to get the task,
// and returns 200 OK with the task response on success.
// Error responses:
//   - 400 Bad Request if ID is invalid
//   - 404 Not Found if task does not exist
//   - 500 Internal Server Error for unexpected errors
func (h *handler) GetByID(c *echo.Context) error {
	id := c.Param("id")
	if id == "" {
		h.getLogger(c).Warn().Msg("missing task id parameter")
		return response.BadRequest(c, "Task ID is required", nil)
	}

	// Call usecase
	result, err := h.usecase.Get(c.Request().Context(), id)
	if err != nil {
		h.getLogger(c).Error().Err(err).Str("task_id", id).Msg("failed to get task")

		// Map domain errors to HTTP responses
		if errors.Is(err, task_entity.ErrTaskNotFound) {
			return response.NotFound(c, "Task")
		}

		// Default to internal error
		return response.InternalError(c, "An unexpected error occurred")
	}

	h.getLogger(c).Info().Str("task_id", result.ID).Msg("task retrieved successfully")
	return response.Success(c, result)
}
