//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // postgres driver
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	catalogmigrations "github.com/zercle/zercle-go-template/internal/features/catalog/repository/postgres/migrations"
	machinesmigrations "github.com/zercle/zercle-go-template/internal/features/machines/repository/postgres/migrations"
	"github.com/zercle/zercle-go-template/internal/features/sales/domain"
	"github.com/zercle/zercle-go-template/internal/features/sales/repository/postgres"
	salesmigrations "github.com/zercle/zercle-go-template/internal/features/sales/repository/postgres/migrations"
	"github.com/zercle/zercle-go-template/internal/fsmerge"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/db"
)

type RepositoryIntegrationSuite struct {
	suite.Suite
	db   *gorm.DB
	repo *postgres.Repository
}

func (s *RepositoryIntegrationSuite) SetupSuite() {
	t := s.T()
	t.Helper()

	cfg, err := config.Load()
	require.NoError(t, err)

	if cfg.App.Environment == "production" {
		t.Fatalf("integration tests must not run against production environment (APP_ENVIRONMENT=production)")
	}

	nop := zerolog.Nop()
	s.db, err = db.NewDB(context.Background(), cfg, &nop)
	require.NoError(t, err)

	runMigrations(t, cfg)

	s.repo = postgres.NewRepository(s.db)
}

func (s *RepositoryIntegrationSuite) TearDownSuite() {
	if s.db != nil {
		sqlDB, err := s.db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}
}

func (s *RepositoryIntegrationSuite) SetupTest() {
	resetSalesState(s.T(), s.db)
}

func (s *RepositoryIntegrationSuite) TestGetProduct() {
	t := s.T()
	ctx := context.Background()

	id := uuid.New()
	seedProduct(s.T(), s.db, id, 100, 4)

	got, err := s.repo.GetProduct(ctx, id)
	require.NoError(t, err)
	require.Equal(t, id, got.ProductID)
	require.EqualValues(t, 100, got.PriceCents)
	require.EqualValues(t, 4, got.Stock)
}

func (s *RepositoryIntegrationSuite) TestGetProduct_NotFound() {
	t := s.T()

	_, err := s.repo.GetProduct(context.Background(), uuid.New())
	require.ErrorIs(t, err, domain.ErrProductNotFound)
}

func (s *RepositoryIntegrationSuite) TestGetMachineBank() {
	t := s.T()
	ctx := context.Background()

	id := uuid.New()
	seedMachine(s.T(), s.db, id, `{"5":10,"25":4}`)

	got, err := s.repo.GetMachineBank(ctx, id)
	require.NoError(t, err)
	require.Equal(t, domain.CoinBank{5: 10, 25: 4}, got)
}

func (s *RepositoryIntegrationSuite) TestGetMachineBank_NotFound() {
	t := s.T()

	_, err := s.repo.GetMachineBank(context.Background(), uuid.New())
	require.ErrorIs(t, err, domain.ErrMachineNotFound)
}

func (s *RepositoryIntegrationSuite) TestCommitPurchase() {
	t := s.T()
	ctx := context.Background()

	productID := uuid.New()
	machineID := uuid.New()
	seedProduct(s.T(), s.db, productID, 100, 1)
	seedMachine(s.T(), s.db, machineID, `{"5":10,"25":10}`)

	record := &domain.PurchaseRecord{
		ID:                 uuid.New(),
		MachineID:          machineID,
		ProductID:          productID,
		PriceCents:         100,
		TotalInsertedCents: 125,
		ChangeCents:        25,
		ChangeCoins:        []int32{25},
		PurchasedAt:        time.Now().UTC().Truncate(time.Microsecond),
	}
	bankAfter := domain.CoinBank{5: 10, 25: 11}

	require.NoError(t, s.repo.CommitPurchase(ctx, machineID, productID, record, bankAfter))

	gotProduct, err := s.repo.GetProduct(ctx, productID)
	require.NoError(t, err)
	require.EqualValues(t, 0, gotProduct.Stock, "stock must be decremented")

	gotBank, err := s.repo.GetMachineBank(ctx, machineID)
	require.NoError(t, err)
	require.Equal(t, bankAfter, gotBank, "coin bank must be replaced with bankAfter")

	var count int64
	require.NoError(t, s.db.WithContext(ctx).Table("sales_purchases").Where("id = ?", record.ID).Count(&count).Error)
	assert.EqualValues(t, 1, count, "the purchase must be persisted")
}

