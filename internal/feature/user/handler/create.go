package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"

	user_entity "github.com/zercle/zercle-go-template/internal/feature/user"
)

// Create handles POST /users - creates a new user.
// It parses and validates the request body, calls the usecase to create the user,
// and returns 201 Created with the user DTO on success.
// Error responses:
//   - 400 Bad Request if validation fails
//   - 409 Conflict if email is already registered
//   - 500 Internal Server Error for unexpected errors
func (h *Handler) Create(c *echo.Context) error {
	var req user_entity.CreateUserDTO

	// Parse request body
	if err := c.Bind(&req); err != nil {
		h.getLogger(c).Warn().Err(err).Msg("failed to parse create user request")
		return c.JSON(http.StatusBadRequest, map[string]any{
			"status": "fail",
			"data":   map[string]string{"message": "Invalid request body", "error": err.Error()},
		})
	}

	// Validate request
	if err := validateCreateRequest(&req); err != nil {
		h.getLogger(c).Warn().Err(err).Msg("validation failed for create user request")
		return c.JSON(http.StatusBadRequest, map[string]any{
			"status": "fail",
			"data":   map[string]any{"validation": map[string]string{"error": err.Error()}},
		})
	}

	// Call usecase
	result, err := h.usecase.Create(c.Request().Context(), &req)
	if err != nil {
		h.getLogger(c).Error().Err(err).Msg("failed to create user")

		// Map domain errors to HTTP responses
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
		if errors.Is(err, user_entity.ErrInvalidPassword) {
			return c.JSON(http.StatusBadRequest, map[string]any{
				"status": "fail",
				"data":   map[string]string{"message": "Invalid password"},
			})
		}

		// Default to internal error
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"status":  "error",
			"message": "An unexpected error occurred",
			"code":    5000,
		})
	}

	h.getLogger(c).Info().Str("user_id", result.ID).Msg("user created successfully")
	return c.JSON(http.StatusCreated, map[string]any{
		"status": "success",
		"data":   result,
	})
}

// validateCreateRequest performs custom validation on CreateUserDTO.
// Returns an error if validation fails.
func validateCreateRequest(req *user_entity.CreateUserDTO) error {
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
