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

// UserHandler handles user management endpoints.
type UserHandler struct {
	userSvc  *service.UserService
	userRepo interface {
		FindByID(ctx interface{}, id string) (*model.User, error)
	}
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userSvc *service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

// Create godoc
// @Summary     Create a new user
// @Tags        users
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body dto.CreateUserRequest true "User data"
// @Success     201 {object} dto.UserCreatedResponse
// @Failure     400,403 {object} dto.ErrorResponse
// @Router      /users [post]
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	requester := h.requesterFromContext(r)

	var req dto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	if errs := validator.Validate(req); errs != nil {
		respondError(w, http.StatusBadRequest, "validation failed", errs)
		return
	}

	user, err := h.userSvc.Create(r.Context(), requester, service.CreateUserInput{
		Name:    req.Name,
		Email:   req.Email,
		Profile: req.Profile,
	})
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	respondJSON(w, http.StatusCreated, dto.UserCreatedResponse{
		UserResponse:      dto.ToUserResponse(user),
		TemporaryPassword: user.PasswordHash, // plain-text temp password returned once
	})
}

// List godoc
// @Summary     List all users
// @Tags        users
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} dto.UserResponse
// @Failure     403 {object} dto.ErrorResponse
// @Router      /users [get]
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	requester := h.requesterFromContext(r)

	users, err := h.userSvc.ListAll(r.Context(), requester)
	if err != nil {
		respondError(w, http.StatusForbidden, err.Error(), nil)
		return
	}

	resp := make([]dto.UserResponse, len(users))
	for i, u := range users {
		resp[i] = dto.ToUserResponse(u)
	}
	respondJSON(w, http.StatusOK, resp)
}

// GetByID godoc
// @Summary     Get user by ID
// @Tags        users
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "User ID"
// @Success     200 {object} dto.UserResponse
// @Failure     403,404 {object} dto.ErrorResponse
// @Router      /users/{id} [get]
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	requester := h.requesterFromContext(r)
	id := chi.URLParam(r, "id")

	user, err := h.userSvc.GetByID(r.Context(), requester, id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error(), nil)
		return
	}
	respondJSON(w, http.StatusOK, dto.ToUserResponse(user))
}

// Delete godoc
// @Summary     Delete a user
// @Tags        users
// @Security    BearerAuth
// @Param       id path string true "User ID"
// @Success     204
// @Failure     403,404 {object} dto.ErrorResponse
// @Router      /users/{id} [delete]
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	requester := h.requesterFromContext(r)
	id := chi.URLParam(r, "id")

	if err := h.userSvc.Delete(r.Context(), requester, id); err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) requesterFromContext(r *http.Request) *model.User {
	claims := middleware.ClaimsFromContext(r.Context())
	return &model.User{ID: claims.UserID, Profile: claims.Profile}
}
