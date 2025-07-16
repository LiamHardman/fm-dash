package main

import (
	"testing"

	"api/proto"
	protobuf "google.golang.org/protobuf/proto"
)

func TestProtobufGeneration(t *testing.T) {
	// Test RoleOverallScore
	roleScore := &proto.RoleOverallScore{
		RoleName: "Striker",
		Score:    85,
	}

	// Test serialization
	data, err := protobuf.Marshal(roleScore)
	if err != nil {
		t.Fatalf("Failed to marshal RoleOverallScore: %v", err)
	}

	// Test deserialization
	var unmarshaled proto.RoleOverallScore
	err = protobuf.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal RoleOverallScore: %v", err)
	}

	// Verify data integrity
	if unmarshaled.RoleName != "Striker" {
		t.Errorf("Expected RoleName 'Striker', got '%s'", unmarshaled.RoleName)
	}
	if unmarshaled.Score != 85 {
		t.Errorf("Expected Score 85, got %d", unmarshaled.Score)
	}
}

func TestPlayerProtobuf(t *testing.T) {
	// Test Player struct
	player := &proto.Player{
		Uid:      12345,
		Name:     "Test Player",
		Position: "ST",
		Age:      "25",
		Club:     "Test FC",
		Overall:  80,
		RoleSpecificOveralls: []*proto.RoleOverallScore{
			{RoleName: "Striker", Score: 85},
			{RoleName: "Winger", Score: 75},
		},
	}

	// Test serialization
	data, err := protobuf.Marshal(player)
	if err != nil {
		t.Fatalf("Failed to marshal Player: %v", err)
	}

	// Test deserialization
	var unmarshaled proto.Player
	err = protobuf.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal Player: %v", err)
	}

	// Verify data integrity
	if unmarshaled.Uid != 12345 {
		t.Errorf("Expected Uid 12345, got %d", unmarshaled.Uid)
	}
	if unmarshaled.Name != "Test Player" {
		t.Errorf("Expected Name 'Test Player', got '%s'", unmarshaled.Name)
	}
	if len(unmarshaled.RoleSpecificOveralls) != 2 {
		t.Errorf("Expected 2 role overalls, got %d", len(unmarshaled.RoleSpecificOveralls))
	}
}

func TestDatasetDataProtobuf(t *testing.T) {
	// Test DatasetData struct
	dataset := &proto.DatasetData{
		Players: []*proto.Player{
			{
				Uid:      1,
				Name:     "Player 1",
				Position: "ST",
				Overall:  80,
			},
			{
				Uid:      2,
				Name:     "Player 2",
				Position: "CM",
				Overall:  75,
			},
		},
		CurrencySymbol: "£",
	}

	// Test serialization
	data, err := protobuf.Marshal(dataset)
	if err != nil {
		t.Fatalf("Failed to marshal DatasetData: %v", err)
	}

	// Test deserialization
	var unmarshaled proto.DatasetData
	err = protobuf.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal DatasetData: %v", err)
	}

	// Verify data integrity
	if len(unmarshaled.Players) != 2 {
		t.Errorf("Expected 2 players, got %d", len(unmarshaled.Players))
	}
	if unmarshaled.CurrencySymbol != "£" {
		t.Errorf("Expected CurrencySymbol '£', got '%s'", unmarshaled.CurrencySymbol)
	}
	if unmarshaled.Players[0].Name != "Player 1" {
		t.Errorf("Expected first player name 'Player 1', got '%s'", unmarshaled.Players[0].Name)
	}
}