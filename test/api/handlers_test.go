package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// TestUploadHandler tests the file upload functionality
func TestUploadHandler(t *testing.T) {
	// Initialize storage for testing
	InitStore()

	tests := []struct {
		name           string
		method         string
		setupFile      func() (*bytes.Buffer, string, error)
		expectedStatus int
		expectedInBody string
	}{
		{
			name:           "Invalid Method",
			method:         "GET",
			setupFile:      func() (*bytes.Buffer, string, error) { return nil, "", nil },
			expectedStatus: http.StatusMethodNotAllowed,
			expectedInBody: "Only POST method is allowed",
		},
		{
			name:   "Valid HTML Upload",
			method: "POST",
			setupFile: func() (*bytes.Buffer, string, error) {
				return createMockHTMLFile("test-file.html", createMockPlayerHTML())
			},
			expectedStatus: http.StatusOK,
			expectedInBody: "File uploaded and parsed successfully",
		},
		{
			name:   "File Too Large",
			method: "POST",
			setupFile: func() (*bytes.Buffer, string, error) {
				// Create a file larger than the limit
				largeContent := strings.Repeat("a", int(getMaxUploadSize())+1)
				return createMockHTMLFile("large-file.html", largeContent)
			},
			expectedStatus: http.StatusRequestEntityTooLarge,
			expectedInBody: "players or less can be in a given dataset",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request

			if tt.setupFile != nil {
				body, contentType, setupErr := tt.setupFile()
				if setupErr != nil && tt.expectedStatus == http.StatusOK {
					t.Fatalf("Failed to setup test file: %v", setupErr)
				}

				if body != nil {
					req = httptest.NewRequest(tt.method, "/upload", body)
					req.Header.Set("Content-Type", contentType)
				} else {
					req = httptest.NewRequest(tt.method, "/upload", nil)
				}
			} else {
				req = httptest.NewRequest(tt.method, "/upload", nil)
			}

			w := httptest.NewRecorder()
			uploadHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedInBody != "" && !strings.Contains(w.Body.String(), tt.expectedInBody) {
				t.Errorf("Expected body to contain %q, got %q", tt.expectedInBody, w.Body.String())
			}
		})
	}
}

// TestPlayerDataHandler tests the player data retrieval functionality
func TestPlayerDataHandler(t *testing.T) {
	InitStore()

	// Setup test data
	testPlayers := []Player{
		{
			UID:             "test-1",
			Name:            "Test Player 1",
			Position:        "Striker",
			Age:             "25",
			Club:            "Test FC",
			Division:        "Premier League",
			Overall:         85,
			ParsedPositions: []string{"Striker"},
			ShortPositions:  []string{"ST"},
			PositionGroups:  []string{"Attackers"},
		},
		{
			UID:             "test-2",
			Name:            "Test Player 2",
			Position:        "Defender",
			Age:             "28",
			Club:            "Test United",
			Division:        "Premier League",
			Overall:         82,
			ParsedPositions: []string{"Centre Back"},
			ShortPositions:  []string{"DC"},
			PositionGroups:  []string{"Defenders"},
		},
	}

	testDatasetID := "test-dataset-123"
	SetPlayerData(testDatasetID, testPlayers, "$")

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		validateBody   func(t *testing.T, body string)
	}{
		{
			name:           "Valid Dataset ID",
			path:           "/api/players/" + testDatasetID,
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				var response PlayerDataWithCurrency
				if err := json.Unmarshal([]byte(body), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
					return
				}
				if len(response.Players) != 2 {
					t.Errorf("Expected 2 players, got %d", len(response.Players))
				}
				if response.CurrencySymbol != "$" {
					t.Errorf("Expected currency symbol '$', got %q", response.CurrencySymbol)
				}
			},
		},
		{
			name:           "Invalid Dataset ID",
			path:           "/api/players/nonexistent",
			expectedStatus: http.StatusNotFound,
			validateBody: func(t *testing.T, body string) {
				if !strings.Contains(body, "Player data not found") {
					t.Errorf("Expected 'Player data not found' in response, got %q", body)
				}
			},
		},
		{
			name:           "Missing Dataset ID",
			path:           "/api/players/",
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				if !strings.Contains(body, "Dataset ID is missing") {
					t.Errorf("Expected 'Dataset ID is missing' in response, got %q", body)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			playerDataHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.validateBody != nil {
				tt.validateBody(t, w.Body.String())
			}
		})
	}
}

