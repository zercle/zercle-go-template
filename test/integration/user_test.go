package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zercle/zercle-go-template/internal/feature/user"
	repository "github.com/zercle/zercle-go-template/internal/feature/user/repository"
)

func TestUserRepository_Integration(t *testing.T) {
	SkipIfShort(t)

	suite := SetupSuite(t)
	defer suite.Teardown(t)

	ctx := NewContext(t)

	// Test repository operations
	t.Run("create and get user", func(t *testing.T) {
		repo := repository.NewPostgresRepository(suite.DB)

		newUser, err := user.NewUser("integration@example.com", "hashedpassword", "Test", "User")
		require.NoError(t, err)

		created, err := repo.Create(ctx, newUser)
		require.NoError(t, err)
		require.NotNil(t, created)
		assert.NotEmpty(t, created.ID)
		assert.Equal(t, "integration@example.com", created.Email)
		assert.Equal(t, "Test", created.FirstName)
		assert.Equal(t, "User", created.LastName)
		assert.Equal(t, user.UserStatusActive, created.Status)

		// Fetch the created user
		fetched, err := repo.GetByID(ctx, created.ID)
		require.NoError(t, err)
		require.NotNil(t, fetched)
		assert.Equal(t, created.ID, fetched.ID)
		assert.Equal(t, created.Email, fetched.Email)
	})

	t.Run("get by email", func(t *testing.T) {
		repo := repository.NewPostgresRepository(suite.DB)

		// Create a user first
		newUser, err := user.NewUser("getbyemail@example.com", "hashedpassword", "GetBy", "Email")
		require.NoError(t, err)
		created, err := repo.Create(ctx, newUser)
		require.NoError(t, err)

		// Get by email
		fetched, err := repo.GetByEmail(ctx, "getbyemail@example.com")
		require.NoError(t, err)
		require.NotNil(t, fetched)
		assert.Equal(t, created.ID, fetched.ID)
		assert.Equal(t, "getbyemail@example.com", fetched.Email)
	})

	t.Run("update user", func(t *testing.T) {
		repo := repository.NewPostgresRepository(suite.DB)

		// Create a user first
		newUser, err := user.NewUser("update@example.com", "hashedpassword", "Original", "Name")
		require.NoError(t, err)
		created, err := repo.Create(ctx, newUser)
		require.NoError(t, err)

		// Update the user
		created.FirstName = "Updated"
		created.LastName = "Name"
		updated, err := repo.Update(ctx, created)
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, "Updated", updated.FirstName)
		assert.Equal(t, "Name", updated.LastName)

		// Verify the update persisted
		fetched, err := repo.GetByID(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated", fetched.FirstName)
	})

	t.Run("delete user", func(t *testing.T) {
		repo := repository.NewPostgresRepository(suite.DB)

		// Create a user first
		newUser, err := user.NewUser("delete@example.com", "hashedpassword", "Delete", "Me")
		require.NoError(t, err)
		created, err := repo.Create(ctx, newUser)
		require.NoError(t, err)

		// Delete the user
		err = repo.Delete(ctx, created.ID)
		require.NoError(t, err)

		// Verify user is not found
		_, err = repo.GetByID(ctx, created.ID)
		require.Error(t, err)
	})

	t.Run("list users with pagination", func(t *testing.T) {
		repo := repository.NewPostgresRepository(suite.DB)

		// Create multiple users
		for i := range 5 {
			newUser, err := user.NewUser("listuser"+string(rune('a'+i))+"@example.com", "hashedpassword", "List", "User")
			require.NoError(t, err)
			_, err = repo.Create(ctx, newUser)
			require.NoError(t, err)
		}

		// List with pagination
		listResp, err := repo.List(ctx, &user.ListParams{
			Limit:  3,
			Offset: 0,
		})
		require.NoError(t, err)
		require.NotNil(t, listResp)
		assert.LessOrEqual(t, len(listResp.Users), 3)
		assert.Greater(t, listResp.Total, int64(0))
	})

	t.Run("list users with status filter", func(t *testing.T) {
		repo := repository.NewPostgresRepository(suite.DB)

		// Create an active user
		newUser, err := user.NewUser("activefilter@example.com", "hashedpassword", "Active", "User")
		require.NoError(t, err)
		_, err = repo.Create(ctx, newUser)
		require.NoError(t, err)

		// List with status filter
		status := user.UserStatusActive
		listResp, err := repo.List(ctx, &user.ListParams{
			Status: &status,
		})
		require.NoError(t, err)
		require.NotNil(t, listResp)

		// All returned users should be active
		for _, u := range listResp.Users {
			assert.Equal(t, user.UserStatusActive, u.Status)
		}
	})

	t.Run("duplicate email error", func(t *testing.T) {
		repo := repository.NewPostgresRepository(suite.DB)

		email := "duplicate@example.com"
		newUser1, err := user.NewUser(email, "hashedpassword", "First", "User")
		require.NoError(t, err)
		_, err = repo.Create(ctx, newUser1)
		require.NoError(t, err)

		// Try to create another user with the same email
		newUser2, err := user.NewUser(email, "hashedpassword2", "Second", "User")
		require.NoError(t, err) // NewUser should succeed
		_, err = repo.Create(ctx, newUser2)
		require.Error(t, err) // Create should fail with duplicate email
	})

	t.Run("user not found", func(t *testing.T) {
		repo := repository.NewPostgresRepository(suite.DB)

		// Try to get a non-existent user
		_, err := repo.GetByID(ctx, "nonexistent-id")
		require.Error(t, err)
	})
}

// TestUserStatusTransitions tests user status transitions via repository.
func TestUserStatusTransitions_Integration(t *testing.T) {
	SkipIfShort(t)

	suite := SetupSuite(t)
	defer suite.Teardown(t)

	ctx := NewContext(t)

	repo := repository.NewPostgresRepository(suite.DB)

	// Create a user
	testUser, err := user.NewUser("statustest@example.com", "hash", "Status", "Test")
	require.NoError(t, err)

	created, err := repo.Create(ctx, testUser)
	require.NoError(t, err)
	assert.Equal(t, user.UserStatusActive, created.Status)

	// Suspend user
	created.Suspend()
	updated, err := repo.Update(ctx, created)
	require.NoError(t, err)
	assert.Equal(t, user.UserStatusSuspended, updated.Status)

	// Reactivate user
	updated.Activate()
	updatedAgain, err := repo.Update(ctx, updated)
	require.NoError(t, err)
	assert.Equal(t, user.UserStatusActive, updatedAgain.Status)

	// Deactivate user
	updatedAgain.Deactivate()
	deactivated, err := repo.Update(ctx, updatedAgain)
	require.NoError(t, err)
	assert.Equal(t, user.UserStatusInactive, deactivated.Status)
}
