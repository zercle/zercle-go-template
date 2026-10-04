//go:build unit

package postgres_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/zercle/zercle-go-template/internal/features/catalog/domain"
	"github.com/zercle/zercle-go-template/internal/features/catalog/repository/mock"
	"github.com/zercle/zercle-go-template/internal/features/catalog/repository/postgres"
	"github.com/zercle/zercle-go-template/internal/testutil"
)

const testTTL = time.Minute

// productCacheKey is the key the decorator is expected to build for a product id.
func productCacheKey(id uuid.UUID) string { return "catalog:product:" + id.String() }

// preMarshaledProduct mirrors the repository's wire shape (json tags are the
// contract under test) so a cache-hit case can supply an encoded payload
// without reaching into the unexported cachedProduct type.
func preMarshaledProduct(t *testing.T, product domain.Product) string {
	t.Helper()
	encoded, err := json.Marshal(struct {
		ID         string    `json:"id"`
		Name       string    `json:"name"`
		PriceCents int32     `json:"price_cents"`
		Stock      int32     `json:"stock"`
		CreatedAt  time.Time `json:"created_at"`
		UpdatedAt  time.Time `json:"updated_at"`
	}{product.ID.String(), product.Name, product.PriceCents, product.Stock, product.CreatedAt, product.UpdatedAt})
	require.NoError(t, err)
	return string(encoded)
}

// TestCachedRepository_GetByID_MissRunsLoader proves the miss path: with a
// pass-through cache the wrapped repository is called, the loader's value is
// decoded back into a domain product, and the decorator addressed the cache
// with the configured ttl and the namespaced id key.
func TestCachedRepository_GetByID_MissRunsLoader(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	repo := mock.NewMockRepository(ctrl)

	id := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)
	repo.EXPECT().GetByID(ctx, id).Return(&domain.Product{ID: id, Name: "from-db", PriceCents: 250, Stock: 4, CreatedAt: now, UpdatedAt: now}, nil)

	var (
		gotTTL time.Duration
		gotKey string
	)
	aside := &testutil.FakeCacheAside{
		GetFn: func(ctx context.Context, ttl time.Duration, key string, fn func(context.Context, string) (string, error)) (string, error) {
			gotTTL, gotKey = ttl, key
			return fn(ctx, key) // pass-through miss
		},
	}

	got, err := postgres.NewCachedRepository(repo, aside, testTTL).GetByID(ctx, id)

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, id, got.ID)
	assert.Equal(t, "from-db", got.Name)
	assert.EqualValues(t, 250, got.PriceCents)
	assert.EqualValues(t, 4, got.Stock)
	assert.Equal(t, now, got.CreatedAt)
	assert.Equal(t, testTTL, gotTTL, "decorator must pass the configured ttl to the cache")
	assert.Equal(t, productCacheKey(id), gotKey, "decorator must namespace the key with the product prefix")
}

// TestCachedRepository_GetByID_HitSkipsRepository proves the hit path: a
// pre-populated cache answers without touching the wrapped repository (gomock
// fails the test on any unexpected call), and the payload decodes to the
// original product.
func TestCachedRepository_GetByID_HitSkipsRepository(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	repo := mock.NewMockRepository(ctrl) // no EXPECT: any call is a failure

	id := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)
	product := domain.Product{ID: id, Name: "from-cache", PriceCents: 100, Stock: 7, CreatedAt: now, UpdatedAt: now}
	aside := &testutil.FakeCacheAside{
		GetFn: func(context.Context, time.Duration, string, func(context.Context, string) (string, error)) (string, error) {
			return preMarshaledProduct(t, product), nil
		},
	}

	got, err := postgres.NewCachedRepository(repo, aside, testTTL).GetByID(ctx, id)

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, id, got.ID)
	assert.Equal(t, "from-cache", got.Name)
	assert.EqualValues(t, 7, got.Stock)
}

// TestCachedRepository_GetByID_LoaderNotFound proves a missing product is not
// masked: the loader's domain.ErrProductNotFound reaches the caller through the
// decorator's wrapping.
func TestCachedRepository_GetByID_LoaderNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	repo := mock.NewMockRepository(ctrl)

	id := uuid.New()
	repo.EXPECT().GetByID(ctx, id).Return(nil, domain.ErrProductNotFound)

	got, err := postgres.NewCachedRepository(repo, &testutil.FakeCacheAside{}, testTTL).GetByID(ctx, id)

	require.ErrorIs(t, err, domain.ErrProductNotFound)
	assert.Nil(t, got)
}

// TestCachedRepository_Passthrough proves the uncached operations forward
// verbatim to the wrapped repository.
func TestCachedRepository_Passthrough(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	repo := mock.NewMockRepository(ctrl)

	product := &domain.Product{ID: uuid.New(), Name: "created"}
	repo.EXPECT().Create(ctx, product).Return(nil)

	products := []domain.Product{{ID: uuid.New(), Name: "listed"}}
	repo.EXPECT().List(ctx, int32(10), int32(0)).Return(products, nil)

	cached := postgres.NewCachedRepository(repo, &testutil.FakeCacheAside{}, testTTL)

	require.NoError(t, cached.Create(ctx, product))

	got, err := cached.List(ctx, 10, 0)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "listed", got[0].Name)
}
