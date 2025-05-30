package main

import (
	"testing"
)

func TestPlayerStructInitialization(t *testing.T) {
	player := Player{
		UID:           "test-uid-123",
		Name:          "Test Player",
		Position:      "Midfielder",
		Age:           "25",
		Club:          "Test FC",
		Division:      "Premier League",
		TransferValue: "£50M",
		Wage:          "£100K",
		Nationality:   "England",
	}

	// Test basic field assignment
	if player.UID != "test-uid-123" {
		t.Errorf("Expected UID 'test-uid-123', got '%s'", player.UID)
	}
	if player.Name != "Test Player" {
		t.Errorf("Expected Name 'Test Player', got '%s'", player.Name)
	}
	if player.Position != "Midfielder" {
		t.Errorf("Expected Position 'Midfielder', got '%s'", player.Position)
	}
	if player.Age != "25" {
		t.Errorf("Expected Age '25', got '%s'", player.Age)
	}
	if player.Club != "Test FC" {
		t.Errorf("Expected Club 'Test FC', got '%s'", player.Club)
	}
	if player.Division != "Premier League" {
		t.Errorf("Expected Division 'Premier League', got '%s'", player.Division)
	}
	if player.TransferValue != "£50M" {
		t.Errorf("Expected TransferValue '£50M', got '%s'", player.TransferValue)
	}
	if player.Wage != "£100K" {
		t.Errorf("Expected Wage '£100K', got '%s'", player.Wage)
	}
	if player.Nationality != "England" {
		t.Errorf("Expected Nationality 'England', got '%s'", player.Nationality)
	}

	// Test that maps initialize properly
	if player.Attributes == nil {
		player.Attributes = make(map[string]string)
	}
	player.Attributes["Passing"] = "18"
	if player.Attributes["Passing"] != "18" {
		t.Errorf("Expected Attributes['Passing'] '18', got '%s'", player.Attributes["Passing"])
	}

	if player.NumericAttributes == nil {
		player.NumericAttributes = make(map[string]int)
	}
	player.NumericAttributes["Passing"] = 18
	if player.NumericAttributes["Passing"] != 18 {
		t.Errorf("Expected NumericAttributes['Passing'] 18, got %d", player.NumericAttributes["Passing"])
	}
}

func TestRoleOverallScore(t *testing.T) {
	roleScore := RoleOverallScore{
		RoleName: "Central Midfielder",
		Score:    85,
	}

	if roleScore.RoleName != "Central Midfielder" {
		t.Errorf("Expected RoleName 'Central Midfielder', got '%s'", roleScore.RoleName)
	}
	if roleScore.Score != 85 {
		t.Errorf("Expected Score 85, got %d", roleScore.Score)
	}
}

func TestPlayerWithRoleSpecificOveralls(t *testing.T) {
	player := Player{
		UID:  "test-uid",
		Name: "Test Player",
		RoleSpecificOveralls: []RoleOverallScore{
			{RoleName: "Central Midfielder", Score: 85},
			{RoleName: "Attacking Midfielder", Score: 82},
			{RoleName: "Deep Lying Playmaker", Score: 88},
		},
	}

	expectedRoles := 3
	if len(player.RoleSpecificOveralls) != expectedRoles {
		t.Errorf("Expected %d role specific overalls, got %d", expectedRoles, len(player.RoleSpecificOveralls))
	}

	// Test first role
	firstRole := player.RoleSpecificOveralls[0]
	if firstRole.RoleName != "Central Midfielder" {
		t.Errorf("Expected first role 'Central Midfielder', got '%s'", firstRole.RoleName)
	}
	if firstRole.Score != 85 {
		t.Errorf("Expected first role score 85, got %d", firstRole.Score)
	}

	// Test last role
	lastRole := player.RoleSpecificOveralls[2]
	if lastRole.RoleName != "Deep Lying Playmaker" {
		t.Errorf("Expected last role 'Deep Lying Playmaker', got '%s'", lastRole.RoleName)
	}
	if lastRole.Score != 88 {
		t.Errorf("Expected last role score 88, got %d", lastRole.Score)
	}
}

func TestPlayerParseResult(t *testing.T) {
	// Test successful parse result
	player := Player{
		UID:  "test-uid",
		Name: "Test Player",
	}

	result := PlayerParseResult{
		Player: player,
		Err:    nil,
	}

	if result.Player.UID != "test-uid" {
		t.Errorf("Expected Player UID 'test-uid', got '%s'", result.Player.UID)
	}
	if result.Err != nil {
		t.Errorf("Expected no error, got %v", result.Err)
	}
}

func TestUploadResponse(t *testing.T) {
	response := UploadResponse{
		DatasetID:              "dataset-123",
		Message:                "Upload successful",
		DetectedCurrencySymbol: "£",
	}

	if response.DatasetID != "dataset-123" {
		t.Errorf("Expected DatasetID 'dataset-123', got '%s'", response.DatasetID)
	}
	if response.Message != "Upload successful" {
		t.Errorf("Expected Message 'Upload successful', got '%s'", response.Message)
	}
	if response.DetectedCurrencySymbol != "£" {
		t.Errorf("Expected DetectedCurrencySymbol '£', got '%s'", response.DetectedCurrencySymbol)
	}
}

func TestPlayerDataWithCurrency(t *testing.T) {
	players := []Player{
		{UID: "player1", Name: "Player One"},
		{UID: "player2", Name: "Player Two"},
	}

	data := PlayerDataWithCurrency{
		Players:        players,
		CurrencySymbol: "€",
	}

	if len(data.Players) != 2 {
		t.Errorf("Expected 2 players, got %d", len(data.Players))
	}
	if data.Players[0].UID != "player1" {
		t.Errorf("Expected first player UID 'player1', got '%s'", data.Players[0].UID)
	}
	if data.CurrencySymbol != "€" {
		t.Errorf("Expected CurrencySymbol '€', got '%s'", data.CurrencySymbol)
	}
}
