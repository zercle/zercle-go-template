package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"zercle-go-template/internal/config"
	"zercle-go-template/internal/logger"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			cfg: &config.Config{
				App: config.AppConfig{
					Name:        "test-app",
					Version:     "1.0.0",
					Environment: "test",
				},
				Log: config.LogConfig{
					Level: "info",
				},
			},
			wantErr: false,
		},
		{
			name: "development environment",
			cfg: &config.Config{
				App: config.AppConfig{
					Name:        "test-app",
					Version:     "1.0.0",
					Environment: "development",
				},
				Log: config.LogConfig{
					Level: "debug",
				},
			},
			wantErr: false,
		},
		{
			name: "production environment",
			cfg: &config.Config{
				App: config.AppConfig{
					Name:        "prod-app",
					Version:     "2.0.0",
					Environment: "production",
				},
				Log: config.LogConfig{
					Level: "info",
				},
			},
			wantErr: false,
		},
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: true,
			errMsg:  "configuration is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container, err := New(tt.cfg)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, container)

			// Verify all dependencies are initialized
			assert.NotNil(t, container.Config)
			assert.NotNil(t, container.Logger)
			assert.NotNil(t, container.UserRepo)
			assert.NotNil(t, container.UserUsecase)
			assert.NotNil(t, container.JWTUsecase)

			// Verify config is the same instance
			assert.Equal(t, tt.cfg, container.Config)

			// Clean up
			err = container.Close()
			assert.NoError(t, err)
		})
	}
}

func TestContainer_DependenciesInitialized(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Name:        "test-app",
			Version:     "1.0.0",
			Environment: "test",
		},
		Log: config.LogConfig{
			Level: "info",
		},
	}

	container, err := New(cfg)
	require.NoError(t, err)
	defer func() { _ = container.Close() }()

	// Test that logger works
	assert.NotPanics(t, func() {
		container.Logger.Info("test message")
	})

	// Test that repository is functional
	assert.NotNil(t, container.UserRepo)

	// Test that usecase is functional
	assert.NotNil(t, container.UserUsecase)
}

func TestContainer_Close(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Name:        "test-app",
			Version:     "1.0.0",
			Environment: "test",
		},
		Log: config.LogConfig{
			Level: "info",
		},
	}

	tests := []struct {
		name      string
		container *Container
		wantErr   bool
	}{
		{
			name: "close with logger",
			container: func() *Container {
				c, _ := New(cfg)
				return c
			}(),
			wantErr: false,
		},
		{
			name: "close with nil logger",
			container: &Container{
				Config: cfg,
				Logger: nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.container.Close()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestContainer_LoggerType(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Name:        "test-app",
			Version:     "1.0.0",
			Environment: "test",
		},
		Log: config.LogConfig{
			Level: "info",
		},
	}

	container, err := New(cfg)
	require.NoError(t, err)
	defer func() { _ = container.Close() }()

	// Verify logger implements the Logger interface
	var _ = container.Logger

	// Test all log levels
	assert.NotPanics(t, func() {
		container.Logger.Debug("debug message")
		container.Logger.Info("info message")
		container.Logger.Warn("warn message")
		container.Logger.Error("error message")
	})
}

func TestContainer_WithFields(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Name:        "test-app",
			Version:     "1.0.0",
			Environment: "test",
		},
		Log: config.LogConfig{
			Level: "info",
		},
	}

	container, err := New(cfg)
	require.NoError(t, err)
	defer func() { _ = container.Close() }()

	// Test WithFields
	loggerWithFields := container.Logger.WithFields(
		logger.String("key1", "value1"),
		logger.Int("key2", 42),
	)
	assert.NotNil(t, loggerWithFields)

	assert.NotPanics(t, func() {
		loggerWithFields.Info("message with fields")
	})
}

func TestNew_MultipleInstances(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Name:        "test-app",
			Version:     "1.0.0",
			Environment: "test",
		},
		Log: config.LogConfig{
			Level: "info",
		},
	}

	// Create multiple containers
	container1, err := New(cfg)
	require.NoError(t, err)
	defer func() { _ = container1.Close() }()

	container2, err := New(cfg)
	require.NoError(t, err)
	defer func() { _ = container2.Close() }()

	// They should be independent instances (different pointer addresses)
	assert.True(t, container1 != container2, "containers should be different instances")
	assert.True(t, container1.Logger != container2.Logger, "loggers should be different instances")
	assert.True(t, container1.UserRepo != container2.UserRepo, "repositories should be different instances")
	assert.True(t, container1.UserUsecase != container2.UserUsecase, "usecases should be different instances")
}
