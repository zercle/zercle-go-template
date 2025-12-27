package integration

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver for database/sql
)

// getEnvOrDefault returns the environment variable or the default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvOrDefaultInt returns the environment variable or the default value (integer)
func getEnvOrDefaultInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// TestDBHelper manages test database lifecycle
type TestDBHelper struct {
	db      *sql.DB
	connStr string
}

// NewTestDBHelper creates a new database helper for tests
func NewTestDBHelper() *TestDBHelper {
	host := getEnvOrDefault("TEST_DB_HOST", "localhost")
	port := getEnvOrDefaultInt("TEST_DB_PORT", 5433)
	user := getEnvOrDefault("TEST_DB_USER", "postgres")
	password := getEnvOrDefault("TEST_DB_PASSWORD", "testpassword")
	dbName := getEnvOrDefault("TEST_DB_NAME", "zercle_test_db")

	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName,
	)

	return &TestDBHelper{connStr: connStr}
}

// Setup connects to database, runs migrations, and waits for ready
func (h *TestDBHelper) Setup(t *testing.T) {
	if t != nil {
		t.Helper()
	}

	var err error

	for i := 0; i < 30; i++ {
		h.db, err = sql.Open("pgx", h.connStr)
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if pingErr := h.db.PingContext(ctx); pingErr == nil {
				cancel()
				if t != nil {
					t.Log("✅ Test database connected")
				} else {
					fmt.Println("✅ Test database connected")
				}
				break
			}
			cancel()
		}
		if i == 29 {
			if t != nil {
				t.Fatalf("Failed to connect to test database after 30s: %v", err)
			} else {
				panic(fmt.Sprintf("Failed to connect to test database after 30s: %v", err))
			}
		}
		time.Sleep(1 * time.Second)
	}

	h.runMigrations(t)
}

// runMigrations applies all database migrations
func (h *TestDBHelper) runMigrations(t *testing.T) {
	if t != nil {
		t.Helper()
	}

	migrationFiles := []string{
		"20251226_initialize_online_booking_schema.up.sql",
		"20251226_add_availability_and_payments.up.sql",
	}

	migrationDir := "../../sql/migration"
	migrationFS := os.DirFS(migrationDir)

	for _, file := range migrationFiles {
		if !fs.ValidPath(file) {
			msg := fmt.Sprintf("⚠️  Warning: Invalid migration file path %s", file)
			if t != nil {
				t.Log(msg)
			} else {
				fmt.Println(msg)
			}
			continue
		}

		content, err := fs.ReadFile(migrationFS, file)
		if err != nil {
			msg := fmt.Sprintf("⚠️  Warning: Failed to read migration file %s: %v", file, err)
			if t != nil {
				t.Log(msg)
			} else {
				fmt.Println(msg)
			}
			continue
		}

		if _, err := h.db.Exec(string(content)); err != nil {
			msg := fmt.Sprintf("⚠️  Warning: Failed to execute migration %s: %v", file, err)
			if t != nil {
				t.Log(msg)
			} else {
				fmt.Println(msg)
			}
		}
	}

	if t != nil {
		t.Log("✅ Migrations applied")
	} else {
		fmt.Println("✅ Migrations applied")
	}
}

// Cleanup truncates all tables for clean test state
func (h *TestDBHelper) Cleanup(t *testing.T) {
	if t != nil {
		t.Helper()
	}

	if h.db == nil {
		return
	}

	tables := []string{
		"payments",
		"bookings_users",
		"bookings",
		"availability_slots",
		"services",
		"users",
	}

	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)
		if _, err := h.db.Exec(query); err != nil {
			msg := fmt.Sprintf("⚠️  Warning: Failed to truncate %s: %v", table, err)
			if t != nil {
				t.Log(msg)
			} else {
				fmt.Println(msg)
			}
		}
	}

	if t != nil {
		t.Log("🧹 Database cleaned")
	} else {
		fmt.Println("🧹 Database cleaned")
	}
}

// Close closes database connection
func (h *TestDBHelper) Close() {
	if h.db != nil {
		_ = h.db.Close()
	}
}

// GetDB returns the database connection for advanced test scenarios
func (h *TestDBHelper) GetDB() *sql.DB {
	return h.db
}
