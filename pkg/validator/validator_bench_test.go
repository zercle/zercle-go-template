// Package validator provides request validation using go-playground/validator.
package validator

import (
	"testing"
)

type benchStruct struct {
	Name     string `json:"name" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Age      int    `json:"age" validate:"gte=0,lte=150"`
	Password string `json:"password" validate:"required,min=8"`
}

func BenchmarkValidateStruct_Valid(b *testing.B) {
	v := New()
	input := benchStruct{
		Name:     "John Doe",
		Email:    "john@example.com",
		Age:      30,
		Password: "securepassword",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.ValidateStruct(input)
	}
}

func BenchmarkValidateStruct_Invalid(b *testing.B) {
	v := New()
	input := benchStruct{
		Name:     "Jo", // Too short
		Email:    "invalid-email",
		Age:      -1,
		Password: "123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.ValidateStruct(input)
	}
}

func BenchmarkValidateVar_Email(b *testing.B) {
	v := New()
	email := "test@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.ValidateVar(email, "email")
	}
}

func BenchmarkValidateVar_Required(b *testing.B) {
	v := New()
	value := "some value"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.ValidateVar(value, "required")
	}
}

func BenchmarkValidateVar_UUID(b *testing.B) {
	v := New()
	uuid := "550e8400-e29b-41d4-a716-446655440000"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.ValidateVar(uuid, "uuid")
	}
}

func BenchmarkNewValidator(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New()
	}
}
