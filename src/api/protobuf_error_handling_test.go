package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	pb "api/proto"
)

// TestProtobufErrorHandling tests the error handling and fallback mechanisms for protobuf operations
func TestProtobufErrorHandling(t *testing.T) {
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
			
		case "client_compatibility":
			// Simulate a client compatibility issue
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
				
				// Handle client compatibility issue
				errorHandler.HandleClientCompatibilityError(
					r.Context(), w, r, playerData, "IE11", "test-dataset")
			} else {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"message": "JSON response",
					"error": nil,
				})
			}
			
		case "error_response":
			// Return a proper protobuf error response
			if wantsProtobuf {
				errorHandler := GetProtobufErrorHandler()
				
				// Create an error response
				errorResponse := errorHandler.CreateErrorResponse(
					r.Context(),
					"test_error",
					"This is a test error message",
					[]string{"Detail 1", "Detail 2"},
					"test-request-id",
				)
				
				// Serialize and send the error response
				data, err := proto.Marshal(errorResponse)
				if err != nil {
					t.Fatalf("Failed to marshal error response: %v", err)
				}
				
				w.Header().Set("Content-Type", "application/x-protobuf")
				w.WriteHeader(http.StatusBadRequest)
				w.Write(data)
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "test_error",
					"message": "This is a test error message",
					"details": []string{"Detail 1", "Detail 2"},
				})
			}
			
		case "success":
			// Return a successful response in the requested format
			if wantsProtobuf {
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
				
				// Serialize and send the response
				data, err := proto.Marshal(playerData)
				if err != nil {
					t.Fatalf("Failed to marshal response: %v", err)
				}
				
				w.Header().Set("Content-Type", "application/x-protobuf")
				w.Write(data)
			} else {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"players": []map[string]interface{}{
						{
							"uid": 12345,
							"name": "Test Player",
						},
					},
					"currency_symbol": "$",
					"metadata": map[string]interface{}{
						"timestamp": 123456789,
						"api_version": "1.0",
						"from_cache": false,
						"request_id": "test-request",
						"total_count": 1,
					},
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
			name:           "Client Compatibility Issue with Protobuf Accept Header",
			scenario:       "client_compatibility",
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
				if fallbackHeader != "client_compatibility" {
					t.Errorf("Expected client_compatibility fallback header, got %s", fallbackHeader)
				}
			},
		},
		{
			name:           "Error Response with Protobuf Accept Header",
			scenario:       "error_response",
			acceptHeader:   "application/x-protobuf",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp *http.Response) {
				// Should return protobuf error
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, "application/x-protobuf") {
					t.Errorf("Expected protobuf content type, got %s", contentType)
				}
				
				// Read and decode the protobuf error response
				buf := new(bytes.Buffer)
				buf.ReadFrom(resp.Body)
				
				errorResponse := &pb.ErrorResponse{}
				err := proto.Unmarshal(buf.Bytes(), errorResponse)
				if err != nil {
					t.Fatalf("Failed to unmarshal error response: %v", err)
				}
				
				// Verify error response fields
				if errorResponse.ErrorCode != "test_error" {
					t.Errorf("Expected error code 'test_error', got %s", errorResponse.ErrorCode)
				}
				
				if errorResponse.Message != "This is a test error message" {
					t.Errorf("Expected error message 'This is a test error message', got %s", 
						errorResponse.Message)
				}
				
				if len(errorResponse.Details) != 2 || 
				   errorResponse.Details[0] != "Detail 1" || 
				   errorResponse.Details[1] != "Detail 2" {
					t.Errorf("Error details don't match expected values: %v", errorResponse.Details)
				}
			},
		},
		{
			name:           "Success with Protobuf Accept Header",
			scenario:       "success",
			acceptHeader:   "application/x-protobuf",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *http.Response) {
				// Should return protobuf
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, "application/x-protobuf") {
					t.Errorf("Expected protobuf content type, got %s", contentType)
				}
				
				// Read and decode the protobuf response
				buf := new(bytes.Buffer)
				buf.ReadFrom(resp.Body)
				
				playerData := &pb.PlayerDataResponse{}
				err := proto.Unmarshal(buf.Bytes(), playerData)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				
				// Verify response fields
				if len(playerData.Players) != 1 {
					t.Errorf("Expected 1 player, got %d", len(playerData.Players))
				}
				
				if playerData.Players[0].Uid != 12345 {
					t.Errorf("Expected player UID 12345, got %d", playerData.Players[0].Uid)
				}
				
				if playerData.Players[0].Name != "Test Player" {
					t.Errorf("Expected player name 'Test Player', got %s", playerData.Players[0].Name)
				}
				
				if playerData.CurrencySymbol != "$" {
					t.Errorf("Expected currency symbol '$', got %s", playerData.CurrencySymbol)
				}
			},
		},
		{
			name:           "Success with JSON Accept Header",
			scenario:       "success",
			acceptHeader:   "application/json",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *http.Response) {
				// Should return JSON
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, "application/json") {
					t.Errorf("Expected JSON content type, got %s", contentType)
				}
				
				// Read and decode the JSON response
				var data map[string]interface{}
				if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
					t.Fatalf("Failed to decode JSON response: %v", err)
				}
				
				// Verify response fields
				players, ok := data["players"].([]interface{})
				if !ok || len(players) != 1 {
					t.Errorf("Expected 1 player, got %v", players)
				}
				
				player, ok := players[0].(map[string]interface{})
				if !ok {
					t.Errorf("Expected player to be a map, got %T", players[0])
				}
				
				if uid, ok := player["uid"].(float64); !ok || uid != 12345 {
					t.Errorf("Expected player UID 12345, got %v", player["uid"])
				}
				
				if name, ok := player["name"].(string); !ok || name != "Test Player" {
					t.Errorf("Expected player name 'Test Player', got %v", player["name"])
				}
				
				if currency, ok := data["currency_symbol"].(string); !ok || currency != "$" {
					t.Errorf("Expected currency symbol '$', got %v", data["currency_symbol"])
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

// TestProtobufErrorHandlerMetrics tests the error metrics collection
func TestProtobufErrorHandlerMetrics(t *testing.T) {
	// Create a new error handler
	handler := NewProtobufErrorHandler()
	
	// Create a test request and response writer
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	ctx := context.Background()
	
	// Simulate various errors
	handler.HandleSerializationError(ctx, w, req, nil, fmt.Errorf("test error"), "dataset1")
	handler.HandleProtobufConversionError(ctx, w, req, nil, fmt.Errorf("test error"), "dataset2", "to_protobuf")
	handler.HandleProtobufCompressionError(ctx, w, req, nil, fmt.Errorf("test error"), "dataset3", "compress")
	handler.HandleClientCompatibilityError(ctx, w, req, nil, "IE11", "dataset4")
	
	// Get the metrics
	metrics := handler.GetErrorMetrics()
	
	// Check the fallback count
	fallbackCount, ok := metrics["fallback_count"].(int64)
	if !ok || fallbackCount != 4 {
		t.Errorf("Expected fallback count 4, got %v", metrics["fallback_count"])
	}
	
	// Check error types
	errorsByType, ok := metrics["errors_by_type"].(map[string]int64)
	if !ok {
		t.Errorf("Expected errors_by_type to be a map, got %T", metrics["errors_by_type"])
	}
	
	// Check specific error types
	if errorsByType["serialization"] != 1 {
		t.Errorf("Expected 1 serialization error, got %d", errorsByType["serialization"])
	}
	
	if errorsByType["conversion_to_protobuf"] != 1 {
		t.Errorf("Expected 1 conversion error, got %d", errorsByType["conversion_to_protobuf"])
	}
	
	if errorsByType["compression_compress"] != 1 {
		t.Errorf("Expected 1 compression error, got %d", errorsByType["compression_compress"])
	}
	
	if errorsByType["client_compatibility"] != 1 {
		t.Errorf("Expected 1 client compatibility error, got %d", errorsByType["client_compatibility"])
	}
}