package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	pb "api/proto"
)

// TestPlayerDataHandlerProtobufSupport tests the enhanced playerDataHandler with protobuf support
func TestPlayerDataHandlerProtobufSupport(t *testing.T) {
	// Setup test data
	testDatasetID := "test-dataset-123"
	testPlayers := []Player{
		{
			UID:                 1,
			Name:               "Test Player 1",
			Position:           "ST",
			Age:                "25",
			Club:               "Test FC",
			Division:           "Premier League",
			TransferValue:      "£50M",
			Wage:               "£100K",
			Nationality:        "England",
			NationalityISO:     "ENG",
			NationalityFIFACode: "ENG",
			Overall:            85,
			PAC:                90,
			SHO:                88,
			PAS:                75,
			DRI:                82,
			DEF:                30,
			PHY:                85,
			TransferValueAmount: 50000000,
			WageAmount:         100000,
			Attributes:         map[string]string{"Finishing": "18", "Pace": "17"},
			NumericAttributes:  map[string]int{"Finishing": 18, "Pace": 17},
			PerformanceStatsNumeric: map[string]float64{"Goals": 25.5, "Assists": 8.2},
			PerformancePercentiles:  map[string]map[string]float64{"Attacking": {"Goals": 95.5, "Assists": 78.3}},
			ParsedPositions:    []string{"ST", "CF"},
			ShortPositions:     []string{"ST"},
			PositionGroups:     []string{"Forward"},
			RoleSpecificOveralls: []RoleOverallScore{
				{RoleName: "Advanced Forward", Score: 87},
				{RoleName: "Poacher", Score: 85},
			},
		},
		{
			UID:                 2,
			Name:               "Test Player 2",
			Position:           "CM",
			Age:                "28",
			Club:               "Test United",
			Division:           "Premier League",
			TransferValue:      "£30M",
			Wage:               "£80K",
			Nationality:        "Spain",
			NationalityISO:     "ESP",
			NationalityFIFACode: "ESP",
			Overall:            82,
			PAC:                70,
			SHO:                75,
			PAS:                90,
			DRI:                85,
			DEF:                78,
			PHY:                80,
			TransferValueAmount: 30000000,
			WageAmount:         80000,
			Attributes:         map[string]string{"Passing": "19", "Vision": "18"},
			NumericAttributes:  map[string]int{"Passing": 19, "Vision": 18},
			PerformanceStatsNumeric: map[string]float64{"Passes": 85.2, "KeyPasses": 3.8},
			PerformancePercentiles:  map[string]map[string]float64{"Midfield": {"Passes": 88.7, "KeyPasses": 82.1}},
			ParsedPositions:    []string{"CM", "CAM"},
			ShortPositions:     []string{"CM"},
			PositionGroups:     []string{"Midfielder"},
			RoleSpecificOveralls: []RoleOverallScore{
				{RoleName: "Deep Lying Playmaker", Score: 84},
				{RoleName: "Box to Box Midfielder", Score: 81},
			},
		},
	}
	testCurrency := "£"

	// Store test data
	SetPlayerData(testDatasetID, testPlayers, testCurrency)
	defer func() {
		// Cleanup
		DeleteDataset(testDatasetID)
	}()

	tests := []struct {
		name           string
		acceptHeader   string
		expectedFormat string
		expectProtobuf bool
	}{
		{
			name:           "JSON request",
			acceptHeader:   "application/json",
			expectedFormat: "application/json",
			expectProtobuf: false,
		},
		{
			name:           "Protobuf request",
			acceptHeader:   "application/x-protobuf",
			expectedFormat: "application/x-protobuf",
			expectProtobuf: true,
		},
		{
			name:           "Protobuf with quality",
			acceptHeader:   "application/x-protobuf;q=0.9, application/json;q=0.8",
			expectedFormat: "application/x-protobuf",
			expectProtobuf: true,
		},
		{
			name:           "JSON fallback",
			acceptHeader:   "text/html, application/json;q=0.9",
			expectedFormat: "application/json",
			expectProtobuf: false,
		},
		{
			name:           "No accept header",
			acceptHeader:   "",
			expectedFormat: "application/json",
			expectProtobuf: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/players/%s", testDatasetID), nil)
			if tt.acceptHeader != "" {
				req.Header.Set("Accept", tt.acceptHeader)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			playerDataHandler(w, req)

			// Check response status
			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
				t.Logf("Response body: %s", w.Body.String())
				return
			}

			// Check content type
			contentType := w.Header().Get("Content-Type")
			if contentType != tt.expectedFormat {
				t.Errorf("Expected content type %s, got %s", tt.expectedFormat, contentType)
			}

			// Verify response data
			if tt.expectProtobuf {
				// Test protobuf response
				var protoResponse pb.PlayerDataResponse
				err := proto.Unmarshal(w.Body.Bytes(), &protoResponse)
				if err != nil {
					t.Fatalf("Failed to unmarshal protobuf response: %v", err)
				}

				// Verify protobuf data
				if len(protoResponse.Players) != len(testPlayers) {
					t.Errorf("Expected %d players, got %d", len(testPlayers), len(protoResponse.Players))
				}

				if protoResponse.CurrencySymbol != testCurrency {
					t.Errorf("Expected currency %s, got %s", testCurrency, protoResponse.CurrencySymbol)
				}

				// Verify metadata
				if protoResponse.Metadata == nil {
					t.Error("Expected metadata in protobuf response")
				} else {
					if protoResponse.Metadata.TotalCount != int32(len(testPlayers)) {
						t.Errorf("Expected total count %d, got %d", len(testPlayers), protoResponse.Metadata.TotalCount)
					}
				}

				// Verify first player data
				if len(protoResponse.Players) > 0 {
					firstPlayer := protoResponse.Players[0]
					if firstPlayer.Name != testPlayers[0].Name {
						t.Errorf("Expected player name %s, got %s", testPlayers[0].Name, firstPlayer.Name)
					}
					if firstPlayer.Overall != int32(testPlayers[0].Overall) {
						t.Errorf("Expected overall %d, got %d", testPlayers[0].Overall, firstPlayer.Overall)
					}
				}
			} else {
				// Test JSON response
				var jsonResponse PlayerDataWithCurrency
				err := json.Unmarshal(w.Body.Bytes(), &jsonResponse)
				if err != nil {
					t.Fatalf("Failed to unmarshal JSON response: %v", err)
				}

				// Verify JSON data
				if len(jsonResponse.Players) != len(testPlayers) {
					t.Errorf("Expected %d players, got %d", len(testPlayers), len(jsonResponse.Players))
				}

				if jsonResponse.CurrencySymbol != testCurrency {
					t.Errorf("Expected currency %s, got %s", testCurrency, jsonResponse.CurrencySymbol)
				}

				// Verify first player data
				if len(jsonResponse.Players) > 0 {
					firstPlayer := jsonResponse.Players[0]
					if firstPlayer.Name != testPlayers[0].Name {
						t.Errorf("Expected player name %s, got %s", testPlayers[0].Name, firstPlayer.Name)
					}
					if firstPlayer.Overall != testPlayers[0].Overall {
						t.Errorf("Expected overall %d, got %d", testPlayers[0].Overall, firstPlayer.Overall)
					}
				}
			}
		})
	}
}

