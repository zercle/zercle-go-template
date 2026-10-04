package testutil

import (
	"context"
	"time"

	valkeygo "github.com/valkey-io/valkey-go"
)

// FakeCacheAside is a scriptable valkeyaside.CacheAsideClient for unit tests of
// cache-aside consumers (for example postgres.CachedRepository) without a live
// Valkey.
//
// A nil GetFn makes Get behave as a cache miss: it runs the loader and returns
// its values, which is the pass-through path most tests want. Set GetFn to
// assert the ttl/key the consumer passed or to simulate a cache hit with a
// pre-encoded payload.
type FakeCacheAside struct {
	GetFn func(
		ctx context.Context,
		ttl time.Duration,
		key string,
		fn func(ctx context.Context, key string) (string, error),
	) (string, error)
}

// Get delegates to GetFn when set, otherwise it invokes fn as a cache miss.
func (f *FakeCacheAside) Get(ctx context.Context, ttl time.Duration, key string, fn func(ctx context.Context, key string) (string, error)) (string, error) {
	if f.GetFn != nil {
		return f.GetFn(ctx, ttl, key, fn)
	}
	return fn(ctx, key)
}

// Del is a no-op: unit tests observe writes through the wrapped repository, not
// through cache invalidation.
func (f *FakeCacheAside) Del(context.Context, string) error { return nil }

// Client is unused by cache-aside consumers under test.
func (f *FakeCacheAside) Client() valkeygo.Client { return nil }

// Close is a no-op: the fake owns no resources.
func (f *FakeCacheAside) Close() {}
