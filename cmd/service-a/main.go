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
	"github.com/victorhugomazzeo/goexpoert_06/internal/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	if err := run(); err != nil {
		slog.Error("failed to run service-a", "error", err)
		os.Exit(1)
	}
}

func run() error {

	_ = godotenv.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownTelemetry, err := telemetry.Init(ctx)
	if err != nil {
		return err
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownTelemetry(ctx); err != nil {
			slog.Error("failed to shutdown telemetry", "error", err)
		}
	}()

	port := os.Getenv("SERVICE_A_PORT")
	if port == "" {
		port = "8080"
	}

	serviceBURL := os.Getenv("SERVICE_B_URL")

	if serviceBURL == "" {
		serviceBURL = "http://localhost:8081"
	}

	httpClient := &http.Client{
		Timeout:   5 * time.Second,
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
	client := gateway.NewWeatherClient(serviceBURL, httpClient)
	handler := gateway.NewHandler(client)

	mux := http.NewServeMux()
	mux.Handle("POST /weather", otelhttp.NewHandler(handler, "POST /weather"))
	ws := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

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