// TestPlayerDataHandlerProtobufFallback tests fallback behavior when protobuf serialization fails
func TestPlayerDataHandlerProtobufFallback(t *testing.T) {
	testDatasetID := "test-fallback-dataset"
	
	// Create test data with potentially problematic values
	testPlayers := []Player{
		{
			UID:      1,
			Name:     "Test Player",
			Position: "ST",
			Age:      "25",
			Overall:  85,
		},
	}
	testCurrency := "£"

	// Store test data
	SetPlayerData(testDatasetID, testPlayers, testCurrency)
	defer func() {
		DeletePlayerData(testDatasetID)
	}()

	// Create request with protobuf preference
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/players/%s", testDatasetID), nil)
	req.Header.Set("Accept", "application/x-protobuf")

	w := httptest.NewRecorder()
	playerDataHandler(w, req)

	// Should succeed with either protobuf or JSON fallback
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		return
	}

	// Response should be valid regardless of format
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/x-protobuf" && contentType != "application/json" {
		t.Errorf("Expected protobuf or JSON content type, got %s", contentType)
	}
}

// TestPlayerDataHandlerErrorHandling tests error handling with protobuf support
func TestPlayerDataHandlerErrorHandling(t *testing.T) {
	tests := []struct {
		name         string
		datasetID    string
		acceptHeader string
		expectedCode int
	}{
		{
			name:         "Dataset not found - JSON",
			datasetID:    "nonexistent-dataset",
			acceptHeader: "application/json",
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "Dataset not found - Protobuf",
			datasetID:    "nonexistent-dataset",
			acceptHeader: "application/x-protobuf",
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "Empty dataset ID - JSON",
			datasetID:    "",
			acceptHeader: "application/json",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Empty dataset ID - Protobuf",
			datasetID:    "",
			acceptHeader: "application/x-protobuf",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var url string
			if tt.datasetID == "" {
				url = "/api/players/"
			} else {
				url = fmt.Sprintf("/api/players/%s", tt.datasetID)
			}

			req := httptest.NewRequest("GET", url, nil)
			req.Header.Set("Accept", tt.acceptHeader)

			w := httptest.NewRecorder()
			playerDataHandler(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Expected status %d, got %d", tt.expectedCode, w.Code)
			}
		})
	}
}

