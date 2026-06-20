//go:build unit

package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/zercle/zercle-go-template/internal/shared/middleware"
)

// tpSpan is a no-op span whose TracerProvider returns the supplied provider.
// It is installed in the request context so the OTel middleware's
// trace.SpanFromContext(ctx).TracerProvider() call resolves to a test provider
// without mutating the global TracerProvider.
type tpSpan struct {
	noop.Span
	tp oteltrace.TracerProvider
}

func (s tpSpan) TracerProvider() oteltrace.TracerProvider { return s.tp }

func newOTelTestEcho(t *testing.T, tp *trace.TracerProvider) (*echo.Echo, *tracetest.InMemoryExporter) {
	t.Helper()
	exporter := tracetest.NewInMemoryExporter()
	*tp = *trace.NewTracerProvider(
		trace.WithSyncer(exporter),
		trace.WithSampler(trace.AlwaysSample()),
	)

	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()
			ctx := oteltrace.ContextWithSpan(req.Context(), tpSpan{tp: tp})
			c.SetRequest(req.WithContext(ctx))
			return next(c)
		}
	})
	e.Use(middleware.OTel())
	return e, exporter
}

func TestOTel_MatchedRouteUsesTemplate(t *testing.T) {
	var tp trace.TracerProvider
	e, exporter := newOTelTestEcho(t, &tp)
	e.GET("/users/:id", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	require.Equal(t, "GET /users/:id", spans[0].Name)

	var hasRoute bool
	var routeVal string
	for _, a := range spans[0].Attributes {
		if a.Key == attribute.Key("http.route") {
			hasRoute = true
			routeVal = a.Value.AsString()
		}
	}
	require.True(t, hasRoute, "expected http.route attribute to be set for matched route")
	require.Equal(t, "/users/:id", routeVal)
	require.NoError(t, tp.Shutdown(context.Background()))
}

func TestOTel_UnmatchedRouteOmitsRouteAttr(t *testing.T) {
	var tp trace.TracerProvider
	e, exporter := newOTelTestEcho(t, &tp)
	// Register an unrelated route so the router exists, but the request path
	// will not match anything (echo returns 404 and c.Path() is empty).
	e.GET("/healthz", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/nonexistent/path", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	require.Equal(t, "GET", spans[0].Name, "unmatched route should use method only as span name")

	for _, a := range spans[0].Attributes {
		require.NotEqual(t, "http.route", string(a.Key), "http.route must NOT be set when no route matched")
	}
	require.NoError(t, tp.Shutdown(context.Background()))
}
