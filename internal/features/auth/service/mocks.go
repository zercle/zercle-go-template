package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/internal/features/auth/domain"
)

type MockUserRepository struct {
	users map[uuid.UUID]*domain.User
}

func NewUserRepoMock() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[uuid.UUID]*domain.User),
	}
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, nil
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, nil
}

func (m *MockUserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, nil
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.users, id)
	return nil
}

type MockSessionRepository struct {
	sessions map[string]*domain.Session
}

func NewSessionRepoMock() *MockSessionRepository {
	return &MockSessionRepository{
		sessions: make(map[string]*domain.Session),
	}
}

func (m *MockSessionRepository) FindByToken(ctx context.Context, token string) (*domain.Session, error) {
	if session, ok := m.sessions[token]; ok {
		return session, nil
	}
	return nil, nil
}

func (m *MockSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	m.sessions[session.Token] = session
	return nil
}

func (m *MockSessionRepository) Delete(ctx context.Context, token string) error {
	delete(m.sessions, token)
	return nil
}

func (m *MockSessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	for token, session := range m.sessions {
		if session.UserID == userID {
			delete(m.sessions, token)
		}
	}
	return nil
}
