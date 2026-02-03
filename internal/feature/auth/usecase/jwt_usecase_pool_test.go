// Package usecase provides concurrent tests for JWT sync.Pool operations.
package usecase

import (
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestJWTClaimsPoolConcurrent tests thread safety of JWTClaims pool.
func TestJWTClaimsPoolConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	numGoroutines := 100
	iterations := 1000

	// Test concurrent get/put operations
	for range numGoroutines {
		wg.Go(func() {
			for range iterations {
				claims := getJWTClaims()
				// Verify the claims can be used
				claims.UserID = "test-user"
				claims.Email = "test@example.com"
				// Reset before returning to pool
				claims.UserID = ""
				claims.Email = ""
				putJWTClaims(claims)
			}
		})
	}

	wg.Wait()
}

// TestJWTClaimsPoolDataIsolation tests that pooled objects are properly reset.
func TestJWTClaimsPoolDataIsolation(t *testing.T) {
	// Get a claims object from the pool
	claims1 := getJWTClaims()
	claims1.UserID = "user-1"
	claims1.Email = "user1@example.com"

	// Return it to the pool
	putJWTClaims(claims1)

	// Get another claims object - should be reset
	claims2 := getJWTClaims()

	// Verify the claims are reset (not containing previous user's data)
	if claims2.UserID != "" {
		t.Errorf("expected UserID to be empty, got %q", claims2.UserID)
	}
	if claims2.Email != "" {
		t.Errorf("expected Email to be empty, got %q", claims2.Email)
	}

	// Clean up
	putJWTClaims(claims2)
}

// TestJWTClaimsPoolMultipleTypes tests pool with different data scenarios.
func TestJWTClaimsPoolMultipleTypes(t *testing.T) {
	testCases := []struct {
		userID string
		email  string
	}{
		{"user-1", "user1@example.com"},
		{"user-2", "user2@example.com"},
		{"", ""},
		{"very-long-user-id-1234567890", "very-long-email-address@example.com"},
	}

	for _, tc := range testCases {
		claims := getJWTClaims()
		claims.UserID = tc.userID
		claims.Email = tc.email

		// Verify data is set correctly
		if claims.UserID != tc.userID {
			t.Errorf("expected UserID %q, got %q", tc.userID, claims.UserID)
		}
		if claims.Email != tc.email {
			t.Errorf("expected Email %q, got %q", tc.email, claims.Email)
		}

		putJWTClaims(claims)
	}
}

// BenchmarkJWTClaimsPoolWithData tests pool with actual claims data.
func BenchmarkJWTClaimsPoolWithData(b *testing.B) {
	now := time.Now()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		claims := getJWTClaims()
		claims.UserID = "user-123"
		claims.Email = "test@example.com"
		claims.RegisteredClaims = jwt.RegisteredClaims{
			ID:        "token-id",
			Subject:   "user-123",
			Issuer:    "test-issuer",
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		}
		_ = claims
		putJWTClaims(claims)
	}
}
