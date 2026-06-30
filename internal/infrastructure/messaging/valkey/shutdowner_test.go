//go:build unit

package valkey_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/infrastructure/messaging/valkey"
)

// TestShutdowner_NilClientIsSafe verifies that constructing a shutdowner
// with a nil valkey client does not panic and returns nil from Shutdown.
// This mirrors the production scenario where the DI container is asked to
// close a never-configured client.
func TestShutdowner_NilClientIsSafe(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	s := valkey.NewShutdowner(nil)
	require.NotNil(t, s, "shutdowner constructor must return non-nil")

	assert.NoError(t, s.Shutdown(ctx), "shutdown with nil client must return nil")
}

// TestShutdowner_Idempotent verifies that Shutdown remains idempotent
// across repeated calls. Constructing a real valkeygo.Client requires a
// live server (covered by e2e), so we exercise the nil-client path twice
// and confirm no panic and no error propagation. The non-nil-client path
// uses the same sync.Once guard, so this is sufficient to assert
// idempotency at the unit level.
func TestShutdowner_Idempotent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	s := valkey.NewShutdowner(nil)
	require.NotNil(t, s)

	assert.NoError(t, s.Shutdown(ctx), "first shutdown on nil client must return nil")
	assert.NoError(t, s.Shutdown(ctx), "second shutdown must remain a no-op")
}
