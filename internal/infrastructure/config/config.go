// Package config loads application configuration from config.yaml and
// environment variables. Validation is performed by Config.Validate.
package config

import (
	"errors"
	"fmt"
	"math"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// leafBinding describes a configuration leaf that is explicitly bound to an
// environment variable name.
type leafBinding struct {
	key     string
	envName string
}

// Config is the single source of truth for application configuration.
//
// Fields carry mapstructure tags (viper decoding) and validate tags only: the
// config is never marshalled to YAML, only unmarshalled from it, so yaml tags
// would be dead weight that drifts from the mapstructure names. Environment
// binding is likewise explicit — leafBindings lists every env-var-to-key pair,
// so a new field is env-readable only once it is added there.
type Config struct {
	App     AppConfig     `mapstructure:"app"`
	HTTP    HTTPConfig    `mapstructure:"http"`
	DB      DBConfig      `mapstructure:"db"`
	Valkey  ValkeyConfig  `mapstructure:"valkey"`
	OTel    OTelConfig    `mapstructure:"otel"`
	Log     LogConfig     `mapstructure:"log"`
	Example ExampleConfig `mapstructure:"example"`
}

// AppConfig holds process-level settings. Service identity (name, listen host,
// port) lives on HTTPConfig and OTelConfig; nothing reads a separate app-level
// copy, so those fields are not declared here.
type AppConfig struct {
	Environment     string        `mapstructure:"environment" validate:"oneof=development staging production test"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" validate:"required,min=1s"`
}

// HTTPConfig holds the HTTP server settings and CORS options.
type HTTPConfig struct {
	Host               string        `mapstructure:"host" validate:"ip|hostname"`
	Port               int           `mapstructure:"port" validate:"required,min=1,max=65535"`
	ReadTimeout        time.Duration `mapstructure:"read_timeout" validate:"required,min=1s"`
	WriteTimeout       time.Duration `mapstructure:"write_timeout" validate:"required,min=1s"`
	IdleTimeout        time.Duration `mapstructure:"idle_timeout" validate:"required,min=1s"`
	BodyLimit          string        `mapstructure:"body_limit" validate:"required"`
	HealthProbeTimeout time.Duration `mapstructure:"health_probe_timeout" validate:"required,min=1s"`
	CORSAllowOrigins   []string      `mapstructure:"cors_allow_origins"`
	CORSAllowMethods   []string      `mapstructure:"cors_allow_methods"`
	CORSAllowHeaders   []string      `mapstructure:"cors_allow_headers"`
}

// DBConfig holds the PostgreSQL connection and pool settings.
type DBConfig struct {
	Host     string `mapstructure:"host" validate:"required,hostname|ip"`
	Port     int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	Name     string `mapstructure:"name" validate:"required"`
	User     string `mapstructure:"user" validate:"required"`
	Password string `mapstructure:"password" validate:"required"`
	SSLMode  string `mapstructure:"ssl_mode" validate:"oneof=disable prefer require verify-ca verify-full"`
	MaxConns int32  `mapstructure:"max_conns" validate:"required,min=1"`
	// MaxIdleConns is the maximum number of idle connections retained in the
	// pool. Maps to database/sql SetMaxIdleConns (idle connection ceiling, not
	// a floor).
	MaxIdleConns   int32         `mapstructure:"max_idle_conns" validate:"min=0"`
	MaxConnIdle    time.Duration `mapstructure:"max_conn_idle" validate:"required,min=1s"`
	MaxConnLife    time.Duration `mapstructure:"max_conn_life" validate:"required,min=1s"`
	ConnectTimeout time.Duration `mapstructure:"connect_timeout" validate:"required,min=1s"`
	// SlowQueryThreshold is how long a query may run before GORM logs it as
	// slow. Zero falls back to the logger's built-in default.
	SlowQueryThreshold time.Duration `mapstructure:"slow_query_threshold" validate:"omitempty,min=1ms"`
}

// ValkeyConfig holds the Valkey client settings. TTL is how long a cache-aside
// entry stays valid in Valkey.
type ValkeyConfig struct {
	Host           string        `mapstructure:"host" validate:"required,hostname|ip"`
	Port           int           `mapstructure:"port" validate:"required,min=1,max=65535"`
	Password       string        `mapstructure:"password"`
	DB             int           `mapstructure:"db" validate:"min=0"`
	ConnectTimeout time.Duration `mapstructure:"connect_timeout" validate:"omitempty,min=1s"`
	TTL            time.Duration `mapstructure:"ttl" validate:"omitempty,min=1s"`
}

// OTelConfig holds OpenTelemetry exporter settings.
type OTelConfig struct {
	Exporter    string  `mapstructure:"exporter" validate:"oneof=otlp none"`
	Endpoint    string  `mapstructure:"endpoint"`
	ServiceName string  `mapstructure:"service_name" validate:"required"`
	Sampling    float64 `mapstructure:"sampling" validate:"min=0,max=1"`
}

// LogConfig holds the zerolog settings.
type LogConfig struct {
	Level  string `mapstructure:"level" validate:"oneof=trace debug info warn error fatal panic"`
	Format string `mapstructure:"format" validate:"oneof=json console"`
}

// ExampleConfig is a feature toggle and settings for the stub feature.
type ExampleConfig struct {
	Enabled         bool  `mapstructure:"enabled"`
	DefaultPageSize int32 `mapstructure:"default_page_size"`
	MaxPageSize     int32 `mapstructure:"max_page_size"`
	MaxNameLength   int32 `mapstructure:"max_name_length"`
}

// exampleMaxPageSizeUpperBound caps EXAMPLE_MAX_PAGE_SIZE to a sane ceiling so
// a misconfiguration cannot request unbounded result sets.
const exampleMaxPageSizeUpperBound int32 = 1000

// exampleMaxNameLengthUpperBound caps EXAMPLE_MAX_NAME_LENGTH to prevent
// unreasonable storage/validation costs per name.
const exampleMaxNameLengthUpperBound int32 = 4096

// validateExamplePositivity returns an error if any of the example config
// fields are less than 1, enforcing a minimum value when the feature is enabled.
func validateExamplePositivity(cfg ExampleConfig) error {
	if cfg.DefaultPageSize < 1 {
		return fmt.Errorf("EXAMPLE_DEFAULT_PAGE_SIZE must be >= 1")
	}
	if cfg.MaxPageSize < 1 {
		return fmt.Errorf("EXAMPLE_MAX_PAGE_SIZE must be >= 1")
	}
	if cfg.MaxNameLength < 1 {
		return fmt.Errorf("EXAMPLE_MAX_NAME_LENGTH must be >= 1")
	}
	return nil
}

// validate is the package-level validator instance.
var validate = validator.New()

// Load reads config.yaml (or CONFIG_FILE) and environment variables and returns
// a typed configuration. Environment variables are unprefixed and use
// SCREAMING_SNAKE names matching the nested config keys (e.g. http.port ->
// HTTP_PORT, otel.service_name -> OTEL_SERVICE_NAME).
func Load() (*Config, error) {
	v := viper.NewWithOptions(viper.ExperimentalBindStruct())

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	configFileExplicit := false
	if configFile, ok := os.LookupEnv("CONFIG_FILE"); ok && configFile != "" {
		configFileExplicit = true
		absPath, err := filepath.Abs(configFile)
		if err != nil {
			return nil, fmt.Errorf("resolve CONFIG_FILE path %q: %w", configFile, err)
		}
		v.SetConfigFile(absPath)
	}

	setDefaults(v)

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	for _, binding := range leafBindings() {
		if err := v.BindEnv(binding.key, binding.envName); err != nil {
			return nil, fmt.Errorf("bind env %s to key %s: %w", binding.envName, binding.key, err)
		}
	}

	if err := v.ReadInConfig(); err != nil {
		if configFileExplicit || !errorsIsConfigNotFound(err) {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	return &cfg, nil
}

// Validate runs go-playground/validator and cross-section checks.
func (c *Config) Validate() error {
	if err := validate.Struct(c); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}
	if err := c.validateOTel(); err != nil {
		return err
	}
	if c.HTTP.BodyLimit != "" {
		if _, err := ParseByteSize(c.HTTP.BodyLimit); err != nil {
			return fmt.Errorf("HTTP_BODY_LIMIT: %w", err)
		}
	}
	if c.DB.MaxConns < c.DB.MaxIdleConns {
		return fmt.Errorf("DB_MAX_CONNS must be >= DB_MAX_IDLE_CONNS")
	}
	if c.Example.Enabled {
		if err := validateExample(c.Example); err != nil {
			return err
		}
	}
	return nil
}

// validateOTel requires a well-formed endpoint when the OTLP exporter is on.
func (c *Config) validateOTel() error {
	if c.OTel.Exporter != "otlp" {
		return nil
	}
	if c.OTel.Endpoint == "" {
		return fmt.Errorf("OTEL_EXPORTER_OTLP_ENDPOINT is required when OTEL_EXPORTER=otlp")
	}
	u, err := url.Parse(c.OTel.Endpoint)
	if err != nil {
		return fmt.Errorf("OTEL_EXPORTER_OTLP_ENDPOINT is invalid: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("OTEL_EXPORTER_OTLP_ENDPOINT must include a scheme and host: %q", c.OTel.Endpoint)
	}
	return nil
}

// validateExample enforces the stub feature's bounds when it is enabled.
func validateExample(cfg ExampleConfig) error {
	if err := validateExamplePositivity(cfg); err != nil {
		return err
	}
	if cfg.DefaultPageSize > cfg.MaxPageSize {
		return fmt.Errorf("EXAMPLE_DEFAULT_PAGE_SIZE must be <= EXAMPLE_MAX_PAGE_SIZE")
	}
	if cfg.MaxPageSize > exampleMaxPageSizeUpperBound {
		return fmt.Errorf("EXAMPLE_MAX_PAGE_SIZE exceeds maximum allowed value %d", exampleMaxPageSizeUpperBound)
	}
	if cfg.MaxNameLength > exampleMaxNameLengthUpperBound {
		return fmt.Errorf("EXAMPLE_MAX_NAME_LENGTH exceeds maximum allowed value %d", exampleMaxNameLengthUpperBound)
	}
	return nil
}

// HTTPAddr returns the HTTP listen address.
func (c *Config) HTTPAddr() string {
	return net.JoinHostPort(c.HTTP.Host, strconv.Itoa(c.HTTP.Port))
}

// DBConnString returns a pgx-compatible DSN, including connect_timeout as an
// integer-second query parameter (minimum 1) so a dial cannot hang past the
// configured bound. The DSN is assembled from the typed fields rather than a
// formatted string, so credentials are percent-encoded and never appear
// unescaped in an error message.
func (c *Config) DBConnString() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.DB.User, c.DB.Password),
		Host:   net.JoinHostPort(c.DB.Host, strconv.Itoa(c.DB.Port)),
		Path:   "/" + c.DB.Name,
	}

	seconds := max(int(c.DB.ConnectTimeout/time.Second), 1)

	q := u.Query()
	q.Set("sslmode", c.DB.SSLMode)
	q.Set("connect_timeout", strconv.Itoa(seconds))
	u.RawQuery = q.Encode()

	return u.String()
}

