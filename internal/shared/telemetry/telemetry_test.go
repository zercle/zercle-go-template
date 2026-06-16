//go:build unit

package telemetry_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/config"
	"github.com/zercle/zercle-go-template/internal/shared/telemetry"
)

func TestNewLogger_JSON(t *testing.T) {
	cfg := &config.Config{Log: config.LogConfig{Level: "info", Format: "json"}}
	logger, err := telemetry.NewLogger(cfg)
	require.NoError(t, err)
	require.NotNil(t, logger)
}

func TestNewLogger_Console(t *testing.T) {
	cfg := &config.Config{Log: config.LogConfig{Level: "debug", Format: "console"}}
	logger, err := telemetry.NewLogger(cfg)
	require.NoError(t, err)
	require.NotNil(t, logger)
}

func TestNewLogger_InvalidLevel(t *testing.T) {
	cfg := &config.Config{Log: config.LogConfig{Level: "unknown", Format: "json"}}
	logger, err := telemetry.NewLogger(cfg)
	require.Error(t, err)
	require.Nil(t, logger)
}

func TestNewTracer_None(t *testing.T) {
	cfg := &config.Config{OTel: config.OTelConfig{Exporter: "none", ServiceName: "test"}}
	provider, shutdown, err := telemetry.NewTracer(context.Background(), cfg)
	require.NoError(t, err)
	require.NotNil(t, provider)
	require.Nil(t, shutdown)
}

func TestNewTracer_OTLPRequiresEndpoint(t *testing.T) {
	cfg := &config.Config{OTel: config.OTelConfig{Exporter: "otlp", ServiceName: "test"}}
	provider, shutdown, err := telemetry.NewTracer(context.Background(), cfg)
	require.Error(t, err)
	require.Nil(t, provider)
	require.Nil(t, shutdown)
}

func TestNewMeterProvider(t *testing.T) {
	cfg := &config.Config{OTel: config.OTelConfig{Exporter: "none", ServiceName: "test"}}
	provider, shutdown, err := telemetry.NewMeterProvider(cfg)
	require.NoError(t, err)
	require.NotNil(t, provider)
	require.NotNil(t, shutdown)
}

func TestMetricsHandler(t *testing.T) {
	handler := telemetry.MetricsHandler()
	require.NotNil(t, handler)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "go_info")
}
