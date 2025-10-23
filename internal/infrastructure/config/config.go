package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server    ServerConfig
	Logging   LoggingConfig
	JWT       JWTConfig
	CORS      CORSConfig
	Database  DatabaseConfig
	RateLimit RateLimitConfig
	Argon2id  Argon2idConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Env  string `mapstructure:"env"`
	Port int    `mapstructure:"port"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host              string        `mapstructure:"host"`
	User              string        `mapstructure:"user"`
	Password          string        `mapstructure:"password"`
	DBName            string        `mapstructure:"dbname"`
	Driver            string        `mapstructure:"driver"`
	Port              int           `mapstructure:"port"`
	MaxConnLifetime   time.Duration `mapstructure:"max_conn_lifetime"`
	MaxConnIdleTime   time.Duration `mapstructure:"max_conn_idletime"`
	HealthCheckPeriod time.Duration `mapstructure:"health_check_period"`
	MaxConns          int32         `mapstructure:"max_conns"`
	MinConns          int32         `mapstructure:"min_conns"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	Expiration int    `mapstructure:"expiration"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Requests int `mapstructure:"requests"`
	Window   int `mapstructure:"window"`
}

// Argon2idConfig holds Argon2id password hashing configuration
// OWASP recommended: memory=19456 (19 MiB), iterations=2, parallelism=1
type Argon2idConfig struct {
	Memory      uint32 `mapstructure:"memory"`
	Iterations  uint32 `mapstructure:"iterations"`
	Parallelism uint8  `mapstructure:"parallelism"`
	SaltLength  uint32 `mapstructure:"salt_length"`
	KeyLength   uint32 `mapstructure:"key_length"`
}

// Load reads configuration from the specified YAML file and merges with environment variables.
// Environment variables take precedence over file-based configuration values.
// Returns an error if the file cannot be read or if validation fails.
func Load(configPath string) (*Config, error) {
	v := viper.New()

	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	setDefaults(v)

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if env := os.Getenv("SERVER_PORT"); env != "" {
		_, _ = fmt.Sscanf(env, "%d", &cfg.Server.Port)
	}
	if env := os.Getenv("SERVER_HOST"); env != "" {
		cfg.Server.Host = env
	}
	if env := os.Getenv("SERVER_ENV"); env != "" {
		cfg.Server.Env = env
	}
	if env := os.Getenv("DB_HOST"); env != "" {
		cfg.Database.Host = env
	}
	if env := os.Getenv("DB_PORT"); env != "" {
		_, _ = fmt.Sscanf(env, "%d", &cfg.Database.Port)
	}
	if env := os.Getenv("DB_USER"); env != "" {
		cfg.Database.User = env
	}
	if env := os.Getenv("DB_PASSWORD"); env != "" {
		cfg.Database.Password = env
	}
	if env := os.Getenv("DB_NAME"); env != "" {
		cfg.Database.DBName = env
	}
	if env := os.Getenv("DB_DRIVER"); env != "" {
		cfg.Database.Driver = env
	}
	if env := os.Getenv("JWT_SECRET"); env != "" {
		cfg.JWT.Secret = env
	}
	if env := os.Getenv("LOG_LEVEL"); env != "" {
		cfg.Logging.Level = env
	}
	if env := os.Getenv("ARGON2ID_MEMORY"); env != "" {
		var val uint32
		_, _ = fmt.Sscanf(env, "%d", &val)
		cfg.Argon2id.Memory = val
	}
	if env := os.Getenv("ARGON2ID_ITERATIONS"); env != "" {
		var val uint32
		_, _ = fmt.Sscanf(env, "%d", &val)
		cfg.Argon2id.Iterations = val
	}
	if env := os.Getenv("ARGON2ID_PARALLELISM"); env != "" {
		var val uint8
		_, _ = fmt.Sscanf(env, "%d", &val)
		cfg.Argon2id.Parallelism = val
	}

	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// setDefaults configures default values for all configuration fields.
// These defaults are used when values are not provided in the config file or environment.
func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", 3000)
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.env", "local")

	v.SetDefault("database.driver", "postgres")
	v.SetDefault("database.port", 5432)

	v.SetDefault("jwt.expiration", 3600)

	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")

	v.SetDefault("cors.allowed_origins", []string{"*"})

	v.SetDefault("rate_limit.requests", 100)
	v.SetDefault("rate_limit.window", 60)

	v.SetDefault("argon2id.memory", 19456)
	v.SetDefault("argon2id.iterations", 2)
	v.SetDefault("argon2id.parallelism", 1)
	v.SetDefault("argon2id.salt_length", 16)
	v.SetDefault("argon2id.key_length", 32)
}

// validate checks that all required configuration values are present and valid.
// Returns an error describing the first validation failure encountered.
func validate(cfg *Config) error {
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}
	if cfg.Server.Host == "" {
		return errors.New("server host cannot be empty")
	}

	if cfg.Database.Host == "" {
		return errors.New("database host cannot be empty")
	}
	if cfg.Database.Port <= 0 || cfg.Database.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", cfg.Database.Port)
	}
	if cfg.Database.DBName == "" {
		return errors.New("database name cannot be empty")
	}

	if cfg.JWT.Secret == "" {
		return errors.New("JWT secret cannot be empty")
	}
	return nil
}
