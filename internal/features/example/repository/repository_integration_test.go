//go:build integration

// STUB FEATURE — delete internal/features/example to start your project.

package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // postgres driver
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/zercle/zercle-go-template/internal/config"
	"github.com/zercle/zercle-go-template/internal/features/example/domain"
	"github.com/zercle/zercle-go-template/internal/features/example/repository"
	"github.com/zercle/zercle-go-template/internal/infrastructure/db"
	"github.com/zercle/zercle-go-template/internal/infrastructure/db/migrations"
)

type RepositoryIntegrationSuite struct {
	suite.Suite
	db   *gorm.DB
	repo *repository.Repository
}

func (s *RepositoryIntegrationSuite) SetupSuite() {
	t := s.T()
	t.Helper()

	cfg, err := config.Load()
	require.NoError(t, err)

	nop := zerolog.Nop()
	s.db, err = db.NewDB(context.Background(), cfg, &nop)
	require.NoError(t, err)

	s.runMigrations(cfg)

	s.repo = repository.NewRepository(s.db)
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
	t := s.T()
	t.Helper()

	err := s.db.WithContext(context.Background()).
		Exec("TRUNCATE TABLE items RESTART IDENTITY CASCADE").
		Error
	require.NoError(t, err)
}

func (s *RepositoryIntegrationSuite) TestCreateAndGetByID() {
	t := s.T()
	ctx := context.Background()

	item := &domain.Item{
		ID:        uuid.New(),
		Name:      "integration-item",
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
		UpdatedAt: time.Now().UTC().Truncate(time.Microsecond),
	}

	err := s.repo.Create(ctx, item)
	require.NoError(t, err)

	got, err := s.repo.GetByID(ctx, item.ID)
	require.NoError(t, err)
	require.Equal(t, item.ID, got.ID)
	require.Equal(t, item.Name, got.Name)
}

func (s *RepositoryIntegrationSuite) TestGetByID_NotFound() {
	t := s.T()
	ctx := context.Background()

	got, err := s.repo.GetByID(ctx, uuid.New())
	require.Nil(t, got)
	require.True(t, errors.Is(err, domain.ErrItemNotFound))
}

func (s *RepositoryIntegrationSuite) TestList() {
	t := s.T()
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		item := &domain.Item{
			ID:        uuid.New(),
			Name:      "list-item",
			CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
			UpdatedAt: time.Now().UTC().Truncate(time.Microsecond),
		}
		require.NoError(t, s.repo.Create(ctx, item))
	}

	items, err := s.repo.List(ctx, 10, 0)
	require.NoError(t, err)
	require.Len(t, items, 3)
}

func (s *RepositoryIntegrationSuite) runMigrations(cfg *config.Config) {
	t := s.T()
	t.Helper()

	src, err := iofs.New(migrations.FS, ".")
	require.NoError(t, err)

	m, err := migrate.NewWithSourceInstance("iofs", src, cfg.DBConnString())
	require.NoError(t, err)
	s.T().Cleanup(func() {
		_, _ = m.Close()
	})

	err = m.Up()
	if err == nil {
		return
	}

	if errors.Is(err, migrate.ErrNoChange) {
		return
	}
	require.NoError(t, err)
}

func TestRepositoryIntegrationSuite(t *testing.T) {
	suite.Run(t, new(RepositoryIntegrationSuite))
}
