package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	pb "api/proto"
)

func TestGeneratePlayerCacheKey(t *testing.T) {
	tests := []struct {
		name      string
		datasetID string
		filters   map[string]string
		expected  string
	}{
		{
			name:      "No filters",
			datasetID: "dataset123",
			filters:   map[string]string{},
			expected:  "players:dataset123",
		},
		{
			name:      "With filters",
			datasetID: "dataset123",
			filters: map[string]string{
				"position": "ST",
				"minAge":   "18",
				"maxAge":   "30",
			},
			expected: "players:dataset123:filter:position=ST;minAge=18;maxAge=30;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GeneratePlayerCacheKey(tt.datasetID, tt.filters)
			// Since the filter hash order is not guaranteed, we can't do an exact string match
			// Instead, check that the key starts with the expected prefix
			expectedPrefix := "players:" + tt.datasetID
			if len(tt.filters) == 0 {
				if result != tt.expected {
					t.Errorf("GeneratePlayerCacheKey() = %q, want %q", result, tt.expected)
				}
			} else {
				if result[:len(expectedPrefix)] != expectedPrefix {
					t.Errorf("GeneratePlayerCacheKey() = %q, doesn't start with %q", result, expectedPrefix)
				}
				if result[len(expectedPrefix):len(expectedPrefix)+8] != ":filter:" {
					t.Errorf("GeneratePlayerCacheKey() = %q, doesn't contain ':filter:' after dataset ID", result)
				}
			}
		})
	}
}

func TestCachedPlayerDataOperations(t *testing.T) {
	// Initialize cache
	InitInMemoryCache()
	defer StopMemCache()
	
	// Create test context
	ctx := context.Background()
	
	// Create test data
	datasetID := "test_dataset"
	cacheKey := GeneratePlayerCacheKey(datasetID, nil)
	players := []Player{
		{
			UID:      1,
			Name:     "Test Player 1",
			Position: "ST",
			Club:     "Test Club",
		},
		{
			UID:      2,
			Name:     "Test Player 2",
			Position: "CB",
			Club:     "Test Club",
		},
	}
	currencySymbol := "$"
	
	// Cache the player data
	CachePlayerData(ctx, cacheKey, players, currencySymbol, "", 5*time.Minute)
	
	// Test JSON format request
	reqJSON := httptest.NewRequest(http.MethodGet, "/api/players/"+datasetID, nil)
	reqJSON.Header.Set("Accept", "application/json")
	
	// Get cached data for JSON request
	cachedJSON, foundJSON := GetCachedPlayerData(ctx, reqJSON, cacheKey)
	
	if !foundJSON {
		t.Errorf("JSON player data not found in cache")
	}
	
	if cachedJSON == nil {
		t.Errorf("JSON player data is nil")
	} else {
		if cachedJSON.Format != FormatTypeJSON {
			t.Errorf("Expected JSON format, got %s", cachedJSON.Format)
		}
		
		if len(cachedJSON.JSONData) != len(players) {
			t.Errorf("Expected %d players, got %d", len(players), len(cachedJSON.JSONData))
		}
		
		if cachedJSON.CurrencySymbol != currencySymbol {
			t.Errorf("Expected currency symbol %s, got %s", currencySymbol, cachedJSON.CurrencySymbol)
		}
	}
	
	// Test Protobuf format request
	reqProto := httptest.NewRequest(http.MethodGet, "/api/players/"+datasetID, nil)
	reqProto.Header.Set("Accept", "application/x-protobuf")
	
	// Get cached data for Protobuf request
	cachedProto, foundProto := GetCachedPlayerData(ctx, reqProto, cacheKey)
	
	if !foundProto {
		t.Errorf("Protobuf player data not found in cache")
	}
	
	if cachedProto == nil {
		t.Errorf("Protobuf player data is nil")
	} else {
		if cachedProto.Format != FormatTypeProtobuf {
			t.Errorf("Expected Protobuf format, got %s", cachedProto.Format)
		}
		
		if cachedProto.ProtobufData == nil {
			t.Errorf("Protobuf data is nil")
		} else {
			if len(cachedProto.ProtobufData.GetPlayers()) != len(players) {
				t.Errorf("Expected %d protobuf players, got %d", 
					len(players), len(cachedProto.ProtobufData.GetPlayers()))
			}
			
			if cachedProto.ProtobufData.GetCurrencySymbol() != currencySymbol {
				t.Errorf("Expected protobuf currency symbol %s, got %s", 
					currencySymbol, cachedProto.ProtobufData.GetCurrencySymbol())
			}
		}
		
		if cachedProto.CurrencySymbol != currencySymbol {
			t.Errorf("Expected currency symbol %s, got %s", currencySymbol, cachedProto.CurrencySymbol)
		}
	}
	
	// Test writing response
	respJSON := httptest.NewRecorder()
	err := WritePlayerDataResponse(ctx, respJSON, reqJSON, cachedJSON)
	if err != nil {
		t.Errorf("Error writing JSON response: %v", err)
	}
	
	if respJSON.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected JSON content type, got %s", respJSON.Header().Get("Content-Type"))
	}
	
	if respJSON.Header().Get("X-Cache-Format") != "json" {
		t.Errorf("Expected X-Cache-Format: json, got %s", respJSON.Header().Get("X-Cache-Format"))
	}
	
	respProto := httptest.NewRecorder()
	err = WritePlayerDataResponse(ctx, respProto, reqProto, cachedProto)
	if err != nil {
		t.Errorf("Error writing Protobuf response: %v", err)
	}
	
	if respProto.Header().Get("Content-Type") != "application/x-protobuf" {
		t.Errorf("Expected Protobuf content type, got %s", respProto.Header().Get("Content-Type"))
	}
	
	if respProto.Header().Get("X-Cache-Format") != "protobuf" {
		t.Errorf("Expected X-Cache-Format: protobuf, got %s", respProto.Header().Get("X-Cache-Format"))
	}
}

func TestOptimizeProtobufPlayerData(t *testing.T) {
	ctx := context.Background()
	
	// Create a test protobuf response
	protoResponse := &pb.PlayerDataResponse{
		Players: []*pb.Player{
			{
				Uid:      1,
				Name:     "Test Player 1",
				Position: "ST",
				Club:     "Test Club",
				PerformancePercentiles: map[string]*pb.PerformancePercentileMap{
					"group1": {
						Percentiles: map[string]float64{
							"stat1": 0.95,
							"stat2": 0.85,
						},
					},
				},
			},
			{
				Uid:      2,
				Name:     "Test Player 2",
				Position: "CB",
				Club:     "Test Club",
				PerformancePercentiles: map[string]*pb.PerformancePercentileMap{
					"group1": {
						Percentiles: map[string]float64{
							"stat1": 0.75,
							"stat2": 0.65,
						},
					},
				},
			},
		},
		CurrencySymbol: "$",
	}
	
	// Optimize the protobuf data
	optimized := OptimizeProtobufPlayerData(ctx, protoResponse)
	
	// Check that the optimization didn't lose essential data
	if len(optimized.GetPlayers()) != len(protoResponse.GetPlayers()) {
		t.Errorf("Expected %d players after optimization, got %d", 
			len(protoResponse.GetPlayers()), len(optimized.GetPlayers()))
	}
	
	// Check that performance percentiles were removed as part of optimization
	for i, player := range optimized.GetPlayers() {
		if player.GetPerformancePercentiles() != nil {
			t.Errorf("Expected nil performance percentiles for player %d after optimization", i)
		}
	}
}