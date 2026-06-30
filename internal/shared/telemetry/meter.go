// OpenTelemetry meter provider and Prometheus metrics HTTP handler.
package telemetry

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"

	"github.com/zercle/zercle-go-template/internal/config"
)

// NewMeterProvider builds a Prometheus exporter-backed meter provider and
// returns it together with a shutdown function.
func NewMeterProvider(_ *config.Config) (*metric.MeterProvider, func(context.Context) error, error) {
	exporter, err := prometheus.New()
	if err != nil {
		return nil, nil, fmt.Errorf("create Prometheus exporter: %w", err)
	}

	provider := metric.NewMeterProvider(metric.WithReader(exporter))

	return provider, provider.Shutdown, nil
}

// MetricsHandler returns an http.Handler that exposes collected Prometheus
// metrics on /metrics.
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
