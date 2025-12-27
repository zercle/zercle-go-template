package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	usermock "github.com/zercle/zercle-go-template/domain/user/mock"
	"github.com/zercle/zercle-go-template/domain/user/model"
	"github.com/zercle/zercle-go-template/domain/user/repository"
	"github.com/zercle/zercle-go-template/domain/user/request"
	"github.com/zercle/zercle-go-template/infrastructure/config"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
)

func setupTestUserUseCase(t *testing.T) (*userUseCase, *usermock.MockRepository) {
	ctrl := gomock.NewController(t)
	mockRepo := usermock.NewMockRepository(ctrl)

	jwtConfig := &config.JWTConfig{
		Secret:     "test-secret-key-for-jwt",
		Expiration: 3600,
	}
	logConfig := &config.LoggingConfig{
		Level:  "debug",
		Format: "console",
	}
	log := logger.NewLogger(logConfig)

	uc := &userUseCase{
		repo: mockRepo,
		jwt:  jwtConfig,
		log:  log,
	}

	return uc, mockRepo
}

func TestUserUseCase_Register(t *testing.T) {
	uc, mockRepo := setupTestUserUseCase(t)

	tests := []struct {
		setup   func()
		request request.RegisterUser
		name    string
		errMsg  string
		wantErr bool
	}{
		{
			name: "successful registration",
			setup: func() {
				mockRepo.EXPECT().GetByEmail(gomock.Any(), "test@example.com").Return(nil, repository.ErrUserNotFound)
				mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&model.User{
					ID:       uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
					Email:    "test@example.com",
					FullName: "Test User",
					Phone:    "1234567890",
				}, nil)
			},
			request: request.RegisterUser{
				Email:    "test@example.com",
				Password: "password123",
				FullName: "Test User",
				Phone:    "1234567890",
			},
			wantErr: false,
		},
		{
			name: "email already exists",
			setup: func() {
				mockRepo.EXPECT().GetByEmail(gomock.Any(), "existing@example.com").Return(&model.User{
					ID:       uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
					Email:    "existing@example.com",
					FullName: "Existing User",
				}, nil)
			},
			request: request.RegisterUser{
				Email:    "existing@example.com",
				Password: "password123",
				FullName: "New User",
			},
			wantErr: true,
			errMsg:  ErrUserAlreadyExists.Error(),
		},
		{
			name: "repository error on get by email",
			setup: func() {
				mockRepo.EXPECT().GetByEmail(gomock.Any(), "test@example.com").Return(nil, errors.New("database connection failed"))
			},
			request: request.RegisterUser{
				Email:    "test@example.com",
				Password: "password123",
				FullName: "Test User",
			},
			wantErr: true,
			errMsg:  "database connection failed",
		},
		{
			name: "repository error on create",
			setup: func() {
				mockRepo.EXPECT().GetByEmail(gomock.Any(), "test@example.com").Return(nil, repository.ErrUserNotFound)
				mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("failed to create user"))
			},
			request: request.RegisterUser{
				Email:    "test@example.com",
				Password: "password123",
				FullName: "Test User",
			},
			wantErr: true,
			errMsg:  "failed to create user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			result, err := uc.Register(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.Token)
				assert.Equal(t, tt.request.Email, result.User.Email)
			}
		})
	}
}

func TestUserUseCase_Login(t *testing.T) {
	uc, mockRepo := setupTestUserUseCase(t)
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		setup   func()
		request request.LoginUser
		name    string
		errMsg  string
		wantErr bool
	}{
		{
			name: "successful login",
			setup: func() {
				mockRepo.EXPECT().GetByEmail(gomock.Any(), "test@example.com").Return(&model.User{
					ID:       uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
					Email:    "test@example.com",
					Password: string(hashedPassword),
					FullName: "Test User",
					Phone:    "1234567890",
				}, nil)
			},
			request: request.LoginUser{
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "user not found",
			setup: func() {
				mockRepo.EXPECT().GetByEmail(gomock.Any(), "nonexistent@example.com").Return(nil, repository.ErrUserNotFound)
			},
			request: request.LoginUser{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			wantErr: true,
			errMsg:  ErrInvalidCredentials.Error(),
		},
		{
			name: "wrong password",
			setup: func() {
				mockRepo.EXPECT().GetByEmail(gomock.Any(), "test@example.com").Return(&model.User{
					ID:       uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
					Email:    "test@example.com",
					Password: string(hashedPassword),
					FullName: "Test User",
				}, nil)
			},
			request: request.LoginUser{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			wantErr: true,
			errMsg:  ErrInvalidCredentials.Error(),
		},
		{
			name: "repository error",
			setup: func() {
				mockRepo.EXPECT().GetByEmail(gomock.Any(), "test@example.com").Return(nil, errors.New("database error"))
			},
			request: request.LoginUser{
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			result, err := uc.Login(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.Token)
			}
		})
	}
}

func TestUserUseCase_GetProfile(t *testing.T) {
	uc, mockRepo := setupTestUserUseCase(t)
	testUserID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		setup   func()
		name    string
		errMsg  string
		userID  uuid.UUID
		wantErr bool
	}{
		{
			name: "successful get profile",
			setup: func() {
				mockRepo.EXPECT().GetByID(gomock.Any(), testUserID).Return(&model.User{
					ID:       testUserID,
					Email:    "test@example.com",
					FullName: "Test User",
					Phone:    "1234567890",
				}, nil)
			},
			userID:  testUserID,
			wantErr: false,
		},
		{
			name: "user not found",
			setup: func() {
				mockRepo.EXPECT().GetByID(gomock.Any(), testUserID).Return(nil, repository.ErrUserNotFound)
			},
			userID:  testUserID,
			wantErr: true,
			errMsg:  repository.ErrUserNotFound.Error(),
		},
		{
			name: "repository error",
			setup: func() {
				mockRepo.EXPECT().GetByID(gomock.Any(), testUserID).Return(nil, errors.New("database error"))
			},
			userID:  testUserID,
			wantErr: true,
			errMsg:  "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			result, err := uc.GetProfile(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.userID, result.ID)
			}
		})
	}
}

func TestUserUseCase_DeleteAccount(t *testing.T) {
	uc, mockRepo := setupTestUserUseCase(t)
	testUserID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		setup   func()
		name    string
		errMsg  string
		userID  uuid.UUID
		wantErr bool
	}{
		{
			name: "successful deletion",
			setup: func() {
				mockRepo.EXPECT().Delete(gomock.Any(), testUserID).Return(nil)
			},
			userID:  testUserID,
			wantErr: false,
		},
		{
			name: "user not found",
			setup: func() {
				mockRepo.EXPECT().Delete(gomock.Any(), testUserID).Return(repository.ErrUserNotFound)
			},
			userID:  testUserID,
			wantErr: true,
			errMsg:  repository.ErrUserNotFound.Error(),
		},
		{
			name: "repository error",
			setup: func() {
				mockRepo.EXPECT().Delete(gomock.Any(), testUserID).Return(errors.New("database error"))
			},
			userID:  testUserID,
			wantErr: true,
			errMsg:  "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			err := uc.DeleteAccount(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
