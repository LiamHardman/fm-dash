package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	pb "api/proto"
)

// TestProtobufRequestResponseCycle tests the complete protobuf request/response cycle
func TestProtobufRequestResponseCycle(t *testing.T) {
	// Set up test data
	testDatasetID := "test-protobuf-cycle"
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
			Overall:            85,
			TransferValueAmount: 50000000,
			WageAmount:         100000,
			ShortPositions:     []string{"ST"},
			PositionGroups:     []string{"Forward"},
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
			Overall:            82,
			TransferValueAmount: 30000000,
			WageAmount:         80000,
			ShortPositions:     []string{"CM"},
			PositionGroups:     []string{"Midfielder"},
		},
	}
	testCurrency := "£"

	// Store test data
	SetPlayerData(testDatasetID, testPlayers, testCurrency)
	defer func() {
		DeleteDataset(testDatasetID)
	}()

	// Create a test server with our API handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/api/players/", playerDataHandler)
	mux.HandleFunc("/api/roles", rolesHandler)
	mux.HandleFunc("/api/leagues/", leaguesHandler)
	mux.HandleFunc("/api/teams/", teamsHandler)
	mux.HandleFunc("/api/search/", searchHandler)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Test cases for different endpoints and accept headers
	testCases := []struct {
		name           string
		endpoint       string
		acceptHeader   string
		expectedStatus int
		validateResponse func(t *testing.T, resp *http.Response)
	}{
		{
			name:           "Player Data with Protobuf Accept",
			endpoint:       fmt.Sprintf("/api/players/%s", testDatasetID),
			acceptHeader:   "application/x-protobuf",
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, resp *http.Response) {
				// Verify content type
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, "application/x-protobuf") {
					t.Errorf("Expected protobuf content type, got %s", contentType)
				}
				
				// Read and decode the protobuf response
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Failed to read response body: %v", err)
				}
				
				playerData := &pb.PlayerDataResponse{}
				if err := proto.Unmarshal(body, playerData); err != nil {
					t.Fatalf("Failed to unmarshal protobuf response: %v", err)
				}
				
				// Verify response data
				if len(playerData.Players) != 2 {
					t.Errorf("Expected 2 players, got %d", len(playerData.Players))
				}
				
				if playerData.CurrencySymbol != testCurrency {
					t.Errorf("Expected currency symbol %s, got %s", testCurrency, playerData.CurrencySymbol)
				}
				
				// Verify metadata
				if playerData.Metadata == nil {
					t.Error("Expected metadata in response")
				} else {
					if playerData.Metadata.TotalCount != 2 {
						t.Errorf("Expected total count 2, got %d", playerData.Metadata.TotalCount)
					}
				}
				
				// Verify first player data
				player1 := playerData.Players[0]
				if player1.Name != "Test Player 1" {
					t.Errorf("Expected player name 'Test Player 1', got %s", player1.Name)
				}
				if player1.Overall != 85 {
					t.Errorf("Expected overall 85, got %d", player1.Overall)
				}
			},
		},
		{
			name:           "Player Data with JSON Accept",
			endpoint:       fmt.Sprintf("/api/players/%s", testDatasetID),
			acceptHeader:   "application/json",
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, resp *http.Response) {
				// Verify content type
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, "application/json") {
					t.Errorf("Expected JSON content type, got %s", contentType)
				}
				
				// Read and decode the JSON response
				var data map[string]interface{}
				if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
					t.Fatalf("Failed to decode JSON response: %v", err)
				}
				
				// Verify response data
				players, ok := data["players"].([]interface{})
				if !ok || len(players) != 2 {
					t.Errorf("Expected 2 players, got %v", players)
				}
				
				if currency, ok := data["currency_symbol"].(string); !ok || currency != testCurrency {
					t.Errorf("Expected currency symbol %s, got %v", testCurrency, data["currency_symbol"])
				}
			},
		},
		{
			name:           "Roles with Protobuf Accept",
			endpoint:       "/api/roles",
			acceptHeader:   "application/x-protobuf",
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, resp *http.Response) {
				// Verify content type
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, "application/x-protobuf") {
					t.Errorf("Expected protobuf content type, got %s", contentType)
				}
				
				// Read and decode the protobuf response
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Failed to read response body: %v", err)
				}
				
				rolesData := &pb.RolesResponse{}
				if err := proto.Unmarshal(body, rolesData); err != nil {
					t.Fatalf("Failed to unmarshal protobuf response: %v", err)
				}
				
				// Verify response has roles
				if len(rolesData.Roles) == 0 {
					t.Error("Expected roles in response, got empty list")
				}
				
				// Verify metadata
				if rolesData.Metadata == nil {
					t.Error("Expected metadata in response")
				}
			},
		},
		{
			name:           "Leagues with Protobuf Accept",
			endpoint:       fmt.Sprintf("/api/leagues/%s", testDatasetID),
			acceptHeader:   "application/x-protobuf",
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, resp *http.Response) {
				// Verify content type
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, "application/x-protobuf") {
					t.Errorf("Expected protobuf content type, got %s", contentType)
				}
				
				// Read and decode the protobuf response
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Failed to read response body: %v", err)
				}
				
				leaguesData := &pb.LeaguesResponse{}
				if err := proto.Unmarshal(body, leaguesData); err != nil {
					t.Fatalf("Failed to unmarshal protobuf response: %v", err)
				}
				
				// Verify response has leagues
				if len(leaguesData.Leagues) == 0 {
					t.Error("Expected leagues in response, got empty list")
				}
				
				// Verify Premier League is included
				found := false
				for _, league := range leaguesData.Leagues {
					if league == "Premier League" {
						found = true
						break
					}
				}
				if !found {
					t.Error("Expected 'Premier League' in leagues response")
				}
				
				// Verify metadata
				if leaguesData.Metadata == nil {
					t.Error("Expected metadata in response")
				}
			},
		},
		{
			name:           "Teams with Protobuf Accept",
			endpoint:       fmt.Sprintf("/api/teams/%s/Premier League", testDatasetID),
			acceptHeader:   "application/x-protobuf",
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, resp *http.Response) {
				// Verify content type
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, "application/x-protobuf") {
					t.Errorf("Expected protobuf content type, got %s", contentType)
				}
				
				// Read and decode the protobuf response
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Failed to read response body: %v", err)
				}
				
				teamsData := &pb.TeamsResponse{}
				if err := proto.Unmarshal(body, teamsData); err != nil {
					t.Fatalf("Failed to unmarshal protobuf response: %v", err)
				}
				
				// Verify response has teams
				if len(teamsData.Teams) != 2 {
					t.Errorf("Expected 2 teams, got %d", len(teamsData.Teams))
				}
				
				// Verify teams are included
				foundFC := false
				foundUnited := false
				for _, team := range teamsData.Teams {
					if team == "Test FC" {
						foundFC = true
					}
					if team == "Test United" {
						foundUnited = true
					}
				}
				if !foundFC {
					t.Error("Expected 'Test FC' in teams response")
				}
				if !foundUnited {
					t.Error("Expected 'Test United' in teams response")
				}
				
				// Verify metadata
				if teamsData.Metadata == nil {
					t.Error("Expected metadata in response")
				}
			},
		},
		{
			name:           "Search with Protobuf Accept",
			endpoint:       fmt.Sprintf("/api/search/%s?q=Test", testDatasetID),
			acceptHeader:   "application/x-protobuf",
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, resp *http.Response) {
				// Verify content type
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, "application/x-protobuf") {
					t.Errorf("Expected protobuf content type, got %s", contentType)
				}
				
				// Read and decode the protobuf response
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Failed to read response body: %v", err)
				}
				
				searchData := &pb.SearchResponse{}
				if err := proto.Unmarshal(body, searchData); err != nil {
					t.Fatalf("Failed to unmarshal protobuf response: %v", err)
				}
				
				// Verify response has players
				if len(searchData.Players) == 0 {
					t.Error("Expected players in search response, got empty list")
				}
				
				// Verify query parameter
				if searchData.Query != "Test" {
					t.Errorf("Expected query 'Test', got %s", searchData.Query)
				}
				
				// Verify metadata
				if searchData.Metadata == nil {
					t.Error("Expected metadata in response")
				}
			},
		},
		{
			name:           "Player Data with Multiple Accept Types",
			endpoint:       fmt.Sprintf("/api/players/%s", testDatasetID),
			acceptHeader:   "application/x-protobuf, application/json;q=0.8",
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, resp *http.Response) {
				// Should prefer protobuf
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, "application/x-protobuf") {
					t.Errorf("Expected protobuf content type, got %s", contentType)
				}
			},
		},
		{
			name:           "Player Data with Quality Values",
			endpoint:       fmt.Sprintf("/api/players/%s", testDatasetID),
			acceptHeader:   "application/json;q=0.9, application/x-protobuf;q=0.8",
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, resp *http.Response) {
				// Should prefer JSON due to higher q value
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, "application/json") {
					t.Errorf("Expected JSON content type, got %s", contentType)
				}
			},
		},
		{
			name:           "Player Data with Wildcard Accept",
			endpoint:       fmt.Sprintf("/api/players/%s", testDatasetID),
			acceptHeader:   "*/*",
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, resp *http.Response) {
				// Should default to JSON
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, "application/json") {
					t.Errorf("Expected JSON content type, got %s", contentType)
				}
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a request to the test server
			url := server.URL + tc.endpoint
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			
			// Set the Accept header
			req.Header.Set("Accept", tc.acceptHeader)
			
			// Send the request
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()
			
			// Check the response status
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}
			
			// Validate the response
			tc.validateResponse(t, resp)
		})
	}
}

