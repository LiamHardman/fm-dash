package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"google.golang.org/protobuf/proto"
)

// TestProtobufStorageValidation validates that protobuf storage actually works correctly
func TestProtobufStorageValidation(t *testing.T) {
	// Initialize test environment
	InitStore()
	InitInMemoryCache()
	InitCacheStorage(context.Background())
	InitializeMemoryOptimizations()

	// Test data integrity between JSON and Protobuf storage
	testData := []Player{
		{
			UID:           12345,
			Name:          "Test Player",
			Position:      "GK",
			Age:           "25",
			Club:          "Test FC",
			Division:      "Test League",
			TransferValue: "£5M",
			Wage:          "£50K",
			Overall:       15,
			Attributes: map[string]string{
				"Ability":   "15",
				"Potential": "18",
			},
			NumericAttributes: map[string]int{
				"Pac": 12,
				"Sho": 8,
				"Pas": 14,
			},
			ParsedPositions: []string{"GK"},
			ShortPositions:  []string{"GK"},
		},
	}

	currency := "£"

	// Test JSON storage
	t.Run("JSON Storage", func(t *testing.T) {
		os.Setenv("USE_PROTOBUF", "false")
		InitStore()

		// Store data
		datasetID := "test-json-dataset"
		SetPlayerData(datasetID, testData, currency)

		// Retrieve data
		retrievedPlayers, retrievedCurrency, found := GetPlayerData(datasetID)
		if !found {
			t.Fatal("Data not found in JSON storage")
		}

		// Validate data integrity
		validateDataIntegrity(t, "JSON", testData, retrievedPlayers, currency, retrievedCurrency)
	})

	// Test Protobuf storage
	t.Run("Protobuf Storage", func(t *testing.T) {
		os.Setenv("USE_PROTOBUF", "true")
		InitStore()

		// Store data
		datasetID := "test-protobuf-dataset"
		SetPlayerData(datasetID, testData, currency)

		// Retrieve data
		retrievedPlayers, retrievedCurrency, found := GetPlayerData(datasetID)
		if !found {
			t.Fatal("Data not found in Protobuf storage")
		}

		// Validate data integrity
		validateDataIntegrity(t, "Protobuf", testData, retrievedPlayers, currency, retrievedCurrency)
	})

	// Test data consistency between backends
	t.Run("Cross-Backend Consistency", func(t *testing.T) {
		// Store with JSON
		os.Setenv("USE_PROTOBUF", "false")
		InitStore()
		jsonDatasetID := "consistency-json"
		SetPlayerData(jsonDatasetID, testData, currency)
		jsonPlayers, jsonCurrency, _ := GetPlayerData(jsonDatasetID)

		// Store with Protobuf
		os.Setenv("USE_PROTOBUF", "true")
		InitStore()
		protobufDatasetID := "consistency-protobuf"
		SetPlayerData(protobufDatasetID, testData, currency)
		protobufPlayers, protobufCurrency, _ := GetPlayerData(protobufDatasetID)

		// Compare results
		if len(jsonPlayers) != len(protobufPlayers) {
			t.Errorf("Player count mismatch: JSON=%d, Protobuf=%d", len(jsonPlayers), len(protobufPlayers))
		}

		if jsonCurrency != protobufCurrency {
			t.Errorf("Currency mismatch: JSON=%s, Protobuf=%s", jsonCurrency, protobufCurrency)
		}

		// Compare first player in detail
		if len(jsonPlayers) > 0 && len(protobufPlayers) > 0 {
			comparePlayerData(t, jsonPlayers[0], protobufPlayers[0])
		}
	})
}

