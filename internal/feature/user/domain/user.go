// Package domain contains the core business entities and domain logic for the user feature.
// This package has no external dependencies and defines the fundamental
// structures of the user domain.
package domain

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

// Argon2id parameters following OWASP recommendations (as of 2023).
// These defaults provide a good balance between security and performance.
// Memory: 64 MB (65536 KB)
// Iterations: 3
// Parallelism: 4 (or runtime.GOMAXPROCS(0))
// Salt length: 16 bytes
// Key length: 32 bytes
// Algorithm: Argon2id
const (
	argon2idVersion   = argon2.Version
	defaultMemory     = 64 * 1024 // 64 MB in KB
	defaultIterations = 3
	defaultParallelism = 4
	defaultSaltLength = 16
	defaultKeyLength  = 32
)

// argon2Params holds the current Argon2id parameters for password hashing.
// Use atomic operations for thread-safe access to the parallelism value.
type argon2Params struct {
	memory     uint32
	iterations uint32
	parallelism uint8
	saltLength uint32
	keyLength  uint32
}

// defaultParams holds the default Argon22id parameters.
var defaultParams = argon2Params{
	memory:     defaultMemory,
	iterations: defaultIterations,
	parallelism: defaultParallelism,
	saltLength: defaultSaltLength,
	keyLength:  defaultKeyLength,
}

// currentParams holds the current Argon2id parameters (can be changed via SetArgon2Params).
var currentParams atomic.Value

func init() {
	currentParams.Store(defaultParams)
}

// SetArgon2Params sets the Argon2id parameters for password hashing.
// This should be called during application initialization.
//
// Recommended settings per OWASP:
//   - Memory: 64 MB (65536 KB) minimum, higher is better if resources allow
//   - Iterations: 3 minimum
//   - Parallelism: 4 or number of CPU cores
//   - Salt length: 16 bytes minimum
//   - Key length: 32 bytes
func SetArgon2Params(memory, iterations int, parallelism uint8, saltLength, keyLength int) {
	if memory < 1024 {
		memory = 1024 // Minimum 1 MB
	}
	if iterations < 1 {
		iterations = 1
	}
	if parallelism < 1 {
		parallelism = 1
	}
	if saltLength < 8 {
		saltLength = 8 // Minimum 8 bytes
	}
	if keyLength < 16 {
		keyLength = 16 // Minimum 16 bytes
	}

	currentParams.Store(argon2Params{
		memory:     uint32(memory),
		iterations: uint32(iterations),
		parallelism: parallelism,
		saltLength: uint32(saltLength),
		keyLength:  uint32(keyLength),
	})
}

// GetArgon2Params returns the current Argon2id parameters.
func GetArgon2Params() (memory, iterations int, parallelism uint8, saltLength, keyLength int) {
	params := currentParams.Load().(argon2Params)
	return int(params.memory), int(params.iterations), params.parallelism, int(params.saltLength), int(params.keyLength)
}

// generateSalt generates a cryptographically secure random salt.
func generateSalt(length int) ([]byte, error) {
	salt := make([]byte, length)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return salt, nil
}

// encodeHash encodes the Argon2id parameters, salt, and hash into a standard format.
// Format: $argon2id$v=<version>$m=<memory>,t=<iterations>,p=<parallelism>$<salt>$<hash>
func encodeHash(params argon2Params, salt, hash []byte) string {
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2idVersion, params.memory, params.iterations, params.parallelism, b64Salt, b64Hash)
}

// decodeHash parses an encoded Argon2id hash string and extracts its components.
func decodeHash(encodedHash string) (params argon2Params, salt, hash []byte, err error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return params, nil, nil, fmt.Errorf("invalid hash format")
	}

	if parts[1] != "argon2id" {
		return params, nil, nil, fmt.Errorf("unsupported algorithm: %s", parts[1])
	}

	var version int
	_, err = fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return params, nil, nil, fmt.Errorf("invalid version: %w", err)
	}
	if version != argon2idVersion {
		return params, nil, nil, fmt.Errorf("unsupported argon2 version: %d", version)
	}

	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &params.memory, &params.iterations, &params.parallelism)
	if err != nil {
		return params, nil, nil, fmt.Errorf("invalid parameters: %w", err)
	}

	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return params, nil, nil, fmt.Errorf("invalid salt: %w", err)
	}
	params.saltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return params, nil, nil, fmt.Errorf("invalid hash: %w", err)
	}
	params.keyLength = uint32(len(hash))

	return params, salt, hash, nil
}

