package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "load with defaults",
			envVars: map[string]string{},
			wantErr: false,
		},
		{
			name: "load with environment variables",
			envVars: map[string]string{
				"APP_APP_NAME":        "test-app",
				"APP_APP_VERSION":     "2.0.0",
				"APP_APP_ENVIRONMENT": "production",
				"APP_SERVER_PORT":     "9090",
				"APP_LOG_LEVEL":       "debug",
			},
			wantErr: false,
		},
		{
			name: "invalid server port",
			envVars: map[string]string{
				"APP_SERVER_PORT": "0",
			},
			wantErr: true,
			errMsg:  "invalid server port",
		},
		{
			name: "invalid log level",
			envVars: map[string]string{
				"APP_LOG_LEVEL": "invalid",
			},
			wantErr: true,
			errMsg:  "invalid log level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment variables after test
			defer func() {
				for k := range tt.envVars {
					_ = os.Unsetenv(k)
				}
			}()

			// Set environment variables
			for k, v := range tt.envVars {
				_ = os.Setenv(k, v)
			}

			cfg, err := Load()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)
		})
	}
}

func TestLoad_WithEnvVars(t *testing.T) {
	// Clean up all env vars after test
	defer func() {
		_ = os.Unsetenv("APP_APP_NAME")
		_ = os.Unsetenv("APP_APP_VERSION")
		_ = os.Unsetenv("APP_APP_ENVIRONMENT")
		_ = os.Unsetenv("APP_SERVER_PORT")
		_ = os.Unsetenv("APP_SERVER_HOST")
		_ = os.Unsetenv("APP_LOG_LEVEL")
		_ = os.Unsetenv("APP_LOG_FORMAT")
		_ = os.Unsetenv("APP_JWT_SECRET")
		_ = os.Unsetenv("APP_JWT_ACCESS_TOKEN_TTL")
		_ = os.Unsetenv("APP_JWT_REFRESH_TOKEN_TTL")
	}()

	_ = os.Setenv("APP_APP_NAME", "test-service")
	_ = os.Setenv("APP_APP_VERSION", "v1.2.3")
	_ = os.Setenv("APP_APP_ENVIRONMENT", "staging")
	_ = os.Setenv("APP_SERVER_PORT", "8888")
	_ = os.Setenv("APP_SERVER_HOST", "127.0.0.1")
	_ = os.Setenv("APP_LOG_LEVEL", "warn")
	_ = os.Setenv("APP_LOG_FORMAT", "console")
	_ = os.Setenv("APP_JWT_SECRET", "test-jwt-secret")
	_ = os.Setenv("APP_JWT_ACCESS_TOKEN_TTL", "30m")
	_ = os.Setenv("APP_JWT_REFRESH_TOKEN_TTL", "24h")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "test-service", cfg.App.Name)
	assert.Equal(t, "v1.2.3", cfg.App.Version)
	assert.Equal(t, "staging", cfg.App.Environment)
	assert.Equal(t, 8888, cfg.Server.Port)
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, "warn", cfg.Log.Level)
	assert.Equal(t, "console", cfg.Log.Format)
	assert.Equal(t, "test-jwt-secret", cfg.JWT.Secret)
	assert.Equal(t, 30*time.Minute, cfg.JWT.AccessTokenTTL)
	assert.Equal(t, 24*time.Hour, cfg.JWT.RefreshTokenTTL)
}

func TestLoadFromEnv(t *testing.T) {
	// Clean up all env vars after test
	defer func() {
		envVars := []string{
			"APP_APP_NAME", "APP_APP_VERSION", "APP_APP_ENVIRONMENT",
			"APP_SERVER_PORT", "APP_SERVER_HOST",
			"APP_DATABASE_HOST", "APP_DATABASE_PORT",
			"APP_LOG_LEVEL", "APP_LOG_FORMAT",
			"APP_JWT_SECRET",
		}
		for _, v := range envVars {
			_ = os.Unsetenv(v)
		}
	}()

	// Set only required environment variables
	_ = os.Setenv("APP_APP_NAME", "env-only-app")
	_ = os.Setenv("APP_APP_VERSION", "3.0.0")
	_ = os.Setenv("APP_APP_ENVIRONMENT", "production")
	_ = os.Setenv("APP_SERVER_PORT", "7777")
	_ = os.Setenv("APP_LOG_LEVEL", "error")
	_ = os.Setenv("APP_JWT_SECRET", "env-jwt-secret")

	cfg, err := LoadFromEnv()

	require.NoError(t, err)
	assert.Equal(t, "env-only-app", cfg.App.Name)
	assert.Equal(t, "3.0.0", cfg.App.Version)
	assert.Equal(t, "production", cfg.App.Environment)
	assert.Equal(t, 7777, cfg.Server.Port)
	assert.Equal(t, "error", cfg.Log.Level)
	assert.Equal(t, "env-jwt-secret", cfg.JWT.Secret)
}

