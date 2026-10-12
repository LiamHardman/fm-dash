package main

import (
	"encoding/json"
	"errors"
	"net/http"

	apperrors "api/errors"
	"github.com/google/uuid"
)

type resolvedPlayerReference struct {
	PlayerRef ShortlistPlayerRef `json:"playerRef"`
	State     string             `json:"state"`
	Player    *referencedPlayer  `json:"player,omitempty"`
}
type referencedPlayer struct {
	UID             int64    `json:"uid"`
	Name            string   `json:"name"`
	Club            string   `json:"club"`
	Position        string   `json:"position"`
	ParsedPositions []string `json:"parsedPositions"`
}

// Resolve live data without copying it into durable profile documents.
func playerReferencesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request struct {
		References []ShortlistPlayerRef `json:"references"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 256*1024)).Decode(&request); err != nil || len(request.References) > 2000 {
		WriteErrorResponse(w, r, "invalid_references", "Invalid Player references", nil, http.StatusBadRequest)
		return
	}
	groups := map[string]map[int64]*referencedPlayer{}
	states := map[string]string{}
	for _, ref := range request.References {
		if _, err := uuid.Parse(ref.DatasetID); err != nil || ref.PlayerUID <= 0 {
			WriteErrorResponse(w, r, "invalid_references", "Invalid Player reference", nil, http.StatusBadRequest)
			return
		}
		if groups[ref.DatasetID] == nil {
			groups[ref.DatasetID] = map[int64]*referencedPlayer{}
		}
		groups[ref.DatasetID][ref.PlayerUID] = nil
	}
	for id, wanted := range groups {
		if r.Context().Err() != nil {
			return
		}
		if storage == nil {
			states[id] = "failed"
			continue
		}
		players, _, err := RetrieveDataset(id)
		if err != nil {
			states[id] = "failed"
			if errors.Is(err, apperrors.ErrDatasetNotFound) {
				states[id] = "expired"
			}
			continue
		}
		states[id] = "available"
		for _, player := range players {
			if _, ok := wanted[player.UID]; ok {
				wanted[player.UID] = &referencedPlayer{player.UID, player.Name, player.Club, player.Position, player.ParsedPositions}
			}
		}
	}
	items := make([]resolvedPlayerReference, 0, len(request.References))
	for _, ref := range request.References {
		item := resolvedPlayerReference{PlayerRef: ref, State: states[ref.DatasetID], Player: groups[ref.DatasetID][ref.PlayerUID]}
		if item.State == "available" && item.Player == nil {
			item.State = "missing"
		}
		items = append(items, item)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(map[string]any{"items": items})
}
