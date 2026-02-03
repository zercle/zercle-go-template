package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zercle/zercle-go-template/internal/feature/task"
)

func TestTaskRepository_Integration(t *testing.T) {
	SkipIfShort(t)
	SkipIfPodman(t)

	suite := SetupSuite(t)
	defer suite.Teardown(t)

	ctx := NewContext(t)

	// Test repository operations
	t.Run("create and get task", func(t *testing.T) {
		repo := task.NewPostgresRepository(suite.DB)

		newTask, err := task.New("Test Task", "Test Description", task.PriorityHigh, "550e8400-e29b-41d4-a716-446655440000")
		require.NoError(t, err)

		created, err := repo.Create(ctx, newTask)
		require.NoError(t, err)
		require.NotNil(t, created)
		assert.NotEmpty(t, created.ID)
		assert.Equal(t, "Test Task", created.Title)
		assert.Equal(t, "Test Description", created.Description)
		assert.Equal(t, task.StatusPending, created.Status)
		assert.Equal(t, task.PriorityHigh, created.Priority)

		// Fetch the created task
		fetched, err := repo.GetByID(ctx, created.ID)
		require.NoError(t, err)
		require.NotNil(t, fetched)
		assert.Equal(t, created.ID, fetched.ID)
		assert.Equal(t, created.Title, fetched.Title)
	})

	t.Run("update task", func(t *testing.T) {
		repo := task.NewPostgresRepository(suite.DB)

		// Create a task first
		newTask, err := task.New("Original Title", "Original Description", task.PriorityLow, "550e8400-e29b-41d4-a716-446655440000")
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
		assert.Equal(t, task.StatusInProgress, updated.Status)

		// Verify the update persisted
		fetched, err := repo.GetByID(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Title", fetched.Title)
	})

	t.Run("delete task", func(t *testing.T) {
		repo := task.NewPostgresRepository(suite.DB)

		// Create a task first
		newTask, err := task.New("Delete Me", "This task will be deleted", task.PriorityMedium, "550e8400-e29b-41d4-a716-446655440000")
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
		repo := task.NewPostgresRepository(suite.DB)
		userID := "550e8400-e29b-41d4-a716-446655440001"

		// Create multiple tasks
		for i := range 5 {
			newTask, err := task.New("List Task "+string(rune('a'+i)), "Description", task.PriorityMedium, userID)
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
		assert.Positive(t, listResp.Total)
	})

	t.Run("list tasks with user_id filter", func(t *testing.T) {
		repo := task.NewPostgresRepository(suite.DB)
		userID := "550e8400-e29b-41d4-a716-446655440002"

		// Create a task for specific user
		newTask, err := task.New("User Filter Task", "Description", task.PriorityHigh, userID)
		require.NoError(t, err)
		_, err = repo.Create(ctx, newTask)
		require.NoError(t, err)

		// List with user_id filter
		listResp, err := repo.List(ctx, &task.ListParams{
			Filter: task.Filter{
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
		repo := task.NewPostgresRepository(suite.DB)
		userID := "550e8400-e29b-41d4-a716-446655440003"

		// Create a completed task
		newTask, err := task.New("Completed Task", "Description", task.PriorityHigh, userID)
		require.NoError(t, err)
		created, err := repo.Create(ctx, newTask)
		require.NoError(t, err)
		created.MarkCompleted()
		_, err = repo.Update(ctx, created)
		require.NoError(t, err)

		// List with status filter
		status := task.StatusCompleted
		listResp, err := repo.List(ctx, &task.ListParams{
			Filter: task.Filter{
				Status: &status,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, listResp)

		// All returned tasks should be completed
		for _, tk := range listResp.Tasks {
			assert.Equal(t, task.StatusCompleted, tk.Status)
		}
	})

	t.Run("list tasks with priority filter", func(t *testing.T) {
		repo := task.NewPostgresRepository(suite.DB)
		userID := "550e8400-e29b-41d4-a716-446655440004"

		// Create a high priority task
		newTask, err := task.New("High Priority Task", "Description", task.PriorityHigh, userID)
		require.NoError(t, err)
		_, err = repo.Create(ctx, newTask)
		require.NoError(t, err)

		// List with priority filter
		priority := task.PriorityHigh
		listResp, err := repo.List(ctx, &task.ListParams{
			Filter: task.Filter{
				Priority: &priority,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, listResp)

		// All returned tasks should be high priority
		for _, tk := range listResp.Tasks {
			assert.Equal(t, task.PriorityHigh, tk.Priority)
		}
	})

	t.Run("count tasks", func(t *testing.T) {
		repo := task.NewPostgresRepository(suite.DB)
		userID := "550e8400-e29b-41d4-a716-446655440005"

		// Create a task
		newTask, err := task.New("Count Task", "Description", task.PriorityMedium, userID)
		require.NoError(t, err)
		_, err = repo.Create(ctx, newTask)
		require.NoError(t, err)

		// Count with user_id filter
		count, err := repo.Count(ctx, task.Filter{
			UserID: userID,
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(1))
	})

	t.Run("exists by id", func(t *testing.T) {
		repo := task.NewPostgresRepository(suite.DB)

		// Create a task
		newTask, err := task.New("Exists Task", "Description", task.PriorityLow, "550e8400-e29b-41d4-a716-446655440000")
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
		repo := task.NewPostgresRepository(suite.DB)

		// Try to get a non-existent task
		_, err := repo.GetByID(ctx, "nonexistent-id")
		require.Error(t, err)
	})

	t.Run("task status transitions", func(t *testing.T) {
		repo := task.NewPostgresRepository(suite.DB)

		// Create a task
		newTask, err := task.New("Status Transition Task", "Description", task.PriorityMedium, "550e8400-e29b-41d4-a716-446655440000")
		require.NoError(t, err)
		created, err := repo.Create(ctx, newTask)
		require.NoError(t, err)
		assert.Equal(t, task.StatusPending, created.Status)

		// Mark in progress
		created.MarkInProgress()
		updated, err := repo.Update(ctx, created)
		require.NoError(t, err)
		assert.Equal(t, task.StatusInProgress, updated.Status)

		// Mark completed
		updated.MarkCompleted()
		updatedAgain, err := repo.Update(ctx, updated)
		require.NoError(t, err)
		assert.Equal(t, task.StatusCompleted, updatedAgain.Status)

		// Cancel
		updatedAgain.Cancel()
		cancelled, err := repo.Update(ctx, updatedAgain)
		require.NoError(t, err)
		assert.Equal(t, task.StatusCancelled, cancelled.Status)
	})
}
