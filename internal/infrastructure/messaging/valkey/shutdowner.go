package valkey

import (
	"context"
	"sync"

	valkeygo "github.com/valkey-io/valkey-go"
)

// Shutdowner adapts valkeygo.Client to samber/do's
// ShutdownerWithContextAndError interface so the DI container owns the
// client lifecycle. This guarantees the client is closed by
// injector.Shutdown() even when server.Application.shutdown() never runs —
// e.g. a partial build failure in app.Build, or tests that call Build
// followed by injector.Shutdown.
//
// Close is guarded by sync.Once so repeated shutdown calls are idempotent
// (valkeygo.Client.Close is itself idempotent, but the guard keeps the
// shutdown deterministic and silent in the happy path where both the
// Application and the DI container close the same client).
type Shutdowner struct {
	client valkeygo.Client
	once   sync.Once
	err    error
}

// NewShutdowner wraps client so the DI container can close it.
func NewShutdowner(client valkeygo.Client) *Shutdowner {
	return &Shutdowner{client: client}
}

// Shutdown implements do.ShutdownerWithContextAndError.
func (s *Shutdowner) Shutdown(context.Context) error {
	s.once.Do(func() {
		if s.client == nil {
			return
		}
		s.client.Close()
	})
	return s.err
}
