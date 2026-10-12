package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

var errProfileConflict = errors.New("profile document changed")

func configNotFound(err error) bool { return errors.Is(err, os.ErrNotExist) }

func profileRevision(data []byte) string {
	if data == nil {
		return `"missing"`
	}
	return fmt.Sprintf(`"%x"`, sha256.Sum256(data))
}

func readProfileDocument(w http.ResponseWriter, r *http.Request, key string, fallback any) {
	data, err := configStorage.RetrieveConfig(key)
	if err != nil && !configNotFound(err) {
		WriteErrorResponse(w, r, "storage_error", "Unable to load your saved data. Please retry.", nil, http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("ETag", profileRevision(data))
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	if configNotFound(err) {
		data, _ = json.Marshal(fallback)
	}
	_, _ = w.Write(data)
}

type conditionalConfigStorage interface {
	CompareAndSwapConfig(name string, data []byte, expected string) error
}

func writeProfileDocument(w http.ResponseWriter, r *http.Request, key string, data []byte) bool {
	expected := r.Header.Get("If-Match")
	if expected == "" {
		WriteErrorResponse(w, r, "revision_required", "Load the document before saving changes.", nil, http.StatusPreconditionRequired)
		return false
	}
	store, ok := configStorage.(conditionalConfigStorage)
	if !ok {
		WriteErrorResponse(w, r, "storage_error", "Storage does not support safe document updates.", nil, http.StatusServiceUnavailable)
		return false
	}
	if err := store.CompareAndSwapConfig(key, data, expected); err != nil {
		if errors.Is(err, errProfileConflict) {
			WriteErrorResponse(w, r, "revision_conflict", "This document changed elsewhere. Review both versions before saving.", nil, http.StatusPreconditionFailed)
		} else {
			WriteErrorResponse(w, r, "storage_error", "Unable to save your changes. Please retry.", nil, http.StatusServiceUnavailable)
		}
		return false
	}
	w.Header().Set("ETag", profileRevision(data))
	w.Header().Set("Cache-Control", "no-store")
	return true
}
