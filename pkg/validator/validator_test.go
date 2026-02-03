// Package validator provides request validation using go-playground/validator.
package validator

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/domain"
)

type TestStruct struct {
	Name     string `json:"name" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Age      int    `json:"age" validate:"gte=0,lte=150"`
	Password string `json:"password" validate:"required,min=8"`
}

type OptionalFields struct {
	Name  string `json:"name" validate:"omitempty,min=3"`
	Email string `json:"email" validate:"omitempty,email"`
}

func TestNew(t *testing.T) {
	v := New()
	assert.NotNil(t, v)
	assert.NotNil(t, v.validate)
}

func TestValidateStruct_ValidData(t *testing.T) {
	v := New()

	testCases := []struct {
		name  string
		input TestStruct
	}{
		{
			name: "valid struct",
			input: TestStruct{
				Name:     "John Doe",
				Email:    "john@example.com",
				Age:      30,
				Password: "securepassword",
			},
		},
		{
			name: "valid struct with edge values",
			input: TestStruct{
				Name:     "ABC",
				Email:    "test@test.org",
				Age:      0,
				Password: "password",
			},
		},
		{
			name: "valid struct with maximum age",
			input: TestStruct{
				Name:     "John Doe",
				Email:    "user@domain.co.uk",
				Age:      150,
				Password: "securepassword",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.ValidateStruct(tc.input)
			assert.NoError(t, err)
		})
	}
}

func TestValidateStruct_InvalidData(t *testing.T) {
	v := New()

	testCases := []struct {
		name          string
		input         TestStruct
		expectedError bool
	}{
		{
			name: "empty name",
			input: TestStruct{
				Name:     "",
				Email:    "john@example.com",
				Age:      30,
				Password: "securepassword",
			},
			expectedError: true,
		},
		{
			name: "name too short",
			input: TestStruct{
				Name:     "Jo",
				Email:    "john@example.com",
				Age:      30,
				Password: "securepassword",
			},
			expectedError: true,
		},
		{
			name: "name too long",
			input: TestStruct{
				Name:     "This is a very long name that exceeds the maximum allowed length of fifty characters",
				Email:    "john@example.com",
				Age:      30,
				Password: "securepassword",
			},
			expectedError: true,
		},
		{
			name: "invalid email",
			input: TestStruct{
				Name:     "John Doe",
				Email:    "invalid-email",
				Age:      30,
				Password: "securepassword",
			},
			expectedError: true,
		},
		{
			name: "empty email",
			input: TestStruct{
				Name:     "John Doe",
				Email:    "",
				Age:      30,
				Password: "securepassword",
			},
			expectedError: true,
		},
		{
			name: "negative age",
			input: TestStruct{
				Name:     "John Doe",
				Email:    "john@example.com",
				Age:      -1,
				Password: "securepassword",
			},
			expectedError: true,
		},
		{
			name: "age too high",
			input: TestStruct{
				Name:     "John Doe",
				Email:    "john@example.com",
				Age:      151,
				Password: "securepassword",
			},
			expectedError: true,
		},
		{
			name: "password too short",
			input: TestStruct{
				Name:     "John Doe",
				Email:    "john@example.com",
				Age:      30,
				Password: "short",
			},
			expectedError: true,
		},
		{
			name: "empty password",
			input: TestStruct{
				Name:     "John Doe",
				Email:    "john@example.com",
				Age:      30,
				Password: "",
			},
			expectedError: true,
		},
		{
			name: "multiple validation errors",
			input: TestStruct{
				Name:     "",
				Email:    "invalid",
				Age:      -5,
				Password: "123",
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.ValidateStruct(tc.input)
			if tc.expectedError {
				assert.Error(t, err)
				var validationErrs domain.ValidationErrors
				assert.ErrorAs(t, err, &validationErrs)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateStruct_WithOptionalFields(t *testing.T) {
	v := New()

	t.Run("all empty optional fields is valid", func(t *testing.T) {
		input := OptionalFields{}
		err := v.ValidateStruct(input)
		assert.NoError(t, err)
	})

	t.Run("partial optional fields is valid", func(t *testing.T) {
		input := OptionalFields{
			Name: "John",
		}
		err := v.ValidateStruct(input)
		assert.NoError(t, err)
	})

	t.Run("optional field with invalid data fails", func(t *testing.T) {
		input := OptionalFields{
			Email: "invalid-email",
		}
		err := v.ValidateStruct(input)
		assert.Error(t, err)
	})
}

func TestValidateStruct_NonStructInput(t *testing.T) {
	v := New()

	t.Run("nil pointer", func(t *testing.T) {
		var ptr *TestStruct
		err := v.ValidateStruct(ptr)
		assert.Error(t, err)
	})

	t.Run("non-struct type", func(t *testing.T) {
		err := v.ValidateStruct("string")
		assert.Error(t, err)
	})
}

func TestValidateVar(t *testing.T) {
	v := New()

	testCases := []struct {
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
			field:   "not-an-email",
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
		{
			name:    "valid numeric comparison",
			field:   10,
			tag:     "gte=5,lte=20",
			wantErr: false,
		},
		{
			name:    "invalid numeric comparison",
			field:   25,
			tag:     "gte=5,lte=20",
			wantErr: true,
		},
		{
			name:    "valid uuid",
			field:   "550e8400-e29b-41d4-a716-446655440000",
			tag:     "uuid",
			wantErr: false,
		},
		{
			name:    "invalid uuid",
			field:   "not-a-uuid",
			tag:     "uuid",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.ValidateVar(tc.field, tc.tag)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFormatValidationError(t *testing.T) {
	v := New()

	t.Run("required error", func(t *testing.T) {
		type TestReq struct {
			Name string `validate:"required"`
		}
		err := v.ValidateStruct(TestReq{})
		require.Error(t, err)
		var validationErrs domain.ValidationErrors
		require.ErrorAs(t, err, &validationErrs)
		assert.Contains(t, validationErrs[0].Message, "required")
	})

	t.Run("email error", func(t *testing.T) {
		type TestEmail struct {
			Email string `validate:"email"`
		}
		err := v.ValidateStruct(TestEmail{Email: "invalid"})
		require.Error(t, err)
		var validationErrs domain.ValidationErrors
		require.ErrorAs(t, err, &validationErrs)
		assert.Contains(t, validationErrs[0].Message, "email address")
	})

	t.Run("min error", func(t *testing.T) {
		type TestMin struct {
			Name string `validate:"min=5"`
		}
		err := v.ValidateStruct(TestMin{Name: "ab"})
		require.Error(t, err)
		var validationErrs domain.ValidationErrors
		require.ErrorAs(t, err, &validationErrs)
		assert.Contains(t, validationErrs[0].Message, "at least 5 characters")
	})

	t.Run("max error", func(t *testing.T) {
		type TestMax struct {
			Name string `validate:"max=5"`
		}
		err := v.ValidateStruct(TestMax{Name: "toolong"})
		require.Error(t, err)
		var validationErrs domain.ValidationErrors
		require.ErrorAs(t, err, &validationErrs)
		assert.Contains(t, validationErrs[0].Message, "at most 5 characters")
	})

	t.Run("gte error", func(t *testing.T) {
		type TestGte struct {
			Age int `validate:"gte=18"`
		}
		err := v.ValidateStruct(TestGte{Age: 16})
		require.Error(t, err)
		var validationErrs domain.ValidationErrors
		require.ErrorAs(t, err, &validationErrs)
		assert.Contains(t, validationErrs[0].Message, "greater than or equal to 18")
	})

	t.Run("lte error", func(t *testing.T) {
		type TestLte struct {
			Age int `validate:"lte=100"`
		}
		err := v.ValidateStruct(TestLte{Age: 150})
		require.Error(t, err)
		var validationErrs domain.ValidationErrors
		require.ErrorAs(t, err, &validationErrs)
		assert.Contains(t, validationErrs[0].Message, "less than or equal to 100")
	})
}

func TestHasErrors(t *testing.T) {
	v := New()

	t.Run("returns true for ValidationErrors", func(t *testing.T) {
		type TestStruct struct {
			Name string `validate:"required"`
		}
		err := v.ValidateStruct(TestStruct{})
		assert.True(t, v.HasErrors(err))
	})

	t.Run("returns false for nil error", func(t *testing.T) {
		type TestStruct struct {
			Name string `validate:"required"`
		}
		err := v.ValidateStruct(TestStruct{Name: "John"})
		assert.False(t, v.HasErrors(err))
	})

	t.Run("returns false for other error types", func(t *testing.T) {
		assert.False(t, v.HasErrors(errors.New("some error")))
	})
}

func TestValidationErrors_Error(t *testing.T) {
	t.Run("single error", func(t *testing.T) {
		errs := domain.ValidationErrors{
			{Field: "email", Message: "is required"},
		}
		assert.Contains(t, errs.Error(), "validation failed")
		assert.Contains(t, errs.Error(), "email")
	})

	t.Run("multiple errors", func(t *testing.T) {
		errs := domain.ValidationErrors{
			{Field: "email", Message: "is required"},
			{Field: "name", Message: "is too short"},
		}
		assert.Contains(t, errs.Error(), "validation failed")
		assert.Contains(t, errs.Error(), "2 errors")
	})

	t.Run("empty errors", func(t *testing.T) {
		errs := domain.ValidationErrors{}
		assert.Equal(t, "validation failed", errs.Error())
	})
}

func TestValidationErrors_Error_Nil(t *testing.T) {
	var errs domain.ValidationErrors
	assert.Equal(t, "validation failed", errs.Error())
}
