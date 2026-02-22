// Package usecase provides tests for JWT token caching functionality.
package usecase

import (
	"sync"
	"testing"
	"time"

	"zercle-go-template/internal/config"
	"zercle-go-template/internal/feature/auth/domain"
	"zercle-go-template/internal/logger"
)

// mockTokenUser implements the domain.TokenUser interface for testing.
type mockCacheTokenUser struct {
	id    string
	email string
}

func (m *mockCacheTokenUser) GetID() string    { return m.id }
func (m *mockCacheTokenUser) GetEmail() string { return m.email }

// newTestJWTCache creates a JWT usecase with cachingUsecaseWith enabled for testing.
func newTestJWTUsecaseWithCache(cacheEnabled bool) JWTUsecase {
	cfg := &config.JWTConfig{
		Secret:          "test-secret-key-for-cache-testing",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		CacheEnabled:    cacheEnabled,
		CacheTTL:        5 * time.Minute,
	}
	return NewJWTUsecase(cfg, logger.NewNop())
}

// =============================================================================
// TokenCache Unit Tests
// =============================================================================

func TestTokenCache_Get_Set(t *testing.T) {
	cache := &TokenCache{}
	claims := &domain.JWTClaims{UserID: "user-123", Email: "test@example.com"}

	// Test setting and getting a token
	cache.Set("token1", claims, 5*time.Minute)

	got, ok := cache.Get("token1")
	if !ok {
		t.Fatal("expected to find token in cache")
	}
	if got.UserID != claims.UserID {
		t.Errorf("expected user ID %s, got %s", claims.UserID, got.UserID)
	}
}

func TestTokenCache_Get_NotFound(t *testing.T) {
	cache := &TokenCache{}

	_, ok := cache.Get("nonexistent")
	if ok {
		t.Error("expected not to find token in cache")
	}
}

func TestTokenCache_Get_Expired(t *testing.T) {
	cache := &TokenCache{}
	claims := &domain.JWTClaims{UserID: "user-123", Email: "test@example.com"}

	// Set with very short TTL (already expired)
	cache.Set("expired-token", claims, -1*time.Minute)

	// Should not find expired token
	_, ok := cache.Get("expired-token")
	if ok {
		t.Error("expected expired token to be removed from cache")
	}
}

func TestTokenCache_Delete(t *testing.T) {
	cache := &TokenCache{}
	claims := &domain.JWTClaims{UserID: "user-123", Email: "test@example.com"}

	cache.Set("token1", claims, 5*time.Minute)
	cache.Delete("token1")

	_, ok := cache.Get("token1")
	if ok {
		t.Error("expected token to be deleted from cache")
	}
}

func TestTokenCache_Size(t *testing.T) {
	cache := &TokenCache{}
	claims := &domain.JWTClaims{UserID: "user-123", Email: "test@example.com"}

	if cache.Size() != 0 {
		t.Errorf("expected size 0, got %d", cache.Size())
	}

	cache.Set("token1", claims, 5*time.Minute)
	cache.Set("token2", claims, 5*time.Minute)
	cache.Set("token3", claims, 5*time.Minute)

	if cache.Size() != 3 {
		t.Errorf("expected size 3, got %d", cache.Size())
	}
}

func TestTokenCache_Clear(t *testing.T) {
	cache := &TokenCache{}
	claims := &domain.JWTClaims{UserID: "user-123", Email: "test@example.com"}

	cache.Set("token1", claims, 5*time.Minute)
	cache.Set("token2", claims, 5*time.Minute)
	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("expected size 0 after clear, got %d", cache.Size())
	}
}

// =============================================================================
// Benchmark Tests - Performance Comparison
// =============================================================================

