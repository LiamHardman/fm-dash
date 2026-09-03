package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

type playerCAS struct {
	PlayerID int64 `json:"playerId"`
	CA       int   `json:"ca"`
}

type casResponse struct {
	Players []playerCAS `json:"players"`
}

// casHandler handles GET /internal/query/{datasetId}/cas: computes
// CurrentAbility (CAS) for every player in datasetID's Query artifact,
// joined against the currently active cas_weights in weights.duckdb.
// Standalone in this phase -- not yet called by src/api.
func casHandler(cfg Config, store *WeightsStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		datasetID := r.PathValue("datasetId")
		if datasetID == "" || strings.ContainsAny(datasetID, "/\\.") {
			http.Error(w, "invalid datasetId", http.StatusBadRequest)
			return
		}

		result, err := computeCAS(r.Context(), store, cfg.DatasetStorageDir, datasetID)
		if err != nil {
			if errors.Is(err, ErrArtifactNotFound) {
				http.Error(w, "dataset artifact not found", http.StatusNotFound)
				return
			}
			slog.Error("cas query failed", "dataset_id", datasetID, "error", err)
			http.Error(w, "cas query failed", http.StatusInternalServerError)
			return
		}

		resp := casResponse{Players: make([]playerCAS, 0, len(result))}
		for playerID, ca := range result {
			resp.Players = append(resp.Players, playerCAS{PlayerID: playerID, CA: ca})
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("encoding cas response failed", "dataset_id", datasetID, "error", err)
		}
	}
}
