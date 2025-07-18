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

// TestEndToEndProtobufErrorHandling tests the complete flow of protobuf error handling
// from client request through server processing and back to client response
func TestEndToEndProtobufErrorHandling(t *testing.T) {
	// Create a test server with our API handlers
	mux := http.NewServeMux()
	
	// Register handlers for different test scenarios
	mux.HandleFunc("/api/players/normal", playerDataHandler)
	mux.HandleFunc("/api/players/error/serialization", serializationErrorHandler)
	mux.HandleFunc("/api/players/error/conversion", conversionErrorHandler)
	mux.HandleFunc("/api/players/error/validation", validationErrorHandler)
	mux.HandleFunc("/api/players/error/server", serverErrorHandler)
	
	server := httptest.NewServer(mux)
	defer server.Close()
	
	// Test cases
	testCases := []struct {
		name           string
		endpoint       string
		acceptHeader   string
		expectedStatus int
		validateResponse func(t *testing.T, resp *http.Response)
	}{
		{
			name:           "Normal Request with Protobuf Accept",
			endpoint:       "/api/players/normal",
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
				if len(playerData.Players) != 1 {
					t.Errorf("Expected 1 player, got %d", len(playerData.Players))
				}
				
				if playerData.Players[0].Name != "Test Player" {
					t.Errorf("Expected player name 'Test Player', got %s", playerData.Players[0].Name)
				}
			},
		},
		{
			name:           "Normal Request with JSON Accept",
			endpoint:       "/api/players/normal",
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
				if !ok || len(players) != 1 {
					t.Errorf("Expected 1 player, got %v", players)
				}
			},
		},
		{
			name:           "Serialization Error with Protobuf Accept",
			endpoint:       "/api/players/error/serialization",
			acceptHeader:   "application/x-protobuf",
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, resp *http.Response) {
				// Verify content type (should fall back to JSON)
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, "application/json") {
					t.Errorf("Expected JSON content type, got %s", contentType)
				}
				
				// Verify fallback header
				fallbackHeader := resp.Header.Get("X-Serialization-Fallback")
				if fallbackHeader != string(FallbackReasonMarshalFailed) {
					t.Errorf("Expected fallback header %s, got %s", 
						FallbackReasonMarshalFailed, fallbackHeader)
				}
				
				// Read and decode the JSON response
				var data map[string]interface{}
				if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
					t.Fatalf("Failed to decode JSON response: %v", err)
				}
				
				// Verify response data
				players, ok := data["players"].([]interface{})
				if !ok || len(players) != 1 {
					t.Errorf("Expected 1 player, got %v", players)
				}
			},
		},
		{
			name:           "Conversion Error with Protobuf Accept",
			endpoint:       "/api/players/error/conversion",
			acceptHeader:   "application/x-protobuf",
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, resp *http.Response) {
				// Verify content type (should fall back to JSON)
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, "application/json") {
					t.Errorf("Expected JSON content type, got %s", contentType)
				}
				
				// Verify fallback header
				fallbackHeader := resp.Header.Get("X-Serialization-Fallback")
				if fallbackHeader != string(FallbackReasonConversionFailed) {
					t.Errorf("Expected fallback header %s, got %s", 
						FallbackReasonConversionFailed, fallbackHeader)
				}
			},
		},
		{
			name:           "Validation Error with Protobuf Accept",
			endpoint:       "/api/players/error/validation",
			acceptHeader:   "application/x-protobuf",
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, resp *http.Response) {
				// Verify content type (should be protobuf error)
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
				
				// Verify error response
				if errorResponse.ErrorCode != "validation_error" {
					t.Errorf("Expected error code 'validation_error', got %s", errorResponse.ErrorCode)
				}
				
				if !strings.Contains(errorResponse.Message, "Invalid player data") {
					t.Errorf("Expected error message to contain 'Invalid player data', got %s", 
						errorResponse.Message)
				}
			},
		},
		{
			name:           "Server Error with Protobuf Accept",
			endpoint:       "/api/players/error/server",
			acceptHeader:   "application/x-protobuf",
			expectedStatus: http.StatusInternalServerError,
			validateResponse: func(t *testing.T, resp *http.Response) {
				// Verify content type (should be protobuf error)
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
				
				// Verify error response
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

// Test handlers for different scenarios

// testPlayerDataHandler returns a normal player data response
func testPlayerDataHandler(w http.ResponseWriter, r *http.Request) {
	// Create test player data
	player := &pb.Player{
		Uid:  12345,
		Name: "Test Player",
		Age:  "25",
		Club: "Test FC",
	}
	
	// Create response
	response := &pb.PlayerDataResponse{
		Players:        []*pb.Player{player},
		CurrencySymbol: "$",
		Metadata:       CreateResponseMetadata("test-request", 1, false),
	}
	
	// Write response using content negotiation
	WriteResponse(w, r, response)
}

// serializationErrorHandler simulates a protobuf serialization error
func serializationErrorHandler(w http.ResponseWriter, r *http.Request) {
	// Create test player data
	player := &pb.Player{
		Uid:  12345,
		Name: "Test Player",
		Age:  "25",
		Club: "Test FC",
	}
	
	// Create response
	response := &pb.PlayerDataResponse{
		Players:        []*pb.Player{player},
		CurrencySymbol: "$",
		Metadata:       CreateResponseMetadata("test-request", 1, false),
	}
	
	// Get error handler
	errorHandler := GetProtobufErrorHandler()
	
	// Simulate serialization error
	if strings.Contains(r.Header.Get("Accept"), "application/x-protobuf") {
		err := fmt.Errorf("simulated serialization error")
		errorHandler.HandleSerializationError(r.Context(), w, r, response, err, "test-dataset")
	} else {
		// For JSON requests, just return normal response
		WriteResponse(w, r, response)
	}
}

// conversionErrorHandler simulates a protobuf conversion error
func conversionErrorHandler(w http.ResponseWriter, r *http.Request) {
	// Create test data (not protobuf)
	data := map[string]interface{}{
		"players": []map[string]interface{}{
			{
				"uid":  12345,
				"name": "Test Player",
				"age":  "25",
				"club": "Test FC",
			},
		},
		"currency_symbol": "$",
	}
	
	// Get error handler
	errorHandler := GetProtobufErrorHandler()
	
	// Simulate conversion error
	if strings.Contains(r.Header.Get("Accept"), "application/x-protobuf") {
		err := fmt.Errorf("simulated conversion error")
		errorHandler.HandleProtobufConversionError(
			r.Context(), w, r, data, err, "test-dataset", "to_protobuf")
	} else {
		// For JSON requests, just return normal response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}
}

// validationErrorHandler returns a validation error response
func validationErrorHandler(w http.ResponseWriter, r *http.Request) {
	// Create error details
	details := []string{
		"Player name is required",
		"Age must be a number",
	}
	
	// Write error response
	WriteErrorResponse(w, r, "validation_error", 
		"Invalid player data: validation failed", 
		details, 
		http.StatusBadRequest)
}

// serverErrorHandler returns a server error response
func serverErrorHandler(w http.ResponseWriter, r *http.Request) {
	// Write error response
	WriteErrorResponse(w, r, "server_error", 
		"Internal server error occurred", 
		[]string{"Database connection failed"}, 
		http.StatusInternalServerError)
}

// TestConcurrentErrorHandling tests error handling under concurrent load
func TestConcurrentErrorHandling(t *testing.T) {
	// Create a test server with our API handlers
	mux := http.NewServeMux()
	
	// Register handlers for different test scenarios
	mux.HandleFunc("/api/players/normal", playerDataHandler)
	mux.HandleFunc("/api/players/error/serialization", serializationErrorHandler)
	mux.HandleFunc("/api/players/error/conversion", conversionErrorHandler)
	
	server := httptest.NewServer(mux)
	defer server.Close()
	
	// Number of concurrent requests
	concurrency := 10
	
	// Create a wait group to synchronize goroutines
	var wg sync.WaitGroup
	wg.Add(concurrency * 3) // 3 endpoints
	
	// Track errors
	errorCh := make(chan error, concurrency*3)
	
	// Test normal endpoint with concurrent requests
	for i := 0; i < concurrency; i++ {
		go func(i int) {
			defer wg.Done()
			
			// Create a request
			req, err := http.NewRequest("GET", server.URL+"/api/players/normal", nil)
			if err != nil {
				errorCh <- fmt.Errorf("failed to create request: %v", err)
				return
			}
			
			// Set Accept header based on request number (alternate)
			if i%2 == 0 {
				req.Header.Set("Accept", "application/x-protobuf")
			} else {
				req.Header.Set("Accept", "application/json")
			}
			
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
			}
		}(i)
	}
	
	// Test serialization error endpoint with concurrent requests
	for i := 0; i < concurrency; i++ {
		go func(i int) {
			defer wg.Done()
			
			// Create a request
			req, err := http.NewRequest("GET", server.URL+"/api/players/error/serialization", nil)
			if err != nil {
				errorCh <- fmt.Errorf("failed to create request: %v", err)
				return
			}
			
			// Set Accept header based on request number (alternate)
			if i%2 == 0 {
				req.Header.Set("Accept", "application/x-protobuf")
			} else {
				req.Header.Set("Accept", "application/json")
			}
			
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
			}
			
			// For protobuf requests, check fallback header
			if i%2 == 0 {
				fallbackHeader := resp.Header.Get("X-Serialization-Fallback")
				if fallbackHeader != string(FallbackReasonMarshalFailed) {
					errorCh <- fmt.Errorf("expected fallback header %s, got %s", 
						FallbackReasonMarshalFailed, fallbackHeader)
				}
			}
		}(i)
	}
	
	// Test conversion error endpoint with concurrent requests
	for i := 0; i < concurrency; i++ {
		go func(i int) {
			defer wg.Done()
			
			// Create a request
			req, err := http.NewRequest("GET", server.URL+"/api/players/error/conversion", nil)
			if err != nil {
				errorCh <- fmt.Errorf("failed to create request: %v", err)
				return
			}
			
			// Set Accept header based on request number (alternate)
			if i%2 == 0 {
				req.Header.Set("Accept", "application/x-protobuf")
			} else {
				req.Header.Set("Accept", "application/json")
			}
			
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
			}
			
			// For protobuf requests, check fallback header
			if i%2 == 0 {
				fallbackHeader := resp.Header.Get("X-Serialization-Fallback")
				if fallbackHeader != string(FallbackReasonConversionFailed) {
					errorCh <- fmt.Errorf("expected fallback header %s, got %s", 
						FallbackReasonConversionFailed, fallbackHeader)
				}
			}
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
}