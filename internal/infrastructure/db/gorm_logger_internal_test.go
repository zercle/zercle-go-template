//go:build unit

package db

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/zercle/zercle-go-template/internal/config"
)

func TestNewGORMLogger_LevelMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		logLevel string
		want     logger.LogLevel
	}{
		{"error", "error", logger.Error},
		{"warn", "warn", logger.Warn},
		{"info", "info", logger.Info},
		{"debug", "debug", logger.Info},
		{"trace", "trace", logger.Info},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{Log: config.LogConfig{Level: tt.logLevel}}
			log := zerolog.Nop()

			gl := newGORMLogger(&log, cfg)
			if gl.level != tt.want {
				t.Errorf("level = %v, want %v", gl.level, tt.want)
			}
		})
	}
}

func TestNewGORMLogger_NilLogger(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{Log: config.LogConfig{Level: "info"}}
	gl := newGORMLogger(nil, cfg)

	if gl.log == nil {
		t.Error("expected a non-nil logger (should fall back to nop)")
	}

	// Should not panic — verify by calling a method.
	gl.Info(context.Background(), "test")
}

func TestNewGORMLogger_NilConfig(t *testing.T) {
	t.Parallel()

	log := zerolog.Nop()
	gl := newGORMLogger(&log, nil)

	// With nil config, level defaults to Info.
	if gl.level != logger.Info {
		t.Errorf("level = %v, want Info (default)", gl.level)
	}

	if gl.slowThreshold != defaultSlowThreshold {
		t.Errorf("slowThreshold = %v, want %v", gl.slowThreshold, defaultSlowThreshold)
	}
}

func TestGORMLogger_TraceWithError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := zerolog.New(&buf).Level(zerolog.ErrorLevel)

	cfg := &config.Config{Log: config.LogConfig{Level: "error"}}
	gl := newGORMLogger(&log, cfg)

	fc := func() (string, int64) { return "SELECT 1", 0 }
	begin := time.Now()

	gl.Trace(context.Background(), begin, fc, errors.New("connection refused"))

	out := buf.String()
	if len(out) == 0 {
		t.Fatal("expected error log output")
	}

	if !bytes.Contains(buf.Bytes(), []byte(`"level":"error"`)) {
		t.Errorf("expected error level in output, got: %s", out)
	}

	if !bytes.Contains(buf.Bytes(), []byte(`"module":"gorm"`)) {
		t.Errorf("expected gorm module in output, got: %s", out)
	}
}

func TestGORMLogger_TraceSlowQuery(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := zerolog.New(&buf).Level(zerolog.WarnLevel)

	cfg := &config.Config{Log: config.LogConfig{Level: "warn"}}
	gl := newGORMLogger(&log, cfg)

	// Override slowThreshold to 1ms for fast testing.
	gl.slowThreshold = 1 * time.Millisecond

	fc := func() (string, int64) { return "SELECT 1", 0 }
	begin := time.Now().Add(-2 * time.Millisecond)

	gl.Trace(context.Background(), begin, fc, nil)

	out := buf.String()
	if len(out) == 0 {
		t.Fatal("expected slow query warning output")
	}

	if !bytes.Contains(buf.Bytes(), []byte(`"level":"warn"`)) {
		t.Errorf("expected warn level in output, got: %s", out)
	}

	if !bytes.Contains(buf.Bytes(), []byte("slow query")) {
		t.Errorf("expected 'slow query' in output, got: %s", out)
	}
}

func TestGORMLogger_TraceSilentLevel(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := zerolog.New(&buf).Level(zerolog.ErrorLevel)

	cfg := &config.Config{Log: config.LogConfig{Level: "error"}}
	gl := newGORMLogger(&log, cfg)

	// Set level to Silent — Trace should be a no-op.
	gl2, ok := gl.LogMode(logger.Silent).(*gormLogger)
	if !ok {
		t.Fatal("LogMode should return *gormLogger")
	}

	fc := func() (string, int64) { return "SELECT 1", 0 }
	begin := time.Now()

	gl2.Trace(context.Background(), begin, fc, nil)

	if len(buf.String()) != 0 {
		t.Errorf("expected no output at Silent level, got: %s", buf.String())
	}
}

