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

	"github.com/joho/godotenv"
	"github.com/victorhugomazzeo/goexpoert_06/internal/telemetry"
	"github.com/victorhugomazzeo/goexpoert_06/internal/viacep"
	"github.com/victorhugomazzeo/goexpoert_06/internal/weather"
	"github.com/victorhugomazzeo/goexpoert_06/internal/weatherapi"
)

func main() {
	if err := run(); err != nil {
		slog.Error("failed to run service-b", "error", err)
		os.Exit(1)
	}
}

func run() error {

	_ = godotenv.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shuwdownTelemetry, err := telemetry.Init(ctx)
	if err != nil {
		return err
	}

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shuwdownTelemetry(ctx); err != nil {
			slog.Error("failed to shutdown telemetry", "error", err)
		}
	}()

	port := os.Getenv("SERVICE_B_PORT")
	if port == "" {
		port = "8081"
	}

	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		return errors.New("WEATHER_API_KEY is required")
	}

	httpClient := &http.Client{Timeout: 5 * time.Second}
	citySearcher := viacep.NewClient(httpClient)
	tempSearcher := weatherapi.NewClient(apiKey, httpClient)
	service := weather.NewService(tempSearcher, citySearcher)
	handler := weather.NewHandler(service)

	mux := http.NewServeMux()
	mux.Handle("GET /weather", handler)
	ws := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)

	go func() {
		slog.Info("service-b is running", "port", port)
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
