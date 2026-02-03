//go:build integration

// Package repository provides data access layer implementations for the user feature.
// This file contains integration tests for the SQLC-based user repository.
package repository

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"zercle-go-template/internal/errors"
	"zercle-go-template/internal/feature/user/domain"
	"zercle-go-template/internal/infrastructure/db/sqlc"
)

// testDB holds the database connection for integration tests.
var testDB *pgxpool.Pool

// getEnvOrDefault returns the value of an environment variable or a default value.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getTestDSN returns the database connection string for tests.
func getTestDSN() string {
	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")
	dbName := getEnvOrDefault("DB_NAME", "zercle_template_test")
	user := getEnvOrDefault("DB_USER", "postgres")
	password := getEnvOrDefault("DB_PASSWORD", "postgres")
	sslMode := getEnvOrDefault("DB_SSL_MODE", "disable")

	return fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		host, port, dbName, user, password, sslMode)
}

// setupTestDB initializes the test database connection.
func setupTestDB() (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, getTestDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

// TestMain runs before and after all tests in this package.
func TestMain(m *testing.M) {
	var err error
	testDB, err = setupTestDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup test database: %v\n", err)
		os.Exit(1)
	}

	// Run all tests
	code := m.Run()

	// Final cleanup after all tests complete
	ctx := context.Background()
	_, _ = testDB.Exec(ctx, "DELETE FROM users WHERE email LIKE 'test-%@example.com'")
	_, _ = testDB.Exec(ctx, "DELETE FROM users WHERE email LIKE 'integration-%@example.com'")
	_, _ = testDB.Exec(ctx, "DELETE FROM users WHERE email LIKE 'concurrent-%@example.com'")

	if testDB != nil {
		testDB.Close()
	}

	os.Exit(code)
}

// createTestUser creates a unique test user for integration tests.
func createTestUser(t *testing.T) *domain.User {
	t.Helper()

	timestamp := time.Now().UnixNano()
	email := fmt.Sprintf("test-%d-%s@example.com", timestamp, uuid.New().String()[:8])
	name := fmt.Sprintf("Test User %d", timestamp)

	user, err := domain.NewUser(email, name, "password123")
	require.NoError(t, err, "should create domain user")

	return user
}

// createTestUserWithPrefix creates a test user with a specific email prefix.
func createTestUserWithPrefix(t *testing.T, prefix string) *domain.User {
	t.Helper()

	timestamp := time.Now().UnixNano()
	email := fmt.Sprintf("%s-%d-%s@example.com", prefix, timestamp, uuid.New().String()[:8])
	name := fmt.Sprintf("Test User %d", timestamp)

	user, err := domain.NewUser(email, name, "password123")
	require.NoError(t, err, "should create domain user")

	return user
}

// cleanupTestUser removes a specific test user by email.
func cleanupTestUser(t *testing.T, ctx context.Context, email string) {
	t.Helper()
	_, _ = testDB.Exec(ctx, "DELETE FROM users WHERE email = $1", email)
}

// cleanupTestUsersByPattern removes test users matching a pattern.
func cleanupTestUsersByPattern(t *testing.T, ctx context.Context, pattern string) {
	t.Helper()
	_, err := testDB.Exec(ctx, "DELETE FROM users WHERE email LIKE $1", pattern)
	if err != nil {
		t.Logf("Warning: failed to cleanup test users with pattern %s: %v", pattern, err)
	}
}

// TestSqlcUserRepository_Create tests user creation.
func TestSqlcUserRepository_Create(t *testing.T) {
	ctx := context.Background()
	querier := sqlc.New(testDB)
	repo := NewSqlcUserRepository(querier)

	t.Run("create_valid_user", func(t *testing.T) {
		user := createTestUser(t)

		// Cleanup after test
		t.Cleanup(func() {
			cleanupTestUser(t, ctx, user.Email)
		})

		err := repo.Create(ctx, user)
		assert.NoError(t, err)
		assert.NotEmpty(t, user.ID, "user ID should be set")
		assert.False(t, user.CreatedAt.IsZero(), "created_at should be set")
		assert.False(t, user.UpdatedAt.IsZero(), "updated_at should be set")
	})

	t.Run("duplicate_email", func(t *testing.T) {
		user := createTestUser(t)

		// Create first user
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		// Cleanup after test
		t.Cleanup(func() {
			cleanupTestUser(t, ctx, user.Email)
		})

		// Try to create another with same email
		user2, _ := domain.NewUser(user.Email, "Another Name", "password123")
		err = repo.Create(ctx, user2)
		assert.Error(t, err)
		assert.True(t, errors.IsConflictError(err), "expected conflict error for duplicate email")
	})
}

