package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

const progressionProjectStorageKey = "progression-projects.json"

type ProgressionProjectDocument struct {
	Version  int                  `json:"version"`
	Projects []ProgressionProject `json:"projects"`
}

type ProgressionProject struct {
	ID        string                `json:"id"`
	Name      string                `json:"name"`
	OrderMode string                `json:"orderMode"`
	Snapshots []ProgressionSnapshot `json:"snapshots"`
}

type ProgressionSnapshot struct {
	DatasetID string `json:"datasetId"`
	Label     string `json:"label"`
}

func progressionProjectsHandler(w http.ResponseWriter, r *http.Request) {
	if configStorage == nil {
		WriteErrorResponse(w, r, "storage_unavailable", "Progression project storage is not initialised", nil, http.StatusServiceUnavailable)
		return
	}
	switch r.Method {
	case http.MethodGet:
		progressionProjectsGet(w, r)
	case http.MethodPut:
		progressionProjectsPut(w, r)
	default:
		WriteErrorResponse(w, r, "method_not_allowed", "Method not allowed", nil, http.StatusMethodNotAllowed)
	}
}

func progressionProjectsGet(w http.ResponseWriter, _ *http.Request) {
	data, err := configStorage.RetrieveConfig(progressionProjectStorageKey)
	if err != nil {
		writeProgressionProjects(w, ProgressionProjectDocument{Version: 1, Projects: []ProgressionProject{}})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

func progressionProjectsPut(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 2*1024*1024))
	if err != nil {
		WriteErrorResponse(w, r, "read_error", "Failed to read request body", nil, http.StatusBadRequest)
		return
	}
	var document ProgressionProjectDocument
	if json.Unmarshal(body, &document) != nil || !validateProgressionProjects(&document) {
		WriteErrorResponse(w, r, "invalid_progression_projects", "Invalid progression project document", nil, http.StatusBadRequest)
		return
	}
	data, _ := json.Marshal(document)
	if err := configStorage.StoreConfig(progressionProjectStorageKey, data); err != nil {
		WriteErrorResponse(w, r, "storage_error", "Failed to save progression projects", nil, http.StatusInternalServerError)
		return
	}
	writeProgressionProjects(w, document)
}

func validateProgressionProjects(document *ProgressionProjectDocument) bool {
	if document.Version != 1 || len(document.Projects) > 100 {
		return false
	}
	seen := map[string]bool{}
	for i := range document.Projects {
		p := &document.Projects[i]
		p.ID = strings.TrimSpace(p.ID)
		p.Name = strings.TrimSpace(p.Name)
		if p.ID == "" || p.Name == "" || len(p.Name) > 100 || seen[p.ID] || len(p.Snapshots) > 100 {
			return false
		}
		seen[p.ID] = true
		if p.OrderMode != "" && p.OrderMode != "inferred" && p.OrderMode != "manual" {
			return false
		}
		for j := range p.Snapshots {
			s := &p.Snapshots[j]
			s.DatasetID = strings.TrimSpace(s.DatasetID)
			s.Label = strings.TrimSpace(s.Label)
			if s.DatasetID == "" || len(s.Label) > 120 {
				return false
			}
		}
	}
	return true
}

func writeProgressionProjects(w http.ResponseWriter, document ProgressionProjectDocument) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(document)
}
