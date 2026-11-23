// OpenTelemetry echo middleware. It starts a server span, links it to an
// inbound parent context via the supplied propagator, records HTTP attributes,
// and records errors on the span.
package middleware

import (
	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// instrumentationScope names the tracer used for server spans.
const instrumentationScope = "github.com/zercle/zercle-go-template"

// OTel returns echo middleware that creates an OpenTelemetry server span for
// each request. The tracer comes from tp; the inbound context is extracted with
// propagator so an upstream service's traceparent becomes the parent span. The
// providers are passed explicitly so the middleware records spans regardless of
// whether the OTel globals have been installed. It sets standard HTTP
// attributes and ends the span after the handler runs, recording any error.
func OTel(tp trace.TracerProvider, propagator propagation.TextMapPropagator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			ctx := propagator.Extract(c.Request().Context(), propagation.HeaderCarrier(c.Request().Header))
			tracer := tp.Tracer(instrumentationScope)

			route := c.Path()
			spanName := c.Request().Method
			if route != "" {
				spanName = c.Request().Method + " " + route
			}
			newCtx, span := tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindServer))
			defer span.End()

			c.SetRequest(c.Request().WithContext(newCtx))

			attrs := []attribute.KeyValue{
				attribute.String("http.method", c.Request().Method),
				attribute.String("url.path", c.Request().URL.Path),
			}
			if route != "" {
				attrs = append(attrs, attribute.String("http.route", route))
			}
			span.SetAttributes(attrs...)

			err := next(c)

			status := responseStatus(c, err)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			span.SetAttributes(attribute.Int("http.response.status_code", status))

			return err
		}
	}
}
