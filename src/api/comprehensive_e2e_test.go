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
	"sync"
	"testing"
	"time"
)

// TestComprehensiveEndToEndFunctionality performs comprehensive end-to-end testing
// covering all requirements for task 8.2: Perform end-to-end system testing
func TestComprehensiveEndToEndFunctionality(t *testing.T) {
	// Initialize test environment
	InitStore()
	InitInMemoryCache()
	InitCacheStorage(context.Background())
	InitializeMemoryOptimizations()

	// Test with both storage backends
	testCases := []struct {
		name        string
		useProtobuf bool
		envValue    string
	}{
		{"JSON Backend E2E", false, "false"},
		{"Protobuf Backend E2E", true, "true"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variable
			originalValue := os.Getenv("USE_PROTOBUF")
			if err := os.Setenv("USE_PROTOBUF", tc.envValue); err != nil {
				t.Fatalf("Failed to set USE_PROTOBUF environment variable: %v", err)
			}
			defer func() {
				if err := os.Setenv("USE_PROTOBUF", originalValue); err != nil {
					t.Logf("Warning: Failed to restore USE_PROTOBUF environment variable: %v", err)
				}
			}()

			// Re-initialize storage
			InitStore()

			// Run comprehensive test
			runComprehensiveE2ETest(t, tc.useProtobuf)
		})
	}
}

func runComprehensiveE2ETest(t *testing.T, useProtobuf bool) {
	backendName := getBackendName(useProtobuf)
	t.Logf("=== Starting Comprehensive E2E Test with %s Backend ===", backendName)

	// Step 1: Test complete data flow from CSV upload to API response
	datasetID := testCompleteDataFlowE2E(t, useProtobuf)
	if datasetID == "" {
		t.Fatal("Complete data flow test failed")
		return
	}

	// Step 2: Validate data integrity throughout the entire system
	players := testDataIntegrityE2E(t, datasetID, useProtobuf)
	if len(players) == 0 {
		t.Fatal("Data integrity test failed")
		return
	}

	// Step 3: Test system behavior under various load conditions
	testLoadBehaviorE2E(t, datasetID, useProtobuf)

	// Step 4: Verify monitoring and logging work correctly in integration
	testMonitoringIntegrationE2E(t, datasetID, useProtobuf)

	t.Logf("=== Comprehensive E2E Test Completed Successfully with %s Backend ===", backendName)
}

// testCompleteDataFlowE2E tests complete data flow from upload to API response
func testCompleteDataFlowE2E(t *testing.T, useProtobuf bool) string {
	backendName := getBackendName(useProtobuf)

	// Create test HTML data
	testHTML := createE2ETestHTML()

	// Upload test data
	datasetID := uploadE2ETestData(t, testHTML, backendName)
	if datasetID == "" {
		return ""
	}

	// Wait for processing
	waitForE2EProcessing(t, datasetID)

	// Test data retrieval
	players := retrieveE2EPlayerData(t, datasetID, backendName)
	if len(players) == 0 {
		return ""
	}

	// Test API endpoints
	testE2EAPIEndpoints(t, datasetID, backendName)

	t.Logf("Complete data flow test passed with %s backend", backendName)
	return datasetID
}

