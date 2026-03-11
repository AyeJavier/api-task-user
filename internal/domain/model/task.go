package model

import (
	"time"

	"github.com/javier/api-task-user/pkg/apperror"
)

// TaskStatus represents the current state of a task.
type TaskStatus string

const (
	TaskStatusAssigned        TaskStatus = "ASSIGNED"
	TaskStatusStarted         TaskStatus = "STARTED"
	TaskStatusFinishedSuccess TaskStatus = "FINISHED_SUCCESS"
	TaskStatusFinishedError   TaskStatus = "FINISHED_ERROR"
	TaskStatusOnHold          TaskStatus = "ON_HOLD"
)

// validTransitions defines the allowed state machine transitions.
var validTransitions = map[TaskStatus][]TaskStatus{
	TaskStatusAssigned:        {TaskStatusStarted},
	TaskStatusStarted:         {TaskStatusFinishedSuccess, TaskStatusFinishedError, TaskStatusOnHold},
	TaskStatusOnHold:          {TaskStatusFinishedSuccess, TaskStatusFinishedError, TaskStatusOnHold},
	TaskStatusFinishedSuccess: {}, // terminal
	TaskStatusFinishedError:   {}, // terminal
}

// Task represents a unit of work assigned to an Executor.
type Task struct {
	ID          string
	Title       string
	Description string
	Status      TaskStatus
	AssigneeID  string
	DueDate     time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// IsExpired reports whether the task has passed its due date.
func (t *Task) IsExpired() bool {
	return time.Now().After(t.DueDate)
}

// CanTransitionTo reports whether the task can move to the target status.
func (t *Task) CanTransitionTo(target TaskStatus) bool {
	for _, allowed := range validTransitions[t.Status] {
		if allowed == target {
			return true
		}
	}
	return false
}

// Transition attempts to move the task to the target status.
// Returns an error if the transition is not allowed or the task is expired.
func (t *Task) Transition(target TaskStatus) error {
	if !t.CanTransitionTo(target) {
		return apperror.ErrInvalidTransition
	}
	if t.IsExpired() {
		return apperror.ErrTaskExpired
	}
	t.Status = target
	t.UpdatedAt = time.Now()
	return nil
}

// IsMutableByAdmin reports whether an Admin can update or delete this task.
// Only tasks in ASSIGNED state are mutable.
func (t *Task) IsMutableByAdmin() bool {
	return t.Status == TaskStatusAssigned
}
