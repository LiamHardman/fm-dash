package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// TestHTMLUploadCompatibility tests HTML upload processing with testdata.html
// using both JSON and protobuf storage backends
func TestHTMLUploadCompatibility(t *testing.T) {
	// Initialize test environment
	InitStore()
	InitInMemoryCache()
	InitCacheStorage(context.Background())
	InitializeMemoryOptimizations()

	// Check if testdata.html exists
	testDataPath := "../../testdata.html"
	if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
		// Try alternative path
		testDataPath = "testdata.html"
		if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
			t.Skip("testdata.html not found, skipping HTML upload test")
			return
		}
	}

	// Read testdata.html
	testHTML, err := os.ReadFile(testDataPath)
	if err != nil {
		t.Fatalf("Failed to read testdata.html: %v", err)
	}

	// Test with both storage backends
	testCases := []struct {
		name        string
		useProtobuf bool
		envVarValue string
	}{
		{
			name:        "JSON Storage Backend",
			useProtobuf: false,
			envVarValue: "false",
		},
		{
			name:        "Protobuf Storage Backend",
			useProtobuf: true,
			envVarValue: "true",
		},
	}

	var jsonResponse, protobufResponse UploadResponse
	var jsonDatasetID, protobufDatasetID string

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variable for storage backend
			originalValue := os.Getenv("USE_PROTOBUF")
			os.Setenv("USE_PROTOBUF", tc.envVarValue)
			defer os.Setenv("USE_PROTOBUF", originalValue)

			// Re-initialize storage with new setting
			InitStore()

			// Test HTML upload
			response, datasetID := testHTMLUpload(t, testHTML, tc.useProtobuf)

			// Store responses for comparison
			if tc.useProtobuf {
				protobufResponse = response
				protobufDatasetID = datasetID
			} else {
				jsonResponse = response
				jsonDatasetID = datasetID
			}

			// Test data retrieval
			testDataRetrieval(t, datasetID, tc.useProtobuf)

			// Test data integrity
			testDataIntegrity(t, datasetID, tc.useProtobuf)
		})
	}

	// Compare responses between backends
	t.Run("Compare Backend Responses", func(t *testing.T) {
		if jsonDatasetID == "" || protobufDatasetID == "" {
			t.Skip("Missing dataset IDs for comparison")
			return
		}

		// Compare upload responses (excluding dataset ID which will be different)
		if jsonResponse.Message != protobufResponse.Message {
			t.Errorf("Upload messages differ: JSON='%s', Protobuf='%s'",
				jsonResponse.Message, protobufResponse.Message)
		}

		if jsonResponse.DetectedCurrencySymbol != protobufResponse.DetectedCurrencySymbol {
			t.Errorf("Currency symbols differ: JSON='%s', Protobuf='%s'",
				jsonResponse.DetectedCurrencySymbol, protobufResponse.DetectedCurrencySymbol)
		}

		// Compare player data responses
		comparePlayerDataResponses(t, jsonDatasetID, protobufDatasetID)
	})
}

// testHTMLUpload tests uploading the HTML file with a specific storage backend
func testHTMLUpload(t *testing.T, testHTML []byte, useProtobuf bool) (UploadResponse, string) {
	// Create multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add file field
	fileWriter, err := writer.CreateFormFile("playerFile", "testdata.html")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	_, err = fileWriter.Write(testHTML)
	if err != nil {
		t.Fatalf("Failed to write test HTML: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close multipart writer: %v", err)
	}

	// Create HTTP request
	req := httptest.NewRequest("POST", "/api/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	uploadHandler(w, req)

	// Check response status
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d with %s backend. Response: %s",
			http.StatusOK, w.Code, getBackendName(useProtobuf), w.Body.String())
	}

	// Parse response
	var response UploadResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse upload response with %s backend: %v",
			getBackendName(useProtobuf), err)
	}

	// Validate response
	if response.DatasetID == "" {
		t.Errorf("DatasetID should not be empty with %s backend", getBackendName(useProtobuf))
	}

	if response.Message == "" {
		t.Errorf("Message should not be empty with %s backend", getBackendName(useProtobuf))
	}

	if response.DetectedCurrencySymbol == "" {
		t.Errorf("DetectedCurrencySymbol should not be empty with %s backend", getBackendName(useProtobuf))
	}

	t.Logf("HTML upload successful with %s backend. DatasetID: %s, Currency: %s",
		getBackendName(useProtobuf), response.DatasetID, response.DetectedCurrencySymbol)

	return response, response.DatasetID
}

