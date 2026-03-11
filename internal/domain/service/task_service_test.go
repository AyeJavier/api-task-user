package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/javier/api-task-user/internal/domain/model"
	"github.com/javier/api-task-user/internal/domain/port/mocks"
	"github.com/javier/api-task-user/internal/domain/service"
	"github.com/javier/api-task-user/pkg/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupTaskService() (*service.TaskService, *mocks.MockTaskRepository, *mocks.MockUserRepository) {
	taskRepo := new(mocks.MockTaskRepository)
	userRepo := new(mocks.MockUserRepository)
	svc := service.NewTaskService(taskRepo, userRepo)
	return svc, taskRepo, userRepo
}

func futureDate() time.Time { return time.Now().Add(7 * 24 * time.Hour) }
func pastDate() time.Time   { return time.Now().Add(-24 * time.Hour) }

// ==================== ADMIN: Create Task ====================

func TestTaskService_Create_AdminSuccess(t *testing.T) {
	svc, taskRepo, userRepo := setupTaskService()
	ctx := context.Background()

	assignee := &model.User{ID: "exec-1", Profile: model.ProfileExecutor}
	userRepo.On("FindByID", ctx, "exec-1").Return(assignee, nil)
	taskRepo.On("Create", ctx, mock.MatchedBy(func(task *model.Task) bool {
		return task.Title == "Deploy v2" &&
			task.AssigneeID == "exec-1" &&
			task.Status == model.TaskStatusAssigned
	})).Return(nil)

	input := service.CreateTaskInput{
		Title:       "Deploy v2",
		Description: "Deploy version 2 to production",
		AssigneeID:  "exec-1",
		DueDate:     futureDate(),
	}

	task, err := svc.Create(ctx, adminUser(), input)

	require.NoError(t, err)
	assert.Equal(t, "Deploy v2", task.Title)
	assert.Equal(t, model.TaskStatusAssigned, task.Status)
	assert.NotEmpty(t, task.ID)
}

func TestTaskService_Create_NonAdminForbidden(t *testing.T) {
	svc, _, _ := setupTaskService()
	ctx := context.Background()

	input := service.CreateTaskInput{Title: "T", AssigneeID: "x", DueDate: futureDate()}

	task, err := svc.Create(ctx, executorUser(), input)
	assert.Nil(t, task)
	assert.ErrorIs(t, err, apperror.ErrForbidden)
}

func TestTaskService_Create_AssigneeNotExecutor(t *testing.T) {
	svc, _, userRepo := setupTaskService()
	ctx := context.Background()

	auditor := &model.User{ID: "aud-1", Profile: model.ProfileAuditor}
	userRepo.On("FindByID", ctx, "aud-1").Return(auditor, nil)

	input := service.CreateTaskInput{Title: "T", AssigneeID: "aud-1", DueDate: futureDate()}

	task, err := svc.Create(ctx, adminUser(), input)
	assert.Nil(t, task)
	assert.ErrorIs(t, err, apperror.ErrForbidden)
}

func TestTaskService_Create_AssigneeNotFound(t *testing.T) {
	svc, _, userRepo := setupTaskService()
	ctx := context.Background()

	userRepo.On("FindByID", ctx, "nonexistent").Return(nil, apperror.ErrUserNotFound)

	input := service.CreateTaskInput{Title: "T", AssigneeID: "nonexistent", DueDate: futureDate()}

	task, err := svc.Create(ctx, adminUser(), input)
	assert.Nil(t, task)
	assert.ErrorIs(t, err, apperror.ErrUserNotFound)
}

// ==================== ADMIN: Update Task ====================

func TestTaskService_Update_AdminAssignedSuccess(t *testing.T) {
	svc, taskRepo, _ := setupTaskService()
	ctx := context.Background()

	existing := &model.Task{ID: "t1", Status: model.TaskStatusAssigned}
	taskRepo.On("FindByID", ctx, "t1").Return(existing, nil)
	taskRepo.On("Update", ctx, mock.Anything).Return(nil)

	updated := &model.Task{ID: "t1", Title: "Updated Title"}
	err := svc.Update(ctx, adminUser(), updated)

	assert.NoError(t, err)
}

