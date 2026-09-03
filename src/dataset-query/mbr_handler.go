package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

type playerMBR struct {
	PlayerID int64 `json:"playerId"`
	MBR      int   `json:"mbr"`
}

type mbrResponse struct {
	Players []playerMBR `json:"players"`
}

// mbrHandler handles GET /internal/query/{datasetId}/mbr: computes the
// final, dataset-normalized Moneyball Rating (MBR) for every player in
// datasetID's Query artifact, joined against the currently active
// role_weights in weights.duckdb. Standalone in this phase -- not yet
// called by src/api.
func mbrHandler(cfg Config, store *WeightsStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		datasetID := r.PathValue("datasetId")
		if datasetID == "" || strings.ContainsAny(datasetID, "/\\.") {
			http.Error(w, "invalid datasetId", http.StatusBadRequest)
			return
		}

		result, err := computeMBR(r.Context(), store, cfg.DatasetStorageDir, datasetID)
		if err != nil {
			if errors.Is(err, ErrArtifactNotFound) {
				http.Error(w, "dataset artifact not found", http.StatusNotFound)
				return
			}
			slog.Error("mbr query failed", "dataset_id", datasetID, "error", err)
			http.Error(w, "mbr query failed", http.StatusInternalServerError)
			return
		}

		resp := mbrResponse{Players: make([]playerMBR, 0, len(result))}
		for playerID, mbr := range result {
			resp.Players = append(resp.Players, playerMBR{PlayerID: playerID, MBR: mbr})
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("encoding mbr response failed", "dataset_id", datasetID, "error", err)
		}
	}
}
