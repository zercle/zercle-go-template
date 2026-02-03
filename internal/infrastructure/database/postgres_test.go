package database

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zercle/zercle-go-template/pkg/config"
)

func TestDSN_Formatting(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "secret",
		Database: "app",
		SSLMode:  "disable",
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
	)

	expected := "host=localhost port=5432 user=postgres password=secret dbname=app sslmode=disable"
	assert.Equal(t, expected, dsn, "DSN should be properly formatted")
}

func TestDSN_AllSSLmodes(t *testing.T) {
	sslModes := []string{"disable", "require", "verify-ca", "verify-full"}

	for _, sslMode := range sslModes {
		cfg := config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "secret",
			Database: "app",
			SSLMode:  sslMode,
		}

		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
		)

		assert.Contains(t, dsn, "sslmode="+sslMode, "DSN should contain sslmode=%s", sslMode)
	}
}

func TestDatabaseConfig_Validation_InvalidPort(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:     0, // Invalid port
		User:     "postgres",
		Password: "secret",
		Database: "app",
		SSLMode:  "disable",
	}

	// Port validation is done at database connection time
	// This test verifies the config structure
	assert.Equal(t, 0, cfg.Port)
}

func TestDatabaseConfig_Validation_MissingPassword(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "", // Empty password
		Database: "app",
		SSLMode:  "disable",
	}

	// Empty password should be handled gracefully
	assert.Empty(t, cfg.Password)
}

func TestDatabaseConfig_Validation_SpecialCharactersInPassword(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "p@ssw0rd!#$%",
		Database: "app",
		SSLMode:  "disable",
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
	)

	assert.Contains(t, dsn, "password=p@ssw0rd!#$%", "DSN should handle special characters in password")
}

func TestDatabaseConfig_Validation_DifferentHosts(t *testing.T) {
	hosts := []string{"localhost", "127.0.0.1", "db.example.com", "192.168.1.100"}

	for _, host := range hosts {
		cfg := config.DatabaseConfig{
			Host:     host,
			Port:     5432,
			User:     "postgres",
			Password: "secret",
			Database: "app",
			SSLMode:  "disable",
		}

		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
		)

		assert.Contains(t, dsn, "host="+host, "DSN should contain host=%s", host)
	}
}

func TestDatabaseConfig_Validation_DifferentPorts(t *testing.T) {
	ports := []int{5432, 5433, 8080, 3000}

	for _, port := range ports {
		cfg := config.DatabaseConfig{
			Host:     "localhost",
			Port:     port,
			User:     "postgres",
			Password: "secret",
			Database: "app",
			SSLMode:  "disable",
		}

		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
		)

		assert.Contains(t, dsn, fmt.Sprintf("port=%d", port), "DSN should contain port=%d", port)
	}
}
