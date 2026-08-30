package db

import (
	"context"
	"fmt"
	"sync"

	"gorm.io/gorm"
)

// Shutdowner adapts *gorm.DB to samber/do's ShutdownerWithContextAndError
// interface so the DI container owns the connection-pool lifecycle. This
// guarantees the pool is released by injector.Shutdown() even when
// server.Application.shutdown() never runs — e.g. a partial build failure
// in app.Build, or tests that call Build followed by injector.Shutdown.
//
// The underlying *sql.DB is closed exactly once via sync.Once; subsequent
// calls are no-ops, keeping shutdown idempotent when both the Application
// and the DI container close the same pool in the happy path.
type Shutdowner struct {
	db   *gorm.DB
	once sync.Once
	err  error
}

// NewShutdowner wraps db so the DI container can close it.
func NewShutdowner(db *gorm.DB) *Shutdowner {
	return &Shutdowner{db: db}
}

// Shutdown implements do.ShutdownerWithContextAndError.
func (s *Shutdowner) Shutdown(context.Context) error {
	s.once.Do(func() {
		if s.db == nil {
			return
		}
		sqlDB, err := s.db.DB()
		if err != nil {
			s.err = fmt.Errorf("gorm db sql handle unavailable: %w", err)
			return
		}
		s.err = sqlDB.Close()
	})
	return s.err
}
