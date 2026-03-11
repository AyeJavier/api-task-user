package middleware

import (
	"net/http"

	"github.com/javier/api-task-user/internal/domain/model"
)

// RequireProfile returns middleware that allows only requests from users
// whose profile matches one of the given allowed profiles.
func RequireProfile(allowed ...model.Profile) func(http.Handler) http.Handler {
	set := make(map[model.Profile]struct{}, len(allowed))
	for _, p := range allowed {
		set[p] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := ClaimsFromContext(r.Context())
			if claims == nil {
				http.Error(w, "unauthorizedxxxx", http.StatusUnauthorized)
				return
			}
			if claims.MustChangePassword {
				http.Error(w, "unauthorizedddd", http.StatusUnauthorized)
				return
			}
			if _, ok := set[claims.Profile]; !ok {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
