package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/zercle/zercle-go-template/api/proto/auth/v1"
	apperrors "github.com/zercle/zercle-go-template/internal/core/errors"
	"github.com/zercle/zercle-go-template/internal/feature/auth/domain"
	"github.com/zercle/zercle-go-template/internal/feature/auth/ports"
)

// Server implements the gRPC authentication service.
type Server struct {
	pb.UnimplementedAuthServiceServer
	authService ports.AuthService
}

// NewServer creates a new gRPC authentication server.
func NewServer(authService ports.AuthService) *Server {
	return &Server{authService: authService}
}

// Register handles user registration via gRPC.
func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthResponse, error) {
	input := ports.RegisterInput{
		Username:    req.Username,
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	}

	result, err := s.authService.Register(ctx, input)
	if err != nil {
		return nil, err
	}

	return &pb.AuthResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		User:         toProtoUser(result.User),
		ExpiresAt:    result.ExpiresAt,
	}, nil
}

// Login handles user login via gRPC.
func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	input := ports.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}

	result, err := s.authService.Login(ctx, input)
	if err != nil {
		return nil, err
	}

	return &pb.AuthResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		User:         toProtoUser(result.User),
		ExpiresAt:    result.ExpiresAt,
	}, nil
}

// ValidateToken validates a JWT token via gRPC.
func (s *Server) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	user, err := s.authService.ValidateToken(ctx, req.Token)
	if err != nil {
		return &pb.ValidateTokenResponse{Valid: false}, err
	}

	return &pb.ValidateTokenResponse{
		Valid:    true,
		UserId:   user.ID.String(),
		Username: user.Username,
	}, nil
}

// RefreshToken handles token refresh via gRPC.
func (s *Server) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.AuthResponse, error) {
	result, err := s.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	return &pb.AuthResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		User:         toProtoUser(result.User),
		ExpiresAt:    result.ExpiresAt,
	}, nil
}

// Logout handles user logout via gRPC.
func (s *Server) Logout(ctx context.Context, req *pb.LogoutRequest) (*emptypb.Empty, error) {
	userID, err := parseUserID(req.UserId)
	if err != nil {
		return nil, apperrors.ErrUnauthorized
	}

	if err := s.authService.Logout(ctx, userID); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func toProtoUser(user *domain.User) *pb.User {
	if user == nil {
		return nil
	}
	return &pb.User{
		Id:          user.ID.String(),
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarUrl:   user.AvatarURL,
		Status:      user.Status,
		CreatedAt:   timestamppb.New(user.CreatedAt),
	}
}

func parseUserID(userIDStr string) (uuid.UUID, error) {
	if userIDStr == "" {
		return uuid.Nil, errors.New("user ID is required")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}
