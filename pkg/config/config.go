// Package config provides configuration management for the application.
// Configuration is loaded from environment variables and config files.
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Cache    CacheConfig    `mapstructure:"cache"`
	Log      LogConfig      `mapstructure:"log"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Host            string        `mapstructure:"host" env:"SERVER_HOST" envDefault:"0.0.0.0"`
	Port            int           `mapstructure:"port" env:"SERVER_PORT" envDefault:"8080"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" env:"SERVER_SHUTDOWN_TIMEOUT" envDefault:"30s"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout" env:"SERVER_READ_TIMEOUT" envDefault:"15s"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout" env:"SERVER_WRITE_TIMEOUT" envDefault:"15s"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout" env:"SERVER_IDLE_TIMEOUT" envDefault:"60s"`
}

// Address returns the server address in host:port format.
func (c ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	Host            string        `mapstructure:"host" env:"DB_HOST" envDefault:"localhost"`
	Port            int           `mapstructure:"port" env:"DB_PORT" envDefault:"5432"`
	User            string        `mapstructure:"user" env:"DB_USER" envDefault:"postgres"`
	Password        string        `mapstructure:"password" env:"DB_PASSWORD"`
	Database        string        `mapstructure:"database" env:"DB_NAME" envDefault:"app"`
	SSLMode         string        `mapstructure:"sslmode" env:"DB_SSLMODE" envDefault:"disable"`
	MaxConns        int32         `mapstructure:"max_conns" env:"DB_MAX_CONNS" envDefault:"25"`
	MinConns        int32         `mapstructure:"min_conns" env:"DB_MIN_CONNS" envDefault:"5"`
	MaxConnLifetime time.Duration `mapstructure:"max_conn_lifetime" env:"DB_MAX_CONN_LIFETIME" envDefault:"1h"`
	MaxConnIdleTime time.Duration `mapstructure:"max_conn_idle_time" env:"DB_MAX_CONN_IDLE_TIME" envDefault:"10m"`
}

// DSN returns the PostgreSQL connection string.
func (c DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.SSLMode,
	)
}

// CacheConfig holds Redis cache configuration.
type CacheConfig struct {
	Host     string `mapstructure:"host" env:"CACHE_HOST" envDefault:"localhost:6379"`
	Password string `mapstructure:"password" env:"CACHE_PASSWORD"`
	DB       int    `mapstructure:"db" env:"CACHE_DB" envDefault:"0"`
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level  string `mapstructure:"level" env:"LOG_LEVEL" envDefault:"info"`
	Format string `mapstructure:"format" env:"LOG_FORMAT" envDefault:"json"`
}

// Load reads configuration from environment variables and config files.
// It looks for config files in the following order:
// 1. Current directory
// 2. ./config
// 3. $HOME/.config/app
// 4. /etc/app
func Load() (*Config, error) {
	v := viper.New()

	// Enable environment variable binding
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.shutdown_timeout", "30s")
	v.SetDefault("server.read_timeout", "15s")
	v.SetDefault("server.write_timeout", "15s")
	v.SetDefault("server.idle_timeout", "60s")
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.name", "app")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.max_conns", 25)
	v.SetDefault("database.min_conns", 5)
	v.SetDefault("database.max_conn_lifetime", "1h")
	v.SetDefault("database.max_conn_idle_time", "10m")
	v.SetDefault("cache.host", "localhost:6379")
	v.SetDefault("cache.db", 0)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")

	// Read config file if present
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("$HOME/.config/app")
	v.AddConfigPath("/etc/app")

	// Ignore config file not found errors
	_ = v.ReadInConfig()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// Validate checks that required configuration values are set.
func (c *Config) Validate() error {
	if c.Database.Password == "" {
		return fmt.Errorf("database.password is required")
	}
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535")
	}
	if c.Database.Port < 1 || c.Database.Port > 65535 {
		return fmt.Errorf("database.port must be between 1 and 65535")
	}
	if c.Log.Level != "debug" && c.Log.Level != "info" && c.Log.Level != "warn" && c.Log.Level != "error" {
		return fmt.Errorf("log.level must be one of: debug, info, warn, error")
	}
	if c.Log.Format != "json" && c.Log.Format != "console" {
		return fmt.Errorf("log.format must be one of: json, console")
	}
	return nil
}
