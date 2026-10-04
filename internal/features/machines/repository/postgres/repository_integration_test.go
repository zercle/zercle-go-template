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
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	catalogmigrations "github.com/zercle/zercle-go-template/internal/features/catalog/repository/postgres/migrations"
	"github.com/zercle/zercle-go-template/internal/features/machines/domain"
	"github.com/zercle/zercle-go-template/internal/features/machines/repository/postgres"
	machinesmigrations "github.com/zercle/zercle-go-template/internal/features/machines/repository/postgres/migrations"
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
	resetMachinesState(s.T(), s.db)
}

func (s *RepositoryIntegrationSuite) TestCreateAndGetByID() {
	t := s.T()
	ctx := context.Background()

	machine := &domain.Machine{
		ID:        uuid.New(),
		Label:     "integration-lobby",
		CoinBank:  domain.CoinBank{5: 3, 25: 2},
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
		UpdatedAt: time.Now().UTC().Truncate(time.Microsecond),
	}

	require.NoError(t, s.repo.Create(ctx, machine))

	got, err := s.repo.GetByID(ctx, machine.ID)
	require.NoError(t, err)
	require.Equal(t, machine.ID, got.ID)
	require.Equal(t, machine.Label, got.Label)
	require.Equal(t, machine.CoinBank, got.CoinBank)
}

func (s *RepositoryIntegrationSuite) TestGetByID_NotFound() {
	t := s.T()

	got, err := s.repo.GetByID(context.Background(), uuid.New())
	require.Nil(t, got)
	require.True(t, errors.Is(err, domain.ErrMachineNotFound))
}

func (s *RepositoryIntegrationSuite) TestList() {
	t := s.T()
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		machine := &domain.Machine{
			ID:        uuid.New(),
			Label:     "list-lobby",
			CoinBank:  domain.CoinBank{5: int32(i)},
			CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
			UpdatedAt: time.Now().UTC().Truncate(time.Microsecond),
		}
		require.NoError(t, s.repo.Create(ctx, machine))
	}

	machines, err := s.repo.List(ctx, 10, 0)
	require.NoError(t, err)
	require.Len(t, machines, 3)
}

// TestRestockBank adds coins and reads the machine back: the bank must hold the
// sum of the seeded coins and the restocked coins.
func (s *RepositoryIntegrationSuite) TestRestockBank() {
	t := s.T()
	ctx := context.Background()

	machine := &domain.Machine{
		ID:        uuid.New(),
		Label:     "restock-lobby",
		CoinBank:  domain.CoinBank{5: 1, 25: 2},
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
		UpdatedAt: time.Now().UTC().Truncate(time.Microsecond),
	}
	require.NoError(t, s.repo.Create(ctx, machine))

	require.NoError(t, s.repo.RestockBank(ctx, machine.ID, []int32{5, 100}))

	got, err := s.repo.GetByID(ctx, machine.ID)
	require.NoError(t, err)
	want := domain.AddCoins(machine.CoinBank, []int32{5, 100})
	require.Equal(t, want, got.CoinBank)
}

func (s *RepositoryIntegrationSuite) TestRestockBank_MachineNotFound() {
	t := s.T()

	err := s.repo.RestockBank(context.Background(), uuid.New(), []int32{5})
	require.ErrorIs(t, err, domain.ErrMachineNotFound)
}

// runMigrations applies every feature's embedded migrations against cfg's
// database from ONE merged source in version order (catalog, machines, sales).
// A single source matters: golang-migrate reads the next schema_migrations
// version from its own file set, so per-feature sources against one shared
// database fail once the first source advances the shared version.
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

// resetMachinesState clears the machines table so each case starts empty.
func resetMachinesState(t *testing.T, db *gorm.DB) {
	t.Helper()

	require.NoError(t, db.WithContext(context.Background()).
		Exec("TRUNCATE TABLE machines RESTART IDENTITY CASCADE").
		Error)
}

func TestRepositoryIntegrationSuite(t *testing.T) {
	suite.Run(t, new(RepositoryIntegrationSuite))
}
