// STUB FEATURE — delete internal/features/example to start your project.

package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/valkey-io/valkey-go/valkeyaside"

	"github.com/zercle/zercle-go-template/internal/features/example/domain"
)

// itemCachePrefix namespaces cache-aside keys for the example feature.
const itemCachePrefix = "example:item:"

// CachedRepository decorates a *Repository with a Valkey cache-aside read path.
// It satisfies the same port, so the application layer is unaware of caching.
//
// The read path is stampede-safe via valkeyaside: on a miss only one caller runs
// the loader while the rest wait for the cache-aside lock to release, and the
// entry is invalidated by Valkey's client-side cache tracking on write. Only the
// by-ID read is cached; List results are not, since a page depends on the whole
// table's write history.
type CachedRepository struct {
	repo  *Repository
	aside valkeyaside.CacheAsideClient
	ttl   time.Duration
}

// NewCachedRepository decorates repo with cache-aside reads through aside.
func NewCachedRepository(repo *Repository, aside valkeyaside.CacheAsideClient, ttl time.Duration) *CachedRepository {
	return &CachedRepository{repo: repo, aside: aside, ttl: ttl}
}

// Create persists a new item. A newly created ID cannot have a cached entry, so
// no invalidation is needed; List results are intentionally not cached.
func (r *CachedRepository) Create(ctx context.Context, item *domain.Item) error {
	return r.repo.Create(ctx, item)
}

// GetByID serves the item from the cache-aside store, loading it from the
// database on a miss. A missing item is not cached: the loader returns
// domain.ErrItemNotFound, which leaves the entry empty and releases the lock.
func (r *CachedRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Item, error) {
	val, err := r.aside.Get(ctx, r.ttl, itemCachePrefix+id.String(), func(ctx context.Context, _ string) (string, error) {
		item, err := r.repo.GetByID(ctx, id)
		if err != nil {
			return "", err
		}
		encoded, err := json.Marshal(cachedItem{
			ID:        item.ID.String(),
			Name:      item.Name,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		})
		if err != nil {
			return "", fmt.Errorf("encode cached item: %w", err)
		}
		return string(encoded), nil
	})
	if err != nil {
		return nil, fmt.Errorf("cache-aside get item %s: %w", id, err)
	}

	var cached cachedItem
	if err := json.Unmarshal([]byte(val), &cached); err != nil {
		return nil, fmt.Errorf("decode cached item: %w", err)
	}
	return cached.toDomain()
}

// List always reads through to the database.
func (r *CachedRepository) List(ctx context.Context, limit, offset int32) ([]domain.Item, error) {
	return r.repo.List(ctx, limit, offset)
}

// cachedItem is the wire shape stored in Valkey. It keeps the cached contract
// independent of both the domain entity and the HTTP contract.
type cachedItem struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (c cachedItem) toDomain() (*domain.Item, error) {
	id, err := uuid.Parse(c.ID)
	if err != nil {
		return nil, fmt.Errorf("parse cached item id: %w", err)
	}
	return &domain.Item{
		ID:        id,
		Name:      c.Name,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}, nil
}
