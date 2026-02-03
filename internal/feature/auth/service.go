package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/zercle/zercle-go-template/internal/feature/user"
	infra_auth "github.com/zercle/zercle-go-template/internal/infrastructure/auth"
	"github.com/zercle/zercle-go-template/pkg/uid"
)

// Service provides authentication business logic.
type Service struct {
	userService *user.Service
	credRepo    CredentialRepository
	tokenRepo   RefreshTokenRepository
	tokenSvc    infra_auth.TokenService
	pwdHasher   infra_auth.PasswordHasher
	logger      zerolog.Logger
}

// NewService creates a new auth service.
func NewService(
	userService *user.Service,
	credRepo CredentialRepository,
	tokenRepo RefreshTokenRepository,
	tokenSvc infra_auth.TokenService,
	pwdHasher infra_auth.PasswordHasher,
) *Service {
	return &Service{
		userService: userService,
		credRepo:    credRepo,
		tokenRepo:   tokenRepo,
		tokenSvc:    tokenSvc,
		pwdHasher:   pwdHasher,
		logger:      zerolog.Nop(),
	}
}

// NewServiceWithLogger creates a new auth service with a custom logger.
func NewServiceWithLogger(
	userService *user.Service,
	credRepo CredentialRepository,
	tokenRepo RefreshTokenRepository,
	tokenSvc infra_auth.TokenService,
	pwdHasher infra_auth.PasswordHasher,
	logger zerolog.Logger,
) *Service {
	return &Service{
		userService: userService,
		credRepo:    credRepo,
		tokenRepo:   tokenRepo,
		tokenSvc:    tokenSvc,
		pwdHasher:   pwdHasher,
		logger:      logger,
	}
}

// Register registers a new user and returns a token pair.
func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*TokenResponse, error) {
	existingUser, err := s.userService.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	passwordHash, err := s.pwdHasher.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	createUserReq := &user.CreateUserInput{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	userDTO, err := s.userService.Create(ctx, createUserReq)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	userID, _ := uuid.Parse(userDTO.ID)

	credential := &Credential{
		ID:           uid.New(),
		UserID:       userID,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	_, err = s.credRepo.Create(ctx, credential)
	if err != nil {
		return nil, err
	}

	tokenPair, err := s.tokenSvc.GenerateTokenPair(ctx, userDTO.ID, userDTO.Email)
	if err != nil {
		return nil, err
	}

	refreshToken := &RefreshToken{
		ID:        uid.New(),
		UserID:    userID,
		TokenHash: infra_auth.HashToken(tokenPair.RefreshToken),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: now,
	}

	_, err = s.tokenRepo.Create(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(tokenPair.ExpiresAt.Sub(now).Seconds()),
		ExpiresAt:    tokenPair.ExpiresAt,
	}, nil
}

// Login authenticates a user and returns a token pair.
func (s *Service) Login(ctx context.Context, req *LoginRequest) (*TokenResponse, error) {
	user, err := s.userService.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	credential, err := s.credRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := s.pwdHasher.Compare(credential.PasswordHash, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	tokenPair, err := s.tokenSvc.GenerateTokenPair(ctx, user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	userID, _ := uuid.Parse(user.ID)
	refreshToken := &RefreshToken{
		ID:        uid.New(),
		UserID:    userID,
		TokenHash: infra_auth.HashToken(tokenPair.RefreshToken),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: now,
	}

	_, err = s.tokenRepo.Create(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(tokenPair.ExpiresAt.Sub(now).Seconds()),
		ExpiresAt:    tokenPair.ExpiresAt,
	}, nil
}

// Refresh generates a new token pair from a valid refresh token.
func (s *Service) Refresh(ctx context.Context, req *RefreshRequest) (*TokenResponse, error) {
	claims, err := s.tokenSvc.ValidateRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, ErrTokenExpired
	}

	tokenHash := infra_auth.HashToken(req.RefreshToken)
	existingToken, err := s.tokenRepo.GetByHash(ctx, tokenHash)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if !existingToken.IsValid() {
		return nil, ErrTokenExpired
	}

	if existingToken.UserID.String() != claims.UserID {
		return nil, ErrInvalidToken
	}

	if err := s.tokenRepo.Revoke(ctx, tokenHash); err != nil {
		return nil, err
	}

	tokenPair, err := s.tokenSvc.GenerateTokenPair(ctx, claims.UserID, claims.Email)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	newRefreshToken := &RefreshToken{
		ID:        uid.New(),
		UserID:    existingToken.UserID,
		TokenHash: infra_auth.HashToken(tokenPair.RefreshToken),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: now,
	}

	_, err = s.tokenRepo.Create(ctx, newRefreshToken)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(tokenPair.ExpiresAt.Sub(now).Seconds()),
		ExpiresAt:    tokenPair.ExpiresAt,
	}, nil
}

// Logout revokes the given refresh token.
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := infra_auth.HashToken(refreshToken)

	token, err := s.tokenRepo.GetByHash(ctx, tokenHash)
	if err != nil {
		return ErrInvalidToken
	}

	if token.IsRevoked() {
		return ErrInvalidToken
	}

	if err := s.tokenRepo.Revoke(ctx, tokenHash); err != nil {
		return err
	}

	return nil
}

// Me returns the profile for the given user ID.
func (s *Service) Me(ctx context.Context, userID string) (*UserResponse, error) {
	user, err := s.userService.GetByID(ctx, userID)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return &UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}
