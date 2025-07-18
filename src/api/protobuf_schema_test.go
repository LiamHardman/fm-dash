package main

import (
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	pb "api/proto"
)

func TestProtobufSchemaExtensions(t *testing.T) {
	// Test ResponseMetadata
	t.Run("ResponseMetadata creation and serialization", func(t *testing.T) {
		metadata := &pb.ResponseMetadata{
			Timestamp:  time.Now().Unix(),
			ApiVersion: "1.0",
			FromCache:  true,
			RequestId:  "test-123",
			TotalCount: 100,
		}

		// Test serialization
		data, err := proto.Marshal(metadata)
		if err != nil {
			t.Fatalf("Failed to marshal ResponseMetadata: %v", err)
		}

		// Test deserialization
		var deserialized pb.ResponseMetadata
		if err := proto.Unmarshal(data, &deserialized); err != nil {
			t.Fatalf("Failed to unmarshal ResponseMetadata: %v", err)
		}

		// Verify fields
		if deserialized.ApiVersion != metadata.ApiVersion {
			t.Errorf("API version mismatch: got %s, want %s", deserialized.ApiVersion, metadata.ApiVersion)
		}
		if deserialized.RequestId != metadata.RequestId {
			t.Errorf("Request ID mismatch: got %s, want %s", deserialized.RequestId, metadata.RequestId)
		}
		if deserialized.TotalCount != metadata.TotalCount {
			t.Errorf("Total count mismatch: got %d, want %d", deserialized.TotalCount, metadata.TotalCount)
		}
	})

	// Test PaginationInfo
	t.Run("PaginationInfo creation and serialization", func(t *testing.T) {
		pagination := &pb.PaginationInfo{
			Page:        2,
			PerPage:     50,
			TotalPages:  10,
			TotalCount:  500,
			HasNext:     true,
			HasPrevious: true,
		}

		// Test serialization
		data, err := proto.Marshal(pagination)
		if err != nil {
			t.Fatalf("Failed to marshal PaginationInfo: %v", err)
		}

		// Test deserialization
		var deserialized pb.PaginationInfo
		if err := proto.Unmarshal(data, &deserialized); err != nil {
			t.Fatalf("Failed to unmarshal PaginationInfo: %v", err)
		}

		// Verify fields
		if deserialized.Page != pagination.Page {
			t.Errorf("Page mismatch: got %d, want %d", deserialized.Page, pagination.Page)
		}
		if deserialized.HasNext != pagination.HasNext {
			t.Errorf("HasNext mismatch: got %v, want %v", deserialized.HasNext, pagination.HasNext)
		}
	})

	// Test PlayerDataResponse
	t.Run("PlayerDataResponse creation and serialization", func(t *testing.T) {
		response := &pb.PlayerDataResponse{
			Players: []*pb.Player{
				{
					Uid:  1,
					Name: "Test Player 1",
					Age:  "25",
				},
				{
					Uid:  2,
					Name: "Test Player 2",
					Age:  "28",
				},
			},
			CurrencySymbol: "£",
			Metadata: &pb.ResponseMetadata{
				Timestamp:  time.Now().Unix(),
				ApiVersion: "1.0",
				RequestId:  "test-456",
				TotalCount: 2,
			},
			Pagination: &pb.PaginationInfo{
				Page:       1,
				PerPage:    10,
				TotalPages: 1,
				TotalCount: 2,
				HasNext:    false,
				HasPrevious: false,
			},
		}

		// Test serialization
		data, err := proto.Marshal(response)
		if err != nil {
			t.Fatalf("Failed to marshal PlayerDataResponse: %v", err)
		}

		// Test deserialization
		var deserialized pb.PlayerDataResponse
		if err := proto.Unmarshal(data, &deserialized); err != nil {
			t.Fatalf("Failed to unmarshal PlayerDataResponse: %v", err)
		}

		// Verify fields
		if len(deserialized.Players) != len(response.Players) {
			t.Errorf("Player count mismatch: got %d, want %d", len(deserialized.Players), len(response.Players))
		}
		if deserialized.CurrencySymbol != response.CurrencySymbol {
			t.Errorf("Currency symbol mismatch: got %s, want %s", deserialized.CurrencySymbol, response.CurrencySymbol)
		}
		if deserialized.Metadata == nil {
			t.Error("Metadata is nil after deserialization")
		}
		if deserialized.Pagination == nil {
			t.Error("Pagination is nil after deserialization")
		}
	})

	// Test RolesResponse
	t.Run("RolesResponse creation and serialization", func(t *testing.T) {
		response := &pb.RolesResponse{
			Roles: []string{"Goalkeeper", "Defender", "Midfielder", "Forward"},
			Metadata: &pb.ResponseMetadata{
				Timestamp:  time.Now().Unix(),
				ApiVersion: "1.0",
				RequestId:  "test-789",
				TotalCount: 4,
			},
		}

		// Test serialization
		data, err := proto.Marshal(response)
		if err != nil {
			t.Fatalf("Failed to marshal RolesResponse: %v", err)
		}

		// Test deserialization
		var deserialized pb.RolesResponse
		if err := proto.Unmarshal(data, &deserialized); err != nil {
			t.Fatalf("Failed to unmarshal RolesResponse: %v", err)
		}

		// Verify fields
		if len(deserialized.Roles) != len(response.Roles) {
			t.Errorf("Roles count mismatch: got %d, want %d", len(deserialized.Roles), len(response.Roles))
		}
		for i, role := range response.Roles {
			if deserialized.Roles[i] != role {
				t.Errorf("Role %d mismatch: got %s, want %s", i, deserialized.Roles[i], role)
			}
		}
	})

	// Test LeaguesResponse
	t.Run("LeaguesResponse creation and serialization", func(t *testing.T) {
		response := &pb.LeaguesResponse{
			Leagues: []string{"Premier League", "Championship", "League One"},
			Metadata: &pb.ResponseMetadata{
				Timestamp:  time.Now().Unix(),
				ApiVersion: "1.0",
				RequestId:  "test-leagues",
				TotalCount: 3,
			},
		}

		// Test serialization
		data, err := proto.Marshal(response)
		if err != nil {
			t.Fatalf("Failed to marshal LeaguesResponse: %v", err)
		}

		// Test deserialization
		var deserialized pb.LeaguesResponse
		if err := proto.Unmarshal(data, &deserialized); err != nil {
			t.Fatalf("Failed to unmarshal LeaguesResponse: %v", err)
		}

		// Verify fields
		if len(deserialized.Leagues) != len(response.Leagues) {
			t.Errorf("Leagues count mismatch: got %d, want %d", len(deserialized.Leagues), len(response.Leagues))
		}
	})

	// Test TeamsResponse
	t.Run("TeamsResponse creation and serialization", func(t *testing.T) {
		response := &pb.TeamsResponse{
			Teams: []string{"Arsenal", "Chelsea", "Liverpool", "Manchester United"},
			Metadata: &pb.ResponseMetadata{
				Timestamp:  time.Now().Unix(),
				ApiVersion: "1.0",
				RequestId:  "test-teams",
				TotalCount: 4,
			},
		}

		// Test serialization
		data, err := proto.Marshal(response)
		if err != nil {
			t.Fatalf("Failed to marshal TeamsResponse: %v", err)
		}

		// Test deserialization
		var deserialized pb.TeamsResponse
		if err := proto.Unmarshal(data, &deserialized); err != nil {
			t.Fatalf("Failed to unmarshal TeamsResponse: %v", err)
		}

		// Verify fields
		if len(deserialized.Teams) != len(response.Teams) {
			t.Errorf("Teams count mismatch: got %d, want %d", len(deserialized.Teams), len(response.Teams))
		}
	})

	// Test SearchResponse
	t.Run("SearchResponse creation and serialization", func(t *testing.T) {
		response := &pb.SearchResponse{
			Players: []*pb.Player{
				{
					Uid:  1,
					Name: "Search Result Player",
					Age:  "26",
				},
			},
			Query: "midfielder",
			Metadata: &pb.ResponseMetadata{
				Timestamp:  time.Now().Unix(),
				ApiVersion: "1.0",
				RequestId:  "test-search",
				TotalCount: 1,
			},
			Pagination: &pb.PaginationInfo{
				Page:       1,
				PerPage:    20,
				TotalPages: 1,
				TotalCount: 1,
				HasNext:    false,
				HasPrevious: false,
			},
		}

		// Test serialization
		data, err := proto.Marshal(response)
		if err != nil {
			t.Fatalf("Failed to marshal SearchResponse: %v", err)
		}

		// Test deserialization
		var deserialized pb.SearchResponse
		if err := proto.Unmarshal(data, &deserialized); err != nil {
			t.Fatalf("Failed to unmarshal SearchResponse: %v", err)
		}

		// Verify fields
		if deserialized.Query != response.Query {
			t.Errorf("Query mismatch: got %s, want %s", deserialized.Query, response.Query)
		}
		if len(deserialized.Players) != len(response.Players) {
			t.Errorf("Players count mismatch: got %d, want %d", len(deserialized.Players), len(response.Players))
		}
	})

	// Test ErrorResponse
	t.Run("ErrorResponse creation and serialization", func(t *testing.T) {
		response := &pb.ErrorResponse{
			ErrorCode: "INVALID_REQUEST",
			Message:   "The request parameters are invalid",
			Details:   []string{"Missing required field: id", "Invalid format: age"},
			Metadata: &pb.ResponseMetadata{
				Timestamp:  time.Now().Unix(),
				ApiVersion: "1.0",
				RequestId:  "test-error",
				TotalCount: 0,
			},
		}

		// Test serialization
		data, err := proto.Marshal(response)
		if err != nil {
			t.Fatalf("Failed to marshal ErrorResponse: %v", err)
		}

		// Test deserialization
		var deserialized pb.ErrorResponse
		if err := proto.Unmarshal(data, &deserialized); err != nil {
			t.Fatalf("Failed to unmarshal ErrorResponse: %v", err)
		}

		// Verify fields
		if deserialized.ErrorCode != response.ErrorCode {
			t.Errorf("Error code mismatch: got %s, want %s", deserialized.ErrorCode, response.ErrorCode)
		}
		if deserialized.Message != response.Message {
			t.Errorf("Message mismatch: got %s, want %s", deserialized.Message, response.Message)
		}
		if len(deserialized.Details) != len(response.Details) {
			t.Errorf("Details count mismatch: got %d, want %d", len(deserialized.Details), len(response.Details))
		}
	})

	// Test GenericResponse
	t.Run("GenericResponse creation and serialization", func(t *testing.T) {
		response := &pb.GenericResponse{
			Data: "Operation completed successfully",
			Metadata: &pb.ResponseMetadata{
				Timestamp:  time.Now().Unix(),
				ApiVersion: "1.0",
				RequestId:  "test-generic",
				TotalCount: 1,
			},
		}

		// Test serialization
		data, err := proto.Marshal(response)
		if err != nil {
			t.Fatalf("Failed to marshal GenericResponse: %v", err)
		}

		// Test deserialization
		var deserialized pb.GenericResponse
		if err := proto.Unmarshal(data, &deserialized); err != nil {
			t.Fatalf("Failed to unmarshal GenericResponse: %v", err)
		}

		// Verify fields
		if deserialized.Data != response.Data {
			t.Errorf("Data mismatch: got %s, want %s", deserialized.Data, response.Data)
		}
	})
}

