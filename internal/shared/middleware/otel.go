// Minimal OpenTelemetry echo middleware. It starts a span from the request
// context, records HTTP attributes, and records errors on the span.
package middleware

import (
	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// OTel returns echo middleware that creates an OpenTelemetry span for each
// request using the trace.TracerProvider available via the request context.
// It sets standard HTTP attributes and ends the span after the handler runs,
// recording any returned error.
func OTel() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			echoCtx := c.Request().Context()
			tracer := trace.SpanFromContext(echoCtx).TracerProvider().Tracer("github.com/zercle/zercle-go-template")

			route := c.Path()
			if route == "" {
				route = c.Request().URL.Path
			}
			newCtx, span := tracer.Start(echoCtx, c.Request().Method+" "+route)
			defer span.End()

			req := c.Request().WithContext(newCtx)
			c.SetRequest(req)

			span.SetAttributes(
				attribute.String("http.method", c.Request().Method),
				attribute.String("http.route", route),
			)

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
