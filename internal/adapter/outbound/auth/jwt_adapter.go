// Package auth implements the outbound authentication adapters.
package auth

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/javier/api-task-user/internal/domain/model"
	"github.com/javier/api-task-user/internal/domain/port"
	"github.com/javier/api-task-user/pkg/apperror"
)

// JWTAdapter implements port.AuthPort using golang-jwt.
type JWTAdapter struct {
	secret     []byte
	expiration time.Duration
	// blacklist stores invalidated tokens (in production, use Redis).
	blacklist map[string]struct{}
	mu        sync.RWMutex
}

// NewJWTAdapter creates a new JWT adapter.
func NewJWTAdapter(secret string, expiration time.Duration) *JWTAdapter {
	return &JWTAdapter{
		secret:     []byte(secret),
		expiration: expiration,
		blacklist:  make(map[string]struct{}),
	}
}

// customClaims extends the standard JWT claims with application-specific data.
type customClaims struct {
	jwt.RegisteredClaims
	UserID  string        `json:"uid"`
	Profile model.Profile `json:"profile"`
	MustChangePassword bool `json:"must_change_password"`
}

// GenerateToken creates a signed JWT for the given user.
func (a *JWTAdapter) GenerateToken(_ context.Context, user *model.User) (string, error) {
	now := time.Now()
	claims := customClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(a.expiration)),
		},
		UserID:  user.ID,
		Profile: user.Profile,
		MustChangePassword: user.MustChangePassword,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(a.secret)
	if err != nil {
		return "", fmt.Errorf("signing JWT: %w", err)
	}
	return signed, nil
}

// ValidateToken parses and validates a JWT, returning the embedded claims.
func (a *JWTAdapter) ValidateToken(_ context.Context, tokenStr string) (*port.TokenClaims, error) {
	a.mu.RLock()
	_, blacklisted := a.blacklist[tokenStr]
	a.mu.RUnlock()
	if blacklisted {
		return nil, apperror.ErrTokenInvalid
	}

	token, err := jwt.ParseWithClaims(tokenStr, &customClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return a.secret, nil
	})
	if err != nil {
		return nil, apperror.ErrTokenInvalid
	}

	claims, ok := token.Claims.(*customClaims)
	if !ok || !token.Valid {
		return nil, apperror.ErrTokenInvalid
	}

	return &port.TokenClaims{
		UserID:  claims.UserID,
		Profile: claims.Profile,
		MustChangePassword: claims.MustChangePassword,
	}, nil
}

// InvalidateToken adds the token to the in-memory blacklist.
func (a *JWTAdapter) InvalidateToken(_ context.Context, tokenStr string) error {
	a.mu.Lock()
	a.blacklist[tokenStr] = struct{}{}
	a.mu.Unlock()
	return nil
}
