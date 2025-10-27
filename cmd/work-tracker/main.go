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
	"github.com/yamada-ai/workspace-backend/presentation/ws"
	"github.com/yamada-ai/workspace-backend/usecase/command"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Printf("Starting work-tracker server on %s", cfg.ServerPort)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run database migrations
	if err := database.RunMigrations(cfg.DatabaseURL, "migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

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
	userRepository := infraRepo.NewUserRepositoryWithPool(pool)
	sessionRepository := infraRepo.NewSessionRepository(queries)

	// 3. Create WebSocket Hub
	wsHub := ws.NewHub()
	go wsHub.Run() // Start hub in background goroutine

	// 4. Create Use Cases (inject WebSocket hub as broadcaster)
	joinUsecase := command.NewJoinCommandUseCase(userRepository, sessionRepository, wsHub)

	// 5. Create HTTP Handlers
	commandHandler := handler.NewCommandHandler(joinUsecase)
	wsHandler := ws.NewHandler(wsHub)

	// 6. Setup Router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Register WebSocket endpoint
	r.Get("/ws", wsHandler.ServeWS)

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
