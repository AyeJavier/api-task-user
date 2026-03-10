// Package http wires together the HTTP routes, middleware and handlers.
package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/javier/api-task-user/internal/adapter/inbound/http/handler"
	"github.com/javier/api-task-user/internal/adapter/inbound/http/middleware"
	"github.com/javier/api-task-user/internal/domain/model"
	"github.com/javier/api-task-user/internal/domain/port"
)

// NewRouter creates and returns the fully configured HTTP router.
func NewRouter(
	authPort  port.AuthPort,
	authH     *handler.AuthHandler,
	userH     *handler.UserHandler,
	taskH     *handler.TaskHandler,
) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.CleanPath)

	// Health check (public)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Public auth routes
	r.Post("/auth/login", authH.Login)

	

	return r
}
