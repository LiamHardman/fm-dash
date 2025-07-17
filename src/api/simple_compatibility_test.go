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
	"time"
)

// TestSimpleAPICompatibility tests basic API compatibility with a shorter timeout
func TestSimpleAPICompatibility(t *testing.T) {
	// Initialize test environment
	InitStore()
	InitInMemoryCache()
	InitCacheStorage(context.Background())
	InitializeMemoryOptimizations()

	// Test with JSON backend first
	t.Run("JSON Backend", func(t *testing.T) {
		if err := os.Setenv("USE_PROTOBUF", "false"); err != nil {
			t.Fatalf("Failed to set USE_PROTOBUF: %v", err)
		}
		InitStore()
		testBasicFlow(t, false)
	})

	// Test with Protobuf backend
	t.Run("Protobuf Backend", func(t *testing.T) {
		if err := os.Setenv("USE_PROTOBUF", "true"); err != nil {
			t.Fatalf("Failed to set USE_PROTOBUF: %v", err)
		}
		InitStore()
		testBasicFlow(t, true)
	})
}

func testBasicFlow(t *testing.T, useProtobuf bool) {
	ctx := context.Background()
	logInfo(ctx, "Starting basic flow test", "use_protobuf", useProtobuf, "backend", getBackendName(useProtobuf))
	start := time.Now()

	// Step 1: Test upload
	datasetID := testSimpleUpload(ctx, t, useProtobuf)
	if datasetID == "" {
		logError(ctx, "Upload failed in basic flow test", "error", "empty dataset ID", "backend", getBackendName(useProtobuf))
		t.Fatal("Upload failed")
		return
	}

	// Step 2: Wait a moment for async processing
	time.Sleep(2 * time.Second)

	// Step 3: Test data retrieval
	testSimpleDataRetrieval(ctx, t, datasetID, useProtobuf)

	// Step 4: Test cache status (doesn't require config)
	testSimpleCacheStatus(ctx, t, useProtobuf)

	logInfo(ctx, "Basic flow completed successfully",
		"backend", getBackendName(useProtobuf),
		"dataset_id", datasetID,
		"duration_ms", time.Since(start).Milliseconds())
	t.Logf("Basic flow completed successfully with %s backend", getBackendName(useProtobuf))
}

func testSimpleUpload(ctx context.Context, t *testing.T, useProtobuf bool) string {
	logInfo(ctx, "Starting simple upload test", "backend", getBackendName(useProtobuf))
	start := time.Now()

	// Create simple test HTML
	testHTML := `<html>
<head><title>Simple Test</title></head>
<body>
<table>
<tr>
	<th>Name</th><th>Position</th><th>Club</th><th>Age</th><th>Transfer Value</th><th>Wage</th><th>Division</th><th>UID</th>
</tr>
<tr>
	<td>Simple Player</td><td>GK</td><td>Simple FC</td><td>25</td><td>£1M</td><td>£10K</td><td>Test League</td><td>999999</td>
</tr>
</table>
</body>
</html>`

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	fileWriter, err := writer.CreateFormFile("playerFile", "simple_test.html")
	if err != nil {
		logError(ctx, "Failed to create form file", "error", err, "backend", getBackendName(useProtobuf))
		t.Fatalf("Failed to create form file: %v", err)
	}

	_, err = fileWriter.Write([]byte(testHTML))
	if err != nil {
		logError(ctx, "Failed to write test HTML", "error", err, "backend", getBackendName(useProtobuf))
		t.Fatalf("Failed to write test HTML: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close multipart writer: %v", err)
	}

	// Make request
	req := httptest.NewRequest("POST", "/api/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	uploadHandler(w, req)

	if w.Code != http.StatusOK {
		logError(ctx, "Upload request failed", "error", fmt.Sprintf("status %d", w.Code),
			"backend", getBackendName(useProtobuf), "response", w.Body.String())
		t.Errorf("Upload failed with %s backend. Status: %d, Response: %s",
			getBackendName(useProtobuf), w.Code, w.Body.String())
		return ""
	}

	var response UploadResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		logError(ctx, "Failed to parse upload response", "error", err, "backend", getBackendName(useProtobuf))
		t.Errorf("Failed to parse upload response: %v", err)
		return ""
	}

	logInfo(ctx, "Simple upload test completed successfully",
		"backend", getBackendName(useProtobuf),
		"dataset_id", response.DatasetID,
		"duration_ms", time.Since(start).Milliseconds())
	t.Logf("Upload successful with %s backend. DatasetID: %s",
		getBackendName(useProtobuf), response.DatasetID)

	return response.DatasetID
}

