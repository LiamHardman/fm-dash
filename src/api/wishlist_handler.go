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

const wishlistKeyPrefix = "wishlists/"

func wishlistStorageKey(datasetID string) string {
	return fmt.Sprintf("%s%s.json", wishlistKeyPrefix, datasetID)
}

func wishlistHandler(w http.ResponseWriter, r *http.Request) {
	datasetID := strings.TrimPrefix(r.URL.Path, "/api/wishlists/")
	datasetID = strings.TrimSuffix(datasetID, "/")

	if _, err := uuid.Parse(datasetID); err != nil {
		WriteErrorResponse(w, r, "invalid_dataset_id", "Invalid dataset ID format", nil, http.StatusBadRequest)
		return
	}

	if configStorage == nil {
		WriteErrorResponse(w, r, "storage_unavailable", "Wishlist storage is not initialised", nil, http.StatusServiceUnavailable)
		return
	}

	switch r.Method {
	case http.MethodGet:
		wishlistGetHandler(w, r, datasetID)
	case http.MethodPut:
		wishlistPutHandler(w, r, datasetID)
	case http.MethodDelete:
		wishlistDeleteHandler(w, r, datasetID)
	default:
		WriteErrorResponse(w, r, "method_not_allowed", "Method not allowed", nil, http.StatusMethodNotAllowed)
	}
}

func wishlistGetHandler(w http.ResponseWriter, r *http.Request, datasetID string) {
	key := wishlistStorageKey(datasetID)
	data, err := configStorage.RetrieveConfig(key)
	if err != nil {
		if os.IsNotExist(err) {
			WriteErrorResponse(w, r, "wishlist_not_found", "Wishlist not found", nil, http.StatusNotFound)
			return
		}
		// For S3, treat any retrieval error as not found so the client falls back to localStorage
		WriteErrorResponse(w, r, "wishlist_not_found", "Wishlist not found", nil, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		LogWarn("wishlist: failed to write GET response: %v", err)
	}
}

func wishlistPutHandler(w http.ResponseWriter, r *http.Request, datasetID string) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 10*1024*1024)) // 10 MB limit
	if err != nil {
		WriteErrorResponse(w, r, "read_error", "Failed to read request body", nil, http.StatusBadRequest)
		return
	}

	if !json.Valid(body) {
		WriteErrorResponse(w, r, "invalid_json", "Request body must be valid JSON", nil, http.StatusBadRequest)
		return
	}

	if err := configStorage.StoreConfig(wishlistStorageKey(datasetID), body); err != nil {
		LogWarn("wishlist: failed to store wishlist for dataset %s: %v", sanitizeForLogging(datasetID), err)
		WriteErrorResponse(w, r, "storage_error", "Failed to save wishlist", nil, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(body); err != nil {
		LogWarn("wishlist: failed to write PUT response: %v", err)
	}
}

func wishlistDeleteHandler(w http.ResponseWriter, r *http.Request, datasetID string) {
	if err := configStorage.DeleteConfig(wishlistStorageKey(datasetID)); err != nil {
		LogWarn("wishlist: failed to delete wishlist for dataset %s: %v", sanitizeForLogging(datasetID), err)
		WriteErrorResponse(w, r, "storage_error", "Failed to delete wishlist", nil, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
