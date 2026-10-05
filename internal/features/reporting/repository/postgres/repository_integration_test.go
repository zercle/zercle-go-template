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
	"github.com/zercle/zercle-go-template/internal/features/reporting/repository/postgres"
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
	resetReportingState(s.T(), s.db)
}

func (s *RepositoryIntegrationSuite) TestGetOverview_Empty() {
	t := s.T()

	got, err := s.repo.GetOverview(context.Background())
	require.NoError(t, err)
	assert.Zero(t, got.ProductCount)
	assert.Zero(t, got.TotalStock)
	assert.Zero(t, got.MachineCount)
	assert.Zero(t, got.TotalBankCents)
	assert.Zero(t, got.PurchaseCount)
	assert.Zero(t, got.RevenueCents)
}

func (s *RepositoryIntegrationSuite) TestGetOverview_Aggregates() {
	t := s.T()
	ctx := context.Background()

	seedProduct(s.T(), s.db, uuid.New(), 100, 3)
	seedProduct(s.T(), s.db, uuid.New(), 250, 5)

	machineID := uuid.New()
	// Coin banks are objects keyed by denomination in cents: 5*10 + 25*4 = 150
	// and 100*2 = 200, so the report's total bank is 350.
	seedMachine(s.T(), s.db, machineID, `{"5":10,"25":4}`)
	seedMachine(s.T(), s.db, uuid.New(), `{"100":2}`)

	seedPurchase(s.T(), s.db, machineID, 100)
	seedPurchase(s.T(), s.db, machineID, 150)

	got, err := s.repo.GetOverview(ctx)
	require.NoError(t, err)
	assert.EqualValues(t, 2, got.ProductCount)
	assert.EqualValues(t, 8, got.TotalStock)
	assert.EqualValues(t, 2, got.MachineCount)
	assert.EqualValues(t, 350, got.TotalBankCents)
	assert.EqualValues(t, 2, got.PurchaseCount)
	assert.EqualValues(t, 250, got.RevenueCents)
}

func (s *RepositoryIntegrationSuite) TestGetTopMachines_OrdersByRevenue() {
	t := s.T()
	ctx := context.Background()

	m1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	m2 := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	seedMachine(s.T(), s.db, m1, `{}`)
	seedMachine(s.T(), s.db, m2, `{}`)

	seedPurchase(s.T(), s.db, m1, 100)
	seedPurchase(s.T(), s.db, m1, 100)
	seedPurchase(s.T(), s.db, m2, 500)

	got, err := s.repo.GetTopMachines(ctx, 10)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, m2, got[0].MachineID)
	assert.EqualValues(t, 500, got[0].RevenueCents)
	assert.Equal(t, m1, got[1].MachineID)
	assert.EqualValues(t, 200, got[1].RevenueCents)
}

// TestGetTopMachines_TieBreakById pins deterministic ordering: equal revenue
// and equal purchase count fall back to the machine id ascending, so the same
// report is returned on every read.
func (s *RepositoryIntegrationSuite) TestGetTopMachines_TieBreakById() {
	t := s.T()
	ctx := context.Background()

	m1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	m2 := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	seedMachine(s.T(), s.db, m1, `{}`)
	seedMachine(s.T(), s.db, m2, `{}`)

	seedPurchase(s.T(), s.db, m1, 100)
	seedPurchase(s.T(), s.db, m1, 100)
	seedPurchase(s.T(), s.db, m2, 100)
	seedPurchase(s.T(), s.db, m2, 100)

	got, err := s.repo.GetTopMachines(ctx, 10)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, m1, got[0].MachineID, "equal rows must order by id ascending")
	assert.Equal(t, m2, got[1].MachineID)
}

func (s *RepositoryIntegrationSuite) TestGetTopMachines_ExcludesMachinesWithoutSales() {
	t := s.T()
	ctx := context.Background()

	seller := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	idle := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	seedMachine(s.T(), s.db, seller, `{}`)
	seedMachine(s.T(), s.db, idle, `{}`)
	seedPurchase(s.T(), s.db, seller, 100)

	got, err := s.repo.GetTopMachines(ctx, 10)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, seller, got[0].MachineID)
	assert.EqualValues(t, 1, got[0].PurchaseCount)
}

func (s *RepositoryIntegrationSuite) TestGetTopMachines_RespectsLimit() {
	t := s.T()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		machineID := uuid.New()
		seedMachine(s.T(), s.db, machineID, `{}`)
		seedPurchase(s.T(), s.db, machineID, 100)
	}

	got, err := s.repo.GetTopMachines(ctx, 2)
	require.NoError(t, err)
	assert.Len(t, got, 2)
}

// runMigrations applies every feature's embedded migrations against cfg's
// database from ONE merged source in version order (catalog, machines, sales).
// Reporting owns no schema; it only reads the tables the other features create.
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

// resetReportingState clears the three tables the report reads so each case
// starts from an empty, deterministic schema.
func resetReportingState(t *testing.T, db *gorm.DB) {
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

func seedPurchase(t *testing.T, db *gorm.DB, machineID uuid.UUID, priceCents int32) {
	t.Helper()

	require.NoError(t, db.WithContext(context.Background()).
		Exec(`INSERT INTO sales_purchases
		        (id, machine_id, product_id, price_cents, total_inserted_cents, change_cents, change_coins, purchased_at)
		      VALUES (?, ?, ?, ?, ?, 0, '[]'::jsonb, ?)`,
			uuid.New().String(), machineID.String(), uuid.New().String(), priceCents, priceCents, time.Now().UTC()).
		Error)
}

func TestRepositoryIntegrationSuite(t *testing.T) {
	suite.Run(t, new(RepositoryIntegrationSuite))
}