// TestRolesHandler tests the roles endpoint
func TestRolesHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/roles", nil)
	w := httptest.NewRecorder()

	rolesHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Parse response to validate structure - it's an array, not a map
	var roles []interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &roles); err != nil {
		t.Errorf("Failed to unmarshal roles response: %v", err)
	}

	// Check that we have some roles
	if len(roles) == 0 {
		t.Error("Expected roles data, got empty response")
	}
}

// TestConfigHandler tests the configuration endpoint
func TestConfigHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/config", nil)
	w := httptest.NewRecorder()

	configHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Parse response to validate structure
	var config map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &config); err != nil {
		t.Errorf("Failed to unmarshal config response: %v", err)
	}
}

// TestSearchHandler tests the universal search functionality
func TestSearchHandler(t *testing.T) {
	InitStore()

	// Setup test data
	testPlayers := []Player{
		{
			UID:         "search-1",
			Name:        "Lionel Messi",
			Club:        "Paris Saint-Germain",
			Division:    "Ligue 1",
			Nationality: "Argentina",
			Overall:     93,
		},
		{
			UID:         "search-2",
			Name:        "Cristiano Ronaldo",
			Club:        "Al Nassr",
			Division:    "Saudi Pro League",
			Nationality: "Portugal",
			Overall:     91,
		},
	}

	testDatasetID := "search-test-dataset"
	SetPlayerData(testDatasetID, testPlayers, "$")

	tests := []struct {
		name           string
		query          string
		datasetID      string
		expectedStatus int
		validateBody   func(t *testing.T, body string)
	}{
		{
			name:           "Valid Player Search",
			query:          "Messi",
			datasetID:      testDatasetID,
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				if !strings.Contains(body, "Lionel Messi") {
					t.Errorf("Expected 'Lionel Messi' in search results, got %q", body)
				}
			},
		},
		{
			name:           "Club Search",
			query:          "Paris",
			datasetID:      testDatasetID,
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				if !strings.Contains(body, "Paris Saint-Germain") {
					t.Errorf("Expected 'Paris Saint-Germain' in search results, got %q", body)
				}
			},
		},
		{
			name:           "Empty Query",
			query:          "",
			datasetID:      testDatasetID,
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				if !strings.Contains(body, "[]") {
					t.Errorf("Expected empty array for empty query, got %q", body)
				}
			},
		},
		{
			name:           "Invalid Dataset",
			query:          "Messi",
			datasetID:      "nonexistent",
			expectedStatus: http.StatusNotFound,
			validateBody: func(t *testing.T, body string) {
				if !strings.Contains(body, "Dataset not found") {
					t.Errorf("Expected dataset not found error, got %q", body)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/search/%s?q=%s", tt.datasetID, tt.query)
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			searchHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.validateBody != nil {
				tt.validateBody(t, w.Body.String())
			}
		})
	}
}

// TestCORSHeaders tests CORS header setting
func TestSetCORSHeaders(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		expectedOrigin string
		shouldSetCORS  bool
	}{
		{
			name:           "Allowed localhost:3000",
			origin:         "http://localhost:3000",
			expectedOrigin: "http://localhost:3000",
			shouldSetCORS:  true,
		},
		{
			name:           "Allowed localhost:8080",
			origin:         "http://localhost:8080",
			expectedOrigin: "http://localhost:8080",
			shouldSetCORS:  true,
		},
		{
			name:           "Disallowed origin",
			origin:         "http://malicious-site.com",
			expectedOrigin: "",
			shouldSetCORS:  false,
		},
		{
			name:           "No origin header",
			origin:         "",
			expectedOrigin: "",
			shouldSetCORS:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			w := httptest.NewRecorder()
			setCORSHeaders(w, req)

			corsOrigin := w.Header().Get("Access-Control-Allow-Origin")

			if tt.shouldSetCORS {
				if corsOrigin != tt.expectedOrigin {
					t.Errorf("Expected CORS origin %q, got %q", tt.expectedOrigin, corsOrigin)
				}
				if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
					t.Error("Expected Access-Control-Allow-Credentials to be true")
				}
			} else {
				if corsOrigin != "" {
					t.Errorf("Expected no CORS origin header, got %q", corsOrigin)
				}
			}
		})
	}
}

