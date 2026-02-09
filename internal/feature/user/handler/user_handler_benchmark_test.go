// Package handler provides benchmark tests for user handler operations.
package handler

import (
	"testing"

	"zercle-go-template/internal/feature/user/dto"
)

// BenchmarkValidateStructCreateUserRequest measures validation performance for CreateUserRequest.
func BenchmarkValidateStructCreateUserRequest(b *testing.B) {
	req := &dto.CreateUserRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "securePassword123!",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		errors := validateStruct(req)
		if errors != nil {
			b.Fatalf("validateStruct returned errors for valid request: %v", errors)
		}
	}
}

// BenchmarkValidateStructCreateUserRequestInvalid measures validation with invalid data.
func BenchmarkValidateStructCreateUserRequestInvalid(b *testing.B) {
	req := &dto.CreateUserRequest{
		Email:    "invalid-email",
		Name:     "A",     // Too short
		Password: "short", // Too short
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		errors := validateStruct(req)
		if errors == nil {
			b.Fatalf("validateStruct should return errors for invalid request")
		}
	}
}

// BenchmarkValidateStructLoginRequest measures validation performance for UserLoginRequest.
func BenchmarkValidateStructLoginRequest(b *testing.B) {
	req := &dto.UserLoginRequest{
		Email:    "test@example.com",
		Password: "securePassword123!",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		errors := validateStruct(req)
		if errors != nil {
			b.Fatalf("validateStruct returned errors for valid request: %v", errors)
		}
	}
}

// BenchmarkValidateStructUpdateUserRequest measures validation performance for UpdateUserRequest.
func BenchmarkValidateStructUpdateUserRequest(b *testing.B) {
	req := &dto.UpdateUserRequest{
		Name: "Updated User Name",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		errors := validateStruct(req)
		if errors != nil {
			b.Fatalf("validateStruct returned errors for valid request: %v", errors)
		}
	}
}

// BenchmarkValidateStructUpdatePasswordRequest measures validation performance for UpdatePasswordRequest.
func BenchmarkValidateStructUpdatePasswordRequest(b *testing.B) {
	req := &dto.UpdatePasswordRequest{
		OldPassword: "oldSecurePassword123!",
		NewPassword: "newSecurePassword123!",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		errors := validateStruct(req)
		if errors != nil {
			b.Fatalf("validateStruct returned errors for valid request: %v", errors)
		}
	}
}

// BenchmarkValidateStructParallel measures validation performance under concurrent load.
func BenchmarkValidateStructParallel(b *testing.B) {
	req := &dto.CreateUserRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "securePassword123!",
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			errors := validateStruct(req)
			if errors != nil {
				b.Fatalf("validateStruct returned errors for valid request: %v", errors)
			}
		}
	})
}

// BenchmarkGetErrorMessageRequired measures getErrorMessage performance for "required" tag.
func BenchmarkGetErrorMessageRequired(b *testing.B) {
	// We can't easily create a validator.FieldError without using the validator,
	// so we'll benchmark the full validation which exercises getErrorMessage
	req := &dto.CreateUserRequest{
		Email:    "test@example.com",
		Name:     "", // Empty to trigger required error
		Password: "securePassword123!",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		validateStruct(req)
	}
}

// BenchmarkGetErrorMessageEmail measures getErrorMessage performance for "email" tag.
func BenchmarkGetErrorMessageEmail(b *testing.B) {
	req := &dto.CreateUserRequest{
		Email:    "invalid-email",
		Name:     "Test User",
		Password: "securePassword123!",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		validateStruct(req)
	}
}

// BenchmarkGetErrorMessageMin measures getErrorMessage performance for "min" tag.
func BenchmarkGetErrorMessageMin(b *testing.B) {
	req := &dto.CreateUserRequest{
		Email:    "test@example.com",
		Name:     "A", // Too short
		Password: "securePassword123!",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		validateStruct(req)
	}
}

// BenchmarkFormatValidationErrors measures formatValidationErrors performance.
func BenchmarkFormatValidationErrors(b *testing.B) {
	errors := map[string]string{
		"email":    "Invalid email format",
		"name":     "Value is too short, minimum length is 2",
		"password": "Value is too short, minimum length is 8",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = formatValidationErrors(errors)
	}
}

// BenchmarkFormatValidationErrorsSingle measures formatValidationErrors with single error.
func BenchmarkFormatValidationErrorsSingle(b *testing.B) {
	errors := map[string]string{
		"email": "Invalid email format",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = formatValidationErrors(errors)
	}
}

// BenchmarkFormatValidationErrorsEmpty measures formatValidationErrors with empty map.
func BenchmarkFormatValidationErrorsEmpty(b *testing.B) {
	errors := map[string]string{}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = formatValidationErrors(errors)
	}
}

// BenchmarkGetValidator measures the performance of getting the singleton validator.
func BenchmarkGetValidator(b *testing.B) {
	// Warm up the singleton
	_ = getValidator()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = getValidator()
	}
}

// BenchmarkValidateStructDifferentRequests measures validation with different request types.
func BenchmarkValidateStructDifferentRequests(b *testing.B) {
	requests := []any{
		&dto.CreateUserRequest{
			Email:    "test1@example.com",
			Name:     "Test User 1",
			Password: "securePassword123!",
		},
		&dto.CreateUserRequest{
			Email:    "test2@example.com",
			Name:     "Test User 2",
			Password: "anotherPassword456!",
		},
		&dto.UserLoginRequest{
			Email:    "login@example.com",
			Password: "loginPassword123!",
		},
		&dto.UpdateUserRequest{
			Name: "Updated Name",
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := requests[i%len(requests)]
		errors := validateStruct(req)
		if errors != nil {
			b.Fatalf("validateStruct returned errors for valid request: %v", errors)
		}
	}
}
