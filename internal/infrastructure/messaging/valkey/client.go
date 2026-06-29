// Package valkey wraps valkey-go client creation for the DI container.
package valkey

import (
	"context"
	"fmt"
	"net"
	"time"

	valkeygo "github.com/valkey-io/valkey-go"

	"github.com/zercle/zercle-go-template/internal/config"
)

const defaultValkeyConnectTimeout = 5 * time.Second

// NewClient returns a connected valkey-go client built from the application
// config. It pings the server before returning and closes the client on ping
// failure. The caller is responsible for calling Close on the returned client.
func NewClient(ctx context.Context, cfg *config.Config) (valkeygo.Client, error) {
	connectTimeout := cfg.Valkey.ConnectTimeout
	if connectTimeout <= 0 {
		connectTimeout = defaultValkeyConnectTimeout
	}

	dialer := net.Dialer{Timeout: connectTimeout}

	client, err := valkeygo.NewClient(valkeygo.ClientOption{
		InitAddress: []string{cfg.ValkeyAddr()},
		Password:    cfg.Valkey.Password,
		SelectDB:    cfg.Valkey.DB,
		Dialer:      dialer,
	})
	if err != nil {
		return nil, fmt.Errorf("create valkey client for %s: %w", cfg.ValkeyAddr(), err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()
	if err := client.Do(pingCtx, client.B().Ping().Build()).Error(); err != nil {
		client.Close()
		return nil, fmt.Errorf("ping valkey %s: %w", cfg.ValkeyAddr(), err)
	}

	return client, nil
}
