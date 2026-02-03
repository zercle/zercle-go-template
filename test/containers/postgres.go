package containers

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func init() {
	// Disable reaper for Podman compatibility
	if IsPodman() {
		_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	}
}

// PostgresContainer wraps the testcontainers postgres module
type PostgresContainer struct {
	*postgres.PostgresContainer
	ConnectionString string
}

// NewPostgresContainer creates a new PostgreSQL container with podman-first fallback to docker
func NewPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").
			WithStartupTimeout(120),
		SkipReaper: true,
	}

	// Create a new context with timeout for container startup
	startupCtx, cancel := context.WithTimeout(ctx, 180)
	defer cancel()

	// Create container using generic container request
	container, err := testcontainers.GenericContainer(startupCtx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Get connection details
	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get host: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		return nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	connStr := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", host, mappedPort.Port())

	return &PostgresContainer{
		PostgresContainer: &postgres.PostgresContainer{Container: container},
		ConnectionString:  connStr,
	}, nil
}

// MustNewPostgresContainer creates a new PostgreSQL container and panics on error
func MustNewPostgresContainer(ctx context.Context) *PostgresContainer {
	container, err := NewPostgresContainer(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to create postgres container: %v", err))
	}
	return container
}

// GetHost returns the container host
func (c *PostgresContainer) GetHost(ctx context.Context) (string, error) {
	return c.Host(ctx)
}

// GetMappedPort returns the mapped port for PostgreSQL
func (c *PostgresContainer) GetMappedPort(ctx context.Context) (int, error) {
	mappedPort, err := c.MappedPort(ctx, "5432/tcp")
	if err != nil {
		return 0, err
	}
	portStr := mappedPort.Port()
	var port int
	if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
		return 0, fmt.Errorf("failed to parse port: %w", err)
	}
	return port, nil
}

// Terminate stops and removes the container
func (c *PostgresContainer) Terminate(ctx context.Context) error {
	return c.PostgresContainer.Terminate(ctx)
}

// IsPodman returns true if podman is available on the system
func IsPodman() bool {
	_, err := exec.LookPath("podman")
	return err == nil
}

// DetectProvider returns the available container provider (podman or docker)
func DetectProvider() testcontainers.ProviderType {
	if _, err := exec.LookPath("podman"); err == nil {
		return testcontainers.ProviderPodman
	}
	return testcontainers.ProviderDocker
}
