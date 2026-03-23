package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"

	user_entity "github.com/zercle/zercle-go-template/internal/feature/user"
)

// Update handles PUT /users/:id - updates an existing user.
// It parses the ID from URL parameter, parses and validates the request body,
// calls the usecase to update the user, and returns 200 OK with the updated user DTO.
// Error responses:
//   - 400 Bad Request if ID is invalid or validation fails
//   - 404 Not Found if user does not exist
//   - 409 Conflict if new email is already registered
//   - 500 Internal Server Error for unexpected errors
func (h *Handler) Update(c *echo.Context) error {
	id := c.Param("id")
	if id == "" {
		h.getLogger(c).Warn().Msg("missing user id parameter")
		return c.JSON(http.StatusBadRequest, map[string]any{
			"status": "fail",
			"data":   map[string]string{"message": "User ID is required"},
		})
	}

	var req user_entity.UpdateUserDTO

	// Parse request body
	if err := c.Bind(&req); err != nil {
		h.getLogger(c).Warn().Err(err).Msg("failed to parse update user request")
		return c.JSON(http.StatusBadRequest, map[string]any{
			"status": "fail",
			"data":   map[string]string{"message": "Invalid request body", "error": err.Error()},
		})
	}

	// Validate request
	if err := validateUpdateRequest(&req); err != nil {
		h.getLogger(c).Warn().Err(err).Msg("validation failed for update user request")
		return c.JSON(http.StatusBadRequest, map[string]any{
			"status": "fail",
			"data":   map[string]any{"validation": map[string]string{"error": err.Error()}},
		})
	}

	// Call usecase
	result, err := h.usecase.Update(c.Request().Context(), id, &req)
	if err != nil {
		h.getLogger(c).Error().Err(err).Str("user_id", id).Msg("failed to update user")

		// Map domain errors to HTTP responses
		if errors.Is(err, user_entity.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound, map[string]any{
				"status": "fail",
				"data":   map[string]string{"message": "User not found"},
			})
		}
		if errors.Is(err, user_entity.ErrDuplicateEmail) {
			return c.JSON(http.StatusConflict, map[string]any{
				"status": "fail",
				"data":   map[string]string{"message": "Email already exists"},
			})
		}
		if errors.Is(err, user_entity.ErrInvalidEmail) {
			return c.JSON(http.StatusBadRequest, map[string]any{
				"status": "fail",
				"data":   map[string]string{"message": "Invalid email format"},
			})
		}

		// Default to internal error
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"status":  "error",
			"message": "An unexpected error occurred",
			"code":    5000,
		})
	}

	h.getLogger(c).Info().Str("user_id", result.ID).Msg("user updated successfully")
	return c.JSON(http.StatusOK, map[string]any{
		"status": "success",
		"data":   result,
	})
}

// validateUpdateRequest performs custom validation on UpdateUserDTO.
// Returns an error if validation fails.
func validateUpdateRequest(req *user_entity.UpdateUserDTO) error {
	// If status is provided, validate it
	if req.Status != nil {
		if *req.Status != "active" && *req.Status != "inactive" && *req.Status != "suspended" {
			return errors.New("status must be active, inactive, or suspended")
		}
	}
	return nil
}