func TestProtobufSchemaCompatibility(t *testing.T) {
	// Test that all response types work with the content negotiation system
	t.Run("All response types work with ProtobufSerializer", func(t *testing.T) {
		serializer := &ProtobufSerializer{}

		testCases := []struct {
			name string
			data interface{}
		}{
			{
				name: "PlayerDataResponse",
				data: &pb.PlayerDataResponse{
					Players:        []*pb.Player{{Uid: 1, Name: "Test"}},
					CurrencySymbol: "£",
					Metadata:       &pb.ResponseMetadata{Timestamp: time.Now().Unix()},
				},
			},
			{
				name: "RolesResponse",
				data: &pb.RolesResponse{
					Roles:    []string{"Test Role"},
					Metadata: &pb.ResponseMetadata{Timestamp: time.Now().Unix()},
				},
			},
			{
				name: "LeaguesResponse",
				data: &pb.LeaguesResponse{
					Leagues:  []string{"Test League"},
					Metadata: &pb.ResponseMetadata{Timestamp: time.Now().Unix()},
				},
			},
			{
				name: "TeamsResponse",
				data: &pb.TeamsResponse{
					Teams:    []string{"Test Team"},
					Metadata: &pb.ResponseMetadata{Timestamp: time.Now().Unix()},
				},
			},
			{
				name: "SearchResponse",
				data: &pb.SearchResponse{
					Players:  []*pb.Player{{Uid: 1, Name: "Test"}},
					Query:    "test",
					Metadata: &pb.ResponseMetadata{Timestamp: time.Now().Unix()},
				},
			},
			{
				name: "ErrorResponse",
				data: &pb.ErrorResponse{
					ErrorCode: "TEST_ERROR",
					Message:   "Test error message",
					Metadata:  &pb.ResponseMetadata{Timestamp: time.Now().Unix()},
				},
			},
			{
				name: "GenericResponse",
				data: &pb.GenericResponse{
					Data:     "Test data",
					Metadata: &pb.ResponseMetadata{Timestamp: time.Now().Unix()},
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				data, err := serializer.Serialize(tc.data)
				if err != nil {
					t.Errorf("Failed to serialize %s: %v", tc.name, err)
				}
				if len(data) == 0 {
					t.Errorf("Serialized data is empty for %s", tc.name)
				}
			})
		}
	})
}