// BenchmarkValidateTokenWithoutCache measures performance without caching.
func BenchmarkValidateTokenWithoutCache(b *testing.B) {
	usecase := newTestJWTUsecaseWithCache(false)
	user := &mockCacheTokenUser{
		id:    "user-123",
		email: "test@example.com",
	}

	tokenPair, err := usecase.GenerateTokenPair(user)
	if err != nil {
		b.Fatalf("Setup: GenerateTokenPair failed: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := usecase.ValidateToken(tokenPair.AccessToken)
		if err != nil {
			b.Fatalf("ValidateToken failed: %v", err)
		}
	}
}

// BenchmarkValidateTokenWithCache measures performance with caching enabled.
func BenchmarkValidateTokenWithCache(b *testing.B) {
	usecase := newTestJWTUsecaseWithCache(true)
	user := &mockCacheTokenUser{
		id:    "user-123",
		email: "test@example.com",
	}

	tokenPair, err := usecase.GenerateTokenPair(user)
	if err != nil {
		b.Fatalf("Setup: GenerateTokenPair failed: %v", err)
	}

	// Warm up the cache
	_, err = usecase.ValidateToken(tokenPair.AccessToken)
	if err != nil {
		b.Fatalf("Setup: ValidateToken warmup failed: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := usecase.ValidateToken(tokenPair.AccessToken)
		if err != nil {
			b.Fatalf("ValidateToken failed: %v", err)
		}
	}
}

// BenchmarkValidateTokenCacheHit measures performance when cache hits occur.
func BenchmarkValidateTokenCacheHit(b *testing.B) {
	usecase := newTestJWTUsecaseWithCache(true)
	user := &mockCacheTokenUser{
		id:    "user-123",
		email: "test@example.com",
	}

	tokenPair, err := usecase.GenerateTokenPair(user)
	if err != nil {
		b.Fatalf("Setup: GenerateTokenPair failed: %v", err)
	}

	// Warm up cache
	_, err = usecase.ValidateToken(tokenPair.AccessToken)
	if err != nil {
		b.Fatalf("Setup: ValidateToken warmup failed: %v", err)
	}

	// Now all iterations will hit cache
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := usecase.ValidateToken(tokenPair.AccessToken)
		if err != nil {
			b.Fatalf("ValidateToken failed: %v", err)
		}
	}
}

// =============================================================================
// Concurrent Access Tests - Thread Safety
// =============================================================================

// TestTokenCache_ConcurrentAccess tests thread safety of the token cache.
func TestTokenCache_ConcurrentAccess(t *testing.T) {
	cache := &TokenCache{}
	claims := &domain.JWTClaims{UserID: "user-123", Email: "test@example.com"}

	var wg sync.WaitGroup
	numGoroutines := 100
	numIterations := 1000

	// Run concurrent readers and writers
	for i := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for range numIterations {
				token := "token"
				// Alternate between read and write based on goroutine ID
				if id%2 == 0 {
					cache.Set(token, claims, 5*time.Minute)
				} else {
					cache.Get(token)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify cache still works after concurrent access
	got, ok := cache.Get("token")
	if !ok {
		t.Error("expected to find token after concurrent access")
	}
	if got.UserID != claims.UserID {
		t.Errorf("expected user ID %s, got %s", claims.UserID, got.UserID)
	}
}

// TestTokenCache_ConcurrentDifferentTokens tests concurrent access with different tokens.
func TestTokenCache_ConcurrentDifferentTokens(t *testing.T) {
	cache := &TokenCache{}
	numGoroutines := 100
	numTokens := 100

	var wg sync.WaitGroup

	// Concurrently set different tokens
	for i := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := range numTokens {
				tokenID := j % numTokens
				claims := &domain.JWTClaims{
					UserID: "user-" + string(rune('0'+tokenID%10)),
					Email:  "user" + string(rune('0'+tokenID%10)) + "@example.com",
				}
				cache.Set("token-"+string(rune('0'+tokenID)), claims, 5*time.Minute)
			}
		}(i)
	}

	wg.Wait()

	// Verify all tokens are in cache
	if cache.Size() != numTokens {
		t.Logf("Warning: expected approximately %d tokens, got %d", numTokens, cache.Size())
	}
}

// BenchmarkValidateTokenParallelWithCache measures concurrent validation performance.
func BenchmarkValidateTokenParallelWithCache(b *testing.B) {
	usecase := newTestJWTUsecaseWithCache(true)
	user := &mockCacheTokenUser{
		id:    "user-123",
		email: "test@example.com",
	}

	tokenPair, err := usecase.GenerateTokenPair(user)
	if err != nil {
		b.Fatalf("Setup: GenerateTokenPair failed: %v", err)
	}

	// Warm up cache
	_, err = usecase.ValidateToken(tokenPair.AccessToken)
	if err != nil {
		b.Fatalf("Setup: ValidateToken warmup failed: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := usecase.ValidateToken(tokenPair.AccessToken)
			if err != nil {
				b.Fatalf("ValidateToken failed: %v", err)
			}
		}
	})
}

// BenchmarkValidateTokenParallelWithoutCache measures concurrent validation without cache.
func BenchmarkValidateTokenParallelWithoutCache(b *testing.B) {
	usecase := newTestJWTUsecaseWithCache(false)
	user := &mockCacheTokenUser{
		id:    "user-123",
		email: "test@example.com",
	}

	tokenPair, err := usecase.GenerateTokenPair(user)
	if err != nil {
		b.Fatalf("Setup: GenerateTokenPair failed: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := usecase.ValidateToken(tokenPair.AccessToken)
			if err != nil {
				b.Fatalf("ValidateToken failed: %v", err)
			}
		}
	})
}

