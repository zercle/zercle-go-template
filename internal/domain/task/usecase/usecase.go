package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/internal/domain/task"
	"github.com/zercle/zercle-go-template/internal/domain/task/entity"
	"github.com/zercle/zercle-go-template/internal/domain/task/repository"
	"github.com/zercle/zercle-go-template/internal/domain/task/request"
	"github.com/zercle/zercle-go-template/internal/domain/task/response"
	"github.com/zercle/zercle-go-template/internal/infrastructure/logger"
)

var (
	ErrTaskNotFound        = errors.New("task not found")
	ErrTaskNotOwned        = errors.New("task not owned by user")
	ErrInvalidTaskStatus   = errors.New("invalid task status")
	ErrInvalidTaskPriority = errors.New("invalid task priority")
)

type taskUseCase struct {
	repo task.TaskRepository
	log  *logger.Logger
}

func NewTaskUseCase(repo task.TaskRepository, log *logger.Logger) task.TaskService {
	return &taskUseCase{repo: repo, log: log}
}

func (uc *taskUseCase) CreateTask(ctx context.Context, userID uuid.UUID, req request.CreateTask) (*response.TaskResponse, error) {
	task := &entity.Task{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		DueDate:     req.DueDate,
		Status:      "pending",
	}

	created, err := uc.repo.Create(ctx, task)
	if err != nil {
		return nil, err
	}

	return toTaskResponse(created), nil
}

func (uc *taskUseCase) GetTask(ctx context.Context, userID, taskID uuid.UUID) (*response.TaskResponse, error) {
	task, err := uc.repo.GetByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	if task.UserID != userID {
		return nil, ErrTaskNotOwned
	}

	return toTaskResponse(task), nil
}

func (uc *taskUseCase) ListTasks(ctx context.Context, userID uuid.UUID, limit, offset int) (*response.ListTasksResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	tasks, total, err := uc.repo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	taskResponses := make([]response.TaskResponse, len(tasks))
	for i, t := range tasks {
		taskResponses[i] = *toTaskResponse(t)
	}

	return &response.ListTasksResponse{
		Tasks: taskResponses,
		Total: total,
	}, nil
}

func (uc *taskUseCase) UpdateTask(ctx context.Context, userID, taskID uuid.UUID, req request.UpdateTask) (*response.TaskResponse, error) {
	task, err := uc.repo.GetByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	if task.UserID != userID {
		return nil, ErrTaskNotOwned
	}

	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Status != nil {
		if !isValidStatus(*req.Status) {
			return nil, ErrInvalidTaskStatus
		}
		task.Status = *req.Status
		if *req.Status == "completed" {
			now := time.Now()
			task.CompletedAt = &now
		}
	}
	if req.Priority != nil {
		if !isValidPriority(*req.Priority) {
			return nil, ErrInvalidTaskPriority
		}
		task.Priority = *req.Priority
	}
	if req.DueDate != nil {
		task.DueDate = req.DueDate
	}

	updated, err := uc.repo.Update(ctx, task)
	if err != nil {
		return nil, err
	}

	return toTaskResponse(updated), nil
}

func (uc *taskUseCase) DeleteTask(ctx context.Context, userID, taskID uuid.UUID) error {
	task, err := uc.repo.GetByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			return ErrTaskNotFound
		}
		return err
	}

	if task.UserID != userID {
		return ErrTaskNotOwned
	}

	return uc.repo.Delete(ctx, taskID)
}

func toTaskResponse(task *entity.Task) *response.TaskResponse {
	return &response.TaskResponse{
		ID:          task.ID,
		UserID:      task.UserID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Priority:    task.Priority,
		DueDate:     task.DueDate,
		CompletedAt: task.CompletedAt,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

func isValidStatus(status string) bool {
	validStatuses := map[string]bool{
		"pending":     true,
		"in_progress": true,
		"completed":   true,
		"cancelled":   true,
	}
	return validStatuses[status]
}

func isValidPriority(priority string) bool {
	validPriorities := map[string]bool{
		"low":    true,
		"medium": true,
		"high":   true,
		"urgent": true,
	}
	return validPriorities[priority]
}
