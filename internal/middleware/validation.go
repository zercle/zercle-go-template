// Package middleware provides HTTP middleware for validation.
package middleware

import (
	"github.com/gofiber/fiber/v2"
	sharedHandler "github.com/zercle/zercle-go-template/internal/shared/handler/response"
	"github.com/zercle/zercle-go-template/pkg/utils/validator"
)

// ValidateRequest validates the request body against a struct.
// This is a generic middleware factory that can be used to validate specific request types.
func ValidateRequest[T any]() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req T
		if err := c.BodyParser(&req); err != nil {
			return sharedHandler.Fail(c, fiber.StatusBadRequest, fiber.Map{"error": "invalid request body"})
		}

		if err := validator.Validate(&req); err != nil {
			return sharedHandler.Fail(c, fiber.StatusBadRequest, fiber.Map{"error": err.Error()})
		}

		// Store validated request in context for handler to use
		c.Locals("validatedRequest", &req)
		return c.Next()
	}
}

// ParseAndValidate is a helper function to parse and validate request bodies in handlers.
func ParseAndValidate[T any](c *fiber.Ctx, req *T) error {
	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if err := validator.Validate(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return nil
}
