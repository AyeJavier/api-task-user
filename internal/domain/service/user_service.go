package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/javier/api-task-user/internal/domain/model"
	"github.com/javier/api-task-user/internal/domain/port"
	"github.com/javier/api-task-user/pkg/apperror"
)

// UserService handles user management business logic.
type UserService struct {
	userRepo port.UserRepository
	hashPort port.HashPort
}

// NewUserService creates a new UserService.
func NewUserService(userRepo port.UserRepository, hashPort port.HashPort) *UserService {
	return &UserService{userRepo: userRepo, hashPort: hashPort}
}

// CreateUserInput holds the data required to create a new user.
type CreateUserInput struct {
	Name    string
	Email   string
	Profile model.Profile
}

// Create creates a new user with a temporary password.
// Only Admins can call this, and they cannot create other Admins.
func (s *UserService) Create(ctx context.Context, requester *model.User, input CreateUserInput) (*model.User, error) {
	if !requester.IsAdmin() {
		return nil, apperror.ErrForbidden
	}
	if !input.Profile.CanBeCreatedBy(requester.Profile) {
		return nil, apperror.ErrCannotCreateAdmin
	}

	existing, _ := s.userRepo.FindByEmail(ctx, input.Email)
	if existing != nil {
		return nil, apperror.ErrEmailAlreadyExists
	}

	tempPassword := generateTemporaryPassword()
	hash, err := s.hashPort.Hash(tempPassword)
	if err != nil {
		return nil, fmt.Errorf("hashing temporary password: %w", err)
	}

	user := &model.User{
		ID:                 uuid.NewString(),
		Name:               input.Name,
		Email:              input.Email,
		PasswordHash:       hash,
		Profile:            input.Profile,
		MustChangePassword: true,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("persisting user: %w", err)
	}

	// Return with plain-text temp password so the caller can communicate it.
	user.PasswordHash = tempPassword
	return user, nil
}

// GetByID retrieves a user by ID.
func (s *UserService) GetByID(ctx context.Context, requester *model.User, id string) (*model.User, error) {
	if !requester.IsAdmin() {
		return nil, apperror.ErrForbidden
	}
	return s.userRepo.FindByID(ctx, id)
}

// ListAll returns all users. Admin only.
func (s *UserService) ListAll(ctx context.Context, requester *model.User) ([]*model.User, error) {
	if !requester.IsAdmin() {
		return nil, apperror.ErrForbidden
	}
	return s.userRepo.FindAll(ctx)
}

// Update updates an existing user's details. Admin only.
func (s *UserService) Update(ctx context.Context, requester *model.User, user *model.User) error {
	if !requester.IsAdmin() {
		return apperror.ErrForbidden
	}
	user.UpdatedAt = time.Now()
	return s.userRepo.Update(ctx, user)
}

// Delete removes a user. Admin only.
func (s *UserService) Delete(ctx context.Context, requester *model.User, id string) error {
	if !requester.IsAdmin() {
		return apperror.ErrForbidden
	}
	return s.userRepo.Delete(ctx, id)
}

func generateTemporaryPassword() string {
	return "Temp@" + uuid.NewString()[:8]
}
