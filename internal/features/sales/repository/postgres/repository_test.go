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

	"github.com/zercle/zercle-go-template/internal/features/sales/domain"
	pgrepo "github.com/zercle/zercle-go-template/internal/features/sales/repository/postgres"
	"github.com/zercle/zercle-go-template/internal/testutil"
)

// Documented SQL expectations (observed with gorm.io/driver/postgres +
// SkipDefaultTransaction=true):
//   - GetProduct emits SELECT * FROM "catalog_products" WHERE id = $1 ORDER BY
//     "catalog_products"."id" LIMIT $2.
//   - GetMachineBank emits SELECT * FROM "machines" WHERE id = $1 ORDER BY
//     "machines"."id" LIMIT $2.
//   - CommitPurchase runs inside a transaction: lock product, guarded stock
//     decrement, machine coin_bank update, purchase insert.
func TestRepository_GetProduct(t *testing.T) {
	t.Parallel()

	gormDB, mock := testutil.NewSQLMockDB(t)
	repo := pgrepo.NewRepository(gormDB)

	id := uuid.New()
	now := time.Now().UTC()

	mock.ExpectQuery(`SELECT \* FROM "catalog_products" WHERE id = \$1 ORDER BY "catalog_products"\."id" LIMIT \$2`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "price_cents", "stock", "created_at", "updated_at"}).
				AddRow(id.String(), int32(150), int32(3), now, now),
		)

	got, err := repo.GetProduct(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, id, got.ProductID)
	assert.EqualValues(t, 150, got.PriceCents)
	assert.EqualValues(t, 3, got.Stock)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetProduct_NotFound(t *testing.T) {
	t.Parallel()

	gormDB, mock := testutil.NewSQLMockDB(t)
	repo := pgrepo.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "catalog_products" WHERE id = \$1 ORDER BY "catalog_products"\."id" LIMIT \$2`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "price_cents", "stock", "created_at", "updated_at"}))

	_, err := repo.GetProduct(context.Background(), uuid.New())
	assert.ErrorIs(t, err, domain.ErrProductNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetMachineBank(t *testing.T) {
	t.Parallel()

	gormDB, mock := testutil.NewSQLMockDB(t)
	repo := pgrepo.NewRepository(gormDB)

	id := uuid.New()
	now := time.Now().UTC()

	mock.ExpectQuery(`SELECT \* FROM "machines" WHERE id = \$1 ORDER BY "machines"\."id" LIMIT \$2`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "coin_bank", "created_at", "updated_at"}).
				AddRow(id.String(), `{"5":10,"25":4}`, now, now),
		)

	got, err := repo.GetMachineBank(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, domain.CoinBank{5: 10, 25: 4}, got)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetMachineBank_NotFound(t *testing.T) {
	t.Parallel()

	gormDB, mock := testutil.NewSQLMockDB(t)
	repo := pgrepo.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "machines" WHERE id = \$1 ORDER BY "machines"\."id" LIMIT \$2`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "coin_bank", "created_at", "updated_at"}))

	_, err := repo.GetMachineBank(context.Background(), uuid.New())
	assert.ErrorIs(t, err, domain.ErrMachineNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func newPurchaseRecord(machineID, productID uuid.UUID) *domain.PurchaseRecord {
	return &domain.PurchaseRecord{
		ID:                 uuid.New(),
		MachineID:          machineID,
		ProductID:          productID,
		PriceCents:         100,
		TotalInsertedCents: 125,
		ChangeCents:        25,
		ChangeCoins:        []int32{25},
		PurchasedAt:        time.Now().UTC(),
	}
}

// expectPurchaseStatements programs the shared happy-path transaction
// expectations: lock, guarded stock decrement, machine bank update, purchase
// insert, commit.
func expectPurchaseStatements(mock sqlmock.Sqlmock, productID, machineID uuid.UUID) {
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "catalog_products" WHERE id = \$1 ORDER BY "catalog_products"\."id" LIMIT \$2 FOR UPDATE`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "price_cents", "stock", "created_at", "updated_at"}).
				AddRow(productID.String(), int32(100), int32(1), time.Now().UTC(), time.Now().UTC()),
		)
	mock.ExpectExec(`UPDATE "catalog_products" SET "stock"=stock - 1,"updated_at"=\$1 WHERE id = \$2 AND stock > 0`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE "machines" SET "coin_bank"=\$1,"updated_at"=\$2 WHERE id = \$3`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO "sales_purchases"`).
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
}

func TestRepository_CommitPurchase(t *testing.T) {
	t.Parallel()

	gormDB, mock := testutil.NewSQLMockDB(t)
	repo := pgrepo.NewRepository(gormDB)

	productID := uuid.New()
	machineID := uuid.New()
	expectPurchaseStatements(mock, productID, machineID)

	err := repo.CommitPurchase(context.Background(), machineID, productID, newPurchaseRecord(machineID, productID), domain.CoinBank{25: 11})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_CommitPurchase_ProductNotFound(t *testing.T) {
	t.Parallel()

	gormDB, mock := testutil.NewSQLMockDB(t)
	repo := pgrepo.NewRepository(gormDB)

	productID := uuid.New()
	machineID := uuid.New()
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "catalog_products" WHERE id = \$1 ORDER BY "catalog_products"\."id" LIMIT \$2 FOR UPDATE`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "price_cents", "stock", "created_at", "updated_at"}))
	mock.ExpectRollback()

	err := repo.CommitPurchase(context.Background(), machineID, productID, newPurchaseRecord(machineID, productID), domain.CoinBank{25: 11})
	assert.ErrorIs(t, err, domain.ErrProductNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_CommitPurchase_OutOfStock(t *testing.T) {
	t.Parallel()

	gormDB, mock := testutil.NewSQLMockDB(t)
	repo := pgrepo.NewRepository(gormDB)

	productID := uuid.New()
	machineID := uuid.New()
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "catalog_products" WHERE id = \$1 ORDER BY "catalog_products"\."id" LIMIT \$2 FOR UPDATE`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "price_cents", "stock", "created_at", "updated_at"}).
				AddRow(productID.String(), int32(100), int32(0), time.Now().UTC(), time.Now().UTC()),
		)
	mock.ExpectExec(`UPDATE "catalog_products" SET "stock"=stock - 1,"updated_at"=\$1 WHERE id = \$2 AND stock > 0`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := repo.CommitPurchase(context.Background(), machineID, productID, newPurchaseRecord(machineID, productID), domain.CoinBank{25: 11})
	assert.ErrorIs(t, err, domain.ErrOutOfStock)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_CommitPurchase_MachineNotFound(t *testing.T) {
	t.Parallel()

	gormDB, mock := testutil.NewSQLMockDB(t)
	repo := pgrepo.NewRepository(gormDB)

	productID := uuid.New()
	machineID := uuid.New()
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "catalog_products" WHERE id = \$1 ORDER BY "catalog_products"\."id" LIMIT \$2 FOR UPDATE`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "price_cents", "stock", "created_at", "updated_at"}).
				AddRow(productID.String(), int32(100), int32(1), time.Now().UTC(), time.Now().UTC()),
		)
	mock.ExpectExec(`UPDATE "catalog_products" SET "stock"=stock - 1,"updated_at"=\$1 WHERE id = \$2 AND stock > 0`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE "machines" SET "coin_bank"=\$1,"updated_at"=\$2 WHERE id = \$3`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := repo.CommitPurchase(context.Background(), machineID, productID, newPurchaseRecord(machineID, productID), domain.CoinBank{25: 11})
	assert.ErrorIs(t, err, domain.ErrMachineNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}
