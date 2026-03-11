package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"github.com/javier/api-task-user/internal/adapter/inbound/http/middleware"
	"github.com/javier/api-task-user/internal/domain/model"
	"github.com/javier/api-task-user/internal/domain/port"
	"github.com/javier/api-task-user/internal/domain/port/mocks"
	"github.com/stretchr/testify/mock" 
)

const (
	tokenAdmin    = "token-admin"
	tokenExecutor = "token-executor"
	tokenAuditor  = "token-auditor"
)

// buildAuthPort returns a MockAuthPort that maps known test tokens to their claims.
func buildAuthPort() *mocks.MockAuthPort {
	// ap := new(mocks.MockAuthPort)
	// ap.On("ValidateToken", context.Background(), tokenAdmin).Return(
	// 	&port.TokenClaims{UserID: "admin-id", Profile: model.ProfileAdmin}, nil,
	// )
	// ap.On("ValidateToken", context.Background(), tokenExecutor).Return(
	// 	&port.TokenClaims{UserID: "exec-id", Profile: model.ProfileExecutor}, nil,
	// )
	// ap.On("ValidateToken", context.Background(), tokenAuditor).Return(
	// 	&port.TokenClaims{UserID: "aud-id", Profile: model.ProfileAuditor}, nil,
	// )
	ap := new(mocks.MockAuthPort)
	ap.On("ValidateToken", mock.Anything, tokenAdmin).Return(
		&port.TokenClaims{UserID: "admin-id", Profile: model.ProfileAdmin}, nil,
	)
	ap.On("ValidateToken", mock.Anything, tokenExecutor).Return(
		&port.TokenClaims{UserID: "exec-id", Profile: model.ProfileExecutor}, nil,
	)
	ap.On("ValidateToken", mock.Anything, tokenAuditor).Return(
		&port.TokenClaims{UserID: "aud-id", Profile: model.ProfileAuditor}, nil,
	)
	return ap
}

// withAuth wraps a handler with the Authenticate middleware and injects claims
// using the given bearer token string (e.g. "token-admin").
func withAuth(ap *mocks.MockAuthPort, h http.Handler) http.Handler {
	return middleware.Authenticate(ap)(h)
}

// newReq is a thin helper to build a request with a Bearer token header.
func newReq(method, path, token string, body *bodyBuffer) *http.Request {
	var r *http.Request
	if body != nil {
		r = httptest.NewRequest(method, path, body)
		r.Header.Set("Content-Type", "application/json")
	} else {
		r = httptest.NewRequest(method, path, nil)
	}
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	return r
}

// bodyBuffer is a thin bytes.Buffer alias with an io.Reader interface.
type bodyBuffer struct {
	data []byte
	pos  int
}

func jsonBody(s string) *bodyBuffer { return &bodyBuffer{data: []byte(s)} }

func (b *bodyBuffer) Read(p []byte) (int, error) {
	if b.pos >= len(b.data) {
		return 0, nil
	}
	n := copy(p, b.data[b.pos:])
	b.pos += n
	return n, nil
}

// addChiParam attaches URL params to the request context (chi router simulation).
func addChiParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}
