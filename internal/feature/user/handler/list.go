package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

	user_entity "github.com/zercle/zercle-go-template/internal/feature/user"
)

// List handles GET /users - lists users with optional filtering and pagination.
// It parses query parameters (limit, offset, email, status), calls the usecase
// to list users, and returns 200 OK with the list result.
// Error responses:
//   - 400 Bad Request if query parameters are invalid
//   - 500 Internal Server Error for unexpected errors
func (h *Handler) List(c *echo.Context) error {
	params := &user_entity.ListParamsDTO{
		Email:  c.QueryParam("email"),
		Status: nil,
	}

	// Parse limit
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		limit, err := strconv.ParseInt(limitStr, 10, 32)
		if err != nil || limit < 1 || limit > 100 {
			h.getLogger(c).Warn().Str("limit", limitStr).Msg("invalid limit parameter")
			return c.JSON(http.StatusBadRequest, map[string]any{
				"status": "fail",
				"data":   map[string]string{"message": "Invalid limit: must be between 1 and 100"},
			})
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
			return c.JSON(http.StatusBadRequest, map[string]any{
				"status": "fail",
				"data":   map[string]string{"message": "Invalid offset: must be non-negative"},
			})
		}
		params.Offset = int32(offset)
	} else {
		params.Offset = 0 // default offset
	}

	// Parse status
	if status := c.QueryParam("status"); status != "" {
		if status != "active" && status != "inactive" && status != "suspended" {
			h.getLogger(c).Warn().Str("status", status).Msg("invalid status parameter")
			return c.JSON(http.StatusBadRequest, map[string]any{
				"status": "fail",
				"data":   map[string]string{"message": "Invalid status: must be active, inactive, or suspended"},
			})
		}
		params.Status = &status
	}

	// Call usecase
	result, err := h.usecase.List(c.Request().Context(), params)
	if err != nil {
		h.getLogger(c).Error().Err(err).Msg("failed to list users")

		// Default to internal error
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"status":  "error",
			"message": "An unexpected error occurred",
			"code":    5000,
		})
	}

	h.getLogger(c).Info().Int64("total", result.Total).Msg("users listed successfully")
	return c.JSON(http.StatusOK, map[string]any{
		"status": "success",
		"data":   result,
	})
}
