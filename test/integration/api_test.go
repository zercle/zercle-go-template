package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zercle/zercle-go-template/internal/app"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/logger"
)

type testApp struct {
	app    *app.App
	echo   *echo.Echo
	helper *TestDBHelper
	ctx    context.Context
	cancel context.CancelFunc
}

func setupTestApp(t *testing.T) *testApp {
	helper := NewTestDBHelper()
	helper.Setup(t)

	ctx, cancel := context.WithCancel(context.Background())

	testConfig := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 3001,
		},
		Database: config.DatabaseConfig{
			Host:              "localhost",
			Port:              5432,
			User:              "postgres",
			Password:          "postgres",
			DBName:            "zercle_test_db",
			Driver:            "postgres",
			MaxConns:          25,
			MinConns:          5,
			MaxConnLifetime:   1 * time.Hour,
			MaxConnIdleTime:   10 * time.Minute,
			HealthCheckPeriod: 1 * time.Minute,
		},
		JWT: config.JWTConfig{
			Secret:     "test-secret-key-for-testing",
			Expiration: 3600,
		},
		Argon2id: config.Argon2idConfig{
			Memory:      19456,
			Iterations:  2,
			Parallelism: 1,
			SaltLength:  16,
			KeyLength:   32,
		},
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"*"},
		},
		RateLimit: config.RateLimitConfig{
			Requests: 100,
			Window:   60,
		},
	}

	connStr, _ := helper.container.ConnectionString(context.Background(), "sslmode=disable")
	testConfig.Database.Host, testConfig.Database.Port = parseConnectionString(connStr)

	logCfg := &config.LoggingConfig{
		Level:  "debug",
		Format: "console",
	}
	log := logger.NewLogger(logCfg)
	application, err := app.NewApp(testConfig, log)
	require.NoError(t, err)

	return &testApp{
		app:    application,
		echo:   application.GetEcho(),
		helper: helper,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (ta *testApp) cleanup(t *testing.T) {
	ta.helper.Cleanup(t)
	if ta.app != nil {
		ta.app.Close()
	}
	ta.cancel()
}

func (ta *testApp) request(t *testing.T, method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	if token != "" {
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	ta.echo.ServeHTTP(rec, req)

	return rec
}

func TestHealthEndpoints(t *testing.T) {
	ta := setupTestApp(t)
	defer ta.cleanup(t)

	tests := []struct {
		name                string
		path                string
		expectedHTTPStatus  int
		expectedJSendStatus string
	}{
		{
			name:                "health check",
			path:                "/health",
			expectedHTTPStatus:  http.StatusOK,
			expectedJSendStatus: "success",
		},
		{
			name:                "readiness check",
			path:                "/readiness",
			expectedHTTPStatus:  http.StatusOK,
			expectedJSendStatus: "success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := ta.request(t, http.MethodGet, tt.path, nil, "")

			assert.Equal(t, tt.expectedHTTPStatus, rec.Code)

			var response map[string]interface{}
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedJSendStatus, response["status"])
		})
	}
}

