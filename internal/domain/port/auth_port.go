package port

import (
	"context"

	"github.com/javier/api-task-user/internal/domain/model"
)

// TokenClaims holds the decoded payload from a JWT.
type TokenClaims struct {
	UserID  string
	Profile model.Profile
}

// AuthPort is the outbound port for token generation and validation.
//
//go:generate mockgen -source=auth_port.go -destination=mocks/mock_auth_port.go -package=mocks
type AuthPort interface {
	GenerateToken(ctx context.Context, user *model.User) (string, error)
	ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
	InvalidateToken(ctx context.Context, token string) error
}
