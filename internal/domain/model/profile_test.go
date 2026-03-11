package model_test

import (
	"testing"

	"github.com/javier/api-task-user/internal/domain/model"
	"github.com/stretchr/testify/assert"
)

func TestProfile_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		profile model.Profile
		want    bool
	}{
		{"admin is valid", model.ProfileAdmin, true},
		{"executor is valid", model.ProfileExecutor, true},
		{"auditor is valid", model.ProfileAuditor, true},
		{"empty is invalid", model.Profile(""), false},
		{"unknown is invalid", model.Profile("SUPERUSER"), false},
		{"lowercase admin is invalid", model.Profile("admin"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.profile.IsValid())
		})
	}
}

func TestProfile_CanBeCreatedBy(t *testing.T) {
	tests := []struct {
		name    string
		profile model.Profile
		creator model.Profile
		want    bool
	}{
		{"admin can create executor", model.ProfileExecutor, model.ProfileAdmin, true},
		{"admin can create auditor", model.ProfileAuditor, model.ProfileAdmin, true},
		{"admin cannot create admin", model.ProfileAdmin, model.ProfileAdmin, false},
		{"executor cannot create anyone", model.ProfileExecutor, model.ProfileExecutor, false},
		{"auditor cannot create anyone", model.ProfileExecutor, model.ProfileAuditor, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.profile.CanBeCreatedBy(tt.creator))
		})
	}
}
