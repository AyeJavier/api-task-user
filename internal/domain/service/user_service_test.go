package service_test

import (
	"context"
	"testing"

	"github.com/javier/api-task-user/internal/domain/model"
	"github.com/javier/api-task-user/internal/domain/port/mocks"
	"github.com/javier/api-task-user/internal/domain/service"
	"github.com/javier/api-task-user/pkg/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupUserService() (*service.UserService, *mocks.MockUserRepository, *mocks.MockHashPort) {
	userRepo := new(mocks.MockUserRepository)
	hashPort := new(mocks.MockHashPort)
	svc := service.NewUserService(userRepo, hashPort)
	return svc, userRepo, hashPort
}

func adminUser() *model.User {
	return &model.User{ID: "admin-1", Name: "Admin", Profile: model.ProfileAdmin}
}

func executorUser() *model.User {
	return &model.User{ID: "exec-1", Name: "Executor", Profile: model.ProfileExecutor}
}

func auditorUser() *model.User {
	return &model.User{ID: "aud-1", Name: "Auditor", Profile: model.ProfileAuditor}
}

// ---------- Create ----------

func TestUserService_Create_ExecutorByAdmin(t *testing.T) {
	svc, userRepo, hashPort := setupUserService()
	ctx := context.Background()

	userRepo.On("FindByEmail", ctx, "new@test.com").Return(nil, apperror.ErrUserNotFound)
	hashPort.On("Hash", mock.AnythingOfType("string")).Return("hashed-temp", nil)
	userRepo.On("Create", ctx, mock.MatchedBy(func(u *model.User) bool {
		return u.Email == "new@test.com" &&
			u.Profile == model.ProfileExecutor &&
			u.MustChangePassword == true
	})).Return(nil)

	input := service.CreateUserInput{
		Name:    "New Executor",
		Email:   "new@test.com",
		Profile: model.ProfileExecutor,
	}

	user, err := svc.Create(ctx, adminUser(), input)

	require.NoError(t, err)
	assert.Equal(t, "new@test.com", user.Email)
	assert.Equal(t, model.ProfileExecutor, user.Profile)
	// The returned user should have the temp password in plain text (not hash)
	assert.NotEqual(t, "hashed-temp", user.PasswordHash)
	assert.Contains(t, user.PasswordHash, "Temp@")
}

func TestUserService_Create_AuditorByAdmin(t *testing.T) {
	svc, userRepo, hashPort := setupUserService()
	ctx := context.Background()

	userRepo.On("FindByEmail", ctx, "aud@test.com").Return(nil, apperror.ErrUserNotFound)
	hashPort.On("Hash", mock.AnythingOfType("string")).Return("hashed", nil)
	userRepo.On("Create", ctx, mock.Anything).Return(nil)

	input := service.CreateUserInput{
		Name:    "New Auditor",
		Email:   "aud@test.com",
		Profile: model.ProfileAuditor,
	}

	user, err := svc.Create(ctx, adminUser(), input)

	require.NoError(t, err)
	assert.Equal(t, model.ProfileAuditor, user.Profile)
}

func TestUserService_Create_AdminCannotCreateAdmin(t *testing.T) {
	svc, _, _ := setupUserService()
	ctx := context.Background()

	input := service.CreateUserInput{
		Name:    "Another Admin",
		Email:   "admin2@test.com",
		Profile: model.ProfileAdmin,
	}

	user, err := svc.Create(ctx, adminUser(), input)

	assert.Nil(t, user)
	assert.ErrorIs(t, err, apperror.ErrCannotCreateAdmin)
}

func TestUserService_Create_ExecutorCannotCreate(t *testing.T) {
	svc, _, _ := setupUserService()
	ctx := context.Background()

	input := service.CreateUserInput{
		Name:    "Someone",
		Email:   "s@test.com",
		Profile: model.ProfileExecutor,
	}

	user, err := svc.Create(ctx, executorUser(), input)

	assert.Nil(t, user)
	assert.ErrorIs(t, err, apperror.ErrForbidden)
}

func TestUserService_Create_DuplicateEmail(t *testing.T) {
	svc, userRepo, _ := setupUserService()
	ctx := context.Background()

	existing := &model.User{ID: "existing", Email: "dup@test.com"}
	userRepo.On("FindByEmail", ctx, "dup@test.com").Return(existing, nil)

	input := service.CreateUserInput{
		Name:    "Dup",
		Email:   "dup@test.com",
		Profile: model.ProfileExecutor,
	}

	user, err := svc.Create(ctx, adminUser(), input)

	assert.Nil(t, user)
	assert.ErrorIs(t, err, apperror.ErrEmailAlreadyExists)
}

// ---------- GetByID ----------

func TestUserService_GetByID_AdminSuccess(t *testing.T) {
	svc, userRepo, _ := setupUserService()
	ctx := context.Background()

	expected := &model.User{ID: "u1", Name: "Target"}
	userRepo.On("FindByID", ctx, "u1").Return(expected, nil)

	user, err := svc.GetByID(ctx, adminUser(), "u1")

	require.NoError(t, err)
	assert.Equal(t, "Target", user.Name)
}

func TestUserService_GetByID_NonAdminForbidden(t *testing.T) {
	svc, _, _ := setupUserService()
	ctx := context.Background()

	user, err := svc.GetByID(ctx, executorUser(), "u1")

	assert.Nil(t, user)
	assert.ErrorIs(t, err, apperror.ErrForbidden)
}

// ---------- ListAll ----------

func TestUserService_ListAll_AdminSuccess(t *testing.T) {
	svc, userRepo, _ := setupUserService()
	ctx := context.Background()

	users := []*model.User{{ID: "u1"}, {ID: "u2"}}
	userRepo.On("FindAll", ctx).Return(users, nil)

	result, err := svc.ListAll(ctx, adminUser())

	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestUserService_ListAll_ExecutorForbidden(t *testing.T) {
	svc, _, _ := setupUserService()
	ctx := context.Background()

	result, err := svc.ListAll(ctx, executorUser())

	assert.Nil(t, result)
	assert.ErrorIs(t, err, apperror.ErrForbidden)
}

// ---------- Update ----------

func TestUserService_Update_AdminSuccess(t *testing.T) {
	svc, userRepo, _ := setupUserService()
	ctx := context.Background()

	target := &model.User{ID: "u1", Name: "Updated"}
	userRepo.On("Update", ctx, mock.MatchedBy(func(u *model.User) bool {
		return u.ID == "u1" && !u.UpdatedAt.IsZero()
	})).Return(nil)

	err := svc.Update(ctx, adminUser(), target)

	assert.NoError(t, err)
}

func TestUserService_Update_NonAdminForbidden(t *testing.T) {
	svc, _, _ := setupUserService()
	ctx := context.Background()

	err := svc.Update(ctx, auditorUser(), &model.User{ID: "u1"})

	assert.ErrorIs(t, err, apperror.ErrForbidden)
}

// ---------- Delete ----------

func TestUserService_Delete_AdminSuccess(t *testing.T) {
	svc, userRepo, _ := setupUserService()
	ctx := context.Background()

	userRepo.On("Delete", ctx, "u1").Return(nil)

	err := svc.Delete(ctx, adminUser(), "u1")

	assert.NoError(t, err)
}

func TestUserService_Delete_NonAdminForbidden(t *testing.T) {
	svc, _, _ := setupUserService()
	ctx := context.Background()

	err := svc.Delete(ctx, executorUser(), "u1")

	assert.ErrorIs(t, err, apperror.ErrForbidden)
}
