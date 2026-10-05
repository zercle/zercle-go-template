//go:build unit

package postgres_test

import (
	"context"
	"testing"
	"uuid"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pgrepo "github.com/zercle/zercle-go-template/internal/features/reporting/repository/postgres"
	"github.com/zercle/zercle-go-template/internal/testutil"
)

// Documented SQL expectations (observed with gorm.io/driver/postgres +
// SkipDefaultTransaction=true): both reads are single raw statements. The
// overview statement takes no arguments and returns one row of scalar
// subqueries (including the coin-bank sum over jsonb_each_text, because
// machines stores coin_bank as a JSON object, not an array); the top-machines
// statement takes LIMIT as $1.
func TestRepository_GetOverview(t *testing.T) {
	t.Parallel()

	gormDB, mock := testutil.NewSQLMockDB(t)
	repo := pgrepo.NewRepository(gormDB)

	mock.ExpectQuery(`FROM catalog_products.*jsonb_each_text\(m\.coin_bank\).*FROM sales_purchases`).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"product_count", "total_stock", "machine_count", "total_bank_cents", "purchase_count", "revenue_cents",
			}).AddRow(int64(4), int64(12), int64(2), int64(350), int64(7), int64(900)),
		)

	got, err := repo.GetOverview(context.Background())
	require.NoError(t, err)
	assert.EqualValues(t, 4, got.ProductCount)
	assert.EqualValues(t, 12, got.TotalStock)
	assert.EqualValues(t, 2, got.MachineCount)
	assert.EqualValues(t, 350, got.TotalBankCents)
	assert.EqualValues(t, 7, got.PurchaseCount)
	assert.EqualValues(t, 900, got.RevenueCents)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetTopMachines(t *testing.T) {
	t.Parallel()

	gormDB, mock := testutil.NewSQLMockDB(t)
	repo := pgrepo.NewRepository(gormDB)

	firstID := uuid.MustParse("12345678-1234-1234-1234-123456789abc")
	secondID := uuid.MustParse("22345678-1234-1234-1234-123456789abc")

	mock.ExpectQuery(`FROM machines m.*JOIN sales_purchases.*LIMIT \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(
			sqlmock.NewRows([]string{"machine_id", "label", "purchase_count", "revenue_cents"}).
				AddRow(firstID.String(), "lobby", int64(5), int64(700)).
				AddRow(secondID.String(), "cafe", int64(2), int64(200)),
		)

	got, err := repo.GetTopMachines(context.Background(), 5)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, firstID, got[0].MachineID)
	assert.Equal(t, "lobby", got[0].Label)
	assert.EqualValues(t, 5, got[0].PurchaseCount)
	assert.EqualValues(t, 700, got[0].RevenueCents)
	assert.Equal(t, secondID, got[1].MachineID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_GetTopMachines_Empty pins the read-only contract: no rows is
// an empty slice and no error, never a not-found.
func TestRepository_GetTopMachines_Empty(t *testing.T) {
	t.Parallel()

	gormDB, mock := testutil.NewSQLMockDB(t)
	repo := pgrepo.NewRepository(gormDB)

	mock.ExpectQuery(`FROM machines m.*LIMIT \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"machine_id", "label", "purchase_count", "revenue_cents"}))

	got, err := repo.GetTopMachines(context.Background(), 5)
	require.NoError(t, err)
	assert.Empty(t, got)
	assert.NoError(t, mock.ExpectationsWereMet())
}
