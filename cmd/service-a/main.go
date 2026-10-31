package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/victorhugomazzeo/goexpoert_06/internal/gateway"
)

func main() {
	if err := run(); err != nil {
		slog.Error("failed to run service-a", "error", err)
		os.Exit(1)
	}
}

func run() error {

	_ = godotenv.Load()

	port := os.Getenv("SERVICE_A_PORT")
	if port == "" {
		port = "8080"
	}

	serviceBURL := os.Getenv("SERVICE_B_URL")

	if serviceBURL == "" {
		serviceBURL = "http://localhost:8081"
	}

	httpClient := &http.Client{Timeout: 5 * time.Second}
	client := gateway.NewWeatherClient(serviceBURL, httpClient)
	handler := gateway.NewHandler(client)

	mux := http.NewServeMux()
	mux.Handle("POST /weather", handler)
	ws := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)

	go func() {
		slog.Info("service-a is running", "port", port)
		errCh <- ws.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		slog.Info("server is shutting down")
	case err := <-errCh:
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return ws.Shutdown(shutdownCtx)
}
