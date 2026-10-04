//go:build unit

package postgres_test

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"reflect"
	"testing"
	"time"
	"uuid"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/features/machines/domain"
	pgrepo "github.com/zercle/zercle-go-template/internal/features/machines/repository/postgres"
	"github.com/zercle/zercle-go-template/internal/testutil"
	"github.com/zercle/zercle-go-template/internal/testutil/fixtures"
)

// Documented SQL expectations (observed with gorm.io/driver/postgres +
// SkipDefaultTransaction=true):
//   - Create emits INSERT with five positional args in model column order
//     (id, label, coin_bank, created_at, updated_at).
//   - GetByID emits SELECT * FROM "machines" WHERE id = $1 ORDER BY
//     "machines"."id" LIMIT $2.
//   - List emits SELECT * FROM "machines" ORDER BY created_at DESC, id DESC
//     LIMIT $1.
//   - RestockBank runs inside a transaction: SELECT ... FOR UPDATE, then the
//     coin_bank UPDATE (updated_at set automatically).
func TestRepository_Create(t *testing.T) {
	t.Parallel()

	gormDB, mock := testutil.NewSQLMockDB(t)
	repo := pgrepo.NewRepository(gormDB)

	machine := fixtures.NewMachine("lobby-1", fixtures.NewBank())

	mock.ExpectExec(`INSERT INTO "machines"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.Create(context.Background(), &machine))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Create_NilMachine(t *testing.T) {
	t.Parallel()

	gormDB, mock := testutil.NewSQLMockDB(t)
	repo := pgrepo.NewRepository(gormDB)

	err := repo.Create(context.Background(), nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "create machine")
	assert.Contains(t, err.Error(), "nil")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetByID(t *testing.T) {
	t.Parallel()

	gormDB, mock := testutil.NewSQLMockDB(t)
	repo := pgrepo.NewRepository(gormDB)

	id := uuid.New()
	now := time.Now().UTC()
	bank := `{"5":10,"25":4}`

	mock.ExpectQuery(`SELECT \* FROM "machines" WHERE id = \$1 ORDER BY "machines"\."id" LIMIT \$2`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "label", "coin_bank", "created_at", "updated_at"}).
				AddRow(id.String(), "found", bank, now, now),
		)

	got, err := repo.GetByID(context.Background(), id)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, id, got.ID)
	assert.Equal(t, "found", got.Label)
	assert.Equal(t, domain.CoinBank{5: 10, 25: 4}, got.CoinBank)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	gormDB, mock := testutil.NewSQLMockDB(t)
	repo := pgrepo.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "machines" WHERE id = \$1 ORDER BY "machines"\."id" LIMIT \$2`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "label", "coin_bank", "created_at", "updated_at"}),
		)

	got, err := repo.GetByID(context.Background(), uuid.New())
	assert.Nil(t, got)
	assert.ErrorIs(t, err, domain.ErrMachineNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_List(t *testing.T) {
	t.Parallel()

	gormDB, mock := testutil.NewSQLMockDB(t)
	repo := pgrepo.NewRepository(gormDB)

	id := uuid.New()
	now := time.Now().UTC()

	mock.ExpectQuery(`SELECT \* FROM "machines" ORDER BY created_at DESC, id DESC LIMIT \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "label", "coin_bank", "created_at", "updated_at"}).
				AddRow(id.String(), "listed", `{"5":1}`, now, now),
		)

	machines, err := repo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	require.Len(t, machines, 1)
	assert.Equal(t, id, machines[0].ID)
	assert.Equal(t, "listed", machines[0].Label)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// coinBankArg matches the serialized coin_bank argument against a wanted bank,
// so the restock test proves the coins were actually added rather than only
// that the UPDATE ran.
type coinBankArg struct {
	want domain.CoinBank
}

func (a coinBankArg) Match(value driver.Value) bool {
	var raw []byte
	switch v := value.(type) {
	case string:
		raw = []byte(v)
	case []byte:
		raw = v
	default:
		return false
	}
	var got domain.CoinBank
	if err := json.Unmarshal(raw, &got); err != nil {
		return false
	}
	return reflect.DeepEqual(a.want, got)
}

func TestRepository_RestockBank(t *testing.T) {
	t.Parallel()

	gormDB, mock := testutil.NewSQLMockDB(t)
	repo := pgrepo.NewRepository(gormDB)

	id := uuid.New()
	now := time.Now().UTC()
	before := domain.CoinBank{5: 1, 25: 2}
	want := domain.AddCoins(before, []int32{5, 100})

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "machines" WHERE id = \$1 ORDER BY "machines"\."id" LIMIT \$2 FOR UPDATE`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "label", "coin_bank", "created_at", "updated_at"}).
				AddRow(id.String(), "restock", `{"5":1,"25":2}`, now, now),
		)
	mock.ExpectExec(`UPDATE "machines" SET "coin_bank"=\$1,"updated_at"=\$2 WHERE id = \$3`).
		WithArgs(coinBankArg{want: want}, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, repo.RestockBank(context.Background(), id, []int32{5, 100}))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_RestockBank_MachineNotFound(t *testing.T) {
	t.Parallel()

	gormDB, mock := testutil.NewSQLMockDB(t)
	repo := pgrepo.NewRepository(gormDB)

	id := uuid.New()
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "machines" WHERE id = \$1 ORDER BY "machines"\."id" LIMIT \$2 FOR UPDATE`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "label", "coin_bank", "created_at", "updated_at"}),
		)
	mock.ExpectRollback()

	err := repo.RestockBank(context.Background(), id, []int32{5})
	assert.ErrorIs(t, err, domain.ErrMachineNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}
