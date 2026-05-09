package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/charbelhanna96/safelyyou-fleet-monitor/internal/api/handlers"
	"github.com/charbelhanna96/safelyyou-fleet-monitor/internal/api/middleware"
	"github.com/charbelhanna96/safelyyou-fleet-monitor/internal/config"
	"github.com/charbelhanna96/safelyyou-fleet-monitor/internal/device"
	"github.com/charbelhanna96/safelyyou-fleet-monitor/internal/store"
)

func main() {
	cfg := config.Load()

	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		logLevel = slog.LevelInfo
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})))

	devices, err := device.LoadDevicesCSV(cfg.AppConfig.CSVPath)
	if err != nil {
		slog.Error("Failed to load devices", slog.String("error", err.Error()))
		os.Exit(1)
	}

	memStore := store.NewMemoryStore()
	memStore.AddDevices(devices)
	handler := handlers.NewHandler(memStore)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/devices/{device_id}/stats", handler.GetStats)
	mux.HandleFunc("POST /api/v1/devices/{device_id}/heartbeat", handler.PostHeartbeat)
	mux.HandleFunc("POST /api/v1/devices/{device_id}/stats", handler.PostUploadStat)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      middleware.LoggingMiddleware(mux),
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	go func() {
		slog.Info("listening", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	slog.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}
