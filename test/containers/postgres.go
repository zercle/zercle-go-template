package containers

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer wraps the testcontainers postgres module
type PostgresContainer struct {
	*postgres.PostgresContainer
	ConnectionString string
}

// NewPostgresContainer creates a new PostgreSQL container with podman-first fallback to docker
func NewPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	// Create container using postgres module
	// The module handles provider detection automatically
	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	return &PostgresContainer{
		PostgresContainer: container,
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
	return mappedPort.Int(), nil
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
