package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/javier/api-task-user/internal/adapter/inbound/http/handler"
	"github.com/javier/api-task-user/internal/domain/model"
	"github.com/javier/api-task-user/internal/domain/port/mocks"
	"github.com/javier/api-task-user/internal/domain/service"
	"github.com/javier/api-task-user/pkg/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTaskHandler() (*handler.TaskHandler, *mocks.MockTaskRepository, *mocks.MockUserRepository) {
	taskRepo := new(mocks.MockTaskRepository)
	userRepo := new(mocks.MockUserRepository)
	svc := service.NewTaskService(taskRepo, userRepo)
	h := handler.NewTaskHandler(svc)
	return h, taskRepo, userRepo
}

func executorModel() *model.User {
	return &model.User{ID: "exec-id", Profile: model.ProfileExecutor}
}

func futureTask(assigneeID string, status model.TaskStatus) *model.Task {
	return &model.Task{
		ID:         "task-1",
		Title:      "Test Task",
		Status:     status,
		AssigneeID: assigneeID,
		DueDate:    time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func pastTask(assigneeID string) *model.Task {
	return &model.Task{
		ID:         "task-1",
		Title:      "Expired Task",
		Status:     model.TaskStatusAssigned,
		AssigneeID: assigneeID,
		DueDate:    time.Now().Add(-24 * time.Hour),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// ---------- Create ----------

func TestTaskHandler_Create_AdminSuccess(t *testing.T) {
	h, taskRepo, userRepo := newTaskHandler()
	ap := buildAuthPort()

	assigneeUUID := "6ba7b810-9dad-41d1-80b4-00c04fd430c8"
	userRepo.On("FindByID", mock.Anything, assigneeUUID).Return(executorModel(), nil)
	taskRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	body := `{"title":"Deploy v2","description":"Deploy to prod","assignee_id":"` + assigneeUUID + `","due_date":"2099-01-01T00:00:00Z"}`

	w := httptest.NewRecorder()
	r := newReq(http.MethodPost, "/tasks", tokenAdmin, jsonBody(body))
	withAuth(ap, http.HandlerFunc(h.Create)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "Deploy v2")
}

func TestTaskHandler_Create_ValidationError(t *testing.T) {
	h, _, _ := newTaskHandler()
	ap := buildAuthPort()

	w := httptest.NewRecorder()
	// missing required fields
	r := newReq(http.MethodPost, "/tasks", tokenAdmin, jsonBody(`{"title":"T"}`))
	withAuth(ap, http.HandlerFunc(h.Create)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_Create_NonAdminForbidden(t *testing.T) {
	h, _, _ := newTaskHandler()
	ap := buildAuthPort()

	assigneeUUID := "6ba7b810-9dad-41d1-80b4-00c04fd430c8"
	body := `{"title":"Deploy v2","description":"Deploy to prod","assignee_id":"` + assigneeUUID + `","due_date":"2099-01-01T00:00:00Z"}`

	w := httptest.NewRecorder()
	r := newReq(http.MethodPost, "/tasks", tokenExecutor, jsonBody(body))
	withAuth(ap, http.HandlerFunc(h.Create)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), apperror.ErrForbidden.Error())
}

// ---------- ListMine ----------

func TestTaskHandler_ListMine_ExecutorSuccess(t *testing.T) {
	h, taskRepo, _ := newTaskHandler()
	ap := buildAuthPort()

	tasks := []*model.Task{
		futureTask("exec-id", model.TaskStatusAssigned),
		futureTask("exec-id", model.TaskStatusStarted),
	}
	taskRepo.On("FindByAssignee", mock.Anything, "exec-id").Return(tasks, nil)

	w := httptest.NewRecorder()
	r := newReq(http.MethodGet, "/tasks/me", tokenExecutor, nil)
	withAuth(ap, http.HandlerFunc(h.ListMine)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTaskHandler_ListMine_AdminForbidden(t *testing.T) {
	h, _, _ := newTaskHandler()
	ap := buildAuthPort()

	w := httptest.NewRecorder()
	r := newReq(http.MethodGet, "/tasks/me", tokenAdmin, nil)
	withAuth(ap, http.HandlerFunc(h.ListMine)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ---------- ListAll ----------

func TestTaskHandler_ListAll_AuditorSuccess(t *testing.T) {
	h, taskRepo, _ := newTaskHandler()
	ap := buildAuthPort()

	tasks := []*model.Task{
		futureTask("exec-1", model.TaskStatusAssigned),
		futureTask("exec-2", model.TaskStatusStarted),
	}
	taskRepo.On("FindAll", mock.Anything).Return(tasks, nil)

	w := httptest.NewRecorder()
	r := newReq(http.MethodGet, "/tasks", tokenAuditor, nil)
	withAuth(ap, http.HandlerFunc(h.ListAll)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTaskHandler_ListAll_ExecutorForbidden(t *testing.T) {
	h, _, _ := newTaskHandler()
	ap := buildAuthPort()

	w := httptest.NewRecorder()
	r := newReq(http.MethodGet, "/tasks", tokenExecutor, nil)
	withAuth(ap, http.HandlerFunc(h.ListAll)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ---------- ChangeStatus ----------

func TestTaskHandler_ChangeStatus_ExecutorSuccess(t *testing.T) {
	h, taskRepo, _ := newTaskHandler()
	ap := buildAuthPort()

	task := futureTask("exec-id", model.TaskStatusAssigned)
	taskRepo.On("FindByID", mock.Anything, "task-1").Return(task, nil)
	taskRepo.On("Update", mock.Anything, mock.MatchedBy(func(t *model.Task) bool {
		return t.Status == model.TaskStatusStarted
	})).Return(nil)

	w := httptest.NewRecorder()
	r := newReq(http.MethodPatch, "/tasks/task-1/status", tokenExecutor, jsonBody(`{"status":"STARTED"}`))
	r = addChiParam(r, "id", "task-1")
	withAuth(ap, http.HandlerFunc(h.ChangeStatus)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestTaskHandler_ChangeStatus_InvalidTransition(t *testing.T) {
	h, taskRepo, _ := newTaskHandler()
	ap := buildAuthPort()

	task := futureTask("exec-id", model.TaskStatusAssigned)
	taskRepo.On("FindByID", mock.Anything, "task-1").Return(task, nil)

	w := httptest.NewRecorder()
	r := newReq(http.MethodPatch, "/tasks/task-1/status", tokenExecutor, jsonBody(`{"status":"FINISHED_SUCCESS"}`))
	r = addChiParam(r, "id", "task-1")
	withAuth(ap, http.HandlerFunc(h.ChangeStatus)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_ChangeStatus_ExpiredTask(t *testing.T) {
	h, taskRepo, _ := newTaskHandler()
	ap := buildAuthPort()

	task := pastTask("exec-id")
	taskRepo.On("FindByID", mock.Anything, "task-1").Return(task, nil)

	w := httptest.NewRecorder()
	r := newReq(http.MethodPatch, "/tasks/task-1/status", tokenExecutor, jsonBody(`{"status":"STARTED"}`))
	r = addChiParam(r, "id", "task-1")
	withAuth(ap, http.HandlerFunc(h.ChangeStatus)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "expired")
}

func TestTaskHandler_ChangeStatus_InvalidStatusValue(t *testing.T) {
	h, _, _ := newTaskHandler()
	ap := buildAuthPort()

	w := httptest.NewRecorder()
	r := newReq(http.MethodPatch, "/tasks/task-1/status", tokenExecutor, jsonBody(`{"status":"INVALID_STATUS"}`))
	r = addChiParam(r, "id", "task-1")
	withAuth(ap, http.HandlerFunc(h.ChangeStatus)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------- AddComment ----------

func TestTaskHandler_AddComment_ExpiredTaskSuccess(t *testing.T) {
	h, taskRepo, _ := newTaskHandler()
	ap := buildAuthPort()

	task := pastTask("exec-id")
	taskRepo.On("FindByID", mock.Anything, "task-1").Return(task, nil)
	taskRepo.On("CreateComment", mock.Anything, mock.MatchedBy(func(c *model.Comment) bool {
		return c.TaskID == "task-1" && c.Body == "API was down during due date"
	})).Return(nil)

	w := httptest.NewRecorder()
	r := newReq(http.MethodPost, "/tasks/task-1/comments", tokenExecutor,
		jsonBody(`{"body":"API was down during due date"}`))
	r = addChiParam(r, "id", "task-1")
	withAuth(ap, http.HandlerFunc(h.AddComment)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestTaskHandler_AddComment_NonExpiredFails(t *testing.T) {
	h, taskRepo, _ := newTaskHandler()
	ap := buildAuthPort()

	task := futureTask("exec-id", model.TaskStatusStarted)
	taskRepo.On("FindByID", mock.Anything, "task-1").Return(task, nil)

	w := httptest.NewRecorder()
	r := newReq(http.MethodPost, "/tasks/task-1/comments", tokenExecutor,
		jsonBody(`{"body":"some comment"}`))
	r = addChiParam(r, "id", "task-1")
	withAuth(ap, http.HandlerFunc(h.AddComment)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_AddComment_EmptyBodyFails(t *testing.T) {
	h, _, _ := newTaskHandler()
	ap := buildAuthPort()

	w := httptest.NewRecorder()
	r := newReq(http.MethodPost, "/tasks/task-1/comments", tokenExecutor,
		jsonBody(`{"body":""}`))
	r = addChiParam(r, "id", "task-1")
	withAuth(ap, http.HandlerFunc(h.AddComment)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------- Delete ----------

func TestTaskHandler_Delete_AdminAssignedSuccess(t *testing.T) {
	h, taskRepo, _ := newTaskHandler()
	ap := buildAuthPort()

	task := futureTask("exec-id", model.TaskStatusAssigned)
	taskRepo.On("FindByID", mock.Anything, "task-1").Return(task, nil)
	taskRepo.On("Delete", mock.Anything, "task-1").Return(nil)

	w := httptest.NewRecorder()
	r := newReq(http.MethodDelete, "/tasks/task-1", tokenAdmin, nil)
	r = addChiParam(r, "id", "task-1")
	withAuth(ap, http.HandlerFunc(h.Delete)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestTaskHandler_Delete_StartedTaskFails(t *testing.T) {
	h, taskRepo, _ := newTaskHandler()
	ap := buildAuthPort()

	task := futureTask("exec-id", model.TaskStatusStarted)
	taskRepo.On("FindByID", mock.Anything, "task-1").Return(task, nil)

	w := httptest.NewRecorder()
	r := newReq(http.MethodDelete, "/tasks/task-1", tokenAdmin, nil)
	r = addChiParam(r, "id", "task-1")
	withAuth(ap, http.HandlerFunc(h.Delete)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_Delete_NotFound(t *testing.T) {
	h, taskRepo, _ := newTaskHandler()
	ap := buildAuthPort()

	taskRepo.On("FindByID", mock.Anything, "missing").Return(nil, apperror.ErrTaskNotFound)

	w := httptest.NewRecorder()
	r := newReq(http.MethodDelete, "/tasks/missing", tokenAdmin, nil)
	r = addChiParam(r, "id", "missing")
	withAuth(ap, http.HandlerFunc(h.Delete)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