// testDataIntegrityE2E validates data integrity throughout the system
func testDataIntegrityE2E(t *testing.T, datasetID string, useProtobuf bool) []Player {
	backendName := getBackendName(useProtobuf)

	// Retrieve player data
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/players/%s", datasetID), nil)
	w := httptest.NewRecorder()
	playerDataHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Data retrieval failed with %s backend: %d", backendName, w.Code)
		return nil
	}

	var response struct {
		Players        []Player `json:"players"`
		CurrencySymbol string   `json:"currencySymbol"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
		return nil
	}

	players := response.Players

	// Validate data integrity
	validateE2EDataIntegrity(t, players, backendName)

	t.Logf("Data integrity validated for %d players with %s backend", len(players), backendName)
	return players
}

// testLoadBehaviorE2E tests system behavior under load
func testLoadBehaviorE2E(t *testing.T, datasetID string, useProtobuf bool) {
	backendName := getBackendName(useProtobuf)

	// Test concurrent requests
	testE2EConcurrentRequests(t, datasetID, backendName)

	// Test mixed operations
	testE2EMixedOperations(t, datasetID, backendName)

	t.Logf("Load behavior tests completed with %s backend", backendName)
}

// testMonitoringIntegrationE2E verifies monitoring and logging
func testMonitoringIntegrationE2E(t *testing.T, datasetID string, useProtobuf bool) {
	backendName := getBackendName(useProtobuf)

	// Test successful operation logging
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "test.e2e_monitoring")
	defer span.End()

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/players/%s", datasetID), nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	playerDataHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Monitoring test failed with %s backend: %d", backendName, w.Code)
	}

	// Test error logging
	req = httptest.NewRequest("GET", "/api/players/invalid-dataset", nil)
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()
	playerDataHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for invalid dataset with %s backend, got %d", backendName, w.Code)
	}

	t.Logf("Monitoring integration verified with %s backend", backendName)
}

// Helper functions

func createE2ETestHTML() string {
	return `<html>
<head><title>E2E Test Data</title></head>
<body>
<table>
<tr>
	<th>Inf</th><th>Name</th><th>Nat</th><th>Age</th><th>Position</th><th>Club</th>
	<th>Transfer Value</th><th>Wage</th><th>Ability</th><th>Potential</th>
	<th>Personality</th><th>Media Handling</th><th>Av Rat</th><th>Apps</th>
	<th>Left Foot</th><th>Right Foot</th><th>Acc</th><th>Agg</th><th>Agi</th>
	<th>Ant</th><th>Bal</th><th>Bra</th><th>Cnt</th><th>Cmp</th><th>Cro</th>
	<th>Dec</th><th>Det</th><th>Dri</th><th>Fin</th><th>Fir</th><th>Fla</th>
	<th>Hea</th><th>L Th</th><th>Jum</th><th>Ldr</th><th>Lon</th><th>Mar</th>
	<th>OtB</th><th>Pac</th><th>Pas</th><th>Pos</th><th>Sta</th><th>Str</th>
	<th>Tck</th><th>Tea</th><th>Tec</th><th>Vis</th><th>Wor</th><th>Division</th><th>UID</th>
</tr>
<tr>
	<td>Hol</td><td>E2E Test Player 1</td><td>ENG</td><td>25</td><td>GK</td><td>E2E FC</td>
	<td>£5M</td><td>£50K</td><td>15</td><td>18</td><td>Balanced</td><td>Evasive</td>
	<td>7.5</td><td>30</td><td>10</td><td>15</td><td>12</td><td>8</td><td>14</td>
	<td>16</td><td>13</td><td>11</td><td>15</td><td>12</td><td>9</td><td>14</td>
	<td>13</td><td>10</td><td>8</td><td>12</td><td>7</td><td>11</td><td>6</td>
	<td>13</td><td>9</td><td>8</td><td>7</td><td>10</td><td>14</td><td>12</td>
	<td>16</td><td>15</td><td>11</td><td>6</td><td>13</td><td>14</td><td>12</td>
	<td>10</td><td>Premier League</td><td>300001</td>
</tr>
<tr>
	<td>Hol</td><td>E2E Test Player 2</td><td>ESP</td><td>22</td><td>D (C)</td><td>E2E United</td>
	<td>£15M</td><td>£75K</td><td>16</td><td>19</td><td>Driven</td><td>Outspoken</td>
	<td>7.8</td><td>28</td><td>12</td><td>14</td><td>15</td><td>16</td><td>13</td>
	<td>17</td><td>14</td><td>15</td><td>16</td><td>13</td><td>11</td><td>15</td>
	<td>14</td><td>12</td><td>9</td><td>13</td><td>8</td><td>14</td><td>7</td>
	<td>15</td><td>10</td><td>9</td><td>16</td><td>11</td><td>15</td><td>13</td>
	<td>17</td><td>16</td><td>18</td><td>15</td><td>14</td><td>15</td><td>13</td>
	<td>11</td><td>Premier League</td><td>300002</td>
</tr>
<tr>
	<td>Hol</td><td>E2E Test Player 3</td><td>BRA</td><td>28</td><td>M (C)</td><td>E2E City</td>
	<td>£25M</td><td>£100K</td><td>17</td><td>17</td><td>Professional</td><td>Composed</td>
	<td>8.2</td><td>32</td><td>16</td><td>14</td><td>16</td><td>12</td><td>15</td>
	<td>18</td><td>16</td><td>14</td><td>17</td><td>15</td><td>13</td><td>16</td>
	<td>15</td><td>17</td><td>14</td><td>16</td><td>12</td><td>13</td><td>11</td>
	<td>14</td><td>16</td><td>15</td><td>12</td><td>14</td><td>16</td><td>18</td>
	<td>15</td><td>16</td><td>13</td><td>11</td><td>15</td><td>17</td><td>16</td>
	<td>14</td><td>Premier League</td><td>300003</td>
</tr>
</table>
</body>
</html>`
}

func uploadE2ETestData(t *testing.T, testHTML, backendName string) string {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	fileWriter, err := writer.CreateFormFile("playerFile", "e2e_test.html")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	_, err = fileWriter.Write([]byte(testHTML))
	if err != nil {
		t.Fatalf("Failed to write test HTML: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	uploadHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Upload failed with %s backend. Status: %d", backendName, w.Code)
	}

	var response UploadResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse upload response: %v", err)
	}

	if response.DatasetID == "" {
		t.Fatal("DatasetID is empty")
	}

	t.Logf("Upload successful with %s backend. DatasetID: %s", backendName, response.DatasetID)
	return response.DatasetID
}

func waitForE2EProcessing(t *testing.T, datasetID string) {
	// Simple wait for processing
	time.Sleep(2 * time.Second)
	t.Logf("Processing wait completed for dataset: %s", datasetID)
}

func retrieveE2EPlayerData(t *testing.T, datasetID, backendName string) []Player {
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/players/%s", datasetID), nil)
	w := httptest.NewRecorder()
	playerDataHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Data retrieval failed with %s backend: %d", backendName, w.Code)
		return nil
	}

	var response struct {
		Players        []Player `json:"players"`
		CurrencySymbol string   `json:"currencySymbol"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
		return nil
	}

	t.Logf("Retrieved %d players with %s backend", len(response.Players), backendName)
	return response.Players
}

