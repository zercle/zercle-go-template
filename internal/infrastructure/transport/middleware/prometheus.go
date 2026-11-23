// Prometheus echo middleware and /metrics handler, built on the official
// echo-prometheus module. It records the standard HTTP server metrics into
// the supplied registry and serves them from /metrics.
package middleware

import (
	echoprometheus "github.com/labstack/echo-prometheus"
	"github.com/labstack/echo/v5"
	"github.com/prometheus/client_golang/prometheus"
)

// Prometheus returns echo middleware that records the HTTP request metrics
// (echo_requests_total, echo_request_duration_seconds, echo_request_size_bytes,
// echo_response_size_bytes) into reg. reg is passed explicitly rather than
// relying on prometheus.DefaultRegisterer: the repository's own registry also
// carries the Go and process collectors, so /metrics serves one consistent
// family. DoNotUseRequestPathFor404 keeps an unmatched route from creating one
// series per raw path: all 404s share one url label. The url label is the
// route template (c.Path()), so /api/v1/products/:id stays one series. The
// metric names can be namespaced with Namespace/Subsystem when several
// services share one Prometheus.
func Prometheus(reg prometheus.Registerer) echo.MiddlewareFunc {
	return echoprometheus.NewMiddlewareWithConfig(echoprometheus.MiddlewareConfig{
		Registerer:                reg,
		DoNotUseRequestPathFor404: true,
	})
}

// MetricsHandler returns an echo handler serving the metrics gathered from
// gatherer, in Prometheus exposition format.
func MetricsHandler(gatherer prometheus.Gatherer) echo.HandlerFunc {
	return echoprometheus.NewHandlerWithConfig(echoprometheus.HandlerConfig{
		Gatherer: gatherer,
	})
}
