package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	pb "api/proto"
)

func TestFormatAwarePlayerDataHandler(t *testing.T) {
	// Initialize cache
	InitInMemoryCache()
	defer StopMemCache()
	
	// Create test dataset
	datasetID := "test_format_aware_dataset"
	players := []Player{
		{
			UID:      1,
			Name:     "Test Player 1",
			Position: "ST",
			Club:     "Test Club",
			Division: "Premier Division",
			Overall:  80,
		},
		{
			UID:      2,
			Name:     "Test Player 2",
			Position: "CB",
			Club:     "Test Club",
			Division: "Premier Division",
			Overall:  75,
		},
	}
	currencySymbol := "$"
	
	// Store the test dataset
	SetPlayerData(datasetID, players, currencySymbol)
	
	// Test JSON request
	reqJSON := httptest.NewRequest(http.MethodGet, "/api/players/"+datasetID, nil)
	reqJSON.Header.Set("Accept", "application/json")
	respJSON := httptest.NewRecorder()
	
	// Call the handler
	formatAwarePlayerDataHandler(respJSON, reqJSON)
	
	// Check response
	if respJSON.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, respJSON.Code)
	}
	
	if respJSON.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type %s, got %s", "application/json", respJSON.Header().Get("Content-Type"))
	}
	
	// Test Protobuf request
	reqProto := httptest.NewRequest(http.MethodGet, "/api/players/"+datasetID, nil)
	reqProto.Header.Set("Accept", "application/x-protobuf")
	respProto := httptest.NewRecorder()
	
	// Call the handler
	formatAwarePlayerDataHandler(respProto, reqProto)
	
	// Check response
	if respProto.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, respProto.Code)
	}
	
	if respProto.Header().Get("Content-Type") != "application/x-protobuf" {
		t.Errorf("Expected Content-Type %s, got %s", "application/x-protobuf", respProto.Header().Get("Content-Type"))
	}
	
	// Test cache hit for JSON
	reqJSON2 := httptest.NewRequest(http.MethodGet, "/api/players/"+datasetID, nil)
	reqJSON2.Header.Set("Accept", "application/json")
	respJSON2 := httptest.NewRecorder()
	
	// Call the handler again
	formatAwarePlayerDataHandler(respJSON2, reqJSON2)
	
	// Check response
	if respJSON2.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, respJSON2.Code)
	}
	
	if respJSON2.Header().Get("X-Cache-Source") != "memory" {
		t.Errorf("Expected X-Cache-Source %s, got %s", "memory", respJSON2.Header().Get("X-Cache-Source"))
	}
	
	// Test cache hit for Protobuf
	reqProto2 := httptest.NewRequest(http.MethodGet, "/api/players/"+datasetID, nil)
	reqProto2.Header.Set("Accept", "application/x-protobuf")
	respProto2 := httptest.NewRecorder()
	
	// Call the handler again
	formatAwarePlayerDataHandler(respProto2, reqProto2)
	
	// Check response
	if respProto2.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, respProto2.Code)
	}
	
	if respProto2.Header().Get("X-Cache-Source") != "memory" {
		t.Errorf("Expected X-Cache-Source %s, got %s", "memory", respProto2.Header().Get("X-Cache-Source"))
	}
	
	if respProto2.Header().Get("X-Cache-Format") != "protobuf" {
		t.Errorf("Expected X-Cache-Format %s, got %s", "protobuf", respProto2.Header().Get("X-Cache-Format"))
	}
}