// ValkeyAddr returns the Valkey server address.
func (c *Config) ValkeyAddr() string {
	return net.JoinHostPort(c.Valkey.Host, strconv.Itoa(c.Valkey.Port))
}

const defaultHost = "0.0.0.0"

// setDefaults registers default values used when a key is missing from both
// config file and environment.
func setDefaults(v *viper.Viper) {
	defaults := map[string]any{
		"app.environment":      "development",
		"app.shutdown_timeout": 15 * time.Second,

		"http.host":                 defaultHost,
		"http.port":                 8080,
		"http.read_timeout":         15 * time.Second,
		"http.write_timeout":        15 * time.Second,
		"http.idle_timeout":         60 * time.Second,
		"http.body_limit":           "1M",
		"http.health_probe_timeout": 5 * time.Second,
		"http.cors_allow_origins":   []string{},
		"http.cors_allow_methods":   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		"http.cors_allow_headers":   []string{"Authorization", "Content-Type", "X-Request-ID"},

		"db.ssl_mode":             "disable",
		"db.max_conns":            10,
		"db.max_idle_conns":       2,
		"db.max_conn_idle":        30 * time.Minute,
		"db.max_conn_life":        1 * time.Hour,
		"db.connect_timeout":      5 * time.Second,
		"db.slow_query_threshold": 200 * time.Millisecond,

		"valkey.db":              0,
		"valkey.connect_timeout": 5 * time.Second,
		"valkey.ttl":             30 * time.Second,

		"otel.exporter":     "none",
		"otel.service_name": "zercle-go-template",
		"otel.sampling":     1.0,

		"log.level":  "info",
		"log.format": "json",

		"example.enabled":           false,
		"example.default_page_size": int32(20),
		"example.max_page_size":     int32(100),
		"example.max_name_length":   int32(255),
	}

	for key, value := range defaults {
		v.SetDefault(key, value)
	}
}

