// Prometheus registry shared by the HTTP metrics and the /metrics handler.
package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// NewPrometheusRegistry returns a registry pre-populated with the standard Go
// and process collectors, so a fresh process exposes go_info and process_*
// without touching prometheus.DefaultRegisterer. The transport registers the
// HTTP metrics into this same registry and serves it from /metrics.
func NewPrometheusRegistry() *prometheus.Registry {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	return reg
}
