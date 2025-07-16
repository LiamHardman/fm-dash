package main

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"api/proto"
)

// Logging functions are already defined in handlers.go

func init() {
	// Set up basic logging for tests
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)
}

func TestRoleOverallScoreToProto(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		role     *RoleOverallScore
		wantErr  bool
		expected *proto.RoleOverallScore
	}{
		{
			name: "valid role conversion",
			role: &RoleOverallScore{
				RoleName: "Striker",
				Score:    85,
			},
			wantErr: false,
			expected: &proto.RoleOverallScore{
				RoleName: "Striker",
				Score:    85,
			},
		},
		{
			name:    "nil role",
			role:    nil,
			wantErr: true,
		},
		{
			name: "empty role name",
			role: &RoleOverallScore{
				RoleName: "",
				Score:    75,
			},
			wantErr: false,
			expected: &proto.RoleOverallScore{
				RoleName: "",
				Score:    75,
			},
		},
		{
			name: "zero score",
			role: &RoleOverallScore{
				RoleName: "Goalkeeper",
				Score:    0,
			},
			wantErr: false,
			expected: &proto.RoleOverallScore{
				RoleName: "Goalkeeper",
				Score:    0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.role == nil {
				// Test nil case separately
				var nilRole *RoleOverallScore
				_, err := nilRole.ToProto(ctx)
				if !tt.wantErr {
					t.Errorf("ToProto() expected no error for nil role but got one")
				}
				if err == nil && tt.wantErr {
					t.Errorf("ToProto() expected error for nil role but got none")
				}
				return
			}
			
			result, err := tt.role.ToProto(ctx)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("ToProto() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ToProto() unexpected error: %v", err)
				return
			}

			if result.RoleName != tt.expected.RoleName {
				t.Errorf("ToProto() RoleName = %v, want %v", result.RoleName, tt.expected.RoleName)
			}

			if result.Score != tt.expected.Score {
				t.Errorf("ToProto() Score = %v, want %v", result.Score, tt.expected.Score)
			}
		})
	}
}

func TestRoleOverallScoreFromProto(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		protoRole *proto.RoleOverallScore
		wantErr   bool
		expected  *RoleOverallScore
	}{
		{
			name: "valid protobuf conversion",
			protoRole: &proto.RoleOverallScore{
				RoleName: "Midfielder",
				Score:    78,
			},
			wantErr: false,
			expected: &RoleOverallScore{
				RoleName: "Midfielder",
				Score:    78,
			},
		},
		{
			name:      "nil protobuf role",
			protoRole: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RoleOverallScoreFromProto(ctx, tt.protoRole)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("RoleOverallScoreFromProto() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("RoleOverallScoreFromProto() unexpected error: %v", err)
				return
			}

			if result.RoleName != tt.expected.RoleName {
				t.Errorf("RoleOverallScoreFromProto() RoleName = %v, want %v", result.RoleName, tt.expected.RoleName)
			}

			if result.Score != tt.expected.Score {
				t.Errorf("RoleOverallScoreFromProto() Score = %v, want %v", result.Score, tt.expected.Score)
			}
		})
	}
}

func TestPlayerToProto(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		player   *Player
		wantErr  bool
	}{
		{
			name: "valid player conversion",
			player: &Player{
				UID:      12345,
				Name:     "Test Player",
				Position: "ST",
				Age:      "25",
				Club:     "Test FC",
				Overall:  80,
				Attributes: map[string]string{
					"Pace": "15",
				},
				NumericAttributes: map[string]int{
					"Pace": 15,
				},
				PerformanceStatsNumeric: map[string]float64{
					"Goals": 10.5,
				},
				PerformancePercentiles: map[string]map[string]float64{
					"Attacking": {
						"Goals": 85.5,
					},
				},
				ParsedPositions: []string{"ST", "CF"},
				RoleSpecificOveralls: []RoleOverallScore{
					{RoleName: "Striker", Score: 85},
					{RoleName: "Winger", Score: 75},
				},
				PAC: 85,
				SHO: 80,
			},
			wantErr: false,
		},
		{
			name:    "nil player",
			player:  nil,
			wantErr: true,
		},
		{
			name: "player with empty fields",
			player: &Player{
				UID:  0,
				Name: "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.player.ToProto(ctx)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("ToProto() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ToProto() unexpected error: %v", err)
				return
			}

			// Verify basic fields
			if result.Uid != tt.player.UID {
				t.Errorf("ToProto() UID = %v, want %v", result.Uid, tt.player.UID)
			}

			if result.Name != tt.player.Name {
				t.Errorf("ToProto() Name = %v, want %v", result.Name, tt.player.Name)
			}

			// Verify role conversions
			if len(result.RoleSpecificOveralls) != len(tt.player.RoleSpecificOveralls) {
				t.Errorf("ToProto() RoleSpecificOveralls length = %v, want %v", 
					len(result.RoleSpecificOveralls), len(tt.player.RoleSpecificOveralls))
			}
		})
	}
}

