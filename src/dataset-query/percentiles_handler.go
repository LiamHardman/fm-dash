package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

type playerPercentiles struct {
	PlayerID    int64                         `json:"playerId"`
	Percentiles map[string]map[string]float64 `json:"percentiles"`
}

type percentilesResponse struct {
	Players []playerPercentiles `json:"players"`
}

// percentilesHandler handles GET /internal/query/{datasetId}/percentiles: it
// computes the 3-tier (Global / broad / detailed position group) performance
// percentiles for every player in datasetID's Query artifact, in DuckDB SQL.
// Standalone in this phase — not yet called by the existing backend.
func percentilesHandler(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		datasetID := r.PathValue("datasetId")
		if datasetID == "" || strings.ContainsAny(datasetID, "/\\.") {
			http.Error(w, "invalid datasetId", http.StatusBadRequest)
			return
		}

		result, err := computePercentiles(r.Context(), cfg.DatasetStorageDir, datasetID)
		if err != nil {
			if errors.Is(err, ErrArtifactNotFound) {
				http.Error(w, "dataset artifact not found", http.StatusNotFound)
				return
			}
			slog.Error("percentiles query failed", "dataset_id", datasetID, "error", err)
			http.Error(w, "percentiles query failed", http.StatusInternalServerError)
			return
		}

		resp := percentilesResponse{Players: make([]playerPercentiles, 0, len(result))}
		for playerID, groups := range result {
			resp.Players = append(resp.Players, playerPercentiles{PlayerID: playerID, Percentiles: groups})
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("encoding percentiles response failed", "dataset_id", datasetID, "error", err)
		}
	}
}
