//go:build unit

package lifecycle_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/infrastructure/lifecycle"
)

// TestCloser_RunsOnceAndPropagatesError pins the Closer contract: the wrapped
// close runs exactly once (idempotent shutdown), its error is returned on every
// call, and a nil close is a safe no-op.
func TestCloser_RunsOnceAndPropagatesError(t *testing.T) {
	t.Parallel()

	calls := 0
	boom := errors.New("close failed")
	c := lifecycle.NewCloser(func() error {
		calls++
		return boom
	})

	require.ErrorIs(t, c.Shutdown(context.Background()), boom)
	require.ErrorIs(t, c.Shutdown(context.Background()), boom)
	require.Equal(t, 1, calls, "close must run exactly once")

	require.NoError(t, lifecycle.NewCloser(nil).Shutdown(context.Background()))
}
