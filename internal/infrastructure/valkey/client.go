// Package valkey wires the Valkey client and its cache-aside facade into the
// DI container. NewClient builds the raw valkey-go client used for health checks
// and command access; NewCacheAside builds the client-side-caching cache-aside
// wrapper features decorate their repositories with.
package valkey

import (
	"context"
	"fmt"
	"net"
	"time"

	valkeygo "github.com/valkey-io/valkey-go"
	"github.com/valkey-io/valkey-go/valkeyaside"

	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
)

const defaultValkeyConnectTimeout = 5 * time.Second

// NewClient returns a connected valkey-go client built from the application
// config. It pings the server before returning and closes the client on ping
// failure. The caller is responsible for calling Close on the returned client.
func NewClient(ctx context.Context, cfg *config.Config) (valkeygo.Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	client, err := valkeygo.NewClient(valkeyClientOption(cfg))
	if err != nil {
		return nil, fmt.Errorf("create valkey client for %s: %w", cfg.ValkeyAddr(), err)
	}

	if err := ping(ctx, cfg, client.B().Ping().Build(), client.Do); err != nil {
		client.Close()
		return nil, err
	}

	return client, nil
}

// valkeyClientOption derives the valkey-go client options from the app config,
// injecting the configured connect timeout as the dialer timeout.
func valkeyClientOption(cfg *config.Config) valkeygo.ClientOption {
	return valkeygo.ClientOption{
		InitAddress: []string{cfg.ValkeyAddr()},
		Password:    cfg.Valkey.Password,
		SelectDB:    cfg.Valkey.DB,
		Dialer:      net.Dialer{Timeout: effectiveConnectTimeout(cfg)},
	}
}

// effectiveConnectTimeout returns the configured Valkey connect timeout, or the
// package default when it is unset (zero or negative).
func effectiveConnectTimeout(cfg *config.Config) time.Duration {
	if cfg.Valkey.ConnectTimeout > 0 {
		return cfg.Valkey.ConnectTimeout
	}
	return defaultValkeyConnectTimeout
}

// ping sends a PING bounded by the connect timeout, returning a wrapped error
// naming the address on failure.
func ping(ctx context.Context, cfg *config.Config, cmd valkeygo.Completed, do func(context.Context, valkeygo.Completed) valkeygo.ValkeyResult) error {
	pingCtx, cancel := context.WithTimeout(ctx, effectiveConnectTimeout(cfg))
	defer cancel()
	if err := do(pingCtx, cmd).Error(); err != nil {
		return fmt.Errorf("ping valkey %s: %w", cfg.ValkeyAddr(), err)
	}
	return nil
}

// NewCacheAside returns a cache-aside client backed by Valkey's client-side
// caching. It provides stampede-safe Get(ttl, key, loader): concurrent misses
// for the same key run the loader once while the rest wait on the invalidation
// notification, and cached values are invalidated by the server on write.
//
// The returned client owns its own connection (it enables client tracking) and
// must be Closed; the DI container registers a closer for it.
func NewCacheAside(ctx context.Context, cfg *config.Config) (valkeyaside.CacheAsideClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	client, err := valkeyaside.NewClient(valkeyaside.ClientOption{
		ClientOption: valkeyClientOption(cfg),
	})
	if err != nil {
		return nil, fmt.Errorf("create cache-aside client for %s: %w", cfg.ValkeyAddr(), err)
	}

	if err := ping(ctx, cfg, client.Client().B().Ping().Build(), client.Client().Do); err != nil {
		client.Close()
		return nil, err
	}

	return client, nil
}
