//go:build integration

// Package handler provides HTTP handlers for the user feature.
// This file contains integration tests for the user HTTP handlers.
package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"zercle-go-template/internal/config"
	authusecase "zercle-go-template/internal/feature/auth/usecase"
	"zercle-go-template/internal/feature/user/dto"
	"zercle-go-template/internal/feature/user/repository"
	userusecase "zercle-go-template/internal/feature/user/usecase"
	"zercle-go-template/internal/infrastructure/db/sqlc"
	"zercle-go-template/internal/logger"
)

// testDB holds the database connection for integration tests.
var testDB *pgxpool.Pool

// testDBMutex ensures tests don't interfere with each other's database operations
var testDBMutex sync.Mutex

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

	// Final cleanup after all tests complete - truncate table for complete isolation
	ctx := context.Background()
	_, _ = testDB.Exec(ctx, "TRUNCATE TABLE users RESTART IDENTITY CASCADE")

	if testDB != nil {
		testDB.Close()
	}

	os.Exit(code)
}

// testJWTConfig returns a JWT config for testing.
func testJWTConfig() *config.JWTConfig {
	return &config.JWTConfig{
		Secret:          "test-secret-key-for-integration-tests-only",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
	}
}

// setupTestHandler creates a handler with real database for integration tests.
func setupTestHandler(t *testing.T) (*UserHandler, *echo.Echo) {
	t.Helper()

	log := logger.NewNop()
	querier := sqlc.New(testDB)

	// Create repository and usecase with real database
	userRepo := repository.NewSqlcUserRepository(querier)
	userUc := userusecase.NewUserUsecase(userRepo, log)

	// Create JWT usecase with test secret
	jwtConfig := testJWTConfig()
	jwtUc := authusecase.NewJWTUsecase(jwtConfig, log)

	handler := NewUserHandler(userUc, jwtUc, log)

	e := echo.New()

	return handler, e
}

// cleanupTestUser removes a specific test user by email.
func cleanupTestUser(t *testing.T, ctx context.Context, email string) {
	t.Helper()
	testDBMutex.Lock()
	defer testDBMutex.Unlock()
	_, _ = testDB.Exec(ctx, "DELETE FROM users WHERE email = $1", email)
}

// cleanupTestUserByID removes a specific test user by ID.
func cleanupTestUserByID(t *testing.T, ctx context.Context, userID string) {
	t.Helper()
	testDBMutex.Lock()
	defer testDBMutex.Unlock()
	_, _ = testDB.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
}

// createUniqueEmail generates a unique email for testing using UUID to ensure isolation.
func createUniqueEmail(prefix string) string {
	return fmt.Sprintf("%s-%s@example.com", prefix, uuid.New().String())
}

// TestUserHandler_CreateUser tests the CreateUser endpoint.
func TestUserHandler_CreateUser(t *testing.T) {
	ctx := context.Background()
	handler, e := setupTestHandler(t)

	tests := []struct {
		name           string
		requestBody    dto.CreateUserRequest
		expectedStatus int
		checkResponse  func(t *testing.T, resp *Response)
	}{
		{
			name: "create_valid_user",
			requestBody: dto.CreateUserRequest{
				Email:    createUniqueEmail("test"),
				Name:     "Test User",
				Password: "password123",
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.True(t, resp.Success)
				assert.NotNil(t, resp.Data)
				data, ok := resp.Data.(map[string]interface{})
				require.True(t, ok)
				assert.NotEmpty(t, data["id"])
				assert.Equal(t, "Test User", data["name"])
			},
		},
		{
			name: "invalid_email",
			requestBody: dto.CreateUserRequest{
				Email:    "invalid-email",
				Name:     "Test User",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.False(t, resp.Success)
				assert.NotNil(t, resp.Error)
			},
		},
		{
			name: "short_password",
			requestBody: dto.CreateUserRequest{
				Email:    createUniqueEmail("test"),
				Name:     "Test User",
				Password: "short",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.False(t, resp.Success)
				assert.NotNil(t, resp.Error)
			},
		},
		{
			name: "short_name",
			requestBody: dto.CreateUserRequest{
				Email:    createUniqueEmail("test"),
				Name:     "A",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.False(t, resp.Success)
				assert.NotNil(t, resp.Error)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.CreateUser(c)
			if err != nil {
				_ = err // echo handles the error
			}

			assert.Equal(t, tt.expectedStatus, rec.Code)

			var resp Response
			err = json.Unmarshal(rec.Body.Bytes(), &resp)
			require.NoError(t, err)
			tt.checkResponse(t, &resp)

			// Cleanup: remove created user
			if tt.expectedStatus == http.StatusCreated {
				cleanupTestUser(t, ctx, tt.requestBody.Email)
			}
		})
	}
}

