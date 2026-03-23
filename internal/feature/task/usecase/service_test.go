package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zercle/zercle-go-template/internal/feature/task"
	mocktask "github.com/zercle/zercle-go-template/internal/feature/task/mock"
	"go.uber.org/mock/gomock"
)

func TestNewService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocktask.NewMockRepository(ctrl)

	// Direct instantiation of service struct for testing
	s := &service{repo: mockRepo}

	assert.NotNil(t, s)
	assert.Equal(t, mockRepo, s.repo)
}

func TestCreate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocktask.NewMockRepository(ctrl)
	s := &service{repo: mockRepo}

	ctx := context.Background()
	input := &task.CreateTaskInput{
		Title:       "Test Task",
		Description: "Test Description",
		Priority:    "high",
		UserID:      "550e8400-e29b-41d4-a716-446655440000",
	}

	expectedTask := &task.Task{
		ID:          task.TaskID("550e8400-e29b-41d4-a716-446655440001"),
		Title:       input.Title,
		Description: input.Description,
		Status:      task.TaskStatusPending,
		Priority:    task.TaskPriorityHigh,
		UserID:      input.UserID,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	mockRepo.EXPECT().
		Create(ctx, gomock.Any()).
		Return(expectedTask, nil)

	result, err := s.Create(ctx, input)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, string(expectedTask.ID), result.ID)
	assert.Equal(t, expectedTask.Title, result.Title)
	assert.Equal(t, expectedTask.Description, result.Description)
	assert.Equal(t, string(expectedTask.Status), result.Status)
	assert.Equal(t, string(expectedTask.Priority), result.Priority)
}

func TestCreate_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocktask.NewMockRepository(ctrl)
	s := &service{repo: mockRepo}

	ctx := context.Background()

	testCases := []struct {
		name   string
		input  *task.CreateTaskInput
		errMsg string
	}{
		{
			name:   "nil input",
			input:  nil,
			errMsg: "create input is required",
		},
		{
			name:   "empty title",
			input:  &task.CreateTaskInput{Title: "", Priority: "high", UserID: "550e8400-e29b-41d4-a716-446655440000"},
			errMsg: "title is required",
		},
		{
			name:   "empty priority",
			input:  &task.CreateTaskInput{Title: "Test", Priority: "", UserID: "550e8400-e29b-41d4-a716-446655440000"},
			errMsg: "priority is required",
		},
		{
			name:   "empty user_id",
			input:  &task.CreateTaskInput{Title: "Test", Priority: "high", UserID: ""},
			errMsg: "user_id is required",
		},
		{
			name:   "invalid priority",
			input:  &task.CreateTaskInput{Title: "Test", Priority: "invalid", UserID: "550e8400-e29b-41d4-a716-446655440000"},
			errMsg: task.ErrInvalidTaskPriority.Error(),
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

func TestGet_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocktask.NewMockRepository(ctrl)
	s := &service{repo: mockRepo}

	ctx := context.Background()
	taskID := "550e8400-e29b-41d4-a716-446655440001"

	expectedTask := &task.Task{
		ID:          task.TaskID(taskID),
		Title:       "Test Task",
		Description: "Test Description",
		Status:      task.TaskStatusPending,
		Priority:    task.TaskPriorityHigh,
		UserID:      "550e8400-e29b-41d4-a716-446655440000",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	mockRepo.EXPECT().
		GetByID(ctx, task.TaskID(taskID)).
		Return(expectedTask, nil)

	result, err := s.Get(ctx, taskID)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, taskID, result.ID)
	assert.Equal(t, expectedTask.Title, result.Title)
}

func TestGet_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocktask.NewMockRepository(ctrl)
	s := &service{repo: mockRepo}

	ctx := context.Background()
	taskID := "nonexistent-id"

	mockRepo.EXPECT().
		GetByID(ctx, task.TaskID(taskID)).
		Return(nil, task.ErrTaskNotFound)

	result, err := s.Get(ctx, taskID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, task.ErrTaskNotFound)
}

func TestList_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocktask.NewMockRepository(ctrl)
	s := &service{repo: mockRepo}

	ctx := context.Background()
	params := &task.ListParamsDTO{
		Limit:  10,
		Offset: 0,
	}

	expectedTasks := []*task.Task{
		{
			ID:       task.TaskID("task-1"),
			Title:    "Task 1",
			Status:   task.TaskStatusPending,
			Priority: task.TaskPriorityHigh,
			UserID:   "user-1",
		},
		{
			ID:       task.TaskID("task-2"),
			Title:    "Task 2",
			Status:   task.TaskStatusInProgress,
			Priority: task.TaskPriorityMedium,
			UserID:   "user-1",
		},
	}

	expectedResult := &task.ListResult{
		Tasks:  expectedTasks,
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
	assert.Len(t, result.Tasks, 2)
	assert.Equal(t, int64(2), result.Total)
	assert.Equal(t, int32(10), result.Limit)
	assert.Equal(t, int32(0), result.Offset)
}

