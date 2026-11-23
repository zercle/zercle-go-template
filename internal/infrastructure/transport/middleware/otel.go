// OpenTelemetry echo middleware, built on the official echo-otel module. It
// starts a server span named from the matched route and links it to an
// inbound parent context via the supplied propagator. Tracing is instrumented
// here; the HTTP metrics come from the Prometheus middleware (prometheus.go),
// so echo-otel records no metrics: a no-op metrics recorder keeps it from
// emitting a second family of HTTP series for the same requests.
package middleware

import (
	echootel "github.com/labstack/echo-otel/v5"
	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// noMetricsRecorder is a MetricsRecorder that records nothing.
type noMetricsRecorder struct{}

func (noMetricsRecorder) Record(*echo.Context, echootel.RecordValues) {}

// OTel returns echo middleware that creates an OpenTelemetry server span for
// each request. The tracer provider is passed explicitly so the middleware
// records spans regardless of whether the OTel globals have been installed. Metrics
// recording is disabled: the Prometheus middleware owns the HTTP metrics, and
// recording both would expose two families of HTTP series for the same
// requests. ServerName is left empty: the server does not know its canonical
// public name, so echo-otel resolves the address from the request.
func OTel(tp trace.TracerProvider, propagator propagation.TextMapPropagator) echo.MiddlewareFunc {
	return echootel.NewMiddlewareWithConfig(echootel.Config{
		TracerProvider: tp,
		Metrics:        noMetricsRecorder{},
		Propagators:    propagator,
	})
}
