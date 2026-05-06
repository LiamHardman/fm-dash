package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	pb "api/proto"
)

func TestFormatAwarePlayerDataHandler(t *testing.T) {
	// Initialize storage system
	InitStore()

	// Initialize cache
	InitInMemoryCache()
	defer StopMemCache()

	// Create test dataset with minimal attributes to avoid configuration dependencies
	datasetID := "test_format_aware_dataset"
	players := []Player{
		{
			UID:      1,
			Name:     "Test Player 1",
			Position: "ST",
			Club:     "Test Club",
			Division: "Premier Division",
			Overall:  80,
			// Add minimal attributes to avoid warnings
			Attributes: map[string]string{
				"Acc": "15", "Pac": "16", "Str": "14", "Sta": "13", "Nat": "12", "Bal": "15", "Jum": "16", "Agi": "14",
				"Agg": "12", "Ant": "15", "Bra": "13", "Cmp": "14", "Cnt": "12", "Dec": "11", "Det": "15", "Fla": "13", "Ldr": "12", "OtB": "16", "Pos": "13", "Tea": "12", "Vis": "14", "Wor": "15",
				"Cor": "12", "Cro": "13", "Dri": "14", "Fin": "16", "Fir": "15", "Fre": "12", "Hea": "14", "Lon": "13", "L Th": "10", "Mar": "12", "Pas": "13", "Pen": "15", "Tck": "10", "Tec": "14",
			},
		},
		{
			UID:      2,
			Name:     "Test Player 2",
			Position: "CB",
			Club:     "Test Club",
			Division: "Premier Division",
			Overall:  75,
			// Add minimal attributes to avoid warnings
			Attributes: map[string]string{
				"Acc": "12", "Pac": "13", "Str": "15", "Sta": "14", "Nat": "13", "Bal": "12", "Jum": "14", "Agi": "11",
				"Agg": "14", "Ant": "13", "Bra": "15", "Cmp": "12", "Cnt": "14", "Dec": "15", "Det": "13", "Fla": "10", "Ldr": "14", "OtB": "12", "Pos": "15", "Tea": "13", "Vis": "12", "Wor": "14",
				"Cor": "13", "Cro": "12", "Dri": "11", "Fin": "10", "Fir": "12", "Fre": "11", "Hea": "15", "Lon": "10", "L Th": "10", "Mar": "15", "Pas": "13", "Pen": "10", "Tck": "14", "Tec": "12",
			},
		},
	}
	currencySymbol := "$"

	// Store the test dataset
	SetPlayerData(datasetID, players, currencySymbol)

	// Retrieve the data to test that enhanced fields are populated
	retrievedPlayers, currencySymbolRetrieved, _ := GetPlayerData(datasetID)
	if len(retrievedPlayers) == 0 {
		t.Fatalf("Failed to retrieve player data or no players returned")
	}

	if len(retrievedPlayers) != 2 {
		t.Fatalf("Expected 2 players, got %d", len(retrievedPlayers))
	}

	// Verify currency symbol is preserved
	if currencySymbolRetrieved != currencySymbol {
		t.Errorf("Expected currency symbol %s, got %s", currencySymbol, currencySymbolRetrieved)
	}

	// Test that the first player has enhanced fields populated
	player1 := retrievedPlayers[0]

	// Check that RoleSpecificOveralls is populated (even if some values might be 0 due to missing role weights)
	if player1.RoleSpecificOveralls == nil {
		t.Error("RoleSpecificOveralls should be populated")
	} else {
		t.Logf("Player 1 RoleSpecificOveralls: %+v", player1.RoleSpecificOveralls)
	}

	// Check that NumericAttributes is populated
	if player1.NumericAttributes == nil {
		t.Error("NumericAttributes should be populated")
	} else {
		t.Logf("Player 1 NumericAttributes: %+v", player1.NumericAttributes)
	}

	// Check that PerformanceStatsNumeric is populated
	if player1.PerformanceStatsNumeric == nil {
		t.Error("PerformanceStatsNumeric should be populated")
	} else {
		t.Logf("Player 1 PerformanceStatsNumeric: %+v", player1.PerformanceStatsNumeric)
	}

	// Test that the second player also has enhanced fields
	player2 := retrievedPlayers[1]

	if player2.RoleSpecificOveralls == nil {
		t.Error("Player 2 RoleSpecificOveralls should be populated")
	}

	if player2.NumericAttributes == nil {
		t.Error("Player 2 NumericAttributes should be populated")
	}

	t.Log("Test completed successfully - enhanced fields are populated when loading from storage")
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
			Age:      fmt.Sprintf("%d", 20+(i%15)),
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

	// Check that optimization did not increase the size
	if optimizedSize > originalSize {
		t.Errorf("Expected optimized size to be no greater than original size, got original=%d, optimized=%d",
			originalSize, optimizedSize)
	}

	// Check that the optimization didn't lose essential data
	if len(optimized.GetPlayers()) != len(protoResponse.GetPlayers()) {
		t.Errorf("Expected %d players after optimization, got %d",
			len(protoResponse.GetPlayers()), len(optimized.GetPlayers()))
	}
}
