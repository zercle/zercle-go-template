// Sentinel registration allows feature packages to register their own typed
// domain errors so the shared HTTP/gRPC mappers can translate them at the
// transport boundary.
package errors

import (
	"errors"
	"sync"
)

// registeredSentinels maps a domain sentinel error to a shared AppError.
var (
	registeredSentinels   = make(map[error]*AppError)
	registeredSentinelsMu sync.RWMutex
)

// RegisterSentinel maps a domain sentinel error to a shared AppError. Call from
// a feature package's init or DI registration so the shared mappers can translate
// domain errors without importing feature packages.
//
// The sentinel must be comparable (typically a package-level var). The mapping
// is checked with errors.Is.
func RegisterSentinel(sentinel error, app *AppError) {
	registeredSentinelsMu.Lock()
	defer registeredSentinelsMu.Unlock()

	registeredSentinels[sentinel] = app
}

// sentinelFor reports whether err matches a registered sentinel using
// errors.Is.
func sentinelFor(err error) *AppError {
	registeredSentinelsMu.RLock()
	defer registeredSentinelsMu.RUnlock()

	for sentinel, app := range registeredSentinels {
		if errors.Is(err, sentinel) {
			return app
		}
	}

	return nil
}
