//go:build unit

package config_test

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
)

func TestLoad_ExplicitConfigFileMissingFails(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "does-not-exist.yaml")

	t.Setenv("CONFIG_FILE", missing)

	cfg, err := config.Load()
	require.Error(t, err)
	require.Nil(t, cfg)
	require.Contains(t, err.Error(), "read config")
}

func TestLoad_ReadsConfigFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	content := `
app:
  environment: test
  shutdown_timeout: 5s
http:
  host: 127.0.0.1
  port: 7001
  read_timeout: 10s
  write_timeout: 10s
  idle_timeout: 30s
  body_limit: 2M
db:
  host: 127.0.0.1
  port: 5432
  name: testdb
  user: testuser
  password: testpass
  ssl_mode: disable
  max_conns: 5
  max_idle_conns: 1
  max_conn_idle: 10m
  max_conn_life: 20m
  connect_timeout: 3s
valkey:
  host: 127.0.0.1
  port: 6379
  password: ""
  db: 0
log:
  level: debug
  format: console
otel:
  exporter: none
  service_name: test-service
  sampling: 0.5
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0o600))

	t.Setenv("CONFIG_FILE", cfgPath)

	cfg, err := config.Load()
	require.NoError(t, err)
	require.NoError(t, cfg.Validate())

	require.Equal(t, "test", cfg.App.Environment)
	require.Equal(t, "127.0.0.1", cfg.HTTP.Host)
	require.Equal(t, 7001, cfg.HTTP.Port)
	require.Equal(t, "127.0.0.1:7001", cfg.HTTPAddr())
	require.Equal(t, "testdb", cfg.DB.Name)
	require.Equal(t, int32(5), cfg.DB.MaxConns)
	require.Equal(t, "127.0.0.1:6379", cfg.ValkeyAddr())
}

func TestLoad_OverridesFromEnv(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	content := `
app:
  environment: test
http:
  port: 1111
db:
  host: 127.0.0.1
  port: 5432
  name: db
  user: u
  password: p
valkey:
  host: 127.0.0.1
  port: 6379
log:
  level: info
  format: json
otel:
  exporter: none
  service_name: svc
  sampling: 1.0
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0o600))
	t.Setenv("CONFIG_FILE", cfgPath)

	t.Setenv("APP_ENVIRONMENT", "production")
	t.Setenv("HTTP_PORT", "2222")
	t.Setenv("DB_NAME", "envdb")
	t.Setenv("OTEL_SERVICE_NAME", "env-service")

	cfg, err := config.Load()
	require.NoError(t, err)
	require.NoError(t, cfg.Validate())

	require.Equal(t, "production", cfg.App.Environment)
	require.Equal(t, 2222, cfg.HTTP.Port)
	require.Equal(t, "envdb", cfg.DB.Name)
	require.Equal(t, "env-service", cfg.OTel.ServiceName)
}