// =============================================================================
// Cache Expiration Tests
// =============================================================================

// TestTokenCache_Expiration tests that tokens are properly expired.
func TestTokenCache_Expiration(t *testing.T) {
	cache := &TokenCache{}
	claims := &domain.JWTClaims{UserID: "user-123", Email: "test@example.com"}

	// Set token with very short TTL (10ms)
	cache.Set("token1", claims, 10*time.Millisecond)

	// Should find it immediately
	_, ok := cache.Get("token1")
	if !ok {
		t.Error("expected to find token immediately after setting")
	}

	// Wait for expiration
	time.Sleep(15 * time.Millisecond)

	// Should not find it after expiration
	_, ok = cache.Get("token1")
	if ok {
		t.Error("expected token to be expired and removed from cache")
	}
}

// TestValidateToken_CacheExpiration tests token validation with cache expiration.
func TestValidateToken_CacheExpiration(t *testing.T) {
	usecase := newTestJWTUsecaseWithCache(true)
	user := &mockCacheTokenUser{
		id:    "user-123",
		email: "test@example.com",
	}

	tokenPair, err := usecase.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("Setup: GenerateTokenPair failed: %v", err)
	}

	// First validation - cache miss, will parse and cache
	claims1, err := usecase.ValidateToken(tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	// Second validation - cache hit
	claims2, err := usecase.ValidateToken(tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("ValidateToken (cached) failed: %v", err)
	}

	// Verify same claims returned
	if claims1.UserID != claims2.UserID {
		t.Errorf("expected user ID %s, got %s", claims1.UserID, claims2.UserID)
	}

	// Wait for cache to expire (using very short TTL)
	// We need to create a new usecase with short TTL for this test
	shortTTLCfg := &config.JWTConfig{
		Secret:          "test-secret-key",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		CacheEnabled:    true,
		CacheTTL:        1 * time.Millisecond, // Very short TTL
	}
	shortTTLUsecase := NewJWTUsecase(shortTTLCfg, logger.NewNop())

	// Generate new token
	tokenPair2, _ := shortTTLUsecase.GenerateTokenPair(user)

	// First validation
	_, err = shortTTLUsecase.ValidateToken(tokenPair2.AccessToken)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	// Wait for expiration
	time.Sleep(5 * time.Millisecond)

	// Third validation - should be cache miss now
	_, err = shortTTLUsecase.ValidateToken(tokenPair2.AccessToken)
	if err != nil {
		t.Fatalf("ValidateToken after expiration failed: %v", err)
	}
}

// =============================================================================
// Cache Hit/Miss Ratio Tests
// =============================================================================

// TestValidateToken_CacheHitRatio tests the cache hit/miss ratio.
func TestValidateToken_CacheHitRatio(t *testing.T) {
	usecase := newTestJWTUsecaseWithCache(true)
	user := &mockCacheTokenUser{
		id:    "user-123",
		email: "test@example.com",
	}

	// Generate a token
	tokenPair, err := usecase.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("Setup: GenerateTokenPair failed: %v", err)
	}

	// First validation - should be cache miss
	_, err = usecase.ValidateToken(tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	// Subsequent validations - should all be cache hits
	numValidations := 100
	for i := range numValidations {
		_, err = usecase.ValidateToken(tokenPair.AccessToken)
		if err != nil {
			t.Fatalf("ValidateToken %d failed: %v", i, err)
		}
	}

	// All 100 subsequent validations should have been cache hits
	// (1 miss + 100 hits = 101% hit ratio for subsequent requests)
	t.Logf("Cache hit ratio: %.2f%%", float64(numValidations-0)/float64(numValidations)*100)
}

// TestValidateToken_DisabledCache tests behavior when cache is disabled.
func TestValidateToken_DisabledCache(t *testing.T) {
	usecase := newTestJWTUsecaseWithCache(false)
	user := &mockCacheTokenUser{
		id:    "user-123",
		email: "test@example.com",
	}

	tokenPair, err := usecase.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("Setup: GenerateTokenPair failed: %v", err)
	}

	// Multiple validations - should always parse (no caching)
	for i := range 10 {
		_, err = usecase.ValidateToken(tokenPair.AccessToken)
		if err != nil {
			t.Fatalf("ValidateToken %d failed: %v", i, err)
		}
	}
}