func TestPlayerFromProto(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		protoPlayer *proto.Player
		wantErr     bool
	}{
		{
			name: "valid protobuf player conversion",
			protoPlayer: &proto.Player{
				Uid:      54321,
				Name:     "Proto Player",
				Position: "CM",
				Age:      "28",
				Club:     "Proto FC",
				Overall:  75,
				Attributes: map[string]string{
					"Passing": "18",
				},
				NumericAttributes: map[string]int32{
					"Passing": 18,
				},
				PerformanceStatsNumeric: map[string]float64{
					"Assists": 8.2,
				},
				PerformancePercentiles: map[string]*proto.PerformancePercentileMap{
					"Passing": {
						Percentiles: map[string]float64{
							"Assists": 92.1,
						},
					},
				},
				ParsedPositions: []string{"CM", "CAM"},
				RoleSpecificOveralls: []*proto.RoleOverallScore{
					{RoleName: "Midfielder", Score: 82},
				},
				Pac: 70,
				Pas: 88,
			},
			wantErr: false,
		},
		{
			name:        "nil protobuf player",
			protoPlayer: nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := PlayerFromProto(ctx, tt.protoPlayer)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("PlayerFromProto() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("PlayerFromProto() unexpected error: %v", err)
				return
			}

			// Verify basic fields
			if result.UID != tt.protoPlayer.Uid {
				t.Errorf("PlayerFromProto() UID = %v, want %v", result.UID, tt.protoPlayer.Uid)
			}

			if result.Name != tt.protoPlayer.Name {
				t.Errorf("PlayerFromProto() Name = %v, want %v", result.Name, tt.protoPlayer.Name)
			}

			// Verify numeric attributes conversion
			if len(result.NumericAttributes) != len(tt.protoPlayer.NumericAttributes) {
				t.Errorf("PlayerFromProto() NumericAttributes length = %v, want %v", 
					len(result.NumericAttributes), len(tt.protoPlayer.NumericAttributes))
			}
		})
	}
}

func TestDatasetDataToProto(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		dataset  *PlayerDataWithCurrency
		wantErr  bool
	}{
		{
			name: "valid dataset conversion",
			dataset: &PlayerDataWithCurrency{
				Players: []Player{
					{
						UID:      1,
						Name:     "Player 1",
						Position: "ST",
						Overall:  80,
					},
					{
						UID:      2,
						Name:     "Player 2",
						Position: "CM",
						Overall:  75,
					},
				},
				CurrencySymbol: "£",
			},
			wantErr: false,
		},
		{
			name:    "nil dataset",
			dataset: nil,
			wantErr: true,
		},
		{
			name: "empty dataset",
			dataset: &PlayerDataWithCurrency{
				Players:        []Player{},
				CurrencySymbol: "$",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.dataset == nil {
				// Test nil case separately
				var nilDataset *PlayerDataWithCurrency
				_, err := nilDataset.ToProto(ctx)
				if !tt.wantErr {
					t.Errorf("ToProto() expected no error for nil dataset but got one")
				}
				if err == nil && tt.wantErr {
					t.Errorf("ToProto() expected error for nil dataset but got none")
				}
				return
			}
			
			result, err := tt.dataset.ToProto(ctx)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("ToProto() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ToProto() unexpected error: %v", err)
				return
			}

			if len(result.Players) != len(tt.dataset.Players) {
				t.Errorf("ToProto() Players length = %v, want %v", 
					len(result.Players), len(tt.dataset.Players))
			}

			if result.CurrencySymbol != tt.dataset.CurrencySymbol {
				t.Errorf("ToProto() CurrencySymbol = %v, want %v", 
					result.CurrencySymbol, tt.dataset.CurrencySymbol)
			}
		})
	}
}

