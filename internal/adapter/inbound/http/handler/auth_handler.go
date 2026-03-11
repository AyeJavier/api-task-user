// Package handler contains the HTTP handlers for the inbound adapter.
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/javier/api-task-user/internal/adapter/inbound/http/dto"
	"github.com/javier/api-task-user/internal/adapter/inbound/http/middleware"
	"github.com/javier/api-task-user/internal/domain/service"
	"github.com/javier/api-task-user/pkg/validator"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	authSvc *service.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// Login godoc
// @Summary     Authenticate user
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body dto.LoginRequest true "Credentials"
// @Success     200 {object} dto.LoginResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Router      /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	if errs := validator.Validate(req); errs != nil {
		respondError(w, http.StatusBadRequest, "validation failed", errs)
		return
	}

	result, err := h.authSvc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	respondJSON(w, http.StatusOK, dto.LoginResponse{
		Token:              result.Token,
		MustChangePassword: result.MustChangePassword,
	})
}

// Logout godoc
// @Summary     Logout user
// @Tags        auth
// @Security    BearerAuth
// @Success     204
// @Failure     401 {object} dto.ErrorResponse
// @Router      /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if err := h.authSvc.Logout(r.Context(), token); err != nil {
		respondError(w, http.StatusInternalServerError, "logout failed", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ChangePassword godoc
// @Summary     Change password
// @Tags        auth
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body dto.ChangePasswordRequest true "Passwords"
// @Success     204
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Router      /auth/password [put]
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())

	var req dto.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	if errs := validator.Validate(req); errs != nil {
		respondError(w, http.StatusBadRequest, "validation failed", errs)
		return
	}

	if err := h.authSvc.ChangePassword(r.Context(), claims.UserID, req.CurrentPassword, req.NewPassword); err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
