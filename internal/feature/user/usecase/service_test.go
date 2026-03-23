package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zercle/zercle-go-template/internal/feature/user"
	mockuser "github.com/zercle/zercle-go-template/internal/feature/user/mock"
	"go.uber.org/mock/gomock"
)

func TestNewService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockuser.NewMockRepository(ctrl)

	// Use the exported NewService constructor
	s := NewService(mockRepo)

	assert.NotNil(t, s)
	// Verify it implements the interface
	var _ = s
}

func TestCreate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockuser.NewMockRepository(ctrl)
	s := NewService(mockRepo)

	ctx := context.Background()
	input := &user.CreateUserDTO{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}

	expectedUser := &user.User{
		ID:           user.UserID("550e8400-e29b-41d4-a716-446655440001"),
		Email:        input.Email,
		PasswordHash: input.Password,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Status:       user.UserStatusActive,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	mockRepo.EXPECT().
		Exists(ctx, input.Email).
		Return(false, nil)

	mockRepo.EXPECT().
		Create(ctx, gomock.Any()).
		Return(expectedUser, nil)

	result, err := s.Create(ctx, input)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, string(expectedUser.ID), result.ID)
	assert.Equal(t, expectedUser.Email, result.Email)
	assert.Equal(t, expectedUser.FirstName, result.FirstName)
	assert.Equal(t, expectedUser.LastName, result.LastName)
}

func TestCreate_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockuser.NewMockRepository(ctrl)
	s := NewService(mockRepo)

	ctx := context.Background()

	testCases := []struct {
		name   string
		input  *user.CreateUserDTO
		errMsg string
	}{
		{
			name:   "nil input",
			input:  nil,
			errMsg: "create request is required",
		},
		{
			name:   "empty email",
			input:  &user.CreateUserDTO{Email: "", Password: "password123", FirstName: "John", LastName: "Doe"},
			errMsg: "email is required",
		},
		{
			name:   "empty password",
			input:  &user.CreateUserDTO{Email: "test@example.com", Password: "", FirstName: "John", LastName: "Doe"},
			errMsg: "password is required",
		},
		{
			name:   "empty first name",
			input:  &user.CreateUserDTO{Email: "test@example.com", Password: "password123", FirstName: "", LastName: "Doe"},
			errMsg: "first name is required",
		},
		{
			name:   "empty last name",
			input:  &user.CreateUserDTO{Email: "test@example.com", Password: "password123", FirstName: "John", LastName: ""},
			errMsg: "last name is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := s.Create(ctx, tc.input)

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Equal(t, tc.errMsg, err.Error())
		})
	}
}

func TestCreate_DuplicateEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockuser.NewMockRepository(ctrl)
	s := NewService(mockRepo)

	ctx := context.Background()
	input := &user.CreateUserDTO{
		Email:     "existing@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}

	mockRepo.EXPECT().
		Exists(ctx, input.Email).
		Return(true, nil)

	result, err := s.Create(ctx, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, user.ErrDuplicateEmail)
}

func TestGetByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockuser.NewMockRepository(ctrl)
	s := NewService(mockRepo)

	ctx := context.Background()
	userID := "550e8400-e29b-41d4-a716-446655440001"

	expectedUser := &user.User{
		ID:        user.UserID(userID),
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Status:    user.UserStatusActive,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	mockRepo.EXPECT().
		GetByID(ctx, user.UserID(userID)).
		Return(expectedUser, nil)

	result, err := s.GetByID(ctx, userID)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.ID)
	assert.Equal(t, expectedUser.Email, result.Email)
}

func TestGetByID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockuser.NewMockRepository(ctrl)
	s := NewService(mockRepo)

	ctx := context.Background()
	userID := "nonexistent-id"

	mockRepo.EXPECT().
		GetByID(ctx, user.UserID(userID)).
		Return(nil, user.ErrUserNotFound)

	result, err := s.GetByID(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, user.ErrUserNotFound)
}

func TestGetByEmail_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockuser.NewMockRepository(ctrl)
	s := NewService(mockRepo)

	ctx := context.Background()
	email := "test@example.com"

	expectedUser := &user.User{
		ID:        user.UserID("550e8400-e29b-41d4-a716-446655440001"),
		Email:     email,
		FirstName: "John",
		LastName:  "Doe",
		Status:    user.UserStatusActive,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	mockRepo.EXPECT().
		GetByEmail(ctx, email).
		Return(expectedUser, nil)

	result, err := s.GetByEmail(ctx, email)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, email, result.Email)
}

func TestGetByEmail_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockuser.NewMockRepository(ctrl)
	s := NewService(mockRepo)

	ctx := context.Background()
	email := "nonexistent@example.com"

	mockRepo.EXPECT().
		GetByEmail(ctx, email).
		Return(nil, user.ErrUserNotFound)

	result, err := s.GetByEmail(ctx, email)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, user.ErrUserNotFound)
}

func TestList_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockuser.NewMockRepository(ctrl)
	s := NewService(mockRepo)

	ctx := context.Background()
	params := &user.ListParamsDTO{
		Limit:  10,
		Offset: 0,
	}

	expectedUsers := []*user.User{
		{
			ID:        user.UserID("user-1"),
			Email:     "user1@example.com",
			FirstName: "User",
			LastName:  "One",
			Status:    user.UserStatusActive,
		},
		{
			ID:        user.UserID("user-2"),
			Email:     "user2@example.com",
			FirstName: "User",
			LastName:  "Two",
			Status:    user.UserStatusActive,
		},
	}

	expectedResult := &user.ListResult{
		Users:  expectedUsers,
		Total:  2,
		Limit:  10,
		Offset: 0,
	}

	mockRepo.EXPECT().
		List(ctx, gomock.Any()).
		Return(expectedResult, nil)

	result, err := s.List(ctx, params)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Users, 2)
	assert.Equal(t, int64(2), result.Total)
	assert.Equal(t, int32(10), result.Limit)
	assert.Equal(t, int32(0), result.Offset)
}

