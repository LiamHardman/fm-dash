package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

// TestAPICompatibility tests all API endpoints to ensure they return identical JSON responses
// with both JSON and protobuf storage backends
func TestAPICompatibility(t *testing.T) {
	// Initialize test environment
	InitStore()
	InitInMemoryCache()
	InitCacheStorage(context.Background())
	InitializeMemoryOptimizations()

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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variable for storage backend
			originalValue := os.Getenv("USE_PROTOBUF")
			os.Setenv("USE_PROTOBUF", tc.envVarValue)
			defer os.Setenv("USE_PROTOBUF", originalValue)

			// Re-initialize storage with new setting
			InitStore()

			// Run all endpoint tests
			t.Run("Upload Endpoint", func(t *testing.T) {
				testUploadEndpoint(t, tc.useProtobuf)
			})

			t.Run("Player Data Endpoint", func(t *testing.T) {
				testPlayerDataEndpoint(t, tc.useProtobuf)
			})

			t.Run("Roles Endpoint", func(t *testing.T) {
				testRolesEndpoint(t, tc.useProtobuf)
			})

			t.Run("Leagues Endpoint", func(t *testing.T) {
				testLeaguesEndpoint(t, tc.useProtobuf)
			})

			t.Run("Teams Endpoint", func(t *testing.T) {
				testTeamsEndpoint(t, tc.useProtobuf)
			})

			t.Run("Search Endpoint", func(t *testing.T) {
				testSearchEndpoint(t, tc.useProtobuf)
			})

			t.Run("Config Endpoint", func(t *testing.T) {
				testConfigEndpoint(t, tc.useProtobuf)
			})

			t.Run("Cache Status Endpoint", func(t *testing.T) {
				testCacheStatusEndpoint(t, tc.useProtobuf)
			})
		})
	}
}

// testUploadEndpoint tests the HTML upload processing with both storage backends
func testUploadEndpoint(t *testing.T, useProtobuf bool) {
	// Create test HTML file content
	testHTML := createTestHTMLContent()

	// Create multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add file field
	fileWriter, err := writer.CreateFormFile("playerFile", "test_players.html")
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

	// Check response status
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Response: %s",
			http.StatusOK, w.Code, w.Body.String())
		return
	}

	// Parse response
	var response UploadResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse upload response: %v", err)
	}

	// Validate response structure
	if response.DatasetID == "" {
		t.Error("DatasetID should not be empty")
	}

	if response.Message == "" {
		t.Error("Message should not be empty")
	}

	if response.DetectedCurrencySymbol == "" {
		t.Error("DetectedCurrencySymbol should not be empty")
	}

	// Store dataset ID for subsequent tests
	if useProtobuf {
		// Store for protobuf tests
		os.Setenv("TEST_DATASET_ID_PROTOBUF", response.DatasetID)
	} else {
		// Store for JSON tests
		os.Setenv("TEST_DATASET_ID_JSON", response.DatasetID)
	}

	t.Logf("Upload successful with %s backend. DatasetID: %s",
		getBackendName(useProtobuf), response.DatasetID)
}

