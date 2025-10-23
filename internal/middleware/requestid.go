package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequestIDConfig defines the config for RequestID middleware
type RequestIDConfig struct {
	// Next defines a function to skip this middleware when returned true.
	Next func(c *fiber.Ctx) bool

	// Header is the header key where to get/set the unique request ID
	// Default: X-Request-ID
	Header string

	// ContextKey is the key used to store the request ID in locals
	// Default: requestid
	ContextKey string

	// Generator defines a function to generate the unique identifier.
	// Default: UUIDv7 generator
	Generator func() string
}

// ConfigDefault is the default config
var ConfigDefault = RequestIDConfig{
	Next:       nil,
	Header:     fiber.HeaderXRequestID,
	ContextKey: "requestid",
	Generator: func() string {
		id, _ := uuid.NewV7()
		return id.String()
	},
}

// RequestID middleware generates a unique ID for each request using UUIDv7
// and adds it to the response header and fiber context.
func RequestID(config ...RequestIDConfig) fiber.Handler {
	// Set default config
	cfg := ConfigDefault

	// Override config if provided
	if len(config) > 0 {
		cfg = config[0]

		// Set default values
		if cfg.Header == "" {
			cfg.Header = ConfigDefault.Header
		}
		if cfg.ContextKey == "" {
			cfg.ContextKey = ConfigDefault.ContextKey
		}
		if cfg.Generator == nil {
			cfg.Generator = ConfigDefault.Generator
		}
	}

	return func(c *fiber.Ctx) error {
		// Don't execute middleware if Next returns true
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}

		// Get ID from request header
		rid := c.Get(cfg.Header)

		// Generate new ID if not present
		if rid == "" {
			rid = cfg.Generator()
		}

		// Set request ID in locals
		c.Locals(cfg.ContextKey, rid)

		// Set request ID in response header
		c.Set(cfg.Header, rid)

		return c.Next()
	}
}
