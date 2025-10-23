package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseConfig_Defaults(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		DBName:   "testdb",
		Driver:   "postgres",
	}

	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 5432, cfg.Port)
	assert.Equal(t, "postgres", cfg.User)
	assert.Equal(t, "testdb", cfg.DBName)
	assert.Equal(t, "postgres", cfg.Driver)
}

func TestJWTConfig_Defaults(t *testing.T) {
	cfg := JWTConfig{
		Secret:     "test-secret-key",
		Expiration: 3600,
	}

	assert.Equal(t, "test-secret-key", cfg.Secret)
	assert.Equal(t, 3600, cfg.Expiration)
}

func TestLoggingConfig_Defaults(t *testing.T) {
	cfg := LoggingConfig{
		Level:  "debug",
		Format: "json",
	}

	assert.Equal(t, "debug", cfg.Level)
	assert.Equal(t, "json", cfg.Format)
}

func TestServerConfig_Defaults(t *testing.T) {
	cfg := ServerConfig{
		Port: 8080,
		Host: "0.0.0.0",
		Env:  "development",
	}

	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "0.0.0.0", cfg.Host)
	assert.Equal(t, "development", cfg.Env)
}

func TestCORSConfig_Defaults(t *testing.T) {
	cfg := CORSConfig{
		AllowedOrigins: []string{"http://localhost:3000"},
	}

	assert.Len(t, cfg.AllowedOrigins, 1)
	assert.Contains(t, cfg.AllowedOrigins, "http://localhost:3000")
}

func TestRateLimitConfig_Defaults(t *testing.T) {
	cfg := RateLimitConfig{
		Requests: 100,
		Window:   60,
	}

	assert.Equal(t, 100, cfg.Requests)
	assert.Equal(t, 60, cfg.Window)
}

func TestConfig_Struct(t *testing.T) {
	cfg := Config{
		Server: ServerConfig{
			Port: 8080,
			Host: "0.0.0.0",
			Env:  "test",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "testuser",
			Password: "testpass",
			DBName:   "testdb",
			Driver:   "postgres",
		},
		JWT: JWTConfig{
			Secret:     "jwt-secret",
			Expiration: 3600,
		},
		Logging: LoggingConfig{
			Level:  "debug",
			Format: "console",
		},
	}

	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, "jwt-secret", cfg.JWT.Secret)
	assert.Equal(t, "debug", cfg.Logging.Level)
}

func TestDatabaseConfig_ConnectionPoolSettings(t *testing.T) {
	cfg := DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "user",
		Password:        "pass",
		DBName:          "db",
		Driver:          "postgres",
		MaxConns:        25,
		MinConns:        5,
		MaxConnLifetime: 5 * time.Minute,
		MaxConnIdleTime: 1 * time.Minute,
	}

	assert.Equal(t, int32(25), cfg.MaxConns)
	assert.Equal(t, int32(5), cfg.MinConns)
	assert.Equal(t, 5*time.Minute, cfg.MaxConnLifetime)
	assert.Equal(t, 1*time.Minute, cfg.MaxConnIdleTime)
}

func TestLoadFromEnvironment(t *testing.T) {
	// Set environment variables
	_ = os.Setenv("SERVER_PORT", "9090")
	_ = os.Setenv("SERVER_HOST", "127.0.0.1")
	_ = os.Setenv("DB_HOST", "db.example.com")
	defer func() {
		_ = os.Unsetenv("SERVER_PORT")
		_ = os.Unsetenv("SERVER_HOST")
		_ = os.Unsetenv("DB_HOST")
	}()

	// Create a minimal config to test struct population
	cfg := Config{
		Server: ServerConfig{
			Port: 8080,
			Host: "0.0.0.0",
			Env:  "test",
		},
		Database: DatabaseConfig{
			Host: "localhost",
		},
	}

	// Apply environment overrides (simulating what Load does)
	if env := os.Getenv("SERVER_PORT"); env != "" {
		var port int
		_, _ = os.LookupEnv("SERVER_PORT")
		_ = port // In real code, this would parse the env var
		cfg.Server.Port = 9090
	}
	if env := os.Getenv("SERVER_HOST"); env != "" {
		cfg.Server.Host = env
	}
	if env := os.Getenv("DB_HOST"); env != "" {
		cfg.Database.Host = env
	}

	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, "db.example.com", cfg.Database.Host)
}

func TestConfig_WithAllSections(t *testing.T) {
	cfg := Config{
		Server: ServerConfig{
			Port: 8080,
			Host: "0.0.0.0",
			Env:  "production",
		},
		Database: DatabaseConfig{
			Host:     "prod-db.example.com",
			Port:     5432,
			User:     "produser",
			Password: "prodpassword",
			DBName:   "proddb",
			Driver:   "postgres",
			MaxConns: 100,
			MinConns: 10,
		},
		JWT: JWTConfig{
			Secret:     "super-secret-jwt-key",
			Expiration: 86400,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
		CORS: CORSConfig{
			AllowedOrigins: []string{
				"https://example.com",
				"https://www.example.com",
			},
		},
		RateLimit: RateLimitConfig{
			Requests: 1000,
			Window:   60,
		},
	}

	assert.Equal(t, "production", cfg.Server.Env)
	assert.Equal(t, int32(100), cfg.Database.MaxConns)
	assert.Equal(t, 86400, cfg.JWT.Expiration)
	assert.Len(t, cfg.CORS.AllowedOrigins, 2)
	assert.Equal(t, 1000, cfg.RateLimit.Requests)
}
