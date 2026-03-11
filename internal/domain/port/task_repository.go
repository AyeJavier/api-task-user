package port

import (
	"context"

	"github.com/javier/api-task-user/internal/domain/model"
)

// TaskRepository is the outbound port for task persistence operations.
//
//go:generate mockgen -source=task_repository.go -destination=mocks/mock_task_repository.go -package=mocks
type TaskRepository interface {
	FindByID(ctx context.Context, id string) (*model.Task, error)
	FindByAssignee(ctx context.Context, assigneeID string) ([]*model.Task, error)
	FindAll(ctx context.Context) ([]*model.Task, error)
	Create(ctx context.Context, task *model.Task) error
	Update(ctx context.Context, task *model.Task) error
	Delete(ctx context.Context, id string) error
	CreateComment(ctx context.Context, comment *model.Comment) error
}
