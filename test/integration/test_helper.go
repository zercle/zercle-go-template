package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// ContainerRuntime represents the container runtime being used
type ContainerRuntime string

const (
	RuntimeDocker  ContainerRuntime = "docker"
	RuntimePodman  ContainerRuntime = "podman"
	RuntimeUnknown ContainerRuntime = "unknown"
)

// TestDBHelper manages test database lifecycle using testcontainers
type TestDBHelper struct {
	container *postgres.PostgresContainer
	db        *sql.DB
	ctx       context.Context
	runtime   ContainerRuntime
}

// getContainerRuntime determines which container runtime to use
// Priority: CONTAINER_RUNTIME env var → DOCKER_HOST detection → Podman → Docker
func getContainerRuntime() (ContainerRuntime, string) {
	// 1. Check explicit environment variable override
	if runtimeEnv := os.Getenv("CONTAINER_RUNTIME"); runtimeEnv != "" {
		runtime := ContainerRuntime(strings.ToLower(runtimeEnv))
		switch runtime {
		case RuntimePodman, RuntimeDocker:
			fmt.Printf("🎯 Container runtime explicitly set to: %s\n", runtime)
			return runtime, ""
		default:
			fmt.Printf("⚠️  Warning: Unknown CONTAINER_RUNTIME value '%s', using auto-detection\n", runtimeEnv)
		}
	}

	// 2. Check if DOCKER_HOST is set (indicates Podman or remote Docker)
	if dockerHost := os.Getenv("DOCKER_HOST"); dockerHost != "" {
		// Check if DOCKER_HOST points to Podman socket
		if strings.Contains(dockerHost, "podman") {
			fmt.Printf("🔧 Detected Podman via DOCKER_HOST: %s\n", dockerHost)
			return RuntimePodman, dockerHost
		}
		// Could be remote Docker, assume Docker
		fmt.Printf("🔧 Detected Docker via DOCKER_HOST: %s\n", dockerHost)
		return RuntimeDocker, dockerHost
	}

	// 3. Check if Podman is available and running
	if isPodmanAvailable() {
		socketPath := detectPodmanSocket()
		if socketPath != "" {
			// Auto-configure DOCKER_HOST for Podman
			if err := os.Setenv("DOCKER_HOST", socketPath); err != nil {
				fmt.Printf("⚠️  Warning: Failed to set DOCKER_HOST: %v\n", err)
			}
			fmt.Printf("🔧 Auto-detected and configured Podman: %s\n", socketPath)
			return RuntimePodman, socketPath
		}
		fmt.Println("🔧 Podman detected but socket not accessible")
		fmt.Println("💡 Tip: Run 'podman machine init' and 'podman machine start' on macOS")
		return RuntimePodman, ""
	}

	// 4. Check if Docker is available and running
	if isDockerAvailable() {
		fmt.Println("🔧 Auto-detected Docker")
		return RuntimeDocker, ""
	}

	// 5. No runtime found
	fmt.Println("⚠️  Warning: No container runtime detected (Docker or Podman)")
	fmt.Println("💡 Tip: Install Docker or Podman, or set CONTAINER_RUNTIME=docker|podman")
	return RuntimeUnknown, ""
}

// isDockerAvailable checks if Docker is available and running
func isDockerAvailable() bool {
	if _, err := exec.LookPath("docker"); err != nil {
		return false
	}

	cmd := exec.Command("docker", "info")
	return cmd.Run() == nil
}

// detectPodmanSocket attempts to find the Podman socket location
func detectPodmanSocket() string {
	// Common Podman socket locations
	possibleSockets := []string{
		"unix:///run/podman/podman.sock",                                         // Linux
		"unix:///run/user/1000/podman/podman.sock",                               // Linux user namespace
		"unix:///var/run/podman/podman.sock",                                     // Alternative Linux
		"unix:///Users/$USER/.local/share/containers/podman/machine/podman.sock", // macOS
	}

	// On macOS, try to get the actual socket from podman machine
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("podman", "machine", "inspect", "--format", "{{.ConnectionInfo.URI.Path}}")
		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			socket := strings.TrimSpace(string(output))
			if socket != "" {
				return "unix://" + socket
			}
		}
	}

	// Try common socket locations
	for _, socket := range possibleSockets {
		// Expand $USER in path
		socket = os.ExpandEnv(socket)

		// Extract the file path from unix:// prefix
		filePath := strings.TrimPrefix(socket, "unix://")

		// Check if the socket file exists and is accessible
		if _, err := os.Stat(filePath); err == nil {
			return socket
		}
	}

	return ""
}

// isPodmanAvailable checks if Podman is available and running
func isPodmanAvailable() bool {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return false
	}

	if _, err := exec.LookPath("podman"); err != nil {
		return false
	}

	cmd := exec.Command("podman", "info")
	return cmd.Run() == nil
}

