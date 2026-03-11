package mocks

import "github.com/stretchr/testify/mock"

// MockHashPort is a mock implementation of port.HashPort.
type MockHashPort struct {
	mock.Mock
}

func (m *MockHashPort) Hash(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockHashPort) Compare(hash, password string) error {
	args := m.Called(hash, password)
	return args.Error(0)
}
