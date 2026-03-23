package test

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	user "github.com/zercle/zercle-go-template/internal/feature/user"
)

// Helper functions for tests

// GenerateRandomID generates a random ID string for testing.
func GenerateRandomID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// GenerateUserID generates a domain UserID for testing.
func GenerateUserID() user.UserID {
	return user.UserID(uuid.New().String())
}

// CreateTestUser creates a valid test user with random email.
func CreateTestUser(tz ...*time.Time) (*user.User, error) {
	now := time.Now()
	if len(tz) > 0 && tz[0] != nil {
		now = *tz[0]
	}

	email := "test." + GenerateRandomID() + "@example.com"
	return user.NewUserWithID(
		user.UserID(uuid.New().String()),
		email,
		"hashedpassword123",
		"John",
		"Doe",
		user.UserStatusActive,
		now,
		now,
	)
}

// CreateTestUsers creates multiple test users.
func CreateTestUsers(count int) ([]*user.User, error) {
	users := make([]*user.User, count)
	for i := range count {
		user, err := CreateTestUser()
		if err != nil {
			return nil, err
		}
		users[i] = user
	}
	return users, nil
}