func testSimpleDataRetrieval(ctx context.Context, t *testing.T, datasetID string, useProtobuf bool) {
	logInfo(ctx, "Starting data retrieval test", "dataset_id", datasetID, "backend", getBackendName(useProtobuf))
	start := time.Now()

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/players/%s", datasetID), nil)
	w := httptest.NewRecorder()

	playerDataHandler(w, req)

	if w.Code != http.StatusOK {
		logError(ctx, "Data retrieval request failed", "error", fmt.Sprintf("status %d", w.Code),
			"dataset_id", datasetID, "backend", getBackendName(useProtobuf), "response", w.Body.String())
		t.Errorf("Data retrieval failed with %s backend. Status: %d, Response: %s",
			getBackendName(useProtobuf), w.Code, w.Body.String())
		return
	}

	var response struct {
		Players        []Player `json:"players"`
		CurrencySymbol string   `json:"currencySymbol"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		logError(ctx, "Failed to parse player data response", "error", err, "dataset_id", datasetID, "backend", getBackendName(useProtobuf))
		t.Errorf("Failed to parse player data response: %v", err)
		return
	}

	if len(response.Players) == 0 {
		logError(ctx, "No players retrieved", "dataset_id", datasetID, "backend", getBackendName(useProtobuf))
		t.Errorf("No players retrieved with %s backend", getBackendName(useProtobuf))
		return
	}

	// Basic validation
	player := response.Players[0]
	if player.Name == "" {
		logError(ctx, "Player name validation failed", "dataset_id", datasetID, "backend", getBackendName(useProtobuf))
		t.Errorf("Player name is empty with %s backend", getBackendName(useProtobuf))
	}

	if player.UID == 0 {
		logError(ctx, "Player UID validation failed", "dataset_id", datasetID, "backend", getBackendName(useProtobuf))
		t.Errorf("Player UID is zero with %s backend", getBackendName(useProtobuf))
	}

	logInfo(ctx, "Data retrieval test completed successfully",
		"dataset_id", datasetID,
		"backend", getBackendName(useProtobuf),
		"player_count", len(response.Players),
		"duration_ms", time.Since(start).Milliseconds())
	t.Logf("Data retrieval successful with %s backend. Players: %d",
		getBackendName(useProtobuf), len(response.Players))
}

func testSimpleCacheStatus(ctx context.Context, t *testing.T, useProtobuf bool) {
	logInfo(ctx, "Starting cache status test", "backend", getBackendName(useProtobuf))
	start := time.Now()

	req := httptest.NewRequest("GET", "/api/cache-status", nil)
	w := httptest.NewRecorder()

	cacheStatusHandler(w, req)

	if w.Code != http.StatusOK {
		logError(ctx, "Cache status request failed", "error", fmt.Sprintf("status %d", w.Code), "backend", getBackendName(useProtobuf))
		t.Errorf("Cache status failed with %s backend. Status: %d",
			getBackendName(useProtobuf), w.Code)
		return
	}

	var status map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &status); err != nil {
		logError(ctx, "Failed to parse cache status response", "error", err, "backend", getBackendName(useProtobuf))
		t.Errorf("Failed to parse cache status response: %v", err)
		return
	}

	logInfo(ctx, "Cache status test completed successfully", 
		"backend", getBackendName(useProtobuf),
		"duration_ms", time.Since(start).Milliseconds())
	t.Logf("Cache status successful with %s backend", getBackendName(useProtobuf))
}

// TestHTMLUploadWithTestData tests HTML upload with the actual testdata.html file
func TestHTMLUploadWithTestData(t *testing.T) {
	// Initialize test environment
	InitStore()
	InitInMemoryCache()
	InitCacheStorage(context.Background())
	InitializeMemoryOptimizations()

	// Check if testdata.html exists
	testDataPath := "../../testdata.html"
	if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
		testDataPath = "testdata.html"
		if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
			t.Skip("testdata.html not found, skipping test")
			return
		}
	}

	// Read testdata.html
	testHTML, err := os.ReadFile(testDataPath)
	if err != nil {
		t.Fatalf("Failed to read testdata.html: %v", err)
	}

	// Test with JSON backend
	t.Run("JSON Backend with testdata.html", func(t *testing.T) {
		if err := os.Setenv("USE_PROTOBUF", "false"); err != nil {
			t.Fatalf("Failed to set USE_PROTOBUF: %v", err)
		}
		InitStore()
		testHTMLUploadWithData(context.Background(), t, testHTML, false)
	})

	// Test with Protobuf backend
	t.Run("Protobuf Backend with testdata.html", func(t *testing.T) {
		if err := os.Setenv("USE_PROTOBUF", "true"); err != nil {
			t.Fatalf("Failed to set USE_PROTOBUF: %v", err)
		}
		InitStore()
		testHTMLUploadWithData(context.Background(), t, testHTML, true)
	})
}

func testHTMLUploadWithData(ctx context.Context, t *testing.T, testHTML []byte, useProtobuf bool) {
	logInfo(ctx, "Starting HTML upload test with real data", "backend", getBackendName(useProtobuf), "data_size_bytes", len(testHTML))
	start := time.Now()

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	fileWriter, err := writer.CreateFormFile("playerFile", "testdata.html")
	if err != nil {
		logError(ctx, "Failed to create form file for HTML upload", "error", err, "backend", getBackendName(useProtobuf))
		t.Fatalf("Failed to create form file: %v", err)
	}

	_, err = fileWriter.Write(testHTML)
	if err != nil {
		logError(ctx, "Failed to write HTML data", "error", err, "backend", getBackendName(useProtobuf))
		t.Fatalf("Failed to write test HTML: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close multipart writer: %v", err)
	}

	// Make request
	req := httptest.NewRequest("POST", "/api/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	uploadHandler(w, req)

	if w.Code != http.StatusOK {
		logError(ctx, "HTML upload request failed", "error", fmt.Sprintf("status %d", w.Code), 
			"backend", getBackendName(useProtobuf), "response", w.Body.String())
		t.Errorf("HTML upload failed with %s backend. Status: %d, Response: %s",
			getBackendName(useProtobuf), w.Code, w.Body.String())
		return
	}

	var response UploadResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		logError(ctx, "Failed to parse HTML upload response", "error", err, "backend", getBackendName(useProtobuf))
		t.Errorf("Failed to parse upload response: %v", err)
		return
	}

	if response.DatasetID == "" {
		logError(ctx, "HTML upload returned empty dataset ID", "backend", getBackendName(useProtobuf))
		t.Errorf("DatasetID is empty with %s backend", getBackendName(useProtobuf))
		return
	}

	logInfo(ctx, "HTML upload completed successfully", "backend", getBackendName(useProtobuf), 
		"dataset_id", response.DatasetID, "currency", response.DetectedCurrencySymbol)
	t.Logf("HTML upload successful with %s backend. DatasetID: %s, Currency: %s",
		getBackendName(useProtobuf), response.DatasetID, response.DetectedCurrencySymbol)

	// Wait for async processing
	time.Sleep(3 * time.Second)

	// Test data retrieval
	logInfo(ctx, "Testing data retrieval after HTML upload", "dataset_id", response.DatasetID, "backend", getBackendName(useProtobuf))
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/players/%s", response.DatasetID), nil)
	w = httptest.NewRecorder()

	playerDataHandler(w, req)

	if w.Code != http.StatusOK {
		logError(ctx, "Data retrieval failed after HTML upload", "error", fmt.Sprintf("status %d", w.Code),
			"dataset_id", response.DatasetID, "backend", getBackendName(useProtobuf))
		t.Errorf("Data retrieval failed after HTML upload with %s backend. Status: %d",
			getBackendName(useProtobuf), w.Code)
		return
	}

	var dataResponse struct {
		Players        []Player `json:"players"`
		CurrencySymbol string   `json:"currencySymbol"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &dataResponse); err != nil {
		logError(ctx, "Failed to parse player data after HTML upload", "error", err, 
			"dataset_id", response.DatasetID, "backend", getBackendName(useProtobuf))
		t.Errorf("Failed to parse player data response: %v", err)
		return
	}

	if len(dataResponse.Players) == 0 {
		logError(ctx, "No players retrieved after HTML upload", "dataset_id", response.DatasetID, "backend", getBackendName(useProtobuf))
		t.Errorf("No players retrieved after HTML upload with %s backend", getBackendName(useProtobuf))
		return
	}

	logInfo(ctx, "HTML upload and retrieval test completed successfully", 
		"backend", getBackendName(useProtobuf), 
		"dataset_id", response.DatasetID,
		"player_count", len(dataResponse.Players),
		"duration_ms", time.Since(start).Milliseconds())
	t.Logf("HTML upload and retrieval successful with %s backend. Players: %d",
		getBackendName(useProtobuf), len(dataResponse.Players))
}
