package mocks

import (
	"context"

	"github.com/zercle/zercle-go-template/internal/feature/user"
)

// MockUserRepository implements user.Repository for testing.
// All methods are function pointers that can be set to customize behavior.
// If a method function is nil, it returns default values.
type MockUserRepository struct {
	CreateFunc     func(ctx context.Context, user *user.User) (*user.User, error)
	GetByIDFunc    func(ctx context.Context, id user.UserID) (*user.User, error)
	GetByEmailFunc func(ctx context.Context, email string) (*user.User, error)
	UpdateFunc     func(ctx context.Context, user *user.User) (*user.User, error)
	DeleteFunc     func(ctx context.Context, id user.UserID) error
	ListFunc       func(ctx context.Context, params *user.ListParams) (*user.ListResult, error)
	ExistsFunc     func(ctx context.Context, email string) (bool, error)
	ExistsByIDFunc func(ctx context.Context, id user.UserID) (bool, error)
}

// Create delegates to CreateFunc if set, otherwise returns nil.
func (m *MockUserRepository) Create(ctx context.Context, user *user.User) (*user.User, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, user)
	}
	return nil, nil
}

// GetByID delegates to GetByIDFunc if set, otherwise returns nil.
func (m *MockUserRepository) GetByID(ctx context.Context, id user.UserID) (*user.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

// GetByEmail delegates to GetByEmailFunc if set, otherwise returns nil.
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(ctx, email)
	}
	return nil, nil
}

// Update delegates to UpdateFunc if set, otherwise returns nil.
func (m *MockUserRepository) Update(ctx context.Context, user *user.User) (*user.User, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, user)
	}
	return nil, nil
}

// Delete delegates to DeleteFunc if set, otherwise returns nil.
func (m *MockUserRepository) Delete(ctx context.Context, id user.UserID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// List delegates to ListFunc if set, otherwise returns nil.
func (m *MockUserRepository) List(ctx context.Context, params *user.ListParams) (*user.ListResult, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, params)
	}
	return nil, nil
}

// Exists delegates to ExistsFunc if set, otherwise returns false.
func (m *MockUserRepository) Exists(ctx context.Context, email string) (bool, error) {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(ctx, email)
	}
	return false, nil
}

// ExistsByID delegates to ExistsByIDFunc if set, otherwise returns false.
func (m *MockUserRepository) ExistsByID(ctx context.Context, id user.UserID) (bool, error) {
	if m.ExistsByIDFunc != nil {
		return m.ExistsByIDFunc(ctx, id)
	}
	return false, nil
}

// Compile-time check to ensure MockUserRepository implements user.Repository.
var _ user.Repository = (*MockUserRepository)(nil)
