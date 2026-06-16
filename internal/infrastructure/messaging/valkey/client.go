// Package valkey wraps valkey-go client creation for the DI container.
package valkey

import (
	"context"
	"fmt"
	"net"

	valkeygo "github.com/valkey-io/valkey-go"

	"github.com/zercle/zercle-go-template/internal/config"
)

// NewClient returns a connected valkey-go client built from the application
// config. It pings the server before returning and closes the client on ping
// failure. The caller is responsible for calling Close on the returned client.
func NewClient(ctx context.Context, cfg *config.Config) (valkeygo.Client, error) {
	dialer := net.Dialer{Timeout: cfg.DB.ConnectTimeout}

	client, err := valkeygo.NewClient(valkeygo.ClientOption{
		InitAddress: []string{cfg.ValkeyAddr()},
		Password:    cfg.Valkey.Password,
		SelectDB:    cfg.Valkey.DB,
		Dialer:      dialer,
	})
	if err != nil {
		return nil, fmt.Errorf("create valkey client: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, cfg.DB.ConnectTimeout)
	defer cancel()
	if err := client.Do(pingCtx, client.B().Ping().Build()).Error(); err != nil {
		client.Close()
		return nil, fmt.Errorf("ping valkey: %w", err)
	}

	return client, nil
}
