// Sentinel registration allows feature packages to register their own typed
// domain errors so the shared HTTP/gRPC mappers can translate them at the
// transport boundary.
package errors

import (
	"errors"
	"sync"
)

type sentinelEntry struct {
	sentinel error
	app      *AppError
}

// registeredSentinels is an ordered slice (registration order) of domain
// sentinel -> shared AppError mappings. Slice iteration is deterministic; the
// first errors.Is match wins.
var (
	registeredSentinels   []sentinelEntry
	registeredSentinelsMu sync.RWMutex
)

// RegisterSentinel maps a domain sentinel error to a shared AppError. Call from
// a feature package's init or DI registration so the shared mappers can translate
// domain errors without importing feature packages.
//
// The sentinel must be comparable (typically a package-level var). The mapping
// is checked with errors.Is. Registration order is preserved so the first
// registered match wins.
//
// If the same sentinel is registered again, the existing entry's *AppError is
// replaced in place (the entry keeps its original position so registration
// order is preserved). The duplicate check uses errors.Is, which for plain
// (non-wrapping) sentinel vars — the documented registration target, typically
// a package-level var created via errors.New or fmt.Errorf without %w — is
// equivalent to identity comparison while satisfying errorlint.
func RegisterSentinel(sentinel error, app *AppError) {
	registeredSentinelsMu.Lock()
	defer registeredSentinelsMu.Unlock()

	for i := range registeredSentinels {
		if errors.Is(registeredSentinels[i].sentinel, sentinel) {
			registeredSentinels[i].app = app
			return
		}
	}
	registeredSentinels = append(registeredSentinels, sentinelEntry{
		sentinel: sentinel,
		app:      app,
	})
}

// sentinelFor reports whether err matches a registered sentinel using
// errors.Is. Returns the first match in registration order, or nil.
func sentinelFor(err error) *AppError {
	registeredSentinelsMu.RLock()
	defer registeredSentinelsMu.RUnlock()

	for _, entry := range registeredSentinels {
		if errors.Is(err, entry.sentinel) {
			return entry.app
		}
	}

	return nil
}