// User represents a user entity in the domain.
// It contains the core user data and business rules.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-"` // Never expose in JSON
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// GetID returns the user ID for interface compatibility.
func (u *User) GetID() string {
	return u.ID
}

// GetEmail returns the user email for interface compatibility.
func (u *User) GetEmail() string {
	return u.Email
}

// emailRegex is the regex pattern for email validation.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// Validate performs validation on the user entity.
// Returns nil if valid, otherwise returns an error describing the validation failure.
func (u *User) Validate() error {
	if u.ID == "" {
		return ErrInvalidID
	}
	if !IsValidEmail(u.Email) {
		return ErrInvalidEmail
	}
	if u.Name == "" {
		return ErrInvalidName
	}
	if len(u.Name) < 2 || len(u.Name) > 100 {
		return ErrInvalidNameLength
	}
	return nil
}

// SetPassword hashes and sets the user's password using Argon2id.
// The password must be at least 8 characters long.
// Uses the configured Argon2id parameters (default: OWASP recommended settings).
func (u *User) SetPassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}

	params := currentParams.Load().(argon2Params)

	// Generate a cryptographically secure random salt
	salt, err := generateSalt(int(params.saltLength))
	if err != nil {
		return ErrPasswordHashFailed
	}

	// Generate the Argon2id hash
	hash := argon2.IDKey([]byte(password), salt, params.iterations, params.memory, params.parallelism, params.keyLength)

	// Encode the hash with parameters for storage
	u.PasswordHash = encodeHash(params, salt, hash)
	u.UpdatedAt = time.Now()
	return nil
}

// VerifyPassword checks if the provided password matches the stored hash.
// Uses constant-time comparison to prevent timing attacks.
func (u *User) VerifyPassword(password string) bool {
	// Decode the stored hash to get parameters and salt
	params, salt, expectedHash, err := decodeHash(u.PasswordHash)
	if err != nil {
		return false
	}

	// Compute the hash with the same parameters and salt
	computedHash := argon2.IDKey([]byte(password), salt, params.iterations, params.memory, params.parallelism, params.keyLength)

	// Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare(computedHash, expectedHash) == 1 {
		return true
	}
	return false
}

// Update updates the user's fields and sets the updated timestamp.
// Only non-empty fields are updated.
func (u *User) Update(name string) {
	if name != "" {
		u.Name = name
	}
	u.UpdatedAt = time.Now()
}

// IsValidEmail validates an email address format.
func IsValidEmail(email string) bool {
	if email == "" {
		return false
	}
	return emailRegex.MatchString(email)
}

// NewUser creates a new user with the given email, name, and password.
// It generates a new UUID for the user and hashes the password.
func NewUser(email, name, password string) (*User, error) {
	if !IsValidEmail(email) {
		return nil, ErrInvalidEmail
	}
	if name == "" {
		return nil, ErrInvalidName
	}
	if len(name) < 2 || len(name) > 100 {
		return nil, ErrInvalidNameLength
	}

	now := time.Now()
	user := &User{
		ID:        uuid.New().String(),
		Email:     email,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := user.SetPassword(password); err != nil {
		return nil, err
	}

	return user, nil
}

// Domain errors for the user entity.
// These are used for validation and business rule violations.
var (
	ErrInvalidID          = NewDomainError("INVALID_ID", "user ID is required")
	ErrInvalidEmail       = NewDomainError("INVALID_EMAIL", "email format is invalid")
	ErrInvalidName        = NewDomainError("INVALID_NAME", "name is required")
	ErrInvalidNameLength  = NewDomainError("INVALID_NAME_LENGTH", "name must be between 2 and 100 characters")
	ErrPasswordTooShort   = NewDomainError("PASSWORD_TOO_SHORT", "password must be at least 8 characters")
	ErrPasswordHashFailed = NewDomainError("PASSWORD_HASH_FAILED", "failed to hash password")
	ErrUserNotFound       = NewDomainError("USER_NOT_FOUND", "user not found")
	ErrDuplicateEmail     = NewDomainError("DUPLICATE_EMAIL", "email already exists")
)

// DomainError represents a domain-specific error.
type DomainError struct {
	Code    string
	Message string
}

// Error implements the error interface.
func (e *DomainError) Error() string {
	return e.Message
}

// NewDomainError creates a new domain error.
func NewDomainError(code, message string) *DomainError {
	return &DomainError{Code: code, Message: message}
}
