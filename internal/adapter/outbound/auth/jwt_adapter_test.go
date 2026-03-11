package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/javier/api-task-user/internal/adapter/outbound/auth"
	"github.com/javier/api-task-user/internal/domain/model"
	"github.com/javier/api-task-user/pkg/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "super-secret-test-key-256-bits!!"

func newJWTAdapter() *auth.JWTAdapter {
	return auth.NewJWTAdapter(testSecret, 1*time.Hour)
}

func TestJWTAdapter_GenerateAndValidate(t *testing.T) {
	adapter := newJWTAdapter()
	ctx := context.Background()

	user := &model.User{
		ID:      "user-123",
		Profile: model.ProfileAdmin,
	}

	token, err := adapter.GenerateToken(ctx, user)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := adapter.ValidateToken(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, model.ProfileAdmin, claims.Profile)
}

func TestJWTAdapter_ValidateToken_Invalid(t *testing.T) {
	adapter := newJWTAdapter()
	ctx := context.Background()

	claims, err := adapter.ValidateToken(ctx, "not-a-jwt")
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, apperror.ErrTokenInvalid)
}

func TestJWTAdapter_ValidateToken_WrongSecret(t *testing.T) {
	generator := auth.NewJWTAdapter("secret-A-32-chars-for-signing!!", 1*time.Hour)
	validator := auth.NewJWTAdapter("secret-B-32-chars-for-signing!!", 1*time.Hour)
	ctx := context.Background()

	user := &model.User{ID: "u1", Profile: model.ProfileExecutor}
	token, err := generator.GenerateToken(ctx, user)
	require.NoError(t, err)

	claims, err := validator.ValidateToken(ctx, token)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, apperror.ErrTokenInvalid)
}

func TestJWTAdapter_ExpiredToken(t *testing.T) {
	adapter := auth.NewJWTAdapter(testSecret, -1*time.Hour) // negative = already expired
	ctx := context.Background()

	user := &model.User{ID: "u1", Profile: model.ProfileAdmin}
	token, err := adapter.GenerateToken(ctx, user)
	require.NoError(t, err)

	claims, err := adapter.ValidateToken(ctx, token)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, apperror.ErrTokenInvalid)
}

func TestJWTAdapter_InvalidateToken(t *testing.T) {
	adapter := newJWTAdapter()
	ctx := context.Background()

	user := &model.User{ID: "u1", Profile: model.ProfileExecutor}
	token, err := adapter.GenerateToken(ctx, user)
	require.NoError(t, err)

	// Token is valid before invalidation
	claims, err := adapter.ValidateToken(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, "u1", claims.UserID)

	// Invalidate
	err = adapter.InvalidateToken(ctx, token)
	require.NoError(t, err)

	// Token is invalid after blacklisting
	claims, err = adapter.ValidateToken(ctx, token)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, apperror.ErrTokenInvalid)
}

func TestJWTAdapter_ProfilesArePreserved(t *testing.T) {
	adapter := newJWTAdapter()
	ctx := context.Background()

	profiles := []model.Profile{model.ProfileAdmin, model.ProfileExecutor, model.ProfileAuditor}

	for _, profile := range profiles {
		t.Run(string(profile), func(t *testing.T) {
			user := &model.User{ID: "u1", Profile: profile}
			token, err := adapter.GenerateToken(ctx, user)
			require.NoError(t, err)

			claims, err := adapter.ValidateToken(ctx, token)
			require.NoError(t, err)
			assert.Equal(t, profile, claims.Profile)
		})
	}
}
