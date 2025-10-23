package middleware

import "github.com/gofiber/fiber/v2"

// getRequestID retrieves request ID from context (set by RequestID middleware)
func getRequestID(c *fiber.Ctx) string {
	if id := c.Locals("requestid"); id != nil {
		if rid, ok := id.(string); ok {
			return rid
		}
	}
	return "unknown"
}
