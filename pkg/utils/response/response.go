package response

import (
	"github.com/gofiber/fiber/v2"
	domerrors "github.com/zercle/zercle-go-template/internal/core/domain/errors"
)

// Response represents a JSend response.
type Response struct {
	Status  string `json:"status"` // success, fail, error
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"` // Only for error
}

// Success returns a JSend success response.
func Success(c *fiber.Ctx, code int, data any) error {
	return c.Status(code).JSON(Response{
		Status: "success",
		Data:   data,
	})
}

// Fail returns a JSend fail response (validation errors, client errors).
func Fail(c *fiber.Ctx, code int, data any) error {
	return c.Status(code).JSON(Response{
		Status: "fail",
		Data:   data,
	})
}

// Error returns a JSend error response (server errors, unhandled exceptions).
func Error(c *fiber.Ctx, code int, message string) error {
	return c.Status(code).JSON(Response{
		Status:  "error",
		Message: message,
	})
}

// HandleError maps domain errors to JSend responses.
func HandleError(c *fiber.Ctx, err error) error {
	switch err {
	case domerrors.ErrNotFound:
		return Fail(c, fiber.StatusNotFound, fiber.Map{"error": err.Error()})
	case domerrors.ErrDuplicate:
		return Fail(c, fiber.StatusConflict, fiber.Map{"error": err.Error()})
	case domerrors.ErrInvalidCreds:
		return Fail(c, fiber.StatusUnauthorized, fiber.Map{"error": err.Error()})
	case domerrors.ErrUnauthorized:
		return Fail(c, fiber.StatusForbidden, fiber.Map{"error": err.Error()})
	default:
		// Check for unmapped errors or internal server errors
		return Error(c, fiber.StatusInternalServerError, "Internal Server Error")
	}
}
