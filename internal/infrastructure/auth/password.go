package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	// ErrInvalidHash indicates the password hash format is invalid.
	ErrInvalidHash = errors.New("invalid password hash format")
	// ErrIncompatibleVersion indicates the argon2 version is incompatible.
	ErrIncompatibleVersion = errors.New("incompatible argon2 version")
	// ErrPasswordMismatch indicates the password does not match.
	ErrPasswordMismatch = errors.New("password does not match")
)

// PasswordHasher provides password hashing and verification.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) error
	Compare(hash, password string) error
}

// Argon2Config holds Argon2id parameters.
type Argon2Config struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultArgon2Config returns recommended production parameters.
// Memory: 64MB, Iterations: 3, Parallelism: 4
func DefaultArgon2Config() *Argon2Config {
	return &Argon2Config{
		Memory:      64 * 1024, // 64MB
		Iterations:  3,
		Parallelism: 4,
		SaltLength:  16,
		KeyLength:   32,
	}
}

// argon2Hasher implements PasswordHasher using Argon2id.
type argon2Hasher struct {
	config *Argon2Config
}

// NewPasswordHasher creates a new Argon2id password hasher.
func NewPasswordHasher() PasswordHasher {
	return &argon2Hasher{
		config: DefaultArgon2Config(),
	}
}

// NewPasswordHasherWithConfig creates a new Argon2id password hasher with custom config.
func NewPasswordHasherWithConfig(config *Argon2Config) PasswordHasher {
	return &argon2Hasher{
		config: config,
	}
}

// Hash creates a password hash using Argon2id.
// Format: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
func (h *argon2Hasher) Hash(password string) (string, error) {
	salt := make([]byte, h.config.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		h.config.Iterations,
		h.config.Memory,
		h.config.Parallelism,
		h.config.KeyLength,
	)

	// Encode to base64
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Format: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		h.config.Memory,
		h.config.Iterations,
		h.config.Parallelism,
		b64Salt,
		b64Hash,
	)

	return encodedHash, nil
}

// Verify checks if a password matches the provided hash.
func (h *argon2Hasher) Verify(password, encodedHash string) error {
	// Parse the encoded hash
	config, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return err
	}

	// Compute the hash of the provided password
	//nolint:gosec // len(hash) is always within uint32 bounds for Argon2
	otherHash := argon2.IDKey([]byte(password), salt, config.Iterations, config.Memory, config.Parallelism, uint32(len(hash)))

	// Constant-time comparison
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return nil
	}

	return ErrPasswordMismatch
}

// Compare checks if a password matches the provided hash.
// This is an alias for Verify with reversed argument order.
func (h *argon2Hasher) Compare(hash, password string) error {
	return h.Verify(password, hash)
}

// decodeHash parses an encoded Argon2id hash.
func decodeHash(encodedHash string) (*Argon2Config, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, ErrInvalidHash
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, ErrInvalidHash
	}

	if version != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	config := &Argon2Config{}
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &config.Memory, &config.Iterations, &config.Parallelism)
	if err != nil {
		return nil, nil, nil, ErrInvalidHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, ErrInvalidHash
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, ErrInvalidHash
	}

	return config, salt, hash, nil
}

// ConstantTimeCompare performs constant-time comparison of two strings.
// This prevents timing attacks by ensuring the comparison takes the same
// amount of time regardless of where the strings differ.
func ConstantTimeCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