// TestSqlcUserRepository_GetByID tests retrieving users by ID.
func TestSqlcUserRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	querier := sqlc.New(testDB)
	repo := NewSqlcUserRepository(querier)

	// Create a test user first
	testUser := createTestUser(t)
	err := repo.Create(ctx, testUser)
	require.NoError(t, err)

	// Cleanup after test
	t.Cleanup(func() {
		cleanupTestUser(t, ctx, testUser.Email)
	})

	t.Run("get_existing_user", func(t *testing.T) {
		user, err := repo.GetByID(ctx, testUser.ID)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, testUser.ID, user.ID)
		assert.Equal(t, testUser.Email, user.Email)
		assert.Equal(t, testUser.Name, user.Name)
	})

	t.Run("get_non-existent_user", func(t *testing.T) {
		user, err := repo.GetByID(ctx, uuid.New().String())
		assert.Error(t, err)
		assert.True(t, errors.IsNotFoundError(err), "expected not found error")
		assert.Nil(t, user)
	})

	t.Run("invalid_uuid_format", func(t *testing.T) {
		user, err := repo.GetByID(ctx, "not-a-uuid")
		assert.Error(t, err)
		assert.True(t, errors.IsValidationError(err), "expected validation error")
		assert.Nil(t, user)
	})

	t.Run("empty_id", func(t *testing.T) {
		user, err := repo.GetByID(ctx, "")
		assert.Error(t, err)
		assert.True(t, errors.IsValidationError(err), "expected validation error")
		assert.Nil(t, user)
	})
}

// TestSqlcUserRepository_GetByEmail tests retrieving users by email.
func TestSqlcUserRepository_GetByEmail(t *testing.T) {
	ctx := context.Background()
	querier := sqlc.New(testDB)
	repo := NewSqlcUserRepository(querier)

	// Create a test user first
	testUser := createTestUser(t)
	err := repo.Create(ctx, testUser)
	require.NoError(t, err)

	// Cleanup after test
	t.Cleanup(func() {
		cleanupTestUser(t, ctx, testUser.Email)
	})

	t.Run("get_existing_user_by_email", func(t *testing.T) {
		user, err := repo.GetByEmail(ctx, testUser.Email)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, testUser.Email, user.Email)
	})

	t.Run("get_non-existent_email", func(t *testing.T) {
		user, err := repo.GetByEmail(ctx, "nonexistent@example.com")
		assert.Error(t, err)
		assert.True(t, errors.IsNotFoundError(err), "expected not found error")
		assert.Nil(t, user)
	})

	t.Run("empty_email", func(t *testing.T) {
		user, err := repo.GetByEmail(ctx, "")
		assert.Error(t, err)
		assert.True(t, errors.IsNotFoundError(err), "expected not found error for empty email")
		assert.Nil(t, user)
	})
}

// TestSqlcUserRepository_GetAll tests retrieving all users with pagination.
func TestSqlcUserRepository_GetAll(t *testing.T) {
	ctx := context.Background()
	querier := sqlc.New(testDB)
	repo := NewSqlcUserRepository(querier)

	// Create test users with unique pattern for this test
	numUsers := 3
	createdUsers := make([]*domain.User, 0, numUsers)

	for i := 0; i < numUsers; i++ {
		user := createTestUserWithPrefix(t, "integration-getall")
		err := repo.Create(ctx, user)
		require.NoError(t, err)
		createdUsers = append(createdUsers, user)
	}

	// Cleanup after test
	t.Cleanup(func() {
		cleanupTestUsersByPattern(t, ctx, "integration-getall-%@example.com")
	})

	t.Run("first_page", func(t *testing.T) {
		result, err := repo.GetAll(ctx, 0, 2)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 2, "should return at least 2 users")
	})

	t.Run("second_page", func(t *testing.T) {
		result, err := repo.GetAll(ctx, 1, 2)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 1, "should return at least 1 user")
	})

	t.Run("large_limit", func(t *testing.T) {
		result, err := repo.GetAll(ctx, 0, 100)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), numUsers, "should return at least the created users")
	})

	t.Run("negative_offset_defaults_to_0", func(t *testing.T) {
		result, err := repo.GetAll(ctx, -1, 10)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), numUsers, "should return at least the created users")
	})

	t.Run("zero_limit_defaults_to_10", func(t *testing.T) {
		result, err := repo.GetAll(ctx, 0, 0)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), numUsers, "should return at least the created users")
	})

	t.Run("limit_over_100_capped_at_100", func(t *testing.T) {
		result, err := repo.GetAll(ctx, 0, 200)
		assert.NoError(t, err)
		assert.LessOrEqual(t, len(result), 100, "should be capped at 100")
	})
}

