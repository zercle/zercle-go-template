//go:generate go run go.uber.org/mock/mockgen@latest -source=$GOFILE -destination=mock/$GOFILE -package=mock

package task

import (
	"context"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/zercle/zercle-go-template/internal/domain/task/entity"
	"github.com/zercle/zercle-go-template/internal/domain/task/request"
	"github.com/zercle/zercle-go-template/internal/domain/task/response"
)

type TaskRepository interface {
	Create(ctx context.Context, task *entity.Task) (*entity.Task, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Task, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Task, int, error)
	Update(ctx context.Context, task *entity.Task) (*entity.Task, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type TaskService interface {
	CreateTask(ctx context.Context, userID uuid.UUID, req request.CreateTask) (*response.TaskResponse, error)
	GetTask(ctx context.Context, userID, taskID uuid.UUID) (*response.TaskResponse, error)
	ListTasks(ctx context.Context, userID uuid.UUID, limit, offset int) (*response.ListTasksResponse, error)
	UpdateTask(ctx context.Context, userID, taskID uuid.UUID, req request.UpdateTask) (*response.TaskResponse, error)
	DeleteTask(ctx context.Context, userID, taskID uuid.UUID) error
}

type TaskHandler interface {
	CreateTask(c echo.Context) error
	GetTask(c echo.Context) error
	ListTasks(c echo.Context) error
	UpdateTask(c echo.Context) error
	DeleteTask(c echo.Context) error
	RegisterRoutes(protected *echo.Group)
}
