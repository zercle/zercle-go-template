package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/zercle/zercle-go-template/internal/core/port"
)

type healthRepository struct {
	db *sql.DB
}

// NewHealthRepository creates a new instance of HealthRepository.
func NewHealthRepository(db *sql.DB) port.HealthRepository {
	return &healthRepository{
		db: db,
	}
}

// CheckDatabase performs a database connectivity check.
func (r *healthRepository) CheckDatabase(ctx context.Context) (string, error) {
	// Set a timeout for the ping
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if err := r.db.PingContext(ctx); err != nil {
		return "unreachable", err
	}

	// Query database version to verify it's working
	var version string
	if err := r.db.QueryRowContext(ctx, "SELECT version()").Scan(&version); err != nil {
		return "unreachable", err
	}

	return version, nil
}
