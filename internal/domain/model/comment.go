package model

import "time"

// Comment represents a note left by an Executor on an expired task.
type Comment struct {
	ID        string
	TaskID    string
	AuthorID  string
	Body      string
	CreatedAt time.Time
}
