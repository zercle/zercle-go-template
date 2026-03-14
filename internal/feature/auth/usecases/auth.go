package usecases

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	apperrors "github.com/zercle/zercle-go-template/internal/core/errors"
	"github.com/zercle/zercle-go-template/internal/feature/auth/domain"
	"github.com/zercle/zercle-go-template/internal/feature/auth/ports"
	"github.com/zercle/zercle-go-template/pkg/uuidgen"
)

// AuthService implements the authentication business logic.
type AuthService struct {
	userRepo      ports.UserRepository
	sessionRepo   ports.SessionRepository
	jwtSecret     []byte
	jwtExpiry     time.Duration
	refreshExpiry time.Duration
}

var _ ports.AuthService = (*AuthService)(nil)

// NewAuthService creates a new authentication service.
func NewAuthService(
	userRepo ports.UserRepository,
	sessionRepo ports.SessionRepository,
	secret string,
	jwtExpiry, refreshExpiry time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:      userRepo,
		sessionRepo:   sessionRepo,
		jwtSecret:     []byte(secret),
		jwtExpiry:     jwtExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// Register registers a new user.
func (s *AuthService) Register(ctx context.Context, input ports.RegisterInput) (*ports.AuthResult, error) {
	existingUser, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err == nil && existingUser != nil {
		return nil, apperrors.ErrAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperrors.ErrInternalError
	}

	user := domain.NewUser(input.Username, input.Email, string(hashedPassword), input.DisplayName)
	if user.DisplayName == "" {
		user.DisplayName = input.Username
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return s.generateAuthResult(ctx, user)
}

// Login authenticates a user and returns tokens.
func (s *AuthService) Login(ctx context.Context, input ports.LoginInput) (*ports.AuthResult, error) {
	user, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	if user == nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return s.generateAuthResult(ctx, user)
}

// ValidateToken validates a JWT token and returns the user.
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*domain.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, domain.ErrTokenInvalid
		}
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, domain.ErrTokenInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, domain.ErrTokenInvalid
	}

	userIDStr, ok := claims["sub"].(string)
	if !ok {
		return nil, domain.ErrTokenInvalid
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, domain.ErrTokenInvalid
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	return user, nil
}

// RefreshToken refreshes access and refresh tokens.
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*ports.AuthResult, error) {
	session, err := s.sessionRepo.FindByToken(ctx, refreshToken)
	if err != nil {
		return nil, domain.ErrTokenInvalid
	}
	if session == nil {
		return nil, domain.ErrTokenInvalid
	}

	if session.ExpiresAt.Before(time.Now()) {
		return nil, domain.ErrTokenExpired
	}

	user, err := s.userRepo.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	return s.generateAuthResult(ctx, user)
}

// Logout logs out a user by deleting all their sessions.
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.sessionRepo.DeleteByUserID(ctx, userID)
}

func (s *AuthService) generateAuthResult(ctx context.Context, user *domain.User) (*ports.AuthResult, error) {
	expiresAt := time.Now().Add(s.jwtExpiry)

	claims := jwt.MapClaims{
		"sub":  user.ID.String(),
		"name": user.Username,
		"exp":  expiresAt.Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, apperrors.ErrInternalError
	}

	refreshToken := uuidgen.NewString()
	refreshExpiresAt := time.Now().Add(s.refreshExpiry)

	session := &domain.Session{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: refreshExpiresAt,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return &ports.AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
		ExpiresAt:    expiresAt.Unix(),
	}, nil
}