// TestValidateEnvironmentVariables tests environment variable validation
func TestValidateEnvironmentVariables(t *testing.T) {
	// Save original environment
	originalOTEL := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	originalS3 := os.Getenv("S3_ENDPOINT")
	originalService := os.Getenv("SERVICE_NAME")

	defer func() {
		// Restore original environment
		if originalOTEL != "" {
			os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", originalOTEL)
		} else {
			os.Unsetenv("OTEL_EXPORTER_OTLP_ENDPOINT")
		}
		if originalS3 != "" {
			os.Setenv("S3_ENDPOINT", originalS3)
		} else {
			os.Unsetenv("S3_ENDPOINT")
		}
		if originalService != "" {
			os.Setenv("SERVICE_NAME", originalService)
		} else {
			os.Unsetenv("SERVICE_NAME")
		}
	}()

	tests := []struct {
		name        string
		setup       func()
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid environment",
			setup: func() {
				os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317")
				os.Setenv("S3_ENDPOINT", "localhost:9000")
				os.Setenv("SERVICE_NAME", "test-service")
			},
			expectError: false,
		},
		{
			name: "Invalid OTEL endpoint",
			setup: func() {
				os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "invalid-endpoint")
				os.Unsetenv("S3_ENDPOINT")
				os.Unsetenv("SERVICE_NAME")
			},
			expectError: true,
			errorMsg:    "invalid OTEL_EXPORTER_OTLP_ENDPOINT",
		},
		{
			name: "Invalid S3 endpoint",
			setup: func() {
				os.Unsetenv("OTEL_EXPORTER_OTLP_ENDPOINT")
				os.Setenv("S3_ENDPOINT", "invalid")
				os.Unsetenv("SERVICE_NAME")
			},
			expectError: true,
			errorMsg:    "invalid S3_ENDPOINT format",
		},
		{
			name: "Invalid service name",
			setup: func() {
				os.Unsetenv("OTEL_EXPORTER_OTLP_ENDPOINT")
				os.Unsetenv("S3_ENDPOINT")
				os.Setenv("SERVICE_NAME", "service with spaces")
			},
			expectError: true,
			errorMsg:    "invalid SERVICE_NAME",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			err := validateEnvironmentVariables()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestGetMaxUploadSize tests upload size limit configuration
func TestGetMaxUploadSize(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected int64
	}{
		{
			name:     "Default value when env not set",
			envValue: "",
			expected: 15 * 1024 * 1024, // 15MB
		},
		{
			name:     "Valid environment value",
			envValue: "20",
			expected: 20 * 1024 * 1024, // 20MB
		},
		{
			name:     "Zero value defaults to 15MB",
			envValue: "0",
			expected: 15 * 1024 * 1024, // 15MB
		},
		{
			name:     "Negative value defaults to 15MB",
			envValue: "-5",
			expected: 15 * 1024 * 1024, // 15MB
		},
		{
			name:     "Invalid string defaults to 15MB",
			envValue: "invalid",
			expected: 15 * 1024 * 1024, // 15MB
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Store original value
			original := os.Getenv("MAX_UPLOAD_SIZE")
			defer func() {
				if original == "" {
					os.Unsetenv("MAX_UPLOAD_SIZE")
				} else {
					os.Setenv("MAX_UPLOAD_SIZE", original)
				}
			}()

			// Set test value
			if tt.envValue == "" {
				os.Unsetenv("MAX_UPLOAD_SIZE")
			} else {
				os.Setenv("MAX_UPLOAD_SIZE", tt.envValue)
			}

			result := getMaxUploadSize()
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

// TestGetFileSizeLimitErrorMessage tests file size limit error message configuration
func TestGetFileSizeLimitErrorMessage(t *testing.T) {
	tests := []struct {
		name          string
		envValue      string
		expectedMB    int64
		shouldContain string
	}{
		{
			name:          "Default 15MB message",
			envValue:      "",
			expectedMB:    15,
			shouldContain: "Max file size: 15MB",
		},
		{
			name:          "Custom 25MB message",
			envValue:      "25",
			expectedMB:    25,
			shouldContain: "Max file size: 25MB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Store original value
			original := os.Getenv("MAX_UPLOAD_SIZE")
			defer func() {
				if original == "" {
					os.Unsetenv("MAX_UPLOAD_SIZE")
				} else {
					os.Setenv("MAX_UPLOAD_SIZE", original)
				}
			}()

			// Set test value
			if tt.envValue == "" {
				os.Unsetenv("MAX_UPLOAD_SIZE")
			} else {
				os.Setenv("MAX_UPLOAD_SIZE", tt.envValue)
			}

			result := getFileSizeLimitErrorMessage()
			if !strings.Contains(result, tt.shouldContain) {
				t.Errorf("Expected message to contain '%s', got '%s'", tt.shouldContain, result)
			}
			if !strings.Contains(result, "10,000 players") {
				t.Errorf("Expected message to mention player limit, got '%s'", result)
			}
		})
	}
}

// TestCalculateOptimalBufferSize tests buffer size calculation
func TestCalculateOptimalBufferSize(t *testing.T) {
	tests := []struct {
		name        string
		numWorkers  int
		fileSize    int64
		expectedMin int
		expectedMax int
	}{
		{
			name:        "Small file, few workers",
			numWorkers:  2,
			fileSize:    1024 * 1024, // 1MB
			expectedMin: 20,
			expectedMax: 50,
		},
		{
			name:        "Large file, many workers",
			numWorkers:  8,
			fileSize:    100 * 1024 * 1024, // 100MB
			expectedMin: 100,
			expectedMax: 1000,
		},
		{
			name:        "Very large file should cap at max",
			numWorkers:  10,
			fileSize:    500 * 1024 * 1024, // 500MB
			expectedMin: 150,
			expectedMax: 1000,
		},
		{
			name:        "Small buffer should use minimum",
			numWorkers:  1,
			fileSize:    100, // Very small file
			expectedMin: 20,
			expectedMax: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateOptimalBufferSize(tt.numWorkers, tt.fileSize)
			if result < tt.expectedMin || result > tt.expectedMax {
				t.Errorf("Expected buffer size between %d and %d, got %d",
					tt.expectedMin, tt.expectedMax, result)
			}
			// Ensure result is within absolute bounds
			if result < 20 || result > 1000 {
				t.Errorf("Buffer size %d is outside absolute bounds [20, 1000]", result)
			}
		})
	}
}

// Helper functions for testing

// createMockHTMLFile creates a multipart form file for testing
func createMockHTMLFile(filename, content string) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("playerFile", filename)
	if err != nil {
		return nil, "", err
	}

	_, err = io.WriteString(part, content)
	if err != nil {
		return nil, "", err
	}

	err = writer.Close()
	if err != nil {
		return nil, "", err
	}

	return body, writer.FormDataContentType(), nil
}

