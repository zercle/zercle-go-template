// Package server constructs and configures the shared HTTP and gRPC servers.
package server

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"

	"github.com/zercle/zercle-go-template/internal/config"
	"github.com/zercle/zercle-go-template/internal/shared/middleware"
	"github.com/zercle/zercle-go-template/internal/shared/telemetry"
)

type echoValidator struct {
	v *validator.Validate
}

// Validate implements echo.Validator by delegating to go-playground/validator.
func (cv *echoValidator) Validate(i any) error {
	if err := cv.v.Struct(i); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	return nil
}

// NewHTTP builds and returns an *echo.Echo with the standard middleware stack
// and shared routes (/healthz, /readyz, /metrics).
func NewHTTP(cfg *config.Config, logger *zerolog.Logger, registry *telemetry.Registry) *echo.Echo {
	e := echo.New()
	e.Validator = &echoValidator{v: validator.New()}

	e.Use(middleware.Recover(logger))
	e.Use(middleware.RequestID())
	e.Use(middleware.OTel())
	e.Use(middleware.AccessLog(logger))
	e.Use(middleware.CORS(cfg))

	e.GET("/healthz", healthzHandler(registry, logger))
	e.GET("/readyz", readyzHandler(registry, logger))
	e.GET("/metrics", echo.WrapHandler(telemetry.MetricsHandler()))

	return e
}

// healthzHandler returns the liveness handler. It returns 200 on success and
// 500 only if the registry itself reports an unexpected error.
func healthzHandler(registry *telemetry.Registry, logger *zerolog.Logger) echo.HandlerFunc {
	return func(c *echo.Context) error {
		if err := registry.Live(c.Request().Context()); err != nil {
			logger.Error().Err(err).Str("request_id", middleware.RequestIDFromContext(c)).Msg("liveness check failed")
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}
}

// readyzHandler returns the readiness handler. It returns 200 when all
// readiness checkers pass and 503 with a JSON body listing failing checkers.
func readyzHandler(registry *telemetry.Registry, logger *zerolog.Logger) echo.HandlerFunc {
	return func(c *echo.Context) error {
		if err := registry.Ready(c.Request().Context()); err != nil {
			logger.Warn().Err(err).Str("request_id", middleware.RequestIDFromContext(c)).Msg("readiness check failed")
			return c.JSON(http.StatusServiceUnavailable, map[string]any{
				"status":  "not ready",
				"details": err.Error(),
			})
		}
		return c.NoContent(http.StatusOK)
	}
}
