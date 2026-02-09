// Package config provides centralized configuration management for the application.
// It supports loading configuration from multiple sources with the following priority:
//  1. Runtime environment variables (os.Getenv) - highest priority
//  2. .env file (loaded by viper)
//  3. YAML config file (configs/config.yaml)
//  4. Default values - lowest priority
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
// It is populated from config files, environment variables, and defaults.
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Log      LogConfig      `mapstructure:"log"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Security SecurityConfig `mapstructure:"security"`
}

// AppConfig contains general application settings.
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
}

// ServerConfig contains HTTP server settings.
type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// DatabaseConfig contains database connection settings.
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

// LogConfig contains logging settings.
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// JWTConfig contains JWT authentication settings.
type JWTConfig struct {
	Secret          string        `mapstructure:"secret"`
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`
}

// SecurityConfig contains security-related settings.
type SecurityConfig struct {
	// Argon2Memory is the memory cost for Argon2id in KB.
	// Default: 65536 (64 MB) for production, 16384 (16 MB) for development/test.
	Argon2Memory int `mapstructure:"argon2_memory"`
	// Argon2Iterations is the number of iterations for Argon2id.
	// Default: 3 (OWASP recommended minimum).
	Argon2Iterations int `mapstructure:"argon2_iterations"`
	// Argon2Parallelism is the parallelism factor for Argon2id.
	// Default: 4 (number of threads).
	Argon2Parallelism int `mapstructure:"argon2_parallelism"`
	// Argon2SaltLength is the salt length in bytes.
	// Default: 16 bytes.
	Argon2SaltLength int `mapstructure:"argon2_salt_length"`
	// Argon2KeyLength is the key length in bytes.
	// Default: 32 bytes.
	Argon2KeyLength int `mapstructure:"argon2_key_length"`
}

// GetArgon2Params returns the configured Argon2id parameters.
// Production defaults: Memory=64MB, Iterations=3, Parallelism=4, SaltLength=16, KeyLength=32
// Development/Test defaults: Memory=16MB for faster tests
func (c *Config) GetArgon2Params() (memory, iterations int, parallelism uint8, saltLength, keyLength int) {
	// Set defaults based on environment
	if c.IsProduction() {
		memory = 64 * 1024 // 64 MB
		iterations = 3
		parallelism = 4
		saltLength = 16
		keyLength = 32
	} else {
		// Development/Test: lower memory for faster tests
		memory = 16 * 1024 // 16 MB
		iterations = 3
		parallelism = 4
		saltLength = 16
		keyLength = 32
	}

	// Override with explicit configuration if set
	if c.Security.Argon2Memory > 0 {
		memory = c.Security.Argon2Memory
	}
	if c.Security.Argon2Iterations > 0 {
		iterations = c.Security.Argon2Iterations
	}
	if c.Security.Argon2Parallelism > 0 {
		parallelism = uint8(c.Security.Argon2Parallelism)
	}
	if c.Security.Argon2SaltLength > 0 {
		saltLength = c.Security.Argon2SaltLength
	}
	if c.Security.Argon2KeyLength > 0 {
		keyLength = c.Security.Argon2KeyLength
	}

	return memory, iterations, parallelism, saltLength, keyLength
}

const envPrefix = "APP"

// Load loads configuration from all sources with the following priority:
//  1. Runtime environment variables (highest priority)
//  2. .env file (if exists)
//  3. YAML config file (if exists)
//  4. Default values (lowest priority)
func Load() (*Config, error) {
	return LoadWithPath("")
}

