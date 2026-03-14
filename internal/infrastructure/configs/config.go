package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Valkey   ValkeyConfig   `mapstructure:"valkey"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

// AppConfig holds application-level configuration.
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
}

// ServerConfig holds server configuration.
type ServerConfig struct {
	GRPC GRPCConfig `mapstructure:"grpc"`
	HTTP HTTPConfig `mapstructure:"http"`
}

// GRPCConfig holds gRPC server configuration.
type GRPCConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// Addr returns the gRPC server address.
func (g GRPCConfig) Addr() string {
	return fmt.Sprintf("%s:%d", g.Host, g.Port)
}

// HTTPConfig holds HTTP server configuration.
type HTTPConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// Addr returns the HTTP server address.
func (h HTTPConfig) Addr() string {
	return fmt.Sprintf("%s:%d", h.Host, h.Port)
}

// DatabaseConfig holds database configuration.
type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Name         string `mapstructure:"name"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	SSLMode      string `mapstructure:"ssl_mode"`
	MaxConns     int32  `mapstructure:"max_conns"`
	MaxIdleConns int32  `mapstructure:"max_idle_conns"`
}

// ConnString returns the PostgreSQL connection string.
func (d DatabaseConfig) ConnString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

// ValkeyConfig holds Valkey (Redis) configuration.
type ValkeyConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// Addr returns the Valkey server address.
func (v ValkeyConfig) Addr() string {
	return fmt.Sprintf("%s:%d", v.Host, v.Port)
}

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	JWTSecret     string        `mapstructure:"jwt_secret"`
	JWTExpiry     time.Duration `mapstructure:"jwt_expiry"`
	RefreshExpiry time.Duration `mapstructure:"refresh_expiry"`
}

// LoggingConfig holds logging configuration.
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Load reads configuration from file and environment variables.
func Load(path string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.config/zercle-go-template")
	viper.AddConfigPath("/etc/zercle-go-template")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("config file not found: %w", err)
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
