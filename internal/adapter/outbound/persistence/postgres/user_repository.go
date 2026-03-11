// Package postgres implements the domain repository ports using PostgreSQL.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/javier/api-task-user/internal/domain/model"
	"github.com/javier/api-task-user/pkg/apperror"
)

// UserRepository implements port.UserRepository using PostgreSQL.
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new PostgreSQL-backed UserRepository.
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, name, email, password_hash, profile, must_change_password, created_at, updated_at
		 FROM users WHERE id = $1`, id)

	u := &model.User{}
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash,
		&u.Profile, &u.MustChangePassword, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, apperror.ErrUserNotFound
	}
	return u, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, name, email, password_hash, profile, must_change_password, created_at, updated_at
		 FROM users WHERE email = $1`, email)

	u := &model.User{}
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash,
		&u.Profile, &u.MustChangePassword, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, apperror.ErrUserNotFound
	}
	return u, nil
}

func (r *UserRepository) FindAll(ctx context.Context) ([]*model.User, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, email, password_hash, profile, must_change_password, created_at, updated_at
		 FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("querying users: %w", err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		u := &model.User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash,
			&u.Profile, &u.MustChangePassword, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning user row: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO users (id, name, email, password_hash, profile, must_change_password, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		user.ID, user.Name, user.Email, user.PasswordHash,
		user.Profile, user.MustChangePassword, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("inserting user: %w", err)
	}
	return nil
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET name=$1, email=$2, password_hash=$3, profile=$4,
		 must_change_password=$5, updated_at=$6 WHERE id=$7`,
		user.Name, user.Email, user.PasswordHash, user.Profile,
		user.MustChangePassword, user.UpdatedAt, user.ID)
	if err != nil {
		return fmt.Errorf("updating user: %w", err)
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting user: %w", err)
	}
	return nil
}
