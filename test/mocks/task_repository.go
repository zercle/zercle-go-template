package mocks

import (
	"context"
	"math"
	"sync"

	"github.com/zercle/zercle-go-template/internal/feature/task"
)

// MockTaskRepository implements task.Repository for testing.
// All methods are function pointers that can be set to customize behavior.
// If a method function is nil, it returns default values.
// Includes thread-safe in-memory storage for basic CRUD operations.
type MockTaskRepository struct {
	CreateFunc     func(ctx context.Context, task *task.Task) (*task.Task, error)
	GetByIDFunc    func(ctx context.Context, id task.TaskID) (*task.Task, error)
	UpdateFunc     func(ctx context.Context, task *task.Task) (*task.Task, error)
	DeleteFunc     func(ctx context.Context, id task.TaskID) error
	ListFunc       func(ctx context.Context, params *task.ListParams) (*task.ListResult, error)
	CountFunc      func(ctx context.Context, filter task.TaskFilter) (int64, error)
	ExistsByIDFunc func(ctx context.Context, id task.TaskID) (bool, error)

	// In-memory storage for basic testing
	mu      sync.RWMutex
	storage map[task.TaskID]*task.Task
}

// NewMockTaskRepository creates a new MockTaskRepository with initialized storage.
func NewMockTaskRepository() *MockTaskRepository {
	return &MockTaskRepository{
		storage: make(map[task.TaskID]*task.Task),
	}
}

// Create delegates to CreateFunc if set, otherwise stores task in memory.
func (m *MockTaskRepository) Create(ctx context.Context, t *task.Task) (*task.Task, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, t)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if task already exists
	if _, exists := m.storage[t.ID]; exists {
		return nil, task.ErrTaskAlreadyExists
	}

	// Store a copy to avoid external modifications
	taskCopy := *t
	m.storage[taskCopy.ID] = &taskCopy
	return &taskCopy, nil
}

// GetByID delegates to GetByIDFunc if set, otherwise retrieves from in-memory storage.
func (m *MockTaskRepository) GetByID(ctx context.Context, id task.TaskID) (*task.Task, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	t, exists := m.storage[id]
	if !exists {
		return nil, task.ErrTaskNotFound
	}

	// Return a copy to avoid external modifications
	taskCopy := *t
	return &taskCopy, nil
}

// Update delegates to UpdateFunc if set, otherwise updates in-memory storage.
func (m *MockTaskRepository) Update(ctx context.Context, t *task.Task) (*task.Task, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, t)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.storage[t.ID]; !exists {
		return nil, task.ErrTaskNotFound
	}

	// Store a copy to avoid external modifications
	taskCopy := *t
	m.storage[taskCopy.ID] = &taskCopy
	return &taskCopy, nil
}

// Delete delegates to DeleteFunc if set, otherwise deletes from in-memory storage.
func (m *MockTaskRepository) Delete(ctx context.Context, id task.TaskID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.storage[id]; !exists {
		return task.ErrTaskNotFound
	}

	delete(m.storage, id)
	return nil
}

// List delegates to ListFunc if set, otherwise returns all tasks from in-memory storage.
func (m *MockTaskRepository) List(ctx context.Context, params *task.ListParams) (*task.ListResult, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, params)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var tasks []*task.Task
	for _, t := range m.storage {
		if params.Filter.UserID != "" && t.UserID != params.Filter.UserID {
			continue
		}
		if params.Filter.Status != nil && t.Status != *params.Filter.Status {
			continue
		}
		if params.Filter.Priority != nil && t.Priority != *params.Filter.Priority {
			continue
		}
		tasks = append(tasks, t)
	}

	// Apply pagination
	total := int64(len(tasks))
	offset := int(params.Offset)
	limit := int(params.Limit)

	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = len(tasks)
	}

	// Ensure offset and limit fit within int32 bounds to prevent overflow
	if offset > math.MaxInt32 {
		offset = math.MaxInt32
	}
	if limit > math.MaxInt32 {
		limit = math.MaxInt32
	}

	if offset >= len(tasks) {
		tasks = []*task.Task{}
	} else {
		end := min(offset+limit, len(tasks))
		tasks = tasks[offset:end]
	}

	return &task.ListResult{
		Tasks:  tasks,
		Total:  total,
		Limit:  int32(limit),
		Offset: int32(offset),
	}, nil
}

// Count delegates to CountFunc if set, otherwise counts tasks in in-memory storage.
func (m *MockTaskRepository) Count(ctx context.Context, filter task.TaskFilter) (int64, error) {
	if m.CountFunc != nil {
		return m.CountFunc(ctx, filter)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var count int64
	for _, t := range m.storage {
		if filter.UserID != "" && t.UserID != filter.UserID {
			continue
		}
		if filter.Status != nil && t.Status != *filter.Status {
			continue
		}
		if filter.Priority != nil && t.Priority != *filter.Priority {
			continue
		}
		count++
	}

	return count, nil
}

// ExistsByID delegates to ExistsByIDFunc if set, otherwise checks in-memory storage.
func (m *MockTaskRepository) ExistsByID(ctx context.Context, id task.TaskID) (bool, error) {
	if m.ExistsByIDFunc != nil {
		return m.ExistsByIDFunc(ctx, id)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.storage[id]
	return exists, nil
}

// Compile-time check to ensure MockTaskRepository implements task.Repository.
var _ task.Repository = (*MockTaskRepository)(nil)
