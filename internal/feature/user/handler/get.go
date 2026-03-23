package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"

	user_entity "github.com/zercle/zercle-go-template/internal/feature/user"
)

// GetByID handles GET /users/:id - retrieves a user by ID.
// It parses the ID from the URL parameter, calls the usecase to get the user,
// and returns 200 OK with the user DTO on success.
// Error responses:
//   - 400 Bad Request if ID is invalid
//   - 404 Not Found if user does not exist
//   - 500 Internal Server Error for unexpected errors
func (h *Handler) GetByID(c *echo.Context) error {
	id := c.Param("id")
	if id == "" {
		h.getLogger(c).Warn().Msg("missing user id parameter")
		return c.JSON(http.StatusBadRequest, map[string]any{
			"status": "fail",
			"data":   map[string]string{"message": "User ID is required"},
		})
	}

	// Call usecase
	result, err := h.usecase.GetByID(c.Request().Context(), id)
	if err != nil {
		h.getLogger(c).Error().Err(err).Str("user_id", id).Msg("failed to get user")

		// Map domain errors to HTTP responses
		if errors.Is(err, user_entity.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound, map[string]any{
				"status": "fail",
				"data":   map[string]string{"message": "User not found"},
			})
		}

		// Default to internal error
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"status":  "error",
			"message": "An unexpected error occurred",
			"code":    5000,
		})
	}

	h.getLogger(c).Info().Str("user_id", result.ID).Msg("user retrieved successfully")
	return c.JSON(http.StatusOK, map[string]any{
		"status": "success",
		"data":   result,
	})
}