// TestCommitPurchase_StockExhaustion sells the last unit, then sells again:
// the second commit loses the guarded stock race and must map to
// domain.ErrOutOfStock with no extra purchase row.
func (s *RepositoryIntegrationSuite) TestCommitPurchase_StockExhaustion() {
	t := s.T()
	ctx := context.Background()

	productID := uuid.New()
	machineID := uuid.New()
	seedProduct(s.T(), s.db, productID, 100, 1)
	seedMachine(s.T(), s.db, machineID, `{"5":10}`)

	newRecord := func() *domain.PurchaseRecord {
		return &domain.PurchaseRecord{
			ID:                 uuid.New(),
			MachineID:          machineID,
			ProductID:          productID,
			PriceCents:         100,
			TotalInsertedCents: 100,
			ChangeCents:        0,
			ChangeCoins:        []int32{},
			PurchasedAt:        time.Now().UTC().Truncate(time.Microsecond),
		}
	}

	require.NoError(t, s.repo.CommitPurchase(ctx, machineID, productID, newRecord(), domain.CoinBank{5: 11}))

	err := s.repo.CommitPurchase(ctx, machineID, productID, newRecord(), domain.CoinBank{5: 11})
	require.ErrorIs(t, err, domain.ErrOutOfStock)

	var count int64
	require.NoError(t, s.db.WithContext(ctx).Table("sales_purchases").Count(&count).Error)
	assert.EqualValues(t, 1, count, "only the winning purchase must be persisted")
}

// TestCommitPurchase_MachineNotFound proves the whole transaction rolls back
// when the machine is missing: the product's stock must be restored.
func (s *RepositoryIntegrationSuite) TestCommitPurchase_MachineNotFound() {
	t := s.T()
	ctx := context.Background()

	productID := uuid.New()
	machineID := uuid.New()
	seedProduct(s.T(), s.db, productID, 100, 1)

	record := &domain.PurchaseRecord{
		ID:                 uuid.New(),
		MachineID:          machineID,
		ProductID:          productID,
		PriceCents:         100,
		TotalInsertedCents: 100,
		ChangeCents:        0,
		ChangeCoins:        []int32{},
		PurchasedAt:        time.Now().UTC().Truncate(time.Microsecond),
	}

	err := s.repo.CommitPurchase(ctx, machineID, productID, record, domain.CoinBank{5: 11})
	require.ErrorIs(t, err, domain.ErrMachineNotFound)

	gotProduct, err := s.repo.GetProduct(ctx, productID)
	require.NoError(t, err)
	require.EqualValues(t, 1, gotProduct.Stock, "the rollback must restore the stock decrement")
}

// runMigrations applies every feature's embedded migrations against cfg's
// database from ONE merged source in version order (catalog, machines, sales).
// A single source matters: golang-migrate reads the next schema_migrations
// version from its own file set, so per-feature sources against one shared
// database fail once the first source advances the shared version. The sales
// tables reference rows written by catalog and machines, so the sales suite
// needs all three schemas present.
func runMigrations(t *testing.T, cfg *config.Config) {
	t.Helper()

	src, err := iofs.New(fsmerge.Merge(
		catalogmigrations.FS,
		machinesmigrations.FS,
		salesmigrations.FS,
	), ".")
	require.NoError(t, err)

	m, err := migrate.NewWithSourceInstance("iofs", src, cfg.DBConnString())
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = m.Close()
	})

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		require.NoError(t, err)
	}
}

// resetSalesState clears the sales, machines, and catalog tables so each case
// starts from an empty, deterministic schema.
func resetSalesState(t *testing.T, db *gorm.DB) {
	t.Helper()

	require.NoError(t, db.WithContext(context.Background()).
		Exec("TRUNCATE TABLE sales_purchases, machines, catalog_products RESTART IDENTITY CASCADE").
		Error)
}

func seedProduct(t *testing.T, db *gorm.DB, id uuid.UUID, priceCents, stock int32) {
	t.Helper()

	require.NoError(t, db.WithContext(context.Background()).
		Exec(`INSERT INTO catalog_products (id, name, price_cents, stock, created_at, updated_at)
		      VALUES (?, 'seed', ?, ?, now(), now())`, id.String(), priceCents, stock).
		Error)
}

func seedMachine(t *testing.T, db *gorm.DB, id uuid.UUID, bankJSON string) {
	t.Helper()

	require.NoError(t, db.WithContext(context.Background()).
		Exec(`INSERT INTO machines (id, label, coin_bank, created_at, updated_at)
		      VALUES (?, 'seed', ?::jsonb, now(), now())`, id.String(), bankJSON).
		Error)
}

func TestRepositoryIntegrationSuite(t *testing.T) {
	suite.Run(t, new(RepositoryIntegrationSuite))
}
