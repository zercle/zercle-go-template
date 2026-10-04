package testutil

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewSQLMockDB builds a *gorm.DB backed by go-sqlmock so tests can assert the
// SQL GORM emits without touching a real database. It centralizes the
// per-test setup that feature repository tests would otherwise duplicate.
//
// The options mirror the ones the example repository's own helper uses:
// QueryMatcherRegexp (the sqlmock default), SkipDefaultTransaction so each
// statement is emitted on its own, and a silent logger so expected errors do
// not spam the test output. The returned mock still expects the caller to
// program its query/exec expectations.
func NewSQLMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	return gormDB, mock
}
