package valkey

import (
	"context"
	"fmt"

	valkeygo "github.com/valkey-io/valkey-go"
)

// valkeyChecker reports Valkey connectivity via the PING command.
type valkeyChecker struct {
	client valkeygo.Client
}

// Name returns the dependency name reported in health output.
func (valkeyChecker) Name() string {
	return "valkey"
}

// Check verifies Valkey is reachable by sending a PING command.
func (c valkeyChecker) Check(ctx context.Context) error {
	if err := c.client.Do(ctx, c.client.B().Ping().Build()).Error(); err != nil {
		return fmt.Errorf("ping valkey: %w", err)
	}
	return nil
}