// TestUserHandler_CreateUser_DuplicateEmail tests duplicate email handling.
func TestUserHandler_CreateUser_DuplicateEmail(t *testing.T) {
	ctx := context.Background()
	handler, e := setupTestHandler(t)

	email := createUniqueEmail("test")

	// Create first user
	body, _ := json.Marshal(dto.CreateUserRequest{
		Email:    email,
		Name:     "First User",
		Password: "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.CreateUser(c)
	if err != nil {
		_ = err
	}

	require.Equal(t, http.StatusCreated, rec.Code)

	// Cleanup after test
	t.Cleanup(func() {
		cleanupTestUser(t, ctx, email)
	})

	// Try to create second user with same email
	body2, _ := json.Marshal(dto.CreateUserRequest{
		Email:    email,
		Name:     "Second User",
		Password: "password456",
	})
	req2 := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body2))
	req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err = handler.CreateUser(c2)
	if err != nil {
		_ = err
	}

	assert.Equal(t, http.StatusConflict, rec2.Code)

	var resp Response
	err = json.Unmarshal(rec2.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.NotNil(t, resp.Error)
}

// TestUserHandler_GetUser tests the GetUser endpoint.
func TestUserHandler_GetUser(t *testing.T) {
	ctx := context.Background()
	handler, e := setupTestHandler(t)

	// First, create a user
	email := createUniqueEmail("test")
	createBody, _ := json.Marshal(dto.CreateUserRequest{
		Email:    email,
		Name:     "Test User",
		Password: "password123",
	})
	createReq := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(createBody))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createRec := httptest.NewRecorder()
	createCtx := e.NewContext(createReq, createRec)

	err := handler.CreateUser(createCtx)
	if err != nil {
		_ = err
	}
	require.Equal(t, http.StatusCreated, createRec.Code)

	var createResp Response
	err = json.Unmarshal(createRec.Body.Bytes(), &createResp)
	require.NoError(t, err)
	userData := createResp.Data.(map[string]interface{})
	userID := userData["id"].(string)

	// Cleanup after test
	t.Cleanup(func() {
		cleanupTestUser(t, ctx, email)
	})

	tests := []struct {
		name           string
		id             string
		expectedStatus int
		checkResponse  func(t *testing.T, resp *Response)
	}{
		{
			name:           "get_existing_user",
			id:             userID,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.True(t, resp.Success)
				data, ok := resp.Data.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, userID, data["id"])
				assert.Equal(t, email, data["email"])
			},
		},
		{
			name:           "get_non-existent_user",
			id:             uuid.New().String(),
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.False(t, resp.Success)
				assert.NotNil(t, resp.Error)
			},
		},
		{
			name:           "invalid_user_id",
			id:             "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.False(t, resp.Success)
				assert.NotNil(t, resp.Error)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.id, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			err := handler.GetUser(c)
			if err != nil {
				_ = err
			}

			assert.Equal(t, tt.expectedStatus, rec.Code)

			var resp Response
			err = json.Unmarshal(rec.Body.Bytes(), &resp)
			require.NoError(t, err)
			tt.checkResponse(t, &resp)
		})
	}
}