func TestLoadFromEnv_MissingRequiredJWTSecret(t *testing.T) {
	// Clean up all env vars after test
	defer func() {
		envVars := []string{
			"APP_APP_NAME", "APP_APP_VERSION", "APP_APP_ENVIRONMENT",
			"APP_SERVER_PORT", "APP_LOG_LEVEL", "APP_JWT_SECRET",
		}
		for _, v := range envVars {
			_ = os.Unsetenv(v)
		}
	}()

	// Set environment variables but NOT JWT_SECRET (rely on default)
	_ = os.Setenv("APP_APP_NAME", "test-app")
	_ = os.Setenv("APP_APP_VERSION", "1.0.0")
	_ = os.Setenv("APP_APP_ENVIRONMENT", "development")
	_ = os.Setenv("APP_SERVER_PORT", "8080")
	_ = os.Setenv("APP_LOG_LEVEL", "info")
	// Do NOT set APP_JWT_SECRET - should use default

	cfg, err := LoadFromEnv()

	require.NoError(t, err)
	assert.Equal(t, "your-secret-key-change-in-production", cfg.JWT.Secret)
}

func TestLoad_Hierarchy_EnvOverridesYAML(t *testing.T) {
	// Create a temporary YAML config file
	tempDir := t.TempDir()
	configContent := `
app:
  name: "yaml-app"
  version: "1.0.0"
  environment: "development"
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: "30s"
  write_timeout: "30s"
  shutdown_timeout: "10s"
log:
  level: "info"
  format: "json"
jwt:
  secret: "yaml-secret"
  access_token_ttl: "15m"
  refresh_token_ttl: "168h"
`
	configPath := filepath.Join(tempDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Clean up env vars after test
	defer func() {
		_ = os.Unsetenv("APP_APP_NAME")
		_ = os.Unsetenv("APP_SERVER_PORT")
		_ = os.Unsetenv("APP_JWT_SECRET")
	}()

	// Set environment variables to override YAML values
	_ = os.Setenv("APP_APP_NAME", "env-override-app")
	_ = os.Setenv("APP_SERVER_PORT", "9999")
	_ = os.Setenv("APP_JWT_SECRET", "env-override-secret")

	// Load with specific config path
	cfg, err := LoadWithPath(configPath)

	require.NoError(t, err)
	// Environment variables should override YAML values
	assert.Equal(t, "env-override-app", cfg.App.Name)
	assert.Equal(t, 9999, cfg.Server.Port)
	assert.Equal(t, "env-override-secret", cfg.JWT.Secret)
	// Other values should come from YAML
	assert.Equal(t, "1.0.0", cfg.App.Version)
	assert.Equal(t, "json", cfg.Log.Format)
}

func TestLoad_Hierarchy_EnvOverridesEnvFile(t *testing.T) {
	// Create a temporary .env file
	tempDir := t.TempDir()
	_ = os.Chdir(tempDir)

	envContent := `APP_APP_NAME=envfile-app
APP_SERVER_PORT=7070
APP_LOG_LEVEL=debug
APP_JWT_SECRET=envfile-secret
`
	envPath := filepath.Join(tempDir, ".env")
	err := os.WriteFile(envPath, []byte(envContent), 0644)
	require.NoError(t, err)

	// Clean up env vars after test
	defer func() {
		_ = os.Unsetenv("APP_APP_NAME")
		_ = os.Unsetenv("APP_SERVER_PORT")
		_ = os.Unsetenv("APP_JWT_SECRET")
		_ = os.Chdir("/")
	}()

	// Set runtime env vars to override .env file values
	_ = os.Setenv("APP_APP_NAME", "runtime-env-app")
	_ = os.Setenv("APP_JWT_SECRET", "runtime-env-secret")

	// Load (should load .env file and then apply runtime env vars)
	cfg, err := Load()

	require.NoError(t, err)
	// Runtime env vars should override .env file values
	assert.Equal(t, "runtime-env-app", cfg.App.Name)
	assert.Equal(t, "runtime-env-secret", cfg.JWT.Secret)
	// Other values should come from .env file
	assert.Equal(t, 7070, cfg.Server.Port)
	assert.Equal(t, "debug", cfg.Log.Level)
}

func TestLoad_Defaults(t *testing.T) {
	// Clean up all env vars to test defaults (both before and after)
	envVars := []string{
		"APP_APP_NAME", "APP_APP_VERSION", "APP_APP_ENVIRONMENT",
		"APP_SERVER_HOST", "APP_SERVER_PORT", "APP_SERVER_READ_TIMEOUT",
		"APP_SERVER_WRITE_TIMEOUT", "APP_SERVER_SHUTDOWN_TIMEOUT",
		"APP_DATABASE_HOST", "APP_DATABASE_PORT", "APP_DATABASE_DATABASE",
		"APP_DATABASE_USERNAME", "APP_DATABASE_PASSWORD", "APP_DATABASE_SSL_MODE",
		"APP_LOG_LEVEL", "APP_LOG_FORMAT",
		"APP_JWT_SECRET", "APP_JWT_ACCESS_TOKEN_TTL", "APP_JWT_REFRESH_TOKEN_TTL",
	}

	// Clean up before test
	for _, v := range envVars {
		_ = os.Unsetenv(v)
	}

	// Clean up after test
	defer func() {
		for _, v := range envVars {
			_ = os.Unsetenv(v)
		}
	}()

	cfg, err := Load()

	require.NoError(t, err)
	// Verify all defaults are set
	assert.Equal(t, "zercle-go-template", cfg.App.Name)
	assert.Equal(t, "1.0.0", cfg.App.Version)
	assert.Equal(t, "development", cfg.App.Environment)
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, 30*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 30*time.Second, cfg.Server.WriteTimeout)
	assert.Equal(t, 10*time.Second, cfg.Server.ShutdownTimeout)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "zercle_template", cfg.Database.Database)
	assert.Equal(t, "postgres", cfg.Database.Username)
	assert.Equal(t, "", cfg.Database.Password)
	assert.Equal(t, "disable", cfg.Database.SSLMode)
	assert.Equal(t, "info", cfg.Log.Level)
	assert.Equal(t, "json", cfg.Log.Format)
	assert.Equal(t, "your-secret-key-change-in-production", cfg.JWT.Secret)
	assert.Equal(t, 15*time.Minute, cfg.JWT.AccessTokenTTL)
	assert.Equal(t, 168*time.Hour, cfg.JWT.RefreshTokenTTL)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &Config{
				App: AppConfig{
					Name:        "test-app",
					Environment: "development",
				},
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
				},
				Log: LogConfig{
					Level: "info",
				},
				JWT: JWTConfig{
					Secret:          "test-secret",
					AccessTokenTTL:  time.Hour,
					RefreshTokenTTL: 168 * time.Hour,
				},
			},
			wantErr: false,
		},
		{
			name: "missing app name",
			config: &Config{
				App: AppConfig{
					Name:        "",
					Environment: "development",
				},
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
				},
				Log: LogConfig{
					Level: "info",
				},
				JWT: JWTConfig{
					Secret:         "test-secret",
					AccessTokenTTL: time.Hour,
				},
			},
			wantErr: true,
			errMsg:  "app name is required",
		},
		{
			name: "missing environment",
			config: &Config{
				App: AppConfig{
					Name:        "test-app",
					Environment: "",
				},
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
				},
				Log: LogConfig{
					Level: "info",
				},
				JWT: JWTConfig{
					Secret:         "test-secret",
					AccessTokenTTL: time.Hour,
				},
			},
			wantErr: true,
			errMsg:  "app environment is required",
		},
		{
			name: "invalid port - negative",
			config: &Config{
				App: AppConfig{
					Name:        "test-app",
					Environment: "development",
				},
				Server: ServerConfig{
					Port:         -1,
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
				},
				Log: LogConfig{
					Level: "info",
				},
				JWT: JWTConfig{
					Secret:         "test-secret",
					AccessTokenTTL: time.Hour,
				},
			},
			wantErr: true,
			errMsg:  "invalid server port",
		},
		{
			name: "invalid port - zero",
			config: &Config{
				App: AppConfig{
					Name:        "test-app",
					Environment: "development",
				},
				Server: ServerConfig{
					Port:         0,
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
				},
				Log: LogConfig{
					Level: "info",
				},
				JWT: JWTConfig{
					Secret:         "test-secret",
					AccessTokenTTL: time.Hour,
				},
			},
			wantErr: true,
			errMsg:  "invalid server port",
		},
		{
			name: "invalid read timeout",
			config: &Config{
				App: AppConfig{
					Name:        "test-app",
					Environment: "development",
				},
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  0,
					WriteTimeout: 30 * time.Second,
				},
				Log: LogConfig{
					Level: "info",
				},
				JWT: JWTConfig{
					Secret:         "test-secret",
					AccessTokenTTL: time.Hour,
				},
			},
			wantErr: true,
			errMsg:  "invalid server read timeout",
		},
		{
			name: "invalid write timeout",
			config: &Config{
				App: AppConfig{
					Name:        "test-app",
					Environment: "development",
				},
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 0,
				},
				Log: LogConfig{
					Level: "info",
				},
				JWT: JWTConfig{
					Secret:         "test-secret",
					AccessTokenTTL: time.Hour,
				},
			},
			wantErr: true,
			errMsg:  "invalid server write timeout",
		},
		{
			name: "missing log level",
			config: &Config{
				App: AppConfig{
					Name:        "test-app",
					Environment: "development",
				},
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
				},
				Log: LogConfig{
					Level: "",
				},
				JWT: JWTConfig{
					Secret:         "test-secret",
					AccessTokenTTL: time.Hour,
				},
			},
			wantErr: true,
			errMsg:  "log level is required",
		},
		{
			name: "invalid log level",
			config: &Config{
				App: AppConfig{
					Name:        "test-app",
					Environment: "development",
				},
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
				},
				Log: LogConfig{
					Level: "trace",
				},
				JWT: JWTConfig{
					Secret:         "test-secret",
					AccessTokenTTL: time.Hour,
				},
			},
			wantErr: true,
			errMsg:  "invalid log level",
		},
		{
			name: "missing JWT secret",
			config: &Config{
				App: AppConfig{
					Name:        "test-app",
					Environment: "development",
				},
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
				},
				Log: LogConfig{
					Level: "info",
				},
				JWT: JWTConfig{
					Secret:         "",
					AccessTokenTTL: time.Hour,
				},
			},
			wantErr: true,
			errMsg:  "JWT secret is required",
		},
		{
			name: "invalid JWT access token TTL",
			config: &Config{
				App: AppConfig{
					Name:        "test-app",
					Environment: "development",
				},
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
				},
				Log: LogConfig{
					Level: "info",
				},
				JWT: JWTConfig{
					Secret:         "test-secret",
					AccessTokenTTL: 0,
				},
			},
			wantErr: true,
			errMsg:  "invalid JWT access token TTL",
		},
		{
			name: "invalid JWT refresh token TTL",
			config: &Config{
				App: AppConfig{
					Name:        "test-app",
					Environment: "development",
				},
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
				},
				Log: LogConfig{
					Level: "info",
				},
				JWT: JWTConfig{
					Secret:          "test-secret",
					AccessTokenTTL:  time.Hour,
					RefreshTokenTTL: 0,
				},
			},
			wantErr: true,
			errMsg:  "invalid JWT refresh token TTL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestConfig_IsDevelopment(t *testing.T) {
	tests := []struct {
		name     string
		env      string
		expected bool
	}{
		{
			name:     "development environment",
			env:      "development",
			expected: true,
		},
		{
			name:     "production environment",
			env:      "production",
			expected: false,
		},
		{
			name:     "staging environment",
			env:      "staging",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				App: AppConfig{
					Environment: tt.env,
				},
			}
			assert.Equal(t, tt.expected, cfg.IsDevelopment())
		})
	}
}

