package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Scout Report persistence & the Scouting Book — implements the design charted in
// .scratch/scout-report-v2/issues/01-persisted-storage-and-endpoint-design.md and
// 02-scouting-book-page-ui-ux.md (Wayfinder map "Scout Report v2"). One blob per
// (player, position) report, plus a separately-maintained lightweight index blob the
// Scouting Book page reads in one call — not one big per-dataset blob like wishlist,
// since generation is server-side and reports are heavier than wishlist entries.

const scoutReportBookKeyPrefix = "scout-reports/"

func scoutReportStorageKey(datasetID string, playerUID int64, position string) string {
	return fmt.Sprintf("%s%s/%d-%s.json", scoutReportBookKeyPrefix, datasetID, playerUID, position)
}

func scoutReportIndexStorageKey(datasetID string) string {
	return fmt.Sprintf("%s%s/_index.json", scoutReportBookKeyPrefix, datasetID)
}

// PersistedScoutReport is the on-disk shape of one saved report.
type PersistedScoutReport struct {
	PlayerUID   int64               `json:"playerUid"`
	Position    string              `json:"position"`
	GeneratedAt time.Time           `json:"generatedAt"`
	Report      ScoutReportResponse `json:"report"`
}

// ScoutReportWithMeta is the wire shape for a single report — the same flat fields the
// frontend already expects from ScoutReportResponse, with generatedAt added at the top
// level (not nested) so existing template bindings don't need a report.report.* rewrite.
type ScoutReportWithMeta struct {
	ScoutReportResponse
	GeneratedAt time.Time `json:"generatedAt"`
}

// ScoutingBookEntry is one row of the Scouting Book's index. PlayerName/Club are snapshot
// at scout time rather than re-resolved live: if a scouted player later transfers clubs,
// the row should still read as "this is what led me to consider them", not silently drift.
type ScoutingBookEntry struct {
	PlayerUID     int64     `json:"playerUid"`
	PlayerName    string    `json:"playerName"`
	Club          string    `json:"club"`
	Position      string    `json:"position"`
	Grade         string    `json:"grade"`
	SquadStars    float64   `json:"squadStars"`
	DivisionStars float64   `json:"divisionStars"`
	GeneratedAt   time.Time `json:"generatedAt"`
}

// scoutReportIndexMu guards the index blob's read-modify-write cycle. One mutex across
// all datasets: expected concurrency here is low (BYOK single-key requests) and the index
// blob itself is tiny. Only protects same-process races — localConfigStorage already
// serializes all configStorage calls behind its own mutex regardless, and the S3 backend
// has no cross-replica protection, same pre-existing caveat wishlist's whole-blob PUT
// already lives with. Worst case on a lost race is a dropped/stale index row; the
// authoritative per-report blob is a single unguarded key-write, never at risk.
var scoutReportIndexMu sync.Mutex

