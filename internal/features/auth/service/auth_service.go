package service

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/zercle/zercle-go-template/internal/features/auth/domain"
	apperrors "github.com/zercle/zercle-go-template/internal/shared/errors"
	"github.com/zercle/zercle-go-template/pkg/uuidgen"
)

// AuthService provides authentication operations including registration, login, and session management.
type AuthService struct {
	userRepo      domain.UserRepository
	sessionRepo   domain.SessionRepository
	jwtSecret     []byte
	jwtExpiry     time.Duration
	refreshExpiry time.Duration
}

var _ AuthServiceInterface = (*AuthService)(nil)

// NewAuthService creates a new AuthService with the given dependencies.
func NewAuthService(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
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

// RegisterInput holds the data required to register a new user.
type RegisterInput struct {
	Username    string
	Email       string
	Password    string
	DisplayName string
}

// LoginInput holds the data required to authenticate a user.
type LoginInput struct {
	Email    string
	Password string
}

// AuthResult contains the authentication tokens and user information.
type AuthResult struct {
	AccessToken  string
	RefreshToken string
	User         *domain.User
	ExpiresAt    int64
}

// Register creates a new user account and returns authentication tokens.
func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*AuthResult, error) {
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

// Login authenticates a user with email and password and returns tokens.
func (s *AuthService) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	user, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, apperrors.ErrInvalidCredentials
	}
	if user == nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	return s.generateAuthResult(ctx, user)
}

// ValidateToken validates a JWT token and returns the associated user.
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*domain.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperrors.ErrTokenInvalid
		}
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, apperrors.ErrTokenInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, apperrors.ErrTokenInvalid
	}

	userIDStr, ok := claims["sub"].(string)
	if !ok {
		return nil, apperrors.ErrTokenInvalid
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, apperrors.ErrTokenInvalid
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, apperrors.ErrUserNotFound
	}

	return user, nil
}

// RefreshToken issues new authentication tokens using a valid refresh token.
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error) {
	session, err := s.sessionRepo.FindByToken(ctx, refreshToken)
	if err != nil {
		return nil, apperrors.ErrTokenInvalid
	}
	if session == nil {
		return nil, apperrors.ErrTokenInvalid
	}

	if session.ExpiresAt.Before(time.Now()) {
		return nil, apperrors.ErrTokenExpired
	}

	user, err := s.userRepo.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, apperrors.ErrUserNotFound
	}

	return s.generateAuthResult(ctx, user)
}

// Logout invalidates all sessions for the specified user.
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.sessionRepo.DeleteByUserID(ctx, userID)
}

func (s *AuthService) generateAuthResult(ctx context.Context, user *domain.User) (*AuthResult, error) {
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

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
		ExpiresAt:    expiresAt.Unix(),
	}, nil
}
