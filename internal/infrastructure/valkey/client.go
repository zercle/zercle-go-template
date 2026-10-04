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

// valkeyClientOption derives the valkey-go client options from the app config,
// injecting the configured connect timeout as the dialer timeout.
func valkeyClientOption(cfg *config.Config) valkeygo.ClientOption {
	connectTimeout := cfg.Valkey.ConnectTimeout
	if connectTimeout <= 0 {
		connectTimeout = defaultValkeyConnectTimeout
	}
	return valkeygo.ClientOption{
		InitAddress: []string{cfg.ValkeyAddr()},
		Password:    cfg.Valkey.Password,
		SelectDB:    cfg.Valkey.DB,
		Dialer:      net.Dialer{Timeout: connectTimeout},
	}
}

// NewCacheAside returns a cache-aside client backed by Valkey's client-side
// caching. It provides stampede-safe Get(ttl, key, loader): concurrent misses
// for the same key run the loader once while the rest wait on the invalidation
// notification, and cached values are invalidated by the server on write.
//
// The returned client owns its own connection (it enables client tracking) and
// must be Closed; the DI container registers a closer for it. With no Valkey
// configured the caller should skip building it.
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

	connectTimeout := cfg.Valkey.ConnectTimeout
	if connectTimeout <= 0 {
		connectTimeout = defaultValkeyConnectTimeout
	}
	pingCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()
	if err := client.Client().Do(pingCtx, client.Client().B().Ping().Build()).Error(); err != nil {
		client.Close()
		return nil, fmt.Errorf("ping valkey %s: %w", cfg.ValkeyAddr(), err)
	}

	return client, nil
}
