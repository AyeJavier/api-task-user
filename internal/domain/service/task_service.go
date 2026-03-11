package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/javier/api-task-user/internal/domain/model"
	"github.com/javier/api-task-user/internal/domain/port"
	"github.com/javier/api-task-user/pkg/apperror"
)

// TaskService handles task management business logic.
type TaskService struct {
	taskRepo port.TaskRepository
	userRepo port.UserRepository
}

// NewTaskService creates a new TaskService.
func NewTaskService(taskRepo port.TaskRepository, userRepo port.UserRepository) *TaskService {
	return &TaskService{taskRepo: taskRepo, userRepo: userRepo}
}

// CreateTaskInput holds the data required to create a task.
type CreateTaskInput struct {
	Title       string
	Description string
	AssigneeID  string
	DueDate     time.Time
}

// Create creates a new task and assigns it to an Executor. Admin only.
func (s *TaskService) Create(ctx context.Context, requester *model.User, input CreateTaskInput) (*model.Task, error) {
	if !requester.IsAdmin() {
		return nil, apperror.ErrForbidden
	}

	assignee, err := s.userRepo.FindByID(ctx, input.AssigneeID)
	if err != nil {
		return nil, apperror.ErrUserNotFound
	}
	if !assignee.IsExecutor() {
		return nil, apperror.ErrForbidden
	}

	task := &model.Task{
		ID:          uuid.NewString(),
		Title:       input.Title,
		Description: input.Description,
		Status:      model.TaskStatusAssigned,
		AssigneeID:  input.AssigneeID,
		DueDate:     input.DueDate,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("persisting task: %w", err)
	}
	return task, nil
}

// Update updates a task's details. Admin only. Task must be in ASSIGNED status.
func (s *TaskService) Update(ctx context.Context, requester *model.User, task *model.Task) error {
	if !requester.IsAdmin() {
		return apperror.ErrForbidden
	}
	existing, err := s.taskRepo.FindByID(ctx, task.ID)
	if err != nil {
		return apperror.ErrTaskNotFound
	}
	if !existing.IsMutableByAdmin() {
		return apperror.ErrTaskNotMutable
	}
	task.UpdatedAt = time.Now()
	return s.taskRepo.Update(ctx, task)
}

// Delete removes a task. Admin only. Task must be in ASSIGNED status.
func (s *TaskService) Delete(ctx context.Context, requester *model.User, id string) error {
	if !requester.IsAdmin() {
		return apperror.ErrForbidden
	}
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return apperror.ErrTaskNotFound
	}
	if !task.IsMutableByAdmin() {
		return apperror.ErrTaskNotMutable
	}
	return s.taskRepo.Delete(ctx, id)
}

// ListMyTasks returns tasks assigned to the requesting Executor.
func (s *TaskService) ListMyTasks(ctx context.Context, requester *model.User) ([]*model.Task, error) {
	if !requester.IsExecutor() {
		return nil, apperror.ErrForbidden
	}
	return s.taskRepo.FindByAssignee(ctx, requester.ID)
}

// ListAllTasks returns all tasks. Auditor only.
func (s *TaskService) ListAllTasks(ctx context.Context, requester *model.User) ([]*model.Task, error) {
	if !requester.IsAuditor() {
		return nil, apperror.ErrForbidden
	}
	return s.taskRepo.FindAll(ctx)
}

// ChangeStatus advances the task state machine. Executor only.
func (s *TaskService) ChangeStatus(ctx context.Context, requester *model.User, taskID string, target model.TaskStatus) error {
	if !requester.IsExecutor() {
		return apperror.ErrForbidden
	}
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return apperror.ErrTaskNotFound
	}
	if task.AssigneeID != requester.ID {
		return apperror.ErrForbidden
	}
	if err := task.Transition(target); err != nil {
		return err
	}
	return s.taskRepo.Update(ctx, task)
}

// AddComment adds a comment to an expired task. Executor only.
func (s *TaskService) AddComment(ctx context.Context, requester *model.User, taskID, body string) (*model.Comment, error) {
	if !requester.IsExecutor() {
		return nil, apperror.ErrForbidden
	}
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, apperror.ErrTaskNotFound
	}
	if task.AssigneeID != requester.ID {
		return nil, apperror.ErrForbidden
	}
	if !task.IsExpired() {
		return nil, fmt.Errorf("comments can only be added to expired tasks")
	}

	comment := &model.Comment{
		ID:        uuid.NewString(),
		TaskID:    taskID,
		AuthorID:  requester.ID,
		Body:      body,
		CreatedAt: time.Now(),
	}
	if err := s.taskRepo.CreateComment(ctx, comment); err != nil {
		return nil, fmt.Errorf("persisting comment: %w", err)
	}
	return comment, nil
}