func TestDatasetDataFromProto(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		protoDataset *proto.DatasetData
		wantErr      bool
	}{
		{
			name: "valid protobuf dataset conversion",
			protoDataset: &proto.DatasetData{
				Players: []*proto.Player{
					{
						Uid:      100,
						Name:     "Proto Player 1",
						Position: "GK",
						Overall:  85,
					},
				},
				CurrencySymbol: "€",
			},
			wantErr: false,
		},
		{
			name:         "nil protobuf dataset",
			protoDataset: nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DatasetDataFromProto(ctx, tt.protoDataset)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("DatasetDataFromProto() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("DatasetDataFromProto() unexpected error: %v", err)
				return
			}

			if len(result.Players) != len(tt.protoDataset.Players) {
				t.Errorf("DatasetDataFromProto() Players length = %v, want %v", 
					len(result.Players), len(tt.protoDataset.Players))
			}

			if result.CurrencySymbol != tt.protoDataset.CurrencySymbol {
				t.Errorf("DatasetDataFromProto() CurrencySymbol = %v, want %v", 
					result.CurrencySymbol, tt.protoDataset.CurrencySymbol)
			}
		})
	}
}

// TestRoundTripConversion tests that data remains intact after converting to protobuf and back
func TestRoundTripConversion(t *testing.T) {
	ctx := context.Background()

	original := &Player{
		UID:      99999,
		Name:     "Round Trip Player",
		Position: "RW",
		Age:      "22",
		Club:     "Test United",
		Overall:  88,
		Attributes: map[string]string{
			"Pace":     "18",
			"Shooting": "16",
		},
		NumericAttributes: map[string]int{
			"Pace":     18,
			"Shooting": 16,
		},
		PerformanceStatsNumeric: map[string]float64{
			"Goals":   15.0,
			"Assists": 8.5,
		},
		PerformancePercentiles: map[string]map[string]float64{
			"Attacking": {
				"Goals":   95.2,
				"Assists": 87.3,
			},
		},
		ParsedPositions: []string{"RW", "RM", "RF"},
		RoleSpecificOveralls: []RoleOverallScore{
			{RoleName: "Winger", Score: 90},
			{RoleName: "Inside Forward", Score: 85},
		},
		PAC: 95,
		SHO: 82,
		PAS: 78,
	}

	// Convert to protobuf
	protoPlayer, err := original.ToProto(ctx)
	if err != nil {
		t.Fatalf("Failed to convert to protobuf: %v", err)
	}

	// Convert back to native struct
	converted, err := PlayerFromProto(ctx, protoPlayer)
	if err != nil {
		t.Fatalf("Failed to convert from protobuf: %v", err)
	}

	// Verify key fields are preserved
	if converted.UID != original.UID {
		t.Errorf("Round trip UID mismatch: got %v, want %v", converted.UID, original.UID)
	}

	if converted.Name != original.Name {
		t.Errorf("Round trip Name mismatch: got %v, want %v", converted.Name, original.Name)
	}

	if len(converted.RoleSpecificOveralls) != len(original.RoleSpecificOveralls) {
		t.Errorf("Round trip RoleSpecificOveralls length mismatch: got %v, want %v", 
			len(converted.RoleSpecificOveralls), len(original.RoleSpecificOveralls))
	}

	// Verify role data integrity
	for i, role := range converted.RoleSpecificOveralls {
		if role.RoleName != original.RoleSpecificOveralls[i].RoleName {
			t.Errorf("Round trip role name mismatch at index %d: got %v, want %v", 
				i, role.RoleName, original.RoleSpecificOveralls[i].RoleName)
		}
		if role.Score != original.RoleSpecificOveralls[i].Score {
			t.Errorf("Round trip role score mismatch at index %d: got %v, want %v", 
				i, role.Score, original.RoleSpecificOveralls[i].Score)
		}
	}
}