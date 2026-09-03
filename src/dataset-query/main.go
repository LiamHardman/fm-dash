// Command dataset-query-server runs the Dataset Query Service: a single,
// shared process embedding DuckDB that materializes and (in a later phase)
// answers queries against FM-Dash Dataset player data.
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
)

func main() {
	cfg := loadConfig()

	if err := os.MkdirAll(cfg.DatasetStorageDir, 0o755); err != nil {
		slog.Error("failed to create dataset storage directory", "dir", cfg.DatasetStorageDir, "error", err)
		os.Exit(1)
	}

	weightsStore, err := openWeightsStore(context.Background(), cfg.DatasetStorageDir)
	if err != nil {
		slog.Error("failed to open weights store", "error", err)
		os.Exit(1)
	}
	defer func() { _ = weightsStore.Close() }()

	if err := weightsStore.seedDefaultsIfEmpty(context.Background()); err != nil {
		slog.Error("failed to seed default weights", "error", err)
		os.Exit(1)
	}

	if err := weightsStore.seedRoleWeightsIfEmpty(context.Background()); err != nil {
		slog.Error("failed to seed default role weights", "error", err)
		os.Exit(1)
	}

	if err := weightsStore.seedCasWeightsIfEmpty(context.Background()); err != nil {
		slog.Error("failed to seed default cas weights", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("POST /internal/materialize/{datasetId}", materializeHandler(cfg))
	mux.HandleFunc("GET /internal/query/{datasetId}/percentiles", percentilesHandler(cfg))
	mux.HandleFunc("GET /internal/query/{datasetId}/fifa-stats", fifaStatsHandler(cfg, weightsStore))
	mux.HandleFunc("GET /internal/query/{datasetId}/role-overalls", roleOverallsHandler(cfg, weightsStore))
	mux.HandleFunc("GET /internal/query/{datasetId}/cas", casHandler(cfg, weightsStore))
	mux.HandleFunc("GET /internal/query/{datasetId}/mbr", mbrHandler(cfg, weightsStore))
	mux.HandleFunc("GET /internal/weights", weightsGetHandler(weightsStore))
	mux.HandleFunc("PUT /internal/weights", weightsPutHandler(weightsStore))

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      130 * time.Second, // materializing a full dataset can take a while
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		slog.Info("dataset query service listening", "port", cfg.Port, "storage_dir", cfg.DatasetStorageDir)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	slog.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
	}
}
