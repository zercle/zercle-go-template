// Command migrate is a self-contained database migration runner. It embeds SQL
// migration files via go:embed and uses the golang-migrate library with the
// iofs source driver.
package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // postgres driver registers "postgres://" DSNs
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/zercle/zercle-go-template/internal/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/db/migrations"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

// run parses arguments, builds a self-contained migrator, and executes the
// requested command. It returns the process exit code.
func run(args []string) (exitCode int) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		return 1
	}

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "invalid config: %v\n", err)
		return 1
	}

	cmd, count, forceVersion, err := parseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		printUsage()
		return 1
	}

	m, err := newMigrator(cfg.DBConnString())
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create migrator: %v\n", err)
		return 1
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil || dbErr != nil {
			fmt.Fprintf(os.Stderr, "failed to close migrator: source=%v db=%v\n", srcErr, dbErr)
			if exitCode == 0 {
				exitCode = 1
			}
		}
	}()

	switch cmd {
	case "up":
		return runUp(m)
	case "down":
		return runDown(m, count)
	case "force":
		return runForce(m, forceVersion)
	case "version":
		return runVersion(m)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printUsage()
		return 1
	}
}

// newMigrator builds a golang-migrate instance backed by the embedded SQL files.
func newMigrator(dsn string) (*migrate.Migrate, error) {
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return nil, fmt.Errorf("create migration source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, dsn)
	if err != nil {
		return nil, fmt.Errorf("create migrator: %w", err)
	}

	return m, nil
}

func runUp(m *migrate.Migrate) int {
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		fmt.Fprintf(os.Stderr, "migration up failed: %v\n", err)
		return 1
	}
	return printVersion(m)
}

func runDown(m *migrate.Migrate, count int) int {
	if err := m.Steps(-count); err != nil {
		fmt.Fprintf(os.Stderr, "migration down failed: %v\n", err)
		return 1
	}
	return printVersion(m)
}

func runForce(m *migrate.Migrate, version int) int {
	if err := m.Force(version); err != nil {
		fmt.Fprintf(os.Stderr, "force migration version failed: %v\n", err)
		return 1
	}
	return printVersion(m)
}

func runVersion(m *migrate.Migrate) int {
	version, dirty, err := m.Version()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read version: %v\n", err)
		return 1
	}
	fmt.Printf("version %d dirty=%v\n", version, dirty)
	return 0
}

// printVersion prints the current migration version after a state change.
func printVersion(m *migrate.Migrate) int {
	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		fmt.Fprintf(os.Stderr, "failed to read version after migration: %v\n", err)
		return 1
	}

	if errors.Is(err, migrate.ErrNilVersion) {
		fmt.Println("no migrations applied")
		return 0
	}

	fmt.Printf("migration complete: version %d dirty=%v\n", version, dirty)
	return 0
}

// parseArgs extracts the command and optional numeric arguments from os.Args.
// Defaults are: command "up", down count 1. Force expects a version integer.
func parseArgs(args []string) (cmd string, count int, forceVersion int, err error) {
	cmd = "up"
	count = 1

	if len(args) == 0 {
		return cmd, count, forceVersion, nil
	}

	cmd = args[0]

	switch cmd {
	case "down":
		if len(args) >= 2 {
			parsed, parseErr := strconv.Atoi(args[1])
			if parseErr != nil || parsed <= 0 {
				return cmd, count, forceVersion, fmt.Errorf("invalid down count %q: must be a positive integer", args[1])
			}
			count = parsed
		}
	case "force":
		if len(args) < 2 {
			return cmd, count, forceVersion, errors.New("force requires a version argument")
		}
		parsed, parseErr := strconv.Atoi(args[1])
		if parseErr != nil {
			return cmd, count, forceVersion, fmt.Errorf("invalid force version %q: must be an integer", args[1])
		}
		forceVersion = parsed
	}

	return cmd, count, forceVersion, nil
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "usage: migrate [up | down [N] | force VERSION | version]")
}