func TestConfig_IsProduction(t *testing.T) {
	tests := []struct {
		name     string
		env      string
		expected bool
	}{
		{
			name:     "production environment",
			env:      "production",
			expected: true,
		},
		{
			name:     "development environment",
			env:      "development",
			expected: false,
		},
		{
			name:     "staging environment",
			env:      "staging",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				App: AppConfig{
					Environment: tt.env,
				},
			}
			assert.Equal(t, tt.expected, cfg.IsProduction())
		})
	}
}

func TestConfig_DatabaseDSN(t *testing.T) {
	cfg := &Config{
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "testdb",
			Username: "testuser",
			Password: "testpass",
			SSLMode:  "disable",
		},
	}

	expected := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
	assert.Equal(t, expected, cfg.DatabaseDSN())
}

func TestConfig_DatabaseDSN_DifferentValues(t *testing.T) {
	cfg := &Config{
		Database: DatabaseConfig{
			Host:     "db.example.com",
			Port:     3306,
			Database: "myapp",
			Username: "admin",
			Password: "secret123",
			SSLMode:  "require",
		},
	}

	expected := "host=db.example.com port=3306 user=admin password=secret123 dbname=myapp sslmode=require"
	assert.Equal(t, expected, cfg.DatabaseDSN())
}

// =============================================================================
// Helper function tests
// =============================================================================