// TestProtobufContentNegotiation tests content negotiation with various Accept headers
func TestProtobufContentNegotiation(t *testing.T) {
	// Create a test handler that returns the negotiated content type
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		negotiator := NewContentNegotiator(r)
		serializer := negotiator.SelectSerializer()
		
		// Write response with the negotiated content type
		w.Header().Set("Content-Type", serializer.ContentType())
		w.Header().Set("X-Supports-Protobuf", fmt.Sprintf("%t", negotiator.SupportsProtobuf()))
		
		// Create a simple response based on the serializer type
		if serializer.ContentType() == "application/x-protobuf" {
			// Create a simple protobuf response
			response := &pb.GenericResponse{
				Data: "This is a protobuf response",
				Metadata: &pb.ResponseMetadata{
					Timestamp:  time.Now().Unix(),
					ApiVersion: "1.0",
					RequestId:  "test-request-id",
				},
			}
			
			data, err := serializer.Serialize(response)
			if err != nil {
				http.Error(w, "Serialization error", http.StatusInternalServerError)
				return
			}
			
			w.Write(data)
		} else {
			// Create a simple JSON response
			response := map[string]interface{}{
				"data": "This is a JSON response",
				"metadata": map[string]interface{}{
					"timestamp":   time.Now().Unix(),
					"api_version": "1.0",
					"request_id":  "test-request-id",
				},
			}
			
			data, err := serializer.Serialize(response)
			if err != nil {
				http.Error(w, "Serialization error", http.StatusInternalServerError)
				return
			}
			
			w.Write(data)
		}
	})
	
	// Create a test server
	server := httptest.NewServer(handler)
	defer server.Close()
	
	// Test cases for different Accept headers
	testCases := []struct {
		name                string
		acceptHeader        string
		expectedContentType string
		supportsProtobuf    bool
	}{
		{
			name:                "Protobuf Accept",
			acceptHeader:        "application/x-protobuf",
			expectedContentType: "application/x-protobuf",
			supportsProtobuf:    true,
		},
		{
			name:                "Alternative Protobuf Accept",
			acceptHeader:        "application/protobuf",
			expectedContentType: "application/x-protobuf",
			supportsProtobuf:    true,
		},
		{
			name:                "JSON Accept",
			acceptHeader:        "application/json",
			expectedContentType: "application/json",
			supportsProtobuf:    false,
		},
		{
			name:                "Multiple Types with Protobuf First",
			acceptHeader:        "application/x-protobuf, application/json",
			expectedContentType: "application/x-protobuf",
			supportsProtobuf:    true,
		},
		{
			name:                "Multiple Types with JSON First",
			acceptHeader:        "application/json, application/x-protobuf",
			expectedContentType: "application/json",
			supportsProtobuf:    true,
		},
		{
			name:                "Quality Values - Protobuf Higher",
			acceptHeader:        "application/json;q=0.8, application/x-protobuf;q=0.9",
			expectedContentType: "application/x-protobuf",
			supportsProtobuf:    true,
		},
		{
			name:                "Quality Values - JSON Higher",
			acceptHeader:        "application/x-protobuf;q=0.7, application/json;q=0.9",
			expectedContentType: "application/json",
			supportsProtobuf:    true,
		},
		{
			name:                "Wildcard Accept",
			acceptHeader:        "*/*",
			expectedContentType: "application/json", // Default fallback
			supportsProtobuf:    false,
		},
		{
			name:                "Empty Accept Header",
			acceptHeader:        "",
			expectedContentType: "application/json", // Default fallback
			supportsProtobuf:    false,
		},
		{
			name:                "Unsupported Type",
			acceptHeader:        "application/xml",
			expectedContentType: "application/json", // Default fallback
			supportsProtobuf:    false,
		},
	}
	
	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a request to the test server
			req, err := http.NewRequest("GET", server.URL, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			
			// Set the Accept header
			if tc.acceptHeader != "" {
				req.Header.Set("Accept", tc.acceptHeader)
			}
			
			// Send the request
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()
			
			// Check the content type
			contentType := resp.Header.Get("Content-Type")
			if contentType != tc.expectedContentType {
				t.Errorf("Expected content type %s, got %s", tc.expectedContentType, contentType)
			}
			
			// Check the supports protobuf header
			supportsProtobuf := resp.Header.Get("X-Supports-Protobuf")
			if supportsProtobuf != fmt.Sprintf("%t", tc.supportsProtobuf) {
				t.Errorf("Expected supports protobuf %t, got %s", tc.supportsProtobuf, supportsProtobuf)
			}
			
			// Verify the response can be parsed correctly
			if strings.Contains(contentType, "application/x-protobuf") {
				// Read and decode the protobuf response
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Failed to read response body: %v", err)
				}
				
				genericResponse := &pb.GenericResponse{}
				if err := proto.Unmarshal(body, genericResponse); err != nil {
					t.Fatalf("Failed to unmarshal protobuf response: %v", err)
				}
				
				// Verify response data
				if genericResponse.Data != "This is a protobuf response" {
					t.Errorf("Expected data 'This is a protobuf response', got %s", genericResponse.Data)
				}
			} else {
				// Read and decode the JSON response
				var data map[string]interface{}
				if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
					t.Fatalf("Failed to decode JSON response: %v", err)
				}
				
				// Verify response data
				if dataStr, ok := data["data"].(string); !ok || dataStr != "This is a JSON response" {
					t.Errorf("Expected data 'This is a JSON response', got %v", data["data"])
				}
			}
		})
	}
}

