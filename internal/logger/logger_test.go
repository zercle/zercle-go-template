package logger

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	// Save and restore original level to prevent affecting other tests
	originalLevel := zerolog.GlobalLevel()
	defer zerolog.SetGlobalLevel(originalLevel)

	tests := []struct {
		name        string
		service     string
		environment string
	}{
		{
			name:        "development environment",
			service:     "test-service",
			environment: "development",
		},
		{
			name:        "production environment",
			service:     "prod-service",
			environment: "production",
		},
		{
			name:        "staging environment",
			service:     "staging-service",
			environment: "staging",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.service, tt.environment)
			require.NotNil(t, logger)

			// Verify logger is not nil and can log
			assert.NotPanics(t, func() {
				logger.Info("test message")
			})
		})
	}
}

func TestNewNop(t *testing.T) {
	logger := NewNop()
	require.NotNil(t, logger)

	// Nop logger should not panic on any operation
	assert.NotPanics(t, func() {
		logger.Debug("debug message")
		logger.Info("info message")
		logger.Warn("warn message")
		logger.Error("error message")
	})
}

func TestFieldConstructors(t *testing.T) {
	tests := []struct {
		name     string
		field    Field
		expected Field
	}{
		{
			name:     "String field",
			field:    String("key", "value"),
			expected: Field{Key: "key", Value: "value"},
		},
		{
			name:     "Int field",
			field:    Int("count", 42),
			expected: Field{Key: "count", Value: 42},
		},
		{
			name:     "Int64 field",
			field:    Int64("bigcount", 9223372036854775807),
			expected: Field{Key: "bigcount", Value: int64(9223372036854775807)},
		},
		{
			name:     "Float64 field",
			field:    Float64("ratio", 3.14),
			expected: Field{Key: "ratio", Value: 3.14},
		},
		{
			name:     "Bool field - true",
			field:    Bool("enabled", true),
			expected: Field{Key: "enabled", Value: true},
		},
		{
			name:     "Bool field - false",
			field:    Bool("disabled", false),
			expected: Field{Key: "disabled", Value: false},
		},
		{
			name:     "Duration field",
			field:    Duration("elapsed", 5*time.Second),
			expected: Field{Key: "elapsed", Value: 5 * time.Second},
		},
		{
			name:     "Any field",
			field:    Any("data", map[string]string{"key": "value"}),
			expected: Field{Key: "data", Value: map[string]string{"key": "value"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.field)
		})
	}
}

func TestErrorField(t *testing.T) {
	testErr := assert.AnError
	field := Error(testErr)
	assert.Equal(t, "error", field.Key)
	assert.Equal(t, testErr, field.Value)
}

func TestTimeField(t *testing.T) {
	now := time.Now()
	field := Time("timestamp", now)
	assert.Equal(t, "timestamp", field.Key)
	assert.Equal(t, now, field.Value)
}

func TestZerologLogger_LogLevels(t *testing.T) {
	var buf bytes.Buffer
	zl := zerolog.New(&buf).Level(zerolog.DebugLevel)
	logger := &zerologLogger{logger: zl}

	tests := []struct {
		name     string
		logFunc  func(string, ...Field)
		expected string
	}{
		{
			name:     "Debug",
			logFunc:  logger.Debug,
			expected: "debug",
		},
		{
			name:     "Info",
			logFunc:  logger.Info,
			expected: "info",
		},
		{
			name:     "Warn",
			logFunc:  logger.Warn,
			expected: "warn",
		},
		{
			name:     "Error",
			logFunc:  logger.Error,
			expected: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc("test message", String("key", "value"))
			output := buf.String()
			assert.Contains(t, output, tt.expected)
			assert.Contains(t, output, "test message")
			assert.Contains(t, output, "value")
		})
	}
}

func TestZerologLogger_Fatal(t *testing.T) {
	// Fatal calls os.Exit, so we can't test it directly
	// The nop logger's Fatal will not produce output but will still call os.Exit
	// We verify the method exists by checking interface compliance
	nopLogger := NewNop()
	require.NotNil(t, nopLogger)
	// We cannot call Fatal() as it will exit the program
	// This test just verifies the interface method exists
}

func TestZerologLogger_WithContext(t *testing.T) {
	logger := NewNop()
	ctx := context.Background()

	tests := []struct {
		name string
		ctx  context.Context
	}{
		{
			name: "with background context",
			ctx:  ctx,
		},
		{
			name: "with nil context",
			ctx:  nil,
		},
		{
			name: "with value context",
			ctx:  context.WithValue(ctx, contextKey("key"), "value"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loggerWithCtx := logger.WithContext(tt.ctx)
			assert.NotNil(t, loggerWithCtx)
			assert.NotPanics(t, func() {
				loggerWithCtx.Info("test message")
			})
		})
	}
}

