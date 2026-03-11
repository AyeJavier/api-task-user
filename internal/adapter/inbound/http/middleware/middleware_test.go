package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/javier/api-task-user/internal/adapter/inbound/http/middleware"
	"github.com/javier/api-task-user/internal/domain/model"
	"github.com/javier/api-task-user/internal/domain/port"
	"github.com/javier/api-task-user/internal/domain/port/mocks"
	"github.com/javier/api-task-user/pkg/apperror"
	"github.com/stretchr/testify/assert"
)

// okHandler is a simple handler that writes 200 OK.
var okHandler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func newRequest(method, authHeader string) *http.Request {
	r := httptest.NewRequest(method, "/", nil)
	if authHeader != "" {
		r.Header.Set("Authorization", authHeader)
	}
	return r
}

// ==================== Authenticate ====================

func TestAuthenticate_ValidToken(t *testing.T) {
	authPort := new(mocks.MockAuthPort)
	claims := &port.TokenClaims{UserID: "u1", Profile: model.ProfileAdmin, MustChangePassword: false}
	authPort.On("ValidateToken", context.Background(), "valid-token").Return(claims, nil)

	handler := middleware.Authenticate(authPort)(okHandler)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, newRequest(http.MethodGet, "Bearer valid-token"))

	assert.Equal(t, http.StatusOK, w.Code)
	authPort.AssertExpectations(t)
}

func TestAuthenticate_MissingToken(t *testing.T) {
	authPort := new(mocks.MockAuthPort)

	handler := middleware.Authenticate(authPort)(okHandler)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, newRequest(http.MethodGet, ""))

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	authPort.AssertNotCalled(t, "ValidateToken")
}

func TestAuthenticate_NoBearerPrefix(t *testing.T) {
	authPort := new(mocks.MockAuthPort)

	handler := middleware.Authenticate(authPort)(okHandler)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, newRequest(http.MethodGet, "Token abc123"))

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthenticate_InvalidToken(t *testing.T) {
	authPort := new(mocks.MockAuthPort)
	authPort.On("ValidateToken", context.Background(), "bad-token").Return(nil, apperror.ErrTokenInvalid)

	handler := middleware.Authenticate(authPort)(okHandler)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, newRequest(http.MethodGet, "Bearer bad-token"))

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthenticate_ClaimsInjectedInContext(t *testing.T) {
	authPort := new(mocks.MockAuthPort)
	expected := &port.TokenClaims{UserID: "user-42", Profile: model.ProfileExecutor, MustChangePassword: false}
	authPort.On("ValidateToken", context.Background(), "my-token").Return(expected, nil)

	var capturedClaims *port.TokenClaims
	capture := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		capturedClaims = middleware.ClaimsFromContext(r.Context())
	})

	handler := middleware.Authenticate(authPort)(capture)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, newRequest(http.MethodGet, "Bearer my-token"))

	assert.NotNil(t, capturedClaims)
	assert.Equal(t, "user-42", capturedClaims.UserID)
	assert.Equal(t, model.ProfileExecutor, capturedClaims.Profile)
}

// ==================== RequireProfile ====================

func newRequestWithClaims(profile model.Profile) *http.Request {
	authPort := new(mocks.MockAuthPort)
	claims := &port.TokenClaims{UserID: "u1", Profile: profile, MustChangePassword: false}
	authPort.On("ValidateToken", context.Background(), "tok").Return(claims, nil)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer tok")

	var out *http.Request
	capture := http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
		out = req
	})
	w := httptest.NewRecorder()
	middleware.Authenticate(authPort)(capture).ServeHTTP(w, r)
	return out
}

func TestRequireProfile_AllowedProfile(t *testing.T) {
	profiles := []model.Profile{model.ProfileAdmin, model.ProfileExecutor, model.ProfileAuditor}

	for _, p := range profiles {
		t.Run(string(p), func(t *testing.T) {
			r := newRequestWithClaims(p)
			w := httptest.NewRecorder()
			middleware.RequireProfile(p)(okHandler).ServeHTTP(w, r)
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestRequireProfile_DisallowedProfile(t *testing.T) {
	r := newRequestWithClaims(model.ProfileExecutor)
	w := httptest.NewRecorder()
	middleware.RequireProfile(model.ProfileAdmin)(okHandler).ServeHTTP(w, r)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireProfile_MultipleAllowed(t *testing.T) {
	r := newRequestWithClaims(model.ProfileAuditor)
	w := httptest.NewRecorder()
	middleware.RequireProfile(model.ProfileAdmin, model.ProfileAuditor)(okHandler).ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireProfile_NoClaims(t *testing.T) {
	// Request without going through Authenticate — no claims in context
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	middleware.RequireProfile(model.ProfileAdmin)(okHandler).ServeHTTP(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