// TestUserHandler_ListUsers tests the ListUsers endpoint.
func TestUserHandler_ListUsers(t *testing.T) {
	ctx := context.Background()
	handler, e := setupTestHandler(t)

	// Create test users with unique emails for this test only
	numUsers := 3
	createdUserIDs := make([]string, 0, numUsers)
	createdEmails := make([]string, 0, numUsers)

	for i := 0; i < numUsers; i++ {
		email := createUniqueEmail("list-test")
		body, _ := json.Marshal(dto.CreateUserRequest{
			Email:    email,
			Name:     fmt.Sprintf("List Test User %d", i),
			Password: "password123",
		})
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.CreateUser(c)
		if err != nil {
			_ = err
		}
		require.Equal(t, http.StatusCreated, rec.Code)

		var resp Response
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		userData := resp.Data.(map[string]interface{})
		createdUserIDs = append(createdUserIDs, userData["id"].(string))
		createdEmails = append(createdEmails, email)
	}

	// Cleanup after test - delete specific users by ID
	t.Cleanup(func() {
		for _, userID := range createdUserIDs {
			cleanupTestUserByID(t, ctx, userID)
		}
	})

	tests := []struct {
		name           string
		query          string
		expectedStatus int
		checkResponse  func(t *testing.T, resp *Response)
	}{
		{
			name:           "list_users_default_pagination",
			query:          "",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.True(t, resp.Success)
				data, ok := resp.Data.(map[string]interface{})
				require.True(t, ok)
				users := data["users"].([]interface{})
				assert.GreaterOrEqual(t, len(users), numUsers)
				assert.NotNil(t, resp.Meta)
			},
		},
		{
			name:           "list_users_with_pagination",
			query:          "?page=1&limit=2",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.True(t, resp.Success)
				data, ok := resp.Data.(map[string]interface{})
				require.True(t, ok)
				users := data["users"].([]interface{})
				assert.LessOrEqual(t, len(users), 2)
			},
		},
		{
			name:           "list_users_with_invalid_page",
			query:          "?page=invalid&limit=10",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.True(t, resp.Success)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/users"+tt.query, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.ListUsers(c)
			if err != nil {
				_ = err
			}

			assert.Equal(t, tt.expectedStatus, rec.Code)

			var resp Response
			err = json.Unmarshal(rec.Body.Bytes(), &resp)
			require.NoError(t, err)
			tt.checkResponse(t, &resp)
		})
	}
}

// TestUserHandler_UpdateUser tests the UpdateUser endpoint.
func TestUserHandler_UpdateUser(t *testing.T) {
	ctx := context.Background()
	handler, e := setupTestHandler(t)

	// Create a user to update
	email := createUniqueEmail("test")
	createBody, _ := json.Marshal(dto.CreateUserRequest{
		Email:    email,
		Name:     "Original Name",
		Password: "password123",
	})
	createReq := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(createBody))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createRec := httptest.NewRecorder()
	createCtx := e.NewContext(createReq, createRec)

	err := handler.CreateUser(createCtx)
	if err != nil {
		_ = err
	}
	require.Equal(t, http.StatusCreated, createRec.Code)

	var createResp Response
	err = json.Unmarshal(createRec.Body.Bytes(), &createResp)
	require.NoError(t, err)
	userData := createResp.Data.(map[string]interface{})
	userID := userData["id"].(string)

	// Cleanup after test
	t.Cleanup(func() {
		cleanupTestUser(t, ctx, email)
	})

	tests := []struct {
		name           string
		id             string
		requestBody    dto.UpdateUserRequest
		expectedStatus int
		checkResponse  func(t *testing.T, resp *Response)
	}{
		{
			name: "update_user_name",
			id:   userID,
			requestBody: dto.UpdateUserRequest{
				Name: "Updated Name",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.True(t, resp.Success)
				data, ok := resp.Data.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "Updated Name", data["name"])
			},
		},
		{
			name:           "update_non-existent_user",
			id:             uuid.New().String(),
			requestBody:    dto.UpdateUserRequest{Name: "New Name"},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.False(t, resp.Success)
				assert.NotNil(t, resp.Error)
			},
		},
		{
			name:           "update_with_invalid_id",
			id:             "invalid-uuid",
			requestBody:    dto.UpdateUserRequest{Name: "New Name"},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.False(t, resp.Success)
				assert.NotNil(t, resp.Error)
			},
		},
		{
			name:           "update_with_short_name",
			id:             userID,
			requestBody:    dto.UpdateUserRequest{Name: "A"},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.False(t, resp.Success)
				assert.NotNil(t, resp.Error)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/users/"+tt.id, bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			err := handler.UpdateUser(c)
			if err != nil {
				_ = err
			}

			assert.Equal(t, tt.expectedStatus, rec.Code)

			var resp Response
			err = json.Unmarshal(rec.Body.Bytes(), &resp)
			require.NoError(t, err)
			tt.checkResponse(t, &resp)
		})
	}
}

