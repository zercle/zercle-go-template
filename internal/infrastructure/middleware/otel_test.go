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
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/zercle/zercle-go-template/internal/infrastructure/middleware"
)

// newOTelTestEcho builds an echo instance whose OTel middleware records into an
// in-memory exporter via an injected SDK provider, so assertions read real
// recorded spans instead of relying on the process-global TracerProvider.
func newOTelTestEcho(t *testing.T) (*echo.Echo, *tracetest.InMemoryExporter, *trace.TracerProvider) {
	t.Helper()

	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
		trace.WithSampler(trace.AlwaysSample()),
	)
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	e := echo.New()
	e.Use(middleware.OTel(tp, propagation.TraceContext{}))
	return e, exporter, tp
}

func TestOTel_MatchedRouteUsesTemplate(t *testing.T) {
	e, exporter, _ := newOTelTestEcho(t)
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
}

func TestOTel_UnmatchedRouteOmitsRouteAttr(t *testing.T) {
	e, exporter, _ := newOTelTestEcho(t)
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
}
