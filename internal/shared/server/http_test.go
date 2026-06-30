//go:build unit

package server_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/config"
	"github.com/zercle/zercle-go-template/internal/shared/server"
	"github.com/zercle/zercle-go-template/internal/shared/telemetry"
)

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

	e := server.NewHTTP(cfg, &logger, registry)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestNewHTTP_Readyz(t *testing.T) {
	cfg := newTestConfig(t)
	logger := zerolog.New(nil)
	registry := telemetry.NewRegistry()

	e := server.NewHTTP(cfg, &logger, registry)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestNewHTTP_Metrics(t *testing.T) {
	cfg := newTestConfig(t)
	logger := zerolog.New(nil)
	registry := telemetry.NewRegistry()

	e := server.NewHTTP(cfg, &logger, registry)

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

	e := server.NewHTTP(cfg, &logger, registry)

	require.NotNil(t, e.Validator, "echo validator must be registered")
}

func TestNewHTTP_ValidatorBinding(t *testing.T) {
	cfg := newTestConfig(t)
	logger := zerolog.New(nil)
	registry := telemetry.NewRegistry()

	e := server.NewHTTP(cfg, &logger, registry)
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

func TestNewGRPC(t *testing.T) {
	logger := zerolog.New(nil)
	gs := server.NewGRPC(&logger)
	require.NotNil(t, gs)
}