// leafBindings returns the explicit env-var-to-config-key bindings used by
// Load.
func leafBindings() []leafBinding {
	return []leafBinding{
		{"app.environment", "APP_ENVIRONMENT"},
		{"app.shutdown_timeout", "APP_SHUTDOWN_TIMEOUT"},

		{"http.host", "HTTP_HOST"},
		{"http.port", "HTTP_PORT"},
		{"http.read_timeout", "HTTP_READ_TIMEOUT"},
		{"http.write_timeout", "HTTP_WRITE_TIMEOUT"},
		{"http.idle_timeout", "HTTP_IDLE_TIMEOUT"},
		{"http.body_limit", "HTTP_BODY_LIMIT"},
		{"http.health_probe_timeout", "HTTP_HEALTH_PROBE_TIMEOUT"},
		{"http.cors_allow_origins", "HTTP_CORS_ALLOW_ORIGINS"},
		{"http.cors_allow_methods", "HTTP_CORS_ALLOW_METHODS"},
		{"http.cors_allow_headers", "HTTP_CORS_ALLOW_HEADERS"},

		{"db.host", "DB_HOST"},
		{"db.port", "DB_PORT"},
		{"db.name", "DB_NAME"},
		{"db.user", "DB_USER"},
		{"db.password", "DB_PASSWORD"},
		{"db.ssl_mode", "DB_SSL_MODE"},
		{"db.max_conns", "DB_MAX_CONNS"},
		{"db.max_idle_conns", "DB_MAX_IDLE_CONNS"},
		{"db.max_conn_idle", "DB_MAX_CONN_IDLE"},
		{"db.max_conn_life", "DB_MAX_CONN_LIFE"},
		{"db.connect_timeout", "DB_CONNECT_TIMEOUT"},
		{"db.slow_query_threshold", "DB_SLOW_QUERY_THRESHOLD"},

		{"valkey.host", "VALKEY_HOST"},
		{"valkey.port", "VALKEY_PORT"},
		{"valkey.password", "VALKEY_PASSWORD"},
		{"valkey.db", "VALKEY_DB"},
		{"valkey.ttl", "VALKEY_TTL"},
		{"valkey.connect_timeout", "VALKEY_CONNECT_TIMEOUT"},

		{"log.level", "LOG_LEVEL"},
		{"log.format", "LOG_FORMAT"},

		{"otel.exporter", "OTEL_EXPORTER"},
		{"otel.endpoint", "OTEL_EXPORTER_OTLP_ENDPOINT"},
		{"otel.service_name", "OTEL_SERVICE_NAME"},
		{"otel.sampling", "OTEL_TRACES_SAMPLER_ARG"},

		{"example.enabled", "EXAMPLE_ENABLED"},
		{"example.default_page_size", "EXAMPLE_DEFAULT_PAGE_SIZE"},
		{"example.max_page_size", "EXAMPLE_MAX_PAGE_SIZE"},
		{"example.max_name_length", "EXAMPLE_MAX_NAME_LENGTH"},
	}
}

