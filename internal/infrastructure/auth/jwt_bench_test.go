package auth

import (
	"context"
	"testing"

	"github.com/zercle/zercle-go-template/pkg/config"
)

func createBenchTokenService() TokenService {
	cfg := config.AuthConfig{
		AccessTokenSecret:  "bench-access-token-secret-key-32b",
		RefreshTokenSecret: "bench-refresh-token-secret-key-32b",
		AccessTokenTTL:     15 * 60,
		RefreshTokenTTL:    7 * 24 * 60 * 60,
		Issuer:             "bench-issuer",
	}
	return NewTokenService(cfg)
}

func BenchmarkGenerateTokenPair(b *testing.B) {
	svc := createBenchTokenService()
	ctx := context.Background()
	userID := "bench-user-123"
	email := "bench@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.GenerateTokenPair(ctx, userID, email)
	}
}

func BenchmarkValidateAccessToken(b *testing.B) {
	svc := createBenchTokenService()
	ctx := context.Background()

	tokenPair, _ := svc.GenerateTokenPair(ctx, "bench-user-123", "bench@example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.ValidateAccessToken(ctx, tokenPair.AccessToken)
	}
}

func BenchmarkValidateAccessToken_ValidToken(b *testing.B) {
	svc := createBenchTokenService()
	ctx := context.Background()

	tokenPair, _ := svc.GenerateTokenPair(ctx, "bench-user-123", "bench@example.com")
	token := tokenPair.AccessToken

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.ValidateAccessToken(ctx, token)
	}
}

func BenchmarkValidateRefreshToken(b *testing.B) {
	svc := createBenchTokenService()
	ctx := context.Background()

	tokenPair, _ := svc.GenerateTokenPair(ctx, "bench-user-123", "bench@example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.ValidateRefreshToken(ctx, tokenPair.RefreshToken)
	}
}

func BenchmarkHashToken(b *testing.B) {
	token := "bench-token-string-for-hashing-purposes"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = HashToken(token)
	}
}

func BenchmarkHashToken_LongToken(b *testing.B) {
	token := "this-is-a-very-long-token-string-that-might-be-used-in-production-scenarios-with-oauth2-or-oidc-tokens"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = HashToken(token)
	}
}

func BenchmarkCompareToken(b *testing.B) {
	token := "bench-token-for-comparison"
	hash := HashToken(token)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CompareToken(token, hash)
	}
}

func BenchmarkGenerateTokenPair_Parallel(b *testing.B) {
	svc := createBenchTokenService()
	ctx := context.Background()
	userID := "bench-user-123"
	email := "bench@example.com"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = svc.GenerateTokenPair(ctx, userID, email)
		}
	})
}

func BenchmarkValidateAccessToken_Parallel(b *testing.B) {
	svc := createBenchTokenService()
	ctx := context.Background()

	tokenPair, _ := svc.GenerateTokenPair(ctx, "bench-user-123", "bench@example.com")
	token := tokenPair.AccessToken

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = svc.ValidateAccessToken(ctx, token)
		}
	})
}
