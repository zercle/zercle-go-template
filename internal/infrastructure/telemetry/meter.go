// OpenTelemetry meter provider and Prometheus metrics HTTP handler.
package telemetry

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"

	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
)

// NewMeterProvider builds a Prometheus exporter-backed meter provider that
// registers its collector on reg, and returns it together with a shutdown
// function. reg is passed explicitly rather than relying on
// prometheus.DefaultRegisterer: building a second provider in one process would
// otherwise fail with "duplicate metrics collector registration attempted" and
// make /metrics return 500.
func NewMeterProvider(_ *config.Config, reg prometheus.Registerer) (*metric.MeterProvider, func(context.Context) error, error) {
	exporter, err := otelprom.New(otelprom.WithRegisterer(reg))
	if err != nil {
		return nil, nil, fmt.Errorf("create Prometheus exporter: %w", err)
	}

	provider := metric.NewMeterProvider(metric.WithReader(exporter))

	return provider, provider.Shutdown, nil
}

// MetricsHandler returns an http.Handler exposing the metrics gathered in reg
// on /metrics.
func MetricsHandler(reg prometheus.Gatherer) http.Handler {
	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
}

// NewPrometheusRegistry returns a registry pre-populated with the standard Go
// and process collectors, so a fresh process exposes go_info and process_*
// without touching prometheus.DefaultRegisterer.
func NewPrometheusRegistry() *prometheus.Registry {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	return reg
}
