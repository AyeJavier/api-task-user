package model_test

import (
	"testing"
	"time"

	"github.com/javier/api-task-user/internal/domain/model"
	"github.com/javier/api-task-user/pkg/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTask(status model.TaskStatus, dueDate time.Time) *model.Task {
	return &model.Task{
		ID:         "task-1",
		Title:      "Test Task",
		Status:     status,
		AssigneeID: "executor-1",
		DueDate:    dueDate,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// ---------- IsExpired ----------

func TestTask_IsExpired_WhenPastDueDate(t *testing.T) {
	task := newTask(model.TaskStatusAssigned, time.Now().Add(-24*time.Hour))
	assert.True(t, task.IsExpired())
}

func TestTask_IsExpired_WhenNotPastDueDate(t *testing.T) {
	task := newTask(model.TaskStatusAssigned, time.Now().Add(24*time.Hour))
	assert.False(t, task.IsExpired())
}

// ---------- CanTransitionTo ----------

func TestTask_CanTransitionTo_ValidTransitions(t *testing.T) {
	tests := []struct {
		name   string
		from   model.TaskStatus
		to     model.TaskStatus
		expect bool
	}{
		// ASSIGNED transitions
		{"assigned -> started", model.TaskStatusAssigned, model.TaskStatusStarted, true},
		{"assigned -> finished_success (invalid)", model.TaskStatusAssigned, model.TaskStatusFinishedSuccess, false},
		{"assigned -> on_hold (invalid)", model.TaskStatusAssigned, model.TaskStatusOnHold, false},

		// STARTED transitions
		{"started -> finished_success", model.TaskStatusStarted, model.TaskStatusFinishedSuccess, true},
		{"started -> finished_error", model.TaskStatusStarted, model.TaskStatusFinishedError, true},
		{"started -> on_hold", model.TaskStatusStarted, model.TaskStatusOnHold, true},
		{"started -> assigned (invalid)", model.TaskStatusStarted, model.TaskStatusAssigned, false},

		// ON_HOLD transitions
		{"on_hold -> started", model.TaskStatusOnHold, model.TaskStatusStarted, true},
		{"on_hold -> finished_success (invalid)", model.TaskStatusOnHold, model.TaskStatusFinishedSuccess, false},

		// Terminal states — no transitions
		{"finished_success -> any (invalid)", model.TaskStatusFinishedSuccess, model.TaskStatusStarted, false},
		{"finished_error -> any (invalid)", model.TaskStatusFinishedError, model.TaskStatusStarted, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := newTask(tt.from, time.Now().Add(24*time.Hour))
			assert.Equal(t, tt.expect, task.CanTransitionTo(tt.to))
		})
	}
}

// ---------- Transition ----------

func TestTask_Transition_Success(t *testing.T) {
	task := newTask(model.TaskStatusAssigned, time.Now().Add(24*time.Hour))
	beforeUpdate := task.UpdatedAt

	err := task.Transition(model.TaskStatusStarted)

	require.NoError(t, err)
	assert.Equal(t, model.TaskStatusStarted, task.Status)
	assert.True(t, task.UpdatedAt.After(beforeUpdate) || task.UpdatedAt.Equal(beforeUpdate))
}

func TestTask_Transition_InvalidTransition(t *testing.T) {
	task := newTask(model.TaskStatusAssigned, time.Now().Add(24*time.Hour))

	err := task.Transition(model.TaskStatusFinishedSuccess)

	assert.ErrorIs(t, err, apperror.ErrInvalidTransition)
	assert.Equal(t, model.TaskStatusAssigned, task.Status, "status should not change")
}

func TestTask_Transition_ExpiredTask(t *testing.T) {
	task := newTask(model.TaskStatusAssigned, time.Now().Add(-24*time.Hour))

	err := task.Transition(model.TaskStatusStarted)

	assert.ErrorIs(t, err, apperror.ErrTaskExpired)
	assert.Equal(t, model.TaskStatusAssigned, task.Status, "status should not change")
}

func TestTask_Transition_OnHoldToStartedCycle(t *testing.T) {
	task := newTask(model.TaskStatusOnHold, time.Now().Add(24*time.Hour))

	err := task.Transition(model.TaskStatusStarted)
	require.NoError(t, err)
	assert.Equal(t, model.TaskStatusStarted, task.Status)

	err = task.Transition(model.TaskStatusOnHold)
	require.NoError(t, err)
	assert.Equal(t, model.TaskStatusOnHold, task.Status)

	err = task.Transition(model.TaskStatusStarted)
	require.NoError(t, err)
	assert.Equal(t, model.TaskStatusStarted, task.Status)
}

func TestTask_Transition_TerminalStatesAreImmutable(t *testing.T) {
	terminals := []model.TaskStatus{model.TaskStatusFinishedSuccess, model.TaskStatusFinishedError}
	allTargets := []model.TaskStatus{
		model.TaskStatusAssigned, model.TaskStatusStarted,
		model.TaskStatusFinishedSuccess, model.TaskStatusFinishedError,
		model.TaskStatusOnHold,
	}

	for _, terminal := range terminals {
		for _, target := range allTargets {
			task := newTask(terminal, time.Now().Add(24*time.Hour))
			err := task.Transition(target)
			assert.ErrorIs(t, err, apperror.ErrInvalidTransition,
				"from %s to %s should fail", terminal, target)
		}
	}
}

// ---------- IsMutableByAdmin ----------

func TestTask_IsMutableByAdmin(t *testing.T) {
	tests := []struct {
		name   string
		status model.TaskStatus
		want   bool
	}{
		{"assigned is mutable", model.TaskStatusAssigned, true},
		{"started is not mutable", model.TaskStatusStarted, false},
		{"on_hold is not mutable", model.TaskStatusOnHold, false},
		{"finished_success is not mutable", model.TaskStatusFinishedSuccess, false},
		{"finished_error is not mutable", model.TaskStatusFinishedError, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := newTask(tt.status, time.Now().Add(24*time.Hour))
			assert.Equal(t, tt.want, task.IsMutableByAdmin())
		})
	}
}
