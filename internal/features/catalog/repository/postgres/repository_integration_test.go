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

	"github.com/zercle/zercle-go-template/internal/features/catalog/domain"
	"github.com/zercle/zercle-go-template/internal/features/catalog/repository/postgres"
	catalogmigrations "github.com/zercle/zercle-go-template/internal/features/catalog/repository/postgres/migrations"
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
	resetCatalogState(s.T(), s.db)
}

func (s *RepositoryIntegrationSuite) TestCreateAndGetByID() {
	t := s.T()
	ctx := context.Background()

	product := &domain.Product{
		ID:         uuid.New(),
		Name:       "integration-cola",
		PriceCents: 150,
		Stock:      4,
		CreatedAt:  time.Now().UTC().Truncate(time.Microsecond),
		UpdatedAt:  time.Now().UTC().Truncate(time.Microsecond),
	}

	require.NoError(t, s.repo.Create(ctx, product))

	got, err := s.repo.GetByID(ctx, product.ID)
	require.NoError(t, err)
	require.Equal(t, product.ID, got.ID)
	require.Equal(t, product.Name, got.Name)
	require.EqualValues(t, product.PriceCents, got.PriceCents)
	require.EqualValues(t, product.Stock, got.Stock)
}

func (s *RepositoryIntegrationSuite) TestGetByID_NotFound() {
	t := s.T()

	got, err := s.repo.GetByID(context.Background(), uuid.New())
	require.Nil(t, got)
	require.True(t, errors.Is(err, domain.ErrProductNotFound))
}

func (s *RepositoryIntegrationSuite) TestList() {
	t := s.T()
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		product := &domain.Product{
			ID:         uuid.New(),
			Name:       "list-cola",
			PriceCents: 100,
			Stock:      int32(i),
			CreatedAt:  time.Now().UTC().Truncate(time.Microsecond),
			UpdatedAt:  time.Now().UTC().Truncate(time.Microsecond),
		}
		require.NoError(t, s.repo.Create(ctx, product))
	}

	products, err := s.repo.List(ctx, 10, 0)
	require.NoError(t, err)
	require.Len(t, products, 3)
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

// resetCatalogState clears the catalog table so each case starts empty.
func resetCatalogState(t *testing.T, db *gorm.DB) {
	t.Helper()

	require.NoError(t, db.WithContext(context.Background()).
		Exec("TRUNCATE TABLE catalog_products RESTART IDENTITY CASCADE").
		Error)
}

func TestRepositoryIntegrationSuite(t *testing.T) {
	suite.Run(t, new(RepositoryIntegrationSuite))
}
