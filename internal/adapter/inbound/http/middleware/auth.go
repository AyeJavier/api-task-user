// Package middleware provides HTTP middleware for the inbound adapter.
package middleware

import (
	"context"
	"net/http"
	"strings"
	"github.com/javier/api-task-user/internal/domain/port"
	"github.com/javier/api-task-user/pkg/apperror"
)

type contextKey string

const claimsKey contextKey = "claims"

// Authenticate validates the Bearer token and injects the claims into the request context.
func Authenticate(authPort port.AuthPort) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r)
			
			if token == "" {
				respondUnauthorized(w, apperror.ErrTokenMissing.Error())
				return
			}

			claims, err := authPort.ValidateToken(r.Context(), token)
			
			if err != nil {
				respondUnauthorized(w, apperror.ErrTokenInvalid.Error())
				return
			}
			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ClaimsFromContext retrieves the token claims stored in the request context.
func ClaimsFromContext(ctx context.Context) *port.TokenClaims {
	v, _ := ctx.Value(claimsKey).(*port.TokenClaims)
	return v
}

func extractBearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(h, "Bearer ")
}

func respondUnauthorized(w http.ResponseWriter, msg string) {
	http.Error(w, msg, http.StatusUnauthorized)
}