func validateDataIntegrity(t *testing.T, backend string, original []Player, retrieved []Player, originalCurrency, retrievedCurrency string) {
	if len(original) != len(retrieved) {
		t.Errorf("%s: Player count mismatch: original=%d, retrieved=%d", backend, len(original), len(retrieved))
		return
	}

	if originalCurrency != retrievedCurrency {
		t.Errorf("%s: Currency mismatch: original=%s, retrieved=%s", backend, originalCurrency, retrievedCurrency)
	}

	for i, originalPlayer := range original {
		retrievedPlayer := retrieved[i]

		// Check basic fields
		if originalPlayer.UID != retrievedPlayer.UID {
			t.Errorf("%s: Player %d UID mismatch: original=%d, retrieved=%d", backend, i, originalPlayer.UID, retrievedPlayer.UID)
		}

		if originalPlayer.Name != retrievedPlayer.Name {
			t.Errorf("%s: Player %d Name mismatch: original=%s, retrieved=%s", backend, i, originalPlayer.Name, retrievedPlayer.Name)
		}

		if originalPlayer.Position != retrievedPlayer.Position {
			t.Errorf("%s: Player %d Position mismatch: original=%s, retrieved=%s", backend, i, originalPlayer.Position, retrievedPlayer.Position)
		}

		if originalPlayer.Overall != retrievedPlayer.Overall {
			t.Errorf("%s: Player %d Overall mismatch: original=%d, retrieved=%d", backend, i, originalPlayer.Overall, retrievedPlayer.Overall)
		}

		// Check maps
		if len(originalPlayer.Attributes) != len(retrievedPlayer.Attributes) {
			t.Errorf("%s: Player %d Attributes count mismatch: original=%d, retrieved=%d", backend, i, len(originalPlayer.Attributes), len(retrievedPlayer.Attributes))
		}

		if len(originalPlayer.NumericAttributes) != len(retrievedPlayer.NumericAttributes) {
			t.Errorf("%s: Player %d NumericAttributes count mismatch: original=%d, retrieved=%d", backend, i, len(originalPlayer.NumericAttributes), len(retrievedPlayer.NumericAttributes))
		}

		// Check slices
		if len(originalPlayer.ParsedPositions) != len(retrievedPlayer.ParsedPositions) {
			t.Errorf("%s: Player %d ParsedPositions count mismatch: original=%d, retrieved=%d", backend, i, len(originalPlayer.ParsedPositions), len(retrievedPlayer.ParsedPositions))
		}
	}

	t.Logf("%s storage validation passed for %d players", backend, len(original))
}

func comparePlayerData(t *testing.T, jsonPlayer, protobufPlayer Player) {
	if jsonPlayer.UID != protobufPlayer.UID {
		t.Errorf("Cross-backend UID mismatch: JSON=%d, Protobuf=%d", jsonPlayer.UID, protobufPlayer.UID)
	}

	if jsonPlayer.Name != protobufPlayer.Name {
		t.Errorf("Cross-backend Name mismatch: JSON=%s, Protobuf=%s", jsonPlayer.Name, protobufPlayer.Name)
	}

	if jsonPlayer.Overall != protobufPlayer.Overall {
		t.Errorf("Cross-backend Overall mismatch: JSON=%d, Protobuf=%d", jsonPlayer.Overall, protobufPlayer.Overall)
	}

	// Compare attributes
	for key, jsonValue := range jsonPlayer.Attributes {
		if protobufValue, exists := protobufPlayer.Attributes[key]; !exists {
			t.Errorf("Cross-backend missing attribute in Protobuf: %s", key)
		} else if jsonValue != protobufValue {
			t.Errorf("Cross-backend attribute mismatch for %s: JSON=%s, Protobuf=%s", key, jsonValue, protobufValue)
		}
	}

	// Compare numeric attributes
	for key, jsonValue := range jsonPlayer.NumericAttributes {
		if protobufValue, exists := protobufPlayer.NumericAttributes[key]; !exists {
			t.Errorf("Cross-backend missing numeric attribute in Protobuf: %s", key)
		} else if jsonValue != protobufValue {
			t.Errorf("Cross-backend numeric attribute mismatch for %s: JSON=%d, Protobuf=%d", key, jsonValue, protobufValue)
		}
	}

	t.Log("Cross-backend player data comparison passed")
}

