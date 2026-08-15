package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
)

// ProgressionAnalyzeRequest is the body of POST /api/progression/analyze.
type ProgressionAnalyzeRequest struct {
	DatasetIDs []string `json:"datasetIds"`
	// Order is an optional explicit chronological order override (dataset IDs, earliest first).
	// Required once the client has resolved an OrderAmbiguous response.
	Order []string `json:"order,omitempty"`
}

// ProgressionPlayer is a player present in every uploaded snapshot, carrying one full
// Player record per snapshot (in chronological order) so the frontend can reuse the same
// filtering, display, and role-overall logic it already has for a single dataset.
type ProgressionPlayer struct {
	UID       int64    `json:"uid"`
	Name      string   `json:"name"`
	Position  string   `json:"position"`
	Club      string   `json:"club"`
	Snapshots []Player `json:"snapshots"`
}

// ProgressionAnalyzeResponse is the response of POST /api/progression/analyze.
type ProgressionAnalyzeResponse struct {
	Order               []string             `json:"order,omitempty"`
	Players             []ProgressionPlayer  `json:"players,omitempty"`
	OrderAmbiguous      bool                 `json:"orderAmbiguous,omitempty"`
	AmbiguousDatasetIDs []string             `json:"ambiguousDatasetIds,omitempty"`
	EmptyIntersection   bool                 `json:"emptyIntersection,omitempty"`
	CurrencySymbol      string               `json:"currencySymbol,omitempty"`
}

// progressionFieldAccessors resolves a "known interesting field" name to a numeric value
// pulled off a Player. Deliberately a lookup table rather than reflection: Player's numeric
// data is split between typed struct fields and the NumericAttributes map, and a lookup
// table makes that split explicit instead of hiding it behind reflect.
var progressionFieldAccessors = map[string]func(Player) (float64, bool){
	"Overall": func(p Player) (float64, bool) { return float64(p.Overall), true },
	"Value":   func(p Player) (float64, bool) { return float64(p.TransferValueAmount), true },
	"Wage":    func(p Player) (float64, bool) { return float64(p.WageAmount), true },
	"Age": func(p Player) (float64, bool) {
		age, err := strconv.Atoi(p.Age)
		if err != nil {
			return 0, false
		}
		return float64(age), true
	},
}

// progressionFieldValue resolves "field" against the accessor registry first, falling back
// to a direct NumericAttributes lookup (covers every FM attribute key without needing an
// entry per attribute).
func progressionFieldValue(p Player, field string) (float64, bool) {
	if accessor, ok := progressionFieldAccessors[field]; ok {
		return accessor(p)
	}
	if v, ok := p.NumericAttributes[field]; ok {
		return float64(v), true
	}
	return 0, false
}

// progressionDelta computes last-snapshot-value minus first-snapshot-value for a field.
func progressionDelta(p ProgressionPlayer, field string) (float64, bool) {
	first, firstOK := progressionFieldValue(p.Snapshots[0], field)
	last, lastOK := progressionFieldValue(p.Snapshots[len(p.Snapshots)-1], field)
	if !firstOK || !lastOK {
		return 0, false
	}
	return last - first, true
}

// determineSnapshotOrder sorts datasetIDs by mean parseable Age (ascending). Players whose
// Age fails to parse are skipped when computing a dataset's mean; if every player in a
// dataset fails to parse, that dataset's mean is treated as unset. Any group of two or more
// datasets that can't be distinguished (equal means, or more than one "unset" dataset) is
// reported as ambiguous instead of guessed.
func determineSnapshotOrder(datasetIDs []string, datasetPlayers map[string][]Player) (order []string, ambiguous bool, ambiguousIDs []string) {
	type entry struct {
		id  string
		age float64
		ok  bool
	}

	entries := make([]entry, len(datasetIDs))
	for i, id := range datasetIDs {
		sum, count := 0, 0
		for _, p := range datasetPlayers[id] {
			if age, err := strconv.Atoi(p.Age); err == nil {
				sum += age
				count++
			}
		}
		if count > 0 {
			entries[i] = entry{id: id, age: float64(sum) / float64(count), ok: true}
		} else {
			entries[i] = entry{id: id, ok: false}
		}
	}

	groups := make(map[string][]string)
	for _, e := range entries {
		key := "unset"
		if e.ok {
			key = strconv.FormatFloat(e.age, 'f', 6, 64)
		}
		groups[key] = append(groups[key], e.id)
	}

	var ambiguousGroup []string
	for _, ids := range groups {
		if len(ids) > 1 {
			ambiguousGroup = append(ambiguousGroup, ids...)
		}
	}
	if len(ambiguousGroup) > 0 {
		return nil, true, ambiguousGroup
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].age < entries[j].age })
	order = make([]string, len(entries))
	for i, e := range entries {
		order[i] = e.id
	}
	return order, false, nil
}

func progressionAnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		WriteErrorResponse(w, r, "method_not_allowed", "Only POST method is allowed", nil, http.StatusMethodNotAllowed)
		return
	}

	var req ProgressionAnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logError(ctx, "Error decoding progression analyze request", "error", err)
		WriteErrorResponse(w, r, "invalid_request", "Invalid request body", nil, http.StatusBadRequest)
		return
	}

	if len(req.DatasetIDs) < 2 {
		WriteErrorResponse(w, r, "insufficient_datasets", "At least 2 dataset IDs are required", nil, http.StatusBadRequest)
		return
	}

	datasetPlayers := make(map[string][]Player, len(req.DatasetIDs))
	var currencySymbol string
	for _, id := range req.DatasetIDs {
		players, currency, found := GetPlayerData(id)
		if !found {
			logWarn(ctx, "Progression dataset not found", "dataset_id", id)
			WriteErrorResponse(w, r, "dataset_not_found", "Dataset not found: "+id, []string{"datasetId: " + id}, http.StatusNotFound)
			return
		}
		datasetPlayers[id] = players
		if currencySymbol == "" {
			currencySymbol = currency
		}
	}

	var order []string
	orderAmbiguous := false
	var ambiguousIDs []string

	if len(req.Order) == len(req.DatasetIDs) {
		order = req.Order
	} else {
		order, orderAmbiguous, ambiguousIDs = determineSnapshotOrder(req.DatasetIDs, datasetPlayers)
	}

	setCORSHeaders(w, r)
	w.Header().Set("Content-Type", "application/json")

	if orderAmbiguous {
		if err := json.NewEncoder(w).Encode(ProgressionAnalyzeResponse{
			OrderAmbiguous:      true,
			AmbiguousDatasetIDs: ambiguousIDs,
		}); err != nil {
			logError(ctx, "Error encoding progression ambiguous-order response", "error", err)
		}
		return
	}

	// Intersect by UID: keep UIDs present in every ordered dataset.
	uidCounts := make(map[int64]int)
	for _, id := range order {
		seen := make(map[int64]bool, len(datasetPlayers[id]))
		for i := range datasetPlayers[id] {
			uid := datasetPlayers[id][i].UID
			if !seen[uid] {
				uidCounts[uid]++
				seen[uid] = true
			}
		}
	}

	intersectedUIDs := make([]int64, 0)
	for uid, count := range uidCounts {
		if count == len(order) {
			intersectedUIDs = append(intersectedUIDs, uid)
		}
	}

	if len(intersectedUIDs) == 0 {
		if err := json.NewEncoder(w).Encode(ProgressionAnalyzeResponse{
			Order:             order,
			EmptyIntersection: true,
			CurrencySymbol:    currencySymbol,
		}); err != nil {
			logError(ctx, "Error encoding progression empty-intersection response", "error", err)
		}
		return
	}

	indexByUID := make(map[string]map[int64]*Player, len(order))
	for _, id := range order {
		idx := make(map[int64]*Player, len(datasetPlayers[id]))
		for i := range datasetPlayers[id] {
			idx[datasetPlayers[id][i].UID] = &datasetPlayers[id][i]
		}
		indexByUID[id] = idx
	}

	players := make([]ProgressionPlayer, 0, len(intersectedUIDs))
	for _, uid := range intersectedUIDs {
		snapshots := make([]Player, 0, len(order))
		for _, id := range order {
			snapshots = append(snapshots, *indexByUID[id][uid])
		}
		latest := snapshots[len(snapshots)-1]
		players = append(players, ProgressionPlayer{
			UID:       uid,
			Name:      latest.Name,
			Position:  latest.Position,
			Club:      latest.Club,
			Snapshots: snapshots,
		})
	}

	sortField := r.URL.Query().Get("sortField")
	sortDir := r.URL.Query().Get("sortDir")
	if sortField != "" {
		sort.SliceStable(players, func(i, j int) bool {
			di, _ := progressionDelta(players[i], sortField)
			dj, _ := progressionDelta(players[j], sortField)
			if sortDir == "asc" {
				return di < dj
			}
			return di > dj
		})
	} else {
		sort.Slice(players, func(i, j int) bool { return players[i].Name < players[j].Name })
	}

	if err := json.NewEncoder(w).Encode(ProgressionAnalyzeResponse{
		Order:          order,
		Players:        players,
		CurrencySymbol: currencySymbol,
	}); err != nil {
		logError(ctx, "Error encoding progression analyze response", "error", err)
	}
}