// configureTestcontainers configures testcontainers for the detected runtime
func configureTestcontainers(runtime ContainerRuntime) {
	// Disable Ryuk (reaper) for Podman as it has issues with socket mounting
	if runtime == RuntimePodman {
		if err := os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true"); err != nil {
			fmt.Printf("⚠️  Warning: Failed to set TESTCONTAINERS_RYUK_DISABLED: %v\n", err)
		}
		fmt.Println("🔧 Disabled Ryuk reaper for Podman compatibility")
	}
}

// NewTestDBHelper creates a new database helper for tests using testcontainers
// It automatically detects Docker or Podman and configures appropriately
func NewTestDBHelper() *TestDBHelper {
	runtime, _ := getContainerRuntime()
	configureTestcontainers(runtime)
	return &TestDBHelper{
		runtime: runtime,
	}
}

// Setup creates a PostgreSQL test container, runs migrations, and waits for ready
func (h *TestDBHelper) Setup(t *testing.T) {
	if t != nil {
		t.Helper()
	}

	var err error
	h.ctx = context.Background()

	// Base container options
	opts := []testcontainers.ContainerCustomizer{
		postgres.WithDatabase("zercle_test_db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60 * time.Second),
		),
	}

	// Apply runtime-specific configuration
	switch h.runtime {
	case RuntimePodman:
		fmt.Println("🔧 Applying Podman-specific container configuration")
		// Podman on macOS uses "podman" network driver by default
		// We don't need to explicitly set network mode as testcontainers handles this
		// The Ryuk reaper is already disabled via configureTestcontainers()
	case RuntimeDocker:
		fmt.Println("🔧 Using Docker container configuration")
		// Docker uses "bridge" network by default, which is the standard
	case RuntimeUnknown:
		fmt.Println("⚠️  Warning: Unknown container runtime, using default configuration")
	}

	h.container, err = postgres.Run(
		h.ctx,
		"postgres:18-alpine",
		opts...,
	)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to create test container with %s: %v", h.runtime, err)
		if t != nil {
			t.Fatalf("%s", errMsg)
		} else {
			panic(errMsg)
		}
	}

	connStr, err := h.container.ConnectionString(h.ctx, "sslmode=disable")
	if err != nil {
		errMsg := fmt.Sprintf("Failed to get connection string: %v", err)
		if t != nil {
			t.Fatalf("%s", errMsg)
		} else {
			panic(errMsg)
		}
	}

	h.db, err = sql.Open("pgx", connStr)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to open database: %v", err)
		if t != nil {
			t.Fatalf("%s", errMsg)
		} else {
			panic(errMsg)
		}
	}

	successMsg := fmt.Sprintf("✅ Test database container started using %s runtime", h.runtime)
	if t != nil {
		t.Log(successMsg)
	} else {
		fmt.Println(successMsg)
	}

	h.runMigrations(t)
}

// runMigrations applies all database migrations
func (h *TestDBHelper) runMigrations(t *testing.T) {
	if t != nil {
		t.Helper()
	}

	migrationDir := "../../sqlc/migrations"

	files, err := os.ReadDir(migrationDir)
	if err != nil {
		msg := fmt.Sprintf("⚠️  Warning: Failed to read migration directory: %v", err)
		if t != nil {
			t.Log(msg)
		} else {
			fmt.Println(msg)
		}
		return
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".up.sql") {
			continue
		}

		filePath := filepath.Join(migrationDir, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			msg := fmt.Sprintf("⚠️  Warning: Failed to read migration file %s: %v", file.Name(), err)
			if t != nil {
				t.Log(msg)
			} else {
				fmt.Println(msg)
			}
			continue
		}

		if _, err := h.db.Exec(string(content)); err != nil {
			msg := fmt.Sprintf("⚠️  Warning: Failed to execute migration %s: %v", file.Name(), err)
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

// Cleanup truncates all tables for clean test state and stops container
func (h *TestDBHelper) Cleanup(t *testing.T) {
	if t != nil {
		t.Helper()
	}

	if h.db == nil {
		return
	}

	tables := []string{
		"tasks",
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

// Close stops the container and closes database connection
func (h *TestDBHelper) Close() {
	if h.db != nil {
		if err := h.db.Close(); err != nil {
			fmt.Printf("⚠️  Warning: Failed to close database connection: %v\n", err)
		}
	}

	if h.container != nil {
		if err := h.container.Terminate(h.ctx); err != nil {
			fmt.Printf("⚠️  Warning: Failed to terminate %s container: %v\n", h.runtime, err)
		} else {
			fmt.Printf("✅ %s container terminated successfully\n", h.runtime)
		}
	}
}

// GetDB returns the database connection for advanced test scenarios
func (h *TestDBHelper) GetDB() *sql.DB {
	return h.db
}

// GetRuntime returns the detected container runtime
func (h *TestDBHelper) GetRuntime() ContainerRuntime {
	return h.runtime
}
