package main

import (
	"bytes"
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
	InitCacheStorage()
	InitializeMemoryOptimizations()

	// Test with JSON backend first
	t.Run("JSON Backend", func(t *testing.T) {
		os.Setenv("USE_PROTOBUF", "false")
		InitStore()
		testBasicFlow(t, false)
	})

	// Test with Protobuf backend
	t.Run("Protobuf Backend", func(t *testing.T) {
		os.Setenv("USE_PROTOBUF", "true")
		InitStore()
		testBasicFlow(t, true)
	})
}

func testBasicFlow(t *testing.T, useProtobuf bool) {
	// Step 1: Test upload
	datasetID := testSimpleUpload(t, useProtobuf)
	if datasetID == "" {
		t.Fatal("Upload failed")
		return
	}

	// Step 2: Wait a moment for async processing
	time.Sleep(2 * time.Second)

	// Step 3: Test data retrieval
	testSimpleDataRetrieval(t, datasetID, useProtobuf)

	// Step 4: Test cache status (doesn't require config)
	testSimpleCacheStatus(t, useProtobuf)

	t.Logf("Basic flow completed successfully with %s backend", getBackendName(useProtobuf))
}

func testSimpleUpload(t *testing.T, useProtobuf bool) string {
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
		t.Fatalf("Failed to create form file: %v", err)
	}
	
	_, err = fileWriter.Write([]byte(testHTML))
	if err != nil {
		t.Fatalf("Failed to write test HTML: %v", err)
	}
	
	writer.Close()

	// Make request
	req := httptest.NewRequest("POST", "/api/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	uploadHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Upload failed with %s backend. Status: %d, Response: %s",
			getBackendName(useProtobuf), w.Code, w.Body.String())
		return ""
	}

	var response UploadResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse upload response: %v", err)
		return ""
	}

	t.Logf("Upload successful with %s backend. DatasetID: %s",
		getBackendName(useProtobuf), response.DatasetID)

	return response.DatasetID
}

func testSimpleDataRetrieval(t *testing.T, datasetID string, useProtobuf bool) {
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/players/%s", datasetID), nil)
	w := httptest.NewRecorder()

	playerDataHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Data retrieval failed with %s backend. Status: %d, Response: %s",
			getBackendName(useProtobuf), w.Code, w.Body.String())
		return
	}

	var response struct {
		Players        []Player `json:"players"`
		CurrencySymbol string   `json:"currencySymbol"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse player data response: %v", err)
		return
	}

	if len(response.Players) == 0 {
		t.Errorf("No players retrieved with %s backend", getBackendName(useProtobuf))
		return
	}

	// Basic validation
	player := response.Players[0]
	if player.Name == "" {
		t.Errorf("Player name is empty with %s backend", getBackendName(useProtobuf))
	}

	if player.UID == 0 {
		t.Errorf("Player UID is zero with %s backend", getBackendName(useProtobuf))
	}

	t.Logf("Data retrieval successful with %s backend. Players: %d",
		getBackendName(useProtobuf), len(response.Players))
}

func testSimpleCacheStatus(t *testing.T, useProtobuf bool) {
	req := httptest.NewRequest("GET", "/api/cache-status", nil)
	w := httptest.NewRecorder()

	cacheStatusHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Cache status failed with %s backend. Status: %d",
			getBackendName(useProtobuf), w.Code)
		return
	}

	var status map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &status); err != nil {
		t.Errorf("Failed to parse cache status response: %v", err)
		return
	}

	t.Logf("Cache status successful with %s backend", getBackendName(useProtobuf))
}

// TestHTMLUploadWithTestData tests HTML upload with the actual testdata.html file
func TestHTMLUploadWithTestData(t *testing.T) {
	// Initialize test environment
	InitStore()
	InitInMemoryCache()
	InitCacheStorage()
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
		os.Setenv("USE_PROTOBUF", "false")
		InitStore()
		testHTMLUploadWithData(t, testHTML, false)
	})

	// Test with Protobuf backend
	t.Run("Protobuf Backend with testdata.html", func(t *testing.T) {
		os.Setenv("USE_PROTOBUF", "true")
		InitStore()
		testHTMLUploadWithData(t, testHTML, true)
	})
}

func testHTMLUploadWithData(t *testing.T, testHTML []byte, useProtobuf bool) {
	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	
	fileWriter, err := writer.CreateFormFile("playerFile", "testdata.html")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	
	_, err = fileWriter.Write(testHTML)
	if err != nil {
		t.Fatalf("Failed to write test HTML: %v", err)
	}
	
	writer.Close()

	// Make request
	req := httptest.NewRequest("POST", "/api/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	uploadHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("HTML upload failed with %s backend. Status: %d, Response: %s",
			getBackendName(useProtobuf), w.Code, w.Body.String())
		return
	}

	var response UploadResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse upload response: %v", err)
		return
	}

	if response.DatasetID == "" {
		t.Errorf("DatasetID is empty with %s backend", getBackendName(useProtobuf))
		return
	}

	t.Logf("HTML upload successful with %s backend. DatasetID: %s, Currency: %s",
		getBackendName(useProtobuf), response.DatasetID, response.DetectedCurrencySymbol)

	// Wait for async processing
	time.Sleep(3 * time.Second)

	// Test data retrieval
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/players/%s", response.DatasetID), nil)
	w = httptest.NewRecorder()

	playerDataHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Data retrieval failed after HTML upload with %s backend. Status: %d",
			getBackendName(useProtobuf), w.Code)
		return
	}

	var dataResponse struct {
		Players        []Player `json:"players"`
		CurrencySymbol string   `json:"currencySymbol"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &dataResponse); err != nil {
		t.Errorf("Failed to parse player data response: %v", err)
		return
	}

	if len(dataResponse.Players) == 0 {
		t.Errorf("No players retrieved after HTML upload with %s backend", getBackendName(useProtobuf))
		return
	}

	t.Logf("HTML upload and retrieval successful with %s backend. Players: %d",
		getBackendName(useProtobuf), len(dataResponse.Players))
}