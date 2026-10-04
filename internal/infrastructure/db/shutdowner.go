package db

import (
	"fmt"

	"github.com/samber/do/v2"
	"gorm.io/gorm"

	"github.com/zercle/zercle-go-template/internal/infrastructure/lifecycle"
)

// NewShutdownCloser adapts a *gorm.DB to samber/do's
// ShutdownerWithContextAndError so the DI container owns the connection-pool
// lifecycle. This guarantees the pool is released by injector.Shutdown() even
// when server.Application.shutdown() never runs — e.g. a partial build failure
// in app.Build, or tests that call Build followed by injector.Shutdown.
//
// A nil *gorm.DB yields a no-op closer; the underlying *sql.DB is closed
// exactly once, keeping shutdown idempotent when both the Application and the
// container close the same pool.
func NewShutdownCloser(db *gorm.DB) do.ShutdownerWithContextAndError {
	return lifecycle.NewCloser(func() error {
		if db == nil {
			return nil
		}
		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("gorm db sql handle unavailable: %w", err)
		}
		return sqlDB.Close()
	})
}
