//go:build unit

package errors

import (
	stderrors "errors"
	"testing"

	"google.golang.org/grpc/codes"
)

var errFoo = stderrors.New("foo")

func withIsolatedSentinels(t *testing.T) {
	t.Helper()
	registeredSentinelsMu.Lock()
	prev := registeredSentinels
	registeredSentinels = nil
	registeredSentinelsMu.Unlock()
	t.Cleanup(func() {
		registeredSentinelsMu.Lock()
		registeredSentinels = prev
		registeredSentinelsMu.Unlock()
	})
}

func TestRegisterSentinel_ReplacesExistingEntryInPlace(t *testing.T) {
	withIsolatedSentinels(t)

	a := &AppError{Code: "A", HTTPStatus: 1, GRPCCode: codes.Unknown}
	b := &AppError{Code: "B", HTTPStatus: 2, GRPCCode: codes.Unknown}

	RegisterSentinel(errFoo, a)

	if got := sentinelFor(errFoo); got != a {
		t.Fatalf("after first registration: sentinelFor(errFoo) = %v, want %v", got, a)
	}

	RegisterSentinel(errFoo, b)

	if got := sentinelFor(errFoo); got != b {
		t.Fatalf("after re-registration: sentinelFor(errFoo) = %v, want %v (older mapping shadowed)", got, b)
	}

	registeredSentinelsMu.RLock()
	defer registeredSentinelsMu.RUnlock()
	if len(registeredSentinels) != 1 {
		t.Fatalf("expected exactly 1 registered entry after duplicate registration, got %d", len(registeredSentinels))
	}
}

func TestRegisterSentinel_AppendsNewSentinel(t *testing.T) {
	withIsolatedSentinels(t)

	a := &AppError{Code: "A", HTTPStatus: 1, GRPCCode: codes.Unknown}
	b := &AppError{Code: "B", HTTPStatus: 2, GRPCCode: codes.Unknown}
	errBar := stderrors.New("bar")

	RegisterSentinel(errFoo, a)
	RegisterSentinel(errBar, b)

	if got := sentinelFor(errFoo); got != a {
		t.Fatalf("sentinelFor(errFoo) = %v, want %v", got, a)
	}
	if got := sentinelFor(errBar); got != b {
		t.Fatalf("sentinelFor(errBar) = %v, want %v", got, b)
	}
}
