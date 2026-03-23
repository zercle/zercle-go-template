package handler

import (
	"errors"

	"github.com/labstack/echo/v5"

	task_entity "github.com/zercle/zercle-go-template/internal/feature/task"
	"github.com/zercle/zercle-go-template/internal/transport/http/response"
)

// Delete handles DELETE /tasks/:id - deletes a task.
// It parses the ID from URL parameter, calls the usecase to delete the task,
// and returns 204 No Content on success.
// Error responses:
//   - 400 Bad Request if ID is invalid
//   - 404 Not Found if task does not exist
//   - 500 Internal Server Error for unexpected errors
func (h *handler) Delete(c *echo.Context) error {
	id := c.Param("id")
	if id == "" {
		h.getLogger(c).Warn().Msg("missing task id parameter")
		return response.BadRequest(c, "Task ID is required", nil)
	}

	// Call usecase
	err := h.usecase.Delete(c.Request().Context(), id)
	if err != nil {
		h.getLogger(c).Error().Err(err).Str("task_id", id).Msg("failed to delete task")

		// Map domain errors to HTTP responses
		if errors.Is(err, task_entity.ErrTaskNotFound) {
			return response.NotFound(c, "Task")
		}

		// Default to internal error
		return response.InternalError(c, "An unexpected error occurred")
	}

	h.getLogger(c).Info().Str("task_id", id).Msg("task deleted successfully")
	return response.NoContent(c)
}