// TestProtobufConversionDirectly tests the protobuf conversion functions directly
func TestProtobufConversionDirectly(t *testing.T) {
	// Test player conversion
	originalPlayer := Player{
		UID:      67890,
		Name:     "Direct Test Player",
		Position: "D (C)",
		Age:      "28",
		Club:     "Direct FC",
		Division: "Direct League",
		Overall:  16,
		Attributes: map[string]string{
			"Personality": "Balanced",
			"Media":       "Evasive",
		},
		NumericAttributes: map[string]int{
			"Pac": 15,
			"Def": 17,
		},
		ParsedPositions: []string{"DC"},
		ShortPositions:  []string{"DC"},
	}

	// Test conversion to protobuf and back
	ctx := context.Background()
	protoPlayer, err := originalPlayer.ToProto(ctx)
	if err != nil {
		t.Fatalf("ToProto() failed: %v", err)
	}
	if protoPlayer == nil {
		t.Fatal("ToProto() returned nil")
	}

	convertedPlayer, err := PlayerFromProto(ctx, protoPlayer)
	if err != nil {
		t.Fatalf("PlayerFromProto() failed: %v", err)
	}

	// Validate conversion
	if originalPlayer.UID != convertedPlayer.UID {
		t.Errorf("UID conversion failed: original=%d, converted=%d", originalPlayer.UID, convertedPlayer.UID)
	}

	if originalPlayer.Name != convertedPlayer.Name {
		t.Errorf("Name conversion failed: original=%s, converted=%s", originalPlayer.Name, convertedPlayer.Name)
	}

	if originalPlayer.Overall != convertedPlayer.Overall {
		t.Errorf("Overall conversion failed: original=%d, converted=%d", originalPlayer.Overall, convertedPlayer.Overall)
	}

	// Test attributes
	for key, originalValue := range originalPlayer.Attributes {
		if convertedValue, exists := convertedPlayer.Attributes[key]; !exists {
			t.Errorf("Missing attribute after conversion: %s", key)
		} else if originalValue != convertedValue {
			t.Errorf("Attribute conversion failed for %s: original=%s, converted=%s", key, originalValue, convertedValue)
		}
	}

	t.Log("Direct protobuf conversion test passed")
}

// TestProtobufSerializationSize compares serialization sizes
func TestProtobufSerializationSize(t *testing.T) {
	// Create test data
	players := make([]Player, 10)
	for i := 0; i < 10; i++ {
		players[i] = Player{
			UID:      int64(i + 1000),
			Name:     fmt.Sprintf("Size Test Player %d", i),
			Position: "M (C)",
			Age:      "25",
			Club:     "Size Test FC",
			Division: "Size Test League",
			Overall:  15 + i,
			Attributes: map[string]string{
				"Personality": "Balanced",
				"Media":       "Evasive",
				"Ability":     fmt.Sprintf("%d", 15+i),
			},
			NumericAttributes: map[string]int{
				"Pac": 12 + i,
				"Sho": 10 + i,
				"Pas": 14 + i,
			},
			ParsedPositions: []string{"MC"},
			ShortPositions:  []string{"MC"},
		}
	}

	datasetData := PlayerDataWithCurrency{
		Players:        players,
		CurrencySymbol: "£",
	}

	// JSON serialization
	jsonData, err := json.Marshal(datasetData)
	if err != nil {
		t.Fatalf("JSON serialization failed: %v", err)
	}

	// Protobuf serialization
	ctx := context.Background()
	protoData, err := datasetData.ToProto(ctx)
	if err != nil {
		t.Fatalf("DatasetData ToProto failed: %v", err)
	}
	protobufData, err := proto.Marshal(protoData)
	if err != nil {
		t.Fatalf("Protobuf serialization failed: %v", err)
	}

	jsonSize := len(jsonData)
	protobufSize := len(protobufData)

	t.Logf("Serialization size comparison:")
	t.Logf("JSON size: %d bytes", jsonSize)
	t.Logf("Protobuf size: %d bytes", protobufSize)

	if protobufSize > 0 {
		reduction := float64(jsonSize-protobufSize) / float64(jsonSize) * 100
		t.Logf("Size reduction: %.2f%%", reduction)

		if reduction < 10 {
			t.Errorf("Expected at least 10%% size reduction, got %.2f%%", reduction)
		}
	} else {
		t.Error("Protobuf serialization produced empty data")
	}
}
