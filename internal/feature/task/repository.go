package task

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	task_sqlc "github.com/zercle/zercle-go-template/internal/infrastructure/database/sqlc/task"
)

// Repository defines the interface for task storage.
type Repository interface {
	Create(ctx context.Context, task *Task) (*Task, error)
	GetByID(ctx context.Context, id ID) (*Task, error)
	Update(ctx context.Context, task *Task) (*Task, error)
	Delete(ctx context.Context, id ID) error
	List(ctx context.Context, params *ListParams) (*ListResult, error)
	Count(ctx context.Context, filter Filter) (int64, error)
	ExistsByID(ctx context.Context, id ID) (bool, error)
}

type postgresRepository struct {
	queries *task_sqlc.Queries
}

// NewPostgresRepository creates a new Postgres task repository.
func NewPostgresRepository(db task_sqlc.DBTX) Repository {
	return &postgresRepository{
		queries: task_sqlc.New(db),
	}
}

func mapPgError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrTaskNotFound
	}

	return err
}

func (r *postgresRepository) Create(ctx context.Context, t *Task) (*Task, error) {
	params := domainTaskToCreateParams(t)

	result, err := r.queries.CreateTask(ctx, &params)
	if err != nil {
		return nil, mapPgError(err)
	}

	return sqlcTaskToDomain(result)
}

func (r *postgresRepository) GetByID(ctx context.Context, id ID) (*Task, error) {
	uid, err := uuid.Parse(string(id))
	if err != nil {
		return nil, err
	}
	result, err := r.queries.GetTaskByID(ctx, uid)
	if err != nil {
		return nil, mapPgError(err)
	}

	return sqlcTaskToDomain(result)
}

func (r *postgresRepository) Update(ctx context.Context, t *Task) (*Task, error) {
	params := domainTaskToUpdateParams(t)

	result, err := r.queries.UpdateTask(ctx, &params)
	if err != nil {
		return nil, mapPgError(err)
	}

	return sqlcTaskToDomain(result)
}

func (r *postgresRepository) Delete(ctx context.Context, id ID) error {
	uid, err := uuid.Parse(string(id))
	if err != nil {
		return err
	}
	err = r.queries.DeleteTask(ctx, uid)
	return mapPgError(err)
}

func (r *postgresRepository) List(ctx context.Context, params *ListParams) (*ListResult, error) {
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

	domainTasks := make([]*Task, 0, len(tasks))
	for _, t := range tasks {
		domainTask, err := sqlcTaskToDomain(t)
		if err != nil {
			return nil, err
		}
		domainTasks = append(domainTasks, domainTask)
	}

	return &ListResult{
		Tasks:  domainTasks,
		Total:  total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}, nil
}

func (r *postgresRepository) Count(ctx context.Context, filter Filter) (int64, error) {
	params := domainFilterToCountParams(filter)

	total, err := r.queries.CountTasks(ctx, params)
	return total, mapPgError(err)
}

func (r *postgresRepository) ExistsByID(ctx context.Context, id ID) (bool, error) {
	uid, err := uuid.Parse(string(id))
	if err != nil {
		return false, err
	}
	exists, err := r.queries.ExistsTaskByID(ctx, uid)
	return exists, mapPgError(err)
}

func sqlcTaskToDomain(m *task_sqlc.Task) (*Task, error) {
	createdAt := time.Time{}
	if m.CreatedAt.Valid {
		createdAt = m.CreatedAt.Time
	}
	updatedAt := time.Time{}
	if m.UpdatedAt.Valid {
		updatedAt = m.UpdatedAt.Time
	}

	var description string
	if m.Description.Valid {
		description = m.Description.String
	}

	return NewWithID(
		ID(m.ID.String()),
		m.Title,
		description,
		Status(m.Status),
		Priority(m.Priority),
		m.UserID.String(),
		createdAt,
		updatedAt,
	)
}

func domainTaskToCreateParams(t *Task) task_sqlc.CreateTaskParams {
	return task_sqlc.CreateTaskParams{
		ID:          uuid.MustParse(string(t.ID)),
		Title:       t.Title,
		Description: pgtypeText(t.Description),
		Status:      string(t.Status),
		Priority:    string(t.Priority),
		UserID:      uuid.MustParse(t.UserID),
		CreatedAt:   pgtypeTimestamptz(t.CreatedAt),
		UpdatedAt:   pgtypeTimestamptz(t.UpdatedAt),
	}
}

func domainTaskToUpdateParams(t *Task) task_sqlc.UpdateTaskParams {
	return task_sqlc.UpdateTaskParams{
		ID:          uuid.MustParse(string(t.ID)),
		Title:       t.Title,
		Description: pgtypeText(t.Description),
		Status:      string(t.Status),
		Priority:    string(t.Priority),
		UpdatedAt:   pgtypeTimestamptz(t.UpdatedAt),
	}
}

func domainFilterToListParams(params *ListParams) *task_sqlc.ListTasksParams {
	var userIDFilter pgtype.UUID
	if params.Filter.UserID != "" {
		userIDFilter = pgtype.UUID{
			Bytes: uuid.MustParse(params.Filter.UserID),
			Valid: true,
		}
	}

	var statusFilter string
	if params.Filter.Status != nil {
		statusFilter = string(*params.Filter.Status)
	}

	var priorityFilter string
	if params.Filter.Priority != nil {
		priorityFilter = string(*params.Filter.Priority)
	}

	return &task_sqlc.ListTasksParams{
		Column1: userIDFilter,
		Column2: statusFilter,
		Column3: priorityFilter,
		Limit:   params.Limit,
		Offset:  params.Offset,
	}
}

func domainFilterToCountParams(filter Filter) *task_sqlc.CountTasksParams {
	var userIDFilter pgtype.UUID
	if filter.UserID != "" {
		userIDFilter = pgtype.UUID{
			Bytes: uuid.MustParse(filter.UserID),
			Valid: true,
		}
	}

	var statusFilter string
	if filter.Status != nil {
		statusFilter = string(*filter.Status)
	}

	var priorityFilter string
	if filter.Priority != nil {
		priorityFilter = string(*filter.Priority)
	}

	return &task_sqlc.CountTasksParams{
		Column1: userIDFilter,
		Column2: statusFilter,
		Column3: priorityFilter,
	}
}

func pgtypeText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{
		String: s,
		Valid:  true,
	}
}

func pgtypeTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  t,
		Valid: true,
	}
}
