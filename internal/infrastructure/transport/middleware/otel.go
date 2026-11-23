// OpenTelemetry echo middleware, built on the official echo-otel module. It
// starts a server span named from the matched route, links it to an inbound
// parent context via the supplied propagator, and records the standard OTel
// HTTP server metrics through the meter provider — so /metrics gains those
// series.
package middleware

import (
	echootel "github.com/labstack/echo-otel/v5"
	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// OTel returns echo middleware that creates an OpenTelemetry server span for
// each request. The providers are passed explicitly so the middleware records
// spans and metrics regardless of whether the OTel globals have been installed.
// ServerName is left empty: the server does not know its canonical public
// name, so echo-otel resolves the address from the request.
func OTel(tp trace.TracerProvider, mp metric.MeterProvider, propagator propagation.TextMapPropagator) echo.MiddlewareFunc {
	return echootel.NewMiddlewareWithConfig(echootel.Config{
		TracerProvider: tp,
		MeterProvider:  mp,
		Propagators:    propagator,
	})
}
