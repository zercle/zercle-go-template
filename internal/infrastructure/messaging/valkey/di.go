// Package valkey wires the Valkey client into the DI container.
package valkey

import (
	"context"

	"github.com/samber/do/v2"
	valkeygo "github.com/valkey-io/valkey-go"

	"github.com/zercle/zercle-go-template/internal/config"
	"github.com/zercle/zercle-go-template/internal/shared/telemetry"
)

// Register provides valkeygo.Client and registers the Valkey readiness checker.
func Register(c do.Injector) error {
	do.Provide(c, func(i do.Injector) (valkeygo.Client, error) {
		cfg := do.MustInvoke[*config.Config](i)
		return NewClient(context.Background(), cfg)
	})

	client, err := do.Invoke[valkeygo.Client](c)
	if err != nil {
		return err
	}

	registry := do.MustInvoke[*telemetry.Registry](c)
	registry.AddReadiness(valkeyChecker{client: client})

	return nil
}