// errorsIsConfigNotFound reports whether err is a viper config file not found
// error. It uses errors.AsType to avoid string comparison.
func errorsIsConfigNotFound(err error) bool {
	if err == nil {
		return false
	}
	_, ok := errors.AsType[viper.ConfigFileNotFoundError](err)
	return ok
}

// ParseByteSize parses a human-friendly byte size such as "1M", "512KiB", or a
// bare byte count into a raw byte count. An empty string yields (0, nil), which
// callers treat as "unset". A non-empty but unparseable, non-positive, or
// overflowing value yields an error, so a misconfigured limit fails loudly at
// startup instead of silently disabling the protection it configures.
func ParseByteSize(s string) (int64, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return 0, nil
	}

	upper := strings.ToUpper(trimmed)
	upper = strings.TrimSuffix(upper, "B")
	upper = strings.TrimSuffix(upper, "I")
	multiplier := int64(1)
	switch {
	case strings.HasSuffix(upper, "K"):
		multiplier = 1024
		upper = strings.TrimSuffix(upper, "K")
	case strings.HasSuffix(upper, "M"):
		multiplier = 1024 * 1024
		upper = strings.TrimSuffix(upper, "M")
	case strings.HasSuffix(upper, "G"):
		multiplier = 1024 * 1024 * 1024
		upper = strings.TrimSuffix(upper, "G")
	}
	upper = strings.TrimSpace(upper)

	n, err := strconv.ParseInt(upper, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid byte size %q: %w", s, err)
	}
	if n <= 0 {
		return 0, fmt.Errorf("invalid byte size %q: must be positive", s)
	}
	if n > math.MaxInt64/multiplier {
		return 0, fmt.Errorf("invalid byte size %q: overflows int64", s)
	}
	return n * multiplier, nil
}
