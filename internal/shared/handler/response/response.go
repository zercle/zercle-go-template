package response

import (
	"github.com/gofiber/fiber/v2"
	sharederrors "github.com/zercle/zercle-go-template/internal/shared/errors"
)

// Response represents a JSend-compliant response
type Response struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// Success sends a success response
func Success(c *fiber.Ctx, status int, data interface{}) error {
	return c.Status(status).JSON(Response{
		Status: "success",
		Data:   data,
	})
}

// Fail sends a fail response
func Fail(c *fiber.Ctx, status int, data interface{}) error {
	return c.Status(status).JSON(Response{
		Status: "fail",
		Data:   data,
	})
}

// Error sends an error response
func Error(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(Response{
		Status:  "error",
		Message: message,
	})
}

// HandleError converts domain errors to HTTP responses
func HandleError(c *fiber.Ctx, err error) error {
	switch err {
	case sharederrors.ErrNotFound:
		return Error(c, fiber.StatusNotFound, "Resource not found")
	case sharederrors.ErrDuplicate:
		return Error(c, fiber.StatusConflict, "Resource already exists")
	case sharederrors.ErrInvalidCreds:
		return Error(c, fiber.StatusUnauthorized, "Invalid credentials")
	default:
		return Error(c, fiber.StatusInternalServerError, "Internal server error")
	}
}