func TestUserAuthEndpoints(t *testing.T) {
	ta := setupTestApp(t)
	defer ta.cleanup(t)

	t.Run("register new user", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"email":     "test@example.com",
			"password":  "password123",
			"full_name": "Test User",
		}

		rec := ta.request(t, http.MethodPost, "/api/v1/auth/register", reqBody, "")

		assert.Equal(t, http.StatusCreated, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "success", response["status"])

		data := response["data"].(map[string]interface{})
		assert.NotEmpty(t, data["token"])
		user := data["user"].(map[string]interface{})
		assert.NotEmpty(t, user["id"])
		assert.Equal(t, "test@example.com", user["email"])
		assert.Equal(t, "Test User", user["full_name"])
	})

	t.Run("register with duplicate email", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"email":     "test@example.com",
			"password":  "password123",
			"full_name": "Test User",
		}

		ta.request(t, http.MethodPost, "/api/v1/auth/register", reqBody, "")
		rec := ta.request(t, http.MethodPost, "/api/v1/auth/register", reqBody, "")

		assert.Equal(t, http.StatusConflict, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "error", response["status"])
	})

	t.Run("register with invalid email", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"email":     "invalid-email",
			"password":  "password123",
			"full_name": "Test User",
		}

		rec := ta.request(t, http.MethodPost, "/api/v1/auth/register", reqBody, "")

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "fail", response["status"])
	})

	t.Run("register with short password", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"email":     "test2@example.com",
			"password":  "123",
			"full_name": "Test User",
		}

		rec := ta.request(t, http.MethodPost, "/api/v1/auth/register", reqBody, "")

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("login with valid credentials", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"email":    "test@example.com",
			"password": "password123",
		}

		rec := ta.request(t, http.MethodPost, "/api/v1/auth/login", reqBody, "")

		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "success", response["status"])

		data := response["data"].(map[string]interface{})
		assert.NotEmpty(t, data["token"])
		user := data["user"].(map[string]interface{})
		assert.Equal(t, "test@example.com", user["email"])
	})

	t.Run("login with invalid credentials", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"email":    "test@example.com",
			"password": "wrongpassword",
		}

		rec := ta.request(t, http.MethodPost, "/api/v1/auth/login", reqBody, "")

		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "error", response["status"])
	})

	t.Run("login with non-existent user", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"email":    "nonexistent@example.com",
			"password": "password123",
		}

		rec := ta.request(t, http.MethodPost, "/api/v1/auth/login", reqBody, "")

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestUserProfileEndpoints(t *testing.T) {
	ta := setupTestApp(t)
	defer ta.cleanup(t)

	registerReq := map[string]interface{}{
		"email":     "profile@example.com",
		"password":  "password123",
		"full_name": "Profile User",
	}

	ta.request(t, http.MethodPost, "/api/v1/auth/register", registerReq, "")

	loginReq := map[string]interface{}{
		"email":    "profile@example.com",
		"password": "password123",
	}

	loginRec := ta.request(t, http.MethodPost, "/api/v1/auth/login", loginReq, "")

	var loginResp map[string]interface{}
	err := json.Unmarshal(loginRec.Body.Bytes(), &loginResp)
	require.NoError(t, err)
	token := loginResp["data"].(map[string]interface{})["token"].(string)

	t.Run("get user profile", func(t *testing.T) {
		rec := ta.request(t, http.MethodGet, "/api/v1/users/profile", nil, token)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "success", response["status"])

		data := response["data"].(map[string]interface{})
		assert.Equal(t, "profile@example.com", data["email"])
		assert.Equal(t, "Profile User", data["full_name"])
	})

	t.Run("get profile without token", func(t *testing.T) {
		rec := ta.request(t, http.MethodGet, "/api/v1/users/profile", nil, "")

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("update user profile", func(t *testing.T) {
		updateReq := map[string]interface{}{
			"full_name": "Updated Name",
			"phone":     "1234567890",
		}

		rec := ta.request(t, http.MethodPut, "/api/v1/users/profile", updateReq, token)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "success", response["status"])

		data := response["data"].(map[string]interface{})
		assert.Equal(t, "Updated Name", data["full_name"])
		assert.Equal(t, "1234567890", data["phone"])
	})

	t.Run("update profile with invalid name", func(t *testing.T) {
		updateReq := map[string]interface{}{
			"full_name": "X",
		}

		rec := ta.request(t, http.MethodPut, "/api/v1/users/profile", updateReq, token)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "fail", response["status"])
	})

	t.Run("delete user account", func(t *testing.T) {
		deleteRec := ta.request(t, http.MethodDelete, "/api/v1/users/profile", nil, token)

		assert.Equal(t, http.StatusNoContent, deleteRec.Code)

		getRec := ta.request(t, http.MethodGet, "/api/v1/users/profile", nil, token)
		assert.Equal(t, http.StatusNotFound, getRec.Code)
	})
}