func TestTaskService_Update_AdminStartedTaskForbidden(t *testing.T) {
	svc, taskRepo, _ := setupTaskService()
	ctx := context.Background()

	existing := &model.Task{ID: "t1", Status: model.TaskStatusStarted}
	taskRepo.On("FindByID", ctx, "t1").Return(existing, nil)

	err := svc.Update(ctx, adminUser(), &model.Task{ID: "t1"})

	assert.ErrorIs(t, err, apperror.ErrTaskNotMutable)
}

func TestTaskService_Update_NonAdminForbidden(t *testing.T) {
	svc, _, _ := setupTaskService()
	ctx := context.Background()

	err := svc.Update(ctx, executorUser(), &model.Task{ID: "t1"})

	assert.ErrorIs(t, err, apperror.ErrForbidden)
}

// ==================== ADMIN: Delete Task ====================

func TestTaskService_Delete_AdminAssignedSuccess(t *testing.T) {
	svc, taskRepo, _ := setupTaskService()
	ctx := context.Background()

	existing := &model.Task{ID: "t1", Status: model.TaskStatusAssigned}
	taskRepo.On("FindByID", ctx, "t1").Return(existing, nil)
	taskRepo.On("Delete", ctx, "t1").Return(nil)

	err := svc.Delete(ctx, adminUser(), "t1")

	assert.NoError(t, err)
}

func TestTaskService_Delete_AdminNonAssignedForbidden(t *testing.T) {
	svc, taskRepo, _ := setupTaskService()
	ctx := context.Background()

	existing := &model.Task{ID: "t1", Status: model.TaskStatusOnHold}
	taskRepo.On("FindByID", ctx, "t1").Return(existing, nil)

	err := svc.Delete(ctx, adminUser(), "t1")

	assert.ErrorIs(t, err, apperror.ErrTaskNotMutable)
}

func TestTaskService_Delete_TaskNotFound(t *testing.T) {
	svc, taskRepo, _ := setupTaskService()
	ctx := context.Background()

	taskRepo.On("FindByID", ctx, "missing").Return(nil, apperror.ErrTaskNotFound)

	err := svc.Delete(ctx, adminUser(), "missing")

	assert.ErrorIs(t, err, apperror.ErrTaskNotFound)
}

// ==================== EXECUTOR: ListMyTasks ====================

