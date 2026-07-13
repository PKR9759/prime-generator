// primeserver is an HTTP server for generating prime numbers.
// provides a REST API at /primes and /executions backed by SQLite.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"prime-generator/internal/server"
	"prime-generator/internal/storage"
)

func main() {
	// setup JSON logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// init the in-memory SQLite database
	ctx := context.Background()
	db, err := storage.NewDB(ctx)
	if err != nil {
		slog.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	// db.Close() is deferred after shutdown completes (see below).

	// setup router with all routes and middleware
	router := server.NewRouter(db)

	// create HTTP server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// start the server in background
	go func() {
		slog.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// wait for shutdown signal (SIGINT, SIGTERM)
	sigCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-sigCtx.Done()
	slog.Info("shutting down", "signal", sigCtx.Err())

	// graceful shutdown with 10s timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}

	// close db after server has fully stopped
	db.Close()
	slog.Info("server stopped")
}