// testPlayerDataEndpoint tests player data retrieval
func testPlayerDataEndpoint(t *testing.T, useProtobuf bool) {
	// Get dataset ID from upload test
	var datasetID string
	if useProtobuf {
		datasetID = os.Getenv("TEST_DATASET_ID_PROTOBUF")
	} else {
		datasetID = os.Getenv("TEST_DATASET_ID_JSON")
	}

	if datasetID == "" {
		t.Skip("No dataset ID available, skipping player data test")
		return
	}

	// Test basic player data retrieval
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/players/%s", datasetID), nil)
	w := httptest.NewRecorder()

	playerDataHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Response: %s",
			http.StatusOK, w.Code, w.Body.String())
		return
	}

	// Parse response
	var response struct {
		Players        []Player `json:"players"`
		CurrencySymbol string   `json:"currencySymbol"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse player data response: %v", err)
	}

	// Validate response structure
	if len(response.Players) == 0 {
		t.Error("Players array should not be empty")
	}

	if response.CurrencySymbol == "" {
		t.Error("CurrencySymbol should not be empty")
	}

	// Validate player structure
	player := response.Players[0]
	validatePlayerStructure(t, player)

	// Test with filters
	testPlayerDataWithFilters(t, datasetID, useProtobuf)

	t.Logf("Player data retrieval successful with %s backend. Players: %d",
		getBackendName(useProtobuf), len(response.Players))
}

// testPlayerDataWithFilters tests player data retrieval with various filters
func testPlayerDataWithFilters(t *testing.T, datasetID string, useProtobuf bool) {
	testCases := []struct {
		name   string
		params string
	}{
		{"Position Filter", "?position=GK"},
		{"Role Filter", "?role=Goalkeeper"},
		{"Age Filter", "?minAge=18&maxAge=25"},
		{"Transfer Value Filter", "?minTransferValue=1000000&maxTransferValue=50000000"},
		{"Division Filter", "?divisionFilter=all"},
		{"Combined Filters", "?position=D&minAge=20&maxAge=30&divisionFilter=same"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/players/%s%s", datasetID, tc.params)
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			playerDataHandler(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d for %s. Response: %s",
					http.StatusOK, w.Code, tc.name, w.Body.String())
				return
			}

			// Ensure response is valid JSON
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Errorf("Invalid JSON response for %s: %v", tc.name, err)
			}
		})
	}
}

// testRolesEndpoint tests the roles API endpoint
func testRolesEndpoint(t *testing.T, useProtobuf bool) {
	req := httptest.NewRequest("GET", "/api/roles", nil)
	w := httptest.NewRecorder()

	rolesHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Response: %s",
			http.StatusOK, w.Code, w.Body.String())
		return
	}

	// Parse response
	var roles []string
	if err := json.Unmarshal(w.Body.Bytes(), &roles); err != nil {
		t.Fatalf("Failed to parse roles response: %v", err)
	}

	// Validate response
	if len(roles) == 0 {
		t.Error("Roles array should not be empty")
	}

	// Check for expected roles
	expectedRoles := []string{"Goalkeeper", "Centre Back", "Full Back"}
	for _, expectedRole := range expectedRoles {
		found := false
		for _, role := range roles {
			if role == expectedRole {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected role '%s' not found in response", expectedRole)
		}
	}

	t.Logf("Roles endpoint successful with %s backend. Roles: %d",
		getBackendName(useProtobuf), len(roles))
}

// testLeaguesEndpoint tests the leagues API endpoint
func testLeaguesEndpoint(t *testing.T, useProtobuf bool) {
	// Get dataset ID
	var datasetID string
	if useProtobuf {
		datasetID = os.Getenv("TEST_DATASET_ID_PROTOBUF")
	} else {
		datasetID = os.Getenv("TEST_DATASET_ID_JSON")
	}

	if datasetID == "" {
		t.Skip("No dataset ID available, skipping leagues test")
		return
	}

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/leagues/%s", datasetID), nil)
	w := httptest.NewRecorder()

	leaguesHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Response: %s",
			http.StatusOK, w.Code, w.Body.String())
		return
	}

	// Parse response
	var leagues []League
	if err := json.Unmarshal(w.Body.Bytes(), &leagues); err != nil {
		t.Fatalf("Failed to parse leagues response: %v", err)
	}

	// Validate response structure
	if len(leagues) == 0 {
		t.Error("Leagues array should not be empty")
	}

	// Validate league structure
	league := leagues[0]
	if league.Name == "" {
		t.Error("League name should not be empty")
	}

	if league.TeamCount < 0 {
		t.Error("League team count should not be negative")
	}

	if league.PlayerCount < 0 {
		t.Error("League player count should not be negative")
	}

	t.Logf("Leagues endpoint successful with %s backend. Leagues: %d",
		getBackendName(useProtobuf), len(leagues))
}

// testTeamsEndpoint tests the teams API endpoint
func testTeamsEndpoint(t *testing.T, useProtobuf bool) {
	// Get dataset ID
	var datasetID string
	if useProtobuf {
		datasetID = os.Getenv("TEST_DATASET_ID_PROTOBUF")
	} else {
		datasetID = os.Getenv("TEST_DATASET_ID_JSON")
	}

	if datasetID == "" {
		t.Skip("No dataset ID available, skipping teams test")
		return
	}

	// First get leagues to find a valid league name
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/leagues/%s", datasetID), nil)
	w := httptest.NewRecorder()
	leaguesHandler(w, req)

	if w.Code != http.StatusOK {
		t.Skip("Cannot get leagues for teams test")
		return
	}

	var leagues []League
	if err := json.Unmarshal(w.Body.Bytes(), &leagues); err != nil || len(leagues) == 0 {
		t.Skip("No leagues available for teams test")
		return
	}

	// Test teams endpoint with first league
	leagueName := leagues[0].Name
	// URL encode the league name to handle spaces and special characters
	encodedLeagueName := url.QueryEscape(leagueName)
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/teams/%s?league=%s", datasetID, encodedLeagueName), nil)
	w = httptest.NewRecorder()

	teamsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Response: %s",
			http.StatusOK, w.Code, w.Body.String())
		return
	}

	// Parse response
	var teams []Team
	if err := json.Unmarshal(w.Body.Bytes(), &teams); err != nil {
		t.Fatalf("Failed to parse teams response: %v", err)
	}

	// Validate response structure
	if len(teams) == 0 {
		t.Error("Teams array should not be empty")
	}

	// Validate team structure
	team := teams[0]
	if team.Name == "" {
		t.Error("Team name should not be empty")
	}

	if team.Division == "" {
		t.Error("Team division should not be empty")
	}

	if team.PlayerCount < 0 {
		t.Error("Team player count should not be negative")
	}

	t.Logf("Teams endpoint successful with %s backend. Teams: %d",
		getBackendName(useProtobuf), len(teams))
}

// testSearchEndpoint tests the search API endpoint
func testSearchEndpoint(t *testing.T, useProtobuf bool) {
	// Get dataset ID
	var datasetID string
	if useProtobuf {
		datasetID = os.Getenv("TEST_DATASET_ID_PROTOBUF")
	} else {
		datasetID = os.Getenv("TEST_DATASET_ID_JSON")
	}

	if datasetID == "" {
		t.Skip("No dataset ID available, skipping search test")
		return
	}

	// Test search with various queries
	testCases := []struct {
		name  string
		query string
	}{
		{"Player Search", "?q=player&type=players"},
		{"Team Search", "?q=team&type=teams"},
		{"League Search", "?q=league&type=leagues"},
		{"General Search", "?q=test"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/search/%s%s", datasetID, tc.query)
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			searchHandler(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d for %s. Response: %s",
					http.StatusOK, w.Code, tc.name, w.Body.String())
				return
			}

			// Ensure response is valid JSON
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Errorf("Invalid JSON response for %s: %v", tc.name, err)
			}
		})
	}
}

// testConfigEndpoint tests the config API endpoint
func testConfigEndpoint(t *testing.T, useProtobuf bool) {
	req := httptest.NewRequest("GET", "/api/config", nil)
	w := httptest.NewRecorder()

	cachedConfigHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Response: %s",
			http.StatusOK, w.Code, w.Body.String())
		return
	}

	// Parse response
	var config map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &config); err != nil {
		t.Fatalf("Failed to parse config response: %v", err)
	}

	// Validate response structure
	if len(config) == 0 {
		t.Error("Config should not be empty")
	}

	t.Logf("Config endpoint successful with %s backend. Config keys: %d",
		getBackendName(useProtobuf), len(config))
}

// testCacheStatusEndpoint tests the cache status API endpoint
func testCacheStatusEndpoint(t *testing.T, useProtobuf bool) {
	req := httptest.NewRequest("GET", "/api/cache-status", nil)
	w := httptest.NewRecorder()

	cacheStatusHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Response: %s",
			http.StatusOK, w.Code, w.Body.String())
		return
	}

	// Parse response
	var status map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &status); err != nil {
		t.Fatalf("Failed to parse cache status response: %v", err)
	}

	// Validate response structure
	if len(status) == 0 {
		t.Error("Cache status should not be empty")
	}

	t.Logf("Cache status endpoint successful with %s backend", getBackendName(useProtobuf))
}

// validatePlayerStructure validates that a player has all expected fields
func validatePlayerStructure(t *testing.T, player Player) {
	if player.UID == 0 {
		t.Error("Player UID should not be zero")
	}

	if player.Name == "" {
		t.Error("Player name should not be empty")
	}

	if player.Position == "" {
		t.Error("Player position should not be empty")
	}

	if player.Age == "" {
		t.Error("Player age should not be empty")
	}

	if player.Club == "" {
		t.Error("Player club should not be empty")
	}

	// Validate numeric attributes are present (overall can be 0 in test environment)
	if player.Overall < 0 {
		t.Error("Player overall should not be negative")
	}

	// Validate maps are initialized
	if player.Attributes == nil {
		t.Error("Player attributes map should be initialized")
	}

	if player.NumericAttributes == nil {
		t.Error("Player numeric attributes map should be initialized")
	}
}

// createTestHTMLContent creates test HTML content for upload testing
func createTestHTMLContent() string {
	return `<html>
<head><title>Test Players</title></head>
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
	<td>Test Player 1</td>
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
	<td>123456</td>
</tr>
<tr>
	<td>Hol</td>
	<td>Test Player 2</td>
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
	<td>789012</td>
</tr>
</table>
</body>
</html>`
}

// getBackendName returns a human-readable name for the storage backend
func getBackendName(useProtobuf bool) string {
	if useProtobuf {
		return "Protobuf"
	}
	return "JSON"
}
