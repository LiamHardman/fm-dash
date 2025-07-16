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
	"strings"
	"sync"
	"testing"
	"time"
)

// TestEndToEndSystemFunctionality performs comprehensive end-to-end testing
// of the complete data flow from CSV upload to API response
func TestEndToEndSystemFunctionality(t *testing.T) {
	// Initialize test environment
	InitStore()
	InitInMemoryCache()
	InitCacheStorage()
	InitializeMemoryOptimizations()

	// Test with both storage backends
	testCases := []struct {
		name           string
		useProtobuf    bool
		envVarValue    string
	}{
		{
			name:           "JSON Storage Backend",
			useProtobuf:    false,
			envVarValue:    "false",
		},
		{
			name:           "Protobuf Storage Backend",
			useProtobuf:    true,
			envVarValue:    "true",
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

			// Run complete end-to-end test
			testCompleteDataFlow(t, tc.useProtobuf)
		})
	}
}

// testCompleteDataFlow tests the complete data flow from upload to API response
func testCompleteDataFlow(t *testing.T, useProtobuf bool) {
	// Step 1: Upload HTML file
	t.Logf("Step 1: Testing HTML upload with %s backend", getBackendName(useProtobuf))
	datasetID := testHTMLUploadFlow(t, useProtobuf)
	if datasetID == "" {
		t.Fatal("Failed to upload HTML file")
		return
	}

	// Step 2: Wait for async processing to complete
	t.Logf("Step 2: Waiting for async processing to complete")
	waitForAsyncProcessing(t, datasetID, 30*time.Second)

	// Step 3: Test data retrieval
	t.Logf("Step 3: Testing data retrieval")
	players := testDataRetrievalFlow(t, datasetID, useProtobuf)
	if len(players) == 0 {
		t.Fatal("No players retrieved")
		return
	}

	// Step 4: Test data integrity
	t.Logf("Step 4: Testing data integrity")
	testDataIntegrityFlow(t, players, useProtobuf)

	// Step 5: Test API endpoints with real data
	t.Logf("Step 5: Testing API endpoints")
	testAPIEndpointsFlow(t, datasetID, useProtobuf)

	// Step 6: Test filtering and search functionality
	t.Logf("Step 6: Testing filtering and search")
	testFilteringAndSearchFlow(t, datasetID, useProtobuf)

	// Step 7: Test performance under load
	t.Logf("Step 7: Testing performance under load")
	testPerformanceUnderLoad(t, datasetID, useProtobuf)

	// Step 8: Test monitoring and logging
	t.Logf("Step 8: Testing monitoring and logging")
	testMonitoringAndLogging(t, datasetID, useProtobuf)

	t.Logf("End-to-end test completed successfully with %s backend", getBackendName(useProtobuf))
}

// testHTMLUploadFlow tests the HTML upload process
func testHTMLUploadFlow(t *testing.T, useProtobuf bool) string {
	// Create comprehensive test HTML content
	testHTML := createComprehensiveTestHTML()

	// Create multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	fileWriter, err := writer.CreateFormFile("playerFile", "comprehensive_test.html")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	_, err = fileWriter.Write([]byte(testHTML))
	if err != nil {
		t.Fatalf("Failed to write test HTML: %v", err)
	}

	writer.Close()

	// Create HTTP request
	req := httptest.NewRequest("POST", "/api/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	uploadHandler(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Fatalf("Upload failed with %s backend. Status: %d, Response: %s",
			getBackendName(useProtobuf), w.Code, w.Body.String())
	}

	var response UploadResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse upload response: %v", err)
	}

	if response.DatasetID == "" {
		t.Fatal("DatasetID is empty")
	}

	t.Logf("Upload successful. DatasetID: %s, Currency: %s",
		response.DatasetID, response.DetectedCurrencySymbol)

	return response.DatasetID
}

// waitForAsyncProcessing waits for async processing to complete
func waitForAsyncProcessing(t *testing.T, datasetID string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Logf("Async processing timeout reached, continuing with test")
			return
		case <-ticker.C:
			// Check if percentiles are calculated by making a request
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/percentiles-status/%s", datasetID), nil)
			w := httptest.NewRecorder()

			percentilesStatusHandler(w, req)

			if w.Code == http.StatusOK {
				var status map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &status); err == nil {
					if ready, ok := status["ready"].(bool); ok && ready {
						t.Logf("Async processing completed")
						return
					}
				}
			}
		}
	}
}

