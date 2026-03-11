package dto

import (
	"time"

	"github.com/javier/api-task-user/internal/domain/model"
)

// ErrorResponse is the standard error envelope.
type ErrorResponse struct {
	Error  string      `json:"error"`
	Detail interface{} `json:"detail,omitempty"`
}

// LoginResponse is returned on successful authentication.
type LoginResponse struct {
	Token              string `json:"token"`
	MustChangePassword bool   `json:"must_change_password"`
}

// UserResponse is the public representation of a user.
type UserResponse struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Email     string       `json:"email"`
	Profile   model.Profile `json:"profile"`
	CreatedAt time.Time    `json:"created_at"`
}

// UserCreatedResponse includes the temporary password (only on creation).
type UserCreatedResponse struct {
	UserResponse
	TemporaryPassword string `json:"temporary_password"`
}

// TaskResponse is the public representation of a task.
type TaskResponse struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Status      model.TaskStatus `json:"status"`
	AssigneeID  string           `json:"assignee_id"`
	DueDate     time.Time        `json:"due_date"`
	IsExpired   bool             `json:"is_expired"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// CommentResponse is the public representation of a comment.
type CommentResponse struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	AuthorID  string    `json:"author_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

// --- Mappers ---

func ToUserResponse(u *model.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Profile:   u.Profile,
		CreatedAt: u.CreatedAt,
	}
}

func ToTaskResponse(t *model.Task) TaskResponse {
	return TaskResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		AssigneeID:  t.AssigneeID,
		DueDate:     t.DueDate,
		IsExpired:   t.IsExpired(),
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func ToCommentResponse(c *model.Comment) CommentResponse {
	return CommentResponse{
		ID:        c.ID,
		TaskID:    c.TaskID,
		AuthorID:  c.AuthorID,
		Body:      c.Body,
		CreatedAt: c.CreatedAt,
	}
}
