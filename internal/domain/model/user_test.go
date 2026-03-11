package model_test

import (
	"testing"

	"github.com/javier/api-task-user/internal/domain/model"
	"github.com/stretchr/testify/assert"
)

func newUser(profile model.Profile) *model.User {
	return &model.User{
		ID:      "user-1",
		Name:    "Test User",
		Email:   "test@example.com",
		Profile: profile,
	}
}

func TestUser_IsAdmin(t *testing.T) {
	assert.True(t, newUser(model.ProfileAdmin).IsAdmin())
	assert.False(t, newUser(model.ProfileExecutor).IsAdmin())
	assert.False(t, newUser(model.ProfileAuditor).IsAdmin())
}

func TestUser_IsExecutor(t *testing.T) {
	assert.True(t, newUser(model.ProfileExecutor).IsExecutor())
	assert.False(t, newUser(model.ProfileAdmin).IsExecutor())
	assert.False(t, newUser(model.ProfileAuditor).IsExecutor())
}

func TestUser_IsAuditor(t *testing.T) {
	assert.True(t, newUser(model.ProfileAuditor).IsAuditor())
	assert.False(t, newUser(model.ProfileAdmin).IsAuditor())
	assert.False(t, newUser(model.ProfileExecutor).IsAuditor())
}