func TestTaskService_ListMyTasks_ExecutorSuccess(t *testing.T) {
	svc, taskRepo, _ := setupTaskService()
	ctx := context.Background()

	tasks := []*model.Task{
		{ID: "t1", AssigneeID: "exec-1"},
		{ID: "t2", AssigneeID: "exec-1"},
	}
	taskRepo.On("FindByAssignee", ctx, "exec-1").Return(tasks, nil)

	result, err := svc.ListMyTasks(ctx, executorUser())

	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestTaskService_ListMyTasks_NonExecutorForbidden(t *testing.T) {
	svc, _, _ := setupTaskService()
	ctx := context.Background()

	result, err := svc.ListMyTasks(ctx, adminUser())

	assert.Nil(t, result)
	assert.ErrorIs(t, err, apperror.ErrForbidden)
}

// ==================== EXECUTOR: ChangeStatus ====================

func TestTaskService_ChangeStatus_AssignedToStarted(t *testing.T) {
	svc, taskRepo, _ := setupTaskService()
	ctx := context.Background()

	task := &model.Task{
		ID:         "t1",
		Status:     model.TaskStatusAssigned,
		AssigneeID: "exec-1",
		DueDate:    futureDate(),
	}
	taskRepo.On("FindByID", ctx, "t1").Return(task, nil)
	taskRepo.On("Update", ctx, mock.MatchedBy(func(t *model.Task) bool {
		return t.Status == model.TaskStatusStarted
	})).Return(nil)

	err := svc.ChangeStatus(ctx, executorUser(), "t1", model.TaskStatusStarted)

	assert.NoError(t, err)
}

func TestTaskService_ChangeStatus_InvalidTransition(t *testing.T) {
	svc, taskRepo, _ := setupTaskService()
	ctx := context.Background()

	task := &model.Task{
		ID:         "t1",
		Status:     model.TaskStatusAssigned,
		AssigneeID: "exec-1",
		DueDate:    futureDate(),
	}
	taskRepo.On("FindByID", ctx, "t1").Return(task, nil)

	err := svc.ChangeStatus(ctx, executorUser(), "t1", model.TaskStatusFinishedSuccess)

	assert.ErrorIs(t, err, apperror.ErrInvalidTransition)
}

func TestTaskService_ChangeStatus_ExpiredTask(t *testing.T) {
	svc, taskRepo, _ := setupTaskService()
	ctx := context.Background()

	task := &model.Task{
		ID:         "t1",
		Status:     model.TaskStatusAssigned,
		AssigneeID: "exec-1",
		DueDate:    pastDate(),
	}
	taskRepo.On("FindByID", ctx, "t1").Return(task, nil)

	err := svc.ChangeStatus(ctx, executorUser(), "t1", model.TaskStatusStarted)

	assert.ErrorIs(t, err, apperror.ErrTaskExpired)
}

func TestTaskService_ChangeStatus_NotOwnTask(t *testing.T) {
	svc, taskRepo, _ := setupTaskService()
	ctx := context.Background()

	task := &model.Task{
		ID:         "t1",
		Status:     model.TaskStatusAssigned,
		AssigneeID: "other-executor",
		DueDate:    futureDate(),
	}
	taskRepo.On("FindByID", ctx, "t1").Return(task, nil)

	err := svc.ChangeStatus(ctx, executorUser(), "t1", model.TaskStatusStarted)

	assert.ErrorIs(t, err, apperror.ErrForbidden)
}

func TestTaskService_ChangeStatus_NonExecutorForbidden(t *testing.T) {
	svc, _, _ := setupTaskService()
	ctx := context.Background()

	err := svc.ChangeStatus(ctx, adminUser(), "t1", model.TaskStatusStarted)

	assert.ErrorIs(t, err, apperror.ErrForbidden)
}

func TestTaskService_ChangeStatus_OnHoldToStartedCycle(t *testing.T) {
	svc, taskRepo, _ := setupTaskService()
	ctx := context.Background()

	task := &model.Task{
		ID:         "t1",
		Status:     model.TaskStatusOnHold,
		AssigneeID: "exec-1",
		DueDate:    futureDate(),
	}
	taskRepo.On("FindByID", ctx, "t1").Return(task, nil)
	taskRepo.On("Update", ctx, mock.Anything).Return(nil)

	err := svc.ChangeStatus(ctx, executorUser(), "t1", model.TaskStatusFinishedError)

	assert.NoError(t, err)
}
func TestTaskService_ChangeStatus_OnHoldTransitions(t *testing.T) {
	tests := []struct {
		name        string
		target      model.TaskStatus
	}{
		{
			name:   "OnHoldToFinishedSuccess",
			target: model.TaskStatusFinishedSuccess,
		},
		{
			name:   "OnHoldToFinishedError",
			target: model.TaskStatusFinishedError,
		},
		{
			name:   "OnHoldToOnHoldAgain",
			target: model.TaskStatusOnHold,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, taskRepo, _ := setupTaskService()
			ctx := context.Background()
			task := &model.Task{
				ID:         "t1",
				Status:     model.TaskStatusOnHold,
				AssigneeID: "exec-1",
				DueDate:    futureDate(),
			}
			taskRepo.On("FindByID", ctx, "t1").Return(task, nil)
			taskRepo.On("Update", ctx, mock.MatchedBy(func(updated *model.Task) bool {
				return updated.Status == tt.target
			})).Return(nil)
			err := svc.ChangeStatus(ctx, executorUser(), "t1", tt.target)
			assert.NoError(t, err)
			taskRepo.AssertExpectations(t)
		})
	}
}

