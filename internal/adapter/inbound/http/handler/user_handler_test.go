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

func newUserHandler() (*handler.UserHandler, *mocks.MockUserRepository, *mocks.MockHashPort) {
	userRepo := new(mocks.MockUserRepository)
	hashPort := new(mocks.MockHashPort)
	svc := service.NewUserService(userRepo, hashPort)
	h := handler.NewUserHandler(svc)
	return h, userRepo, hashPort
}

// userWithProfile builds a model.User with the given profile string.
func userWithProfile(profile string) *model.User {
	return &model.User{
		ID:        "some-id",
		Name:      "Test User",
		Email:     "test@example.com",
		Profile:   model.Profile(profile),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// ---------- Create ----------

func TestUserHandler_Create_AdminCreatesExecutor(t *testing.T) {
	h, userRepo, hashPort := newUserHandler()
	ap := buildAuthPort()

	userRepo.On("FindByEmail", mock.Anything, "executor@test.com").Return(nil, apperror.ErrUserNotFound)
	hashPort.On("Hash", mock.AnythingOfType("string")).Return("hashed", nil)
	userRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *model.User) bool {
		return u.Email == "executor@test.com" && u.Profile == model.ProfileExecutor
	})).Return(nil)

	w := httptest.NewRecorder()
	r := newReq(http.MethodPost, "/users", tokenAdmin, jsonBody(
		`{"name":"New Executor","email":"executor@test.com","profile":"EXECUTOR"}`,
	))
	withAuth(ap, http.HandlerFunc(h.Create)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "executor@test.com")
}

func TestUserHandler_Create_AdminCannotCreateAdmin(t *testing.T) {
	h, _, _ := newUserHandler()
	ap := buildAuthPort()

	w := httptest.NewRecorder()
	r := newReq(http.MethodPost, "/users", tokenAdmin, jsonBody(
		`{"name":"Bad Admin","email":"admin2@test.com","profile":"ADMIN"}`,
	))
	withAuth(ap, http.HandlerFunc(h.Create)).ServeHTTP(w, r)

	// profile validation rejects ADMIN at dto level (oneof=EXECUTOR AUDITOR)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Create_InvalidBody(t *testing.T) {
	h, _, _ := newUserHandler()
	ap := buildAuthPort()

	w := httptest.NewRecorder()
	r := newReq(http.MethodPost, "/users", tokenAdmin, jsonBody(`not json`))
	withAuth(ap, http.HandlerFunc(h.Create)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Create_DuplicateEmail(t *testing.T) {
	h, userRepo, hashPort := newUserHandler()
	ap := buildAuthPort()

	existing := userWithProfile("EXECUTOR")
	existing.Email = "dup@test.com"
	userRepo.On("FindByEmail", mock.Anything, "dup@test.com").Return(existing, nil)
	hashPort.On("Hash", mock.Anything).Return("h", nil)

	w := httptest.NewRecorder()
	r := newReq(http.MethodPost, "/users", tokenAdmin, jsonBody(
		`{"name":"Dup User","email":"dup@test.com","profile":"EXECUTOR"}`,
	))
	withAuth(ap, http.HandlerFunc(h.Create)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "email")
}

// ---------- List ----------

func TestUserHandler_List_AdminSuccess(t *testing.T) {
	h, userRepo, _ := newUserHandler()
	ap := buildAuthPort()

	users := []*model.User{userWithProfile("EXECUTOR"), userWithProfile("AUDITOR")}
	userRepo.On("FindAll", mock.Anything).Return(users, nil)

	w := httptest.NewRecorder()
	r := newReq(http.MethodGet, "/users", tokenAdmin, nil)
	withAuth(ap, http.HandlerFunc(h.List)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandler_List_ExecutorForbidden(t *testing.T) {
	h, _, _ := newUserHandler()
	ap := buildAuthPort()

	w := httptest.NewRecorder()
	r := newReq(http.MethodGet, "/users", tokenExecutor, nil)
	withAuth(ap, http.HandlerFunc(h.List)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ---------- GetByID ----------

func TestUserHandler_GetByID_AdminSuccess(t *testing.T) {
	h, userRepo, _ := newUserHandler()
	ap := buildAuthPort()

	user := userWithProfile("EXECUTOR")
	user.ID = "u-target"
	userRepo.On("FindByID", mock.Anything, "u-target").Return(user, nil)

	w := httptest.NewRecorder()
	r := newReq(http.MethodGet, "/users/u-target", tokenAdmin, nil)
	r = addChiParam(r, "id", "u-target")
	withAuth(ap, http.HandlerFunc(h.GetByID)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandler_GetByID_NotFound(t *testing.T) {
	h, userRepo, _ := newUserHandler()
	ap := buildAuthPort()

	userRepo.On("FindByID", mock.Anything, "missing").Return(nil, apperror.ErrUserNotFound)

	w := httptest.NewRecorder()
	r := newReq(http.MethodGet, "/users/missing", tokenAdmin, nil)
	r = addChiParam(r, "id", "missing")
	withAuth(ap, http.HandlerFunc(h.GetByID)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------- Delete ----------

func TestUserHandler_Delete_AdminSuccess(t *testing.T) {
	h, userRepo, _ := newUserHandler()
	ap := buildAuthPort()

	userRepo.On("Delete", mock.Anything, "u-del").Return(nil)

	w := httptest.NewRecorder()
	r := newReq(http.MethodDelete, "/users/u-del", tokenAdmin, nil)
	r = addChiParam(r, "id", "u-del")
	withAuth(ap, http.HandlerFunc(h.Delete)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestUserHandler_Delete_NonAdminForbidden(t *testing.T) {
	h, _, _ := newUserHandler()
	ap := buildAuthPort()

	w := httptest.NewRecorder()
	r := newReq(http.MethodDelete, "/users/u-del", tokenExecutor, nil)
	r = addChiParam(r, "id", "u-del")
	withAuth(ap, http.HandlerFunc(h.Delete)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