func TestLoad_SliceEnvVariable(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	content := `
app:
  environment: test
http:
  port: 8080
  read_timeout: 10s
  write_timeout: 10s
  idle_timeout: 30s
  body_limit: 1M
db:
  host: 127.0.0.1
  port: 5432
  name: db
  user: u
  password: p
valkey:
  host: 127.0.0.1
  port: 6379
log:
  level: info
  format: json
otel:
  exporter: none
  service_name: svc
  sampling: 1.0
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0o600))
	t.Setenv("CONFIG_FILE", cfgPath)

	t.Setenv("HTTP_CORS_ALLOW_ORIGINS", "https://example.com,https://app.example.com")
	t.Setenv("HTTP_CORS_ALLOW_METHODS", "GET,POST")
	t.Setenv("HTTP_CORS_ALLOW_HEADERS", "X-Custom")

	cfg, err := config.Load()
	require.NoError(t, err)
	require.NoError(t, cfg.Validate())

	require.Equal(t, []string{"https://example.com", "https://app.example.com"}, cfg.HTTP.CORSAllowOrigins)
	require.Equal(t, []string{"GET", "POST"}, cfg.HTTP.CORSAllowMethods)
	require.Equal(t, []string{"X-Custom"}, cfg.HTTP.CORSAllowHeaders)
}

func TestValidate_InvalidEnvironment(t *testing.T) {
	cfg := validConfig()
	cfg.App.Environment = "invalid"

	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "config validation failed")
}

func TestValidate_MaxConnsBelowMaxIdleConns(t *testing.T) {
	cfg := validConfig()
	cfg.DB.MaxConns = 1
	cfg.DB.MaxIdleConns = 5

	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "DB_MAX_CONNS must be >= DB_MAX_IDLE_CONNS")
}

func TestValidate_OTLPWithoutEndpoint(t *testing.T) {
	cfg := validConfig()
	cfg.OTel.Exporter = "otlp"
	cfg.OTel.Endpoint = ""

	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "OTEL_EXPORTER_OTLP_ENDPOINT is required when OTEL_EXPORTER=otlp")
}

func TestValidate_InvalidOTLPURL(t *testing.T) {
	cfg := validConfig()
	cfg.OTel.Exporter = "otlp"
	cfg.OTel.Endpoint = "://missing-scheme"

	err := cfg.Validate()
	require.Error(t, err)
}

func TestValidate_AcceptsValidConfig(t *testing.T) {
	cfg := validConfig()
	require.NoError(t, cfg.Validate())
}

func TestValidate_InvalidBodyLimitFails(t *testing.T) {
	cfg := validConfig()
	cfg.HTTP.BodyLimit = "1XB"

	err := cfg.Validate()

	require.Error(t, err, "an unparseable body limit must fail startup, not silently disable the middleware")
	require.Contains(t, err.Error(), "HTTP_BODY_LIMIT")
}

func TestValidate_OTLPEndpointRequiresSchemeAndHost(t *testing.T) {
	cfg := validConfig()
	cfg.OTel.Exporter = "otlp"
	cfg.OTel.Endpoint = "not-a-url"

	err := cfg.Validate()

	require.Error(t, err)
	require.Contains(t, err.Error(), "OTEL_EXPORTER_OTLP_ENDPOINT")
}

func TestDBConnString(t *testing.T) {
	cfg := validConfig()
	cfg.DB.Password = "p@ss w#rd"

	dsn := cfg.DBConnString()

	parsed, err := url.Parse(dsn)
	require.NoError(t, err)
	require.Equal(t, "postgres", parsed.User.Username())
	password, hasPassword := parsed.User.Password()
	require.True(t, hasPassword)
	require.Equal(t, "p@ss w#rd", password)
	require.Equal(t, "disable", parsed.Query().Get("sslmode"))
}

// TestLoad_FeatureDefaults pins the new feature sections' defaults: disabled
// with the same page-size/name-length defaults as the example feature.
func TestLoad_FeatureDefaults(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	content := `
app:
  environment: test
http:
  port: 8080
db:
  host: 127.0.0.1
  port: 5432
  name: db
  user: u
  password: p
valkey:
  host: 127.0.0.1
  port: 6379
log:
  level: info
  format: json
otel:
  exporter: none
  service_name: svc
  sampling: 1.0
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0o600))
	t.Setenv("CONFIG_FILE", cfgPath)

	cfg, err := config.Load()
	require.NoError(t, err)
	require.NoError(t, cfg.Validate())

	require.False(t, cfg.Catalog.Enabled)
	require.Equal(t, int32(20), cfg.Catalog.DefaultPageSize)
	require.Equal(t, int32(100), cfg.Catalog.MaxPageSize)
	require.Equal(t, int32(255), cfg.Catalog.MaxNameLength)

	require.False(t, cfg.Machines.Enabled)
	require.Equal(t, int32(20), cfg.Machines.DefaultPageSize)
	require.Equal(t, int32(100), cfg.Machines.MaxPageSize)
	require.Equal(t, int32(255), cfg.Machines.MaxLabelLength)

	require.False(t, cfg.Sales.Enabled)
}

