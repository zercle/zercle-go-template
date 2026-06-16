// Package config loads application configuration from config.yaml and
// environment variables. Validation is performed by Config.Validate.
package config

import (
	"errors"
	"fmt"
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
type Config struct {
	App     AppConfig     `mapstructure:"app" yaml:"app" validate:"required"`
	HTTP    HTTPConfig    `mapstructure:"http" yaml:"http" validate:"required"`
	GRPC    GRPCConfig    `mapstructure:"grpc" yaml:"grpc" validate:"required"`
	DB      DBConfig      `mapstructure:"db" yaml:"db" validate:"required"`
	Valkey  ValkeyConfig  `mapstructure:"valkey" yaml:"valkey" validate:"required"`
	OTel    OTelConfig    `mapstructure:"otel" yaml:"otel" validate:"required"`
	Log     LogConfig     `mapstructure:"log" yaml:"log" validate:"required"`
	Example ExampleConfig `mapstructure:"example" yaml:"example"`
}

// AppConfig holds process-level settings.
type AppConfig struct {
	Name            string        `mapstructure:"name" yaml:"name" env:"APP_NAME" validate:"required"`
	Environment     string        `mapstructure:"environment" yaml:"environment" env:"APP_ENVIRONMENT" validate:"oneof=development staging production test"`
	Host            string        `mapstructure:"host" yaml:"host" env:"APP_HOST" validate:"ip|hostname"`
	Port            int           `mapstructure:"port" yaml:"port" env:"APP_PORT" validate:"required,min=1,max=65535"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" yaml:"shutdown_timeout" env:"APP_SHUTDOWN_TIMEOUT" validate:"required,min=1s"`
}

// HTTPConfig holds the HTTP server settings and CORS options.
type HTTPConfig struct {
	Host               string        `mapstructure:"host" yaml:"host" env:"HTTP_HOST" validate:"ip|hostname"`
	Port               int           `mapstructure:"port" yaml:"port" env:"HTTP_PORT" validate:"required,min=1,max=65535"`
	ReadTimeout        time.Duration `mapstructure:"read_timeout" yaml:"read_timeout" env:"HTTP_READ_TIMEOUT" validate:"required,min=1s"`
	WriteTimeout       time.Duration `mapstructure:"write_timeout" yaml:"write_timeout" env:"HTTP_WRITE_TIMEOUT" validate:"required,min=1s"`
	IdleTimeout        time.Duration `mapstructure:"idle_timeout" yaml:"idle_timeout" env:"HTTP_IDLE_TIMEOUT" validate:"required,min=1s"`
	BodyLimit          string        `mapstructure:"body_limit" yaml:"body_limit" env:"HTTP_BODY_LIMIT" validate:"required"`
	HealthProbeTimeout time.Duration `mapstructure:"health_probe_timeout" yaml:"health_probe_timeout" env:"HTTP_HEALTH_PROBE_TIMEOUT" validate:"required,min=1s"`
	CORSAllowOrigins   []string      `mapstructure:"cors_allow_origins" yaml:"cors_allow_origins" env:"HTTP_CORS_ALLOW_ORIGINS"`
	CORSAllowMethods   []string      `mapstructure:"cors_allow_methods" yaml:"cors_allow_methods" env:"HTTP_CORS_ALLOW_METHODS"`
	CORSAllowHeaders   []string      `mapstructure:"cors_allow_headers" yaml:"cors_allow_headers" env:"HTTP_CORS_ALLOW_HEADERS"`
}

// GRPCConfig holds the gRPC server settings.
type GRPCConfig struct {
	Host string `mapstructure:"host" yaml:"host" env:"GRPC_HOST" validate:"ip|hostname"`
	Port int    `mapstructure:"port" yaml:"port" env:"GRPC_PORT" validate:"required,min=1,max=65535"`
}

// DBConfig holds the PostgreSQL connection and pool settings.
type DBConfig struct {
	Host           string        `mapstructure:"host" yaml:"host" env:"DB_HOST" validate:"required,hostname|ip"`
	Port           int           `mapstructure:"port" yaml:"port" env:"DB_PORT" validate:"required,min=1,max=65535"`
	Name           string        `mapstructure:"name" yaml:"name" env:"DB_NAME" validate:"required"`
	User           string        `mapstructure:"user" yaml:"user" env:"DB_USER" validate:"required"`
	Password       string        `mapstructure:"password" yaml:"password" env:"DB_PASSWORD" validate:"required"`
	SSLMode        string        `mapstructure:"ssl_mode" yaml:"ssl_mode" env:"DB_SSL_MODE" validate:"oneof=disable prefer require verify-ca verify-full"`
	MaxConns       int32         `mapstructure:"max_conns" yaml:"max_conns" env:"DB_MAX_CONNS" validate:"required,min=1"`
	MinConns       int32         `mapstructure:"min_conns" yaml:"min_conns" env:"DB_MIN_CONNS" validate:"min=0"`
	MaxConnIdle    time.Duration `mapstructure:"max_conn_idle" yaml:"max_conn_idle" env:"DB_MAX_CONN_IDLE" validate:"required,min=1s"`
	MaxConnLife    time.Duration `mapstructure:"max_conn_life" yaml:"max_conn_life" env:"DB_MAX_CONN_LIFE" validate:"required,min=1s"`
	ConnectTimeout time.Duration `mapstructure:"connect_timeout" yaml:"connect_timeout" env:"DB_CONNECT_TIMEOUT" validate:"required,min=1s"`
}

// ValkeyConfig holds the Valkey client settings.
type ValkeyConfig struct {
	Host           string        `mapstructure:"host" yaml:"host" env:"VALKEY_HOST" validate:"required,hostname|ip"`
	Port           int           `mapstructure:"port" yaml:"port" env:"VALKEY_PORT" validate:"required,min=1,max=65535"`
	Password       string        `mapstructure:"password" yaml:"password" env:"VALKEY_PASSWORD"`
	DB             int           `mapstructure:"db" yaml:"db" env:"VALKEY_DB" validate:"min=0"`
	ConnectTimeout time.Duration `mapstructure:"connect_timeout" yaml:"connect_timeout" env:"VALKEY_CONNECT_TIMEOUT" validate:"omitempty,min=1s"`
}

// OTelConfig holds OpenTelemetry exporter settings.
type OTelConfig struct {
	Exporter    string  `mapstructure:"exporter" yaml:"exporter" env:"OTEL_EXPORTER" validate:"oneof=otlp none"`
	Endpoint    string  `mapstructure:"endpoint" yaml:"endpoint" env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	ServiceName string  `mapstructure:"service_name" yaml:"service_name" env:"OTEL_SERVICE_NAME" validate:"required"`
	Sampling    float64 `mapstructure:"sampling" yaml:"sampling" env:"OTEL_TRACES_SAMPLER_ARG" validate:"min=0,max=1"`
}

// LogConfig holds the zerolog settings.
type LogConfig struct {
	Level  string `mapstructure:"level" yaml:"level" env:"LOG_LEVEL" validate:"oneof=trace debug info warn error fatal panic"`
	Format string `mapstructure:"format" yaml:"format" env:"LOG_FORMAT" validate:"oneof=json console"`
}

// ExampleConfig is a feature toggle and settings for the stub feature.
type ExampleConfig struct {
	Enabled         bool  `mapstructure:"enabled" yaml:"enabled" env:"EXAMPLE_ENABLED"`
	DefaultPageSize int32 `mapstructure:"default_page_size" yaml:"default_page_size" env:"EXAMPLE_DEFAULT_PAGE_SIZE" validate:"required,min=1"`
	MaxPageSize     int32 `mapstructure:"max_page_size" yaml:"max_page_size" env:"EXAMPLE_MAX_PAGE_SIZE" validate:"required,min=1"`
	MaxNameLength   int32 `mapstructure:"max_name_length" yaml:"max_name_length" env:"EXAMPLE_MAX_NAME_LENGTH" validate:"required,min=1"`
}

// validate is the package-level validator instance.
var validate = validator.New()

// Load reads config.yaml (or CONFIG_FILE) and environment variables and returns
// a typed configuration. Environment variables are unprefixed and use
// SCREAMING_SNAKE names matching the nested config keys (e.g. app.name ->
// APP_NAME, http.port -> HTTP_PORT).
func Load() (*Config, error) {
	v := viper.NewWithOptions(viper.ExperimentalBindStruct())

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	if configFile, ok := os.LookupEnv("CONFIG_FILE"); ok && configFile != "" {
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
		if !errorsIsConfigNotFound(err) {
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

	if c.OTel.Exporter == "otlp" && c.OTel.Endpoint == "" {
		return fmt.Errorf("OTEL_EXPORTER_OTLP_ENDPOINT is required when OTEL_EXPORTER=otlp")
	}

	if c.OTel.Exporter == "otlp" {
		if _, err := url.Parse(c.OTel.Endpoint); err != nil {
			return fmt.Errorf("OTEL_EXPORTER_OTLP_ENDPOINT is invalid: %w", err)
		}
	}

	if c.DB.MaxConns < c.DB.MinConns {
		return fmt.Errorf("DB_MAX_CONNS must be >= DB_MIN_CONNS")
	}

	return nil
}

// HTTPAddr returns the HTTP listen address.
func (c *Config) HTTPAddr() string {
	return net.JoinHostPort(c.HTTP.Host, strconv.Itoa(c.HTTP.Port))
}

// GRPCAddr returns the gRPC listen address.
func (c *Config) GRPCAddr() string {
	return net.JoinHostPort(c.GRPC.Host, strconv.Itoa(c.GRPC.Port))
}

// DBConnString returns a pgx-compatible DSN.
func (c *Config) DBConnString() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.DB.User, c.DB.Password),
		Host:   net.JoinHostPort(c.DB.Host, strconv.Itoa(c.DB.Port)),
		Path:   c.DB.Name,
	}
	q := u.Query()
	q.Set("sslmode", c.DB.SSLMode)
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
		"app.name":             "zercle-go-template",
		"app.environment":      "development",
		"app.host":             defaultHost,
		"app.port":             8080,
		"app.shutdown_timeout": 15 * time.Second,

		"http.host":               defaultHost,
		"http.port":               8080,
		"http.read_timeout":       15 * time.Second,
		"http.write_timeout":      15 * time.Second,
		"http.idle_timeout":       60 * time.Second,
		"http.body_limit":         "1M",
		"http.health_probe_timeout": 5 * time.Second,
		"http.cors_allow_origins": []string{},
		"http.cors_allow_methods": []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		"http.cors_allow_headers": []string{"Authorization", "Content-Type", "X-Request-ID"},

		"grpc.host": defaultHost,
		"grpc.port": 50051,

		"db.ssl_mode":        "disable",
		"db.max_conns":       10,
		"db.min_conns":       2,
		"db.max_conn_idle":   30 * time.Minute,
		"db.max_conn_life":   1 * time.Hour,
		"db.connect_timeout": 5 * time.Second,

		"valkey.db":              0,
		"valkey.connect_timeout": 5 * time.Second,

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
		{"app.name", "APP_NAME"},
		{"app.environment", "APP_ENVIRONMENT"},
		{"app.host", "APP_HOST"},
		{"app.port", "APP_PORT"},
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

		{"grpc.host", "GRPC_HOST"},
		{"grpc.port", "GRPC_PORT"},

		{"db.host", "DB_HOST"},
		{"db.port", "DB_PORT"},
		{"db.name", "DB_NAME"},
		{"db.user", "DB_USER"},
		{"db.password", "DB_PASSWORD"},
		{"db.ssl_mode", "DB_SSL_MODE"},
		{"db.max_conns", "DB_MAX_CONNS"},
		{"db.min_conns", "DB_MIN_CONNS"},
		{"db.max_conn_idle", "DB_MAX_CONN_IDLE"},
		{"db.max_conn_life", "DB_MAX_CONN_LIFE"},
		{"db.connect_timeout", "DB_CONNECT_TIMEOUT"},

		{"valkey.host", "VALKEY_HOST"},
		{"valkey.port", "VALKEY_PORT"},
		{"valkey.password", "VALKEY_PASSWORD"},
		{"valkey.db", "VALKEY_DB"},
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
// error. It uses errors.As to avoid string comparison.
func errorsIsConfigNotFound(err error) bool {
	var configFileNotFoundError viper.ConfigFileNotFoundError
	return err != nil && errors.As(err, &configFileNotFoundError)
}