func TestList_WithPagination(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockuser.NewMockRepository(ctrl)
	s := NewService(mockRepo)

	ctx := context.Background()

	t.Run("default pagination when limit is zero", func(t *testing.T) {
		params := &user.ListParamsDTO{
			Limit:  0,
			Offset: 0,
		}

		expectedResult := &user.ListResult{
			Users:  []*user.User{},
			Total:  0,
			Limit:  20, // default limit
			Offset: 0,
		}

		mockRepo.EXPECT().
			List(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, params *user.ListParams) (*user.ListResult, error) {
				assert.Equal(t, int32(20), params.Limit) // default limit applied
				return expectedResult, nil
			})

		result, err := s.List(ctx, params)

		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("limit capped at 100", func(t *testing.T) {
		params := &user.ListParamsDTO{
			Limit:  200, // exceeds max
			Offset: 0,
		}

		expectedResult := &user.ListResult{
			Users:  []*user.User{},
			Total:  0,
			Limit:  100, // capped limit
			Offset: 0,
		}

		mockRepo.EXPECT().
			List(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, params *user.ListParams) (*user.ListResult, error) {
				assert.Equal(t, int32(100), params.Limit) // capped limit
				return expectedResult, nil
			})

		result, err := s.List(ctx, params)

		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestUpdate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockuser.NewMockRepository(ctrl)
	s := NewService(mockRepo)

	ctx := context.Background()
	userID := "550e8400-e29b-41d4-a716-446655440001"

	existingUser := &user.User{
		ID:           user.UserID(userID),
		Email:        "old@example.com",
		FirstName:    "Old",
		LastName:     "Name",
		Status:       user.UserStatusActive,
		PasswordHash: "hash",
		CreatedAt:    time.Now().UTC().Add(-24 * time.Hour),
		UpdatedAt:    time.Now().UTC().Add(-24 * time.Hour),
	}

	newFirstName := "New"
	newLastName := "Name"
	input := &user.UpdateUserDTO{
		FirstName: &newFirstName,
		LastName:  &newLastName,
	}

	updatedUser := &user.User{
		ID:           existingUser.ID,
		Email:        existingUser.Email,
		FirstName:    newFirstName,
		LastName:     newLastName,
		Status:       existingUser.Status,
		PasswordHash: existingUser.PasswordHash,
		CreatedAt:    existingUser.CreatedAt,
		UpdatedAt:    time.Now().UTC(),
	}

	mockRepo.EXPECT().
		GetByID(ctx, user.UserID(userID)).
		Return(existingUser, nil)

	mockRepo.EXPECT().
		Update(ctx, gomock.Any()).
		Return(updatedUser, nil)

	result, err := s.Update(ctx, userID, input)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, newFirstName, result.FirstName)
	assert.Equal(t, newLastName, result.LastName)
}

func TestUpdate_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockuser.NewMockRepository(ctrl)
	s := NewService(mockRepo)

	ctx := context.Background()
	userID := "nonexistent-id"

	newFirstName := "New"
	input := &user.UpdateUserDTO{
		FirstName: &newFirstName,
	}

	mockRepo.EXPECT().
		GetByID(ctx, user.UserID(userID)).
		Return(nil, user.ErrUserNotFound)

	result, err := s.Update(ctx, userID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, user.ErrUserNotFound)
}

func TestUpdate_DuplicateEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockuser.NewMockRepository(ctrl)
	s := NewService(mockRepo)

	ctx := context.Background()
	userID := "550e8400-e29b-41d4-a716-446655440001"

	existingUser := &user.User{
		ID:           user.UserID(userID),
		Email:        "old@example.com",
		FirstName:    "Old",
		LastName:     "Name",
		Status:       user.UserStatusActive,
		PasswordHash: "hash",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	newEmail := "existing@example.com"
	input := &user.UpdateUserDTO{
		Email: &newEmail,
	}

	mockRepo.EXPECT().
		GetByID(ctx, user.UserID(userID)).
		Return(existingUser, nil)

	mockRepo.EXPECT().
		Exists(ctx, newEmail).
		Return(true, nil)

	result, err := s.Update(ctx, userID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, user.ErrDuplicateEmail)
}

func TestUpdate_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockuser.NewMockRepository(ctrl)
	s := NewService(mockRepo)

	ctx := context.Background()
	userID := "550e8400-e29b-41d4-a716-446655440001"

	t.Run("nil input", func(t *testing.T) {
		result, err := s.Update(ctx, userID, nil)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "update request is required", err.Error())
	})

	t.Run("empty user id", func(t *testing.T) {
		result, err := s.Update(ctx, "", &user.UpdateUserDTO{})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, user.ErrUserNotFound)
	})
}

func TestDelete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockuser.NewMockRepository(ctrl)
	s := NewService(mockRepo)

	ctx := context.Background()
	userID := "550e8400-e29b-41d4-a716-446655440001"

	mockRepo.EXPECT().
		ExistsByID(ctx, user.UserID(userID)).
		Return(true, nil)

	mockRepo.EXPECT().
		Delete(ctx, user.UserID(userID)).
		Return(nil)

	err := s.Delete(ctx, userID)

	require.NoError(t, err)
}

func TestDelete_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockuser.NewMockRepository(ctrl)
	s := NewService(mockRepo)

	ctx := context.Background()
	userID := "nonexistent-id"

	mockRepo.EXPECT().
		ExistsByID(ctx, user.UserID(userID)).
		Return(false, nil)

	err := s.Delete(ctx, userID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, user.ErrUserNotFound)
}
