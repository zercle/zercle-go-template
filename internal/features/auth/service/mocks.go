package service

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/zercle/zercle-go-template/internal/features/auth/domain"
)

// MockUserRepository is a mock implementation of UserRepository for testing.
type MockUserRepository struct {
	mu    sync.RWMutex
	users map[uuid.UUID]*domain.User
}

// NewUserRepoMock creates a new MockUserRepository instance.
func NewUserRepoMock() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[uuid.UUID]*domain.User),
	}
}

// FindByID retrieves a user by their ID.
func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, nil
}

// FindByEmail retrieves a user by their email address.
func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, nil
}

// FindByUsername retrieves a user by their username.
func (m *MockUserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, nil
}

// Create adds a new user to the mock repository.
func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[user.ID] = user
	return nil
}

// Update modifies an existing user in the mock repository.
func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[user.ID] = user
	return nil
}

// Delete removes a user from the mock repository.
func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.users, id)
	return nil
}

// MockSessionRepository is a mock implementation of SessionRepository for testing.
type MockSessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]*domain.Session
}

// NewSessionRepoMock creates a new MockSessionRepository instance.
func NewSessionRepoMock() *MockSessionRepository {
	return &MockSessionRepository{
		sessions: make(map[string]*domain.Session),
	}
}

// FindByToken retrieves a session by its token.
func (m *MockSessionRepository) FindByToken(ctx context.Context, token string) (*domain.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if session, ok := m.sessions[token]; ok {
		return session, nil
	}
	return nil, nil
}

// Create adds a new session to the mock repository.
func (m *MockSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.Token] = session
	return nil
}

// Delete removes a session from the mock repository.
func (m *MockSessionRepository) Delete(ctx context.Context, token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, token)
	return nil
}

// DeleteByUserID removes all sessions associated with a user ID.
func (m *MockSessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for token, session := range m.sessions {
		if session.UserID == userID {
			delete(m.sessions, token)
		}
	}
	return nil
}
