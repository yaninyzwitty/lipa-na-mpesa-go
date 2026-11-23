package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yw/lipa-na-mpesa-go/internal/c2b"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /c2b/validation", c2b.HandleValidation)
	mux.HandleFunc("POST /c2b/confirmation", c2b.HandleConfirmation)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ListenAndServe()
	}()
	slog.Info("C2B server listening on port :8080")

	select {
	case err := <-serverErr:
		slog.Error("Server stopped unexpectedly", "error", err)
		os.Exit(1)
	case <-ctx.Done():
		slog.Info("Shutting down C2B server")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Graceful shutdown failed", "error", err)
		_ = server.Close()
		os.Exit(1)
	}
	<-serverErr
	slog.Info("C2B server stopped")
}
