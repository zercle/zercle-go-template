package domain

import (
	std_errors "errors"
	"testing"

	apperrors "github.com/zercle/zercle-go-template/internal/shared/errors"
)

func TestUser_Validate_EmptyUsername(t *testing.T) {
	t.Parallel()
	user := &User{
		Username: "",
		Email:    "test@example.com",
	}
	err := user.Validate()
	if !std_errors.Is(err, apperrors.ErrUsernameRequired) {
		t.Errorf("expected ErrUsernameRequired, got %v", err)
	}
}

func TestUser_Validate_EmptyEmail(t *testing.T) {
	t.Parallel()
	user := &User{
		Username: "testuser",
		Email:    "",
	}
	err := user.Validate()
	if !std_errors.Is(err, apperrors.ErrEmailRequired) {
		t.Errorf("expected ErrEmailRequired, got %v", err)
	}
}

func TestUser_Validate_PasswordTooShort(t *testing.T) {
	t.Parallel()
	user := &User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "short",
	}
	err := user.Validate()
	if !std_errors.Is(err, apperrors.ErrPasswordTooShort) {
		t.Errorf("expected ErrPasswordTooShort, got %v", err)
	}
}

func TestUser_Validate_ValidUser(t *testing.T) {
	t.Parallel()
	user := &User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	if err := user.Validate(); err != nil {
		t.Errorf("expected nil for valid user, got %v", err)
	}
}

func TestUser_Validate_ValidUserWithoutPassword(t *testing.T) {
	t.Parallel()
	user := &User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "",
	}
	if err := user.Validate(); err != nil {
		t.Errorf("expected nil for user without password, got %v", err)
	}
}

func TestNewUser(t *testing.T) {
	t.Parallel()
	user := NewUser("testuser", "test@example.com", "password123", "Test User")

	if user.Username != "testuser" {
		t.Errorf("expected Username=testuser, got %s", user.Username)
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected Email=test@example.com, got %s", user.Email)
	}
	if user.Password != "password123" {
		t.Errorf("expected Password=password123, got %s", user.Password)
	}
	if user.DisplayName != "Test User" {
		t.Errorf("expected DisplayName=Test User, got %s", user.DisplayName)
	}
	if user.Status != string(StatusOffline) {
		t.Errorf("expected Status=offline, got %s", user.Status)
	}
	if user.ID.String() == "" {
		t.Error("expected generated UUID, got empty")
	}
	if user.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if user.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
}