// ==================== EXECUTOR: AddComment ====================

func TestTaskService_AddComment_ExpiredTaskSuccess(t *testing.T) {
	svc, taskRepo, _ := setupTaskService()
	ctx := context.Background()

	task := &model.Task{
		ID:         "t1",
		AssigneeID: "exec-1",
		DueDate:    pastDate(),
	}
	taskRepo.On("FindByID", ctx, "t1").Return(task, nil)
	taskRepo.On("CreateComment", ctx, mock.MatchedBy(func(c *model.Comment) bool {
		return c.TaskID == "t1" && c.AuthorID == "exec-1" && c.Body == "Could not complete in time"
	})).Return(nil)

	comment, err := svc.AddComment(ctx, executorUser(), "t1", "Could not complete in time")

	require.NoError(t, err)
	assert.Equal(t, "t1", comment.TaskID)
	assert.NotEmpty(t, comment.ID)
}

func TestTaskService_AddComment_NonExpiredTaskFails(t *testing.T) {
	svc, taskRepo, _ := setupTaskService()
	ctx := context.Background()

	task := &model.Task{
		ID:         "t1",
		AssigneeID: "exec-1",
		DueDate:    futureDate(),
	}
	taskRepo.On("FindByID", ctx, "t1").Return(task, nil)

	comment, err := svc.AddComment(ctx, executorUser(), "t1", "Some comment")

	assert.Nil(t, comment)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestTaskService_AddComment_NotOwnTask(t *testing.T) {
	svc, taskRepo, _ := setupTaskService()
	ctx := context.Background()

	task := &model.Task{ID: "t1", AssigneeID: "other", DueDate: pastDate()}
	taskRepo.On("FindByID", ctx, "t1").Return(task, nil)

	comment, err := svc.AddComment(ctx, executorUser(), "t1", "Comment")

	assert.Nil(t, comment)
	assert.ErrorIs(t, err, apperror.ErrForbidden)
}

func TestTaskService_AddComment_NonExecutorForbidden(t *testing.T) {
	svc, _, _ := setupTaskService()
	ctx := context.Background()

	comment, err := svc.AddComment(ctx, auditorUser(), "t1", "Comment")

	assert.Nil(t, comment)
	assert.ErrorIs(t, err, apperror.ErrForbidden)
}

// ==================== AUDITOR: ListAllTasks ====================

func TestTaskService_ListAllTasks_AuditorSuccess(t *testing.T) {
	svc, taskRepo, _ := setupTaskService()
	ctx := context.Background()

	tasks := []*model.Task{{ID: "t1"}, {ID: "t2"}, {ID: "t3"}}
	taskRepo.On("FindAll", ctx).Return(tasks, nil)

	result, err := svc.ListAllTasks(ctx, auditorUser())

	require.NoError(t, err)
	assert.Len(t, result, 3)
}

func TestTaskService_ListAllTasks_ExecutorForbidden(t *testing.T) {
	svc, _, _ := setupTaskService()
	ctx := context.Background()

	result, err := svc.ListAllTasks(ctx, executorUser())

	assert.Nil(t, result)
	assert.ErrorIs(t, err, apperror.ErrForbidden)
}

func TestTaskService_ListAllTasks_AdminForbidden(t *testing.T) {
	svc, _, _ := setupTaskService()
	ctx := context.Background()

	result, err := svc.ListAllTasks(ctx, adminUser())

	assert.Nil(t, result)
	assert.ErrorIs(t, err, apperror.ErrForbidden)
}
