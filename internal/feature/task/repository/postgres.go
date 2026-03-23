package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	task_entity "github.com/zercle/zercle-go-template/internal/feature/task"
	task_sqlc "github.com/zercle/zercle-go-template/internal/infrastructure/database/sqlc/task"
)

// postgresRepository implements task.Repository using PostgreSQL.
type postgresRepository struct {
	queries *task_sqlc.Queries
}

// NewPostgresRepository creates a new PostgreSQL implementation of task.Repository.
func NewPostgresRepository(db task_sqlc.DBTX) task_entity.Repository {
	return &postgresRepository{
		queries: task_sqlc.New(db),
	}
}

// mapPgError maps PostgreSQL errors to domain errors.
func mapPgError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return task_entity.ErrTaskNotFound
	}

	return err
}

// Create inserts a new task into the data store.
func (r *postgresRepository) Create(ctx context.Context, t *task_entity.Task) (*task_entity.Task, error) {
	params := domainTaskToCreateParams(t)

	result, err := r.queries.CreateTask(ctx, &params)
	if err != nil {
		return nil, mapPgError(err)
	}

	return sqlcTaskToDomain(result)
}

// GetByID retrieves a task by its ID.
func (r *postgresRepository) GetByID(ctx context.Context, id task_entity.TaskID) (*task_entity.Task, error) {
	result, err := r.queries.GetTaskByID(ctx, uuid.MustParse(string(id)))
	if err != nil {
		return nil, mapPgError(err)
	}

	return sqlcTaskToDomain(result)
}

// Update modifies an existing task in the data store.
func (r *postgresRepository) Update(ctx context.Context, t *task_entity.Task) (*task_entity.Task, error) {
	params := domainTaskToUpdateParams(t)

	result, err := r.queries.UpdateTask(ctx, &params)
	if err != nil {
		return nil, mapPgError(err)
	}

	return sqlcTaskToDomain(result)
}

// Delete removes a task from the data store.
func (r *postgresRepository) Delete(ctx context.Context, id task_entity.TaskID) error {
	err := r.queries.DeleteTask(ctx, uuid.MustParse(string(id)))
	return mapPgError(err)
}

// List returns a paginated list of tasks matching the given parameters.
func (r *postgresRepository) List(ctx context.Context, params *task_entity.ListParams) (*task_entity.ListResult, error) {
	listParams := domainFilterToListParams(params)

	tasks, err := r.queries.ListTasks(ctx, listParams)
	if err != nil {
		return nil, mapPgError(err)
	}

	countParams := domainFilterToCountParams(params.Filter)

	total, err := r.queries.CountTasks(ctx, countParams)
	if err != nil {
		return nil, mapPgError(err)
	}

	domainTasks := make([]*task_entity.Task, 0, len(tasks))
	for _, t := range tasks {
		domainTask, err := sqlcTaskToDomain(t)
		if err != nil {
			return nil, err
		}
		domainTasks = append(domainTasks, domainTask)
	}

	return &task_entity.ListResult{
		Tasks:  domainTasks,
		Total:  total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}, nil
}

// Count returns the total number of tasks matching the given filter.
func (r *postgresRepository) Count(ctx context.Context, filter task_entity.TaskFilter) (int64, error) {
	params := domainFilterToCountParams(filter)

	total, err := r.queries.CountTasks(ctx, params)
	return total, mapPgError(err)
}

// ExistsByID checks if a task with the given ID exists.
func (r *postgresRepository) ExistsByID(ctx context.Context, id task_entity.TaskID) (bool, error) {
	exists, err := r.queries.ExistsTaskByID(ctx, uuid.MustParse(string(id)))
	return exists, mapPgError(err)
}
