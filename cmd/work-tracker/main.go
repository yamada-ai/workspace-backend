package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/yamada-ai/workspace-backend/infrastructure/config"
	"github.com/yamada-ai/workspace-backend/infrastructure/database"
	infraRepo "github.com/yamada-ai/workspace-backend/infrastructure/database/repository"
	"github.com/yamada-ai/workspace-backend/infrastructure/database/sqlc"
	"github.com/yamada-ai/workspace-backend/presentation/http/dto"
	"github.com/yamada-ai/workspace-backend/presentation/http/handler"
	"github.com/yamada-ai/workspace-backend/usecase/command"
)

func main() {
	// Load configuration
	cfg := config.Load()
	log.Printf("Starting work-tracker server on %s", cfg.ServerPort)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to database
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()
	log.Println("✅ Connected to database")

	// === Dependency Injection (Bottom-up) ===

	// 1. Create sqlc Queries (Infrastructure layer)
	queries := sqlc.New(pool)

	// 2. Create Repository implementations
	userRepo := infraRepo.NewUserRepository(queries)
	sessionRepo := infraRepo.NewSessionRepository(queries)

	// 3. Create Use Cases
	joinUseCase := command.NewJoinCommandUseCase(userRepo, sessionRepo)

	// 4. Create HTTP Handlers
	commandHandler := handler.NewCommandHandler(joinUseCase)

	// 5. Setup Router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Register OpenAPI-generated routes
	handlerFunc := dto.HandlerFromMux(commandHandler, r)

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.ServerPort,
		Handler:      handlerFunc,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Server listening on %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited gracefully")
}