// TestSqlcUserRepository_Count tests counting users.
func TestSqlcUserRepository_Count(t *testing.T) {
	ctx := context.Background()
	querier := sqlc.New(testDB)
	repo := NewSqlcUserRepository(querier)

	// Get initial count (before creating test users)
	initialCount, err := repo.Count(ctx)
	require.NoError(t, err)

	// Create test users with unique pattern for this test
	numUsers := 2
	createdEmails := make([]string, 0, numUsers)

	for i := 0; i < numUsers; i++ {
		user := createTestUserWithPrefix(t, fmt.Sprintf("integration-count-%d", time.Now().UnixNano()))
		err := repo.Create(ctx, user)
		require.NoError(t, err)
		createdEmails = append(createdEmails, user.Email)
	}

	// Cleanup after test
	t.Cleanup(func() {
		for _, email := range createdEmails {
			cleanupTestUser(t, ctx, email)
		}
	})

	// Get new count
	newCount, err := repo.Count(ctx)
	require.NoError(t, err)

	// Verify count increased by at least the number of users we created
	// (not exactly equal because other tests may have created users too)
	assert.GreaterOrEqual(t, newCount, initialCount+numUsers, "count should increase by at least the number of created users")
}

// TestSqlcUserRepository_Update tests updating users.
func TestSqlcUserRepository_Update(t *testing.T) {
	ctx := context.Background()
	querier := sqlc.New(testDB)
	repo := NewSqlcUserRepository(querier)

	// Create a test user first
	testUser := createTestUser(t)
	err := repo.Create(ctx, testUser)
	require.NoError(t, err)

	// Cleanup after test
	t.Cleanup(func() {
		cleanupTestUser(t, ctx, testUser.Email)
	})

	t.Run("update_name", func(t *testing.T) {
		originalUpdatedAt := testUser.UpdatedAt
		time.Sleep(10 * time.Millisecond) // Ensure time difference

		testUser.Name = "Updated Name"
		err := repo.Update(ctx, testUser)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Name", testUser.Name)
		assert.True(t, testUser.UpdatedAt.After(originalUpdatedAt), "updated_at should be newer")
	})

	t.Run("invalid_user_id", func(t *testing.T) {
		invalidUser := *testUser
		invalidUser.ID = "invalid-uuid"
		err := repo.Update(ctx, &invalidUser)
		assert.Error(t, err)
		assert.True(t, errors.IsValidationError(err), "expected validation error")
	})
}

// TestSqlcUserRepository_Delete tests deleting users.
func TestSqlcUserRepository_Delete(t *testing.T) {
	ctx := context.Background()
	querier := sqlc.New(testDB)
	repo := NewSqlcUserRepository(querier)

	t.Run("delete_existing_user", func(t *testing.T) {
		user := createTestUser(t)
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		// Cleanup just in case (though we're testing delete)
		t.Cleanup(func() {
			cleanupTestUser(t, ctx, user.Email)
		})

		err = repo.Delete(ctx, user.ID)
		assert.NoError(t, err)

		// Verify user is deleted
		_, err = repo.GetByID(ctx, user.ID)
		assert.True(t, errors.IsNotFoundError(err), "user should be deleted")
	})

	t.Run("delete_non-existent_user", func(t *testing.T) {
		err := repo.Delete(ctx, uuid.New().String())
		assert.Error(t, err)
		assert.True(t, errors.IsNotFoundError(err), "expected not found error")
	})

	t.Run("delete_with_invalid_uuid", func(t *testing.T) {
		err := repo.Delete(ctx, "not-a-uuid")
		assert.Error(t, err)
		assert.True(t, errors.IsValidationError(err), "expected validation error")
	})
}

