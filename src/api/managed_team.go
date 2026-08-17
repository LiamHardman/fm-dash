package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
)

// Managed team + division per dataset — backs the AI Scout Report feature's blocking
// setup modal and Who to Sign's team pre-fill. Implements the design charted in
// .scratch/scout-report/issues/01-managed-team-storage-and-prompt.md: mirrors
// wishlist_handler.go's configStorage-backed pattern exactly (not fields on the
// in-memory, periodically-cleaned DatasetData player cache).

const managedTeamKeyPrefix = "managed-team/"

type ManagedTeam struct {
	Club     string `json:"club"`
	Division string `json:"division"`
}

func managedTeamStorageKey(datasetID string) string {
	return fmt.Sprintf("%s%s.json", managedTeamKeyPrefix, datasetID)
}

func managedTeamHandler(w http.ResponseWriter, r *http.Request) {
	datasetID := strings.TrimPrefix(r.URL.Path, "/api/managed-team/")
	datasetID = strings.TrimSuffix(datasetID, "/")

	if _, err := uuid.Parse(datasetID); err != nil {
		WriteErrorResponse(w, r, "invalid_dataset_id", "Invalid dataset ID format", nil, http.StatusBadRequest)
		return
	}

	if configStorage == nil {
		WriteErrorResponse(w, r, "storage_unavailable", "Managed team storage is not initialised", nil, http.StatusServiceUnavailable)
		return
	}

	switch r.Method {
	case http.MethodGet:
		managedTeamGetHandler(w, r, datasetID)
	case http.MethodPut:
		managedTeamPutHandler(w, r, datasetID)
	case http.MethodDelete:
		managedTeamDeleteHandler(w, r, datasetID)
	default:
		WriteErrorResponse(w, r, "method_not_allowed", "Method not allowed", nil, http.StatusMethodNotAllowed)
	}
}

func managedTeamGetHandler(w http.ResponseWriter, r *http.Request, datasetID string) {
	key := managedTeamStorageKey(datasetID)
	data, err := configStorage.RetrieveConfig(key)
	if err != nil {
		if os.IsNotExist(err) {
			WriteErrorResponse(w, r, "managed_team_not_found", "Managed team not set for this dataset", nil, http.StatusNotFound)
			return
		}
		WriteErrorResponse(w, r, "managed_team_not_found", "Managed team not set for this dataset", nil, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		LogWarn("managed-team: failed to write GET response: %v", err)
	}
}

func managedTeamPutHandler(w http.ResponseWriter, r *http.Request, datasetID string) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 64*1024))
	if err != nil {
		WriteErrorResponse(w, r, "read_error", "Failed to read request body", nil, http.StatusBadRequest)
		return
	}

	var team ManagedTeam
	if err := json.Unmarshal(body, &team); err != nil {
		WriteErrorResponse(w, r, "invalid_json", "Request body must be valid JSON", nil, http.StatusBadRequest)
		return
	}
	if team.Club == "" {
		WriteErrorResponse(w, r, "club_required", "club is required", nil, http.StatusBadRequest)
		return
	}

	encoded, err := json.Marshal(team)
	if err != nil {
		WriteErrorResponse(w, r, "encode_error", "Failed to encode managed team", nil, http.StatusInternalServerError)
		return
	}

	if err := configStorage.StoreConfig(managedTeamStorageKey(datasetID), encoded); err != nil {
		LogWarn("managed-team: failed to store managed team for dataset %s: %v", sanitizeForLogging(datasetID), err)
		WriteErrorResponse(w, r, "storage_error", "Failed to save managed team", nil, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(encoded); err != nil {
		LogWarn("managed-team: failed to write PUT response: %v", err)
	}
}

func managedTeamDeleteHandler(w http.ResponseWriter, r *http.Request, datasetID string) {
	if err := configStorage.DeleteConfig(managedTeamStorageKey(datasetID)); err != nil {
		LogWarn("managed-team: failed to delete managed team for dataset %s: %v", sanitizeForLogging(datasetID), err)
		WriteErrorResponse(w, r, "storage_error", "Failed to delete managed team", nil, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
