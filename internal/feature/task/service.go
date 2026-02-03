package task

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"
)

// Service provides task business logic.
type Service struct {
	repo   Repository
	logger zerolog.Logger
}

// NewService creates a new task service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo, logger: zerolog.Nop()}
}

// NewServiceWithLogger creates a new task service with a custom logger.
func NewServiceWithLogger(repo Repository, logger zerolog.Logger) *Service {
	return &Service{repo: repo, logger: logger}
}

// Create creates a new task.
func (s *Service) Create(ctx context.Context, input *CreateTaskInput) (*Response, error) {
	if input == nil {
		return nil, errors.New("create input is required")
	}
	if input.Title == "" {
		return nil, errors.New("title is required")
	}
	if input.Priority == "" {
		return nil, errors.New("priority is required")
	}
	if input.UserID == "" {
		return nil, errors.New("user_id is required")
	}

	if !Priority(input.Priority).IsValid() {
		return nil, ErrInvalidTask
	}

	newTask, err := createInputToDomain(input)
	if err != nil {
		return nil, err
	}

	created, err := s.repo.Create(ctx, newTask)
	if err != nil {
		return nil, err
	}

	return domainToResponse(created), nil
}

// Get retrieves a task by ID.
func (s *Service) Get(ctx context.Context, id string) (*Response, error) {
	if id == "" {
		return nil, ErrTaskNotFound
	}

	taskID := ID(id)
	found, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			return nil, err
		}
		return nil, err
	}

	return domainToResponse(found), nil
}

// List retrieves a paginated list of tasks.
func (s *Service) List(ctx context.Context, params *ListQuery) (*ListResponse, error) {
	if params == nil {
		params = &ListQuery{}
	}

	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	if params.Offset < 0 {
		params.Offset = 0
	}

	repoParams := listParamsToRepo(*params)

	result, err := s.repo.List(ctx, &repoParams)
	if err != nil {
		return nil, err
	}

	return listResultToDTO(result), nil
}

// Update updates an existing task.
func (s *Service) Update(ctx context.Context, id string, input *UpdateTaskInput) (*Response, error) {
	if id == "" {
		return nil, ErrTaskNotFound
	}
	if input == nil {
		return nil, errors.New("update input is required")
	}

	taskID := ID(id)

	existing, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			return nil, err
		}
		return nil, err
	}

	if input.Status != nil {
		if !Status(*input.Status).IsValid() {
			return nil, ErrInvalidTask
		}
	}

	if input.Priority != nil {
		if !Priority(*input.Priority).IsValid() {
			return nil, ErrInvalidTask
		}
	}

	updated := &Task{
		ID:          existing.ID,
		Title:       existing.Title,
		Description: existing.Description,
		Status:      existing.Status,
		Priority:    existing.Priority,
		UserID:      existing.UserID,
		CreatedAt:   existing.CreatedAt,
		UpdatedAt:   time.Now().UTC(),
	}

	if input.Title != nil {
		updated.Title = *input.Title
	}
	if input.Description != nil {
		updated.Description = *input.Description
	}
	if input.Status != nil {
		updated.Status = Status(*input.Status)
	}
	if input.Priority != nil {
		updated.Priority = Priority(*input.Priority)
	}

	result, err := s.repo.Update(ctx, updated)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			return nil, err
		}
		return nil, err
	}

	return domainToResponse(result), nil
}

// Delete removes a task by ID.
func (s *Service) Delete(ctx context.Context, id string) error {
	if id == "" {
		return ErrTaskNotFound
	}

	taskID := ID(id)

	exists, err := s.repo.ExistsByID(ctx, taskID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrTaskNotFound
	}

	if err := s.repo.Delete(ctx, taskID); err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			return err
		}
		return err
	}

	return nil
}

func domainToResponse(t *Task) *Response {
	if t == nil {
		return nil
	}
	return &Response{
		ID:          string(t.ID),
		Title:       t.Title,
		Description: t.Description,
		Status:      string(t.Status),
		Priority:    string(t.Priority),
		UserID:      t.UserID,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func createInputToDomain(dto *CreateTaskInput) (*Task, error) {
	if dto == nil {
		return nil, errors.New("input is nil")
	}
	return New(dto.Title, dto.Description, Priority(dto.Priority), dto.UserID)
}

func listParamsToRepo(dto ListQuery) ListParams {
	params := ListParams{
		Limit:  dto.Limit,
		Offset: dto.Offset,
	}

	if dto.UserID != "" {
		params.Filter.UserID = dto.UserID
	}
	if dto.Status != "" {
		status := Status(dto.Status)
		params.Filter.Status = &status
	}
	if dto.Priority != "" {
		priority := Priority(dto.Priority)
		params.Filter.Priority = &priority
	}

	return params
}

func listResultToDTO(result *ListResult) *ListResponse {
	if result == nil {
		return nil
	}

	tasks := make([]*Response, 0, len(result.Tasks))
	for _, t := range result.Tasks {
		tasks = append(tasks, domainToResponse(t))
	}

	return &ListResponse{
		Tasks:  tasks,
		Total:  result.Total,
		Limit:  result.Limit,
		Offset: result.Offset,
	}
}
