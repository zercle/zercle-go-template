package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/samber/oops"
	"github.com/zercle/zercle-go-template/pkg/utils/response"
)

// RecoveryMiddleware recovers from panics and logs them using zerolog and oops
func RecoveryMiddleware(log zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				// Get stack trace
				stack := debug.Stack()

				// Create oops error for better error handling
				var panicErr error
				if e, ok := r.(error); ok {
					panicErr = e
				} else {
					panicErr = fmt.Errorf("%v", r)
				}

				// Build oops error with context using builder pattern
				builder := oops.
					Code("panic_recovered").
					In("middleware").
					With("request_id", getRequestID(c)).
					With("method", c.Method()).
					With("path", c.Path()).
					With("stack_trace", string(stack)).
					Wrap(panicErr)

				// Log the panic with full context
				log.Error().
					Str("request_id", getRequestID(c)).
					Str("method", c.Method()).
					Str("path", c.Path()).
					Interface("panic", r).
					Str("stack_trace", string(stack)).
					Msg("Panic recovered")

				// Return JSend error response
				err = response.Error(c, fiber.StatusInternalServerError,
					fmt.Sprintf("Internal server error: %s", builder.Error()))
			}
		}()

		return c.Next()
	}
}