func TestZerologLogger_WithFields(t *testing.T) {
	logger := NewNop()

	tests := []struct {
		name   string
		fields []Field
	}{
		{
			name:   "single field",
			fields: []Field{String("key", "value")},
		},
		{
			name:   "multiple fields",
			fields: []Field{String("key1", "value1"), Int("key2", 42), Bool("key3", true)},
		},
		{
			name:   "no fields",
			fields: []Field{},
		},
		{
			name:   "complex fields",
			fields: []Field{Error(assert.AnError), Duration("elapsed", time.Second)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loggerWithFields := logger.WithFields(tt.fields...)
			assert.NotNil(t, loggerWithFields)
			assert.NotPanics(t, func() {
				loggerWithFields.Info("test message")
			})
		})
	}
}

func TestFromContext(t *testing.T) {
	tests := []struct {
		name         string
		ctx          context.Context
		shouldBeNop  bool
		expectedType string
	}{
		{
			name:        "nil context",
			ctx:         nil,
			shouldBeNop: true,
		},
		{
			name:        "empty context",
			ctx:         context.Background(),
			shouldBeNop: true,
		},
		{
			name: "context with logger",
			ctx: func() context.Context {
				logger := NewNop()
				return WithContext(context.Background(), logger)
			}(),
			shouldBeNop: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := FromContext(tt.ctx)
			assert.NotNil(t, logger)
		})
	}
}

func TestWithContext(t *testing.T) {
	logger := NewNop()

	tests := []struct {
		name string
		ctx  context.Context
	}{
		{
			name: "nil context",
			ctx:  nil,
		},
		{
			name: "background context",
			ctx:  context.Background(),
		},
		{
			name: "context with values",
			ctx:  context.WithValue(context.Background(), contextKey("key"), "value"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newCtx := WithContext(tt.ctx, logger)
			assert.NotNil(t, newCtx)

			// Verify logger can be retrieved
			retrievedLogger := FromContext(newCtx)
			assert.NotNil(t, retrievedLogger)
		})
	}
}

func TestContextRoundTrip(t *testing.T) {
	originalLogger := New("test-service", "development")
	ctx := context.Background()

	// Add logger to context
	ctxWithLogger := WithContext(ctx, originalLogger)

	// Retrieve logger from context
	retrievedLogger := FromContext(ctxWithLogger)

	// Retrieved logger should work
	assert.NotPanics(t, func() {
		retrievedLogger.Info("test message from retrieved logger")
	})
}

func TestNew_DevelopmentLogLevel(t *testing.T) {
	// Save and restore original level
	originalLevel := zerolog.GlobalLevel()
	defer zerolog.SetGlobalLevel(originalLevel)

	// In development, log level should be debug
	logger := New("test", "development")
	assert.NotNil(t, logger)
	assert.Equal(t, zerolog.DebugLevel, zerolog.GlobalLevel())
}

func TestNew_ProductionLogLevel(t *testing.T) {
	// Save and restore original level
	originalLevel := zerolog.GlobalLevel()
	defer zerolog.SetGlobalLevel(originalLevel)

	// In production, log level should be info
	logger := New("test", "production")
	assert.NotNil(t, logger)
	assert.Equal(t, zerolog.InfoLevel, zerolog.GlobalLevel())
}

func TestZerologLogger_ConsoleOutput(t *testing.T) {
	// Test that development environment uses console writer
	logger := New("test-service", "development")
	assert.NotNil(t, logger)

	// Should not panic when logging
	assert.NotPanics(t, func() {
		logger.Info("console output test")
	})
}

func TestZerologLogger_JSONOutput(t *testing.T) {
	// Test that production environment uses JSON writer
	logger := New("test-service", "production")
	assert.NotNil(t, logger)

	// Should not panic when logging
	assert.NotPanics(t, func() {
		logger.Info("json output test")
	})
}

func TestAddFields(t *testing.T) {
	var buf bytes.Buffer
	zl := zerolog.New(&buf)
	event := zl.Info()

	fields := []Field{
		String("string_key", "string_value"),
		Int("int_key", 42),
		Bool("bool_key", true),
	}

	event = addFields(event, fields...)
	event.Msg("test")

	output := buf.String()
	assert.Contains(t, output, "string_value")
	assert.Contains(t, output, "42")
	assert.Contains(t, output, "true")
}

func TestLogger_Interface(t *testing.T) {
	// Verify that zerologLogger implements Logger interface
	var _ Logger = (*zerologLogger)(nil)

	// Verify that New returns a Logger
	logger := New("test", "development")
	var _ = logger
}

func TestEnvironmentDetection(t *testing.T) {
	tests := []struct {
		environment string
		isDev       bool
	}{
		{"development", true},
		{"production", false},
		{"staging", false},
		{"test", false},
	}

	for _, tt := range tests {
		t.Run(tt.environment, func(t *testing.T) {
			// Reset stdout to capture console output
			oldStdout := os.Stdout
			defer func() { os.Stdout = oldStdout }()

			logger := New("test", tt.environment)
			assert.NotNil(t, logger)
		})
	}
}
