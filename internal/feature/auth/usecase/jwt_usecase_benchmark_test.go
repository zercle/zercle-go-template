// Package usecase provides benchmark tests for JWT operations.
package usecase

import (
	"context"
	"testing"
	"time"

	"zercle-go-template/internal/config"
	"zercle-go-template/internal/feature/auth/domain"
	"zercle-go-template/internal/logger"
)

// mockTokenUser implements the domain.TokenUser interface for testing.
type mockTokenUser struct {
	id    string
	email string
}

func (m *mockTokenUser) GetID() string    { return m.id }
func (m *mockTokenUser) GetEmail() string { return m.email }

// newTestJWTUsecase creates a JWT usecase for benchmarking.
func newTestJWTUsecase() JWTUsecase {
	cfg := &config.JWTConfig{
		Secret:          "benchmark-secret-key-for-testing-only",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}
	return NewJWTUsecase(cfg, logger.NewNop())
}

// BenchmarkGenerateTokenPair measures the performance of generating JWT token pairs.
func BenchmarkGenerateTokenPair(b *testing.B) {
	usecase := newTestJWTUsecase()
	user := &mockTokenUser{
		id:    "user-123",
		email: "test@example.com",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := usecase.GenerateTokenPair(user)
		if err != nil {
			b.Fatalf("GenerateTokenPair failed: %v", err)
		}
	}
}

// BenchmarkValidateToken measures the performance of validating JWT tokens.
func BenchmarkValidateToken(b *testing.B) {
	usecase := newTestJWTUsecase()
	user := &mockTokenUser{
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

// BenchmarkGenerateAccessToken measures the performance of generating access tokens from claims.
func BenchmarkGenerateAccessToken(b *testing.B) {
	usecase := newTestJWTUsecase()
	user := &mockTokenUser{
		id:    "user-123",
		email: "test@example.com",
	}

	tokenPair, err := usecase.GenerateTokenPair(user)
	if err != nil {
		b.Fatalf("Setup: GenerateTokenPair failed: %v", err)
	}

	claims, err := usecase.ValidateToken(tokenPair.AccessToken)
	if err != nil {
		b.Fatalf("Setup: ValidateToken failed: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := usecase.GenerateAccessToken(claims)
		if err != nil {
			b.Fatalf("GenerateAccessToken failed: %v", err)
		}
	}
}

// BenchmarkTokenRoundTrip measures the full round-trip of generating and validating tokens.
func BenchmarkTokenRoundTrip(b *testing.B) {
	usecase := newTestJWTUsecase()
	user := &mockTokenUser{
		id:    "user-123",
		email: "test@example.com",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Generate token pair
		tokenPair, err := usecase.GenerateTokenPair(user)
		if err != nil {
			b.Fatalf("GenerateTokenPair failed: %v", err)
		}

		// Validate access token
		_, err = usecase.ValidateToken(tokenPair.AccessToken)
		if err != nil {
			b.Fatalf("ValidateToken failed: %v", err)
		}
	}
}

// BenchmarkValidateTokenParallel measures token validation performance under concurrent load.
func BenchmarkValidateTokenParallel(b *testing.B) {
	usecase := newTestJWTUsecase()
	user := &mockTokenUser{
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

// BenchmarkGenerateTokenPairParallel measures token generation performance under concurrent load.
func BenchmarkGenerateTokenPairParallel(b *testing.B) {
	usecase := newTestJWTUsecase()
	user := &mockTokenUser{
		id:    "user-123",
		email: "test@example.com",
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := usecase.GenerateTokenPair(user)
			if err != nil {
				b.Fatalf("GenerateTokenPair failed: %v", err)
			}
		}
	})
}

// BenchmarkGenerateTokenPairDifferentUsers measures token generation for different users.
func BenchmarkGenerateTokenPairDifferentUsers(b *testing.B) {
	usecase := newTestJWTUsecase()

	// Pre-create users
	users := make([]*mockTokenUser, 100)
	for i := range users {
		users[i] = &mockTokenUser{
			id:    string(rune('a' + (i % 26))),
			email: "user@example.com",
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		user := users[i%len(users)]
		_, err := usecase.GenerateTokenPair(user)
		if err != nil {
			b.Fatalf("GenerateTokenPair failed: %v", err)
		}
	}
}

// mockBenchmarkLogger implements a minimal logger for benchmarking.
// This avoids the overhead of the full logger implementation.
type mockBenchmarkLogger struct{}

func (m *mockBenchmarkLogger) Debug(msg string, fields ...logger.Field)        {}
func (m *mockBenchmarkLogger) Info(msg string, fields ...logger.Field)         {}
func (m *mockBenchmarkLogger) Warn(msg string, fields ...logger.Field)         {}
func (m *mockBenchmarkLogger) Error(msg string, fields ...logger.Field)        {}
func (m *mockBenchmarkLogger) Fatal(msg string, fields ...logger.Field)        {}
func (m *mockBenchmarkLogger) WithContext(ctx context.Context) logger.Logger   { return m }
func (m *mockBenchmarkLogger) WithFields(fields ...logger.Field) logger.Logger { return m }

// Ensure mockLogger implements the interface
var _ logger.Logger = (*mockBenchmarkLogger)(nil)

// BenchmarkGenerateTokenPairNoLogging measures token generation without logging overhead.
func BenchmarkGenerateTokenPairNoLogging(b *testing.B) {
	cfg := &config.JWTConfig{
		Secret:          "benchmark-secret-key-for-testing-only",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}
	usecase := NewJWTUsecase(cfg, &mockBenchmarkLogger{})

	user := &mockTokenUser{
		id:    "user-123",
		email: "test@example.com",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := usecase.GenerateTokenPair(user)
		if err != nil {
			b.Fatalf("GenerateTokenPair failed: %v", err)
		}
	}
}

// BenchmarkJWTClaimsPoolGet measures the performance of getting a claims object from the pool.
func BenchmarkJWTClaimsPoolGet(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		claims := getJWTClaims()
		putJWTClaims(claims)
	}
}

// BenchmarkJWTClaimsDirectAllocation measures the performance of direct JWTClaims allocation.
func BenchmarkJWTClaimsDirectAllocation(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		claims := &domain.JWTClaims{}
		_ = claims
	}
}

// BenchmarkJWTClaimsPoolParallel measures the pool under concurrent load.
func BenchmarkJWTClaimsPoolParallel(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			claims := getJWTClaims()
			putJWTClaims(claims)
		}
	})
}
