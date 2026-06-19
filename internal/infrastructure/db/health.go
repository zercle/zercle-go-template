package db

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// gormChecker reports PostgreSQL connectivity by pinging the underlying
// *sql.DB exposed by *gorm.DB.
type gormChecker struct {
	db *gorm.DB
}

// Name returns the dependency name reported in health output.
func (gormChecker) Name() string {
	return "postgres"
}

// Check verifies PostgreSQL is reachable by pinging the connection pool.
func (c gormChecker) Check(ctx context.Context) error {
	if c.db == nil {
		return fmt.Errorf("gorm db is not initialized")
	}
	sqlDB, err := c.db.DB()
	if err != nil {
		return fmt.Errorf("get sql db: %w", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}
	return nil
}
