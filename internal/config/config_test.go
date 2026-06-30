//go:build unit

package config_test

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/config"
)

func TestLoad_ReadsConfigFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	content := `
app:
  name: test-app
  environment: test
  host: 127.0.0.1
  port: 7000
  shutdown_timeout: 5s
http:
  host: 127.0.0.1
  port: 7001
  read_timeout: 10s
  write_timeout: 10s
  idle_timeout: 30s
  body_limit: 2M
grpc:
  host: 127.0.0.1
  port: 7002
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

	require.Equal(t, "test-app", cfg.App.Name)
	require.Equal(t, "test", cfg.App.Environment)
	require.Equal(t, "127.0.0.1", cfg.HTTP.Host)
	require.Equal(t, 7001, cfg.HTTP.Port)
	require.Equal(t, "127.0.0.1:7001", cfg.HTTPAddr())
	require.Equal(t, "testdb", cfg.DB.Name)
	require.Equal(t, int32(5), cfg.DB.MaxConns)
	require.Equal(t, "127.0.0.1:6379", cfg.ValkeyAddr())
	require.False(t, cfg.Example.Enabled)
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

	t.Setenv("APP_NAME", "env-app")
	t.Setenv("HTTP_PORT", "2222")
	t.Setenv("DB_NAME", "envdb")
	t.Setenv("OTEL_SERVICE_NAME", "env-service")
	t.Setenv("EXAMPLE_ENABLED", "false")

	cfg, err := config.Load()
	require.NoError(t, err)
	require.NoError(t, cfg.Validate())

	require.Equal(t, "env-app", cfg.App.Name)
	require.Equal(t, 2222, cfg.HTTP.Port)
	require.Equal(t, "envdb", cfg.DB.Name)
	require.Equal(t, "env-service", cfg.OTel.ServiceName)
	require.False(t, cfg.Example.Enabled)
}

func TestLoad_SliceEnvVariable(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	content := `
app:
  environment: test
  port: 8080
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

func TestLoad_ExampleDefaults(t *testing.T) {
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

	require.Equal(t, int32(20), cfg.Example.DefaultPageSize)
	require.Equal(t, int32(100), cfg.Example.MaxPageSize)
	require.Equal(t, int32(255), cfg.Example.MaxNameLength)
	require.Equal(t, 5*time.Second, cfg.HTTP.HealthProbeTimeout)
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

func TestValidate_ExampleDefaultPageSizeExceedsMax(t *testing.T) {
	cfg := validConfig()
	cfg.Example.Enabled = true
	cfg.Example.DefaultPageSize = 100
	cfg.Example.MaxPageSize = 10

	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "EXAMPLE_DEFAULT_PAGE_SIZE must be <= EXAMPLE_MAX_PAGE_SIZE")
}

func TestValidate_ExampleMaxPageSizeExceedsUpperBound(t *testing.T) {
	cfg := validConfig()
	cfg.Example.Enabled = true
	cfg.Example.MaxPageSize = 100000

	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "EXAMPLE_MAX_PAGE_SIZE exceeds")
}

func TestValidate_ExampleDisabledSkipsChecks(t *testing.T) {
	cfg := validConfig()
	cfg.Example.Enabled = false
	cfg.Example.DefaultPageSize = 1000
	cfg.Example.MaxPageSize = 10

	require.NoError(t, cfg.Validate())
}

// TestValidate_ExampleDisabledAllowsZeroValues verifies that when the example
// feature is disabled, zero-valued ExampleConfig fields do not fail validation.
// This lets users delete the example: block from config.yaml without startup
// failing on required,min=1 tags.
func TestValidate_ExampleDisabledAllowsZeroValues(t *testing.T) {
	t.Parallel()

	cfg := validConfig()
	cfg.Example.Enabled = false
	cfg.Example.DefaultPageSize = 0
	cfg.Example.MaxPageSize = 0
	cfg.Example.MaxNameLength = 0

	require.NoError(t, cfg.Validate())
}

// TestValidate_ExampleEnabledRejectsZeroValues verifies that explicit validation
// rejects zero-valued ExampleConfig fields when the feature is enabled.
func TestValidate_ExampleEnabledRejectsZeroValues(t *testing.T) {
	t.Parallel()

	cfg := validConfig()
	cfg.Example.Enabled = true
	cfg.Example.DefaultPageSize = 0

	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "EXAMPLE_DEFAULT_PAGE_SIZE must be >= 1")
}

// TestValidate_ExampleEnabledRejectsNegativeValues verifies that explicit validation
// rejects negative-valued ExampleConfig fields when the feature is enabled,
// confirming the min=1 positivity guarantee.
func TestValidate_ExampleEnabledRejectsNegativeValues(t *testing.T) {
	t.Parallel()

	cfg := validConfig()
	cfg.Example.Enabled = true
	cfg.Example.DefaultPageSize = -1

	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "EXAMPLE_DEFAULT_PAGE_SIZE must be >= 1")
}

func TestValidate_AcceptsValidConfig(t *testing.T) {
	cfg := validConfig()
	require.NoError(t, cfg.Validate())
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

func TestGRPCAddr(t *testing.T) {
	cfg := validConfig()
	cfg.GRPC.Host = "127.0.0.1"
	cfg.GRPC.Port = 50051
	require.Equal(t, "127.0.0.1:50051", cfg.GRPCAddr())
}

func validConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Name:            "test",
			Environment:     "test",
			Host:            "127.0.0.1",
			Port:            8080,
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
		GRPC: config.GRPCConfig{
			Host: "127.0.0.1",
			Port: 50051,
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
		Example: config.ExampleConfig{
			Enabled:         true,
			DefaultPageSize: 20,
			MaxPageSize:     100,
			MaxNameLength:   255,
		},
	}
}