// TestProtobufFallbackMechanisms tests fallback mechanisms for protobuf errors
func TestProtobufFallbackMechanisms(t *testing.T) {
	// Create a test handler that simulates different error scenarios
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the scenario from the URL path
		pathParts := strings.Split(r.URL.Path, "/")
		scenario := pathParts[len(pathParts)-1]
		
		// Get the accept header to determine client preference
		acceptHeader := r.Header.Get("Accept")
		wantsProtobuf := strings.Contains(acceptHeader, "application/x-protobuf")
		
		switch scenario {
		case "serialization_error":
			// Simulate a protobuf serialization error with JSON fallback
			if wantsProtobuf {
				// Create an invalid protobuf message (will fail serialization)
				errorHandler := GetProtobufErrorHandler()
				
				// This is a valid protobuf message but we'll force an error
				playerData := &pb.PlayerDataResponse{
					Players: []*pb.Player{
						{
							Uid: 12345,
							Name: "Test Player",
						},
					},
					CurrencySymbol: "$",
					Metadata: CreateResponseMetadata("test-request", 1, false),
				}
				
				// Simulate serialization error and handle with fallback
				err := fmt.Errorf("simulated serialization error")
				errorHandler.HandleSerializationError(r.Context(), w, r, playerData, err, "test-dataset")
			} else {
				// Client wants JSON, just return JSON directly
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"message": "JSON response",
					"error": nil,
				})
			}
			
		case "conversion_error":
			// Simulate a protobuf conversion error with JSON fallback
			if wantsProtobuf {
				errorHandler := GetProtobufErrorHandler()
				
				// Create some data that would normally be converted to protobuf
				data := map[string]interface{}{
					"players": []map[string]interface{}{
						{
							"uid": 12345,
							"name": "Test Player",
						},
					},
					"currency_symbol": "$",
				}
				
				// Simulate conversion error and handle with fallback
				err := fmt.Errorf("simulated conversion error")
				errorHandler.HandleProtobufConversionError(
					r.Context(), w, r, data, err, "test-dataset", "to_protobuf")
			} else {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"message": "JSON response",
					"error": nil,
				})
			}
			
		case "compression_error":
			// Simulate a protobuf compression error with uncompressed fallback
			if wantsProtobuf {
				errorHandler := GetProtobufErrorHandler()
				
				// Create a valid protobuf message
				playerData := &pb.PlayerDataResponse{
					Players: []*pb.Player{
						{
							Uid: 12345,
							Name: "Test Player",
						},
					},
					CurrencySymbol: "$",
					Metadata: CreateResponseMetadata("test-request", 1, false),
				}
				
				// Simulate compression error and handle with fallback
				err := fmt.Errorf("simulated compression error")
				errorHandler.HandleProtobufCompressionError(
					r.Context(), w, r, playerData, err, "test-dataset", "compress")
			} else {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"message": "JSON response",
					"error": nil,
				})
			}
			
		case "validation_error":
			// Return a validation error response
			WriteErrorResponse(w, r, "validation_error", 
				"Invalid player data: validation failed", 
				[]string{"Player name is required", "Age must be a number"}, 
				http.StatusBadRequest)
			
		case "server_error":
			// Return a server error response
			WriteErrorResponse(w, r, "server_error", 
				"Internal server error occurred", 
				[]string{"Database connection failed"}, 
				http.StatusInternalServerError)
			
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Unknown scenario: %s", scenario)
		}
	})
	
	// Create a test server
	server := httptest.NewServer(handler)
	defer server.Close()
	
	// Test cases
	testCases := []struct {
		name           string
		scenario       string
		acceptHeader   string
		expectedStatus int
		checkResponse  func(t *testing.T, resp *http.Response)
	}{
		{
			name:           "Serialization Error with Protobuf Accept Header",
			scenario:       "serialization_error",
			acceptHeader:   "application/x-protobuf",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *http.Response) {
				// Should fall back to JSON
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, "application/json") {
					t.Errorf("Expected JSON content type, got %s", contentType)
				}
				
				// Should have fallback header
				fallbackHeader := resp.Header.Get("X-Serialization-Fallback")
				if fallbackHeader != string(FallbackReasonMarshalFailed) {
					t.Errorf("Expected fallback header %s, got %s", 
						FallbackReasonMarshalFailed, fallbackHeader)
				}
			},
		},
		{
			name:           "Conversion Error with Protobuf Accept Header",
			scenario:       "conversion_error",
			acceptHeader:   "application/x-protobuf",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *http.Response) {
				// Should fall back to JSON
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, "application/json") {
					t.Errorf("Expected JSON content type, got %s", contentType)
				}
				
				// Should have fallback header
				fallbackHeader := resp.Header.Get("X-Serialization-Fallback")
				if fallbackHeader != string(FallbackReasonConversionFailed) {
					t.Errorf("Expected fallback header %s, got %s", 
						FallbackReasonConversionFailed, fallbackHeader)
				}
			},
		},
		{
			name:           "Compression Error with Protobuf Accept Header",
			scenario:       "compression_error",
			acceptHeader:   "application/x-protobuf",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *http.Response) {
				// Should still return protobuf but uncompressed
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, "application/x-protobuf") {
					t.Errorf("Expected protobuf content type, got %s", contentType)
				}
				
				// Should have compression disabled header
				compressionHeader := resp.Header.Get("X-Compression-Status")
				if compressionHeader != "disabled" {
					t.Errorf("Expected compression disabled header, got %s", compressionHeader)
				}
			},
		},
		{
			name:           "Validation Error with Protobuf Accept Header",
			scenario:       "validation_error",
			acceptHeader:   "application/x-protobuf",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp *http.Response) {
				// Should return protobuf error
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, "application/x-protobuf") {
					t.Errorf("Expected protobuf content type, got %s", contentType)
				}
				
				// Read and decode the protobuf error response
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Failed to read response body: %v", err)
				}
				
				errorResponse := &pb.ErrorResponse{}
				if err := proto.Unmarshal(body, errorResponse); err != nil {
					t.Fatalf("Failed to unmarshal protobuf error response: %v", err)
				}
				
				// Verify error response fields
				if errorResponse.ErrorCode != "validation_error" {
					t.Errorf("Expected error code 'validation_error', got %s", errorResponse.ErrorCode)
				}
				
				if !strings.Contains(errorResponse.Message, "Invalid player data") {
					t.Errorf("Expected error message to contain 'Invalid player data', got %s", 
						errorResponse.Message)
				}
				
				if len(errorResponse.Details) != 2 {
					t.Errorf("Expected 2 error details, got %d", len(errorResponse.Details))
				}
			},
		},
		{
			name:           "Validation Error with JSON Accept Header",
			scenario:       "validation_error",
			acceptHeader:   "application/json",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp *http.Response) {
				// Should return JSON error
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, "application/json") {
					t.Errorf("Expected JSON content type, got %s", contentType)
				}
				
				// Read and decode the JSON error response
				var errorData map[string]interface{}
				if err := json.NewDecoder(resp.Body).Decode(&errorData); err != nil {
					t.Fatalf("Failed to decode JSON error response: %v", err)
				}
				
				// Verify error response fields
				if errorCode, ok := errorData["error_code"].(string); !ok || errorCode != "validation_error" {
					t.Errorf("Expected error code 'validation_error', got %v", errorData["error_code"])
				}
				
				if message, ok := errorData["message"].(string); !ok || !strings.Contains(message, "Invalid player data") {
					t.Errorf("Expected error message to contain 'Invalid player data', got %v", errorData["message"])
				}
				
				details, ok := errorData["details"].([]interface{})
				if !ok || len(details) != 2 {
					t.Errorf("Expected 2 error details, got %v", errorData["details"])
				}
			},
		},
		{
			name:           "Server Error with Protobuf Accept Header",
			scenario:       "server_error",
			acceptHeader:   "application/x-protobuf",
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, resp *http.Response) {
				// Should return protobuf error
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, "application/x-protobuf") {
					t.Errorf("Expected protobuf content type, got %s", contentType)
				}
				
				// Read and decode the protobuf error response
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Failed to read response body: %v", err)
				}
				
				errorResponse := &pb.ErrorResponse{}
				if err := proto.Unmarshal(body, errorResponse); err != nil {
					t.Fatalf("Failed to unmarshal protobuf error response: %v", err)
				}
				
				// Verify error response fields
				if errorResponse.ErrorCode != "server_error" {
					t.Errorf("Expected error code 'server_error', got %s", errorResponse.ErrorCode)
				}
			},
		},
	}
	
	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a request to the test server
			url := fmt.Sprintf("%s/%s", server.URL, tc.scenario)
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			
			// Set the Accept header
			req.Header.Set("Accept", tc.acceptHeader)
			
			// Send the request
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()
			
			// Check the response status
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}
			
			// Run the response check function
			tc.checkResponse(t, resp)
		})
	}
}

