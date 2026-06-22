// Package main is the entry point for the log query API service.
//
// It loads configuration from environment variables, initializes a ClickHouse
// connection pool, wires up the HTTP router with middleware, and starts the
// server with graceful shutdown support.
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

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/your-username/log-platform/config"
	"github.com/your-username/log-platform/handlers"
	"github.com/your-username/log-platform/middleware"
	"github.com/your-username/log-platform/repository"
)

func main() {
	// ---------------------------------------------------------------
	// Configuration
	// ---------------------------------------------------------------
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// ---------------------------------------------------------------
	// ClickHouse connection pool
	// ---------------------------------------------------------------
	repo, err := repository.New(
		cfg.ClickHouseAddr,
		cfg.ClickHouseDatabase,
		cfg.ClickHouseUsername,
		cfg.ClickHousePassword,
	)
	if err != nil {
		log.Fatalf("failed to connect to ClickHouse: %v", err)
	}
	defer repo.Close()

	log.Printf("connected to ClickHouse at %s (db=%s)", cfg.ClickHouseAddr, cfg.ClickHouseDatabase)

	// ---------------------------------------------------------------
	// Router & middleware
	// ---------------------------------------------------------------
	r := chi.NewRouter()

	// Middleware chain (order matters).
	r.Use(chimw.RequestID)   // Inject X-Request-Id.
	r.Use(chimw.RealIP)      // Trust X-Forwarded-For / X-Real-IP.
	r.Use(middleware.StructuredLogger()) // JSON request logging (skips /health).
	r.Use(chimw.Recoverer)   // Catch panics, return 500.
	r.Use(middleware.APIKeyAuth(cfg.APIKey))
	r.Use(middleware.RateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst))

	// ---------------------------------------------------------------
	// Routes
	// ---------------------------------------------------------------
	r.Get("/health", handlers.Health())
	r.Get("/ready", handlers.Ready(repo))

	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/search", handlers.Search(repo))
		api.Get("/trace/{traceId}", handlers.Trace(repo))
		api.Get("/stats", handlers.Stats(repo))
		api.Get("/stats/services", handlers.Services(repo))
	})

	// ---------------------------------------------------------------
	// HTTP server with timeouts
	// ---------------------------------------------------------------
	addr := fmt.Sprintf(":%d", cfg.APIPort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  60 * time.Second,
	}

	// ---------------------------------------------------------------
	// Graceful shutdown
	// ---------------------------------------------------------------
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine so we can block on shutdown signals.
	go func() {
		log.Printf("query-api listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Block until we receive a termination signal.
	sig := <-shutdownCh
	log.Printf("received signal %v, shutting down gracefully…", sig)

	// Give in-flight requests up to 15 seconds to complete.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}

	log.Println("server stopped")
}
