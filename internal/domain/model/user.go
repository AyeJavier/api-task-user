package model

import "time"

// User represents a system user with a specific access profile.
type User struct {
	ID                 string
	Name               string
	Email              string
	PasswordHash       string
	Profile            Profile
	MustChangePassword bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// IsAdmin reports whether the user holds the Admin profile.
func (u *User) IsAdmin() bool { return u.Profile == ProfileAdmin }

// IsExecutor reports whether the user holds the Executor profile.
func (u *User) IsExecutor() bool { return u.Profile == ProfileExecutor }

// IsAuditor reports whether the user holds the Auditor profile.
func (u *User) IsAuditor() bool { return u.Profile == ProfileAuditor }
