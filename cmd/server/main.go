// Command server runs the PR Reviewer Assignment Service.
//
// It connects to PostgreSQL, applies migrations, and starts the HTTP server
// on the configured port. The service is designed to be run via docker-compose.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kihcnxlehp/pr-reviewer-service/internal/config"
)

func main() {
	cfg := config.Load()

	startupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to PostgreSQL.
	pool, err := pgxpool.New(startupCtx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(startupCtx); err != nil {
		log.Fatalf("could not ping database: %v", err)
	}
	log.Println("connected to database")

	//Apply database migrations.
	if err := runMigrations(startupCtx, cfg.DatabaseURL); err != nil {
		log.Fatalf("could not run migrations: %v", err)
	}
	log.Println("migrations applied")

	// Set up HTTP server.
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server in a separate goroutine.
	go func() {
		log.Printf("server listening on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("could not listen on port %s: %v", cfg.ServerPort, err)
		}
	}()

	// Wait for termination signal and perform graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("could not shutdown server: %v", err)
	}
	log.Println("server stopped")
}

// runMigrations applies all pending SQL migrations from the migrations/ directory.
// Returns nil if migrations are already up-to-date (migrate.ErrNoChange is ignored).
func runMigrations(ctx context.Context, databaseURL string) error {
	m, err := migrate.New("file://migrations", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	select {
	case <-ctx.Done():
		return fmt.Errorf("migration cancelled: %w", ctx.Err())
	default:
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	return nil
}
