//go:build unit

package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/server"
	"github.com/zercle/zercle-go-template/internal/infrastructure/telemetry"
	errcodes "github.com/zercle/zercle-go-template/pkg/api/errcodes"
)

// newTestHTTP builds an echo instance with the shared middleware using a no-op
// tracer provider and TraceContext propagator; these tests exercise routing,
// health, and binding, not span export.
func newTestHTTP(cfg *config.Config, logger *zerolog.Logger, registry *telemetry.Registry) *echo.Echo {
	return server.NewHTTP(cfg, logger, registry, trace.NewTracerProvider(), propagation.TraceContext{}, telemetry.NewPrometheusRegistry())
}

func newTestConfig(t *testing.T) *config.Config {
	t.Helper()

	return &config.Config{
		App: config.AppConfig{ShutdownTimeout: 5 * time.Second},
		HTTP: config.HTTPConfig{
			CORSAllowOrigins: []string{"*"},
			CORSAllowMethods: []string{"GET"},
			CORSAllowHeaders: []string{"X-Request-ID"},
		},
		OTel: config.OTelConfig{Exporter: "none", ServiceName: "test", Sampling: 1.0},
		Log:  config.LogConfig{Level: "debug", Format: "console"},
	}
}

func TestNewHTTP_Healthz(t *testing.T) {
	cfg := newTestConfig(t)
	logger := zerolog.New(nil)
	registry := telemetry.NewRegistry()

	e := newTestHTTP(cfg, &logger, registry)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestNewHTTP_Readyz(t *testing.T) {
	cfg := newTestConfig(t)
	logger := zerolog.New(nil)
	registry := telemetry.NewRegistry()

	e := newTestHTTP(cfg, &logger, registry)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestNewHTTP_Metrics(t *testing.T) {
	cfg := newTestConfig(t)
	logger := zerolog.New(nil)
	registry := telemetry.NewRegistry()

	e := newTestHTTP(cfg, &logger, registry)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "go_info")
}

func TestNewHTTP_ValidatorRegistered(t *testing.T) {
	cfg := newTestConfig(t)
	logger := zerolog.New(nil)
	registry := telemetry.NewRegistry()

	e := newTestHTTP(cfg, &logger, registry)

	require.NotNil(t, e.Validator, "echo validator must be registered")
}

func TestNewHTTP_ValidatorBinding(t *testing.T) {
	cfg := newTestConfig(t)
	logger := zerolog.New(nil)
	registry := telemetry.NewRegistry()

	e := newTestHTTP(cfg, &logger, registry)
	e.POST("/validate", func(c *echo.Context) error {
		var req struct {
			Name string `json:"name" validate:"required"`
		}
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]any{"error": "bind"})
		}
		if err := c.Validate(req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid"})
		}
		return c.JSON(http.StatusOK, req)
	})

	valid := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(`{"name":"ok"}`))
	valid.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, valid)
	require.Equal(t, http.StatusOK, rec.Code)

	invalid := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(`{"name":""}`))
	invalid.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, invalid)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestNewHTTP_RecordsSpanAndLinksParent verifies a served request produces one
// server span named after the route, and that an inbound W3C traceparent is
// adopted as the parent so distributed traces link across services.
func TestNewHTTP_RecordsSpanAndLinksParent(t *testing.T) {
	cfg := newTestConfig(t)
	logger := zerolog.New(nil)
	registry := telemetry.NewRegistry()

	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(trace.WithSyncer(exporter), trace.WithSampler(trace.AlwaysSample()))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	e := server.NewHTTP(cfg, &logger, registry, tp, propagation.TraceContext{}, telemetry.NewPrometheusRegistry())
	e.GET("/items/:id", func(c *echo.Context) error { return c.NoContent(http.StatusNoContent) })

	req := httptest.NewRequest(http.MethodGet, "/items/42", nil)
	req.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	require.Equal(t, "GET /items/:id", spans[0].Name)
	require.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", spans[0].SpanContext.TraceID().String(), "span must join the inbound trace")
	require.Equal(t, "00f067aa0ba902b7", spans[0].Parent.SpanID().String(), "inbound span must be the parent")
}

// TestNewHTTP_PanickedRequestIsAccessLogged pins the shared middleware order:
// a route panic must still yield one access-log line with status 500, which
// only holds when AccessLog wraps Recover.
func TestNewHTTP_PanickedRequestIsAccessLogged(t *testing.T) {
	cfg := newTestConfig(t)
	var buf bytes.Buffer
	logger := zerolog.New(&buf)
	registry := telemetry.NewRegistry()

	e := newTestHTTP(cfg, &logger, registry)
	e.GET("/panic", func(c *echo.Context) error { panic("boom") })

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Contains(t, buf.String(), "http request", "panicked requests must be access-logged")
	require.Contains(t, buf.String(), `"status":500`)
}

// TestNewHTTP_NotFoundUsesErrorEnvelope pins that a router-generated 404 is
// served in the shared {"error","message"} envelope, not echo's default
// {"message":"Not Found"} shape.
func TestNewHTTP_NotFoundUsesErrorEnvelope(t *testing.T) {
	cfg := newTestConfig(t)
	logger := zerolog.New(nil)
	registry := telemetry.NewRegistry()

	e := newTestHTTP(cfg, &logger, registry)

	req := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, errcodes.NotFound, body["error"])
	require.NotEmpty(t, body["message"])
}

// TestNewHTTP_HandlerErrorUsesErrorEnvelope pins the same envelope for an error
// returned from a handler: a generic error becomes INTERNAL, not echo's default
// shape, and the internal message is not leaked to the client.
func TestNewHTTP_HandlerErrorUsesErrorEnvelope(t *testing.T) {
	cfg := newTestConfig(t)
	logger := zerolog.New(nil)
	registry := telemetry.NewRegistry()

	e := newTestHTTP(cfg, &logger, registry)
	e.GET("/boom", func(c *echo.Context) error { return errors.New("kaboom") })

	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, errcodes.Internal, body["error"])
	require.Equal(t, "internal error", body["message"])
	require.NotContains(t, rec.Body.String(), "kaboom", "internal details must not leak to clients")
}
