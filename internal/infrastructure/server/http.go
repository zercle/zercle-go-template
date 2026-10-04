// Package server constructs and configures the shared HTTP server.
package server

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	echomw "github.com/labstack/echo/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	apperrors "github.com/zercle/zercle-go-template/internal/infrastructure/errors"
	"github.com/zercle/zercle-go-template/internal/infrastructure/middleware"
	"github.com/zercle/zercle-go-template/internal/infrastructure/telemetry"
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

// defaultProbeTimeout is the fallback health-probe timeout used when the
// configured value is zero or negative. It caps how long a health probe will
// wait on registered checkers before returning, so a blocking dependency
// cannot hang /healthz or /readyz.
const defaultProbeTimeout = 5 * time.Second

// NewHTTP builds and returns an *echo.Echo with the standard middleware stack
// and shared routes (/healthz, /readyz, /metrics). The tracer provider,
// propagator, and Prometheus gatherer are injected so request spans are
// recorded, inbound trace context is honored, and /metrics serves the same
// registry the meter provider exports into — without depending on
// process-global OTel or Prometheus state.
func NewHTTP(cfg *config.Config, logger *zerolog.Logger, registry *telemetry.Registry, tp trace.TracerProvider, propagator propagation.TextMapPropagator, gatherer prometheus.Gatherer) *echo.Echo {
	e := echo.New()
	e.Validator = &echoValidator{v: validator.New()}
	e.HTTPErrorHandler = errorHandler(logger)
	// Middleware enters in registration order, so AccessLog is registered first
	// to wrap Recover: a panicking request is recovered inside its span and still
	// produces exactly one "http request" line (status 500) rather than being
	// visible only as a recovered-panic log.
	e.Use(middleware.AccessLog(logger))
	e.Use(middleware.Recover(logger))
	e.Use(middleware.RequestID())
	e.Use(middleware.OTel(tp, propagator))
	e.Use(middleware.CORS(cfg))
	if limit := parseBodyLimitBytes(cfg.HTTP.BodyLimit); limit > 0 {
		e.Use(echomw.BodyLimit(limit))
	}

	probeTimeout := cfg.HTTP.HealthProbeTimeout
	if probeTimeout <= 0 {
		probeTimeout = defaultProbeTimeout
	}

	e.GET("/healthz", healthzHandler(registry, logger, probeTimeout))
	e.GET("/readyz", readyzHandler(registry, logger, probeTimeout))
	e.GET("/metrics", echo.WrapHandler(telemetry.MetricsHandler(gatherer)))

	return e
}

// errorHandler is echo's centralized error handler. It routes every error that
// reaches the framework — including router-generated 404/405 errors — through
// the shared apperrors mapper so all failures share the {"error", "message"}
// envelope instead of echo's default {"message"} shape. Errors are logged here
// because echo's default handler does not log.
func errorHandler(logger *zerolog.Logger) echo.HTTPErrorHandler {
	return func(c *echo.Context, err error) {
		if resp, ok := c.Response().(*echo.Response); ok && resp.Committed {
			// The handler already wrote a response; a middleware error on the
			// way out cannot change it.
			return
		}

		var (
			status int
			body   map[string]any
		)
		// Framework errors (router 404/405, body-limit 413) carry only a status,
		// not a domain meaning. echo.StatusCode extracts it from any error
		// implementing HTTPStatusCoder (v5's *echo.HTTPError and its internal
		// httpError both do), so map that status to a shared code.
		if code := echo.StatusCode(err); code != 0 {
			mapped := apperrors.ForHTTPStatus(code)
			status = code
			body = map[string]any{"error": mapped.Code, "message": mapped.Message}
		} else {
			status, body = apperrors.HTTPError(err)
		}

		logger.Error().
			Err(err).
			Str("request_id", middleware.RequestIDFromContext(c)).
			Str("method", c.Request().Method).
			Str("path", c.Request().URL.Path).
			Int("status", status).
			Msg("request failed")

		if c.Request().Method == http.MethodHead {
			_ = c.NoContent(status)
			return
		}
		if jsonErr := c.JSON(status, body); jsonErr != nil {
			logger.Error().Err(jsonErr).Msg("failed to write error response")
		}
	}
}

// healthzHandler returns the liveness handler. It returns 200 on success and
// 500 only if the registry itself reports an unexpected error.
func healthzHandler(registry *telemetry.Registry, logger *zerolog.Logger, probeTimeout time.Duration) echo.HandlerFunc {
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
func readyzHandler(registry *telemetry.Registry, logger *zerolog.Logger, probeTimeout time.Duration) echo.HandlerFunc {
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
	upper = strings.TrimSuffix(upper, "B")
	upper = strings.TrimSuffix(upper, "I")
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
	if n > math.MaxInt64/multiplier {
		return 0
	}
	return n * multiplier
}
