package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

const savedSearchStorageKey = "saved-searches.json"

type SavedSearchDocument struct {
	Version  int           `json:"version"`
	Searches []SavedSearch `json:"searches"`
}

type SavedSearch struct {
	ID      string          `json:"id"`
	Name    string          `json:"name"`
	Filters json.RawMessage `json:"filters"`
}

func savedSearchesHandler(w http.ResponseWriter, r *http.Request) {
	if configStorage == nil {
		WriteErrorResponse(w, r, "storage_unavailable", "Saved-search storage is not initialised", nil, http.StatusServiceUnavailable)
		return
	}
	switch r.Method {
	case http.MethodGet:
		savedSearchesGetHandler(w, r)
	case http.MethodPut:
		savedSearchesPutHandler(w, r)
	default:
		WriteErrorResponse(w, r, "method_not_allowed", "Method not allowed", nil, http.StatusMethodNotAllowed)
	}
}

func savedSearchesGetHandler(w http.ResponseWriter, r *http.Request) {
	data, err := configStorage.RetrieveConfig(savedSearchStorageKey)
	if err != nil {
		writeSavedSearchesResponse(w, SavedSearchDocument{Version: 1, Searches: []SavedSearch{}})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		LogWarn("saved-searches: failed to write GET response: %v", err)
	}
}

func savedSearchesPutHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 2*1024*1024))
	if err != nil {
		WriteErrorResponse(w, r, "read_error", "Failed to read request body", nil, http.StatusBadRequest)
		return
	}
	var document SavedSearchDocument
	if err := json.Unmarshal(body, &document); err != nil {
		WriteErrorResponse(w, r, "invalid_json", "Request body must be valid JSON", nil, http.StatusBadRequest)
		return
	}
	if err := validateSavedSearchDocument(&document); err != nil {
		WriteErrorResponse(w, r, "invalid_saved_search", err.Error(), nil, http.StatusBadRequest)
		return
	}
	data, err := json.Marshal(document)
	if err != nil {
		WriteErrorResponse(w, r, "serialization_error", "Failed to save searches", nil, http.StatusInternalServerError)
		return
	}
	if err := configStorage.StoreConfig(savedSearchStorageKey, data); err != nil {
		LogWarn("saved-searches: failed to store: %v", err)
		WriteErrorResponse(w, r, "storage_error", "Failed to save searches", nil, http.StatusInternalServerError)
		return
	}
	writeSavedSearchesResponse(w, document)
}

func writeSavedSearchesResponse(w http.ResponseWriter, document SavedSearchDocument) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(document); err != nil {
		LogWarn("saved-searches: failed to write response: %v", err)
	}
}

func validateSavedSearchDocument(document *SavedSearchDocument) error {
	if document.Version != 1 || len(document.Searches) > 200 {
		return errInvalidSavedSearch
	}
	seen := make(map[string]bool)
	for index := range document.Searches {
		search := &document.Searches[index]
		search.ID = strings.TrimSpace(search.ID)
		search.Name = strings.TrimSpace(search.Name)
		if search.ID == "" || search.Name == "" || len(search.Name) > 100 || seen[search.ID] || len(search.Filters) == 0 || len(search.Filters) > 200000 || !json.Valid(search.Filters) {
			return errInvalidSavedSearch
		}
		seen[search.ID] = true
		if containsSavedSearchForbiddenData(search.Filters) {
			return errInvalidSavedSearch
		}
	}
	return nil
}

func containsSavedSearchForbiddenData(data []byte) bool {
	var value any
	if json.Unmarshal(data, &value) != nil {
		return true
	}
	var inspect func(any) bool
	inspect = func(node any) bool {
		switch current := node.(type) {
		case map[string]any:
			for key, child := range current {
				lower := strings.ToLower(key)
				if lower == "datasetid" || lower == "playeruid" || lower == "playerid" || lower == "playerdata" || lower == "snapshot" {
					return true
				}
				if inspect(child) {
					return true
				}
			}
		case []any:
			for _, child := range current {
				if inspect(child) {
					return true
				}
			}
		}
		return false
	}
	return inspect(value)
}