// testDataRetrieval tests retrieving the uploaded data
func testDataRetrieval(t *testing.T, datasetID string, useProtobuf bool) {
	// Test basic player data retrieval
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/players/%s", datasetID), nil)
	w := httptest.NewRecorder()

	playerDataHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d with %s backend. Response: %s",
			http.StatusOK, w.Code, getBackendName(useProtobuf), w.Body.String())
		return
	}

	// Parse response
	var response struct {
		Players        []Player `json:"players"`
		CurrencySymbol string   `json:"currencySymbol"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse player data response with %s backend: %v",
			getBackendName(useProtobuf), err)
	}

	// Validate response
	if len(response.Players) == 0 {
		t.Errorf("Players array should not be empty with %s backend", getBackendName(useProtobuf))
		return
	}

	if response.CurrencySymbol == "" {
		t.Errorf("CurrencySymbol should not be empty with %s backend", getBackendName(useProtobuf))
	}

	t.Logf("Data retrieval successful with %s backend. Players: %d, Currency: %s",
		getBackendName(useProtobuf), len(response.Players), response.CurrencySymbol)

	// Test with various filters to ensure compatibility
	testFilters := []string{
		"?position=GK",
		"?role=Goalkeeper",
		"?minAge=18&maxAge=30",
		"?divisionFilter=all",
		"?position=D&minAge=20",
	}

	for _, filter := range testFilters {
		url := fmt.Sprintf("/api/players/%s%s", datasetID, filter)
		req := httptest.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()

		playerDataHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Filter test failed with %s backend for filter %s. Status: %d, Response: %s",
				getBackendName(useProtobuf), filter, w.Code, w.Body.String())
		}
	}
}

// testDataIntegrity tests that the data maintains integrity across storage backends
func testDataIntegrity(t *testing.T, datasetID string, useProtobuf bool) {
	// Get player data
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/players/%s", datasetID), nil)
	w := httptest.NewRecorder()

	playerDataHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Failed to get player data for integrity test with %s backend", getBackendName(useProtobuf))
		return
	}

	var response struct {
		Players        []Player `json:"players"`
		CurrencySymbol string   `json:"currencySymbol"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse player data for integrity test with %s backend: %v",
			getBackendName(useProtobuf), err)
	}

	// Validate data integrity
	for i, player := range response.Players {
		// Check required fields
		if player.UID == 0 {
			t.Errorf("Player %d has zero UID with %s backend", i, getBackendName(useProtobuf))
		}

		if player.Name == "" {
			t.Errorf("Player %d has empty name with %s backend", i, getBackendName(useProtobuf))
		}

		if player.Position == "" {
			t.Errorf("Player %d has empty position with %s backend", i, getBackendName(useProtobuf))
		}

		if player.Club == "" {
			t.Errorf("Player %d has empty club with %s backend", i, getBackendName(useProtobuf))
		}

		// Check that maps are initialized
		if player.Attributes == nil {
			t.Errorf("Player %d has nil Attributes map with %s backend", i, getBackendName(useProtobuf))
		}

		if player.NumericAttributes == nil {
			t.Errorf("Player %d has nil NumericAttributes map with %s backend", i, getBackendName(useProtobuf))
		}

		// Check numeric values are reasonable
		if player.Overall < 0 || player.Overall > 100 {
			t.Errorf("Player %d has invalid overall rating %d with %s backend",
				i, player.Overall, getBackendName(useProtobuf))
		}

		// Check that position parsing worked
		if len(player.ParsedPositions) == 0 {
			t.Errorf("Player %d has no parsed positions with %s backend", i, getBackendName(useProtobuf))
		}
	}

	t.Logf("Data integrity validated for %d players with %s backend",
		len(response.Players), getBackendName(useProtobuf))
}