func TestFormatAwarePlayerDataHandlerWithFilters(t *testing.T) {
	// Initialize cache
	InitInMemoryCache()
	defer StopMemCache()
	
	// Create test dataset
	datasetID := "test_format_aware_dataset_filters"
	players := []Player{
		{
			UID:      1,
			Name:     "Test Player 1",
			Position: "ST",
			Club:     "Test Club",
			Division: "Premier Division",
			Overall:  80,
			Age:      "20",
		},
		{
			UID:      2,
			Name:     "Test Player 2",
			Position: "CB",
			Club:     "Test Club",
			Division: "Premier Division",
			Overall:  75,
			Age:      "25",
		},
		{
			UID:      3,
			Name:     "Test Player 3",
			Position: "ST",
			Club:     "Another Club",
			Division: "Second Division",
			Overall:  70,
			Age:      "30",
		},
	}
	currencySymbol := "$"
	
	// Store the test dataset
	SetPlayerData(datasetID, players, currencySymbol)
	
	// Test with position filter
	reqFiltered := httptest.NewRequest(http.MethodGet, "/api/players/"+datasetID+"?position=ST", nil)
	reqFiltered.Header.Set("Accept", "application/json")
	respFiltered := httptest.NewRecorder()
	
	// Call the handler
	formatAwarePlayerDataHandler(respFiltered, reqFiltered)
	
	// Check response
	if respFiltered.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, respFiltered.Code)
	}
	
	// Test with different filter
	reqFiltered2 := httptest.NewRequest(http.MethodGet, "/api/players/"+datasetID+"?minAge=25", nil)
	reqFiltered2.Header.Set("Accept", "application/json")
	respFiltered2 := httptest.NewRecorder()
	
	// Call the handler
	formatAwarePlayerDataHandler(respFiltered2, reqFiltered2)
	
	// Check response
	if respFiltered2.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, respFiltered2.Code)
	}
	
	// Test with combined filters
	reqFiltered3 := httptest.NewRequest(http.MethodGet, "/api/players/"+datasetID+"?position=ST&minAge=25", nil)
	reqFiltered3.Header.Set("Accept", "application/json")
	respFiltered3 := httptest.NewRecorder()
	
	// Call the handler
	formatAwarePlayerDataHandler(respFiltered3, reqFiltered3)
	
	// Check response
	if respFiltered3.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, respFiltered3.Code)
	}
	
	// Test cache hit with filters
	reqFiltered4 := httptest.NewRequest(http.MethodGet, "/api/players/"+datasetID+"?position=ST&minAge=25", nil)
	reqFiltered4.Header.Set("Accept", "application/json")
	respFiltered4 := httptest.NewRecorder()
	
	// Call the handler again
	formatAwarePlayerDataHandler(respFiltered4, reqFiltered4)
	
	// Check response
	if respFiltered4.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, respFiltered4.Code)
	}
	
	if respFiltered4.Header().Get("X-Cache-Source") != "memory" {
		t.Errorf("Expected X-Cache-Source %s, got %s", "memory", respFiltered4.Header().Get("X-Cache-Source"))
	}
}

func TestFormatAwarePlayerDataHandlerMemoryOptimization(t *testing.T) {
	// Initialize cache
	InitInMemoryCache()
	defer StopMemCache()
	
	// Create test context
	ctx := context.Background()
	
	// Create test dataset with many players to test memory optimization
	datasetID := "test_format_aware_dataset_memory"
	players := make([]Player, 100)
	for i := 0; i < 100; i++ {
		players[i] = Player{
			UID:      int64(i + 1),
			Name:     fmt.Sprintf("Test Player %d", i+1),
			Position: "ST",
			Club:     "Test Club",
			Division: "Premier Division",
			Overall:  75 + (i % 10),
			Age:      fmt.Sprintf("%d", 20 + (i % 15)),
			PerformancePercentiles: map[string]map[string]float64{
				"group1": {
					"stat1": 0.75 + float64(i%10)/100.0,
					"stat2": 0.65 + float64(i%10)/100.0,
				},
			},
		}
	}
	currencySymbol := "$"
	
	// Store the test dataset
	SetPlayerData(datasetID, players, currencySymbol)
	
	// Create a protobuf response for memory testing
	protoResponse := &pb.PlayerDataResponse{
		Players:        make([]*pb.Player, 0, len(players)),
		CurrencySymbol: currencySymbol,
		Metadata:       CreateResponseMetadata("test-request", int32(len(players)), false),
	}
	
	// Convert each player to protobuf
	for _, player := range players {
		protoPlayer, err := player.ToProto(ctx)
		if err != nil {
			t.Errorf("Failed to convert player to protobuf: %v", err)
			continue
		}
		protoResponse.Players = append(protoResponse.Players, protoPlayer)
	}
	
	// Get the original size
	originalSize := estimateSize(protoResponse)
	
	// Optimize the protobuf data
	optimized := OptimizeProtobufPlayerData(ctx, protoResponse)
	
	// Get the optimized size
	optimizedSize := estimateSize(optimized)
	
	// Check that optimization reduced the size
	if optimizedSize >= originalSize {
		t.Errorf("Expected optimized size to be less than original size, got original=%d, optimized=%d",
			originalSize, optimizedSize)
	}
	
	// Check that the optimization didn't lose essential data
	if len(optimized.GetPlayers()) != len(protoResponse.GetPlayers()) {
		t.Errorf("Expected %d players after optimization, got %d", 
			len(protoResponse.GetPlayers()), len(optimized.GetPlayers()))
	}
}