// createMockPlayerHTML creates mock HTML content for testing
func createMockPlayerHTML() string {
	return `<!DOCTYPE html>
<html>
<body>
<table>
<tr class="player-row">
	<td class="player-name">Test Player</td>
	<td class="player-position">ST</td>
	<td class="player-age">25</td>
	<td class="player-club">Test FC</td>
	<td class="player-division">Premier League</td>
	<td class="player-value">£5.0M</td>
	<td class="player-wage">£50K</td>
	<td class="player-nationality">England</td>
</tr>
</table>
</body>
</html>`
}

// BenchmarkHandlers provides performance benchmarks for handlers
func BenchmarkUploadHandler(b *testing.B) {
	InitStore()

	// Create test file content
	body, contentType, err := createMockHTMLFile("benchmark.html", createMockPlayerHTML())
	if err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/upload", bytes.NewReader(body.Bytes()))
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()

		uploadHandler(w, req)

		if w.Code != http.StatusOK {
			b.Errorf("Upload failed with status %d", w.Code)
		}
	}
}

func BenchmarkPlayerDataHandler(b *testing.B) {
	InitStore()

	// Setup test data
	testPlayers := createTestPlayers(1000)
	testDatasetID := "benchmark-dataset"
	SetPlayerData(testDatasetID, testPlayers, "$")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/players/"+testDatasetID, nil)
		w := httptest.NewRecorder()

		playerDataHandler(w, req)

		if w.Code != http.StatusOK {
			b.Errorf("Player data request failed with status %d", w.Code)
		}
	}
}
