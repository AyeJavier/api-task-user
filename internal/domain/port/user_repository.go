// Package port defines the driven ports (outbound interfaces) of the domain.
package port

import (
	"context"

	"github.com/javier/api-task-user/internal/domain/model"
)

// UserRepository is the outbound port for user persistence operations.
//
//go:generate mockgen -source=user_repository.go -destination=mocks/mock_user_repository.go -package=mocks
type UserRepository interface {
	FindByID(ctx context.Context, id string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindAll(ctx context.Context) ([]*model.User, error)
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id string) error
}
