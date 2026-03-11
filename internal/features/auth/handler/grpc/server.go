package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/zercle/zercle-go-template/api/pb"
	"github.com/zercle/zercle-go-template/internal/features/auth/domain"
	"github.com/zercle/zercle-go-template/internal/features/auth/service"
	apperrors "github.com/zercle/zercle-go-template/internal/shared/errors"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	authService *service.AuthService
}

func NewAuthServer(authService *service.AuthService) *AuthServer {
	return &AuthServer{authService: authService}
}

func (s *AuthServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthResponse, error) {
	input := service.RegisterInput{
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

func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	input := service.LoginInput{
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

func (s *AuthServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	user, err := s.authService.ValidateToken(ctx, req.Token)
	if err != nil {
		return &pb.ValidateTokenResponse{Valid: false}, nil
	}

	return &pb.ValidateTokenResponse{
		Valid:    true,
		UserId:   user.ID.String(),
		Username: user.Username,
	}, nil
}

func (s *AuthServer) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.AuthResponse, error) {
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

func (s *AuthServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*emptypb.Empty, error) {
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