// TestUserHandler_DeleteUser tests the DeleteUser endpoint.
func TestUserHandler_DeleteUser(t *testing.T) {
	ctx := context.Background()
	handler, e := setupTestHandler(t)

	// Create a user to delete
	email := createUniqueEmail("test")
	createBody, _ := json.Marshal(dto.CreateUserRequest{
		Email:    email,
		Name:     "User to Delete",
		Password: "password123",
	})
	createReq := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(createBody))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createRec := httptest.NewRecorder()
	createCtx := e.NewContext(createReq, createRec)

	err := handler.CreateUser(createCtx)
	if err != nil {
		_ = err
	}
	require.Equal(t, http.StatusCreated, createRec.Code)

	var createResp Response
	err = json.Unmarshal(createRec.Body.Bytes(), &createResp)
	require.NoError(t, err)
	userData := createResp.Data.(map[string]interface{})
	userID := userData["id"].(string)

	// Cleanup after test (in case delete fails)
	t.Cleanup(func() {
		cleanupTestUser(t, ctx, email)
	})

	tests := []struct {
		name           string
		id             string
		expectedStatus int
	}{
		{
			name:           "delete_existing_user",
			id:             userID,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "delete_already_deleted_user",
			id:             userID,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "delete_non-existent_user",
			id:             uuid.New().String(),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "delete_with_invalid_id",
			id:             "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/users/"+tt.id, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			err := handler.DeleteUser(c)
			if err != nil {
				_ = err
			}

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// TestUserHandler_Login tests the Login endpoint.
func TestUserHandler_Login(t *testing.T) {
	ctx := context.Background()
	handler, e := setupTestHandler(t)

	// Create a user to login
	email := createUniqueEmail("test")
	password := "password123"
	createBody, _ := json.Marshal(dto.CreateUserRequest{
		Email:    email,
		Name:     "Login Test User",
		Password: password,
	})
	createReq := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(createBody))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createRec := httptest.NewRecorder()
	createCtx := e.NewContext(createReq, createRec)

	err := handler.CreateUser(createCtx)
	if err != nil {
		_ = err
	}
	require.Equal(t, http.StatusCreated, createRec.Code)

	// Cleanup after test
	t.Cleanup(func() {
		cleanupTestUser(t, ctx, email)
	})

	tests := []struct {
		name           string
		requestBody    dto.UserLoginRequest
		expectedStatus int
		checkResponse  func(t *testing.T, resp *Response)
	}{
		{
			name: "successful_login",
			requestBody: dto.UserLoginRequest{
				Email:    email,
				Password: password,
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.True(t, resp.Success)
				data, ok := resp.Data.(map[string]interface{})
				require.True(t, ok)
				assert.NotEmpty(t, data["token"])
				userData, ok := data["user"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, email, userData["email"])
			},
		},
		{
			name: "wrong_password",
			requestBody: dto.UserLoginRequest{
				Email:    email,
				Password: "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.False(t, resp.Success)
				assert.NotNil(t, resp.Error)
			},
		},
		{
			name: "non-existent_user",
			requestBody: dto.UserLoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.False(t, resp.Success)
				assert.NotNil(t, resp.Error)
			},
		},
		{
			name: "invalid_email_format",
			requestBody: dto.UserLoginRequest{
				Email:    "invalid-email",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.False(t, resp.Success)
				assert.NotNil(t, resp.Error)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.Login(c)
			if err != nil {
				_ = err
			}

			assert.Equal(t, tt.expectedStatus, rec.Code)

			var resp Response
			err = json.Unmarshal(rec.Body.Bytes(), &resp)
			require.NoError(t, err)
			tt.checkResponse(t, &resp)
		})
	}
}

