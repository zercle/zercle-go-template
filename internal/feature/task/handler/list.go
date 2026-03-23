package handler

import (
	"strconv"

	"github.com/labstack/echo/v5"

	task_entity "github.com/zercle/zercle-go-template/internal/feature/task"
	"github.com/zercle/zercle-go-template/internal/transport/http/response"
)

// List handles GET /tasks - lists tasks with optional filtering and pagination.
// It parses query parameters (limit, offset, user_id, status, priority), calls the usecase
// to list tasks, and returns 200 OK with the list result.
// Error responses:
//   - 400 Bad Request if query parameters are invalid
//   - 500 Internal Server Error for unexpected errors
func (h *handler) List(c *echo.Context) error {
	params := &task_entity.ListParamsDTO{}

	// Parse limit
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		limit, err := strconv.ParseInt(limitStr, 10, 32)
		if err != nil || limit < 1 || limit > 100 {
			h.getLogger(c).Warn().Str("limit", limitStr).Msg("invalid limit parameter")
			return response.BadRequest(c, "Invalid limit: must be between 1 and 100", nil)
		}
		params.Limit = int32(limit)
	} else {
		params.Limit = 20 // default limit
	}

	// Parse offset
	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		offset, err := strconv.ParseInt(offsetStr, 10, 32)
		if err != nil || offset < 0 {
			h.getLogger(c).Warn().Str("offset", offsetStr).Msg("invalid offset parameter")
			return response.BadRequest(c, "Invalid offset: must be non-negative", nil)
		}
		params.Offset = int32(offset)
	} else {
		params.Offset = 0 // default offset
	}

	// Parse user_id
	params.UserID = c.QueryParam("user_id")

	// Parse status
	if status := c.QueryParam("status"); status != "" {
		if status != "pending" && status != "in_progress" && status != "completed" && status != "cancelled" {
			h.getLogger(c).Warn().Str("status", status).Msg("invalid status parameter")
			return response.BadRequest(c, "Invalid status: must be pending, in_progress, completed, or cancelled", nil)
		}
		params.Status = status
	}

	// Parse priority
	if priority := c.QueryParam("priority"); priority != "" {
		if priority != "low" && priority != "medium" && priority != "high" {
			h.getLogger(c).Warn().Str("priority", priority).Msg("invalid priority parameter")
			return response.BadRequest(c, "Invalid priority: must be low, medium, or high", nil)
		}
		params.Priority = priority
	}

	// Call usecase
	result, err := h.usecase.List(c.Request().Context(), params)
	if err != nil {
		h.getLogger(c).Error().Err(err).Msg("failed to list tasks")

		// Default to internal error
		return response.InternalError(c, "An unexpected error occurred")
	}

	h.getLogger(c).Info().Int64("total", result.Total).Msg("tasks listed successfully")
	return response.Success(c, result)
}
