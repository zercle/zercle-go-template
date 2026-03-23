package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zercle/zercle-go-template/internal/feature/task"
	repository "github.com/zercle/zercle-go-template/internal/feature/task/repository"
)

func TestTaskRepository_Integration(t *testing.T) {
	SkipIfShort(t)

	suite := SetupSuite(t)
	defer suite.Teardown(t)

	ctx := NewContext(t)

	// Test repository operations
	t.Run("create and get task", func(t *testing.T) {
		repo := repository.NewPostgresRepository(suite.DB)

		newTask, err := task.NewTask("Test Task", "Test Description", task.TaskPriorityHigh, "550e8400-e29b-41d4-a716-446655440000")
		require.NoError(t, err)

		created, err := repo.Create(ctx, newTask)
		require.NoError(t, err)
		require.NotNil(t, created)
		assert.NotEmpty(t, created.ID)
		assert.Equal(t, "Test Task", created.Title)
		assert.Equal(t, "Test Description", created.Description)
		assert.Equal(t, task.TaskStatusPending, created.Status)
		assert.Equal(t, task.TaskPriorityHigh, created.Priority)

		// Fetch the created task
		fetched, err := repo.GetByID(ctx, created.ID)
		require.NoError(t, err)
		require.NotNil(t, fetched)
		assert.Equal(t, created.ID, fetched.ID)
		assert.Equal(t, created.Title, fetched.Title)
	})

	t.Run("update task", func(t *testing.T) {
		repo := repository.NewPostgresRepository(suite.DB)

		// Create a task first
		newTask, err := task.NewTask("Original Title", "Original Description", task.TaskPriorityLow, "550e8400-e29b-41d4-a716-446655440000")
		require.NoError(t, err)
		created, err := repo.Create(ctx, newTask)
		require.NoError(t, err)

		// Update the task
		created.Title = "Updated Title"
		created.Description = "Updated Description"
		created.MarkInProgress()
		updated, err := repo.Update(ctx, created)
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, "Updated Title", updated.Title)
		assert.Equal(t, "Updated Description", updated.Description)
		assert.Equal(t, task.TaskStatusInProgress, updated.Status)

		// Verify the update persisted
		fetched, err := repo.GetByID(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Title", fetched.Title)
	})

	t.Run("delete task", func(t *testing.T) {
		repo := repository.NewPostgresRepository(suite.DB)

		// Create a task first
		newTask, err := task.NewTask("Delete Me", "This task will be deleted", task.TaskPriorityMedium, "550e8400-e29b-41d4-a716-446655440000")
		require.NoError(t, err)
		created, err := repo.Create(ctx, newTask)
		require.NoError(t, err)

		// Delete the task
		err = repo.Delete(ctx, created.ID)
		require.NoError(t, err)

		// Verify task is not found
		_, err = repo.GetByID(ctx, created.ID)
		require.Error(t, err)
	})

	t.Run("list tasks with pagination", func(t *testing.T) {
		repo := repository.NewPostgresRepository(suite.DB)
		userID := "550e8400-e29b-41d4-a716-446655440001"

		// Create multiple tasks
		for i := range 5 {
			newTask, err := task.NewTask("List Task "+string(rune('a'+i)), "Description", task.TaskPriorityMedium, userID)
			require.NoError(t, err)
			_, err = repo.Create(ctx, newTask)
			require.NoError(t, err)
		}

		// List with pagination
		listResp, err := repo.List(ctx, &task.ListParams{
			Limit:  3,
			Offset: 0,
		})
		require.NoError(t, err)
		require.NotNil(t, listResp)
		assert.LessOrEqual(t, len(listResp.Tasks), 3)
		assert.Greater(t, listResp.Total, int64(0))
	})

	t.Run("list tasks with user_id filter", func(t *testing.T) {
		repo := repository.NewPostgresRepository(suite.DB)
		userID := "550e8400-e29b-41d4-a716-446655440002"

		// Create a task for specific user
		newTask, err := task.NewTask("User Filter Task", "Description", task.TaskPriorityHigh, userID)
		require.NoError(t, err)
		_, err = repo.Create(ctx, newTask)
		require.NoError(t, err)

		// List with user_id filter
		listResp, err := repo.List(ctx, &task.ListParams{
			Filter: task.TaskFilter{
				UserID: userID,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, listResp)

		// All returned tasks should have the correct user_id
		for _, tk := range listResp.Tasks {
			assert.Equal(t, userID, tk.UserID)
		}
	})

	t.Run("list tasks with status filter", func(t *testing.T) {
		repo := repository.NewPostgresRepository(suite.DB)
		userID := "550e8400-e29b-41d4-a716-446655440003"

		// Create a completed task
		newTask, err := task.NewTask("Completed Task", "Description", task.TaskPriorityHigh, userID)
		require.NoError(t, err)
		created, err := repo.Create(ctx, newTask)
		require.NoError(t, err)
		created.MarkCompleted()
		_, err = repo.Update(ctx, created)
		require.NoError(t, err)

		// List with status filter
		status := task.TaskStatusCompleted
		listResp, err := repo.List(ctx, &task.ListParams{
			Filter: task.TaskFilter{
				Status: &status,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, listResp)

		// All returned tasks should be completed
		for _, tk := range listResp.Tasks {
			assert.Equal(t, task.TaskStatusCompleted, tk.Status)
		}
	})

	t.Run("list tasks with priority filter", func(t *testing.T) {
		repo := repository.NewPostgresRepository(suite.DB)
		userID := "550e8400-e29b-41d4-a716-446655440004"

		// Create a high priority task
		newTask, err := task.NewTask("High Priority Task", "Description", task.TaskPriorityHigh, userID)
		require.NoError(t, err)
		_, err = repo.Create(ctx, newTask)
		require.NoError(t, err)

		// List with priority filter
		priority := task.TaskPriorityHigh
		listResp, err := repo.List(ctx, &task.ListParams{
			Filter: task.TaskFilter{
				Priority: &priority,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, listResp)

		// All returned tasks should be high priority
		for _, tk := range listResp.Tasks {
			assert.Equal(t, task.TaskPriorityHigh, tk.Priority)
		}
	})

	t.Run("count tasks", func(t *testing.T) {
		repo := repository.NewPostgresRepository(suite.DB)
		userID := "550e8400-e29b-41d4-a716-446655440005"

		// Create a task
		newTask, err := task.NewTask("Count Task", "Description", task.TaskPriorityMedium, userID)
		require.NoError(t, err)
		_, err = repo.Create(ctx, newTask)
		require.NoError(t, err)

		// Count with user_id filter
		count, err := repo.Count(ctx, task.TaskFilter{
			UserID: userID,
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(1))
	})

	t.Run("exists by id", func(t *testing.T) {
		repo := repository.NewPostgresRepository(suite.DB)

		// Create a task
		newTask, err := task.NewTask("Exists Task", "Description", task.TaskPriorityLow, "550e8400-e29b-41d4-a716-446655440000")
		require.NoError(t, err)
		created, err := repo.Create(ctx, newTask)
		require.NoError(t, err)

		// Check exists
		exists, err := repo.ExistsByID(ctx, created.ID)
		require.NoError(t, err)
		assert.True(t, exists)

		// Check non-existent
		exists, err = repo.ExistsByID(ctx, "nonexistent-id")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("task not found", func(t *testing.T) {
		repo := repository.NewPostgresRepository(suite.DB)

		// Try to get a non-existent task
		_, err := repo.GetByID(ctx, "nonexistent-id")
		require.Error(t, err)
	})

	t.Run("task status transitions", func(t *testing.T) {
		repo := repository.NewPostgresRepository(suite.DB)

		// Create a task
		newTask, err := task.NewTask("Status Transition Task", "Description", task.TaskPriorityMedium, "550e8400-e29b-41d4-a716-446655440000")
		require.NoError(t, err)
		created, err := repo.Create(ctx, newTask)
		require.NoError(t, err)
		assert.Equal(t, task.TaskStatusPending, created.Status)

		// Mark in progress
		created.MarkInProgress()
		updated, err := repo.Update(ctx, created)
		require.NoError(t, err)
		assert.Equal(t, task.TaskStatusInProgress, updated.Status)

		// Mark completed
		updated.MarkCompleted()
		updatedAgain, err := repo.Update(ctx, updated)
		require.NoError(t, err)
		assert.Equal(t, task.TaskStatusCompleted, updatedAgain.Status)

		// Cancel
		updatedAgain.Cancel()
		cancelled, err := repo.Update(ctx, updatedAgain)
		require.NoError(t, err)
		assert.Equal(t, task.TaskStatusCancelled, cancelled.Status)
	})
}
