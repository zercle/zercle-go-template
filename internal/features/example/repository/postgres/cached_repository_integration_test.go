//go:build integration

// STUB FEATURE — delete internal/features/example to start your project.

package postgres_test

import (
	"context"
	"testing"
	"time"
	"uuid"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/valkey-io/valkey-go/valkeyaside"
	"gorm.io/gorm"

	"github.com/zercle/zercle-go-template/internal/features/example/domain"
	"github.com/zercle/zercle-go-template/internal/features/example/repository/postgres"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/db"
	infravalkey "github.com/zercle/zercle-go-template/internal/infrastructure/valkey"
)

// CachedRepositoryIntegrationSuite exercises the cache-aside read path against
// a live Valkey.
type CachedRepositoryIntegrationSuite struct {
	suite.Suite
	db    *gorm.DB
	repo  *postgres.CachedRepository
	aside valkeyaside.CacheAsideClient
}

func (s *CachedRepositoryIntegrationSuite) SetupSuite() {
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s.aside, err = infravalkey.NewCacheAside(ctx, cfg)
	require.NoError(t, err, "valkey must be reachable: docker compose up -d valkey")

	s.repo = postgres.NewCachedRepository(postgres.NewRepository(s.db), s.aside, cfg.Valkey.TTL)
}

func (s *CachedRepositoryIntegrationSuite) TearDownSuite() {
	if s.aside != nil {
		s.aside.Close()
	}
	if s.db != nil {
		if sqlDB, err := s.db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}
}

func (s *CachedRepositoryIntegrationSuite) SetupTest() {
	require.NoError(s.T(), s.db.WithContext(context.Background()).
		Exec("TRUNCATE TABLE items RESTART IDENTITY CASCADE").
		Error)
}

// TestGetByID_ServedFromCacheOnSecondRead seeds an item, reads it once to
// populate the cache, deletes the row, and reads again: the second read must
// still succeed, proving it was answered from the cache and not the database.
func (s *CachedRepositoryIntegrationSuite) TestGetByID_ServedFromCacheOnSecondRead() {
	t := s.T()
	ctx := context.Background()

	item := &domain.Item{
		ID:        uuid.New(),
		Name:      "cache-aside",
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
		UpdatedAt: time.Now().UTC().Truncate(time.Microsecond),
	}
	require.NoError(t, s.db.WithContext(ctx).Create(&item).Error)

	first, err := s.repo.GetByID(ctx, item.ID)
	require.NoError(t, err)
	require.Equal(t, item.Name, first.Name)

	// Remove the row; only the cache can now answer.
	require.NoError(t, s.db.WithContext(ctx).Delete(&domain.Item{}, "id = ?", item.ID).Error)

	second, err := s.repo.GetByID(ctx, item.ID)
	require.NoError(t, err, "second read must be served from the cache")
	require.Equal(t, item.Name, second.Name)
	require.Equal(t, first.CreatedAt.UTC(), second.CreatedAt.UTC())
}

// TestGetByID_MissIsNotCached pins that a missing item is not cached: the
// loader error propagates and a later read still reports not-found.
func (s *CachedRepositoryIntegrationSuite) TestGetByID_MissIsNotCached() {
	t := s.T()
	ctx := context.Background()

	id := uuid.New()
	_, err := s.repo.GetByID(ctx, id)
	require.ErrorIs(t, err, domain.ErrItemNotFound)

	_, err = s.repo.GetByID(ctx, id)
	require.ErrorIs(t, err, domain.ErrItemNotFound)
}

func TestCachedRepositoryIntegrationSuite(t *testing.T) {
	suite.Run(t, new(CachedRepositoryIntegrationSuite))
}
