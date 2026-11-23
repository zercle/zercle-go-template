//go:build unit

package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/infrastructure/transport/middleware"
)

// TestPrometheus_RegistersAndCountsRequest verifies the middleware registers its
// collectors into the given registry (not prometheus.DefaultRegisterer) and
// counts a request, so a subsequent /metrics handler serves the recorded series.
func TestPrometheus_RegistersAndCountsRequest(t *testing.T) {
	t.Parallel()

	reg := prometheus.NewRegistry()

	e := echo.New()
	e.Use(middleware.Prometheus(reg))
	e.GET("/items/:id", func(c *echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	// A raw-path request on a parameterized route must be counted under the
	// route template, not the concrete path.
	req := httptest.NewRequest(http.MethodGet, "/items/42", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	// The registry must now expose the request counter with the template label.
	// Serve through the shared handler, exercising the same gatherer /metrics uses.
	e.GET("/metrics", middleware.MetricsHandler(reg))
	hrec := httptest.NewRecorder()
	e.ServeHTTP(hrec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	require.Equal(t, http.StatusOK, hrec.Code)

	assert.Contains(t, hrec.Body.String(), "echo_requests_total", "the request counter must be registered into the given registry")
	assert.Contains(t, hrec.Body.String(), `url="/items/:id"`, "the url label must be the route template")
}
