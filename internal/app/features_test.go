//go:build unit

package app

import (
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"
	"github.com/valkey-io/valkey-go/valkeyaside"

	catalogusecase "github.com/zercle/zercle-go-template/internal/features/catalog/usecase"
	machinesusecase "github.com/zercle/zercle-go-template/internal/features/machines/usecase"
	reportingusecase "github.com/zercle/zercle-go-template/internal/features/reporting/usecase"
	salesusecase "github.com/zercle/zercle-go-template/internal/features/sales/usecase"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/testutil"
)

// findRoute reports whether a route with the given method and path is mounted.
func findRoute(e *echo.Echo, method, path string) bool {
	for _, r := range e.Router().Routes() {
		if r.Method == method && r.Path == path {
			return true
		}
	}
	return false
}

// TestRegisterCatalog_DepsMissing returns an error when required DI dependencies
// are not registered while the feature is enabled.
func TestRegisterCatalog_DepsMissing(t *testing.T) {
	t.Parallel()

	injector := do.New()
	do.ProvideValue(injector, &config.Config{Catalog: config.CatalogConfig{Enabled: true}})

	require.Error(t, registerCatalog(injector))
}

// TestRegisterCatalog_DisabledSkipsFeature pins the feature-flag contract: with
// Enabled=false the feature registers nothing, resolves nothing, and mounts no
// routes.
func TestRegisterCatalog_DisabledSkipsFeature(t *testing.T) {
	t.Parallel()

	injector := do.New()
	do.ProvideValue(injector, &config.Config{Catalog: config.CatalogConfig{Enabled: false}})

	require.NoError(t, registerCatalog(injector))

	_, err := do.Invoke[catalogusecase.Service](injector)
	require.Error(t, err, "disabled feature must not provide its service")
}

// TestRegisterCatalog_WiresWithoutInfra pins the happy path: with only config,
// a sqlmock-backed *gorm.DB, and an echo instance registered, registerCatalog
// wires the whole feature (no live Postgres or Valkey). The service resolves
// and the routes mount.
func TestRegisterCatalog_WiresWithoutInfra(t *testing.T) {
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

	require.NoError(t, registerCatalog(injector))

	_, err := do.Invoke[catalogusecase.Service](injector)
	require.NoError(t, err, "enabled feature must provide its usecase service")

	e, err := do.Invoke[*echo.Echo](injector)
	require.NoError(t, err)
	require.True(t, findRoute(e, "POST", "/api/v1/products"), "feature routes must be mounted at /api/v1")
}

// TestRegisterCatalog_WiresWithCacheAside pins the cache-aside branch: when a
// valkeyaside.CacheAsideClient is registered, registerCatalog decorates the
// repository with it and the feature still wires without live infrastructure.
func TestRegisterCatalog_WiresWithCacheAside(t *testing.T) {
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

	require.NoError(t, registerCatalog(injector))

	_, err := do.Invoke[catalogusecase.Service](injector)
	require.NoError(t, err, "cache-aside branch must still provide the service")
}

// TestRegisterMachines_DepsMissing returns an error when required DI
// dependencies are not registered while the feature is enabled.
func TestRegisterMachines_DepsMissing(t *testing.T) {
	t.Parallel()

	injector := do.New()
	do.ProvideValue(injector, &config.Config{Machines: config.MachinesConfig{Enabled: true}})

	require.Error(t, registerMachines(injector))
}

// TestRegisterMachines_DisabledSkipsFeature pins the feature-flag contract:
// with Enabled=false the feature registers nothing and mounts no routes.
func TestRegisterMachines_DisabledSkipsFeature(t *testing.T) {
	t.Parallel()

	injector := do.New()
	do.ProvideValue(injector, &config.Config{Machines: config.MachinesConfig{Enabled: false}})

	require.NoError(t, registerMachines(injector))

	_, err := do.Invoke[machinesusecase.Service](injector)
	require.Error(t, err, "disabled feature must not provide its service")
}

