package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zercle/zercle-go-template/internal/domain/task"
	"github.com/zercle/zercle-go-template/internal/domain/task/entity"
)

type taskRepository struct {
	db *pgxpool.Pool
}

func NewTaskRepository(db *pgxpool.Pool) task.TaskRepository {
	return &taskRepository{db: db}
}

var (
	ErrTaskNotFound = pgx.ErrNoRows
)

func (r *taskRepository) Create(ctx context.Context, task *entity.Task) (*entity.Task, error) {
	query := `
		INSERT INTO tasks (id, user_id, title, description, status, priority, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, user_id, title, description, status, priority, due_date, completed_at, created_at, updated_at
	`
	now := time.Now()
	task.ID = uuid.New()
	task.CreatedAt = now
	task.UpdatedAt = now
	if task.Status == "" {
		task.Status = "pending"
	}
	if task.Priority == "" {
		task.Priority = "medium"
	}

	err := r.db.QueryRow(ctx, query,
		task.ID, task.UserID, task.Title, task.Description, task.Status, task.Priority,
		task.DueDate, task.CreatedAt, task.UpdatedAt,
	).Scan(
		&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status, &task.Priority,
		&task.DueDate, &task.CompletedAt, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (r *taskRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Task, error) {
	query := `
		SELECT id, user_id, title, description, status, priority, due_date, completed_at, created_at, updated_at
		FROM tasks WHERE id = $1
	`
	task := &entity.Task{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status, &task.Priority,
		&task.DueDate, &task.CompletedAt, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (r *taskRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Task, int, error) {
	countQuery := `SELECT COUNT(*) FROM tasks WHERE user_id = $1`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, title, description, status, priority, due_date, completed_at, created_at, updated_at
		FROM tasks WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []*entity.Task
	for rows.Next() {
		task := &entity.Task{}
		if err := rows.Scan(
			&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status, &task.Priority,
			&task.DueDate, &task.CompletedAt, &task.CreatedAt, &task.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, task)
	}
	return tasks, total, nil
}

func (r *taskRepository) Update(ctx context.Context, task *entity.Task) (*entity.Task, error) {
	query := `
		UPDATE tasks SET title = $2, description = $3, status = $4, priority = $5, due_date = $6,
			completed_at = $7, updated_at = $8
		WHERE id = $1
		RETURNING id, user_id, title, description, status, priority, due_date, completed_at, created_at, updated_at
	`
	task.UpdatedAt = time.Now()
	err := r.db.QueryRow(ctx, query,
		task.ID, task.Title, task.Description, task.Status, task.Priority,
		task.DueDate, task.CompletedAt, task.UpdatedAt,
	).Scan(
		&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status, &task.Priority,
		&task.DueDate, &task.CompletedAt, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (r *taskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tasks WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
