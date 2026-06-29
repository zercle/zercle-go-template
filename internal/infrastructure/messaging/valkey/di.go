// Package valkey wires the Valkey client into the DI container.
package valkey

import (
	"context"
	"fmt"

	"github.com/samber/do/v2"

	"github.com/zercle/zercle-go-template/internal/config"
	"github.com/zercle/zercle-go-template/internal/shared/telemetry"
)

// Register provides valkeygo.Client and registers the Valkey readiness
// checker. The ctx is used to drive the initial client construction so
// startup cancellation/timeouts propagate.
func Register(ctx context.Context, c do.Injector) error {
	cfg, err := do.Invoke[*config.Config](c)
	if err != nil {
		return fmt.Errorf("resolve config: %w", err)
	}

	registry, err := do.Invoke[*telemetry.Registry](c)
	if err != nil {
		return fmt.Errorf("resolve health registry: %w", err)
	}

	client, err := NewClient(ctx, cfg)
	if err != nil {
		return err
	}
	do.ProvideValue(c, client)
	do.ProvideValue(c, NewShutdowner(client))

	registry.AddReadiness(valkeyChecker{client: client})

	return nil
}
