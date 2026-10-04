//go:build unit

package postgres_test

import (
	"context"
	"testing"
	"time"
	"uuid"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/features/catalog/domain"
	pgrepo "github.com/zercle/zercle-go-template/internal/features/catalog/repository/postgres"
	"github.com/zercle/zercle-go-template/internal/testutil"
	"github.com/zercle/zercle-go-template/internal/testutil/fixtures"
)

// Documented SQL expectations (observed with gorm.io/driver/postgres +
// SkipDefaultTransaction=true):
//   - Create emits INSERT with six positional args in model column order
//     (id, name, price_cents, stock, created_at, updated_at).
//   - GetByID emits SELECT * FROM "catalog_products" WHERE id = $1 ORDER BY
//     "catalog_products"."id" LIMIT $2 (the id and the literal 1).
//   - List emits SELECT * FROM "catalog_products" ORDER BY created_at DESC,
//     id DESC LIMIT $1.
func TestRepository_Create(t *testing.T) {
	t.Parallel()

	gormDB, mock := testutil.NewSQLMockDB(t)
	repo := pgrepo.NewRepository(gormDB)

	product := fixtures.NewProduct("Cola", 100, 5)

	mock.ExpectExec(`INSERT INTO "catalog_products"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.Create(context.Background(), &product))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Create_NilProduct(t *testing.T) {
	t.Parallel()

	gormDB, mock := testutil.NewSQLMockDB(t)
	repo := pgrepo.NewRepository(gormDB)

	err := repo.Create(context.Background(), nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "create product")
	assert.Contains(t, err.Error(), "nil")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetByID(t *testing.T) {
	t.Parallel()

	gormDB, mock := testutil.NewSQLMockDB(t)
	repo := pgrepo.NewRepository(gormDB)

	id := uuid.New()
	now := time.Now().UTC()

	mock.ExpectQuery(`SELECT \* FROM "catalog_products" WHERE id = \$1 ORDER BY "catalog_products"\."id" LIMIT \$2`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "price_cents", "stock", "created_at", "updated_at"}).
				AddRow(id.String(), "found", int32(150), int32(3), now, now),
		)

	got, err := repo.GetByID(context.Background(), id)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, id, got.ID)
	assert.Equal(t, "found", got.Name)
	assert.EqualValues(t, 150, got.PriceCents)
	assert.EqualValues(t, 3, got.Stock)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	gormDB, mock := testutil.NewSQLMockDB(t)
	repo := pgrepo.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "catalog_products" WHERE id = \$1 ORDER BY "catalog_products"\."id" LIMIT \$2`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "price_cents", "stock", "created_at", "updated_at"}),
		)

	got, err := repo.GetByID(context.Background(), uuid.New())
	assert.Nil(t, got)
	assert.ErrorIs(t, err, domain.ErrProductNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_List(t *testing.T) {
	t.Parallel()

	gormDB, mock := testutil.NewSQLMockDB(t)
	repo := pgrepo.NewRepository(gormDB)

	id := uuid.New()
	now := time.Now().UTC()

	mock.ExpectQuery(`SELECT \* FROM "catalog_products" ORDER BY created_at DESC, id DESC LIMIT \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "price_cents", "stock", "created_at", "updated_at"}).
				AddRow(id.String(), "listed", int32(100), int32(1), now, now),
		)

	products, err := repo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	require.Len(t, products, 1)
	assert.Equal(t, id, products[0].ID)
	assert.Equal(t, "listed", products[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}