// TestRegisterMachines_WiresWithoutInfra pins the happy path: the service
// resolves and the routes mount without live Postgres.
func TestRegisterMachines_WiresWithoutInfra(t *testing.T) {
	t.Parallel()

	gormDB, _ := testutil.NewSQLMockDB(t)

	injector := do.New()
	do.ProvideValue(injector, &config.Config{
		Machines: config.MachinesConfig{
			Enabled:         true,
			DefaultPageSize: 1,
			MaxPageSize:     2,
			MaxLabelLength:  5,
		},
	})
	do.ProvideValue(injector, gormDB)
	do.ProvideValue(injector, echo.New())

	require.NoError(t, registerMachines(injector))

	_, err := do.Invoke[machinesusecase.Service](injector)
	require.NoError(t, err, "enabled feature must provide its usecase service")

	e, err := do.Invoke[*echo.Echo](injector)
	require.NoError(t, err)
	require.True(t, findRoute(e, "POST", "/api/v1/machines"), "feature routes must be mounted at /api/v1")
}

// TestRegisterSales_DepsMissing returns an error when required DI dependencies
// are not registered while the feature is enabled.
func TestRegisterSales_DepsMissing(t *testing.T) {
	t.Parallel()

	injector := do.New()
	do.ProvideValue(injector, &config.Config{Sales: config.SalesConfig{Enabled: true}})

	require.Error(t, registerSales(injector))
}

// TestRegisterSales_DisabledSkipsFeature pins the feature-flag contract: with
// Enabled=false the feature registers nothing and mounts no routes.
func TestRegisterSales_DisabledSkipsFeature(t *testing.T) {
	t.Parallel()

	injector := do.New()
	do.ProvideValue(injector, &config.Config{Sales: config.SalesConfig{Enabled: false}})

	require.NoError(t, registerSales(injector))

	_, err := do.Invoke[salesusecase.Service](injector)
	require.Error(t, err, "disabled feature must not provide its service")
}

// TestRegisterSales_WiresWithoutInfra pins the happy path: the service
// resolves and the route mounts without live Postgres.
func TestRegisterSales_WiresWithoutInfra(t *testing.T) {
	t.Parallel()

	gormDB, _ := testutil.NewSQLMockDB(t)

	injector := do.New()
	do.ProvideValue(injector, &config.Config{Sales: config.SalesConfig{Enabled: true}})
	do.ProvideValue(injector, gormDB)
	do.ProvideValue(injector, echo.New())

	require.NoError(t, registerSales(injector))

	_, err := do.Invoke[salesusecase.Service](injector)
	require.NoError(t, err, "enabled feature must provide its usecase service")

	e, err := do.Invoke[*echo.Echo](injector)
	require.NoError(t, err)
	require.True(t, findRoute(e, "POST", "/api/v1/purchases"), "feature route must be mounted at /api/v1")
}

// TestRegisterReporting_DepsMissing returns an error when required DI
// dependencies are not registered while the feature is enabled.
func TestRegisterReporting_DepsMissing(t *testing.T) {
	t.Parallel()

	injector := do.New()
	do.ProvideValue(injector, &config.Config{Reporting: config.ReportingConfig{Enabled: true}})

	require.Error(t, registerReporting(injector))
}

// TestRegisterReporting_DisabledSkipsFeature pins the feature-flag contract:
// with Enabled=false the feature registers nothing and mounts no routes.
func TestRegisterReporting_DisabledSkipsFeature(t *testing.T) {
	t.Parallel()

	injector := do.New()
	do.ProvideValue(injector, &config.Config{Reporting: config.ReportingConfig{Enabled: false}})

	require.NoError(t, registerReporting(injector))

	_, err := do.Invoke[reportingusecase.Service](injector)
	require.Error(t, err, "disabled feature must not provide its service")
}

// TestRegisterReporting_WiresWithoutInfra pins the happy path: the service
// resolves and the route mounts without live Postgres.
func TestRegisterReporting_WiresWithoutInfra(t *testing.T) {
	t.Parallel()

	gormDB, _ := testutil.NewSQLMockDB(t)

	injector := do.New()
	do.ProvideValue(injector, &config.Config{
		Reporting: config.ReportingConfig{Enabled: true, DefaultTopMachines: 5, MaxTopMachines: 20},
	})
	do.ProvideValue(injector, gormDB)
	do.ProvideValue(injector, echo.New())

	require.NoError(t, registerReporting(injector))

	_, err := do.Invoke[reportingusecase.Service](injector)
	require.NoError(t, err, "enabled feature must provide its usecase service")

	e, err := do.Invoke[*echo.Echo](injector)
	require.NoError(t, err)
	require.True(t, findRoute(e, "GET", "/api/v1/reports/summary"), "feature route must be mounted at /api/v1")
}
