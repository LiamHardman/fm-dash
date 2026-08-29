package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

const shortlistStorageKey = "shortlists.json"

var shortlistStatuses = map[string]bool{
	"watching": true, "scouting": true, "bid_planned": true, "signed": true, "not_pursuing": true,
}

var shortlistPriorities = map[string]bool{
	"high": true, "medium": true, "low": true,
}

// ShortlistDocument is the installation-level recruitment board. It deliberately
// contains player references, rather than copies of Dataset player data.
type ShortlistDocument struct {
	Version int         `json:"version"`
	Lists   []Shortlist `json:"lists"`
}

type Shortlist struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Items []ShortlistItem `json:"items"`
}

type ShortlistItem struct {
	PlayerRef  ShortlistPlayerRef `json:"playerRef"`
	Status     string             `json:"status"`
	Priority   string             `json:"priority"`
	Tags       []string           `json:"tags"`
	Notes      string             `json:"notes"`
	TargetFee  float64            `json:"targetFee"`
	TargetWage float64            `json:"targetWage"`
}

type ShortlistPlayerRef struct {
	DatasetID string `json:"datasetId"`
	PlayerUID int64  `json:"playerUid"`
}

func defaultShortlistDocument() ShortlistDocument {
	return ShortlistDocument{
		Version: 1,
		Lists:   []Shortlist{{ID: "default", Name: "My shortlist", Items: []ShortlistItem{}}},
	}
}

func shortlistsHandler(w http.ResponseWriter, r *http.Request) {
	if configStorage == nil {
		WriteErrorResponse(w, r, "storage_unavailable", "Shortlist storage is not initialised", nil, http.StatusServiceUnavailable)
		return
	}

	switch r.Method {
	case http.MethodGet:
		shortlistsGetHandler(w, r)
	case http.MethodPut:
		shortlistsPutHandler(w, r)
	default:
		WriteErrorResponse(w, r, "method_not_allowed", "Method not allowed", nil, http.StatusMethodNotAllowed)
	}
}

func shortlistsGetHandler(w http.ResponseWriter, r *http.Request) {
	data, err := configStorage.RetrieveConfig(shortlistStorageKey)
	if err != nil {
		writeShortlistsResponse(w, defaultShortlistDocument())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		LogWarn("shortlists: failed to write GET response: %v", err)
	}
}

func shortlistsPutHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 2*1024*1024))
	if err != nil {
		WriteErrorResponse(w, r, "read_error", "Failed to read request body", nil, http.StatusBadRequest)
		return
	}

	var document ShortlistDocument
	if err := json.Unmarshal(body, &document); err != nil {
		WriteErrorResponse(w, r, "invalid_json", "Request body must be valid JSON", nil, http.StatusBadRequest)
		return
	}
	if err := validateShortlistDocument(&document); err != nil {
		WriteErrorResponse(w, r, "invalid_shortlist", err.Error(), nil, http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(document)
	if err != nil {
		WriteErrorResponse(w, r, "serialization_error", "Failed to save shortlist", nil, http.StatusInternalServerError)
		return
	}
	if err := configStorage.StoreConfig(shortlistStorageKey, data); err != nil {
		LogWarn("shortlists: failed to store: %v", err)
		WriteErrorResponse(w, r, "storage_error", "Failed to save shortlist", nil, http.StatusInternalServerError)
		return
	}

	writeShortlistsResponse(w, document)
}

func writeShortlistsResponse(w http.ResponseWriter, document ShortlistDocument) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(document); err != nil {
		LogWarn("shortlists: failed to write response: %v", err)
	}
}

func validateShortlistDocument(document *ShortlistDocument) error {
	if document.Version != 1 || len(document.Lists) == 0 || len(document.Lists) > 50 {
		return errInvalidShortlist
	}
	seenLists := make(map[string]bool)
	for listIndex := range document.Lists {
		list := &document.Lists[listIndex]
		list.ID = strings.TrimSpace(list.ID)
		list.Name = strings.TrimSpace(list.Name)
		if list.ID == "" || list.Name == "" || len(list.Name) > 80 || seenLists[list.ID] || len(list.Items) > 2000 {
			return errInvalidShortlist
		}
		seenLists[list.ID] = true
		seenItems := make(map[string]bool)
		for itemIndex := range list.Items {
			item := &list.Items[itemIndex]
			item.PlayerRef.DatasetID = strings.TrimSpace(item.PlayerRef.DatasetID)
			item.Status = strings.TrimSpace(item.Status)
			item.Priority = strings.TrimSpace(item.Priority)
			item.Notes = strings.TrimSpace(item.Notes)
			if _, err := uuid.Parse(item.PlayerRef.DatasetID); err != nil || item.PlayerRef.PlayerUID <= 0 || !shortlistStatuses[item.Status] || !shortlistPriorities[item.Priority] || len(item.Tags) > 20 || len(item.Notes) > 2000 || item.TargetFee < 0 || item.TargetWage < 0 {
				return errInvalidShortlist
			}
			identity := fmt.Sprintf("%s:%d", item.PlayerRef.DatasetID, item.PlayerRef.PlayerUID)
			if seenItems[identity] {
				return errInvalidShortlist
			}
			seenItems[identity] = true
			for tagIndex, tag := range item.Tags {
				item.Tags[tagIndex] = strings.TrimSpace(tag)
				if item.Tags[tagIndex] == "" || len(item.Tags[tagIndex]) > 40 {
					return errInvalidShortlist
				}
			}
		}
	}
	return nil
}
