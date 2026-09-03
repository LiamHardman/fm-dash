package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

// weightsGetHandler handles GET /internal/weights: returns every category's
// currently active attribute weights, map[string]map[string]int.
func weightsGetHandler(store *WeightsStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		weights, err := store.GetAll(r.Context())
		if err != nil {
			slog.Error("reading weights failed", "error", err)
			http.Error(w, "reading weights failed", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(weights); err != nil {
			slog.Error("encoding weights response failed", "error", err)
		}
	}
}

// weightsPutHandler handles PUT /internal/weights: category-replace upsert,
// matching src/api/config.go's SetAttributeWeights semantics exactly. Body:
// {"PAC": {"Acc": 8, "Pac": 8, "Agi": 5}, ...}. Responds 200 with the full
// resulting weights map (mirroring POST /api/config's response shape).
func weightsPutHandler(store *WeightsStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var update map[string]map[string]int
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		if err := store.SetCategories(r.Context(), update); err != nil {
			var verr *WeightsValidationError
			if errors.As(err, &verr) {
				http.Error(w, verr.Error(), http.StatusBadRequest)
				return
			}
			slog.Error("updating weights failed", "error", err)
			http.Error(w, "updating weights failed", http.StatusInternalServerError)
			return
		}

		weights, err := store.GetAll(r.Context())
		if err != nil {
			slog.Error("reading weights after update failed", "error", err)
			http.Error(w, "reading weights after update failed", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(weights); err != nil {
			slog.Error("encoding weights response failed", "error", err)
		}
	}
}