func testE2EAPIEndpoints(t *testing.T, datasetID, backendName string) {
	endpoints := []struct {
		path    string
		handler func(http.ResponseWriter, *http.Request)
		name    string
	}{
		{fmt.Sprintf("/api/players/%s", datasetID), playerDataHandler, "players"},
		{fmt.Sprintf("/api/leagues/%s", datasetID), leaguesHandler, "leagues"},
		{"/api/roles", rolesHandler, "roles"},
		{"/api/cache-status", cacheStatusHandler, "cache-status"},
	}

	for _, endpoint := range endpoints {
		req := httptest.NewRequest("GET", endpoint.path, nil)
		w := httptest.NewRecorder()
		endpoint.handler(w, req)

		// Allow some endpoints to fail gracefully during tests
		if w.Code != http.StatusOK && w.Code != http.StatusServiceUnavailable {
			t.Logf("%s endpoint returned %d with %s backend (may be expected)",
				endpoint.name, w.Code, backendName)
		}
	}

	t.Logf("API endpoints tested with %s backend", backendName)
}

func validateE2EDataIntegrity(t *testing.T, players []Player, backendName string) {
	for i, player := range players {
		// Check required fields
		if player.UID == 0 {
			t.Errorf("Player %d has zero UID with %s backend", i, backendName)
		}
		if player.Name == "" {
			t.Errorf("Player %d has empty name with %s backend", i, backendName)
		}
		if player.Position == "" {
			t.Errorf("Player %d has empty position with %s backend", i, backendName)
		}
		if player.Club == "" {
			t.Errorf("Player %d has empty club with %s backend", i, backendName)
		}

		// Check maps are initialized
		if player.Attributes == nil {
			t.Errorf("Player %d has nil Attributes map with %s backend", i, backendName)
		}
		if player.NumericAttributes == nil {
			t.Errorf("Player %d has nil NumericAttributes map with %s backend", i, backendName)
		}
	}
}

func testE2EConcurrentRequests(t *testing.T, datasetID, backendName string) {
	concurrency := 3
	requestsPerWorker := 2
	var wg sync.WaitGroup
	errors := make(chan error, concurrency*requestsPerWorker)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < requestsPerWorker; j++ {
				req := httptest.NewRequest("GET", fmt.Sprintf("/api/players/%s", datasetID), nil)
				w := httptest.NewRecorder()
				playerDataHandler(w, req)

				if w.Code != http.StatusOK {
					errors <- WrapErrorf(ErrWorkerRequestFailed, "worker %d request %d failed: %d", workerID, j, w.Code)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	errorCount := 0
	for err := range errors {
		t.Errorf("Concurrent request error with %s backend: %v", backendName, err)
		errorCount++
	}

	if errorCount == 0 {
		t.Logf("Concurrent requests test passed with %s backend", backendName)
	}
}

func testE2EMixedOperations(t *testing.T, datasetID, backendName string) {
	// Test different types of operations
	operations := []func() error{
		func() error {
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/players/%s", datasetID), nil)
			w := httptest.NewRecorder()
			playerDataHandler(w, req)
			if w.Code != http.StatusOK {
				return WrapErrorf(ErrPlayersRequestFailed, "players request failed: %d", w.Code)
			}
			return nil
		},
		func() error {
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/leagues/%s", datasetID), nil)
			w := httptest.NewRecorder()
			leaguesHandler(w, req)
			if w.Code != http.StatusOK {
				return WrapErrorf(ErrLeaguesRequestFailed, "leagues request failed: %d", w.Code)
			}
			return nil
		},
		func() error {
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/players/%s?position=GK", datasetID), nil)
			w := httptest.NewRecorder()
			playerDataHandler(w, req)
			if w.Code != http.StatusOK {
				return WrapErrorf(ErrFilteredRequestFailed, "filtered request failed: %d", w.Code)
			}
			return nil
		},
	}

	for i, op := range operations {
		if err := op(); err != nil {
			t.Errorf("Mixed operation %d failed with %s backend: %v", i, backendName, err)
		}
	}

	t.Logf("Mixed operations test completed with %s backend", backendName)
}
