package auth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordHasher_Hash_Success(t *testing.T) {
	hasher := NewPasswordHasher()

	password := "testPassword123"
	hash, err := hasher.Hash(password)

	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.Contains(t, hash, "$argon2id$")
}

func TestPasswordHasher_Hash_DifferentPasswordsProduceDifferentHashes(t *testing.T) {
	hasher := NewPasswordHasher()

	password1 := "password123"
	password2 := "differentPassword456"

	hash1, err1 := hasher.Hash(password1)
	hash2, err2 := hasher.Hash(password2)

	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.NotEqual(t, hash1, hash2, "Same password should produce different hashes due to random salt")
}

func TestPasswordHasher_Hash_SamePasswordProducesDifferentHashes(t *testing.T) {
	hasher := NewPasswordHasher()

	password := "samePassword123"

	hash1, err1 := hasher.Hash(password)
	hash2, err2 := hasher.Hash(password)

	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.NotEqual(t, hash1, hash2, "Same password should produce different hashes due to random salt")
}

func TestPasswordHasher_Verify_CorrectPassword(t *testing.T) {
	hasher := NewPasswordHasher()

	password := "testPassword123"
	hash, err := hasher.Hash(password)
	require.NoError(t, err)

	err = hasher.Verify(password, hash)

	assert.NoError(t, err)
}

func TestPasswordHasher_Verify_WrongPassword(t *testing.T) {
	hasher := NewPasswordHasher()

	password := "testPassword123"
	wrongPassword := "wrongPassword456"
	hash, err := hasher.Hash(password)
	require.NoError(t, err)

	err = hasher.Verify(wrongPassword, hash)

	assert.Error(t, err)
	assert.Equal(t, ErrPasswordMismatch, err)
}

func TestPasswordHasher_Verify_InvalidHash(t *testing.T) {
	hasher := NewPasswordHasher()

	invalidHash := "invalid-hash-format"

	err := hasher.Verify("password", invalidHash)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidHash, err)
}

func TestPasswordHasher_Verify_EmptyPassword(t *testing.T) {
	hasher := NewPasswordHasher()

	password := "testPassword123"
	hash, err := hasher.Hash(password)
	require.NoError(t, err)

	err = hasher.Verify("", hash)

	assert.Error(t, err)
}

func TestPasswordHasher_Verify_EmptyPasswordWithValidHash(t *testing.T) {
	hasher := NewPasswordHasher()

	hash, err := hasher.Hash("testPassword123")
	require.NoError(t, err)

	err = hasher.Verify("", hash)

	assert.Error(t, err)
}

func TestPasswordHasher_Hash_VeryLongPassword(t *testing.T) {
	hasher := NewPasswordHasher()

	// Create a very long password
	longPassword := strings.Repeat("a", 1000)

	hash, err := hasher.Hash(longPassword)

	require.NoError(t, err)
	assert.NotEmpty(t, hash)

	err = hasher.Verify(longPassword, hash)
	assert.NoError(t, err)
}

func TestPasswordHasher_Verify_AlmostCorrectPassword(t *testing.T) {
	hasher := NewPasswordHasher()

	password := "testPassword123"
	almostCorrect := "testPassword124"
	hash, err := hasher.Hash(password)
	require.NoError(t, err)

	err = hasher.Verify(almostCorrect, hash)

	assert.Error(t, err)
	assert.Equal(t, ErrPasswordMismatch, err)
}

func TestPasswordHasher_WithCustomConfig(t *testing.T) {
	config := &Argon2Config{
		Memory:      32 * 1024, // 32MB
		Iterations:  1,
		Parallelism: 2,
		SaltLength:  8,
		KeyLength:   16,
	}
	hasher := NewPasswordHasherWithConfig(config)

	password := "testPassword123"
	hash, err := hasher.Hash(password)

	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.Contains(t, hash, "$argon2id$")

	err = hasher.Verify(password, hash)
	assert.NoError(t, err)
}

func TestConstantTimeCompare(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{
			name:     "equal strings",
			a:        "hello",
			b:        "hello",
			expected: true,
		},
		{
			name:     "different strings",
			a:        "hello",
			b:        "world",
			expected: false,
		},
		{
			name:     "empty strings",
			a:        "",
			b:        "",
			expected: true,
		},
		{
			name:     "one empty string",
			a:        "hello",
			b:        "",
			expected: false,
		},
		{
			name:     "different lengths",
			a:        "short",
			b:        "much longer string",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConstantTimeCompare(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHashToken(t *testing.T) {
	token := "test-token-123"
	hash := HashToken(token)

	assert.NotEmpty(t, hash)
	assert.NotEqual(t, token, hash, "Hash should be different from original token")
}

func TestCompareToken(t *testing.T) {
	token := "test-token-123"
	hash := HashToken(token)

	result := CompareToken(token, hash)
	assert.True(t, result)

	wrongToken := "wrong-token"
	result = CompareToken(wrongToken, hash)
	assert.False(t, result)
}

func TestDecodeHash_ValidHash(t *testing.T) {
	hasher := NewPasswordHasher()
	password := "testPassword123"
	hash, err := hasher.Hash(password)
	require.NoError(t, err)

	config, salt, decodedHash, err := decodeHash(hash)

	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.NotEmpty(t, salt)
	assert.NotEmpty(t, decodedHash)
}

func TestDecodeHash_InvalidHash_WrongNumberOfParts(t *testing.T) {
	_, _, _, err := decodeHash("invalid")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidHash, err)
}

func TestDecodeHash_InvalidHash_WrongVersion(t *testing.T) {
	_, _, _, err := decodeHash("$argon2id$v=0$m=65536,t=3,p=4$abcdefghijklmnop$qrstuvwxyz")

	assert.Error(t, err)
	assert.Equal(t, ErrIncompatibleVersion, err)
}

func TestDecodeHash_InvalidHash_BadFormat(t *testing.T) {
	_, _, _, err := decodeHash("$argon2id$v=19$invalid$m=65536,t=3,p=4$abcdefghijklmnop$qrstuvwxyz")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidHash, err)
}