// TestConcurrentProtobufRequests tests handling of concurrent requests with mixed formats
func TestConcurrentProtobufRequests(t *testing.T) {
	// Set up test data
	testDatasetID := "test-concurrent-protobuf"
	testPlayers := make([]Player, 50)
	for i := 0; i < 50; i++ {
		testPlayers[i] = Player{
			UID:      int64(i + 1),
			Name:     fmt.Sprintf("Player %d", i+1),
			Position: "ST",
			Age:      "25",
			Overall:  80 + (i % 10),
		}
	}
	testCurrency := "£"

	// Store test data
	SetPlayerData(testDatasetID, testPlayers, testCurrency)
	defer func() {
		DeleteDataset(testDatasetID)
	}()

	// Create a test server with our API handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/api/players/", playerDataHandler)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Number of concurrent requests
	concurrency := 10
	
	// Create a wait group to synchronize goroutines
	var wg sync.WaitGroup
	wg.Add(concurrency * 2) // JSON and protobuf requests
	
	// Track errors
	errorCh := make(chan error, concurrency*2)
	
	// Track response formats
	jsonResponses := 0
	protobufResponses := 0
	var responseMutex sync.Mutex
	
	// Make concurrent protobuf requests
	for i := 0; i < concurrency; i++ {
		go func(i int) {
			defer wg.Done()
			
			// Create a request
			url := fmt.Sprintf("%s/api/players/%s", server.URL, testDatasetID)
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				errorCh <- fmt.Errorf("failed to create request: %v", err)
				return
			}
			
			// Set Accept header for protobuf
			req.Header.Set("Accept", "application/x-protobuf")
			
			// Send the request
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				errorCh <- fmt.Errorf("failed to send request: %v", err)
				return
			}
			defer resp.Body.Close()
			
			// Check status code
			if resp.StatusCode != http.StatusOK {
				errorCh <- fmt.Errorf("expected status 200, got %d", resp.StatusCode)
				return
			}
			
			// Check content type
			contentType := resp.Header.Get("Content-Type")
			if !strings.Contains(contentType, "application/x-protobuf") {
				errorCh <- fmt.Errorf("expected protobuf content type, got %s", contentType)
				return
			}
			
			// Read and decode the protobuf response
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				errorCh <- fmt.Errorf("failed to read response body: %v", err)
				return
			}
			
			playerData := &pb.PlayerDataResponse{}
			if err := proto.Unmarshal(body, playerData); err != nil {
				errorCh <- fmt.Errorf("failed to unmarshal protobuf response: %v", err)
				return
			}
			
			// Verify response data
			if len(playerData.Players) != len(testPlayers) {
				errorCh <- fmt.Errorf("expected %d players, got %d", len(testPlayers), len(playerData.Players))
				return
			}
			
			// Track response format
			responseMutex.Lock()
			protobufResponses++
			responseMutex.Unlock()
		}(i)
	}
	
	// Make concurrent JSON requests
	for i := 0; i < concurrency; i++ {
		go func(i int) {
			defer wg.Done()
			
			// Create a request
			url := fmt.Sprintf("%s/api/players/%s", server.URL, testDatasetID)
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				errorCh <- fmt.Errorf("failed to create request: %v", err)
				return
			}
			
			// Set Accept header for JSON
			req.Header.Set("Accept", "application/json")
			
			// Send the request
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				errorCh <- fmt.Errorf("failed to send request: %v", err)
				return
			}
			defer resp.Body.Close()
			
			// Check status code
			if resp.StatusCode != http.StatusOK {
				errorCh <- fmt.Errorf("expected status 200, got %d", resp.StatusCode)
				return
			}
			
			// Check content type
			contentType := resp.Header.Get("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				errorCh <- fmt.Errorf("expected JSON content type, got %s", contentType)
				return
			}
			
			// Read and decode the JSON response
			var data map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
				errorCh <- fmt.Errorf("failed to decode JSON response: %v", err)
				return
			}
			
			// Verify response data
			players, ok := data["players"].([]interface{})
			if !ok || len(players) != len(testPlayers) {
				errorCh <- fmt.Errorf("expected %d players, got %v", len(testPlayers), len(players))
				return
			}
			
			// Track response format
			responseMutex.Lock()
			jsonResponses++
			responseMutex.Unlock()
		}(i)
	}
	
	// Wait for all goroutines to finish
	wg.Wait()
	close(errorCh)
	
	// Check for errors
	var errors []error
	for err := range errorCh {
		errors = append(errors, err)
	}
	
	if len(errors) > 0 {
		for _, err := range errors {
			t.Errorf("Concurrent test error: %v", err)
		}
	}
	
	// Verify response counts
	if protobufResponses != concurrency {
		t.Errorf("Expected %d protobuf responses, got %d", concurrency, protobufResponses)
	}
	
	if jsonResponses != concurrency {
		t.Errorf("Expected %d JSON responses, got %d", concurrency, jsonResponses)
	}
}