// TestPlayerDataHandlerFiltering tests filtering functionality with protobuf responses
func TestPlayerDataHandlerFiltering(t *testing.T) {
	testDatasetID := "test-filtering-dataset"
	
	// Create test data with different positions and roles
	testPlayers := []Player{
		{
			UID:          1,
			Name:         "Striker",
			Position:     "ST",
			Age:          "25",
			Overall:      85,
			ShortPositions: []string{"ST"},
			RoleSpecificOveralls: []RoleOverallScore{
				{RoleName: "Advanced Forward", Score: 87},
			},
		},
		{
			UID:          2,
			Name:         "Midfielder",
			Position:     "CM",
			Age:          "28",
			Overall:      82,
			ShortPositions: []string{"CM"},
			RoleSpecificOveralls: []RoleOverallScore{
				{RoleName: "Deep Lying Playmaker", Score: 84},
			},
		},
	}
	testCurrency := "£"

	SetPlayerData(testDatasetID, testPlayers, testCurrency)
	defer func() {
		DeletePlayerData(testDatasetID)
	}()

	tests := []struct {
		name           string
		queryParams    string
		acceptHeader   string
		expectedCount  int
	}{
		{
			name:          "Filter by position - JSON",
			queryParams:   "position=ST",
			acceptHeader:  "application/json",
			expectedCount: 1,
		},
		{
			name:          "Filter by position - Protobuf",
			queryParams:   "position=ST",
			acceptHeader:  "application/x-protobuf",
			expectedCount: 1,
		},
		{
			name:          "Filter by role - JSON",
			queryParams:   "role=Advanced Forward",
			acceptHeader:  "application/json",
			expectedCount: 1,
		},
		{
			name:          "Filter by role - Protobuf",
			queryParams:   "role=Advanced Forward",
			acceptHeader:  "application/x-protobuf",
			expectedCount: 1,
		},
		{
			name:          "No filters - JSON",
			queryParams:   "",
			acceptHeader:  "application/json",
			expectedCount: 2,
		},
		{
			name:          "No filters - Protobuf",
			queryParams:   "",
			acceptHeader:  "application/x-protobuf",
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/players/%s", testDatasetID)
			if tt.queryParams != "" {
				url += "?" + tt.queryParams
			}

			req := httptest.NewRequest("GET", url, nil)
			req.Header.Set("Accept", tt.acceptHeader)

			w := httptest.NewRecorder()
			playerDataHandler(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
				return
			}

			var playerCount int
			if strings.Contains(tt.acceptHeader, "protobuf") {
				var protoResponse pb.PlayerDataResponse
				err := proto.Unmarshal(w.Body.Bytes(), &protoResponse)
				if err != nil {
					t.Fatalf("Failed to unmarshal protobuf response: %v", err)
				}
				playerCount = len(protoResponse.Players)
			} else {
				var jsonResponse PlayerDataWithCurrency
				err := json.Unmarshal(w.Body.Bytes(), &jsonResponse)
				if err != nil {
					t.Fatalf("Failed to unmarshal JSON response: %v", err)
				}
				playerCount = len(jsonResponse.Players)
			}

			if playerCount != tt.expectedCount {
				t.Errorf("Expected %d players, got %d", tt.expectedCount, playerCount)
			}
		})
	}
}

