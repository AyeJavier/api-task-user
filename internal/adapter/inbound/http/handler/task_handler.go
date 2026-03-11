package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/javier/api-task-user/internal/adapter/inbound/http/dto"
	"github.com/javier/api-task-user/internal/adapter/inbound/http/middleware"
	"github.com/javier/api-task-user/internal/domain/model"
	"github.com/javier/api-task-user/internal/domain/service"
	"github.com/javier/api-task-user/pkg/validator"
)

// TaskHandler handles task management endpoints.
type TaskHandler struct {
	taskSvc *service.TaskService
}

// NewTaskHandler creates a new TaskHandler.
func NewTaskHandler(taskSvc *service.TaskService) *TaskHandler {
	return &TaskHandler{taskSvc: taskSvc}
}

// Create godoc
// @Summary     Create and assign a task
// @Tags        tasks
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body dto.CreateTaskRequest true "Task data"
// @Success     201 {object} dto.TaskResponse
// @Failure     400,403 {object} dto.ErrorResponse
// @Router      /tasks [post]
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	requester := h.requesterFromContext(r)

	var req dto.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	if errs := validator.Validate(req); errs != nil {
		respondError(w, http.StatusBadRequest, "validation failed", errs)
		return
	}

	task, err := h.taskSvc.Create(r.Context(), requester, service.CreateTaskInput{
		Title:       req.Title,
		Description: req.Description,
		AssigneeID:  req.AssigneeID,
		DueDate:     req.DueDate,
	})
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	respondJSON(w, http.StatusCreated, dto.ToTaskResponse(task))
}

// ListMine godoc
// @Summary     List my assigned tasks
// @Tags        tasks
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} dto.TaskResponse
// @Failure     403 {object} dto.ErrorResponse
// @Router      /tasks/me [get]
func (h *TaskHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	requester := h.requesterFromContext(r)

	tasks, err := h.taskSvc.ListMyTasks(r.Context(), requester)
	if err != nil {
		respondError(w, http.StatusForbidden, err.Error(), nil)
		return
	}
	respondJSON(w, http.StatusOK, tasksToResponse(tasks))
}

// ListAll godoc
// @Summary     List all tasks (Auditor only)
// @Tags        tasks
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} dto.TaskResponse
// @Failure     403 {object} dto.ErrorResponse
// @Router      /tasks [get]
func (h *TaskHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	requester := h.requesterFromContext(r)

	tasks, err := h.taskSvc.ListAllTasks(r.Context(), requester)
	if err != nil {
		respondError(w, http.StatusForbidden, err.Error(), nil)
		return
	}
	respondJSON(w, http.StatusOK, tasksToResponse(tasks))
}

// ChangeStatus godoc
// @Summary     Update task status
// @Tags        tasks
// @Security    BearerAuth
// @Accept      json
// @Param       id   path string                    true "Task ID"
// @Param       body body dto.ChangeTaskStatusRequest true "New status"
// @Success     204
// @Failure     400,403,404 {object} dto.ErrorResponse
// @Router      /tasks/{id}/status [patch]
func (h *TaskHandler) ChangeStatus(w http.ResponseWriter, r *http.Request) {
	requester := h.requesterFromContext(r)
	id := chi.URLParam(r, "id")

	var req dto.ChangeTaskStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	if errs := validator.Validate(req); errs != nil {
		respondError(w, http.StatusBadRequest, "validation failed", errs)
		return
	}

	if err := h.taskSvc.ChangeStatus(r.Context(), requester, id, req.Status); err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// AddComment godoc
// @Summary     Add comment to an expired task
// @Tags        tasks
// @Security    BearerAuth
// @Accept      json
// @Param       id   path string              true "Task ID"
// @Param       body body dto.AddCommentRequest true "Comment body"
// @Success     201 {object} dto.CommentResponse
// @Failure     400,403,404 {object} dto.ErrorResponse
// @Router      /tasks/{id}/comments [post]
func (h *TaskHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	requester := h.requesterFromContext(r)
	id := chi.URLParam(r, "id")

	var req dto.AddCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	if errs := validator.Validate(req); errs != nil {
		respondError(w, http.StatusBadRequest, "validation failed", errs)
		return
	}

	comment, err := h.taskSvc.AddComment(r.Context(), requester, id, req.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	respondJSON(w, http.StatusCreated, dto.ToCommentResponse(comment))
}

// Delete godoc
// @Summary     Delete a task (Admin only, must be ASSIGNED)
// @Tags        tasks
// @Security    BearerAuth
// @Param       id path string true "Task ID"
// @Success     204
// @Failure     400,403,404 {object} dto.ErrorResponse
// @Router      /tasks/{id} [delete]
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	requester := h.requesterFromContext(r)
	id := chi.URLParam(r, "id")

	if err := h.taskSvc.Delete(r.Context(), requester, id); err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) requesterFromContext(r *http.Request) *model.User {
	claims := middleware.ClaimsFromContext(r.Context())
	return &model.User{ID: claims.UserID, Profile: claims.Profile}
}

func tasksToResponse(tasks []*model.Task) []dto.TaskResponse {
	resp := make([]dto.TaskResponse, len(tasks))
	for i, t := range tasks {
		resp[i] = dto.ToTaskResponse(t)
	}
	return resp
}
