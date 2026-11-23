package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nitin1chandani/payments/internal/platform/postgres"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	dsn := os.Getenv("WALLET_DATABASE_URL")
	if dsn == "" {
		// local development default
		dsn = "postgres://postgres:postgres@localhost:5432/wallet?sslmode=disable"
	}

	pool, err := postgres.NewPool(context.Background(), dsn)
	if err != nil {
		logger.Error("connect to postgres", "err", err)
		os.Exit(1)
	}

	logger.Info("connected to postgres", "max_conns", pool.Config().MaxConns)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("health check", "remote", r.RemoteAddr)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}` + "\n"))
	})

	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		w.Header().Set("Content-Type", "application/json")
		if err := pool.Ping(ctx); err != nil {
			logger.Warn("not ready: database unreachable", "err", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"service":"unavailable"}` + "\n"))
			return
		}
		w.Write([]byte(`{"service":"ready"}`))
	})

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("wallet service listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	// close the server

	// waiting for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// inside shutdown we have ctx.done it blocks there, after timeout code will be continued
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown failed", "err", err)
	}

	logger.Info("server stopped")

	// close the pool
	pool.Close()
	logger.Info("database pool closed")

}
