package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/zercle/zercle-go-template/internal/core/port/mocks"
	healthDomain "github.com/zercle/zercle-go-template/internal/features/health/domain"
	postDomain "github.com/zercle/zercle-go-template/internal/features/post/domain"
	postDto "github.com/zercle/zercle-go-template/internal/features/post/dto"
	userDomain "github.com/zercle/zercle-go-template/internal/features/user/domain"
	userDto "github.com/zercle/zercle-go-template/internal/features/user/dto"
	sharedDomain "github.com/zercle/zercle-go-template/internal/shared/domain"
	sharederrors "github.com/zercle/zercle-go-template/internal/shared/errors"
	"go.uber.org/mock/gomock"
)

// ===========================================
// Mock Data Fixtures
// ===========================================

// Mock Users
var (
	MockUserID1 = uuid.Must(uuid.Parse("11111111-1111-1111-1111-111111111111"))
	MockUserID2 = uuid.Must(uuid.Parse("22222222-2222-2222-2222-222222222222"))
	MockUserID3 = uuid.Must(uuid.Parse("33333333-3333-3333-3333-333333333333"))

	MockUser1 = &userDomain.User{
		ID:        MockUserID1,
		Name:      "Test User 1",
		Email:     "test1@example.com",
		Password:  "$2a$10$rQn3z1z1z1z1z1z1z1z1z.OxQXQXQXQXQXQXQXQXQXQXQXQXQXQX", // "password123"
		CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	MockUser2 = &userDomain.User{
		ID:        MockUserID2,
		Name:      "Test User 2",
		Email:     "test2@example.com",
		Password:  "$2a$10$rQn3z1z1z1z1z1z1z1z1z.OxQXQXQXQXQXQXQXQXQXQXQXQXQXQX",
		CreatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	MockUser3 = &userDomain.User{
		ID:        MockUserID3,
		Name:      "Test User 3",
		Email:     "test3@example.com",
		Password:  "$2a$10$rQn3z1z1z1z1z1z1z1z1z.OxQXQXQXQXQXQXQXQXQXQXQXQXQXQX",
		CreatedAt: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
	}

	MockUsers = []*userDomain.User{MockUser1, MockUser2, MockUser3}
)

// Mock Posts
var (
	MockPostID1 = uuid.Must(uuid.Parse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"))
	MockPostID2 = uuid.Must(uuid.Parse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"))
	MockPostID3 = uuid.Must(uuid.Parse("cccccccc-cccc-cccc-cccc-cccccccccccc"))

	MockPost1 = &postDomain.Post{
		ID:        MockPostID1,
		Title:     "Test Post 1",
		Content:   "This is the content of test post 1. It has enough length.",
		AuthorID:  sharedDomain.NewUserID(MockUserID1),
		CreatedAt: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
	}

	MockPost2 = &postDomain.Post{
		ID:        MockPostID2,
		Title:     "Test Post 2",
		Content:   "This is the content of test post 2. It also has enough length.",
		AuthorID:  sharedDomain.NewUserID(MockUserID2),
		CreatedAt: time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC),
	}

	MockPost3 = &postDomain.Post{
		ID:        MockPostID3,
		Title:     "Test Post 3",
		Content:   "This is the content of test post 3. It has sufficient length.",
		AuthorID:  sharedDomain.NewUserID(MockUserID3),
		CreatedAt: time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC),
	}

	MockPosts = []*postDomain.Post{MockPost1, MockPost2, MockPost3}
)

// Mock Health Status
var MockHealthStatus = &healthDomain.HealthStatus{
	Status:    "ok",
	Database:  "connected",
	Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
}

// ===========================================
// JWT Token Generator
// ===========================================

// GenerateJWTToken generates a valid JWT token for testing
// Note: This is a simplified version for testing purposes
// In real scenarios, use proper JWT library with secret key
func GenerateJWTToken(userID uuid.UUID) string {
	// This is a placeholder. In actual implementation, you would use:
	// token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
	//     "user_id": userID.String(),
	//     "exp":     time.Now().Add(time.Hour).Unix(),
	// })
	// tokenString, _ := token.SignedString([]byte("test-secret"))
	// return tokenString

	// For testing without actual JWT library, return a mock token
	return "mock-jwt-token-for-user-" + userID.String()
}

// ===========================================
// HTTP Request Helpers
// ===========================================

// NewRequest creates a new HTTP request with optional body and auth
func NewRequest(method, path string, body interface{}, withAuth bool, userID uuid.UUID) *http.Request {
	var req *http.Request
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewReader(bodyBytes))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	req.Header.Set("Content-Type", "application/json")

	if withAuth {
		token := GenerateJWTToken(userID)
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req
}

// ===========================================
// Response Validation Helpers
// ===========================================

// JSendResponse represents a JSend-compliant response
type JSendResponse struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ParseJSendResponse parses a JSend response from HTTP response
func ParseJSendResponse(t *testing.T, resp *http.Response) *JSendResponse {
	t.Helper()
	defer resp.Body.Close()

	var jsend JSendResponse
	err := json.NewDecoder(resp.Body).Decode(&jsend)
	assert.NoError(t, err, "Failed to parse JSend response")
	return &jsend
}

// ValidateSuccessResponse validates a successful JSend response
func ValidateSuccessResponse(t *testing.T, resp *http.Response, expectedStatusCode int) *JSendResponse {
	t.Helper()
	assert.Equal(t, expectedStatusCode, resp.StatusCode, "Status code mismatch")

	jsend := ParseJSendResponse(t, resp)
	assert.Equal(t, "success", jsend.Status, "Expected success status")
	return jsend
}

// ValidateErrorResponse validates an error JSend response
func ValidateErrorResponse(t *testing.T, resp *http.Response, expectedStatusCode int) *JSendResponse {
	t.Helper()
	assert.Equal(t, expectedStatusCode, resp.StatusCode, "Status code mismatch")

	jsend := ParseJSendResponse(t, resp)
	assert.Equal(t, "error", jsend.Status, "Expected error status")
	return jsend
}

// ValidateFailResponse validates a fail JSend response
func ValidateFailResponse(t *testing.T, resp *http.Response, expectedStatusCode int) *JSendResponse {
	t.Helper()
	assert.Equal(t, expectedStatusCode, resp.StatusCode, "Status code mismatch")

	jsend := ParseJSendResponse(t, resp)
	assert.Equal(t, "fail", jsend.Status, "Expected fail status")
	return jsend
}

// ===========================================
// Mock Repository Setup Helpers
// ===========================================

// SetupMockUserRepo setups a mock user repository with common expectations
func SetupMockUserRepo(ctrl *gomock.Controller) *mocks.MockUserRepository {
	mockRepo := mocks.NewMockUserRepository(ctrl)
	return mockRepo
}

// SetupMockPostRepo setups a mock post repository with common expectations
func SetupMockPostRepo(ctrl *gomock.Controller) *mocks.MockPostRepository {
	mockRepo := mocks.NewMockPostRepository(ctrl)
	return mockRepo
}

// SetupMockHealthRepo setups a mock health repository with common expectations
func SetupMockHealthRepo(ctrl *gomock.Controller) *mocks.MockHealthRepository {
	mockRepo := mocks.NewMockHealthRepository(ctrl)
	return mockRepo
}

// ===========================================
// DTO Conversion Helpers
// ===========================================

// UserToResponse converts a User domain model to UserResponse DTO
func UserToResponse(user *userDomain.User) *userDto.UserResponse {
	if user == nil {
		return nil
	}
	return &userDto.UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// PostToResponse converts a Post domain model to PostResponse DTO
func PostToResponse(post *postDomain.Post) *postDto.PostResponse {
	if post == nil {
		return nil
	}
	return &postDto.PostResponse{
		ID:        post.ID.String(),
		Title:     post.Title,
		Content:   post.Content,
		AuthorID:  post.AuthorID.String(),
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}
}

// PostsToResponse converts a slice of Post domain models to PostResponse DTOs
func PostsToResponse(posts []*postDomain.Post) []*postDto.PostResponse {
	if posts == nil {
		return nil
	}
	responses := make([]*postDto.PostResponse, len(posts))
	for i, post := range posts {
		responses[i] = PostToResponse(post)
	}
	return responses
}

// ===========================================
// Common Test Data
// ===========================================

// Common test requests
var (
	ValidRegisterRequest = userDto.RegisterRequest{
		Email:    "newuser@example.com",
		Password: "password123",
		Name:     "New User",
	}

	ValidLoginRequest = userDto.LoginRequest{
		Email:    "test1@example.com",
		Password: "password123",
	}

	ValidCreatePostRequest = postDto.CreatePostRequest{
		Title:   "New Post Title",
		Content: "This is the content of the new post. It is long enough.",
	}

	InvalidRegisterRequest_EmptyEmail = userDto.RegisterRequest{
		Email:    "",
		Password: "password123",
		Name:     "New User",
	}

	InvalidRegisterRequest_InvalidEmail = userDto.RegisterRequest{
		Email:    "not-an-email",
		Password: "password123",
		Name:     "New User",
	}

	InvalidRegisterRequest_ShortPassword = userDto.RegisterRequest{
		Email:    "user@example.com",
		Password: "short",
		Name:     "New User",
	}

	InvalidRegisterRequest_ShortName = userDto.RegisterRequest{
		Email:    "user@example.com",
		Password: "password123",
		Name:     "X",
	}

	InvalidCreatePostRequest_EmptyTitle = postDto.CreatePostRequest{
		Title:   "",
		Content: "This is the content of the new post. It is long enough.",
	}

	InvalidCreatePostRequest_ShortTitle = postDto.CreatePostRequest{
		Title:   "Hi",
		Content: "This is the content of the new post. It is long enough.",
	}

	InvalidCreatePostRequest_EmptyContent = postDto.CreatePostRequest{
		Title:   "New Post Title",
		Content: "",
	}

	InvalidCreatePostRequest_ShortContent = postDto.CreatePostRequest{
		Title:   "New Post Title",
		Content: "Short",
	}
)

// ===========================================
// Error Expectations
// ===========================================

// Common mock error expectations
var (
	ErrNotFound       = sharederrors.ErrNotFound
	ErrDuplicate      = sharederrors.ErrDuplicate
	ErrInvalidCreds   = sharederrors.ErrInvalidCreds
	ErrInternalServer = sharederrors.ErrInternalServer
)
