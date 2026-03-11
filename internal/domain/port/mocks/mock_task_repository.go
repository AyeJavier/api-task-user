package mocks

import (
	"context"

	"github.com/javier/api-task-user/internal/domain/model"
	"github.com/stretchr/testify/mock"
)

// MockTaskRepository is a mock implementation of port.TaskRepository.
type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) FindByID(ctx context.Context, id string) (*model.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func (m *MockTaskRepository) FindByAssignee(ctx context.Context, assigneeID string) ([]*model.Task, error) {
	args := m.Called(ctx, assigneeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Task), args.Error(1)
}

func (m *MockTaskRepository) FindAll(ctx context.Context) ([]*model.Task, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Task), args.Error(1)
}

func (m *MockTaskRepository) Create(ctx context.Context, task *model.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepository) Update(ctx context.Context, task *model.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTaskRepository) CreateComment(ctx context.Context, comment *model.Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}
