package valkey

import (
	"github.com/valkey-io/valkey-go/valkeyaside"

	"github.com/zercle/zercle-go-template/internal/infrastructure/lifecycle"
)

// NewShutdownCloser adapts a valkeyaside.CacheAsideClient to samber/do's
// ShutdownerWithContextAndError so the DI container owns the client lifecycle.
// This guarantees the client is closed by injector.Shutdown() even when
// server.Application.shutdown() never runs — e.g. a partial build failure in
// app.Build, or tests that call Build followed by injector.Shutdown.
//
// Close does not return an error; the shared Closer's sync.Once keeps repeated
// shutdowns idempotent when both the Application and the container close the
// same client. Closing the cache-aside client also closes its underlying
// valkeygo.Client.
func NewShutdownCloser(client valkeyaside.CacheAsideClient) *lifecycle.Closer {
	return lifecycle.NewCloser(func() error {
		if client == nil {
			return nil
		}
		client.Close()
		return nil
	})
}
