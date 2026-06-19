//go:build unit

// STUB FEATURE — delete internal/features/example to start your project.

package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/zercle/zercle-go-template/internal/features/example/domain"
	"github.com/zercle/zercle-go-template/internal/features/example/repository"
)

// newTestDB builds a *gorm.DB backed by go-sqlmock so each test can assert
// exact SQL emitted by GORM without touching a real database.
//
// Notes on matching (observed empirically with gorm.io/driver/postgres +
// SkipDefaultTransaction=true):
//   - QueryMatcherRegexp is the default; the regex patterns below mirror the
//     SQL GORM actually emits.
//   - Create: ExpectExec `INSERT INTO "items" ... VALUES (...)` with four
//     positional args. The ORDER of args matches the column order in the
//     GORM model (id, name, created_at, updated_at).
//   - GetByID: GORM emits
//       SELECT * FROM "items" WHERE id = $1 ORDER BY "items"."id" LIMIT $2
//     i.e. TWO bound args (the id and the literal 1 for LIMIT). The
//     expectation passes AnyArg() twice.
//   - List with offset=0 omits the OFFSET clause entirely, so the regex
//     tolerates the OFFSET being absent:
//       SELECT * FROM "items" ORDER BY created_at DESC, id DESC LIMIT $1
//   - For uuid args we still use AnyArg() to avoid driver-level type
//     mismatch (uuid.UUID vs string vs [16]byte representations).
//   - sqlmock.NewRows(...).AddRow(id.String(), ...) returns the uuid as a
//     string; GORM's postgres driver scans the column back into uuid.UUID
//     without complaint.
func newTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
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

func TestRepository_Create(t *testing.T) {
	gormDB, mock := newTestDB(t)
	repo := repository.NewRepository(gormDB)

	item := &domain.Item{
		ID:        uuid.New(),
		Name:      "repo-item",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	mock.ExpectExec(`INSERT INTO "items"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Create(context.Background(), item)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Create_Error(t *testing.T) {
	gormDB, mock := newTestDB(t)
	repo := repository.NewRepository(gormDB)

	item := &domain.Item{
		ID:        uuid.New(),
		Name:      "x",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	mock.ExpectExec(`INSERT INTO "items"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("exec failed"))

	err := repo.Create(context.Background(), item)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create item")
	assert.Contains(t, err.Error(), "exec failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetByID(t *testing.T) {
	gormDB, mock := newTestDB(t)
	repo := repository.NewRepository(gormDB)

	id := uuid.New()
	now := time.Now().UTC()
	name := "found"

	mock.ExpectQuery(`SELECT \* FROM "items" WHERE id = \$1 ORDER BY "items"\."id" LIMIT \$2`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"}).
				AddRow(id.String(), name, now, now),
		)

	got, err := repo.GetByID(context.Background(), id)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, id, got.ID)
	assert.Equal(t, name, got.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	gormDB, mock := newTestDB(t)
	repo := repository.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "items" WHERE id = \$1 ORDER BY "items"\."id" LIMIT \$2`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"}),
		)

	got, err := repo.GetByID(context.Background(), uuid.New())
	assert.Nil(t, got)
	assert.True(t, errors.Is(err, domain.ErrItemNotFound))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_List(t *testing.T) {
	gormDB, mock := newTestDB(t)
	repo := repository.NewRepository(gormDB)

	id := uuid.New()
	now := time.Now().UTC()
	limit, offset := int32(10), int32(0)

	mock.ExpectQuery(`SELECT \* FROM "items" ORDER BY created_at DESC, id DESC LIMIT \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"}).
				AddRow(id.String(), "listed", now, now),
		)

	items, err := repo.List(context.Background(), limit, offset)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, id, items[0].ID)
	assert.Equal(t, "listed", items[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_List_WithOffset(t *testing.T) {
	gormDB, mock := newTestDB(t)
	repo := repository.NewRepository(gormDB)

	limit, offset := int32(10), int32(5)

	mock.ExpectQuery(`SELECT \* FROM "items" ORDER BY created_at DESC, id DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"}),
		)

	items, err := repo.List(context.Background(), limit, offset)
	require.NoError(t, err)
	assert.Empty(t, items)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_List_Error(t *testing.T) {
	gormDB, mock := newTestDB(t)
	repo := repository.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "items" ORDER BY created_at DESC, id DESC LIMIT \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(errors.New("query failed"))

	items, err := repo.List(context.Background(), 10, 0)
	assert.Error(t, err)
	assert.Nil(t, items)
	assert.Contains(t, err.Error(), "list items")
	assert.NoError(t, mock.ExpectationsWereMet())
}
