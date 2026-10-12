package main

import (
	"encoding/json"
	"io"
	"net/http"
)

type ProfileBackup struct {
	Version             int                        `json:"version"`
	Shortlists          ShortlistDocument          `json:"shortlists"`
	SavedSearches       SavedSearchDocument        `json:"savedSearches"`
	ProgressionProjects ProgressionProjectDocument `json:"progressionProjects"`
}

// Validation is read-only. Imports use the existing revision-checked endpoints.
func profileBackupValidateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 6*1024*1024))
	decoder.DisallowUnknownFields()
	var backup ProfileBackup
	err := decoder.Decode(&backup)
	if err != nil || decoder.Decode(new(any)) != io.EOF || backup.Version != 1 || validateShortlistDocument(&backup.Shortlists) != nil || validateSavedSearchDocument(&backup.SavedSearches) != nil || !validateProgressionProjects(&backup.ProgressionProjects) {
		WriteErrorResponse(w, r, "invalid_backup", "Invalid profile backup. Only supported manager-authored data and Player references can be imported.", nil, http.StatusBadRequest)
		return
	}
	// Match the individual save limits, including after a merge.
	for _, document := range []any{backup.Shortlists, backup.SavedSearches, backup.ProgressionProjects} {
		data, _ := json.Marshal(document)
		if len(data) > 2*1024*1024 {
			WriteErrorResponse(w, r, "backup_too_large", "One profile document exceeds its size limit.", nil, http.StatusBadRequest)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(backup)
}
