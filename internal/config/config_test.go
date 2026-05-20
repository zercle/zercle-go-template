package config

import (
	"os"
	"testing"
	"time"
)

func TestConfig_Validate_RequiresDBName(t *testing.T) {
	t.Parallel()
	cfg := Config{
		DBUser:                 "user",
		AuthAccessTokenSecret:  "access-secret",
		AuthRefreshTokenSecret: "refresh-secret",
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for missing DB_NAME, got nil")
	}
}

func TestConfig_Validate_RequiresDBUser(t *testing.T) {
	t.Parallel()
	cfg := Config{
		DBName:                 "testdb",
		AuthAccessTokenSecret:  "access-secret",
		AuthRefreshTokenSecret: "refresh-secret",
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for missing DB_USER, got nil")
	}
}

func TestConfig_Validate_RequiresAccessTokenSecret(t *testing.T) {
	t.Parallel()
	cfg := Config{
		DBName:                 "testdb",
		DBUser:                 "user",
		AuthRefreshTokenSecret: "refresh-secret",
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for missing AUTH_ACCESS_TOKEN_SECRET, got nil")
	}
}

func TestConfig_Validate_RequiresRefreshTokenSecret(t *testing.T) {
	t.Parallel()
	cfg := Config{
		DBName:                "testdb",
		DBUser:                "user",
		AuthAccessTokenSecret: "access-secret",
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for missing AUTH_REFRESH_TOKEN_SECRET, got nil")
	}
}

func TestConfig_Validate_AllFieldsPresent(t *testing.T) {
	t.Parallel()
	cfg := Config{
		DBName:                 "testdb",
		DBUser:                 "user",
		AuthAccessTokenSecret:  "access-secret",
		AuthRefreshTokenSecret: "refresh-secret",
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("expected nil error for valid config, got %v", err)
	}
}

func TestConfig_ServerAddr(t *testing.T) {
	t.Parallel()
	cfg := Config{
		ServerHost: "localhost",
		ServerPort: 8080,
	}
	if got := cfg.ServerAddr(); got != "localhost:8080" {
		t.Errorf("expected localhost:8080, got %s", got)
	}
}

func TestConfig_GRPCAddr(t *testing.T) {
	t.Parallel()
	cfg := Config{
		GRPCHost: "0.0.0.0",
		GRPCPort: 9090,
	}
	if got := cfg.GRPCAddr(); got != "0.0.0.0:9090" {
		t.Errorf("expected 0.0.0.0:9090, got %s", got)
	}
}

func TestConfig_DBConnString(t *testing.T) {
	t.Parallel()
	cfg := Config{
		DBUser:     "user",
		DBPassword: "password",
		DBHost:     "localhost",
		DBPort:     5432,
		DBName:     "testdb",
		DBSSLMode:  "disable",
	}
	expected := "postgres://user:password@localhost:5432/testdb?sslmode=disable" //nolint
	if got := cfg.DBConnString(); got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestConfig_CacheAddr(t *testing.T) {
	t.Parallel()
	cfg := Config{
		CacheHost: "localhost",
		CachePort: 6379,
	}
	if got := cfg.CacheAddr(); got != "localhost:6379" {
		t.Errorf("expected localhost:6379, got %s", got)
	}
}

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer os.Chdir(cwd)

	t.Setenv("DB_NAME", "testdb")
	t.Setenv("DB_USER", "testuser")
	t.Setenv("AUTH_ACCESS_TOKEN_SECRET", "access")
	t.Setenv("AUTH_REFRESH_TOKEN_SECRET", "refresh")

	cfg := Load()

	if cfg.DBName != "testdb" {
		t.Errorf("expected DBName=testdb, got %s", cfg.DBName)
	}
	if cfg.DBUser != "testuser" {
		t.Errorf("expected DBUser=testuser, got %s", cfg.DBUser)
	}
	if cfg.AuthAccessTokenSecret != "access" {
		t.Errorf("expected AuthAccessTokenSecret=access, got %s", cfg.AuthAccessTokenSecret)
	}
	if cfg.AuthRefreshTokenSecret != "refresh" {
		t.Errorf("expected AuthRefreshTokenSecret=refresh, got %s", cfg.AuthRefreshTokenSecret)
	}
	if cfg.ServerPort != 8080 {
		t.Errorf("expected default ServerPort=8080, got %d", cfg.ServerPort)
	}
	if cfg.AuthAccessTokenTTL != 24*time.Hour {
		t.Errorf("expected default AuthAccessTokenTTL=24h, got %v", cfg.AuthAccessTokenTTL)
	}
}
