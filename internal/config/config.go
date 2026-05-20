package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Config holds application configuration loaded from file and environment.
type Config struct {
	ServerHost             string        `mapstructure:"server_host"`
	ServerPort             int           `mapstructure:"server_port"`
	GRPCHost               string        `mapstructure:"grpc_host"`
	GRPCPort               int           `mapstructure:"grpc_port"`
	DBHost                 string        `mapstructure:"db_host"`
	DBPort                 int           `mapstructure:"db_port"`
	DBName                 string        `mapstructure:"db_name"`
	DBUser                 string        `mapstructure:"db_user"`
	DBPassword             string        `mapstructure:"db_password"`
	DBSSLMode              string        `mapstructure:"db_ssl_mode"`
	DBMaxConns             int32         `mapstructure:"db_max_conns"`
	DBMaxIdleConns         int32         `mapstructure:"db_max_idle_conns"`
	CacheHost              string        `mapstructure:"cache_host"`
	CachePort              int           `mapstructure:"cache_port"`
	CachePassword          string        `mapstructure:"cache_password"`
	CacheDB                int           `mapstructure:"cache_db"`
	AuthAccessTokenSecret  string        `mapstructure:"auth_access_token_secret"`
	AuthRefreshTokenSecret string        `mapstructure:"auth_refresh_token_secret"`
	AuthAccessTokenTTL     time.Duration `mapstructure:"auth_access_token_ttl"`
	AuthRefreshTokenTTL    time.Duration `mapstructure:"auth_refresh_token_ttl"`
	AuthIssuer             string        `mapstructure:"auth_issuer"`
	LogLevel               string        `mapstructure:"log_level"`
	LogFormat              string        `mapstructure:"log_format"`
	MetricsEnabled         bool          `mapstructure:"metrics_enabled"`
	MetricsPort            int           `mapstructure:"metrics_port"`
	TracingEnabled         bool          `mapstructure:"tracing_enabled"`
	TracingEndpoint        string        `mapstructure:"tracing_endpoint"`
	OTELServiceName        string        `mapstructure:"otel_service_name"`
	AppEnvironment         string        `mapstructure:"app_environment"`
}

// Load reads configuration from file and environment variables.
func Load() Config {
	v := viper.New() //nolint:forbidigo // config layer abstracts viper usage

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	if configFile := os.Getenv("CONFIG_FILE"); configFile != "" {
		v.SetConfigFile(configFile)
	}
	_ = v.ReadInConfig()

	v.AutomaticEnv()
	envBindings := []struct {
		key string
		env string
	}{
		{"server_host", "SERVER_HOST"},
		{"server_port", "SERVER_PORT"},
		{"grpc_host", "GRPC_HOST"},
		{"grpc_port", "GRPC_PORT"},
		{"db_host", "DB_HOST"},
		{"db_port", "DB_PORT"},
		{"db_name", "DB_NAME"},
		{"db_user", "DB_USER"},
		{"db_password", "DB_PASSWORD"},
		{"db_ssl_mode", "DB_SSL_MODE"},
		{"db_max_conns", "DB_MAX_CONNS"},
		{"db_max_idle_conns", "DB_MAX_IDLE_CONNS"},
		{"cache_host", "CACHE_HOST"},
		{"cache_port", "CACHE_PORT"},
		{"cache_password", "CACHE_PASSWORD"},
		{"cache_db", "CACHE_DB"},
		{"auth_access_token_secret", "AUTH_ACCESS_TOKEN_SECRET"},
		{"auth_refresh_token_secret", "AUTH_REFRESH_TOKEN_SECRET"},
		{"auth_access_token_ttl", "AUTH_ACCESS_TOKEN_TTL"},
		{"auth_refresh_token_ttl", "AUTH_REFRESH_TOKEN_TTL"},
		{"auth_issuer", "AUTH_ISSUER"},
		{"log_level", "LOG_LEVEL"},
		{"log_format", "LOG_FORMAT"},
		{"metrics_enabled", "METRICS_ENABLED"},
		{"metrics_port", "METRICS_PORT"},
		{"tracing_enabled", "TRACING_ENABLED"},
		{"tracing_endpoint", "TRACING_ENDPOINT"},
		{"otel_service_name", "OTEL_SERVICE_NAME"},
		{"app_environment", "APP_ENVIRONMENT"},
	}
	for _, binding := range envBindings {
		if err := v.BindEnv(binding.key, binding.env); err != nil {
			panic(fmt.Sprintf("failed to bind env %s: %v", binding.env, err))
		}
	}

	v.SetDefault("server_host", "0.0.0.0")
	v.SetDefault("server_port", 8080)
	v.SetDefault("grpc_host", "0.0.0.0")
	v.SetDefault("grpc_port", 9090)
	v.SetDefault("db_host", "localhost")
	v.SetDefault("db_port", 5432)
	v.SetDefault("db_ssl_mode", "disable")
	v.SetDefault("db_max_conns", 10)
	v.SetDefault("db_max_idle_conns", 5)
	v.SetDefault("cache_host", "localhost")
	v.SetDefault("cache_port", 6379)
	v.SetDefault("cache_db", 0)
	v.SetDefault("auth_access_token_ttl", "24h")
	v.SetDefault("auth_refresh_token_ttl", "168h")
	v.SetDefault("auth_issuer", "zercle-go-template")
	v.SetDefault("log_level", "info")
	v.SetDefault("log_format", "json")
	v.SetDefault("metrics_port", 9090)
	v.SetDefault("otel_service_name", "zercle-go-template")
	v.SetDefault("app_environment", "development")

	var cfg Config
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:     &cfg,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(mapstructure.StringToTimeDurationHookFunc()),
	})
	if err != nil {
		panic(err)
	}
	if err := decoder.Decode(v.AllSettings()); err != nil {
		panic(err)
	}

	return cfg
}

// Validate checks that required configuration fields are set.
func (c *Config) Validate() error {
	if c.DBName == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if c.DBUser == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if c.AuthAccessTokenSecret == "" {
		return fmt.Errorf("AUTH_ACCESS_TOKEN_SECRET is required")
	}
	if c.AuthRefreshTokenSecret == "" {
		return fmt.Errorf("AUTH_REFRESH_TOKEN_SECRET is required")
	}
	return nil
}

// ServerAddr returns the HTTP server address in host:port format.
func (c *Config) ServerAddr() string {
	return net.JoinHostPort(c.ServerHost, strconv.Itoa(c.ServerPort))
}

// GRPCAddr returns the gRPC server address in host:port format.
func (c *Config) GRPCAddr() string {
	return net.JoinHostPort(c.GRPCHost, strconv.Itoa(c.GRPCPort))
}

// DBConnString returns the PostgreSQL connection string.
func (c *Config) DBConnString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

// CacheAddr returns the cache server address in host:port format.
func (c *Config) CacheAddr() string {
	return net.JoinHostPort(c.CacheHost, strconv.Itoa(c.CachePort))
}