func TestProtobufImportStructure(t *testing.T) {
	// Test that the protobuf import structure is correct
	t.Run("Player import works correctly", func(t *testing.T) {
		// Create a PlayerDataResponse with Player data
		response := &pb.PlayerDataResponse{
			Players: []*pb.Player{
				{
					Uid:                1,
					Name:               "Test Player",
					Position:           "Midfielder",
					Age:                "25",
					Club:               "Test FC",
					Division:           "Premier League",
					TransferValue:      "£10M",
					Wage:               "£50K",
					Nationality:        "England",
					NationalityIso:     "ENG",
					NationalityFifaCode: "ENG",
					AttributeMasked:    false,
					Attributes:         map[string]string{"Pace": "85"},
					NumericAttributes:  map[string]int32{"Pace": 85},
					ParsedPositions:    []string{"CM", "CAM"},
					ShortPositions:     []string{"CM"},
					PositionGroups:     []string{"Midfield"},
					Pac:                85,
					Sho:                75,
					Pas:                80,
					Dri:                78,
					Def:                65,
					Phy:                82,
					Overall:            79,
					BestRoleOverall:    "Central Midfielder",
					TransferValueAmount: 10000000,
					WageAmount:         50000,
				},
			},
			CurrencySymbol: "£",
			Metadata: &pb.ResponseMetadata{
				Timestamp:  time.Now().Unix(),
				ApiVersion: "1.0",
				RequestId:  "test-import",
				TotalCount: 1,
			},
		}

		// Test serialization
		data, err := proto.Marshal(response)
		if err != nil {
			t.Fatalf("Failed to marshal PlayerDataResponse with full Player data: %v", err)
		}

		// Test deserialization
		var deserialized pb.PlayerDataResponse
		if err := proto.Unmarshal(data, &deserialized); err != nil {
			t.Fatalf("Failed to unmarshal PlayerDataResponse: %v", err)
		}

		// Verify player data
		if len(deserialized.Players) != 1 {
			t.Fatalf("Expected 1 player, got %d", len(deserialized.Players))
		}

		player := deserialized.Players[0]
		if player.Name != "Test Player" {
			t.Errorf("Player name mismatch: got %s, want Test Player", player.Name)
		}
		if player.Overall != 79 {
			t.Errorf("Player overall mismatch: got %d, want 79", player.Overall)
		}
		if len(player.ParsedPositions) != 2 {
			t.Errorf("Expected 2 parsed positions, got %d", len(player.ParsedPositions))
		}
	})
}