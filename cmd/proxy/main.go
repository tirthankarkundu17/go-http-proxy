package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tirthankarkundu17/go-http-proxy/internal/config"
	"tirthankarkundu17/go-http-proxy/internal/proxy"
)

func main() {
	// Initialize structured logging (Industry standard)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize proxy handler
	proxyHandler := proxy.NewHandler(cfg.DefaultHeaders)

	// Setup HTTP server with explicit timeouts (Crucial for production)
	mux := http.NewServeMux()
	mux.Handle("/", proxyHandler)

	server := &http.Server{
		Addr:         cfg.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for errors coming from the listener
	serverErrors := make(chan error, 1)

	// Start the server
	go func() {
		slog.Info("Starting proxy server", "port", cfg.Port)
		serverErrors <- server.ListenAndServe()
	}()

	// Channel to listen for an interrupt or terminate signal from the OS
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking main and waiting for shutdown
	select {
	case err := <-serverErrors:
		slog.Error("Error starting server", "error", err)
		os.Exit(1)

	case sig := <-shutdown:
		slog.Info("Start shutdown", "signal", sig.String())

		// Create context with timeout for graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Asking listener to shut down and shed load
		err := server.Shutdown(ctx)
		if err != nil {
			slog.Error("Graceful shutdown did not complete in time", "error", err)
			err = server.Close()
		}

		if err != nil {
			slog.Error("Could not stop server gracefully", "error", err)
			os.Exit(1)
		}

		slog.Info("Server stopped cleanly")
	}
}
