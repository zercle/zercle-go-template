// Package main provides the database migration CLI tool.
// Uses golang-migrate/migrate for versioned database migrations.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
)

func main() {
	action, version, steps := parseFlags()
	m := initMigration()

	switch action {
	case "up":
		runUp(m, steps)
	case "down":
		runDown(m, steps)
	case "version":
		showVersion(m)
	case "force":
		runForce(m, version)
	default:
		fmt.Fprintf(os.Stderr, "Unknown action: %s\n", action)
		fmt.Fprintf(os.Stderr, "Usage: %s [-action=up|down|version|force] [-version=N] [-steps=N]\n", os.Args[0])
		os.Exit(1)
	}
}

func parseFlags() (action string, version int, steps int) {
	actionFlag := flag.String("action", "up", "Migration action: up, down, version, force")
	versionFlag := flag.Int("version", 0, "Target version for force action")
	stepsFlag := flag.Int("steps", 0, "Number of steps for up/down (0 = all)")
	flag.Parse()
	return *actionFlag, *versionFlag, *stepsFlag
}

func initMigration() *migrate.Migrate {
	cfg, err := config.Load(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	dsn := cfg.Database.ConnString()
	m, err := migrate.New(
		"file:///migrations",
		dsn,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create migrate instance: %v\n", err)
		os.Exit(1)
	}
	return m
}

func runUp(m *migrate.Migrate, steps int) {
	var err error
	if steps > 0 {
		err = m.Steps(steps)
	} else {
		err = m.Up()
	}
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		fmt.Fprintf(os.Stderr, "Migration up failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Migrations applied successfully")
}

func runDown(m *migrate.Migrate, steps int) {
	var err error
	if steps > 0 {
		err = m.Steps(-steps)
	} else {
		err = m.Down()
	}
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		fmt.Fprintf(os.Stderr, "Migration down failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Migrations rolled back successfully")
}

func showVersion(m *migrate.Migrate) {
	v, dirty, err := m.Version()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get version: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Current version: %d (dirty: %v)\n", v, dirty)
}

func runForce(m *migrate.Migrate, version int) {
	if version == 0 {
		fmt.Fprintf(os.Stderr, "Version is required for force action\n")
		os.Exit(1)
	}
	if err := m.Force(version); err != nil {
		fmt.Fprintf(os.Stderr, "Force failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Forced to version %d\n", version)
}
