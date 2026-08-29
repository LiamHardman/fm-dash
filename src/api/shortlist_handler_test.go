package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShortlistsGetReturnsDefaultDocument(t *testing.T) {
	tempDir := t.TempDir()
	previousStorage := configStorage
	configStorage = &localConfigStorage{dir: tempDir}
	t.Cleanup(func() { configStorage = previousStorage })

	req := httptest.NewRequest(http.MethodGet, "/api/shortlists", nil)
	response := httptest.NewRecorder()
	shortlistsHandler(response, req)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "My shortlist") {
		t.Fatalf("expected default shortlist, got %s", response.Body.String())
	}
}

func TestShortlistsPutRejectsPlayerSnapshots(t *testing.T) {
	tempDir := t.TempDir()
	previousStorage := configStorage
	configStorage = &localConfigStorage{dir: tempDir}
	t.Cleanup(func() { configStorage = previousStorage })

	body := `{"version":1,"lists":[{"id":"default","name":"My shortlist","items":[{"playerRef":{"datasetId":"ce27fb2f-5d0f-4a09-b11c-591dcbf25d91","playerUid":7},"status":"watching","priority":"medium","tags":[],"notes":"","targetFee":0,"targetWage":0,"player":{"name":"Must not persist"}}]}]}`
	req := httptest.NewRequest(http.MethodPut, "/api/shortlists", strings.NewReader(body))
	response := httptest.NewRecorder()
	shortlistsHandler(response, req)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}
	if strings.Contains(response.Body.String(), "Must not persist") {
		t.Fatalf("player snapshot leaked into stored response: %s", response.Body.String())
	}
}

func TestShortlistsPutRejectsInvalidStatus(t *testing.T) {
	tempDir := t.TempDir()
	previousStorage := configStorage
	configStorage = &localConfigStorage{dir: tempDir}
	t.Cleanup(func() { configStorage = previousStorage })

	body := `{"version":1,"lists":[{"id":"default","name":"My shortlist","items":[{"playerRef":{"datasetId":"ce27fb2f-5d0f-4a09-b11c-591dcbf25d91","playerUid":7},"status":"maybe","priority":"medium","tags":[],"notes":"","targetFee":0,"targetWage":0}]}]}`
	req := httptest.NewRequest(http.MethodPut, "/api/shortlists", strings.NewReader(body))
	response := httptest.NewRecorder()
	shortlistsHandler(response, req)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}
