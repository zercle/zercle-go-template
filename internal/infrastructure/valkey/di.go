// Package valkey wires the Valkey client into the DI container.
package valkey

import (
	"context"
	"fmt"

	"github.com/samber/do/v2"

	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/telemetry"
)

// Register provides the cache-aside client, its underlying valkeygo.Client, and
// registers the Valkey readiness checker. The ctx drives the initial client
// construction so startup cancellation/timeouts propagate.
func Register(ctx context.Context, c do.Injector) error {
	cfg, err := do.Invoke[*config.Config](c)
	if err != nil {
		return fmt.Errorf("resolve config: %w", err)
	}

	registry, err := do.Invoke[*telemetry.Registry](c)
	if err != nil {
		return fmt.Errorf("resolve health registry: %w", err)
	}

	aside, err := NewCacheAside(ctx, cfg)
	if err != nil {
		return err
	}
	do.ProvideValue(c, aside)
	do.ProvideValue(c, aside.Client())
	do.ProvideValue(c, NewShutdownCloser(aside))

	registry.AddReadiness(valkeyChecker{client: aside.Client()})

	return nil
}