// TestSqlcUserRepository_Exists tests checking user existence.
func TestSqlcUserRepository_Exists(t *testing.T) {
	ctx := context.Background()
	querier := sqlc.New(testDB)
	repo := NewSqlcUserRepository(querier)

	// Create a test user first
	testUser := createTestUser(t)
	err := repo.Create(ctx, testUser)
	require.NoError(t, err)

	// Cleanup after test
	t.Cleanup(func() {
		cleanupTestUser(t, ctx, testUser.Email)
	})

	t.Run("existing_user", func(t *testing.T) {
		exists, err := repo.Exists(ctx, testUser.Email)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("non-existent_user", func(t *testing.T) {
		exists, err := repo.Exists(ctx, "nonexistent@example.com")
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("empty_email", func(t *testing.T) {
		exists, err := repo.Exists(ctx, "")
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

// TestSqlcUserRepository_ConcurrentOperations tests concurrent database operations.
func TestSqlcUserRepository_ConcurrentOperations(t *testing.T) {
	ctx := context.Background()
	querier := sqlc.New(testDB)
	repo := NewSqlcUserRepository(querier)

	// Create test users with unique pattern
	numUsers := 5
	createdEmails := make([]string, 0, numUsers)
	errors := make(chan error, numUsers)

	for i := 0; i < numUsers; i++ {
		go func(index int) {
			timestamp := time.Now().UnixNano()
			email := fmt.Sprintf("concurrent-%d-%d@example.com", timestamp, index)
			user, _ := domain.NewUser(
				email,
				fmt.Sprintf("Concurrent User %d", index),
				"password123",
			)
			createdEmails = append(createdEmails, email)
			errors <- repo.Create(ctx, user)
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numUsers; i++ {
		err := <-errors
		assert.NoError(t, err, "concurrent create should succeed")
	}

	// Cleanup after test
	t.Cleanup(func() {
		cleanupTestUsersByPattern(t, ctx, "concurrent-%@example.com")
	})

	// Verify at least some users were created
	count, err := repo.Count(ctx)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, count, numUsers, "should have at least the created users")
}

// TestSqlcUserRepository_CRUDWorkflow tests a complete CRUD workflow.
func TestSqlcUserRepository_CRUDWorkflow(t *testing.T) {
	ctx := context.Background()
	querier := sqlc.New(testDB)
	repo := NewSqlcUserRepository(querier)

	// 1. Create
	user := createTestUser(t)
	err := repo.Create(ctx, user)
	require.NoError(t, err)
	assert.NotEmpty(t, user.ID)

	// Cleanup after test
	t.Cleanup(func() {
		cleanupTestUser(t, ctx, user.Email)
	})

	// 2. Read by ID
	retrievedByID, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.Email, retrievedByID.Email)
	assert.Equal(t, user.Name, retrievedByID.Name)

	// 3. Read by Email
	retrievedByEmail, err := repo.GetByEmail(ctx, user.Email)
	require.NoError(t, err)
	assert.Equal(t, user.ID, retrievedByEmail.ID)

	// 4. Check exists
	exists, err := repo.Exists(ctx, user.Email)
	require.NoError(t, err)
	assert.True(t, exists)

	// 5. Update
	originalName := user.Name
	user.Name = "Updated Name"
	err = repo.Update(ctx, user)
	require.NoError(t, err)

	// Verify update
	updated, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.NotEqual(t, originalName, updated.Name)

	// 6. Delete
	err = repo.Delete(ctx, user.ID)
	require.NoError(t, err)

	// 7. Verify deletion
	_, err = repo.GetByID(ctx, user.ID)
	assert.True(t, errors.IsNotFoundError(err), "should return not found after deletion")

	exists, err = repo.Exists(ctx, user.Email)
	require.NoError(t, err)
	assert.False(t, exists, "should not exist after deletion")
}