func TestListUsersEndpoint(t *testing.T) {
	ta := setupTestApp(t)
	defer ta.cleanup(t)

	for i := 1; i <= 5; i++ {
		reqBody := map[string]interface{}{
			"email":     fmt.Sprintf("user%d@example.com", i),
			"password":  "password123",
			"full_name": fmt.Sprintf("User %d", i),
		}
		ta.request(t, http.MethodPost, "/api/v1/auth/register", reqBody, "")
	}

	loginReq := map[string]interface{}{
		"email":    "user1@example.com",
		"password": "password123",
	}

	loginRec := ta.request(t, http.MethodPost, "/api/v1/auth/login", loginReq, "")

	var loginResp map[string]interface{}
	err := json.Unmarshal(loginRec.Body.Bytes(), &loginResp)
	require.NoError(t, err)
	token := loginResp["data"].(map[string]interface{})["token"].(string)

	t.Run("list users with default pagination", func(t *testing.T) {
		rec := ta.request(t, http.MethodGet, "/api/v1/users", nil, token)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "success", response["status"])

		data := response["data"].(map[string]interface{})
		users := data["users"].([]interface{})
		total := int(data["total"].(float64))

		assert.GreaterOrEqual(t, total, 5)
		assert.NotEmpty(t, users)
	})

	t.Run("list users with custom pagination", func(t *testing.T) {
		rec := ta.request(t, http.MethodGet, "/api/v1/users?limit=2&offset=0", nil, token)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		users := data["users"].([]interface{})
		assert.Len(t, users, 2)
	})

	t.Run("list users without token", func(t *testing.T) {
		rec := ta.request(t, http.MethodGet, "/api/v1/users", nil, "")

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestTaskEndpoints(t *testing.T) {
	ta := setupTestApp(t)
	defer ta.cleanup(t)

	registerReq := map[string]interface{}{
		"email":     "taskuser@example.com",
		"password":  "password123",
		"full_name": "Task User",
	}

	ta.request(t, http.MethodPost, "/api/v1/auth/register", registerReq, "")

	loginReq := map[string]interface{}{
		"email":    "taskuser@example.com",
		"password": "password123",
	}

	loginRec := ta.request(t, http.MethodPost, "/api/v1/auth/login", loginReq, "")

	var loginResp map[string]interface{}
	err := json.Unmarshal(loginRec.Body.Bytes(), &loginResp)
	require.NoError(t, err)
	token := loginResp["data"].(map[string]interface{})["token"].(string)

	t.Run("create task", func(t *testing.T) {
		taskReq := map[string]interface{}{
			"title":       "Test Task",
			"description": "This is a test task",
			"priority":    "high",
		}

		rec := ta.request(t, http.MethodPost, "/api/v1/tasks", taskReq, token)

		assert.Equal(t, http.StatusCreated, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "success", response["status"])

		data := response["data"].(map[string]interface{})
		assert.NotEmpty(t, data["id"])
		assert.Equal(t, "Test Task", data["title"])
		assert.Equal(t, "high", data["priority"])
		assert.Equal(t, "pending", data["status"])
	})

	t.Run("create task with invalid priority", func(t *testing.T) {
		taskReq := map[string]interface{}{
			"title":    "Invalid Task",
			"priority": "invalid",
		}

		rec := ta.request(t, http.MethodPost, "/api/v1/tasks", taskReq, token)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("create task with due date", func(t *testing.T) {
		dueDate := time.Now().Add(24 * time.Hour)
		taskReq := map[string]interface{}{
			"title":    "Task with Due Date",
			"due_date": dueDate.Format(time.RFC3339),
		}

		rec := ta.request(t, http.MethodPost, "/api/v1/tasks", taskReq, token)

		assert.Equal(t, http.StatusCreated, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		assert.NotNil(t, data["due_date"])
	})

	t.Run("create task without token", func(t *testing.T) {
		taskReq := map[string]interface{}{
			"title": "Unauthorized Task",
		}

		rec := ta.request(t, http.MethodPost, "/api/v1/tasks", taskReq, "")

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("list tasks", func(t *testing.T) {
		for i := 1; i <= 3; i++ {
			taskReq := map[string]interface{}{
				"title": fmt.Sprintf("Task %d", i),
			}
			ta.request(t, http.MethodPost, "/api/v1/tasks", taskReq, token)
		}

		rec := ta.request(t, http.MethodGet, "/api/v1/tasks", nil, token)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "success", response["status"])

		data := response["data"].(map[string]interface{})
		tasks := data["tasks"].([]interface{})
		assert.GreaterOrEqual(t, len(tasks), 3)
	})

	t.Run("list tasks with pagination", func(t *testing.T) {
		rec := ta.request(t, http.MethodGet, "/api/v1/tasks?limit=2&offset=0", nil, token)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		tasks := data["tasks"].([]interface{})
		assert.LessOrEqual(t, len(tasks), 2)
	})

	t.Run("get task by id", func(t *testing.T) {
		taskReq := map[string]interface{}{
			"title": "Get This Task",
		}

		createRec := ta.request(t, http.MethodPost, "/api/v1/tasks", taskReq, token)

		var createResp map[string]interface{}
		err := json.Unmarshal(createRec.Body.Bytes(), &createResp)
		require.NoError(t, err)
		taskID := createResp["data"].(map[string]interface{})["id"].(string)

		rec := ta.request(t, http.MethodGet, "/api/v1/tasks/"+taskID, nil, token)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		assert.Equal(t, "Get This Task", data["title"])
	})

	t.Run("get non-existent task", func(t *testing.T) {
		rec := ta.request(t, http.MethodGet, "/api/v1/tasks/00000000-0000-0000-0000-000000000000", nil, token)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("get task with invalid id", func(t *testing.T) {
		rec := ta.request(t, http.MethodGet, "/api/v1/tasks/invalid-id", nil, token)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("update task", func(t *testing.T) {
		taskReq := map[string]interface{}{
			"title": "Original Title",
		}

		createRec := ta.request(t, http.MethodPost, "/api/v1/tasks", taskReq, token)

		var createResp map[string]interface{}
		err := json.Unmarshal(createRec.Body.Bytes(), &createResp)
		require.NoError(t, err)
		taskID := createResp["data"].(map[string]interface{})["id"].(string)

		updateReq := map[string]interface{}{
			"title":  "Updated Title",
			"status": "in_progress",
		}

		rec := ta.request(t, http.MethodPut, "/api/v1/tasks/"+taskID, updateReq, token)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		assert.Equal(t, "Updated Title", data["title"])
		assert.Equal(t, "in_progress", data["status"])
	})

	t.Run("update task to completed", func(t *testing.T) {
		taskReq := map[string]interface{}{
			"title": "Task to Complete",
		}

		createRec := ta.request(t, http.MethodPost, "/api/v1/tasks", taskReq, token)

		var createResp map[string]interface{}
		err := json.Unmarshal(createRec.Body.Bytes(), &createResp)
		require.NoError(t, err)
		taskID := createResp["data"].(map[string]interface{})["id"].(string)

		updateReq := map[string]interface{}{
			"status": "completed",
		}

		rec := ta.request(t, http.MethodPut, "/api/v1/tasks/"+taskID, updateReq, token)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		assert.Equal(t, "completed", data["status"])
		assert.NotEmpty(t, data["completed_at"])
	})

	t.Run("update task with invalid status", func(t *testing.T) {
		taskReq := map[string]interface{}{
			"title": "Test Task",
		}

		createRec := ta.request(t, http.MethodPost, "/api/v1/tasks", taskReq, token)

		var createResp map[string]interface{}
		err := json.Unmarshal(createRec.Body.Bytes(), &createResp)
		require.NoError(t, err)
		taskID := createResp["data"].(map[string]interface{})["id"].(string)

		updateReq := map[string]interface{}{
			"status": "invalid_status",
		}

		rec := ta.request(t, http.MethodPut, "/api/v1/tasks/"+taskID, updateReq, token)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("delete task", func(t *testing.T) {
		taskReq := map[string]interface{}{
			"title": "Task to Delete",
		}

		createRec := ta.request(t, http.MethodPost, "/api/v1/tasks", taskReq, token)

		var createResp map[string]interface{}
		err := json.Unmarshal(createRec.Body.Bytes(), &createResp)
		require.NoError(t, err)
		taskID := createResp["data"].(map[string]interface{})["id"].(string)

		rec := ta.request(t, http.MethodDelete, "/api/v1/tasks/"+taskID, nil, token)

		assert.Equal(t, http.StatusNoContent, rec.Code)

		getRec := ta.request(t, http.MethodGet, "/api/v1/tasks/"+taskID, nil, token)
		assert.Equal(t, http.StatusNotFound, getRec.Code)
	})

	t.Run("delete non-existent task", func(t *testing.T) {
		rec := ta.request(t, http.MethodDelete, "/api/v1/tasks/00000000-0000-0000-0000-000000000000", nil, token)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestTaskOwnership(t *testing.T) {
	ta := setupTestApp(t)
	defer ta.cleanup(t)

	user1Req := map[string]interface{}{
		"email":     "user1@example.com",
		"password":  "password123",
		"full_name": "User One",
	}

	ta.request(t, http.MethodPost, "/api/v1/auth/register", user1Req, "")

	user1LoginReq := map[string]interface{}{
		"email":    "user1@example.com",
		"password": "password123",
	}

	user1LoginRec := ta.request(t, http.MethodPost, "/api/v1/auth/login", user1LoginReq, "")

	var user1LoginResp map[string]interface{}
	err := json.Unmarshal(user1LoginRec.Body.Bytes(), &user1LoginResp)
	require.NoError(t, err)
	user1Token := user1LoginResp["data"].(map[string]interface{})["token"].(string)

	taskReq := map[string]interface{}{
		"title": "User1 Task",
	}

	createRec := ta.request(t, http.MethodPost, "/api/v1/tasks", taskReq, user1Token)

	var createResp map[string]interface{}
	err = json.Unmarshal(createRec.Body.Bytes(), &createResp)
	require.NoError(t, err)
	taskID := createResp["data"].(map[string]interface{})["id"].(string)

	user2Req := map[string]interface{}{
		"email":     "user2@example.com",
		"password":  "password123",
		"full_name": "User Two",
	}

	ta.request(t, http.MethodPost, "/api/v1/auth/register", user2Req, "")

	user2LoginReq := map[string]interface{}{
		"email":    "user2@example.com",
		"password": "password123",
	}

	user2LoginRec := ta.request(t, http.MethodPost, "/api/v1/auth/login", user2LoginReq, "")

	var user2LoginResp map[string]interface{}
	err = json.Unmarshal(user2LoginRec.Body.Bytes(), &user2LoginResp)
	require.NoError(t, err)
	user2Token := user2LoginResp["data"].(map[string]interface{})["token"].(string)

	t.Run("user2 cannot access user1 task", func(t *testing.T) {
		rec := ta.request(t, http.MethodGet, "/api/v1/tasks/"+taskID, nil, user2Token)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("user2 cannot update user1 task", func(t *testing.T) {
		updateReq := map[string]interface{}{
			"title": "Hacked Title",
		}

		rec := ta.request(t, http.MethodPut, "/api/v1/tasks/"+taskID, updateReq, user2Token)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("user2 cannot delete user1 task", func(t *testing.T) {
		rec := ta.request(t, http.MethodDelete, "/api/v1/tasks/"+taskID, nil, user2Token)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("user1 can still access their task", func(t *testing.T) {
		rec := ta.request(t, http.MethodGet, "/api/v1/tasks/"+taskID, nil, user1Token)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		assert.Equal(t, "User1 Task", data["title"])
	})
}

func parseConnectionString(connStr string) (string, int) {
	parts := strings.Split(connStr, "@")
	if len(parts) < 2 {
		return "localhost", 5432
	}

	hostPort := strings.Split(parts[1], "/")
	hostPortParts := strings.Split(hostPort[0], ":")

	if len(hostPortParts) < 2 {
		return hostPortParts[0], 5432
	}

	host := hostPortParts[0]
	var port int
	_, err := fmt.Sscanf(hostPortParts[1], "%d", &port)
	if err != nil {
		return host, 5432
	}

	return host, port
}
