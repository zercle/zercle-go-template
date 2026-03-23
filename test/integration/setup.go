package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zercle/zercle-go-template/test/containers"
)

// IntegrationSuite holds the test environment resources.
type IntegrationSuite struct {
	Container *containers.PostgresContainer
	DB        *pgxpool.Pool
	Host      string
	Port      int
}

// SetupSuite creates a new integration test suite with PostgreSQL container.
func SetupSuite(t *testing.T) *IntegrationSuite {
	ctx := context.Background()

	// Start PostgreSQL container using testcontainers
	container, err := containers.NewPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	// Create connection pool
	pool, err := pgxpool.New(ctx, container.ConnectionString)
	if err != nil {
		if termErr := container.Terminate(ctx); termErr != nil {
			t.Logf("failed to terminate container: %v", termErr)
		}
		t.Fatalf("failed to create connection pool: %v", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		if termErr := container.Terminate(ctx); termErr != nil {
			t.Logf("failed to terminate container: %v", termErr)
		}
		t.Fatalf("failed to ping database: %v", err)
	}

	// Run migrations
	if err := runMigrations(ctx, pool); err != nil {
		pool.Close()
		if termErr := container.Terminate(ctx); termErr != nil {
			t.Logf("failed to terminate container: %v", termErr)
		}
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Get host and port for debugging
	host, _ := container.GetHost(ctx)
	port, _ := container.GetMappedPort(ctx)

	return &IntegrationSuite{
		Container: container,
		DB:        pool,
		Host:      host,
		Port:      port,
	}
}

// Teardown tears down the test environment.
func (s *IntegrationSuite) Teardown(t *testing.T) {
	if s.DB != nil {
		s.DB.Close()
	}
	if s.Container != nil {
		if err := s.Container.Terminate(context.Background()); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}
}

// runMigrations executes database migrations for testing.
func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	// Create users table if not exists using sqlc schema
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			first_name VARCHAR(100) NOT NULL,
			last_name VARCHAR(100) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_users_status ON users(status)`,
		// Tasks table
		`CREATE TABLE IF NOT EXISTS tasks (
			id UUID PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			priority VARCHAR(20) NOT NULL DEFAULT 'medium',
			user_id UUID NOT NULL REFERENCES users(id),
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at)`,
	}

	for _, migration := range migrations {
		if _, err := pool.Exec(ctx, migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

// SetupWithRetry sets up the test environment with retry logic.
func SetupWithRetry(t *testing.T, maxRetries int) *IntegrationSuite {
	var suite *IntegrationSuite
	var err error

	for i := range maxRetries {
		suite = SetupSuite(t)
		if suite != nil && suite.DB != nil {
			return suite
		}
		if i < maxRetries-1 {
			t.Logf("Retrying setup (attempt %d/%d)", i+2, maxRetries)
			time.Sleep(time.Second)
		}
	}

	if err != nil {
		t.Fatalf("failed to setup test environment after %d retries: %v", maxRetries, err)
	}

	return suite
}

// SkipIfShort skips the test if -short flag is used.
func SkipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test (use -short=false to run)")
	}
}

// NewContext creates a context with timeout for tests.
func NewContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// SetupTestEnv is an alias for SetupSuite for backwards compatibility.
func SetupTestEnv(t *testing.T) *IntegrationSuite {
	return SetupSuite(t)
}

// Cleanup is an alias for Teardown for backwards compatibility.
func (s *IntegrationSuite) Cleanup(t *testing.T) {
	s.Teardown(t)
}