// testDataRetrievalFlow tests data retrieval functionality
func testDataRetrievalFlow(t *testing.T, datasetID string, useProtobuf bool) []Player {
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/players/%s", datasetID), nil)
	w := httptest.NewRecorder()

	playerDataHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Data retrieval failed with %s backend. Status: %d, Response: %s",
			getBackendName(useProtobuf), w.Code, w.Body.String())
	}

	var response struct {
		Players        []Player `json:"players"`
		CurrencySymbol string   `json:"currencySymbol"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse player data response: %v", err)
	}

	if len(response.Players) == 0 {
		t.Fatal("No players in response")
	}

	t.Logf("Retrieved %d players with currency %s",
		len(response.Players), response.CurrencySymbol)

	return response.Players
}

// testDataIntegrityFlow validates data integrity throughout the system
func testDataIntegrityFlow(t *testing.T, players []Player, useProtobuf bool) {
	// Check that all players have required fields
	for i, player := range players {
		if player.UID == 0 {
			t.Errorf("Player %d has zero UID", i)
		}

		if player.Name == "" {
			t.Errorf("Player %d has empty name", i)
		}

		if player.Position == "" {
			t.Errorf("Player %d has empty position", i)
		}

		if player.Club == "" {
			t.Errorf("Player %d has empty club", i)
		}

		// Check numeric attributes are reasonable
		if player.Overall < 1 || player.Overall > 20 {
			t.Errorf("Player %d has invalid overall rating: %d", i, player.Overall)
		}

		// Check maps are initialized
		if player.Attributes == nil {
			t.Errorf("Player %d has nil Attributes map", i)
		}

		if player.NumericAttributes == nil {
			t.Errorf("Player %d has nil NumericAttributes map", i)
		}

		// Check position parsing
		if len(player.ParsedPositions) == 0 {
			t.Errorf("Player %d has no parsed positions", i)
		}

		// Check that performance stats are present
		if player.PerformanceStatsNumeric == nil {
			t.Errorf("Player %d has nil PerformanceStatsNumeric", i)
		}

		// Validate transfer value and wage parsing
		if player.TransferValueAmount == 0 && player.TransferValue != "" && player.TransferValue != "0" {
			t.Errorf("Player %d has unparsed transfer value: %s", i, player.TransferValue)
		}

		if player.WageAmount == 0 && player.Wage != "" && player.Wage != "0" {
			t.Errorf("Player %d has unparsed wage: %s", i, player.Wage)
		}
	}

	t.Logf("Data integrity validated for %d players", len(players))
}

// testAPIEndpointsFlow tests all API endpoints with real data
func testAPIEndpointsFlow(t *testing.T, datasetID string, useProtobuf bool) {
	// Test roles endpoint
	req := httptest.NewRequest("GET", "/api/roles", nil)
	w := httptest.NewRecorder()
	rolesHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Roles endpoint failed: %d", w.Code)
	}

	// Test leagues endpoint
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/leagues/%s", datasetID), nil)
	w = httptest.NewRecorder()
	leaguesHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Leagues endpoint failed: %d", w.Code)
	}

	// Parse leagues response to get a league name for teams test
	var leagues []League
	if err := json.Unmarshal(w.Body.Bytes(), &leagues); err == nil && len(leagues) > 0 {
		// Test teams endpoint
		leagueName := leagues[0].Name
		req = httptest.NewRequest("GET", fmt.Sprintf("/api/teams/%s?league=%s", datasetID, leagueName), nil)
		w = httptest.NewRecorder()
		teamsHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Teams endpoint failed: %d", w.Code)
		}
	}

	// Test config endpoint
	req = httptest.NewRequest("GET", "/api/config", nil)
	w = httptest.NewRecorder()
	cachedConfigHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Config endpoint failed: %d", w.Code)
	}

	// Test cache status endpoint
	req = httptest.NewRequest("GET", "/api/cache-status", nil)
	w = httptest.NewRecorder()
	cacheStatusHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Cache status endpoint failed: %d", w.Code)
	}

	t.Logf("All API endpoints tested successfully")
}

// testFilteringAndSearchFlow tests filtering and search functionality
func testFilteringAndSearchFlow(t *testing.T, datasetID string, useProtobuf bool) {
	// Test various filters
	filters := []string{
		"?position=GK",
		"?position=D",
		"?position=M",
		"?position=F",
		"?role=Goalkeeper",
		"?role=Centre+Back",
		"?minAge=18&maxAge=25",
		"?minAge=26&maxAge=35",
		"?divisionFilter=all",
		"?divisionFilter=same&targetDivision=Premier+League",
		"?positionCompare=all",
		"?positionCompare=broad",
		"?positionCompare=detailed",
	}

	for _, filter := range filters {
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/players/%s%s", datasetID, filter), nil)
		w := httptest.NewRecorder()

		playerDataHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Filter test failed for %s: %d", filter, w.Code)
			continue
		}

		// Ensure response is valid JSON
		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Invalid JSON response for filter %s: %v", filter, err)
		}
	}

	// Test search functionality
	searchQueries := []string{
		"?q=test&type=players",
		"?q=premier&type=teams",
		"?q=league&type=leagues",
		"?q=england&type=nations",
		"?q=goalkeeper",
	}

	for _, query := range searchQueries {
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/search/%s%s", datasetID, query), nil)
		w := httptest.NewRecorder()

		searchHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Search test failed for %s: %d", query, w.Code)
			continue
		}

		// Ensure response is valid JSON
		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Invalid JSON response for search %s: %v", query, err)
		}
	}

	t.Logf("Filtering and search functionality tested successfully")
}

// testPerformanceUnderLoad tests system behavior under various load conditions
func testPerformanceUnderLoad(t *testing.T, datasetID string, useProtobuf bool) {
	// Test concurrent requests
	concurrency := 10
	requestsPerWorker := 5

	var wg sync.WaitGroup
	errors := make(chan error, concurrency*requestsPerWorker)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < requestsPerWorker; j++ {
				// Make concurrent requests to different endpoints
				endpoints := []string{
					fmt.Sprintf("/api/players/%s", datasetID),
					fmt.Sprintf("/api/players/%s?position=GK", datasetID),
					fmt.Sprintf("/api/leagues/%s", datasetID),
					"/api/roles",
					"/api/config",
				}

				endpoint := endpoints[j%len(endpoints)]
				req := httptest.NewRequest("GET", endpoint, nil)
				w := httptest.NewRecorder()

				switch {
				case strings.Contains(endpoint, "/players/"):
					playerDataHandler(w, req)
				case strings.Contains(endpoint, "/leagues/"):
					leaguesHandler(w, req)
				case strings.Contains(endpoint, "/roles"):
					rolesHandler(w, req)
				case strings.Contains(endpoint, "/config"):
					cachedConfigHandler(w, req)
				}

				if w.Code != http.StatusOK {
					errors <- fmt.Errorf("worker %d request %d failed: %d", workerID, j, w.Code)
				}
			}
		}(i)
	}

	// Wait for all workers to complete
	wg.Wait()
	close(errors)

	// Check for errors
	errorCount := 0
	for err := range errors {
		t.Errorf("Concurrent request error: %v", err)
		errorCount++
	}

	if errorCount > 0 {
		t.Errorf("Performance test failed with %d errors out of %d requests",
			errorCount, concurrency*requestsPerWorker)
	} else {
		t.Logf("Performance test passed: %d concurrent requests completed successfully",
			concurrency*requestsPerWorker)
	}
}

// testMonitoringAndLogging verifies monitoring and logging work correctly
func testMonitoringAndLogging(t *testing.T, datasetID string, useProtobuf bool) {
	// Test that requests generate appropriate logs and metrics
	// This is a basic test - in a real environment you'd check actual log output

	// Make a request that should generate logs
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/players/%s", datasetID), nil)
	w := httptest.NewRecorder()

	// Add request context for tracing
	ctx := req.Context()
	ctx, span := StartSpan(ctx, "test.monitoring")
	req = req.WithContext(ctx)

	playerDataHandler(w, req)
	span.End()

	if w.Code != http.StatusOK {
		t.Errorf("Monitoring test request failed: %d", w.Code)
		return
	}

	// Test error logging by making an invalid request
	req = httptest.NewRequest("GET", "/api/players/invalid-dataset-id", nil)
	w = httptest.NewRecorder()

	ctx = req.Context()
	ctx, span = StartSpan(ctx, "test.error_monitoring")
	req = req.WithContext(ctx)

	playerDataHandler(w, req)
	span.End()

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for invalid dataset, got %d", w.Code)
	}

	t.Logf("Monitoring and logging test completed")
}

// createComprehensiveTestHTML creates comprehensive test HTML with various player types
func createComprehensiveTestHTML() string {
	return `<html>
<head><title>Comprehensive Test Players</title></head>
<body>
<table>
<tr>
	<th>Inf</th>
	<th>Name</th>
	<th>Nat</th>
	<th>Age</th>
	<th>Position</th>
	<th>Club</th>
	<th>Transfer Value</th>
	<th>Wage</th>
	<th>Ability</th>
	<th>Potential</th>
	<th>Personality</th>
	<th>Media Handling</th>
	<th>Av Rat</th>
	<th>Apps</th>
	<th>Left Foot</th>
	<th>Right Foot</th>
	<th>Acc</th>
	<th>Agg</th>
	<th>Agi</th>
	<th>Ant</th>
	<th>Bal</th>
	<th>Bra</th>
	<th>Cnt</th>
	<th>Cmp</th>
	<th>Cro</th>
	<th>Dec</th>
	<th>Det</th>
	<th>Dri</th>
	<th>Fin</th>
	<th>Fir</th>
	<th>Fla</th>
	<th>Hea</th>
	<th>L Th</th>
	<th>Jum</th>
	<th>Ldr</th>
	<th>Lon</th>
	<th>Mar</th>
	<th>OtB</th>
	<th>Pac</th>
	<th>Pas</th>
	<th>Pos</th>
	<th>Sta</th>
	<th>Str</th>
	<th>Tck</th>
	<th>Tea</th>
	<th>Tec</th>
	<th>Vis</th>
	<th>Wor</th>
	<th>Division</th>
	<th>UID</th>
</tr>
<tr>
	<td>Hol</td>
	<td>Test Goalkeeper</td>
	<td>ENG</td>
	<td>25</td>
	<td>GK</td>
	<td>Test FC</td>
	<td>£5M</td>
	<td>£50K</td>
	<td>15</td>
	<td>18</td>
	<td>Balanced</td>
	<td>Evasive</td>
	<td>7.5</td>
	<td>30</td>
	<td>10</td>
	<td>15</td>
	<td>12</td>
	<td>8</td>
	<td>14</td>
	<td>16</td>
	<td>13</td>
	<td>11</td>
	<td>15</td>
	<td>12</td>
	<td>9</td>
	<td>14</td>
	<td>13</td>
	<td>10</td>
	<td>8</td>
	<td>12</td>
	<td>7</td>
	<td>11</td>
	<td>6</td>
	<td>13</td>
	<td>9</td>
	<td>8</td>
	<td>7</td>
	<td>10</td>
	<td>14</td>
	<td>12</td>
	<td>16</td>
	<td>15</td>
	<td>11</td>
	<td>6</td>
	<td>13</td>
	<td>14</td>
	<td>12</td>
	<td>10</td>
	<td>Premier League</td>
	<td>100001</td>
</tr>
<tr>
	<td>Hol</td>
	<td>Test Centre Back</td>
	<td>ESP</td>
	<td>22</td>
	<td>D (C)</td>
	<td>Test United</td>
	<td>£15M</td>
	<td>£75K</td>
	<td>16</td>
	<td>19</td>
	<td>Driven</td>
	<td>Outspoken</td>
	<td>7.8</td>
	<td>28</td>
	<td>12</td>
	<td>14</td>
	<td>15</td>
	<td>16</td>
	<td>13</td>
	<td>17</td>
	<td>14</td>
	<td>15</td>
	<td>16</td>
	<td>13</td>
	<td>11</td>
	<td>15</td>
	<td>14</td>
	<td>12</td>
	<td>9</td>
	<td>13</td>
	<td>8</td>
	<td>14</td>
	<td>7</td>
	<td>15</td>
	<td>10</td>
	<td>9</td>
	<td>16</td>
	<td>11</td>
	<td>15</td>
	<td>13</td>
	<td>17</td>
	<td>16</td>
	<td>18</td>
	<td>15</td>
	<td>14</td>
	<td>15</td>
	<td>13</td>
	<td>11</td>
	<td>Premier League</td>
	<td>100002</td>
</tr>
<tr>
	<td>Hol</td>
	<td>Test Midfielder</td>
	<td>BRA</td>
	<td>28</td>
	<td>M (C)</td>
	<td>Test City</td>
	<td>£25M</td>
	<td>£100K</td>
	<td>17</td>
	<td>17</td>
	<td>Professional</td>
	<td>Composed</td>
	<td>8.2</td>
	<td>32</td>
	<td>16</td>
	<td>14</td>
	<td>16</td>
	<td>12</td>
	<td>15</td>
	<td>18</td>
	<td>16</td>
	<td>14</td>
	<td>17</td>
	<td>15</td>
	<td>13</td>
	<td>16</td>
	<td>15</td>
	<td>17</td>
	<td>14</td>
	<td>16</td>
	<td>12</td>
	<td>13</td>
	<td>11</td>
	<td>14</td>
	<td>16</td>
	<td>15</td>
	<td>12</td>
	<td>14</td>
	<td>16</td>
	<td>18</td>
	<td>15</td>
	<td>16</td>
	<td>13</td>
	<td>11</td>
	<td>15</td>
	<td>17</td>
	<td>16</td>
	<td>14</td>
	<td>Premier League</td>
	<td>100003</td>
</tr>
<tr>
	<td>Hol</td>
	<td>Test Forward</td>
	<td>ARG</td>
	<td>24</td>
	<td>F (C)</td>
	<td>Test Athletic</td>
	<td>£35M</td>
	<td>£125K</td>
	<td>18</td>
	<td>20</td>
	<td>Ambitious</td>
	<td>Charismatic</td>
	<td>8.5</td>
	<td>35</td>
	<td>15</td>
	<td>17</td>
	<td>17</td>
	<td>14</td>
	<td>16</td>
	<td>18</td>
	<td>15</td>
	<td>16</td>
	<td>13</td>
	<td>14</td>
	<td>15</td>
	<td>17</td>
	<td>16</td>
	<td>18</td>
	<td>19</td>
	<td>17</td>
	<td>16</td>
	<td>15</td>
	<td>14</td>
	<td>13</td>
	<td>15</td>
	<td>12</td>
	<td>11</td>
	<td>18</td>
	<td>18</td>
	<td>16</td>
	<td>14</td>
	<td>15</td>
	<td>13</td>
	<td>10</td>
	<td>14</td>
	<td>17</td>
	<td>16</td>
	<td>15</td>
	<td>Premier League</td>
	<td>100004</td>
</tr>
<tr>
	<td>Hol</td>
	<td>Test Wingback</td>
	<td>FRA</td>
	<td>26</td>
	<td>D/M (RL), WB (RL)</td>
	<td>Test Rovers</td>
	<td>£12M</td>
	<td>£65K</td>
	<td>16</td>
	<td>17</td>
	<td>Resolute</td>
	<td>Respectful</td>
	<td>7.9</td>
	<td>29</td>
	<td>13</td>
	<td>15</td>
	<td>16</td>
	<td>13</td>
	<td>17</td>
	<td>15</td>
	<td>16</td>
	<td>14</td>
	<td>15</td>
	<td>14</td>
	<td>16</td>
	<td>15</td>
	<td>14</td>
	<td>15</td>
	<td>12</td>
	<td>14</td>
	<td>11</td>
	<td>13</td>
	<td>10</td>
	<td>12</td>
	<td>13</td>
	<td>11</td>
	<td>14</td>
	<td>13</td>
	<td>17</td>
	<td>15</td>
	<td>16</td>
	<td>17</td>
	<td>15</td>
	<td>14</td>
	<td>16</td>
	<td>14</td>
	<td>15</td>
	<td>16</td>
	<td>Championship</td>
	<td>100005</td>
</tr>
</table>
</body>
</html>`
}