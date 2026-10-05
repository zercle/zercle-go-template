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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/app"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	apiv1 "github.com/zercle/zercle-go-template/pkg/api/v1"
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

	// POST /api/v1/products adds a product to the global catalog pool.
	var product apiv1.ProductResponse
	postJSON(t, client, server.URL+"/api/v1/products", apiv1.CreateProductRequest{
		Name: "cola", PriceCents: 50, Stock: 10,
	}, http.StatusCreated, &product)
	require.NotEmpty(t, product.ID)
	require.Equal(t, int32(50), product.PriceCents)

	// POST /api/v1/machines registers a machine with an initial coin bank deep
	// enough to compose any single purchase's change.
	var machine apiv1.MachineResponse
	postJSON(t, client, server.URL+"/api/v1/machines", apiv1.CreateMachineRequest{
		Label: "lobby", InitialCoins: []int32{25, 25, 50, 100},
	}, http.StatusCreated, &machine)
	require.NotEmpty(t, machine.ID)

	// Exact payment: the sale succeeds and yields no change.
	var exact apiv1.PurchaseResponse
	postJSON(t, client, server.URL+"/api/v1/purchases", apiv1.PurchaseRequest{
		MachineID: machine.ID, ProductID: product.ID, Coins: []int32{50},
	}, http.StatusCreated, &exact)
	require.Equal(t, int32(50), exact.TotalInsertedCents)
	require.Equal(t, int32(0), exact.ChangeCents)
	require.Empty(t, exact.ChangeCoins)

	// Overpayment: change is composed from the machine bank and the coin
	// breakdown sums back to the change amount.
	var change apiv1.PurchaseResponse
	postJSON(t, client, server.URL+"/api/v1/purchases", apiv1.PurchaseRequest{
		MachineID: machine.ID, ProductID: product.ID, Coins: []int32{100},
	}, http.StatusCreated, &change)
	require.Equal(t, int32(50), change.ChangeCents)
	require.NotEmpty(t, change.ChangeCoins)
	var changeSum int32
	for _, coin := range change.ChangeCoins {
		changeSum += coin
	}
	require.Equal(t, change.ChangeCents, changeSum)

	// Insufficient payment: 400 in the shared error envelope.
	assertErrorEnvelope(t, client, server.URL+"/api/v1/purchases", apiv1.PurchaseRequest{
		MachineID: machine.ID, ProductID: product.ID, Coins: []int32{25},
	}, http.StatusBadRequest, apiv1.ErrCodeInvalidInput)

	// Unknown machine: 404, because the product resolves but the bank does not.
	assertErrorEnvelope(t, client, server.URL+"/api/v1/purchases", apiv1.PurchaseRequest{
		MachineID: "00000000-0000-0000-0000-000000000000", ProductID: product.ID, Coins: []int32{50},
	}, http.StatusNotFound, apiv1.ErrCodeNotFound)

	// GET /api/v1/products lists the product just created.
	var list apiv1.ListProductsResponse
	getJSON(t, client, server.URL+"/api/v1/products", http.StatusOK, &list)
	require.NotEmpty(t, list.Products)

	// GET /api/v1/machines/:id retrieves the machine by id.
	var fetched apiv1.MachineResponse
	getJSON(t, client, server.URL+"/api/v1/machines/"+machine.ID, http.StatusOK, &fetched)
	require.Equal(t, machine.ID, fetched.ID)

	// GET /api/v1/reports/summary aggregates catalog, machines, and sales. The
	// assertions use >= because earlier e2e runs may leave rows behind; the two
	// purchases above guarantee the lower bounds.
	var summary apiv1.SummaryResponse
	getJSON(t, client, server.URL+"/api/v1/reports/summary", http.StatusOK, &summary)
	require.GreaterOrEqual(t, summary.Sales.PurchaseCount, int64(2))
	require.GreaterOrEqual(t, summary.Sales.RevenueCents, int64(100))
	require.GreaterOrEqual(t, summary.Catalog.ProductCount, int64(1))
	require.GreaterOrEqual(t, summary.Machines.MachineCount, int64(1))
	require.Contains(t, topMachineLabels(summary.TopMachines), "lobby")

	// A non-positive top value fails request validation: 400 in the envelope.
	assertErrorEnvelopeGET(t, client, server.URL+"/api/v1/reports/summary?top=-1",
		http.StatusBadRequest, apiv1.ErrCodeInvalidInput)

	// An explicit valid top caps the leaderboard length.
	var capped apiv1.SummaryResponse
	getJSON(t, client, server.URL+"/api/v1/reports/summary?top=1", http.StatusOK, &capped)
	require.LessOrEqual(t, len(capped.TopMachines), 1)
}

// topMachineLabels projects the leaderboard entries to their labels.
func topMachineLabels(machines []apiv1.MachineSales) []string {
	labels := make([]string, len(machines))
	for i := range machines {
		labels[i] = machines[i].Label
	}
	return labels
}

// postJSON marshals body as JSON, POSTs it, asserts the status, and decodes a
// non-nil out into the response body.
func postJSON(t *testing.T, client *http.Client, url string, body any, wantStatus int, out any) {
	t.Helper()

	data, err := json.Marshal(body)
	require.NoError(t, err)

	resp, err := client.Post(url, "application/json", bytes.NewReader(data))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, wantStatus, resp.StatusCode)
	if out != nil {
		require.NoError(t, json.NewDecoder(resp.Body).Decode(out))
	}
}

// getJSON GETs url, asserts the status, and decodes a non-nil out.
func getJSON(t *testing.T, client *http.Client, url string, wantStatus int, out any) {
	t.Helper()

	resp, err := client.Get(url)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, wantStatus, resp.StatusCode)
	if out != nil {
		require.NoError(t, json.NewDecoder(resp.Body).Decode(out))
	}
}

// assertErrorEnvelope POSTs body and asserts the response is the shared
// {"error": code, "message": msg} envelope with the expected code and a
// non-empty message.
func assertErrorEnvelope(t *testing.T, client *http.Client, url string, body any, wantStatus int, wantCode string) {
	t.Helper()

	data, err := json.Marshal(body)
	require.NoError(t, err)

	resp, err := client.Post(url, "application/json", bytes.NewReader(data))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, wantStatus, resp.StatusCode)

	var envelope struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&envelope))
	require.Equal(t, wantCode, envelope.Error)
	require.NotEmpty(t, envelope.Message)
}

// assertErrorEnvelopeGET GETs url (no body) and asserts the response is the
// shared {"error": code, "message": msg} envelope with the expected code and a
// non-empty message.
func assertErrorEnvelopeGET(t *testing.T, client *http.Client, url string, wantStatus int, wantCode string) {
	t.Helper()

	resp, err := client.Get(url)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, wantStatus, resp.StatusCode)

	var envelope struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&envelope))
	require.Equal(t, wantCode, envelope.Error)
	require.NotEmpty(t, envelope.Message)
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
