package server

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/rs/zerolog"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/middleware"
)

// New creates a new Fiber App with middleware configured.
func New(cfg *config.Config, log zerolog.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		ErrorHandler: globalErrorHandler,
	})

	// Middleware chain with zerolog
	app.Use(middleware.RequestID())             // Generate UUIDv7 request IDs
	app.Use(middleware.RecoveryMiddleware(log)) // Panic recovery with oops
	app.Use(middleware.LoggerMiddleware(log))   // Request logging

	app.Use(cors.New(cors.Config{
		AllowOrigins:     arrayToString(cfg.CORS.AllowedOrigins),
		AllowMethods:     arrayToString(cfg.CORS.AllowedMethods),
		AllowHeaders:     arrayToString(cfg.CORS.AllowedHeaders),
		AllowCredentials: cfg.CORS.AllowedCredentials,
	}))

	return app
}

func globalErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
	}
	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}

func arrayToString(a []string) string {
	return strings.Join(a, ", ")
}