// TestPlayerDataHandlerPerformanceMetrics tests performance tracking for protobuf vs JSON
func TestPlayerDataHandlerPerformanceMetrics(t *testing.T) {
	testDatasetID := "test-performance-dataset"
	
	// Create larger test dataset for performance comparison
	testPlayers := make([]Player, 100)
	for i := 0; i < 100; i++ {
		testPlayers[i] = Player{
			UID:                 int64(i + 1),
			Name:               fmt.Sprintf("Player %d", i+1),
			Position:           "ST",
			Age:                "25",
			Club:               "Test FC",
			Overall:            80 + (i % 20),
			TransferValueAmount: int64(1000000 * (i + 1)),
			WageAmount:         int64(10000 * (i + 1)),
			Attributes:         map[string]string{"Finishing": "15", "Pace": "16"},
			NumericAttributes:  map[string]int{"Finishing": 15, "Pace": 16},
			ShortPositions:     []string{"ST"},
		}
	}
	testCurrency := "£"

	SetPlayerData(testDatasetID, testPlayers, testCurrency)
	defer func() {
		DeletePlayerData(testDatasetID)
	}()

	// Test JSON performance
	jsonStart := time.Now()
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/players/%s", testDatasetID), nil)
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	playerDataHandler(w, req)
	jsonDuration := time.Since(jsonStart)
	jsonSize := w.Body.Len()

	if w.Code != http.StatusOK {
		t.Fatalf("JSON request failed with status %d", w.Code)
	}

	// Test Protobuf performance
	protobufStart := time.Now()
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/players/%s", testDatasetID), nil)
	req.Header.Set("Accept", "application/x-protobuf")
	w = httptest.NewRecorder()
	playerDataHandler(w, req)
	protobufDuration := time.Since(protobufStart)
	protobufSize := w.Body.Len()

	if w.Code != http.StatusOK {
		t.Fatalf("Protobuf request failed with status %d", w.Code)
	}

	// Log performance comparison
	t.Logf("Performance comparison for %d players:", len(testPlayers))
	t.Logf("JSON: %v duration, %d bytes", jsonDuration, jsonSize)
	t.Logf("Protobuf: %v duration, %d bytes", protobufDuration, protobufSize)
	
	if protobufSize > 0 && jsonSize > 0 {
		compressionRatio := float64(protobufSize) / float64(jsonSize)
		t.Logf("Compression ratio: %.2f (protobuf/json)", compressionRatio)
		
		// Protobuf should generally be smaller
		if compressionRatio > 1.0 {
			t.Logf("Warning: Protobuf response is larger than JSON (ratio: %.2f)", compressionRatio)
		}
	}
}

// BenchmarkPlayerDataHandlerJSON benchmarks JSON response performance
func BenchmarkPlayerDataHandlerJSON(b *testing.B) {
	testDatasetID := "bench-json-dataset"
	testPlayers := make([]Player, 50)
	for i := 0; i < 50; i++ {
		testPlayers[i] = Player{
			UID:      int64(i + 1),
			Name:     fmt.Sprintf("Player %d", i+1),
			Position: "ST",
			Age:      "25",
			Overall:  80,
		}
	}
	SetPlayerData(testDatasetID, testPlayers, "£")
	defer DeletePlayerData(testDatasetID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/players/%s", testDatasetID), nil)
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()
		playerDataHandler(w, req)
	}
}

// BenchmarkPlayerDataHandlerProtobuf benchmarks protobuf response performance
func BenchmarkPlayerDataHandlerProtobuf(b *testing.B) {
	testDatasetID := "bench-protobuf-dataset"
	testPlayers := make([]Player, 50)
	for i := 0; i < 50; i++ {
		testPlayers[i] = Player{
			UID:      int64(i + 1),
			Name:     fmt.Sprintf("Player %d", i+1),
			Position: "ST",
			Age:      "25",
			Overall:  80,
		}
	}
	SetPlayerData(testDatasetID, testPlayers, "£")
	defer DeletePlayerData(testDatasetID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/players/%s", testDatasetID), nil)
		req.Header.Set("Accept", "application/x-protobuf")
		w := httptest.NewRecorder()
		playerDataHandler(w, req)
	}
}