package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/javier/api-task-user/internal/domain/model"
	"github.com/javier/api-task-user/internal/domain/port/mocks"
	"github.com/javier/api-task-user/internal/domain/service"
	"github.com/javier/api-task-user/pkg/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupAuthService() (*service.AuthService, *mocks.MockUserRepository, *mocks.MockAuthPort, *mocks.MockHashPort) {
	userRepo := new(mocks.MockUserRepository)
	authPort := new(mocks.MockAuthPort)
	hashPort := new(mocks.MockHashPort)
	svc := service.NewAuthService(userRepo, authPort, hashPort)
	return svc, userRepo, authPort, hashPort
}

// ---------- Login ----------

func TestAuthService_Login_Success(t *testing.T) {
	svc, userRepo, authPort, hashPort := setupAuthService()
	ctx := context.Background()

	user := &model.User{
		ID:                 "u1",
		Email:              "admin@test.com",
		PasswordHash:       "hashed-pw",
		Profile:            model.ProfileAdmin,
		MustChangePassword: false,
	}

	userRepo.On("FindByEmail", ctx, "admin@test.com").Return(user, nil)
	hashPort.On("Compare", "hashed-pw", "password123").Return(nil)
	authPort.On("GenerateToken", ctx, user).Return("jwt-token-123", nil)

	result, err := svc.Login(ctx, "admin@test.com", "password123")

	require.NoError(t, err)
	assert.Equal(t, "jwt-token-123", result.Token)
	assert.False(t, result.MustChangePassword)
	userRepo.AssertExpectations(t)
	hashPort.AssertExpectations(t)
	authPort.AssertExpectations(t)
}

func TestAuthService_Login_MustChangePassword(t *testing.T) {
	svc, userRepo, authPort, hashPort := setupAuthService()
	ctx := context.Background()

	user := &model.User{
		ID:                 "u1",
		Email:              "new@test.com",
		PasswordHash:       "temp-hash",
		Profile:            model.ProfileExecutor,
		MustChangePassword: true,
	}

	userRepo.On("FindByEmail", ctx, "new@test.com").Return(user, nil)
	hashPort.On("Compare", "temp-hash", "temppass").Return(nil)
	authPort.On("GenerateToken", ctx, user).Return("jwt-token", nil)

	result, err := svc.Login(ctx, "new@test.com", "temppass")

	require.NoError(t, err)
	assert.True(t, result.MustChangePassword)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	svc, userRepo, _, _ := setupAuthService()
	ctx := context.Background()

	userRepo.On("FindByEmail", ctx, "nobody@test.com").Return(nil, apperror.ErrUserNotFound)

	result, err := svc.Login(ctx, "nobody@test.com", "password123")

	assert.Nil(t, result)
	assert.ErrorIs(t, err, apperror.ErrInvalidCredentials)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	svc, userRepo, _, hashPort := setupAuthService()
	ctx := context.Background()

	user := &model.User{ID: "u1", Email: "admin@test.com", PasswordHash: "hashed"}
	userRepo.On("FindByEmail", ctx, "admin@test.com").Return(user, nil)
	hashPort.On("Compare", "hashed", "wrong-pw").Return(errors.New("mismatch"))

	result, err := svc.Login(ctx, "admin@test.com", "wrong-pw")

	assert.Nil(t, result)
	assert.ErrorIs(t, err, apperror.ErrInvalidCredentials)
}

func TestAuthService_Login_TokenGenerationFails(t *testing.T) {
	svc, userRepo, authPort, hashPort := setupAuthService()
	ctx := context.Background()

	user := &model.User{ID: "u1", Email: "a@test.com", PasswordHash: "h"}
	userRepo.On("FindByEmail", ctx, "a@test.com").Return(user, nil)
	hashPort.On("Compare", "h", "p").Return(nil)
	authPort.On("GenerateToken", ctx, user).Return("", errors.New("signing error"))

	result, err := svc.Login(ctx, "a@test.com", "p")

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "generating token")
}

// ---------- Logout ----------

func TestAuthService_Logout_Success(t *testing.T) {
	svc, _, authPort, _ := setupAuthService()
	ctx := context.Background()

	authPort.On("InvalidateToken", ctx, "valid-token").Return(nil)

	err := svc.Logout(ctx, "valid-token")

	assert.NoError(t, err)
	authPort.AssertExpectations(t)
}

func TestAuthService_Logout_InvalidationFails(t *testing.T) {
	svc, _, authPort, _ := setupAuthService()
	ctx := context.Background()

	authPort.On("InvalidateToken", ctx, "token").Return(errors.New("redis down"))

	err := svc.Logout(ctx, "token")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalidating token")
}

// ---------- ChangePassword ----------

func TestAuthService_ChangePassword_Success(t *testing.T) {
	svc, userRepo, _, hashPort := setupAuthService()
	ctx := context.Background()

	user := &model.User{
		ID:                 "u1",
		PasswordHash:       "old-hash",
		MustChangePassword: true,
	}

	userRepo.On("FindByID", ctx, "u1").Return(user, nil)
	hashPort.On("Compare", "old-hash", "oldpass").Return(nil)
	hashPort.On("Hash", "newpass123").Return("new-hash", nil)
	userRepo.On("Update", ctx, mock.MatchedBy(func(u *model.User) bool {
		return u.PasswordHash == "new-hash" && !u.MustChangePassword
	})).Return(nil)

	err := svc.ChangePassword(ctx, "u1", "oldpass", "newpass123")

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func TestAuthService_ChangePassword_UserNotFound(t *testing.T) {
	svc, userRepo, _, _ := setupAuthService()
	ctx := context.Background()

	userRepo.On("FindByID", ctx, "unknown").Return(nil, apperror.ErrUserNotFound)

	err := svc.ChangePassword(ctx, "unknown", "old", "new")

	assert.ErrorIs(t, err, apperror.ErrUserNotFound)
}

func TestAuthService_ChangePassword_WrongCurrentPassword(t *testing.T) {
	svc, userRepo, _, hashPort := setupAuthService()
	ctx := context.Background()

	user := &model.User{ID: "u1", PasswordHash: "h"}
	userRepo.On("FindByID", ctx, "u1").Return(user, nil)
	hashPort.On("Compare", "h", "wrong").Return(errors.New("mismatch"))

	err := svc.ChangePassword(ctx, "u1", "wrong", "new")

	assert.ErrorIs(t, err, apperror.ErrInvalidCredentials)
}
