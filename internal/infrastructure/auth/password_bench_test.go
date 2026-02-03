package auth

import (
	"testing"
)

func BenchmarkPasswordHash(b *testing.B) {
	hasher := NewPasswordHasher()
	password := "benchPassword123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = hasher.Hash(password)
	}
}

func BenchmarkPasswordHash_VariousLengths(b *testing.B) {
	hasher := NewPasswordHasher()
	testCases := []struct {
		name     string
		password string
	}{
		{"short", "pass"},
		{"medium", "mediumLengthPassword123"},
		{"long", "thisIsAVeryLongPasswordThatExceedsTypicalLengthsAndTestsPerformance"},
		{"veryLong", "thisIsAnExtremelyLongPasswordThatIsUnlikelyToBeUsedButTestsPerformanceBoundsWithExcessiveLength"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = hasher.Hash(tc.password)
			}
		})
	}
}

func BenchmarkPasswordVerify(b *testing.B) {
	hasher := NewPasswordHasher()
	password := "benchPassword123"
	hash, _ := hasher.Hash(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hasher.Verify(password, hash)
	}
}

func BenchmarkPasswordVerify_CorrectPassword(b *testing.B) {
	hasher := NewPasswordHasher()
	password := "benchPassword123"
	hash, _ := hasher.Hash(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hasher.Verify(password, hash)
	}
}

func BenchmarkPasswordVerify_WrongPassword(b *testing.B) {
	hasher := NewPasswordHasher()
	password := "benchPassword123"
	wrongPassword := "wrongPassword456"
	hash, _ := hasher.Hash(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hasher.Verify(wrongPassword, hash)
	}
}

func BenchmarkPasswordHash_WithCustomConfig(b *testing.B) {
	config := &Argon2Config{
		Memory:      32 * 1024,
		Iterations:  1,
		Parallelism: 2,
		SaltLength:  8,
		KeyLength:   16,
	}
	hasher := NewPasswordHasherWithConfig(config)
	password := "benchPassword123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = hasher.Hash(password)
	}
}

func BenchmarkPasswordVerify_WithCustomConfig(b *testing.B) {
	config := &Argon2Config{
		Memory:      32 * 1024,
		Iterations:  1,
		Parallelism: 2,
		SaltLength:  8,
		KeyLength:   16,
	}
	hasher := NewPasswordHasherWithConfig(config)
	password := "benchPassword123"
	hash, _ := hasher.Hash(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hasher.Verify(password, hash)
	}
}

func BenchmarkPasswordHash_Parallel(b *testing.B) {
	hasher := NewPasswordHasher()
	password := "benchPassword123"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = hasher.Hash(password)
		}
	})
}

func BenchmarkPasswordVerify_Parallel(b *testing.B) {
	hasher := NewPasswordHasher()
	password := "benchPassword123"
	hash, _ := hasher.Hash(password)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = hasher.Verify(password, hash)
		}
	})
}
