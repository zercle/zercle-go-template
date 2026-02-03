package service

import (
	"context"
	"testing"
	"time"

	"github.com/zercle/zercle-go-template/internal/features/auth/domain"
	apperrors "github.com/zercle/zercle-go-template/internal/shared/errors"
)

func TestAuthService_Register(t *testing.T) {
	userRepo := NewUserRepoMock()
	sessionRepo := NewSessionRepoMock()
	authSvc := NewAuthService(userRepo, sessionRepo, "test-secret", time.Hour, time.Hour*24*7)

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		input := RegisterInput{
			Username:    "testuser",
			Email:       "test@example.com",
			Password:    "password123",
			DisplayName: "Test User",
		}

		result, err := authSvc.Register(ctx, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.AccessToken == "" {
			t.Error("expected access token")
		}
		if result.RefreshToken == "" {
			t.Error("expected refresh token")
		}
		if result.User == nil {
			t.Error("expected user")
		}
		if result.User.Username != input.Username {
			t.Errorf("expected username %s, got %s", input.Username, result.User.Username)
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		input := RegisterInput{
			Username:    "testuser2",
			Email:       "test@example.com",
			Password:    "password123",
			DisplayName: "Test User 2",
		}

		_, err := authSvc.Register(ctx, input)
		if err != apperrors.ErrAlreadyExists {
			t.Errorf("expected ErrAlreadyExists, got %v", err)
		}
	})
}

func TestAuthService_Login(t *testing.T) {
	userRepo := NewUserRepoMock()
	sessionRepo := NewSessionRepoMock()
	authSvc := NewAuthService(userRepo, sessionRepo, "test-secret", time.Hour, time.Hour*24*7)

	ctx := context.Background()

	user := domain.NewUser("testuser", "test@example.com", "hashedpassword", "Test User")
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	t.Run("success", func(t *testing.T) {
		input := LoginInput{
			Email:    "test@example.com",
			Password: "password123",
		}

		_, err := authSvc.Login(ctx, input)
		if err != apperrors.ErrInvalidCredentials {
			t.Fatalf("expected ErrInvalidCredentials for wrong password, got %v", err)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		input := LoginInput{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}

		_, err := authSvc.Login(ctx, input)
		if err != apperrors.ErrInvalidCredentials {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})
}

func TestAuthService_ValidateToken(t *testing.T) {
	userRepo := NewUserRepoMock()
	sessionRepo := NewSessionRepoMock()
	authSvc := NewAuthService(userRepo, sessionRepo, "test-secret", time.Hour, time.Hour*24*7)

	ctx := context.Background()

	user := domain.NewUser("testuser", "test@example.com", "hashedpassword", "Test User")
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	registerResult, err := authSvc.Register(ctx, RegisterInput{
		Username: "testuser",
		Email:    "test2@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	t.Run("valid token", func(t *testing.T) {
		resultUser, err := authSvc.ValidateToken(ctx, registerResult.AccessToken)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resultUser.ID != registerResult.User.ID {
			t.Errorf("expected user ID %s, got %s", registerResult.User.ID, resultUser.ID)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		_, err := authSvc.ValidateToken(ctx, "invalid-token")
		if err != apperrors.ErrTokenInvalid {
			t.Errorf("expected ErrTokenInvalid, got %v", err)
		}
	})
}

func TestAuthService_RefreshToken(t *testing.T) {
	userRepo := NewUserRepoMock()
	sessionRepo := NewSessionRepoMock()
	authSvc := NewAuthService(userRepo, sessionRepo, "test-secret", time.Hour, time.Hour*24*7)

	ctx := context.Background()

	registerResult, err := authSvc.Register(ctx, RegisterInput{
		Username: "testuser",
		Email:    "test3@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	t.Run("valid refresh token", func(t *testing.T) {
		result, err := authSvc.RefreshToken(ctx, registerResult.RefreshToken)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.AccessToken == "" {
			t.Error("expected new access token")
		}
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		_, err := authSvc.RefreshToken(ctx, "invalid-token")
		if err != apperrors.ErrTokenInvalid {
			t.Errorf("expected ErrTokenInvalid, got %v", err)
		}
	})
}

func TestAuthService_Logout(t *testing.T) {
	userRepo := NewUserRepoMock()
	sessionRepo := NewSessionRepoMock()
	authSvc := NewAuthService(userRepo, sessionRepo, "test-secret", time.Hour, time.Hour*24*7)

	ctx := context.Background()

	registerResult, err := authSvc.Register(ctx, RegisterInput{
		Username: "testuser",
		Email:    "test4@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	user, _ := authSvc.ValidateToken(ctx, registerResult.AccessToken)

	err = authSvc.Logout(ctx, user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = authSvc.RefreshToken(ctx, registerResult.RefreshToken)
	if err != apperrors.ErrTokenInvalid {
		t.Errorf("expected ErrTokenInvalid after logout, got %v", err)
	}
}
