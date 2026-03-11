package model_test

import (
	"testing"
	"time"

	"github.com/javier/api-task-user/internal/domain/model"
	"github.com/stretchr/testify/assert"
)

func TestComment_FieldsAreSet(t *testing.T) {
	now := time.Now()
	c := model.Comment{
		ID:        "comment-1",
		TaskID:    "task-1",
		AuthorID:  "user-1",
		Body:      "This task could not be completed due to API downtime.",
		CreatedAt: now,
	}

	assert.Equal(t, "comment-1", c.ID)
	assert.Equal(t, "task-1", c.TaskID)
	assert.Equal(t, "user-1", c.AuthorID)
	assert.NotEmpty(t, c.Body)
	assert.Equal(t, now, c.CreatedAt)
}