// TestLoad_FeatureEnvOverrides proves every new leaf is explicit-bound: setting
// the env vars overrides both the file and the defaults.
func TestLoad_FeatureEnvOverrides(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	content := `
app:
  environment: test
http:
  port: 8080
db:
  host: 127.0.0.1
  port: 5432
  name: db
  user: u
  password: p
valkey:
  host: 127.0.0.1
  port: 6379
log:
  level: info
  format: json
otel:
  exporter: none
  service_name: svc
  sampling: 1.0
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0o600))
	t.Setenv("CONFIG_FILE", cfgPath)

	t.Setenv("CATALOG_ENABLED", "true")
	t.Setenv("CATALOG_DEFAULT_PAGE_SIZE", "30")
	t.Setenv("CATALOG_MAX_PAGE_SIZE", "300")
	t.Setenv("CATALOG_MAX_NAME_LENGTH", "512")
	t.Setenv("MACHINES_ENABLED", "true")
	t.Setenv("MACHINES_DEFAULT_PAGE_SIZE", "40")
	t.Setenv("MACHINES_MAX_PAGE_SIZE", "400")
	t.Setenv("MACHINES_MAX_LABEL_LENGTH", "128")
	t.Setenv("SALES_ENABLED", "true")

	cfg, err := config.Load()
	require.NoError(t, err)
	require.NoError(t, cfg.Validate())

	require.True(t, cfg.Catalog.Enabled)
	require.Equal(t, int32(30), cfg.Catalog.DefaultPageSize)
	require.Equal(t, int32(300), cfg.Catalog.MaxPageSize)
	require.Equal(t, int32(512), cfg.Catalog.MaxNameLength)

	require.True(t, cfg.Machines.Enabled)
	require.Equal(t, int32(40), cfg.Machines.DefaultPageSize)
	require.Equal(t, int32(400), cfg.Machines.MaxPageSize)
	require.Equal(t, int32(128), cfg.Machines.MaxLabelLength)

	require.True(t, cfg.Sales.Enabled)
}

func TestValidate_CatalogDefaultPageSizeExceedsMax(t *testing.T) {
	t.Parallel()

	cfg := validConfig()
	cfg.Catalog.Enabled = true
	cfg.Catalog.DefaultPageSize = 100
	cfg.Catalog.MaxPageSize = 10

	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "CATALOG_DEFAULT_PAGE_SIZE must be <= CATALOG_MAX_PAGE_SIZE")
}

func TestValidate_CatalogMaxPageSizeExceedsUpperBound(t *testing.T) {
	t.Parallel()

	cfg := validConfig()
	cfg.Catalog.Enabled = true
	cfg.Catalog.MaxPageSize = 100000

	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "CATALOG_MAX_PAGE_SIZE exceeds")
}

func TestValidate_CatalogMaxNameLengthExceedsUpperBound(t *testing.T) {
	t.Parallel()

	cfg := validConfig()
	cfg.Catalog.Enabled = true
	cfg.Catalog.MaxNameLength = 100000

	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "CATALOG_MAX_NAME_LENGTH exceeds")
}

// TestValidate_CatalogEnabledRejectsZeroValues pins the positivity guarantee
// when the feature is on.
func TestValidate_CatalogEnabledRejectsZeroValues(t *testing.T) {
	t.Parallel()

	cfg := validConfig()
	cfg.Catalog.Enabled = true
	cfg.Catalog.DefaultPageSize = 0

	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "CATALOG_DEFAULT_PAGE_SIZE must be >= 1")
}

// TestValidate_CatalogDisabledSkipsChecks lets the catalog block be omitted
// from config.yaml without startup failing on zero values.
func TestValidate_CatalogDisabledSkipsChecks(t *testing.T) {
	t.Parallel()

	cfg := validConfig()
	cfg.Catalog.Enabled = false
	cfg.Catalog.DefaultPageSize = 0
	cfg.Catalog.MaxPageSize = 0
	cfg.Catalog.MaxNameLength = 0

	require.NoError(t, cfg.Validate())
}

func TestValidate_MachinesDefaultPageSizeExceedsMax(t *testing.T) {
	t.Parallel()

	cfg := validConfig()
	cfg.Machines.Enabled = true
	cfg.Machines.DefaultPageSize = 100
	cfg.Machines.MaxPageSize = 10

	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "MACHINES_DEFAULT_PAGE_SIZE must be <= MACHINES_MAX_PAGE_SIZE")
}

func TestValidate_MachinesMaxLabelLengthExceedsUpperBound(t *testing.T) {
	t.Parallel()

	cfg := validConfig()
	cfg.Machines.Enabled = true
	cfg.Machines.MaxLabelLength = 100000

	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "MACHINES_MAX_LABEL_LENGTH exceeds")
}

// TestValidate_MachinesDisabledSkipsChecks mirrors the catalog disabled case.
func TestValidate_MachinesDisabledSkipsChecks(t *testing.T) {
	t.Parallel()

	cfg := validConfig()
	cfg.Machines.Enabled = false
	cfg.Machines.DefaultPageSize = 0
	cfg.Machines.MaxPageSize = 0
	cfg.Machines.MaxLabelLength = 0

	require.NoError(t, cfg.Validate())
}

// TestValidate_SalesEnabledAddsNoFieldConstraints confirms the sales toggle
// carries no numeric fields to validate: enabling it alone stays valid.
func TestValidate_SalesEnabledAddsNoFieldConstraints(t *testing.T) {
	t.Parallel()

	cfg := validConfig()
	cfg.Sales.Enabled = true

	require.NoError(t, cfg.Validate())
}

func validConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Environment:     "test",
			ShutdownTimeout: 15 * time.Second,
		},
		HTTP: config.HTTPConfig{
			Host:               "127.0.0.1",
			Port:               8080,
			ReadTimeout:        15 * time.Second,
			WriteTimeout:       15 * time.Second,
			IdleTimeout:        60 * time.Second,
			BodyLimit:          "1M",
			HealthProbeTimeout: 5 * time.Second,
		},
		DB: config.DBConfig{
			Host:           "127.0.0.1",
			Port:           5432,
			Name:           "app",
			User:           "postgres",
			Password:       "postgres",
			SSLMode:        "disable",
			MaxConns:       10,
			MaxIdleConns:   2,
			MaxConnIdle:    30 * time.Minute,
			MaxConnLife:    1 * time.Hour,
			ConnectTimeout: 5 * time.Second,
		},
		Valkey: config.ValkeyConfig{
			Host: "127.0.0.1",
			Port: 6379,
			DB:   0,
		},
		OTel: config.OTelConfig{
			Exporter:    "none",
			Endpoint:    "http://localhost:4318",
			ServiceName: "test-service",
			Sampling:    1.0,
		},
		Log: config.LogConfig{
			Level:  "info",
			Format: "json",
		},
		Catalog: config.CatalogConfig{
			Enabled:         true,
			DefaultPageSize: 20,
			MaxPageSize:     100,
			MaxNameLength:   255,
		},
		Machines: config.MachinesConfig{
			Enabled:         true,
			DefaultPageSize: 20,
			MaxPageSize:     100,
			MaxLabelLength:  255,
		},
		Sales: config.SalesConfig{
			Enabled: true,
		},
	}
}
