// Package service contains the business logic of the domain.
package service

import (
	"context"
	"fmt"

	"github.com/javier/api-task-user/internal/domain/model"
	"github.com/javier/api-task-user/internal/domain/port"
	"github.com/javier/api-task-user/pkg/apperror"
)

// AuthService handles authentication business logic.
type AuthService struct {
	userRepo  port.UserRepository
	authPort  port.AuthPort
	hashPort  port.HashPort
}

// NewAuthService creates a new AuthService.
func NewAuthService(
	userRepo port.UserRepository,
	authPort port.AuthPort,
	hashPort port.HashPort,
) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		authPort: authPort,
		hashPort: hashPort,
	}
}

// LoginResult is returned after a successful login.
type LoginResult struct {
	Token              string
	MustChangePassword bool
}

// Login authenticates a user by email and password.
// Returns a signed JWT and a flag indicating whether the user must change their password.
func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, apperror.ErrInvalidCredentials
	}
	
	if err := s.hashPort.Compare(user.PasswordHash, password); err != nil {
		return nil, apperror.ErrInvalidCredentials
	}

	token, err := s.authPort.GenerateToken(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}

	return &LoginResult{
		Token:              token,
		MustChangePassword: user.MustChangePassword,
	}, nil
}

// Logout invalidates the provided token.
func (s *AuthService) Logout(ctx context.Context, token string) error {
	if err := s.authPort.InvalidateToken(ctx, token); err != nil {
		return fmt.Errorf("invalidating token: %w", err)
	}
	return nil
}

// ChangePassword updates the user's password and clears the must-change flag.
func (s *AuthService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return apperror.ErrUserNotFound
	}

	if err := s.hashPort.Compare(user.PasswordHash, currentPassword); err != nil {
		return apperror.ErrInvalidCredentials
	}	

	hash, err := s.hashPort.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	user.PasswordHash = hash
	user.MustChangePassword = false

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("updating user: %w", err)
	}
	return nil
}

// ensure model is used (avoid import cycle in stubs)
var _ *model.User
