// Package lifecycle holds cross-cutting resource-lifecycle adapters shared by
// infrastructure packages.
package lifecycle

import (
	"context"
	"sync"
)

// Closer adapts a resource's close function to samber/do's
// ShutdownerWithContextAndError interface, so the DI container owns the
// lifecycle of infrastructure clients even when the Application's graceful
// shutdown never runs (e.g. a partial build failure in app.Build).
//
// The close function runs at most once via sync.Once, making repeated shutdowns
// idempotent when both the Application and the container close the same
// resource.
type Closer struct {
	closeFn func() error
	once    sync.Once
	err     error
}

// NewCloser returns a Closer that runs closeFn on the first Shutdown call. A
// nil closeFn is a no-op, so a never-configured resource shuts down cleanly.
func NewCloser(closeFn func() error) *Closer {
	return &Closer{closeFn: closeFn}
}

// Shutdown implements do.ShutdownerWithContextAndError.
func (c *Closer) Shutdown(context.Context) error {
	c.once.Do(func() {
		if c.closeFn == nil {
			return
		}
		c.err = c.closeFn()
	})
	return c.err
}
