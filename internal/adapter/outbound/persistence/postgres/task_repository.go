package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/javier/api-task-user/internal/domain/model"
	"github.com/javier/api-task-user/pkg/apperror"
)

// TaskRepository implements port.TaskRepository using PostgreSQL.
type TaskRepository struct {
	db *pgxpool.Pool
}

// NewTaskRepository creates a new PostgreSQL-backed TaskRepository.
func NewTaskRepository(db *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{db: db}
}

const taskSelectCols = `id, title, description, status, assignee_id, due_date, created_at, updated_at`

func (r *TaskRepository) FindByID(ctx context.Context, id string) (*model.Task, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+taskSelectCols+` FROM tasks WHERE id = $1`, id)
	return scanTask(row)
}

func (r *TaskRepository) FindByAssignee(ctx context.Context, assigneeID string) ([]*model.Task, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+taskSelectCols+` FROM tasks WHERE assignee_id = $1 ORDER BY created_at DESC`, assigneeID)
	if err != nil {
		return nil, fmt.Errorf("querying tasks by assignee: %w", err)
	}
	defer rows.Close()
	return scanTasks(rows)
}

func (r *TaskRepository) FindAll(ctx context.Context) ([]*model.Task, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+taskSelectCols+` FROM tasks ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("querying all tasks: %w", err)
	}
	defer rows.Close()
	return scanTasks(rows)
}

func (r *TaskRepository) Create(ctx context.Context, task *model.Task) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO tasks (id, title, description, status, assignee_id, due_date, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		task.ID, task.Title, task.Description, task.Status,
		task.AssigneeID, task.DueDate, task.CreatedAt, task.UpdatedAt)
	if err != nil {
		return fmt.Errorf("inserting task: %w", err)
	}
	return nil
}

func (r *TaskRepository) Update(ctx context.Context, task *model.Task) error {
	_, err := r.db.Exec(ctx,
		`UPDATE tasks SET title=$1, description=$2, status=$3, due_date=$4, updated_at=$5
		 WHERE id=$6`,
		task.Title, task.Description, task.Status, task.DueDate, task.UpdatedAt, task.ID)
	if err != nil {
		return fmt.Errorf("updating task: %w", err)
	}
	return nil
}

func (r *TaskRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting task: %w", err)
	}
	return nil
}

func (r *TaskRepository) CreateComment(ctx context.Context, comment *model.Comment) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO comments (id, task_id, author_id, body, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		comment.ID, comment.TaskID, comment.AuthorID, comment.Body, comment.CreatedAt)
	if err != nil {
		return fmt.Errorf("inserting comment: %w", err)
	}
	return nil
}

// --- scan helpers ---

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTask(row rowScanner) (*model.Task, error) {
	t := &model.Task{}
	err := row.Scan(&t.ID, &t.Title, &t.Description, &t.Status,
		&t.AssigneeID, &t.DueDate, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, apperror.ErrTaskNotFound
	}
	return t, nil
}

func scanTasks(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]*model.Task, error) {
	var tasks []*model.Task
	for rows.Next() {
		t := &model.Task{}
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status,
			&t.AssigneeID, &t.DueDate, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning task row: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}
