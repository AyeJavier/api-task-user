package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/javier/api-task-user/internal/adapter/inbound/http/handler"
	"github.com/javier/api-task-user/internal/domain/port/mocks"
	"github.com/javier/api-task-user/internal/domain/service"
	"github.com/javier/api-task-user/pkg/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newAuthHandler() (*handler.AuthHandler, *mocks.MockUserRepository, *mocks.MockAuthPort, *mocks.MockHashPort) {
	userRepo := new(mocks.MockUserRepository)
	authPort := new(mocks.MockAuthPort)
	hashPort := new(mocks.MockHashPort)
	svc := service.NewAuthService(userRepo, authPort, hashPort)
	h := handler.NewAuthHandler(svc)
	return h, userRepo, authPort, hashPort
}

// ---------- Login ----------

func TestAuthHandler_Login_Success(t *testing.T) {
	h, userRepo, authPort, hashPort := newAuthHandler()

	user := userWithProfile("ADMIN")
	user.PasswordHash = "hashed"
	userRepo.On("FindByEmail", mock.Anything, "admin@test.com").Return(user, nil)
	hashPort.On("Compare", "hashed", "password123").Return(nil)
	authPort.On("GenerateToken", mock.Anything, user).Return("jwt-abc", nil)

	// ap := buildAuthPort()
	// w := httptest.NewRecorder()
	// r := newReq(http.MethodPost, "/auth/login", "", jsonBody(`{"email":"admin@test.com","password":"password123"}`))
	// //withAuth(ap, http.HandlerFunc(h.Login)).ServeHTTP(w, r)

	// assert.Equal(t, http.StatusOK, w.Code)
	// assert.Contains(t, w.Body.String(), "jwt-abc")
	w := httptest.NewRecorder()
	r := newReq(http.MethodPost, "/auth/login", "", jsonBody(`{"email":"admin@test.com","password":"password123"}`))
	http.HandlerFunc(h.Login).ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "jwt-abc")
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	h, userRepo, _, _ := newAuthHandler()

	userRepo.On("FindByEmail", mock.Anything, "unknown@test.com").Return(nil, apperror.ErrUserNotFound)

	w := httptest.NewRecorder()
	r := newReq(http.MethodPost, "/auth/login", "", jsonBody(`{"email":"unknown@test.com","password":"password123"}`))
	http.HandlerFunc(h.Login).ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_Login_MissingFields(t *testing.T) {
	h, _, _, _ := newAuthHandler()

	w := httptest.NewRecorder()
	r := newReq(http.MethodPost, "/auth/login", "", jsonBody(`{"email":"notanemail","password":"short"}`))
	http.HandlerFunc(h.Login).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	h, _, _, _ := newAuthHandler()

	w := httptest.NewRecorder()
	r := newReq(http.MethodPost, "/auth/login", "", jsonBody(`not json`))
	http.HandlerFunc(h.Login).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------- Logout ----------

func TestAuthHandler_Logout_Success(t *testing.T) {
	h, _, authPort, _ := newAuthHandler()
	ap := buildAuthPort()

	authPort.On("InvalidateToken", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	w := httptest.NewRecorder()
	r := newReq(http.MethodPost, "/auth/logout", tokenAdmin, nil)
	withAuth(ap, http.HandlerFunc(h.Logout)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

// ---------- ChangePassword ----------

func TestAuthHandler_ChangePassword_Success(t *testing.T) {
	h, userRepo, _, hashPort := newAuthHandler()
	ap := buildAuthPort()

	user := userWithProfile("ADMIN")
	user.ID = "admin-id"
	user.PasswordHash = "old-hash"
	userRepo.On("FindByID", mock.Anything, mock.AnythingOfType("string")).Return(user, nil)
	hashPort.On("Compare", "old-hash", "OldPass123").Return(nil)
	hashPort.On("Hash", "NewPass456!").Return("new-hash", nil)
	userRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	r := newReq(http.MethodPut, "/auth/password", tokenAdmin,
		jsonBody(`{"current_password":"OldPass123","new_password":"NewPass456!"}`))
	withAuth(ap, http.HandlerFunc(h.ChangePassword)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestAuthHandler_ChangePassword_ValidationFails(t *testing.T) {
	h, _, _, _ := newAuthHandler()
	ap := buildAuthPort()

	w := httptest.NewRecorder()
	// new_password is too short (< 8 chars)
	r := newReq(http.MethodPut, "/auth/password", tokenAdmin,
		jsonBody(`{"current_password":"OldPass123","new_password":"short"}`))
	withAuth(ap, http.HandlerFunc(h.ChangePassword)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
