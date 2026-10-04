package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"uuid"

	"github.com/valkey-io/valkey-go/valkeyaside"

	"github.com/zercle/zercle-go-template/internal/features/catalog/domain"
	"github.com/zercle/zercle-go-template/internal/features/catalog/repository"
)

// productCachePrefix namespaces cache-aside keys for the catalog feature.
const productCachePrefix = "catalog:product:"

// CachedRepository decorates a repository.Repository with a Valkey cache-aside
// read path. It satisfies the same interface, so the usecase layer is unaware
// of caching.
//
// Only GetByID is cached. List results are not, because a page depends on the
// whole table's write history and a stale page would misreport global stock.
type CachedRepository struct {
	repo  repository.Repository
	aside valkeyaside.CacheAsideClient
	ttl   time.Duration
}

// NewCachedRepository decorates repo with cache-aside reads through aside.
func NewCachedRepository(repo repository.Repository, aside valkeyaside.CacheAsideClient, ttl time.Duration) *CachedRepository {
	return &CachedRepository{repo: repo, aside: aside, ttl: ttl}
}

// Create persists a new product. A newly created ID cannot have a cached entry,
// so no invalidation is needed; List results are intentionally not cached.
func (r *CachedRepository) Create(ctx context.Context, product *domain.Product) error {
	if err := r.repo.Create(ctx, product); err != nil {
		return fmt.Errorf("create product: %w", err)
	}
	return nil
}

// GetByID serves the product from the cache-aside store, loading it from the
// database on a miss. A missing product is not cached: the loader returns
// domain.ErrProductNotFound, which leaves the entry empty and releases the lock.
func (r *CachedRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	val, err := r.aside.Get(ctx, r.ttl, productCachePrefix+id.String(), func(ctx context.Context, _ string) (string, error) {
		product, err := r.repo.GetByID(ctx, id)
		if err != nil {
			return "", fmt.Errorf("get product: %w", err)
		}
		encoded, err := json.Marshal(cachedProduct{
			ID:         product.ID.String(),
			Name:       product.Name,
			PriceCents: product.PriceCents,
			Stock:      product.Stock,
			CreatedAt:  product.CreatedAt,
			UpdatedAt:  product.UpdatedAt,
		})
		if err != nil {
			return "", fmt.Errorf("encode cached product: %w", err)
		}
		return string(encoded), nil
	})
	if err != nil {
		return nil, fmt.Errorf("cache-aside get product %s: %w", id, err)
	}

	var cached cachedProduct
	if err := json.Unmarshal([]byte(val), &cached); err != nil {
		return nil, fmt.Errorf("decode cached product: %w", err)
	}
	return cached.toDomain()
}

// List always reads through to the database.
func (r *CachedRepository) List(ctx context.Context, limit, offset int32) ([]domain.Product, error) {
	products, err := r.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	return products, nil
}

// cachedProduct is the wire shape stored in Valkey. It keeps the cached
// contract independent of both the domain entity and the HTTP contract.
type cachedProduct struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	PriceCents int32     `json:"price_cents"`
	Stock      int32     `json:"stock"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (c cachedProduct) toDomain() (*domain.Product, error) {
	id, err := uuid.Parse(c.ID)
	if err != nil {
		return nil, fmt.Errorf("parse cached product id: %w", err)
	}
	return &domain.Product{
		ID:         id,
		Name:       c.Name,
		PriceCents: c.PriceCents,
		Stock:      c.Stock,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
	}, nil
}