// LoadWithPath loads configuration with an optional config file path.
// If path is empty, it searches for config file in default locations.
func LoadWithPath(configPath string) (*Config, error) {
	v := viper.New()

	// Step 1: Set default values (lowest priority)
	setDefaults(v)

	// Step 2: Load from YAML config file if exists
	if err := loadFromYAML(v, configPath); err != nil {
		return nil, err
	}

	// Step 3: Load from .env file if exists (overrides YAML)
	if err := loadFromEnvFile(v); err != nil {
		return nil, err
	}

	// Step 4: Configure and load from environment variables (highest priority)
	// This will override any values from files
	if err := loadFromEnvVars(v); err != nil {
		return nil, err
	}

	// Step 5: Unmarshal config into struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Step 6: Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// LoadFromEnv loads configuration from environment variables only.
// This is useful for containerized environments where config is injected via env vars.
func LoadFromEnv() (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Configure and load from environment variables
	if err := loadFromEnvVars(v); err != nil {
		return nil, err
	}

	// Unmarshal config into struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// loadFromYAML loads configuration from YAML file.
func loadFromYAML(v *viper.Viper, configPath string) error {
	v.SetConfigType("yaml")

	if configPath != "" {
		// Use provided path
		v.SetConfigFile(configPath)
	} else {
		// Search in default locations
		v.SetConfigName("config")
		v.AddConfigPath(".")
		v.AddConfigPath("./configs")
	}

	if err := v.ReadInConfig(); err != nil {
		// Config file not found is not a fatal error
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}

	return nil
}

// loadFromEnvFile loads environment variables from .env file if it exists.
// It silently ignores if .env file doesn't exist.
// It tries multiple possible locations in order of priority.
func loadFromEnvFile(v *viper.Viper) error {
	// List of possible .env file locations (in order of priority)
	envPaths := []string{
		".env",       // Current directory
		"../.env",    // Parent directory (for tests)
		"../../.env", // Two levels up
	}

	for _, path := range envPaths {
		if _, err := os.Stat(path); err == nil {
			// File exists, try to load it
			envViper := viper.New()
			envViper.SetConfigFile(path)
			envViper.SetConfigType("env")

			if err := envViper.ReadInConfig(); err != nil {
				// Continue to next location if this one fails
				continue
			}

			// Merge the .env values into the main viper instance
			// Check if runtime env var exists first, otherwise use .env value
			// This ensures runtime env vars take priority over .env file
			for _, key := range envViper.AllKeys() {
				// Transform key from APP_SERVER_PORT format to server.port format
				transformedKey := transformEnvKey(key)
				// Construct the env var name (e.g., APP_SERVER_PORT)
				envVarName := envPrefix + "_" + strings.ToUpper(strings.ReplaceAll(transformedKey, ".", "_"))
				// If runtime env var is set, use it; otherwise use .env value
				if envVal := os.Getenv(envVarName); envVal != "" {
					v.Set(transformedKey, envVal)
				} else {
					v.Set(transformedKey, envViper.Get(key))
				}
			}

			// Only load the first found .env file
			break
		}
	}

	return nil
}

// transformEnvKey converts environment variable key format (APP_SERVER_PORT)
// to viper key format (server.port)
// When loading from .env file, viper keeps underscores (e.g., app_server_port)
// so we need to strip the "app_" prefix and convert underscores to dots.
func transformEnvKey(key string) string {
	// Viper keeps the keys as-is from .env (e.g., app_server_port)
	// Strip the "app_" prefix which corresponds to the APP_ env prefix
	prefix := strings.ToLower(envPrefix) + "_"
	key = strings.TrimPrefix(key, prefix)
	// Replace remaining underscores with dots for nested keys
	return strings.ReplaceAll(key, "_", ".")
}

// loadFromEnvVars configures viper to read from environment variables.
func loadFromEnvVars(v *viper.Viper) error {
	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	return nil
}

// setDefaults sets default configuration values.
func setDefaults(v *viper.Viper) {
	// App defaults
	v.SetDefault("app.name", "zercle-go-template")
	v.SetDefault("app.version", "1.0.0")
	v.SetDefault("app.environment", "development")

	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("server.shutdown_timeout", "10s")

	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.database", "zercle_template")
	v.SetDefault("database.username", "postgres")
	v.SetDefault("database.password", "")
	v.SetDefault("database.ssl_mode", "disable")

	// Log defaults
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")

	// JWT defaults
	v.SetDefault("jwt.secret", "your-secret-key-change-in-production")
	v.SetDefault("jwt.access_token_ttl", "15m")
	v.SetDefault("jwt.refresh_token_ttl", "168h")

	// Security defaults - Argon2id parameters
	// Production: Memory=64MB, Iterations=3, Parallelism=4
	// Development/Test: Memory=16MB for faster tests
	v.SetDefault("security.argon2_memory", 0)      // 0 means use environment-based default
	v.SetDefault("security.argon2_iterations", 0)  // 0 means use environment-based default
	v.SetDefault("security.argon2_parallelism", 0) // 0 means use environment-based default
	v.SetDefault("security.argon2_salt_length", 0) // 0 means use environment-based default
	v.SetDefault("security.argon2_key_length", 0)  // 0 means use environment-based default
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.App.Name == "" {
		return fmt.Errorf("app name is required")
	}

	if c.App.Environment == "" {
		return fmt.Errorf("app environment is required")
	}

	if c.Server.Port <= 0 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Server.ReadTimeout <= 0 {
		return fmt.Errorf("invalid server read timeout: %v", c.Server.ReadTimeout)
	}

	if c.Server.WriteTimeout <= 0 {
		return fmt.Errorf("invalid server write timeout: %v", c.Server.WriteTimeout)
	}

	if c.Log.Level == "" {
		return fmt.Errorf("log level is required")
	}

	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[c.Log.Level] {
		return fmt.Errorf("invalid log level: %s", c.Log.Level)
	}

	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}

	if c.JWT.AccessTokenTTL <= 0 {
		return fmt.Errorf("invalid JWT access token TTL: %v", c.JWT.AccessTokenTTL)
	}

	if c.JWT.RefreshTokenTTL <= 0 {
		return fmt.Errorf("invalid JWT refresh token TTL: %v", c.JWT.RefreshTokenTTL)
	}

	return nil
}

// IsDevelopment returns true if running in development environment.
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// IsProduction returns true if running in production environment.
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

// DatabaseDSN returns the PostgreSQL connection string.
func (c *Config) DatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.Username,
		c.Database.Password,
		c.Database.Database,
		c.Database.SSLMode,
	)
}

// =============================================================================
// Helper functions for environment variable parsing
// =============================================================================

// GetEnvString retrieves a string value from environment variable.
// Returns defaultValue if the environment variable is not set or empty.
func GetEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvInt retrieves an integer value from environment variable.
// Returns defaultValue if the environment variable is not set or invalid.
func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetEnvBool retrieves a boolean value from environment variable.
// Returns defaultValue if the environment variable is not set or invalid.
// Recognizes: "true", "1", "yes", "on" as true; "false", "0", "no", "off" as false.
func GetEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch strings.ToLower(value) {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		}
	}
	return defaultValue
}

// GetEnvDuration retrieves a duration value from environment variable.
// Returns defaultValue if the environment variable is not set or invalid.
// Duration format examples: "30s", "1m", "2h", "24h"
func GetEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
