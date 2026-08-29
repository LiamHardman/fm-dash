package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSavedSearchesGetReturnsEmptyDocument(t *testing.T) {
	previousStorage := configStorage
	configStorage = &localConfigStorage{dir: t.TempDir()}
	t.Cleanup(func() { configStorage = previousStorage })
	response := httptest.NewRecorder()
	savedSearchesHandler(response, httptest.NewRequest(http.MethodGet, "/api/saved-searches", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"version":1`) {
		t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
	}
}

func TestSavedSearchesRejectDatasetData(t *testing.T) {
	previousStorage := configStorage
	configStorage = &localConfigStorage{dir: t.TempDir()}
	t.Cleanup(func() { configStorage = previousStorage })
	body := `{"version":1,"searches":[{"id":"one","name":"One","filters":{"position":[],"datasetId":"private"}}]}`
	response := httptest.NewRecorder()
	savedSearchesHandler(response, httptest.NewRequest(http.MethodPut, "/api/saved-searches", strings.NewReader(body)))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}