func TestGetEnvString(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		value        string
		defaultValue string
		expected     string
	}{
		{
			name:         "env var set",
			key:          "TEST_STRING",
			value:        "test-value",
			defaultValue: "default",
			expected:     "test-value",
		},
		{
			name:         "env var not set",
			key:          "TEST_STRING_MISSING",
			value:        "",
			defaultValue: "default",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				_ = os.Setenv(tt.key, tt.value)
				defer func() { _ = os.Unsetenv(tt.key) }()
			}

			result := GetEnvString(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		value        string
		defaultValue int
		expected     int
	}{
		{
			name:         "env var set with valid int",
			key:          "TEST_INT",
			value:        "42",
			defaultValue: 0,
			expected:     42,
		},
		{
			name:         "env var set with invalid int",
			key:          "TEST_INT_INVALID",
			value:        "not-a-number",
			defaultValue: 10,
			expected:     10,
		},
		{
			name:         "env var not set",
			key:          "TEST_INT_MISSING",
			value:        "",
			defaultValue: 5,
			expected:     5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				_ = os.Setenv(tt.key, tt.value)
				defer func() { _ = os.Unsetenv(tt.key) }()
			}

			result := GetEnvInt(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		value        string
		defaultValue bool
		expected     bool
	}{
		{
			name:         "env var set - true",
			key:          "TEST_BOOL",
			value:        "true",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "env var set - 1",
			key:          "TEST_BOOL",
			value:        "1",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "env var set - yes",
			key:          "TEST_BOOL",
			value:        "yes",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "env var set - on",
			key:          "TEST_BOOL",
			value:        "on",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "env var set - false",
			key:          "TEST_BOOL",
			value:        "false",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "env var set - 0",
			key:          "TEST_BOOL",
			value:        "0",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "env var set - no",
			key:          "TEST_BOOL",
			value:        "no",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "env var set - off",
			key:          "TEST_BOOL",
			value:        "off",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "env var set - invalid value",
			key:          "TEST_BOOL",
			value:        "invalid",
			defaultValue: true,
			expected:     true,
		},
		{
			name:         "env var not set",
			key:          "TEST_BOOL_MISSING",
			value:        "",
			defaultValue: true,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				_ = os.Setenv(tt.key, tt.value)
				defer func() { _ = os.Unsetenv(tt.key) }()
			}

			result := GetEnvBool(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvDuration(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		value        string
		defaultValue time.Duration
		expected     time.Duration
	}{
		{
			name:         "env var set - seconds",
			key:          "TEST_DURATION",
			value:        "30s",
			defaultValue: 0,
			expected:     30 * time.Second,
		},
		{
			name:         "env var set - minutes",
			key:          "TEST_DURATION",
			value:        "5m",
			defaultValue: 0,
			expected:     5 * time.Minute,
		},
		{
			name:         "env var set - hours",
			key:          "TEST_DURATION",
			value:        "2h",
			defaultValue: 0,
			expected:     2 * time.Hour,
		},
		{
			name:         "env var set - invalid format",
			key:          "TEST_DURATION",
			value:        "invalid",
			defaultValue: time.Minute,
			expected:     time.Minute,
		},
		{
			name:         "env var not set",
			key:          "TEST_DURATION_MISSING",
			value:        "",
			defaultValue: time.Hour,
			expected:     time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				_ = os.Setenv(tt.key, tt.value)
				defer func() { _ = os.Unsetenv(tt.key) }()
			}

			result := GetEnvDuration(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}
