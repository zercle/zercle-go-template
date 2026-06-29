//go:build unit

package db_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/zercle/zercle-go-template/internal/infrastructure/db"
)

// newMockGormDB builds a *gorm.DB backed by sqlmock so we can exercise the
// shutdowner against a real *sql.DB handle without requiring a live database.
// The DI container's shutdown path needs the underlying *sql.DB to be
// closable; sqlmock satisfies that contract.
//
// The returned sqlmock controller must have its expectations managed by the
// caller (e.g. via ExpectClose) before invoking the shutdowner; sqlmock
// rejects unexpected calls by default.
func newMockGormDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(false))
	require.NoError(t, err, "create sqlmock")
	t.Cleanup(func() { _ = sqlDB.Close() })

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	require.NoError(t, err, "open gorm against sqlmock")

	return gormDB, mock
}

// TestShutdowner_NilDBIsSafe verifies that constructing a shutdowner with
// a nil *gorm.DB does not panic and returns nil from Shutdown — covering
// the case where the DI container is asked to close a never-configured DB.
func TestShutdowner_NilDBIsSafe(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	s := db.NewShutdowner(nil)
	require.NotNil(t, s, "shutdowner constructor must return non-nil")

	assert.NoError(t, s.Shutdown(ctx), "shutdown with nil db must return nil")
	assert.NoError(t, s.Shutdown(ctx), "second shutdown call must remain a no-op")
}

// TestShutdowner_ClosesUnderlyingPool verifies that Shutdown closes the
// underlying *sql.DB exactly once and remains idempotent across repeated
// calls. The second call must return no error without panicking, mirroring
// the production scenario where both the Application's graceful shutdown
// and the DI container's shutdown close the same pool.
func TestShutdowner_ClosesUnderlyingPool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	gormDB, mock := newMockGormDB(t)

	mock.ExpectClose()

	s := db.NewShutdowner(gormDB)
	require.NotNil(t, s)

	firstErr := s.Shutdown(ctx)
	assert.NoError(t, firstErr, "first shutdown must close the pool cleanly")
	assert.NoError(t, mock.ExpectationsWereMet(), "sqlmock must observe exactly one Close call")

	secondErr := s.Shutdown(ctx)
	assert.NoError(t, secondErr, "second shutdown must be a no-op and must not propagate sql.ErrConnDone")
}
