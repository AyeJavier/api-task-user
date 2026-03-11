package mocks

import (
	"context"

	"github.com/javier/api-task-user/internal/domain/model"
	"github.com/javier/api-task-user/internal/domain/port"
	"github.com/stretchr/testify/mock"
)

// MockAuthPort is a mock implementation of port.AuthPort.
type MockAuthPort struct {
	mock.Mock
}

func (m *MockAuthPort) GenerateToken(ctx context.Context, user *model.User) (string, error) {
	args := m.Called(ctx, user)
	return args.String(0), args.Error(1)
}

func (m *MockAuthPort) ValidateToken(ctx context.Context, token string) (*port.TokenClaims, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.TokenClaims), args.Error(1)
}

func (m *MockAuthPort) InvalidateToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}
