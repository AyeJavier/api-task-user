// Package dto defines the HTTP request and response data transfer objects.
package dto

import (
	"time"

	"github.com/javier/api-task-user/internal/domain/model"
)

// --- Auth ---

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password"     validate:"required,min=8"`
}

// --- Users ---

type CreateUserRequest struct {
	Name    string        `json:"name"    validate:"required,min=2,max=100"`
	Email   string        `json:"email"   validate:"required,email"`
	Profile model.Profile `json:"profile" validate:"required,oneof=EXECUTOR AUDITOR"`
}

type UpdateUserRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

// --- Tasks ---

type CreateTaskRequest struct {
	Title       string    `json:"title"       validate:"required,min=3,max=200"`
	Description string    `json:"description" validate:"required"`
	AssigneeID  string    `json:"assignee_id" validate:"required,uuid4"`
	DueDate     time.Time `json:"due_date"    validate:"required"`
}

type UpdateTaskRequest struct {
	Title       string    `json:"title"       validate:"required,min=3,max=200"`
	Description string    `json:"description" validate:"required"`
	DueDate     time.Time `json:"due_date"    validate:"required"`
}

type ChangeTaskStatusRequest struct {
	Status model.TaskStatus `json:"status" validate:"required,oneof=STARTED FINISHED_SUCCESS FINISHED_ERROR ON_HOLD"`
}

type AddCommentRequest struct {
	Body string `json:"body" validate:"required,min=1,max=1000"`
}