func loadScoutReportIndex(datasetID string) ([]ScoutingBookEntry, error) {
	if configStorage == nil {
		return nil, http.ErrNoLocation
	}
	data, err := configStorage.RetrieveConfig(scoutReportIndexStorageKey(datasetID))
	if err != nil {
		if os.IsNotExist(err) {
			return []ScoutingBookEntry{}, nil
		}
		return nil, err
	}
	var entries []ScoutingBookEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func writeScoutReportIndex(datasetID string, entries []ScoutingBookEntry) error {
	encoded, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	return configStorage.StoreConfig(scoutReportIndexStorageKey(datasetID), encoded)
}

// upsertScoutReportIndexEntry replaces any existing row for the same (playerUID,
// position) or appends a new one, then writes the index back.
func upsertScoutReportIndexEntry(datasetID string, entry ScoutingBookEntry) error {
	scoutReportIndexMu.Lock()
	defer scoutReportIndexMu.Unlock()

	entries, err := loadScoutReportIndex(datasetID)
	if err != nil {
		return err
	}
	replaced := false
	for i := range entries {
		if entries[i].PlayerUID == entry.PlayerUID && entries[i].Position == entry.Position {
			entries[i] = entry
			replaced = true
			break
		}
	}
	if !replaced {
		entries = append(entries, entry)
	}
	return writeScoutReportIndex(datasetID, entries)
}

func removeScoutReportIndexEntry(datasetID string, playerUID int64, position string) error {
	scoutReportIndexMu.Lock()
	defer scoutReportIndexMu.Unlock()

	entries, err := loadScoutReportIndex(datasetID)
	if err != nil {
		return err
	}
	out := entries[:0]
	for _, e := range entries {
		if e.PlayerUID == playerUID && e.Position == position {
			continue
		}
		out = append(out, e)
	}
	return writeScoutReportIndex(datasetID, out)
}

// saveScoutReport persists one report's full blob and upserts its Scouting Book index
// row. Called by both scoutReportPostHandler (HTTP path) and the chatbot's
// generate_scout_report tool via generateScoutReport, so persistence is never duplicated.
func saveScoutReport(datasetID string, playerUID int64, position, playerName, club string, report ScoutReportResponse, squadStars, divisionStars float64, generatedAt time.Time) error {
	if configStorage == nil {
		return http.ErrNoLocation
	}

	persisted := PersistedScoutReport{
		PlayerUID:   playerUID,
		Position:    position,
		GeneratedAt: generatedAt,
		Report:      report,
	}
	encoded, err := json.Marshal(persisted)
	if err != nil {
		return err
	}
	if err := configStorage.StoreConfig(scoutReportStorageKey(datasetID, playerUID, position), encoded); err != nil {
		return err
	}

	return upsertScoutReportIndexEntry(datasetID, ScoutingBookEntry{
		PlayerUID:     playerUID,
		PlayerName:    playerName,
		Club:          club,
		Position:      position,
		Grade:         report.SubjectGrade,
		SquadStars:    squadStars,
		DivisionStars: divisionStars,
		GeneratedAt:   generatedAt,
	})
}

func loadPersistedScoutReport(datasetID string, playerUID int64, position string) (*PersistedScoutReport, error) {
	if configStorage == nil {
		return nil, http.ErrNoLocation
	}
	data, err := configStorage.RetrieveConfig(scoutReportStorageKey(datasetID, playerUID, position))
	if err != nil {
		return nil, err
	}
	var persisted PersistedScoutReport
	if err := json.Unmarshal(data, &persisted); err != nil {
		return nil, err
	}
	return &persisted, nil
}

// deleteScoutReport removes the report blob first, then its index row. If the process
// dies between the two, a stale index row can reference a now-missing blob — the GET
// single-report endpoint's 404 in that case is treated as "already removed" and dropped
// client-side (self-healing on the Scouting Book's next load), not surfaced as an error.
func deleteScoutReport(datasetID string, playerUID int64, position string) error {
	if configStorage == nil {
		return http.ErrNoLocation
	}
	if err := configStorage.DeleteConfig(scoutReportStorageKey(datasetID, playerUID, position)); err != nil {
		return err
	}
	return removeScoutReportIndexEntry(datasetID, playerUID, position)
}

// --- HTTP handlers ---

// scoutReportGetHandler serves GET /api/scout-report/{datasetId}?playerUid=&position=,
// dispatched from scoutReportHandler (scout_report.go). Used by ScoutReportTab.vue's
// backend-GET-first auto-load (Scout Report v2 map ticket 04).
func scoutReportGetHandler(w http.ResponseWriter, r *http.Request, datasetID string) {
	playerUID, position, ok := parseScoutReportQueryParams(w, r)
	if !ok {
		return
	}

	persisted, err := loadPersistedScoutReport(datasetID, playerUID, position)
	if err != nil {
		WriteErrorResponse(w, r, "scout_report_not_found", "No saved scout report for this player/position", nil, http.StatusNotFound)
		return
	}

	setCORSHeaders(w, r)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(ScoutReportWithMeta{ScoutReportResponse: persisted.Report, GeneratedAt: persisted.GeneratedAt}); err != nil {
		LogWarn("scout-report: failed to write GET response: %v", err)
	}
}

// scoutReportDeleteHandler serves DELETE /api/scout-report/{datasetId}?playerUid=&position=
// — the Scouting Book's row-remove action, no confirmation required (low stakes,
// regeneratable).
func scoutReportDeleteHandler(w http.ResponseWriter, r *http.Request, datasetID string) {
	playerUID, position, ok := parseScoutReportQueryParams(w, r)
	if !ok {
		return
	}

	if err := deleteScoutReport(datasetID, playerUID, position); err != nil {
		LogWarn("scout-report: failed to delete report for dataset %s: %v", sanitizeForLogging(datasetID), err)
		WriteErrorResponse(w, r, "storage_error", "Failed to remove scout report", nil, http.StatusInternalServerError)
		return
	}
	setCORSHeaders(w, r)
	w.WriteHeader(http.StatusNoContent)
}

func parseScoutReportQueryParams(w http.ResponseWriter, r *http.Request) (playerUID int64, position string, ok bool) {
	playerUIDStr := r.URL.Query().Get("playerUid")
	position = r.URL.Query().Get("position")
	if playerUIDStr == "" || position == "" {
		WriteErrorResponse(w, r, "missing_params", "playerUid and position are both required", nil, http.StatusBadRequest)
		return 0, "", false
	}
	playerUID, err := strconv.ParseInt(playerUIDStr, 10, 64)
	if err != nil {
		WriteErrorResponse(w, r, "invalid_player_uid", "playerUid must be an integer", nil, http.StatusBadRequest)
		return 0, "", false
	}
	return playerUID, position, true
}

// scoutReportsBookHandler serves GET /api/scout-reports/{datasetId} (plural, deliberately
// distinct from the singular /api/scout-report/{datasetId}) — the Scouting Book page's
// entire data source in one call.
func scoutReportsBookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteErrorResponse(w, r, "method_not_allowed", "Only GET method is allowed", nil, http.StatusMethodNotAllowed)
		return
	}
	datasetID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/scout-reports/"), "/")
	if datasetID == "" {
		WriteErrorResponse(w, r, "invalid_dataset_id", "Dataset ID is missing in the request path", nil, http.StatusBadRequest)
		return
	}

	entries, err := loadScoutReportIndex(datasetID)
	if err != nil {
		WriteErrorResponse(w, r, "storage_error", "Failed to load Scouting Book", nil, http.StatusInternalServerError)
		return
	}

	setCORSHeaders(w, r)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(entries); err != nil {
		LogWarn("scout-reports: failed to write response: %v", err)
	}
}