// comparePlayerDataResponses compares player data responses between JSON and protobuf backends
func comparePlayerDataResponses(t *testing.T, jsonDatasetID, protobufDatasetID string) {
	// Get JSON response
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/players/%s", jsonDatasetID), nil)
	w := httptest.NewRecorder()

	// Set to JSON backend
	if err := os.Setenv("USE_PROTOBUF", "false"); err != nil {
		t.Fatalf("Failed to set USE_PROTOBUF to false: %v", err)
	}
	InitStore()
	playerDataHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Failed to get JSON player data for comparison")
		return
	}

	var jsonResponse struct {
		Players        []Player `json:"players"`
		CurrencySymbol string   `json:"currencySymbol"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &jsonResponse); err != nil {
		t.Fatalf("Failed to parse JSON player data for comparison: %v", err)
	}

	// Get Protobuf response
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/players/%s", protobufDatasetID), nil)
	w = httptest.NewRecorder()

	// Set to Protobuf backend
	if err := os.Setenv("USE_PROTOBUF", "true"); err != nil {
		t.Fatalf("Failed to set USE_PROTOBUF to true: %v", err)
	}
	InitStore()
	playerDataHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Failed to get Protobuf player data for comparison")
		return
	}

	var protobufResponse struct {
		Players        []Player `json:"players"`
		CurrencySymbol string   `json:"currencySymbol"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &protobufResponse); err != nil {
		t.Fatalf("Failed to parse Protobuf player data for comparison: %v", err)
	}

	// Compare responses
	if len(jsonResponse.Players) != len(protobufResponse.Players) {
		t.Errorf("Player count differs: JSON=%d, Protobuf=%d",
			len(jsonResponse.Players), len(protobufResponse.Players))
		return
	}

	if jsonResponse.CurrencySymbol != protobufResponse.CurrencySymbol {
		t.Errorf("Currency symbols differ: JSON='%s', Protobuf='%s'",
			jsonResponse.CurrencySymbol, protobufResponse.CurrencySymbol)
	}

	// Compare first few players in detail
	compareCount := len(jsonResponse.Players)
	if compareCount > 5 {
		compareCount = 5 // Limit detailed comparison to first 5 players
	}

	for i := 0; i < compareCount; i++ {
		jsonPlayer := jsonResponse.Players[i]
		protobufPlayer := protobufResponse.Players[i]

		// Compare key fields
		if jsonPlayer.UID != protobufPlayer.UID {
			t.Errorf("Player %d UID differs: JSON=%d, Protobuf=%d",
				i, jsonPlayer.UID, protobufPlayer.UID)
		}

		if jsonPlayer.Name != protobufPlayer.Name {
			t.Errorf("Player %d name differs: JSON='%s', Protobuf='%s'",
				i, jsonPlayer.Name, protobufPlayer.Name)
		}

		if jsonPlayer.Position != protobufPlayer.Position {
			t.Errorf("Player %d position differs: JSON='%s', Protobuf='%s'",
				i, jsonPlayer.Position, protobufPlayer.Position)
		}

		if jsonPlayer.Club != protobufPlayer.Club {
			t.Errorf("Player %d club differs: JSON='%s', Protobuf='%s'",
				i, jsonPlayer.Club, protobufPlayer.Club)
		}

		if jsonPlayer.Overall != protobufPlayer.Overall {
			t.Errorf("Player %d overall differs: JSON=%d, Protobuf=%d",
				i, jsonPlayer.Overall, protobufPlayer.Overall)
		}

		// Compare attribute counts
		if len(jsonPlayer.Attributes) != len(protobufPlayer.Attributes) {
			t.Errorf("Player %d attribute count differs: JSON=%d, Protobuf=%d",
				i, len(jsonPlayer.Attributes), len(protobufPlayer.Attributes))
		}

		if len(jsonPlayer.NumericAttributes) != len(protobufPlayer.NumericAttributes) {
			t.Errorf("Player %d numeric attribute count differs: JSON=%d, Protobuf=%d",
				i, len(jsonPlayer.NumericAttributes), len(protobufPlayer.NumericAttributes))
		}
	}

	t.Logf("Successfully compared %d players between JSON and Protobuf backends", compareCount)
}

// TestErrorResponseCompatibility tests that error responses are identical between backends
func TestErrorResponseCompatibility(t *testing.T) {
	// Initialize test environment
	InitStore()
	InitInMemoryCache()
	InitCacheStorage(context.Background())
	InitializeMemoryOptimizations()

	testCases := []struct {
		name        string
		useProtobuf bool
		envVarValue string
	}{
		{
			name:        "JSON Storage Backend",
			useProtobuf: false,
			envVarValue: "false",
		},
		{
			name:        "Protobuf Storage Backend",
			useProtobuf: true,
			envVarValue: "true",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variable for storage backend
			originalValue := os.Getenv("USE_PROTOBUF")
			os.Setenv("USE_PROTOBUF", tc.envVarValue)
			defer os.Setenv("USE_PROTOBUF", originalValue)

			// Re-initialize storage with new setting
			InitStore()

			// Test 404 error for non-existent dataset
			req := httptest.NewRequest("GET", "/api/players/non-existent-id", nil)
			w := httptest.NewRecorder()

			playerDataHandler(w, req)

			if w.Code != http.StatusNotFound {
				t.Errorf("Expected status %d for non-existent dataset with %s backend, got %d",
					http.StatusNotFound, getBackendName(tc.useProtobuf), w.Code)
			}

			// Test method not allowed
			req = httptest.NewRequest("DELETE", "/api/players/test-id", nil)
			w = httptest.NewRecorder()

			playerDataHandler(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status %d for invalid method with %s backend, got %d",
					http.StatusMethodNotAllowed, getBackendName(tc.useProtobuf), w.Code)
			}

			// Test invalid upload (empty request)
			req = httptest.NewRequest("POST", "/api/upload", nil)
			w = httptest.NewRecorder()

			uploadHandler(w, req)

			if w.Code == http.StatusOK {
				t.Errorf("Expected error status for empty upload with %s backend, got %d",
					getBackendName(tc.useProtobuf), w.Code)
			}
		})
	}
}
