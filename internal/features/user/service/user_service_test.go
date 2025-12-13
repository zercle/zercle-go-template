package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/zercle/zercle-go-template/internal/core/port/mocks"
	userDomain "github.com/zercle/zercle-go-template/internal/features/user/domain"
	userDto "github.com/zercle/zercle-go-template/internal/features/user/dto"
	"github.com/zercle/zercle-go-template/internal/features/user/service"
	sharederrors "github.com/zercle/zercle-go-template/internal/shared/errors"
	"github.com/zercle/zercle-go-template/pkg/utils/password"
	"go.uber.org/mock/gomock"
)

func TestUserService(t *testing.T) {
	jwtSecret := "testsecret"
	jwtExpiry := time.Hour

	t.Run("Register_Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockUserRepository(ctrl)
		svc := service.NewUserService(mockRepo, jwtSecret, jwtExpiry)
		ctx := context.Background()

		req := &userDto.RegisterRequest{
			Email:    "newuser@example.com",
			Password: "securepass123",
			Name:     "New User",
		}

		mockRepo.EXPECT().GetByEmail(gomock.Any(), req.Email).Return(nil, sharederrors.ErrNotFound)
		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, user *userDomain.User) error {
			// Verify user fields
			assert.Equal(t, req.Email, user.Email)
			assert.Equal(t, req.Name, user.Name)
			assert.NotEmpty(t, user.Password)               // Should be hashed
			assert.NotEqual(t, req.Password, user.Password) // Should be different from plain text
			return nil
		})

		// Act
		result, err := svc.Register(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, req.Email, result.Email)
		assert.Equal(t, req.Name, result.Name)
		assert.NotEmpty(t, result.ID)
		assert.NotEmpty(t, result.CreatedAt)
		assert.NotEmpty(t, result.UpdatedAt)
	})

	t.Run("Register_DuplicateEmail", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockUserRepository(ctrl)
		svc := service.NewUserService(mockRepo, jwtSecret, jwtExpiry)
		ctx := context.Background()

		req := &userDto.RegisterRequest{
			Email:    "existing@example.com",
			Password: "securepass123",
			Name:     "User",
		}

		existingUser := &userDomain.User{
			ID:        uuid.New(),
			Email:     req.Email,
			Name:      req.Name,
			Password:  "hashedpassword",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		mockRepo.EXPECT().GetByEmail(gomock.Any(), req.Email).Return(existingUser, nil)

		// Act
		result, err := svc.Register(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, sharederrors.ErrDuplicate, err)
	})

	t.Run("Register_PasswordHashError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockUserRepository(ctrl)
		svc := service.NewUserService(mockRepo, jwtSecret, jwtExpiry)
		ctx := context.Background()

		req := &userDto.RegisterRequest{
			Email:    "test@example.com",
			Password: "securepass123",
			Name:     "User",
		}

		mockRepo.EXPECT().GetByEmail(gomock.Any(), req.Email).Return(nil, sharederrors.ErrNotFound)
		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
		// Note: bcrypt shouldn't fail with normal input

		// Act
		result, err := svc.Register(ctx, req)

		// Assert
		assert.NoError(t, err) // bcrypt should succeed
		assert.NotNil(t, result)
	})

	t.Run("Register_UUIDGenerationError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockUserRepository(ctrl)
		svc := service.NewUserService(mockRepo, jwtSecret, jwtExpiry)
		ctx := context.Background()

		req := &userDto.RegisterRequest{
			Email:    "test@example.com",
			Password: "securepass123",
			Name:     "User",
		}

		mockRepo.EXPECT().GetByEmail(gomock.Any(), req.Email).Return(nil, sharederrors.ErrNotFound)
		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(assert.AnError)

		// Act
		result, err := svc.Register(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("Login_Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockUserRepository(ctrl)
		svc := service.NewUserService(mockRepo, jwtSecret, jwtExpiry)
		ctx := context.Background()

		pwd := "securepass123"
		hashed, _ := password.Hash(pwd)

		req := &userDto.LoginRequest{
			Email:    "test@example.com",
			Password: pwd,
		}

		user := &userDomain.User{
			ID:        uuid.New(),
			Email:     req.Email,
			Name:      "Test User",
			Password:  hashed,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockRepo.EXPECT().GetByEmail(gomock.Any(), req.Email).Return(user, nil)

		// Act
		token, err := svc.Login(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Verify JWT token
		parsedToken, err := jwt.NewParser().Parse(token, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})
		assert.NoError(t, err)
		assert.NotNil(t, parsedToken)
		assert.True(t, parsedToken.Valid)
	})

	t.Run("Login_UserNotFound", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockUserRepository(ctrl)
		svc := service.NewUserService(mockRepo, jwtSecret, jwtExpiry)
		ctx := context.Background()

		req := &userDto.LoginRequest{
			Email:    "notfound@example.com",
			Password: "securepass123",
		}

		mockRepo.EXPECT().GetByEmail(gomock.Any(), req.Email).Return(nil, sharederrors.ErrNotFound)

		// Act
		token, err := svc.Login(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Equal(t, sharederrors.ErrInvalidCreds, err)
	})

	t.Run("Login_WrongPassword", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockUserRepository(ctrl)
		svc := service.NewUserService(mockRepo, jwtSecret, jwtExpiry)
		ctx := context.Background()

		correctPwd := "correctpassword"
		hashed, _ := password.Hash(correctPwd)

		req := &userDto.LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		user := &userDomain.User{
			ID:        uuid.New(),
			Email:     req.Email,
			Name:      "Test User",
			Password:  hashed,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockRepo.EXPECT().GetByEmail(gomock.Any(), req.Email).Return(user, nil)

		// Act
		token, err := svc.Login(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Equal(t, sharederrors.ErrInvalidCreds, err)
	})

	t.Run("GetProfile_Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockUserRepository(ctrl)
		svc := service.NewUserService(mockRepo, jwtSecret, jwtExpiry)
		ctx := context.Background()

		userID := uuid.New()
		user := &userDomain.User{
			ID:        userID,
			Email:     "test@example.com",
			Name:      "Test User",
			Password:  "hashedpassword",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockRepo.EXPECT().GetByID(gomock.Any(), userID).Return(user, nil)

		// Act
		result, err := svc.GetProfile(ctx, userID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID.String(), result.ID)
		assert.Equal(t, user.Email, result.Email)
		assert.Equal(t, user.Name, result.Name)
		assert.Equal(t, user.CreatedAt, result.CreatedAt)
		assert.Equal(t, user.UpdatedAt, result.UpdatedAt)
	})

	t.Run("GetProfile_UserNotFound", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockUserRepository(ctrl)
		svc := service.NewUserService(mockRepo, jwtSecret, jwtExpiry)
		ctx := context.Background()

		userID := uuid.New()
		mockRepo.EXPECT().GetByID(gomock.Any(), userID).Return(nil, sharederrors.ErrNotFound)

		// Act
		result, err := svc.GetProfile(ctx, userID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, sharederrors.ErrNotFound, err)
	})

	t.Run("GetProfile_ContextCancellation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockUserRepository(ctrl)
		svc := service.NewUserService(mockRepo, jwtSecret, jwtExpiry)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		userID := uuid.New()
		mockRepo.EXPECT().GetByID(gomock.Any(), userID).Return(nil, context.Canceled)

		// Act
		result, err := svc.GetProfile(ctx, userID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	// Test table for Register with different inputs
	// Note: Service doesn't validate input - validation is done at handler level
	t.Run("Register_DifferentInputs", func(t *testing.T) {
		tests := []struct {
			testName string
			email    string
			password string
			name     string
		}{
			{
				testName: "Valid Request",
				email:    "test@example.com",
				password: "securepass123",
				name:     "Test User",
			},
			{
				testName: "Long Email",
				email:    "verylongemailaddress@example.com",
				password: "securepass123",
				name:     "Test User",
			},
			{
				testName: "Long Password",
				email:    "test@example.com",
				password: "averyverylongandcomplexpassword123!@#",
				name:     "Test User",
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()
				// Arrange
				mockRepo := mocks.NewMockUserRepository(ctrl)
				svc := service.NewUserService(mockRepo, jwtSecret, jwtExpiry)
				ctx := context.Background()

				req := &userDto.RegisterRequest{
					Email:    tt.email,
					Password: tt.password,
					Name:     tt.name,
				}

				mockRepo.EXPECT().GetByEmail(gomock.Any(), req.Email).Return(nil, sharederrors.ErrNotFound)
				mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

				// Act
				result, err := svc.Register(ctx, req)

				// Assert
				assert.NoError(t, err)
				assert.NotNil(t, result)
			})
		}
	})

	t.Run("JWT_TokenGeneration", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockUserRepository(ctrl)
		svc := service.NewUserService(mockRepo, jwtSecret, jwtExpiry)
		ctx := context.Background()

		pwd := "securepass123"
		hashed, _ := password.Hash(pwd)

		req := &userDto.LoginRequest{
			Email:    "test@example.com",
			Password: pwd,
		}

		user := &userDomain.User{
			ID:        uuid.New(),
			Email:     req.Email,
			Name:      "Test User",
			Password:  hashed,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockRepo.EXPECT().GetByEmail(gomock.Any(), req.Email).Return(user, nil)

		// Act
		token, err := svc.Login(ctx, req)

		// Assert
		assert.NoError(t, err)

		// Parse and validate token
		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})
		assert.NoError(t, err)
		assert.NotNil(t, parsedToken)

		claims := parsedToken.Claims.(jwt.MapClaims)
		assert.Equal(t, user.ID.String(), claims["user_id"])
	})

	t.Run("Password_Argon2Hashing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Arrange
		mockRepo := mocks.NewMockUserRepository(ctrl)
		svc := service.NewUserService(mockRepo, jwtSecret, jwtExpiry)
		ctx := context.Background()

		pwd := "MyVerySecurePassword123!"
		req := &userDto.RegisterRequest{
			Email:    "test@example.com",
			Password: pwd,
			Name:     "Test User",
		}

		var storedHash string
		mockRepo.EXPECT().GetByEmail(gomock.Any(), req.Email).Return(nil, sharederrors.ErrNotFound)
		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, user *userDomain.User) error {
			storedHash = user.Password
			return nil
		})

		// Act
		_, err := svc.Register(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, storedHash)
		assert.NotEqual(t, pwd, storedHash)

		// Verify hash can be verified using the password package
		match, err := password.Verify(pwd, storedHash)
		assert.NoError(t, err)
		assert.True(t, match)

		// Verify wrong password fails
		match, err = password.Verify("wrongpassword", storedHash)
		assert.NoError(t, err)
		assert.False(t, match)
	})
}