func TestList_WithPagination(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocktask.NewMockRepository(ctrl)
	s := &service{repo: mockRepo}

	ctx := context.Background()

	t.Run("default pagination when limit is zero", func(t *testing.T) {
		params := &task.ListParamsDTO{
			Limit:  0,
			Offset: 0,
		}

		expectedResult := &task.ListResult{
			Tasks:  []*task.Task{},
			Total:  0,
			Limit:  20, // default limit
			Offset: 0,
		}

		mockRepo.EXPECT().
			List(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, params *task.ListParams) (*task.ListResult, error) {
				assert.Equal(t, int32(20), params.Limit) // default limit applied
				return expectedResult, nil
			})

		result, err := s.List(ctx, params)

		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("limit capped at 100", func(t *testing.T) {
		params := &task.ListParamsDTO{
			Limit:  200, // exceeds max
			Offset: 0,
		}

		expectedResult := &task.ListResult{
			Tasks:  []*task.Task{},
			Total:  0,
			Limit:  100, // capped limit
			Offset: 0,
		}

		mockRepo.EXPECT().
			List(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, params *task.ListParams) (*task.ListResult, error) {
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

	mockRepo := mocktask.NewMockRepository(ctrl)
	s := &service{repo: mockRepo}

	ctx := context.Background()
	taskID := "550e8400-e29b-41d4-a716-446655440001"

	existingTask := &task.Task{
		ID:          task.TaskID(taskID),
		Title:       "Original Title",
		Description: "Original Description",
		Status:      task.TaskStatusPending,
		Priority:    task.TaskPriorityLow,
		UserID:      "550e8400-e29b-41d4-a716-446655440000",
		CreatedAt:   time.Now().UTC().Add(-24 * time.Hour),
		UpdatedAt:   time.Now().UTC().Add(-24 * time.Hour),
	}

	newTitle := "Updated Title"
	newStatus := "completed"
	input := &task.UpdateTaskInput{
		Title:  &newTitle,
		Status: &newStatus,
	}

	updatedTask := &task.Task{
		ID:          existingTask.ID,
		Title:       newTitle,
		Description: existingTask.Description,
		Status:      task.TaskStatusCompleted,
		Priority:    existingTask.Priority,
		UserID:      existingTask.UserID,
		CreatedAt:   existingTask.CreatedAt,
		UpdatedAt:   time.Now().UTC(),
	}

	mockRepo.EXPECT().
		GetByID(ctx, task.TaskID(taskID)).
		Return(existingTask, nil)

	mockRepo.EXPECT().
		Update(ctx, gomock.Any()).
		Return(updatedTask, nil)

	result, err := s.Update(ctx, taskID, input)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, newTitle, result.Title)
	assert.Equal(t, string(task.TaskStatusCompleted), result.Status)
}

func TestUpdate_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocktask.NewMockRepository(ctrl)
	s := &service{repo: mockRepo}

	ctx := context.Background()
	taskID := "nonexistent-id"

	newTitle := "Updated Title"
	input := &task.UpdateTaskInput{
		Title: &newTitle,
	}

	mockRepo.EXPECT().
		GetByID(ctx, task.TaskID(taskID)).
		Return(nil, task.ErrTaskNotFound)

	result, err := s.Update(ctx, taskID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, task.ErrTaskNotFound)
}

func TestUpdate_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocktask.NewMockRepository(ctrl)
	s := &service{repo: mockRepo}

	ctx := context.Background()
	taskID := "550e8400-e29b-41d4-a716-446655440001"

	t.Run("nil input", func(t *testing.T) {
		result, err := s.Update(ctx, taskID, nil)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "update input is required", err.Error())
	})

	t.Run("empty task id", func(t *testing.T) {
		result, err := s.Update(ctx, "", &task.UpdateTaskInput{})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, task.ErrTaskNotFound)
	})

	t.Run("invalid status", func(t *testing.T) {
		existingTask := &task.Task{
			ID:       task.TaskID(taskID),
			Title:    "Original Title",
			Status:   task.TaskStatusPending,
			Priority: task.TaskPriorityLow,
			UserID:   "550e8400-e29b-41d4-a716-446655440000",
		}

		mockRepo.EXPECT().
			GetByID(ctx, task.TaskID(taskID)).
			Return(existingTask, nil)

		result, err := s.Update(ctx, taskID, &task.UpdateTaskInput{
			Status: new("invalid_status"),
		})

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, task.ErrInvalidTaskStatus.Error(), err.Error())
	})

	t.Run("invalid priority", func(t *testing.T) {
		existingTask := &task.Task{
			ID:       task.TaskID(taskID),
			Title:    "Original Title",
			Status:   task.TaskStatusPending,
			Priority: task.TaskPriorityLow,
			UserID:   "550e8400-e29b-41d4-a716-446655440000",
		}

		mockRepo.EXPECT().
			GetByID(ctx, task.TaskID(taskID)).
			Return(existingTask, nil)

		result, err := s.Update(ctx, taskID, &task.UpdateTaskInput{
			Priority: new("invalid_priority"),
		})

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, task.ErrInvalidTaskPriority.Error(), err.Error())
	})
}

func TestDelete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocktask.NewMockRepository(ctrl)
	s := &service{repo: mockRepo}

	ctx := context.Background()
	taskID := "550e8400-e29b-41d4-a716-446655440001"

	mockRepo.EXPECT().
		ExistsByID(ctx, task.TaskID(taskID)).
		Return(true, nil)

	mockRepo.EXPECT().
		Delete(ctx, task.TaskID(taskID)).
		Return(nil)

	err := s.Delete(ctx, taskID)

	require.NoError(t, err)
}

func TestDelete_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocktask.NewMockRepository(ctrl)
	s := &service{repo: mockRepo}

	ctx := context.Background()
	taskID := "nonexistent-id"

	mockRepo.EXPECT().
		ExistsByID(ctx, task.TaskID(taskID)).
		Return(false, nil)

	err := s.Delete(ctx, taskID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, task.ErrTaskNotFound)
}

// Helper function to create string pointer
//
//go:fix inline
func strPtr(s string) *string {
	return new(s)
}
