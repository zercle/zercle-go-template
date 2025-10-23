package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/zercle/zercle-go-template/internal/core/domain"
	domerrors "github.com/zercle/zercle-go-template/internal/core/domain/errors"
	"github.com/zercle/zercle-go-template/internal/core/service"
	"github.com/zercle/zercle-go-template/pkg/dto"
	"github.com/zercle/zercle-go-template/pkg/utils/password"
	"github.com/zercle/zercle-go-template/test/mocks"
	"go.uber.org/mock/gomock"
)

func TestUserService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	svc := service.NewUserService(mockRepo, "secret", time.Hour)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		req := &dto.RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
			Name:     "Test User",
		}

		mockRepo.EXPECT().GetByEmail(ctx, req.Email).Return(nil, domerrors.ErrNotFound)
		mockRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)

		res, err := svc.Register(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, req.Email, res.Email)
	})

	t.Run("DuplicateEmail", func(t *testing.T) {
		req := &dto.RegisterRequest{
			Email:    "existing@example.com",
			Password: "password123",
			Name:     "Existing User",
		}

		// Simulate user found (no error)
		mockRepo.EXPECT().GetByEmail(ctx, req.Email).Return(nil, nil)

		res, err := svc.Register(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, domerrors.ErrDuplicate, err)
		assert.Nil(t, res)
	})
}

func TestUserService_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	svc := service.NewUserService(mockRepo, "secret", time.Hour)
	ctx := context.Background()

	plainPassword := "password123"
	hashedPassword, _ := password.Hash(plainPassword)

	user := &domain.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Password: hashedPassword,
		Name:     "Test User",
	}

	t.Run("Success", func(t *testing.T) {
		req := &dto.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		mockRepo.EXPECT().GetByEmail(ctx, req.Email).Return(user, nil)

		token, err := svc.Login(ctx, req)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("InvalidEmail", func(t *testing.T) {
		req := &dto.LoginRequest{
			Email:    "wrong@example.com",
			Password: "password123",
		}

		mockRepo.EXPECT().GetByEmail(ctx, req.Email).Return(nil, domerrors.ErrNotFound)

		token, err := svc.Login(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, domerrors.ErrInvalidCreds, err)
		assert.Empty(t, token)
	})

	t.Run("InvalidPassword", func(t *testing.T) {
		req := &dto.LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		mockRepo.EXPECT().GetByEmail(ctx, req.Email).Return(user, nil)

		token, err := svc.Login(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, domerrors.ErrInvalidCreds, err)
		assert.Empty(t, token)
	})
}

func TestUserService_GetProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	svc := service.NewUserService(mockRepo, "secret", time.Hour)
	ctx := context.Background()

	user := &domain.User{
		ID:    uuid.New(),
		Name:  "Test User",
		Email: "test@example.com",
	}

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().GetByID(ctx, user.ID).Return(user, nil)

		res, err := svc.GetProfile(ctx, user.ID)
		assert.NoError(t, err)
		assert.Equal(t, user.ID.String(), res.ID)
		assert.Equal(t, user.Email, res.Email)
	})

	t.Run("NotFound", func(t *testing.T) {
		id := uuid.New()
		mockRepo.EXPECT().GetByID(ctx, id).Return(nil, domerrors.ErrNotFound)

		res, err := svc.GetProfile(ctx, id)
		assert.Error(t, err)
		assert.Equal(t, domerrors.ErrNotFound, err)
		assert.Nil(t, res)
	})
}
