// Package apperror defines typed domain errors used across the application.
package apperror

import "errors"

// Domain errors — use errors.Is() to compare.
var (
	// User errors
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already in use")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrForbidden          = errors.New("action not allowed for this profile")
	ErrCannotCreateAdmin  = errors.New("admins cannot create other admins")

	// Task errors
	ErrTaskNotFound      = errors.New("task not found")
	ErrTaskNotMutable    = errors.New("task can only be modified when in ASSIGNED status")
	ErrInvalidTransition = errors.New("invalid state transition")
	ErrTaskExpired       = errors.New("task is expired — status change not allowed")

	// Auth errors
	ErrTokenInvalid = errors.New("token is invalid or expired")
	ErrTokenMissing = errors.New("authorization token missing")
)

// AppError wraps a domain error with an HTTP-friendly status code and message.
type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }

// New creates a new AppError.
func New(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}
