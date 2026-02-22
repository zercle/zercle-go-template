// Package utils provides utility functions for the application.
package utils

import (
	"errors"
	"testing"
)

func TestGetValidator(t *testing.T) {
	// Test that GetValidator returns a singleton instance
	v1 := GetValidator()
	v2 := GetValidator()

	if v1 == nil {
		t.Error("expected validator instance, got nil")
	}

	if v1 != v2 {
		t.Error("expected GetValidator to return the same instance")
	}
}

func TestValidator_ValidateStruct(t *testing.T) {
	v := GetValidator()

	type TestStruct struct {
		Name  string `validate:"required"`
		Email string `validate:"required,email"`
		Age   int    `validate:"gte=0,lte=150"`
	}

	tests := []struct {
		name    string
		input   TestStruct
		wantErr bool
	}{
		{
			name: "valid struct",
			input: TestStruct{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   30,
			},
			wantErr: false,
		},
		{
			name: "missing required field",
			input: TestStruct{
				Name:  "",
				Email: "john@example.com",
				Age:   30,
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			input: TestStruct{
				Name:  "John Doe",
				Email: "invalid-email",
				Age:   30,
			},
			wantErr: true,
		},
		{
			name: "age too high",
			input: TestStruct{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   200,
			},
			wantErr: true,
		},
		{
			name: "negative age",
			input: TestStruct{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   -5,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateStruct(&tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateVar(t *testing.T) {
	v := GetValidator()

	tests := []struct {
		name    string
		field   any
		tag     string
		wantErr bool
	}{
		{
			name:    "valid email",
			field:   "test@example.com",
			tag:     "email",
			wantErr: false,
		},
		{
			name:    "invalid email",
			field:   "invalid-email",
			tag:     "email",
			wantErr: true,
		},
		{
			name:    "valid min length",
			field:   "hello",
			tag:     "min=3",
			wantErr: false,
		},
		{
			name:    "invalid min length",
			field:   "hi",
			tag:     "min=3",
			wantErr: true,
		},
		{
			name:    "valid required",
			field:   "value",
			tag:     "required",
			wantErr: false,
		},
		{
			name:    "invalid required",
			field:   "",
			tag:     "required",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateVar(tt.field, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVar() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidationErrors(t *testing.T) {
	v := GetValidator()

	type TestStruct struct {
		Name  string `validate:"required"`
		Email string `validate:"required,email"`
	}

	input := TestStruct{
		Name:  "",
		Email: "invalid",
	}

	err := v.ValidateStruct(&input)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	errors := v.ValidationErrors(err)

	if len(errors) == 0 {
		t.Error("expected validation errors, got none")
	}

	// Check for expected fields
	if _, ok := errors["name"]; !ok {
		t.Error("expected 'name' field error")
	}

	if _, ok := errors["email"]; !ok {
		t.Error("expected 'email' field error")
	}
}

func TestValidator_ValidationErrors_NoErrors(t *testing.T) {
	v := GetValidator()

	type TestStruct struct {
		Name string `validate:"required"`
	}

	input := TestStruct{Name: "John"}
	err := v.ValidateStruct(&input)

	errors := v.ValidationErrors(err)

	if len(errors) != 0 {
		t.Errorf("expected no validation errors, got %d", len(errors))
	}
}

func TestValidator_ValidationErrors_NonValidationError(t *testing.T) {
	v := GetValidator()

	nonValidationErr := errors.New("some other error")
	errors := v.ValidationErrors(nonValidationErr)

	if len(errors) != 0 {
		t.Errorf("expected no validation errors for non-validation error, got %d", len(errors))
	}
}

func TestGetErrorMessage(t *testing.T) {
	v := GetValidator()

	type TestStruct struct {
		Required string `validate:"required"`
		Email    string `validate:"email"`
		Min      string `validate:"min=5"`
		Max      string `validate:"max=10"`
		Len      string `validate:"len=5"`
		Alpha    string `validate:"alphanumspace"`
	}

	tests := []struct {
		name     string
		input    TestStruct
		field    string
		expected string
	}{
		{
			name:     "required error",
			input:    TestStruct{},
			field:    "required",
			expected: "This field is required",
		},
		{
			name:     "email error",
			input:    TestStruct{Required: "test", Email: "invalid"},
			field:    "email",
			expected: "Invalid email format",
		},
		{
			name:     "min error",
			input:    TestStruct{Required: "test", Min: "abc"},
			field:    "min",
			expected: "Value is too short, minimum length is 5",
		},
		{
			name:     "max error",
			input:    TestStruct{Required: "test", Max: "this is a very long string"},
			field:    "max",
			expected: "Value is too long, maximum length is 10",
		},
		{
			name:     "len error",
			input:    TestStruct{Required: "test", Len: "abc"},
			field:    "len",
			expected: "Value must be exactly 5 characters",
		},
		{
			name:     "alphanumspace error",
			input:    TestStruct{Required: "test", Alpha: "test@123"},
			field:    "alpha",
			expected: "Value can only contain letters, numbers, and spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateStruct(&tt.input)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}

			errors := v.ValidationErrors(err)
			if _, ok := errors[tt.field]; !ok {
				t.Errorf("expected error for field %s, got errors: %v", tt.field, errors)
			}
		})
	}
}

func TestValidateAlphaNumSpace(t *testing.T) {
	v := GetValidator()

	type TestStruct struct {
		Value string `validate:"alphanumspace"`
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid alphanumeric",
			value:   "Hello123",
			wantErr: false,
		},
		{
			name:    "valid with space",
			value:   "Hello World 123",
			wantErr: false,
		},
		{
			name:    "valid lowercase",
			value:   "abc xyz",
			wantErr: false,
		},
		{
			name:    "valid numbers only",
			value:   "123 456",
			wantErr: false,
		},
		{
			name:    "invalid special char",
			value:   "Hello@World",
			wantErr: true,
		},
		{
			name:    "invalid hyphen",
			value:   "Hello-World",
			wantErr: true,
		},
		{
			name:    "invalid underscore",
			value:   "Hello_World",
			wantErr: true,
		},
		{
			name:    "empty string",
			value:   "",
			wantErr: false, // empty is valid for alphanumspace
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := TestStruct{Value: tt.value}
			err := v.ValidateStruct(&input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsValidationError(t *testing.T) {
	v := GetValidator()

	type TestStruct struct {
		Name string `validate:"required"`
	}

	t.Run("validation error", func(t *testing.T) {
		input := TestStruct{}
		err := v.ValidateStruct(&input)

		if !IsValidationError(err) {
			t.Error("expected IsValidationError to return true for validation error")
		}
	})

	t.Run("non-validation error", func(t *testing.T) {
		err := errors.New("some error")

		if IsValidationError(err) {
			t.Error("expected IsValidationError to return false for non-validation error")
		}
	})

	t.Run("nil error", func(t *testing.T) {
		if IsValidationError(nil) {
			t.Error("expected IsValidationError to return false for nil error")
		}
	})
}

func TestInitFiberValidator(t *testing.T) {
	// This test just ensures InitFiberValidator doesn't panic
	InitFiberValidator()

	// Verify the validator was initialized
	v := GetValidator()
	if v == nil {
		t.Error("expected validator to be initialized")
	}
}

func TestInitGinValidator(t *testing.T) {
	// This test just ensures InitGinValidator doesn't panic (backward compatibility)
	InitGinValidator()

	// Verify the validator was initialized
	v := GetValidator()
	if v == nil {
		t.Error("expected validator to be initialized")
	}
}

func TestValidator_SingletonThreadSafety(t *testing.T) {
	// Test that concurrent calls to GetValidator are safe
	done := make(chan bool, 10)

	for range 10 {
		go func() {
			_ = GetValidator()
			done <- true
		}()
	}

	for range 10 {
		<-done
	}
}

func TestValidator_MultipleValidationCalls(t *testing.T) {
	v := GetValidator()

	type TestStruct struct {
		Name  string `validate:"required"`
		Email string `validate:"required,email"`
	}

	// Test multiple validations
	for i := range 100 {
		input := TestStruct{
			Name:  "John Doe",
			Email: "john@example.com",
		}

		err := v.ValidateStruct(&input)
		if err != nil {
			t.Errorf("unexpected error on iteration %d: %v", i, err)
		}
	}
}

func TestValidationErrors_EmptyStruct(t *testing.T) {
	v := GetValidator()

	type EmptyStruct struct{}

	input := EmptyStruct{}
	err := v.ValidateStruct(&input)

	if err != nil {
		t.Errorf("expected no error for empty struct, got: %v", err)
	}

	errors := v.ValidationErrors(err)
	if len(errors) != 0 {
		t.Errorf("expected no validation errors for empty struct, got %d", len(errors))
	}
}

func TestValidateAlphaNumSpace_EdgeCases(t *testing.T) {
	v := GetValidator()

	type TestStruct struct {
		Value string `validate:"alphanumspace"`
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "only spaces",
			value:   "   ",
			wantErr: false,
		},
		{
			name:    "single char",
			value:   "a",
			wantErr: false,
		},
		{
			name:    "unicode letter",
			value:   "Hello 世界",
			wantErr: true, // unicode not allowed
		},
		{
			name:    "tab character",
			value:   "Hello\tWorld",
			wantErr: true, // tab not allowed
		},
		{
			name:    "newline character",
			value:   "Hello\nWorld",
			wantErr: true, // newline not allowed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := TestStruct{Value: tt.value}
			err := v.ValidateStruct(&input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ComplexStruct(t *testing.T) {
	v := GetValidator()

	type Address struct {
		Street string `validate:"required"`
		City   string `validate:"required"`
	}

	type Person struct {
		Name    string  `validate:"required"`
		Email   string  `validate:"required,email"`
		Address Address `validate:"required"`
	}

	tests := []struct {
		name    string
		input   Person
		wantErr bool
	}{
		{
			name: "valid nested struct",
			input: Person{
				Name:  "John Doe",
				Email: "john@example.com",
				Address: Address{
					Street: "123 Main St",
					City:   "New York",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid nested struct",
			input: Person{
				Name:  "John Doe",
				Email: "john@example.com",
				Address: Address{
					Street: "",
					City:   "New York",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateStruct(&tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
