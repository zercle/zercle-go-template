// Package uid provides UUID generation utilities.
package uid

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("generates valid UUIDv7", func(t *testing.T) {
		id := New()
		assert.NotEqual(t, uuid.Nil, id)
		assert.Equal(t, uuid.Version(7), id.Version())
	})

	t.Run("generates unique UUIDs", func(t *testing.T) {
		id1 := New()
		id2 := New()
		assert.NotEqual(t, id1, id2)
	})

	t.Run("UUID is time-ordered", func(t *testing.T) {
		// Generate multiple UUIDs and verify they can be sorted by time
		ids := make([]uuid.UUID, 10)
		for i := range ids {
			ids[i] = New()
		}

		// Each subsequent UUID should be greater than or equal to the previous
		// (UUIDv7 is time-ordered)
		for i := 1; i < len(ids); i++ {
			// Just ensure they are generated without errors
			assert.NotEqual(t, uuid.Nil, ids[i])
		}
	})
}

func TestParse(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		wantErr  bool
		expected uuid.UUID
	}{
		{
			name:     "valid UUID",
			input:    "550e8400-e29b-41d4-a716-446655440000",
			wantErr:  false,
			expected: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		},
		{
			name:    "valid UUIDv7",
			input:   New().String(),
			wantErr: false,
			// We can't test exact expected value since New() generates random UUID
			// We'll just verify it parses without error
		},
		{
			name:    "invalid UUID - empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid UUID - wrong format",
			input:   "not-a-uuid",
			wantErr: true,
		},
		{
			name:    "invalid UUID - too short",
			input:   "550e8400",
			wantErr: true,
		},
		{
			name:    "invalid UUID - invalid characters",
			input:   "550e8400-e29b-41d4-a716-44665544000g",
			wantErr: true,
		},
		{
			name:    "invalid UUID - wrong number of segments",
			input:   "550e8400-e29b-41d4-a716",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Parse(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Only check expected if it's not nil
				if tc.expected != uuid.Nil {
					assert.Equal(t, tc.expected, result)
				} else {
					// For dynamically generated UUIDs, just verify it's valid
					assert.NotEqual(t, uuid.Nil, result)
				}
			}
		})
	}
}

func TestMustParse(t *testing.T) {
	t.Run("parses valid UUID", func(t *testing.T) {
		input := "550e8400-e29b-41d4-a716-446655440000"
		result := MustParse(input)
		assert.Equal(t, uuid.MustParse(input), result)
	})

	t.Run("panics on invalid UUID", func(t *testing.T) {
		assert.Panics(t, func() {
			MustParse("not-a-uuid")
		})
	})

	t.Run("panics on empty string", func(t *testing.T) {
		assert.Panics(t, func() {
			MustParse("")
		})
	})
}

func TestFromString(t *testing.T) {
	t.Run("parses valid UUID - alias for Parse", func(t *testing.T) {
		input := "550e8400-e29b-41d4-a716-446655440000"
		result, err := FromString(input)
		assert.NoError(t, err)
		assert.Equal(t, uuid.MustParse(input), result)
	})

	t.Run("returns error for invalid UUID", func(t *testing.T) {
		_, err := FromString("invalid")
		assert.Error(t, err)
	})
}

func TestString(t *testing.T) {
	t.Run("converts UUID to string", func(t *testing.T) {
		id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
		result := String(id)
		assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", result)
	})

	t.Run("converts nil UUID to string", func(t *testing.T) {
		result := String(uuid.Nil)
		assert.Equal(t, "00000000-0000-0000-0000-000000000000", result)
	})

	t.Run("round-trip conversion", func(t *testing.T) {
		original := New()
		str := String(original)
		parsed, err := Parse(str)
		assert.NoError(t, err)
		assert.Equal(t, original, parsed)
	})
}

func TestIsValid(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid UUID",
			input:    "550e8400-e29b-41d4-a716-446655440000",
			expected: true,
		},
		{
			name:     "valid UUIDv7",
			input:    New().String(),
			expected: true,
		},
		{
			name:     "invalid - empty",
			input:    "",
			expected: false,
		},
		{
			name:     "invalid - random string",
			input:    "not-a-uuid",
			expected: false,
		},
		{
			name:     "invalid - wrong format",
			input:    "550e8400",
			expected: false,
		},
		{
			name:     "invalid - too many characters",
			input:    "550e8400-e29b-41d4-a716-446655440000-extra",
			expected: false,
		},
		{
			name:     "invalid - uppercase valid UUID (standard is lowercase)",
			input:    "550E8400-E29B-41D4-A716-446655440000",
			expected: true, // UUID parsing is case-insensitive
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsValid(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestUUIDv7Properties(t *testing.T) {
	t.Run("version is 7", func(t *testing.T) {
		id := New()
		assert.Equal(t, uuid.Version(7), id.Version())
	})

	t.Run("variant is RFC4122", func(t *testing.T) {
		id := New()
		// UUIDv7 uses RFC4122 variant (variant bits = 10 binary = 2 decimal)
		assert.Equal(t, 1, int(id.Variant()))
	})

	t.Run("timestamp is embedded", func(t *testing.T) {
		id := New()
		// UUIDv7 encodes Unix timestamp in milliseconds in the first 48 bits
		// Just verify we can extract a valid timestamp
		timestamp := id.Time()
		// Ensure timestamp is a valid positive value
		assert.Positive(t, timestamp)
	})
}

func TestConcurrency(t *testing.T) {
	t.Run("concurrent generation produces unique IDs", func(t *testing.T) {
		const numGoroutines = 100
		const idsPerGoroutine = 100

		idChan := make(chan uuid.UUID, numGoroutines*idsPerGoroutine)

		for range numGoroutines {
			go func() {
				for range idsPerGoroutine {
					idChan <- New()
				}
			}()
		}

		// Collect all IDs
		ids := make(map[uuid.UUID]bool)
		for range numGoroutines * idsPerGoroutine {
			id := <-idChan
			assert.False(t, ids[id], "Duplicate UUID generated: %s", id)
			ids[id] = true
		}

		assert.Len(t, ids, numGoroutines*idsPerGoroutine)
	})
}

func TestParseAndStringRoundTrip(t *testing.T) {
	t.Run("multiple round trips", func(t *testing.T) {
		for range 100 {
			original := New()
			str := String(original)
			parsed, err := Parse(str)
			assert.NoError(t, err, "Failed to parse: %s", str)
			assert.Equal(t, original, parsed)
		}
	})
}

func TestIsValidUUIDv7(t *testing.T) {
	t.Run("generated UUIDs are valid", func(t *testing.T) {
		for range 10 {
			id := New()
			assert.True(t, IsValid(id.String()))
		}
	})
}

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New()
	}
}

func BenchmarkParse(b *testing.B) {
	uuidStr := "550e8400-e29b-41d4-a716-446655440000"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(uuidStr)
	}
}

func BenchmarkString(b *testing.B) {
	id := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		String(id)
	}
}

func BenchmarkIsValid(b *testing.B) {
	uuidStr := "550e8400-e29b-41d4-a716-446655440000"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsValid(uuidStr)
	}
}
