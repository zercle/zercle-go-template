//go:build unit

package di_test

import (
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"
	"github.com/valkey-io/valkey-go/valkeyaside"

	"github.com/zercle/zercle-go-template/internal/features/catalog/di"
	"github.com/zercle/zercle-go-template/internal/features/catalog/usecase"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/testutil"
)

// TestRegister_DepsMissing returns an error when required DI dependencies are
// not registered while the feature is enabled.
func TestRegister_DepsMissing(t *testing.T) {
	t.Parallel()

	injector := do.New()
	do.ProvideValue(injector, &config.Config{Catalog: config.CatalogConfig{Enabled: true}})

	err := di.Register(injector)
	require.Error(t, err)
}

// TestRegister_DisabledSkipsFeature pins the feature-flag contract: with
// Enabled=false the feature registers nothing, resolves nothing, and mounts no
// routes.
func TestRegister_DisabledSkipsFeature(t *testing.T) {
	t.Parallel()

	injector := do.New()
	do.ProvideValue(injector, &config.Config{Catalog: config.CatalogConfig{Enabled: false}})

	require.NoError(t, di.Register(injector))

	_, err := do.Invoke[usecase.Service](injector)
	require.Error(t, err, "disabled feature must not provide its service")
}

// TestRegister_WiresWithoutInfra pins the happy path: with only config, a
// sqlmock-backed *gorm.DB, and an echo instance registered, Register wires the
// whole feature (no live Postgres or Valkey). The service resolves and the
// routes mount.
func TestRegister_WiresWithoutInfra(t *testing.T) {
	t.Parallel()

	gormDB, _ := testutil.NewSQLMockDB(t)

	injector := do.New()
	do.ProvideValue(injector, &config.Config{
		Valkey: config.ValkeyConfig{TTL: time.Minute},
		Catalog: config.CatalogConfig{
			Enabled:         true,
			DefaultPageSize: 1,
			MaxPageSize:     2,
			MaxNameLength:   5,
		},
	})
	do.ProvideValue(injector, gormDB)
	do.ProvideValue(injector, echo.New())

	require.NoError(t, di.Register(injector))

	_, err := do.Invoke[usecase.Service](injector)
	require.NoError(t, err, "enabled feature must provide its usecase service")

	e, err := do.Invoke[*echo.Echo](injector)
	require.NoError(t, err)

	var found bool
	for _, r := range e.Router().Routes() {
		if r.Method == "POST" && r.Path == "/api/v1/products" {
			found = true
			break
		}
	}
	require.True(t, found, "feature routes must be mounted at /api/v1")
}

// TestRegister_WiresWithCacheAside pins the cache-aside branch: when a
// valkeyaside.CacheAsideClient is registered, Register decorates the
// repository with it and the feature still wires without live infrastructure.
func TestRegister_WiresWithCacheAside(t *testing.T) {
	t.Parallel()

	gormDB, _ := testutil.NewSQLMockDB(t)

	injector := do.New()
	do.ProvideValue(injector, &config.Config{
		Valkey: config.ValkeyConfig{TTL: time.Minute},
		Catalog: config.CatalogConfig{
			Enabled:         true,
			DefaultPageSize: 1,
			MaxPageSize:     2,
			MaxNameLength:   5,
		},
	})
	do.ProvideValue(injector, gormDB)
	do.ProvideValue(injector, echo.New())
	do.ProvideValue[valkeyaside.CacheAsideClient](injector, &testutil.FakeCacheAside{})

	require.NoError(t, di.Register(injector))

	_, err := do.Invoke[usecase.Service](injector)
	require.NoError(t, err, "cache-aside branch must still provide the service")
}
