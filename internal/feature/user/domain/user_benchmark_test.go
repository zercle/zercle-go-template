// Package domain provides benchmark tests for user domain operations.
package domain

import (
	"fmt"
	"testing"
)

// BenchmarkSetPasswordDefaultParams measures password hashing with default Argon2id parameters.
func BenchmarkSetPasswordDefaultParams(b *testing.B) {
	user := &User{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Test User",
	}

	password := "securePassword123!"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Create a new user for each iteration to avoid state issues
		testUser := &User{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
		}
		err := testUser.SetPassword(password)
		if err != nil {
			b.Fatalf("SetPassword failed: %v", err)
		}
	}
}

// BenchmarkSetPasswordLowMemory measures password hashing with lower memory (16 MB).
func BenchmarkSetPasswordLowMemory(b *testing.B) {
	// Set lower memory for faster benchmarking (16 MB)
	SetArgon2Params(16*1024, 3, 4, 16, 32)
	defer SetArgon2Params(defaultMemory, defaultIterations, defaultParallelism, defaultSaltLength, defaultKeyLength)

	password := "securePassword123!"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		user := &User{
			ID:    "user-123",
			Email: "test@example.com",
			Name:  "Test User",
		}
		err := user.SetPassword(password)
		if err != nil {
			b.Fatalf("SetPassword failed: %v", err)
		}
	}
}

// BenchmarkSetPasswordHighMemory measures password hashing with higher memory (128 MB).
func BenchmarkSetPasswordHighMemory(b *testing.B) {
	// Set higher memory for stronger security (128 MB)
	SetArgon2Params(128*1024, 3, 4, 16, 32)
	defer SetArgon2Params(defaultMemory, defaultIterations, defaultParallelism, defaultSaltLength, defaultKeyLength)

	password := "securePassword123!"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		user := &User{
			ID:    "user-123",
			Email: "test@example.com",
			Name:  "Test User",
		}
		err := user.SetPassword(password)
		if err != nil {
			b.Fatalf("SetPassword failed: %v", err)
		}
	}
}

// BenchmarkVerifyPassword measures password verification performance.
func BenchmarkVerifyPassword(b *testing.B) {
	user := &User{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Test User",
	}

	password := "securePassword123!"
	err := user.SetPassword(password)
	if err != nil {
		b.Fatalf("Setup: SetPassword failed: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		valid := user.VerifyPassword(password)
		if !valid {
			b.Fatalf("VerifyPassword returned false for correct password")
		}
	}
}

// BenchmarkVerifyPasswordWrong measures password verification with wrong password.
func BenchmarkVerifyPasswordWrong(b *testing.B) {
	user := &User{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Test User",
	}

	password := "securePassword123!"
	err := user.SetPassword(password)
	if err != nil {
		b.Fatalf("Setup: SetPassword failed: %v", err)
	}

	wrongPassword := "wrongPassword123!"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		valid := user.VerifyPassword(wrongPassword)
		if valid {
			b.Fatalf("VerifyPassword returned true for wrong password")
		}
	}
}

// BenchmarkNewUser measures the performance of creating a new user.
func BenchmarkNewUser(b *testing.B) {
	email := "test@example.com"
	name := "Test User"
	password := "securePassword123!"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := NewUser(email, name, password)
		if err != nil {
			b.Fatalf("NewUser failed: %v", err)
		}
	}
}

// BenchmarkIsValidEmailValid measures email validation with valid emails.
func BenchmarkIsValidEmailValid(b *testing.B) {
	emails := []string{
		"test@example.com",
		"user.name@domain.co.uk",
		"user+tag@example.org",
		"first.last@company.io",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		email := emails[i%len(emails)]
		valid := IsValidEmail(email)
		if !valid {
			b.Fatalf("IsValidEmail returned false for valid email: %s", email)
		}
	}
}

// BenchmarkIsValidEmailInvalid measures email validation with invalid emails.
func BenchmarkIsValidEmailInvalid(b *testing.B) {
	emails := []string{
		"",
		"invalid",
		"@example.com",
		"test@",
		"test@.com",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		email := emails[i%len(emails)]
		valid := IsValidEmail(email)
		if valid {
			b.Fatalf("IsValidEmail returned true for invalid email: %s", email)
		}
	}
}

// BenchmarkUserValidate measures user validation performance.
func BenchmarkUserValidate(b *testing.B) {
	user := &User{
		ID:           "user-123",
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "hashed_password_here",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := user.Validate()
		if err != nil {
			b.Fatalf("Validate failed: %v", err)
		}
	}
}

// BenchmarkUserUpdate measures user update performance.
func BenchmarkUserUpdate(b *testing.B) {
	user := &User{
		ID:           "user-123",
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "hashed_password_here",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Update with a different name each time to ensure the operation is performed
		user.Update(fmt.Sprintf("Updated User %d", i))
	}
}

// BenchmarkArgon2idParamsComparison compares different Argon2id parameter sets.
func BenchmarkArgon2idParamsComparison(b *testing.B) {
	password := "benchmarkPassword123!"

	// Different parameter sets for comparison
	paramSets := []struct {
		name        string
		memory      int
		iterations  int
		parallelism uint8
	}{
		{"Low_16MB_3iter", 16 * 1024, 3, 4},
		{"Default_64MB_3iter", 64 * 1024, 3, 4},
		{"High_128MB_3iter", 128 * 1024, 3, 4},
		{"High_64MB_4iter", 64 * 1024, 4, 4},
		{"Ultra_256MB_3iter", 256 * 1024, 3, 4},
	}

	for _, ps := range paramSets {
		b.Run(ps.name, func(b *testing.B) {
			SetArgon2Params(ps.memory, ps.iterations, ps.parallelism, 16, 32)
			defer SetArgon2Params(defaultMemory, defaultIterations, defaultParallelism, defaultSaltLength, defaultKeyLength)

			user := &User{
				ID:    "user-123",
				Email: "test@example.com",
				Name:  "Test User",
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				err := user.SetPassword(password)
				if err != nil {
					b.Fatalf("SetPassword failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkSetPasswordParallel measures password hashing under concurrent load.
func BenchmarkSetPasswordParallel(b *testing.B) {
	password := "securePassword123!"

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			user := &User{
				ID:    "user-123",
				Email: "test@example.com",
				Name:  "Test User",
			}
			err := user.SetPassword(password)
			if err != nil {
				b.Fatalf("SetPassword failed: %v", err)
			}
		}
	})
}

// BenchmarkVerifyPasswordParallel measures password verification under concurrent load.
func BenchmarkVerifyPasswordParallel(b *testing.B) {
	user := &User{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Test User",
	}

	password := "securePassword123!"
	err := user.SetPassword(password)
	if err != nil {
		b.Fatalf("Setup: SetPassword failed: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			valid := user.VerifyPassword(password)
			if !valid {
				b.Fatalf("VerifyPassword returned false for correct password")
			}
		}
	})
}
