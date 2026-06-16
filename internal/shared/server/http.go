// Package server constructs and configures the shared HTTP and gRPC servers.
package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	echomw "github.com/labstack/echo/v5/middleware"
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

// probeTimeout caps how long a health probe will wait on registered checkers
// before returning, so a blocking dependency cannot hang /healthz or /readyz.
const probeTimeout = 5 * time.Second

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
	if limit := parseBodyLimitBytes(cfg.HTTP.BodyLimit); limit > 0 {
		e.Use(echomw.BodyLimit(limit))
	}

	// TODO: apply HTTP read/write/idle timeouts once echo v5 exposes the
	// underlying *http.Server (echo v5 StartConfig.BeforeServeFunc is the
	// supported hook, but it is set where the server is started, not here).
	_ = cfg

	e.GET("/healthz", healthzHandler(registry, logger))
	e.GET("/readyz", readyzHandler(registry, logger))
	e.GET("/metrics", echo.WrapHandler(telemetry.MetricsHandler()))

	return e
}

// healthzHandler returns the liveness handler. It returns 200 on success and
// 500 only if the registry itself reports an unexpected error.
func healthzHandler(registry *telemetry.Registry, logger *zerolog.Logger) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), probeTimeout)
		defer cancel()
		if err := registry.Live(ctx); err != nil {
			logger.Error().Err(err).Str("request_id", middleware.RequestIDFromContext(c)).Msg("liveness check failed")
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}
}

// readyzHandler returns the readiness handler. It returns 200 when all
// readiness checkers pass and 503 with a generic body when any fail. The
// detailed error is logged server-side but never returned to the caller.
func readyzHandler(registry *telemetry.Registry, logger *zerolog.Logger) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), probeTimeout)
		defer cancel()
		if err := registry.Ready(ctx); err != nil {
			logger.Warn().Err(err).Str("request_id", middleware.RequestIDFromContext(c)).Msg("readiness check failed")
			return c.JSON(http.StatusServiceUnavailable, map[string]any{
				"status": "not ready",
			})
		}
		return c.NoContent(http.StatusOK)
	}
}

// parseBodyLimitBytes converts a human-friendly byte size string such as
// "1M" or "512K" into the raw byte count accepted by echo's BodyLimit
// middleware. It returns 0 (i.e. "skip") for empty or unparseable input.
func parseBodyLimitBytes(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	upper := strings.ToUpper(s)
	multiplier := int64(1)
	switch {
	case strings.HasSuffix(upper, "K"):
		multiplier = 1024
		upper = strings.TrimSuffix(upper, "K")
	case strings.HasSuffix(upper, "M"):
		multiplier = 1024 * 1024
		upper = strings.TrimSuffix(upper, "M")
	case strings.HasSuffix(upper, "G"):
		multiplier = 1024 * 1024 * 1024
		upper = strings.TrimSuffix(upper, "G")
	}
	upper = strings.TrimSpace(upper)
	n, err := strconv.ParseInt(upper, 10, 64)
	if err != nil || n <= 0 {
		return 0
	}
	return n * multiplier
}
