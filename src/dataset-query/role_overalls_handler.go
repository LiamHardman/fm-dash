package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

type playerRoleScore struct {
	RoleName string `json:"roleName"`
	Score    int    `json:"score"`
}

type playerRoleOveralls struct {
	PlayerID             int64             `json:"playerId"`
	Overall              int               `json:"overall"`
	BestRoleOverall      string            `json:"bestRoleOverall"`
	RoleSpecificOveralls []playerRoleScore `json:"roleSpecificOveralls"`
}

type roleOverallsResponse struct {
	Players []playerRoleOveralls `json:"players"`
}

// roleOverallsHandler handles GET /internal/query/{datasetId}/role-overalls:
// computes, for every player in datasetID's Query artifact, every applicable
// role's score, the mean-of-categories Overall, and BestRoleOverall, joined
// against the currently active weights in weights.duckdb. Standalone in this
// phase -- not yet called by src/api.
func roleOverallsHandler(cfg Config, store *WeightsStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		datasetID := r.PathValue("datasetId")
		if datasetID == "" || strings.ContainsAny(datasetID, "/\\.") {
			http.Error(w, "invalid datasetId", http.StatusBadRequest)
			return
		}

		result, err := computeRoleOveralls(r.Context(), store, cfg.DatasetStorageDir, datasetID)
		if err != nil {
			if errors.Is(err, ErrArtifactNotFound) {
				http.Error(w, "dataset artifact not found", http.StatusNotFound)
				return
			}
			slog.Error("role overalls query failed", "dataset_id", datasetID, "error", err)
			http.Error(w, "role overalls query failed", http.StatusInternalServerError)
			return
		}

		resp := roleOverallsResponse{Players: make([]playerRoleOveralls, 0, len(result))}
		for playerID, ro := range result {
			scores := make([]playerRoleScore, 0, len(ro.RoleSpecificOveralls))
			for _, rs := range ro.RoleSpecificOveralls {
				scores = append(scores, playerRoleScore(rs))
			}
			resp.Players = append(resp.Players, playerRoleOveralls{
				PlayerID:             playerID,
				Overall:              ro.Overall,
				BestRoleOverall:      ro.BestRoleOverall,
				RoleSpecificOveralls: scores,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("encoding role overalls response failed", "dataset_id", datasetID, "error", err)
		}
	}
}
