package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// BcryptAdapter implements port.HashPort using bcrypt.
type BcryptAdapter struct {
	cost int
}

// NewBcryptAdapter creates a new bcrypt adapter with the given cost factor.
func NewBcryptAdapter(cost int) *BcryptAdapter {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	return &BcryptAdapter{cost: cost}
}

// Hash generates a bcrypt hash of the given password.
func (a *BcryptAdapter) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), a.cost)
	if err != nil {
		return "", fmt.Errorf("bcrypt hash: %w", err)
	}
	return string(hash), nil
}

// Compare checks whether the given password matches the bcrypt hash.
// Returns nil on success or an error if they don't match.
func (a *BcryptAdapter) Compare(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
