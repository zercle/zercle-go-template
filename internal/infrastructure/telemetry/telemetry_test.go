//go:build unit

package telemetry_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/telemetry"
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
	provider, shutdown, err := telemetry.NewTracerProvider(context.Background(), cfg)
	require.NoError(t, err)
	require.NotNil(t, provider)
	require.Nil(t, shutdown)
}

func TestNewTracer_OTLPRequiresEndpoint(t *testing.T) {
	cfg := &config.Config{OTel: config.OTelConfig{Exporter: "otlp", ServiceName: "test"}}
	provider, shutdown, err := telemetry.NewTracerProvider(context.Background(), cfg)
	require.Error(t, err)
	require.Nil(t, provider)
	require.Nil(t, shutdown)
}

func TestNewMeterProvider(t *testing.T) {
	provider, shutdown, err := telemetry.NewMeterProvider(telemetry.NewPrometheusRegistry())
	require.NoError(t, err)
	require.NotNil(t, provider)
	require.NotNil(t, shutdown)
}

func TestMetricsHandler(t *testing.T) {
	handler := telemetry.MetricsHandler(telemetry.NewPrometheusRegistry())
	require.NotNil(t, handler)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "go_info")
}

// TestMeterProvider_InstancesAreIsolated is the regression guard for duplicate
// Prometheus collector registration: every meter provider gets its own registry
// (mirroring one per app instance), so a second provider in the same process
// cannot collide on prometheus.DefaultRegisterer and break /metrics. Before the
// registry was injectable both providers registered on the default registerer
// and /metrics returned 500 with a duplicate "target_info" collection error.
func TestMeterProvider_InstancesAreIsolated(t *testing.T) {
	reg1 := telemetry.NewPrometheusRegistry()
	p1, _, err := telemetry.NewMeterProvider(reg1)
	require.NoError(t, err)
	reg2 := telemetry.NewPrometheusRegistry()
	p2, _, err := telemetry.NewMeterProvider(reg2)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = p1.Shutdown(context.Background())
		_ = p2.Shutdown(context.Background())
	})

	for i, reg := range []*prometheus.Registry{reg1, reg2} {
		rec := httptest.NewRecorder()
		telemetry.MetricsHandler(reg).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
		require.Equal(t, http.StatusOK, rec.Code, "provider %d must serve /metrics: %s", i, rec.Body.String())
		require.Contains(t, rec.Body.String(), "target_info", "each registry must own the OTel exporter collector")
	}
}
