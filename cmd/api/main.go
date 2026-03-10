// Package main is the entry point of the api-task-user application.
// It wires all dependencies (hexagonal composition root) and starts the HTTP server.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/javier/api-task-user/internal/adapter/inbound/http/handler"
	inboundhttp "github.com/javier/api-task-user/internal/adapter/inbound/http"
	"github.com/javier/api-task-user/internal/adapter/outbound/persistence/postgres"
	"github.com/javier/api-task-user/internal/config"
	"github.com/javier/api-task-user/internal/domain/service"
)

func main() {
	// Load .env (dev only — in production use real env vars)
	_ = godotenv.Load()

	cfg := config.Load()

	// Database connection pool
	db, err := pgxpool.New(context.Background(), cfg.Database.URL)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		log.Fatalf("pinging database: %v", err)
	}
	log.Println("connected to PostgreSQL")

	// --- Outbound adapters (repositories) ---
	userRepo := postgres.NewUserRepository(db)
	taskRepo := postgres.NewTaskRepository(db)

	// --- Auth outbound adapter (JWT + bcrypt) ---
	// TODO: implement JWTAdapter and BcryptAdapter in outbound/auth/
	// authAdapter := auth.NewJWTAdapter(cfg.JWT.Secret, cfg.JWT.ExpirationHours)
	// hashAdapter := auth.NewBcryptAdapter(cfg.Security.BcryptCost)

	// --- Domain services ---
	// authSvc  := service.NewAuthService(userRepo, authAdapter, hashAdapter)
	// userSvc  := service.NewUserService(userRepo, hashAdapter)
	taskSvc := service.NewTaskService(taskRepo, userRepo)

	// --- Inbound handlers ---
	// authH := handler.NewAuthHandler(authSvc)
	// userH := handler.NewUserHandler(userSvc)
	taskH := handler.NewTaskHandler(taskSvc)

	// --- Router ---
	// router := inboundhttp.NewRouter(authAdapter, authH, userH, taskH)
	_ = taskH
	_ = inboundhttp.NewRouter

	// TODO: remove placeholder once auth adapter is implemented
	router := http.NewServeMux()
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// --- HTTP Server with graceful shutdown ---
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.App.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("server listening on :%s (env: %s)", cfg.App.Port, cfg.App.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("server stopped")
}