// TestProtobufErrorHandlingWithRetries tests error handling with retries
func TestProtobufErrorHandlingWithRetries(t *testing.T) {
	// Create a test handler that simulates transient errors
	var requestCount int
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the scenario from the URL path
		pathParts := strings.Split(r.URL.Path, "/")
		scenario := pathParts[len(pathParts)-1]
		
		// Get the accept header to determine client preference
		acceptHeader := r.Header.Get("Accept")
		wantsProtobuf := strings.Contains(acceptHeader, "application/x-protobuf")
		
		// Increment request count
		requestCount++
		
		switch scenario {
		case "transient_error":
			// Simulate a transient error that succeeds on retry
			if requestCount <= 1 {
				// First request fails
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte("Service temporarily unavailable"))
				return
			}
			
			// Subsequent requests succeed
			if wantsProtobuf {
				// Return protobuf response
				response := &pb.PlayerDataResponse{
					Players: []*pb.Player{
						{
							Uid: 12345,
							Name: "Test Player",
						},
					},
					CurrencySymbol: "$",
					Metadata: CreateResponseMetadata("test-request", 1, false),
				}
				
				data, err := proto.Marshal(response)
				if err != nil {
					http.Error(w, "Serialization error", http.StatusInternalServerError)
					return
				}
				
				w.Header().Set("Content-Type", "application/x-protobuf")
				w.Write(data)
			} else {
				// Return JSON response
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"players": []map[string]interface{}{
						{
							"uid": 12345,
							"name": "Test Player",
						},
					},
					"currency_symbol": "$",
				})
			}
			
		case "partial_response":
			// Simulate a partial response that requires retry
			if requestCount <= 1 {
				// First request returns partial data
				w.Header().Set("Content-Type", "application/x-protobuf")
				w.Write([]byte{0x0A, 0x03}) // Incomplete protobuf message
				return
			}
			
			// Subsequent requests succeed
			if wantsProtobuf {
				// Return complete protobuf response
				response := &pb.PlayerDataResponse{
					Players: []*pb.Player{
						{
							Uid: 12345,
							Name: "Test Player",
						},
					},
					CurrencySymbol: "$",
					Metadata: CreateResponseMetadata("test-request", 1, false),
				}
				
				data, err := proto.Marshal(response)
				if err != nil {
					http.Error(w, "Serialization error", http.StatusInternalServerError)
					return
				}
				
				w.Header().Set("Content-Type", "application/x-protobuf")
				w.Write(data)
			} else {
				// Return JSON response
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"players": []map[string]interface{}{
						{
							"uid": 12345,
							"name": "Test Player",
						},
					},
					"currency_symbol": "$",
				})
			}
			
		case "permanent_error":
			// Simulate a permanent error that should not be retried
			w.WriteHeader(http.StatusBadRequest)
			
			if wantsProtobuf {
				// Return protobuf error
				errorResponse := &pb.ErrorResponse{
					ErrorCode: "validation_error",
					Message:   "Invalid request parameters",
					Details:   []string{"Parameter 'id' is required"},
					Metadata:  CreateResponseMetadata("test-request", 0, false),
				}
				
				data, err := proto.Marshal(errorResponse)
				if err != nil {
					http.Error(w, "Serialization error", http.StatusInternalServerError)
					return
				}
				
				w.Header().Set("Content-Type", "application/x-protobuf")
				w.Write(data)
			} else {
				// Return JSON error
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error_code": "validation_error",
					"message":    "Invalid request parameters",
					"details":    []string{"Parameter 'id' is required"},
				})
			}
			
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Unknown scenario: %s", scenario)
		}
	})
	
	// Create a test server
	server := httptest.NewServer(handler)
	defer server.Close()
	
	// Test cases
	testCases := []struct {
		name           string
		scenario       string
		acceptHeader   string
		expectedStatus int
		expectedRetries int
	}{
		{
			name:           "Transient Error with Protobuf Accept",
			scenario:       "transient_error",
			acceptHeader:   "application/x-protobuf",
			expectedStatus: http.StatusOK,
			expectedRetries: 1,
		},
		{
			name:           "Partial Response with Protobuf Accept",
			scenario:       "partial_response",
			acceptHeader:   "application/x-protobuf",
			expectedStatus: http.StatusOK,
			expectedRetries: 1,
		},
		{
			name:           "Permanent Error with Protobuf Accept",
			scenario:       "permanent_error",
			acceptHeader:   "application/x-protobuf",
			expectedStatus: http.StatusBadRequest,
			expectedRetries: 0,
		},
	}
	
	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset request count
			requestCount = 0
			
			// Create a request to the test server
			url := fmt.Sprintf("%s/%s", server.URL, tc.scenario)
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			
			// Set the Accept header
			req.Header.Set("Accept", tc.acceptHeader)
			
			// Create a client with retry capability
			client := &http.Client{
				Timeout: 5 * time.Second,
			}
			
			// Send the request
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()
			
			// Check the response status
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}
			
			// Check the number of requests made (original + retries)
			expectedRequests := tc.expectedRetries + 1
			if requestCount != expectedRequests {
				t.Errorf("Expected %d requests, got %d", expectedRequests, requestCount)
			}
		})
	}
}