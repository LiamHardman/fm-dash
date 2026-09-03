package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

type playerFifaStats struct {
	PlayerID int64 `json:"playerId"`
	PAC      int   `json:"pac"`
	SHO      int   `json:"sho"`
	PAS      int   `json:"pas"`
	DRI      int   `json:"dri"`
	DEF      int   `json:"def"`
	PHY      int   `json:"phy"`
	GK       int   `json:"gk"`
	DIV      int   `json:"div"`
	HAN      int   `json:"han"`
	REF      int   `json:"ref"`
	KIC      int   `json:"kic"`
	SPD      int   `json:"spd"`
	POS      int   `json:"pos"`
}

type fifaStatsResponse struct {
	Players []playerFifaStats `json:"players"`
}

// fifaStatsHandler handles GET /internal/query/{datasetId}/fifa-stats: it
// computes every applicable FIFA-style stat for every player in datasetID's
// Query artifact, joined against the currently active weights in
// weights.duckdb. Standalone in this phase -- not yet called by src/api.
func fifaStatsHandler(cfg Config, store *WeightsStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		datasetID := r.PathValue("datasetId")
		if datasetID == "" || strings.ContainsAny(datasetID, "/\\.") {
			http.Error(w, "invalid datasetId", http.StatusBadRequest)
			return
		}

		result, err := computeFifaStats(r.Context(), store, cfg.DatasetStorageDir, datasetID)
		if err != nil {
			if errors.Is(err, ErrArtifactNotFound) {
				http.Error(w, "dataset artifact not found", http.StatusNotFound)
				return
			}
			slog.Error("fifa stats query failed", "dataset_id", datasetID, "error", err)
			http.Error(w, "fifa stats query failed", http.StatusInternalServerError)
			return
		}

		resp := fifaStatsResponse{Players: make([]playerFifaStats, 0, len(result))}
		for playerID, s := range result {
			resp.Players = append(resp.Players, playerFifaStats{
				PlayerID: playerID,
				PAC:      s.PAC, SHO: s.SHO, PAS: s.PAS, DRI: s.DRI, DEF: s.DEF, PHY: s.PHY,
				GK: s.GK, DIV: s.DIV, HAN: s.HAN, REF: s.REF, KIC: s.KIC, SPD: s.SPD, POS: s.POS,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("encoding fifa stats response failed", "dataset_id", datasetID, "error", err)
		}
	}
}
