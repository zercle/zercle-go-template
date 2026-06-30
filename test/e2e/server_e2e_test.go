//go:build e2e
// +build e2e

// Package e2e_test boots the full application against real infrastructure.
//
// To run these tests locally:
//
//	docker compose up -d postgres valkey
//	go test -tags=e2e ./test/e2e/...
//
// If postgres or valkey are unreachable the tests skip cleanly instead of
// failing.
package e2e_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/app"
	"github.com/zercle/zercle-go-template/internal/config"
)

func TestServer_EndToEnd(t *testing.T) {
	cfg, err := config.Load()
	require.NoError(t, err)

	if !infraReachable(t, cfg) {
		t.Skip("requires: docker compose up postgres valkey")
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	application, injector, err := app.Build(ctx, cfg)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := injector.Shutdown(); err != nil {
			t.Logf("injector shutdown error: %v", err)
		}
	})

	require.NotNil(t, application.Echo(), "application.Echo() must be resolved before running httptest")

	server := httptest.NewServer(application.Echo())
	t.Cleanup(server.Close)

	go func() {
		if err := application.Run(ctx); err != nil {
			t.Logf("application run stopped: %v", err)
		}
	}()

	// Wait until the background goroutine has started the real HTTP listener.
	select {
	case <-application.HasHTTPStarted():
	case <-time.After(2 * time.Second):
		t.Fatal("application HTTP server never started")
	}

	client := server.Client()

	// Liveness should always report 200.
	resp, err := client.Get(server.URL + "/healthz")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	// Readiness should report 200 once DB and Valkey are healthy.
	require.Eventually(t, func() bool {
		r, err := client.Get(server.URL + "/readyz")
		if err != nil {
			return false
		}
		_ = r.Body.Close()
		return r.StatusCode == http.StatusOK
	}, 5*time.Second, 250*time.Millisecond, "readiness probe never passed")

	// POST /api/v1/items creates an item.
	resp, err = client.Post(server.URL+"/api/v1/items", "application/json", strings.NewReader(`{"name":"stub"}`))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	_ = resp.Body.Close()

	// GET /api/v1/items/:id retrieves it.
	resp, err = client.Get(server.URL + "/api/v1/items")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()
}

// infraReachable returns true when both postgres and valkey respond to TCP
// probes. It is used to decide whether to skip the e2e suite because the
// required backing services are not running.
func infraReachable(t *testing.T, cfg *config.Config) bool {
	t.Helper()

	dbAddr := fmt.Sprintf("%s:%d", cfg.DB.Host, cfg.DB.Port)
	valkeyAddr := fmt.Sprintf("%s:%d", cfg.Valkey.Host, cfg.Valkey.Port)

	dbOK := tcpReachable(dbAddr, 2*time.Second)
	valkeyOK := tcpReachable(valkeyAddr, 2*time.Second)

	return dbOK && valkeyOK
}

func tcpReachable(addr string, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
