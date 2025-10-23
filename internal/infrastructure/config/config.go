package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
	customValidator "github.com/zercle/zercle-go-template/pkg/utils/validator"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	CORS     CORSConfig
	Logging  LoggingConfig
}

// ServerConfig contains server configuration
type ServerConfig struct {
	Host            string        `validate:"required,hostname|ip"`
	Port            int           `validate:"required,gte=1,lte=65535"`
	Environment     string        `validate:"required,oneof=dev test production"`
	ReadTimeout     time.Duration `validate:"required,gte=1s"`
	WriteTimeout    time.Duration `validate:"required,gte=1s"`
	ShutdownTimeout time.Duration `validate:"required,gte=1s"`
}

// DatabaseConfig contains database configuration
type DatabaseConfig struct {
	Host            string        `validate:"required"`
	Port            int           `validate:"required,gte=1,lte=65535"`
	User            string        `validate:"required"`
	Password        string        `validate:"required"`
	Name            string        `validate:"required"`
	Driver          string        `validate:"required,oneof=postgres mysql sqlite"`
	MaxOpenConns    int           `validate:"gte=0"`
	MaxIdleConns    int           `validate:"gte=0"`
	ConnMaxLifetime time.Duration `validate:"gte=0s"`
}

// JWTConfig contains JWT configuration
type JWTConfig struct {
	Secret     string        `validate:"required,min=32"`
	Expiration time.Duration `validate:"required,gte=1m"`
}

// CORSConfig contains CORS configuration
type CORSConfig struct {
	AllowedOrigins     []string `validate:"required,min=1,dive,required"`
	AllowedMethods     []string `validate:"required,min=1,dive,required"`
	AllowedHeaders     []string `validate:"required,min=1,dive,required"`
	AllowedCredentials bool
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level  string `validate:"required,oneof=debug info warn error"` // debug, info, warn, error
	Format string `validate:"required,oneof=json text"`             // json, text
}

// Load loads configuration from environment variables and config files
// Loading precedence (lowest to highest priority):
// 1. Defaults (hardcoded values)
// 2. YAML config file (configs/{env}.yaml)
// 3. Environment variables (highest priority)
func Load() (*Config, error) {
	// 1. Set Defaults first
	setDefaults()

	// 2. Determine environment (from env var or default)
	// We need to check env var before loading YAML
	env := getEnv("SERVER_ENV", "dev")

	// 3. Load from YAML config file
	viper.SetConfigName(env)
	viper.SetConfigType("yaml")
	viper.AddConfigPath("configs")
	viper.AddConfigPath(".")

	// Try reading config file (YAML overrides defaults)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// It's okay if file not found, we rely on defaults/env
	}

	// 4. Bind environment variables (env vars override YAML)
	// This ensures SERVER_PORT overrides server.port from YAML
	bindEnvs()

	// 5. Enable automatic env reading (final override)
	viper.AutomaticEnv()

	// 6. Unmarshal into Struct
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Fix for Environment not being in YAML usually
	if cfg.Server.Environment == "" {
		cfg.Server.Environment = env
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// bindEnvs binds environment variables to struct fields
func bindEnvs() {
	_ = viper.BindEnv("server.port", "SERVER_PORT")
	_ = viper.BindEnv("server.host", "SERVER_HOST")
	_ = viper.BindEnv("server.environment", "SERVER_ENV")

	_ = viper.BindEnv("database.host", "DB_HOST")
	_ = viper.BindEnv("database.port", "DB_PORT")
	_ = viper.BindEnv("database.user", "DB_USER")
	_ = viper.BindEnv("database.password", "DB_PASSWORD")
	_ = viper.BindEnv("database.name", "DB_NAME")
	_ = viper.BindEnv("database.driver", "DB_DRIVER")

	_ = viper.BindEnv("jwt.secret", "JWT_SECRET")
	_ = viper.BindEnv("logging.level", "LOG_LEVEL")
}

// setDefaults sets default configuration values
func setDefaults() {
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 3000)
	viper.SetDefault("server.readTimeout", "10s")
	viper.SetDefault("server.writeTimeout", "10s")

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.name", "zercle_db")
	viper.SetDefault("database.driver", "postgres")

	viper.SetDefault("jwt.secret", "default")
	viper.SetDefault("logging.level", "info")
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Validate validates critical configuration using both struct tags and custom business rules.
func (c *Config) Validate() error {
	// First, run struct tag validation
	if err := customValidator.Validate(c); err != nil {
		return err
	}

	// Additional business logic validation
	if c.JWT.Secret == "default" && c.Server.Environment == "production" {
		return fmt.Errorf("JWT_SECRET must be changed in production")
	}

	if c.JWT.Secret == "your-super-secret-jwt-key" && c.Server.Environment == "production" {
		return fmt.Errorf("JWT_SECRET must be changed in production")
	}

	// Validate DB connection pool sanity
	if c.Database.MaxIdleConns > c.Database.MaxOpenConns && c.Database.MaxOpenConns > 0 {
		return fmt.Errorf("database max idle connections (%d) cannot exceed max open connections (%d)",
			c.Database.MaxIdleConns, c.Database.MaxOpenConns)
	}

	return nil
}