// TestUserHandler_UpdatePassword tests the UpdatePassword endpoint.
func TestUserHandler_UpdatePassword(t *testing.T) {
	ctx := context.Background()
	handler, e := setupTestHandler(t)

	// Create a user
	email := createUniqueEmail("test")
	oldPassword := "password123"
	createBody, _ := json.Marshal(dto.CreateUserRequest{
		Email:    email,
		Name:     "Password Test User",
		Password: oldPassword,
	})
	createReq := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(createBody))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createRec := httptest.NewRecorder()
	createCtx := e.NewContext(createReq, createRec)

	err := handler.CreateUser(createCtx)
	if err != nil {
		_ = err
	}
	require.Equal(t, http.StatusCreated, createRec.Code)

	var createResp Response
	err = json.Unmarshal(createRec.Body.Bytes(), &createResp)
	require.NoError(t, err)
	userData := createResp.Data.(map[string]interface{})
	userID := userData["id"].(string)

	// Cleanup after test
	t.Cleanup(func() {
		cleanupTestUser(t, ctx, email)
	})

	tests := []struct {
		name           string
		id             string
		requestBody    dto.UpdatePasswordRequest
		expectedStatus int
	}{
		{
			name: "update_password_successfully",
			id:   userID,
			requestBody: dto.UpdatePasswordRequest{
				OldPassword: oldPassword,
				NewPassword: "newpassword123",
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "update_password_with_wrong_old_password",
			id:   userID,
			requestBody: dto.UpdatePasswordRequest{
				OldPassword: "wrongpassword",
				NewPassword: "newpassword123",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "update_password_for_non-existent_user",
			id:             uuid.New().String(),
			requestBody:    dto.UpdatePasswordRequest{OldPassword: oldPassword, NewPassword: "newpassword123"},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "update_password_with_short_new_password",
			id:   userID,
			requestBody: dto.UpdatePasswordRequest{
				OldPassword: oldPassword,
				NewPassword: "short",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/users/"+tt.id+"/password", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			err := handler.UpdatePassword(c)
			if err != nil {
				_ = err
			}

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// TestUserHandler_RegisterRoutes tests route registration.
func TestUserHandler_RegisterRoutes(t *testing.T) {
	handler, e := setupTestHandler(t)

	// Create router groups
	api := e.Group("/api")

	// Register routes
	handler.RegisterRoutes(api)

	// Test that routes are registered by checking the echo router
	routes := e.Routes()
	require.Greater(t, len(routes), 0)

	// Check for expected routes
	expectedRoutes := []string{
		"POST /api/auth/login",
		"POST /api/users",
		"GET /api/users",
		"GET /api/users/:id",
		"PUT /api/users/:id",
		"DELETE /api/users/:id",
		"PUT /api/users/:id/password",
	}

	routeMap := make(map[string]bool)
	for _, route := range routes {
		routeMap[route.Method+" "+route.Path] = true
	}

	for _, expected := range expectedRoutes {
		assert.True(t, routeMap[expected], "Route %s should be registered", expected)
	}
}

// TestUserHandler_FullWorkflow tests a complete user workflow.
func TestUserHandler_FullWorkflow(t *testing.T) {
	ctx := context.Background()
	handler, e := setupTestHandler(t)

	email := createUniqueEmail("workflow")
	password := "password123"
	name := "Workflow Test User"

	// Cleanup after test
	t.Cleanup(func() {
		cleanupTestUser(t, ctx, email)
	})

	// 1. Create User
	createBody, _ := json.Marshal(dto.CreateUserRequest{
		Email:    email,
		Name:     name,
		Password: password,
	})
	createReq := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(createBody))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createRec := httptest.NewRecorder()
	createCtx := e.NewContext(createReq, createRec)

	err := handler.CreateUser(createCtx)
	if err != nil {
		_ = err
	}
	require.Equal(t, http.StatusCreated, createRec.Code)

	var createResp Response
	err = json.Unmarshal(createRec.Body.Bytes(), &createResp)
	require.NoError(t, err)
	require.True(t, createResp.Success)
	userData := createResp.Data.(map[string]interface{})
	userID := userData["id"].(string)
	assert.Equal(t, name, userData["name"])
	assert.Equal(t, email, userData["email"])

	// 2. Login
	loginBody, _ := json.Marshal(dto.UserLoginRequest{
		Email:    email,
		Password: password,
	})
	loginReq := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	loginRec := httptest.NewRecorder()
	loginCtx := e.NewContext(loginReq, loginRec)

	err = handler.Login(loginCtx)
	if err != nil {
		_ = err
	}
	require.Equal(t, http.StatusOK, loginRec.Code)

	var loginResp Response
	err = json.Unmarshal(loginRec.Body.Bytes(), &loginResp)
	require.NoError(t, err)
	require.True(t, loginResp.Success)
	loginData := loginResp.Data.(map[string]interface{})
	assert.NotEmpty(t, loginData["token"])

	// 3. Get User
	getReq := httptest.NewRequest(http.MethodGet, "/users/"+userID, nil)
	getRec := httptest.NewRecorder()
	getCtx := e.NewContext(getReq, getRec)
	getCtx.SetParamNames("id")
	getCtx.SetParamValues(userID)

	err = handler.GetUser(getCtx)
	if err != nil {
		_ = err
	}
	require.Equal(t, http.StatusOK, getRec.Code)

	var getResp Response
	err = json.Unmarshal(getRec.Body.Bytes(), &getResp)
	require.NoError(t, err)
	require.True(t, getResp.Success)

	// 4. Update User
	updateBody, _ := json.Marshal(dto.UpdateUserRequest{
		Name: "Updated Name",
	})
	updateReq := httptest.NewRequest(http.MethodPut, "/users/"+userID, bytes.NewReader(updateBody))
	updateReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	updateRec := httptest.NewRecorder()
	updateCtx := e.NewContext(updateReq, updateRec)
	updateCtx.SetParamNames("id")
	updateCtx.SetParamValues(userID)

	err = handler.UpdateUser(updateCtx)
	if err != nil {
		_ = err
	}
	require.Equal(t, http.StatusOK, updateRec.Code)

	var updateResp Response
	err = json.Unmarshal(updateRec.Body.Bytes(), &updateResp)
	require.NoError(t, err)
	require.True(t, updateResp.Success)
	updateData := updateResp.Data.(map[string]interface{})
	assert.Equal(t, "Updated Name", updateData["name"])

	// 5. Update Password
	newPassword := "newpassword123"
	passwordBody, _ := json.Marshal(dto.UpdatePasswordRequest{
		OldPassword: password,
		NewPassword: newPassword,
	})
	passwordReq := httptest.NewRequest(http.MethodPut, "/users/"+userID+"/password", bytes.NewReader(passwordBody))
	passwordReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	passwordRec := httptest.NewRecorder()
	passwordCtx := e.NewContext(passwordReq, passwordRec)
	passwordCtx.SetParamNames("id")
	passwordCtx.SetParamValues(userID)

	err = handler.UpdatePassword(passwordCtx)
	if err != nil {
		_ = err
	}
	require.Equal(t, http.StatusNoContent, passwordRec.Code)

	// 6. Login with new password
	loginBody2, _ := json.Marshal(dto.UserLoginRequest{
		Email:    email,
		Password: newPassword,
	})
	loginReq2 := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(loginBody2))
	loginReq2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	loginRec2 := httptest.NewRecorder()
	loginCtx2 := e.NewContext(loginReq2, loginRec2)

	err = handler.Login(loginCtx2)
	if err != nil {
		_ = err
	}
	require.Equal(t, http.StatusOK, loginRec2.Code)

	// 7. Delete User
	deleteReq := httptest.NewRequest(http.MethodDelete, "/users/"+userID, nil)
	deleteRec := httptest.NewRecorder()
	deleteCtx := e.NewContext(deleteReq, deleteRec)
	deleteCtx.SetParamNames("id")
	deleteCtx.SetParamValues(userID)

	err = handler.DeleteUser(deleteCtx)
	if err != nil {
		_ = err
	}
	require.Equal(t, http.StatusNoContent, deleteRec.Code)

	// 8. Verify deletion
	getReq2 := httptest.NewRequest(http.MethodGet, "/users/"+userID, nil)
	getRec2 := httptest.NewRecorder()
	getCtx2 := e.NewContext(getReq2, getRec2)
	getCtx2.SetParamNames("id")
	getCtx2.SetParamValues(userID)

	err = handler.GetUser(getCtx2)
	if err != nil {
		_ = err
	}
	require.Equal(t, http.StatusNotFound, getRec2.Code)
}

// intPtr returns a pointer to an int value.
func intPtr(i int) *int {
	return &i
}

// float64Ptr returns a pointer to a float64 value.
func float64Ptr(f float64) *float64 {
	return &f
}

// jsonNumberToInt converts json.Number to int.
func jsonNumberToInt(n json.Number) int {
	i, _ := strconv.Atoi(string(n))
	return i
}

// parseJWT parses a JWT token string (without verification for testing).
func parseJWT(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret-key-for-integration-tests-only"), nil
	})
}