func TestGORMLogger_TraceRecordNotFoundIgnored(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := zerolog.New(&buf).Level(zerolog.WarnLevel)

	cfg := &config.Config{Log: config.LogConfig{Level: "warn"}}
	gl := newGORMLogger(&log, cfg)

	fc := func() (string, int64) { return "SELECT 1", 0 }
	begin := time.Now()

	gl.Trace(context.Background(), begin, fc, gorm.ErrRecordNotFound)

	// With ignoreRecordNotFoundError=true (default), no output expected.
	if len(buf.String()) != 0 {
		t.Errorf("expected no output for ignored ErrRecordNotFound, got: %s", buf.String())
	}
}

func TestGORMLogger_TraceDebugSQL(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := zerolog.New(&buf).Level(zerolog.DebugLevel)

	cfg := &config.Config{Log: config.LogConfig{Level: "debug"}}
	gl := newGORMLogger(&log, cfg)

	fc := func() (string, int64) { return "SELECT 1", 0 }
	begin := time.Now()

	gl.Trace(context.Background(), begin, fc, nil)

	out := buf.String()
	if len(out) == 0 {
		t.Fatal("expected debug SQL output")
	}

	if !bytes.Contains(buf.Bytes(), []byte(`"level":"debug"`)) {
		t.Errorf("expected debug level in output, got: %s", out)
	}

	if !bytes.Contains(buf.Bytes(), []byte("SELECT 1")) {
		t.Errorf("expected SQL in output, got: %s", out)
	}
}

func TestGORMLogger_Info(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := zerolog.New(&buf).Level(zerolog.InfoLevel)

	cfg := &config.Config{Log: config.LogConfig{Level: "info"}}
	gl := newGORMLogger(&log, cfg)

	gl.Info(context.Background(), "test message")

	if len(buf.String()) == 0 {
		t.Fatal("expected info log output")
	}

	if !bytes.Contains(buf.Bytes(), []byte(`"level":"info"`)) {
		t.Errorf("expected info level in output, got: %s", buf.String())
	}

	if !bytes.Contains(buf.Bytes(), []byte("test message")) {
		t.Errorf("expected message in output, got: %s", buf.String())
	}
}

func TestGORMLogger_InfoLevelGuard(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := zerolog.New(&buf).Level(zerolog.ErrorLevel)

	cfg := &config.Config{Log: config.LogConfig{Level: "error"}}
	gl := newGORMLogger(&log, cfg)

	// Info should not log when level is Error.
	gl.Info(context.Background(), "should not appear")

	if len(buf.String()) != 0 {
		t.Errorf("expected no output at Error level, got: %s", buf.String())
	}
}

func TestGORMLogger_Warn(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := zerolog.New(&buf).Level(zerolog.WarnLevel)

	cfg := &config.Config{Log: config.LogConfig{Level: "warn"}}
	gl := newGORMLogger(&log, cfg)

	gl.Warn(context.Background(), "test warning")

	if len(buf.String()) == 0 {
		t.Fatal("expected warn log output")
	}

	if !bytes.Contains(buf.Bytes(), []byte(`"level":"warn"`)) {
		t.Errorf("expected warn level in output, got: %s", buf.String())
	}
}

func TestGORMLogger_Error(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := zerolog.New(&buf).Level(zerolog.ErrorLevel)

	cfg := &config.Config{Log: config.LogConfig{Level: "error"}}
	gl := newGORMLogger(&log, cfg)

	gl.Error(context.Background(), "test error")

	if len(buf.String()) == 0 {
		t.Fatal("expected error log output")
	}

	if !bytes.Contains(buf.Bytes(), []byte(`"level":"error"`)) {
		t.Errorf("expected error level in output, got: %s", buf.String())
	}
}

func TestGORMLogger_LogMode(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := zerolog.New(&buf).Level(zerolog.WarnLevel)

	cfg := &config.Config{Log: config.LogConfig{Level: "warn"}}
	gl := newGORMLogger(&log, cfg)

	// LogMode should return a copy with the new level.
	gl2, ok := gl.LogMode(logger.Warn).(*gormLogger)
	if !ok {
		t.Fatal("LogMode should return *gormLogger")
	}

	// Original should still be Warn level.
	if gl.level != logger.Warn {
		t.Errorf("original level = %v, want Warn", gl.level)
	}

	// Copy should be Warn level.
	if gl2.level != logger.Warn {
		t.Errorf("copy level = %v, want Warn", gl2.level)
	}

	// The copy should log warnings.
	gl2.Warn(context.Background(), "warning from copy")

	if len(buf.String()) == 0 {
		t.Fatal("expected warn output from copy")
	}

	if !bytes.Contains(buf.Bytes(), []byte(`"level":"warn"`)) {
		t.Errorf("expected warn level in output, got: %s", buf.String())
	}
}
