package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

type materializeRequest struct {
	Players []PlayerRow `json:"players"`
}

// materializeHandler handles POST /internal/materialize/{datasetId}: it
// accepts a Dataset's raw player data and writes its Query artifact. It is
// standalone in this phase — not yet called by the existing upload pipeline.
func materializeHandler(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		datasetID := r.PathValue("datasetId")
		if datasetID == "" || strings.ContainsAny(datasetID, "/\\.") {
			http.Error(w, "invalid datasetId", http.StatusBadRequest)
			return
		}

		var body materializeRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		if err := writeParquet(r.Context(), cfg.DatasetStorageDir, datasetID, body.Players); err != nil {
			slog.Error("materialize failed", "dataset_id", datasetID, "error", err)
			http.Error(w, "materialization failed", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}
