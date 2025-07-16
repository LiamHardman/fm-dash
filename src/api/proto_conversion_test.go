package main

import (
	"context"
	"log/slog"
	"math"
	"os"
	"reflect"
	"strings"
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
		name    string
		player  *Player
		wantErr bool
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
		name    string
		dataset *PlayerDataWithCurrency
		wantErr bool
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

// TestRoleOverallScoreEdgeCases tests edge cases for RoleOverallScore conversion
func TestRoleOverallScoreEdgeCases(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		role    *RoleOverallScore
		wantErr bool
	}{
		{
			name: "negative score",
			role: &RoleOverallScore{
				RoleName: "Test Role",
				Score:    -10,
			},
			wantErr: false,
		},
		{
			name: "maximum int score",
			role: &RoleOverallScore{
				RoleName: "Max Role",
				Score:    math.MaxInt32,
			},
			wantErr: false,
		},
		{
			name: "very long role name",
			role: &RoleOverallScore{
				RoleName: strings.Repeat("A", 1000),
				Score:    50,
			},
			wantErr: false,
		},
		{
			name: "unicode characters in role name",
			role: &RoleOverallScore{
				RoleName: "Rôle Spécialisé 中文 🏆",
				Score:    75,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test conversion to protobuf
			protoRole, err := tt.role.ToProto(ctx)
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

			// Test conversion back from protobuf
			converted, err := RoleOverallScoreFromProto(ctx, protoRole)
			if err != nil {
				t.Errorf("RoleOverallScoreFromProto() unexpected error: %v", err)
				return
			}

			// Verify data integrity
			if converted.RoleName != tt.role.RoleName {
				t.Errorf("Round trip RoleName mismatch: got %v, want %v", converted.RoleName, tt.role.RoleName)
			}
			if converted.Score != tt.role.Score {
				t.Errorf("Round trip Score mismatch: got %v, want %v", converted.Score, tt.role.Score)
			}
		})
	}
}

// TestPlayerEdgeCases tests edge cases for Player conversion
func TestPlayerEdgeCases(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		player  *Player
		wantErr bool
	}{
		{
			name: "player with nil maps",
			player: &Player{
				UID:                     1,
				Name:                    "Test Player",
				Attributes:              nil,
				NumericAttributes:       nil,
				PerformanceStatsNumeric: nil,
				PerformancePercentiles:  nil,
				ParsedPositions:         nil,
				RoleSpecificOveralls:    nil,
			},
			wantErr: false,
		},
		{
			name: "player with empty maps",
			player: &Player{
				UID:                     2,
				Name:                    "Empty Maps Player",
				Attributes:              make(map[string]string),
				NumericAttributes:       make(map[string]int),
				PerformanceStatsNumeric: make(map[string]float64),
				PerformancePercentiles:  make(map[string]map[string]float64),
				ParsedPositions:         []string{},
				RoleSpecificOveralls:    []RoleOverallScore{},
			},
			wantErr: false,
		},
		{
			name: "player with extreme values",
			player: &Player{
				UID:                 math.MaxInt64,
				Name:                strings.Repeat("X", 500),
				TransferValueAmount: math.MaxInt64,
				WageAmount:          math.MaxInt64,
				PAC:                 math.MaxInt32,
				SHO:                 math.MinInt32,
				Overall:             -100,
			},
			wantErr: false,
		},
		{
			name: "player with special characters",
			player: &Player{
				UID:         3,
				Name:        "Jürgen Müller-Øvergård 中文名字 🏆⚽",
				Club:        "FC Zürich-München",
				Nationality: "España/Deutschland",
				Attributes: map[string]string{
					"Técnica":    "18",
					"Velocità":   "16",
					"力量":         "14",
					"🏃‍♂️ Speed": "20",
				},
			},
			wantErr: false,
		},
		{
			name: "player with large nested maps",
			player: &Player{
				UID:  4,
				Name: "Large Data Player",
				PerformancePercentiles: map[string]map[string]float64{
					"Attacking": {
						"Goals":         95.5,
						"Assists":       87.2,
						"Shots":         92.1,
						"KeyPasses":     89.7,
						"Dribbles":      91.3,
						"OffensiveRuns": 88.9,
					},
					"Defending": {
						"Tackles":       76.4,
						"Interceptions": 82.1,
						"Clearances":    79.8,
						"Blocks":        74.2,
					},
					"Passing": {
						"ShortPasses":   94.7,
						"LongPasses":    87.3,
						"CrossAccuracy": 85.9,
					},
				},
				NumericAttributes: func() map[string]int {
					attrs := make(map[string]int)
					for i := 0; i < 100; i++ {
						attrs[strings.Repeat("A", i+1)] = i
					}
					return attrs
				}(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test conversion to protobuf
			protoPlayer, err := tt.player.ToProto(ctx)
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

			// Test conversion back from protobuf
			converted, err := PlayerFromProto(ctx, protoPlayer)
			if err != nil {
				t.Errorf("PlayerFromProto() unexpected error: %v", err)
				return
			}

			// Verify basic data integrity
			if converted.UID != tt.player.UID {
				t.Errorf("Round trip UID mismatch: got %v, want %v", converted.UID, tt.player.UID)
			}
			if converted.Name != tt.player.Name {
				t.Errorf("Round trip Name mismatch: got %v, want %v", converted.Name, tt.player.Name)
			}

			// Verify map lengths
			if len(converted.Attributes) != len(tt.player.Attributes) {
				t.Errorf("Round trip Attributes length mismatch: got %v, want %v",
					len(converted.Attributes), len(tt.player.Attributes))
			}
			if len(converted.NumericAttributes) != len(tt.player.NumericAttributes) {
				t.Errorf("Round trip NumericAttributes length mismatch: got %v, want %v",
					len(converted.NumericAttributes), len(tt.player.NumericAttributes))
			}
		})
	}
}

// TestDatasetDataEdgeCases tests edge cases for DatasetData conversion
func TestDatasetDataEdgeCases(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		dataset *PlayerDataWithCurrency
		wantErr bool
	}{
		{
			name: "dataset with unicode currency",
			dataset: &PlayerDataWithCurrency{
				Players:        []Player{},
				CurrencySymbol: "¥€£$₹₽₩",
			},
			wantErr: false,
		},
		{
			name: "dataset with very long currency symbol",
			dataset: &PlayerDataWithCurrency{
				Players:        []Player{},
				CurrencySymbol: strings.Repeat("$", 100),
			},
			wantErr: false,
		},
		{
			name: "dataset with single player",
			dataset: &PlayerDataWithCurrency{
				Players: []Player{
					{UID: 1, Name: "Solo Player"},
				},
				CurrencySymbol: "€",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test conversion to protobuf
			protoDataset, err := tt.dataset.ToProto(ctx)
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

			// Test conversion back from protobuf
			converted, err := DatasetDataFromProto(ctx, protoDataset)
			if err != nil {
				t.Errorf("DatasetDataFromProto() unexpected error: %v", err)
				return
			}

			// Verify data integrity
			if converted.CurrencySymbol != tt.dataset.CurrencySymbol {
				t.Errorf("Round trip CurrencySymbol mismatch: got %v, want %v",
					converted.CurrencySymbol, tt.dataset.CurrencySymbol)
			}
			if len(converted.Players) != len(tt.dataset.Players) {
				t.Errorf("Round trip Players length mismatch: got %v, want %v",
					len(converted.Players), len(tt.dataset.Players))
			}
		})
	}
}

// TestLargeDatasetConversion tests conversion with large datasets
func TestLargeDatasetConversion(t *testing.T) {
	ctx := context.Background()

	// Create a large dataset with many players
	players := make([]Player, 1000)
	for i := 0; i < 1000; i++ {
		players[i] = Player{
			UID:      int64(i + 1),
			Name:     "Player " + strings.Repeat("X", i%50+1),
			Position: []string{"ST", "CM", "CB", "GK"}[i%4],
			Age:      "25",
			Club:     "Club " + strings.Repeat("Y", i%20+1),
			Overall:  i % 100,
			Attributes: map[string]string{
				"Pace":     "15",
				"Shooting": "16",
				"Passing":  "17",
			},
			NumericAttributes: map[string]int{
				"Pace":     15 + i%5,
				"Shooting": 16 + i%3,
				"Passing":  17 + i%7,
			},
			PerformanceStatsNumeric: map[string]float64{
				"Goals":   float64(i % 30),
				"Assists": float64(i % 20),
			},
			PerformancePercentiles: map[string]map[string]float64{
				"Attacking": {
					"Goals":   float64(i%100) + 0.5,
					"Assists": float64((i*2)%100) + 0.3,
				},
			},
			ParsedPositions: []string{[]string{"ST", "CM", "CB", "GK"}[i%4]},
			RoleSpecificOveralls: []RoleOverallScore{
				{RoleName: "Primary", Score: i % 100},
				{RoleName: "Secondary", Score: (i + 10) % 100},
			},
			PAC: i % 20,
			SHO: (i + 5) % 20,
			PAS: (i + 10) % 20,
		}
	}

	dataset := &PlayerDataWithCurrency{
		Players:        players,
		CurrencySymbol: "£",
	}

	// Test conversion to protobuf
	protoDataset, err := dataset.ToProto(ctx)
	if err != nil {
		t.Fatalf("Failed to convert large dataset to protobuf: %v", err)
	}

	// Verify protobuf dataset structure
	if len(protoDataset.Players) != len(dataset.Players) {
		t.Errorf("Large dataset protobuf conversion: Players length = %v, want %v",
			len(protoDataset.Players), len(dataset.Players))
	}

	// Test conversion back from protobuf
	converted, err := DatasetDataFromProto(ctx, protoDataset)
	if err != nil {
		t.Fatalf("Failed to convert large protobuf dataset back: %v", err)
	}

	// Verify converted dataset structure
	if len(converted.Players) != len(dataset.Players) {
		t.Errorf("Large dataset round trip: Players length = %v, want %v",
			len(converted.Players), len(dataset.Players))
	}

	// Spot check some players for data integrity
	checkIndices := []int{0, 100, 500, 999}
	for _, idx := range checkIndices {
		original := &dataset.Players[idx]
		convertedPlayer := &converted.Players[idx]

		if convertedPlayer.UID != original.UID {
			t.Errorf("Large dataset player %d UID mismatch: got %v, want %v",
				idx, convertedPlayer.UID, original.UID)
		}
		if convertedPlayer.Name != original.Name {
			t.Errorf("Large dataset player %d Name mismatch: got %v, want %v",
				idx, convertedPlayer.Name, original.Name)
		}
		if len(convertedPlayer.RoleSpecificOveralls) != len(original.RoleSpecificOveralls) {
			t.Errorf("Large dataset player %d RoleSpecificOveralls length mismatch: got %v, want %v",
				idx, len(convertedPlayer.RoleSpecificOveralls), len(original.RoleSpecificOveralls))
		}
	}
}

// TestCompleteRoundTripIntegrity performs comprehensive round-trip testing
func TestCompleteRoundTripIntegrity(t *testing.T) {
	ctx := context.Background()

	// Create a comprehensive test player with all possible field types
	original := &Player{
		UID:                 123456789,
		Name:                "Comprehensive Test Player",
		Position:            "CAM",
		Age:                 "24",
		Club:                "Test FC United",
		Division:            "Premier League",
		TransferValue:       "£50M",
		Wage:                "£200K",
		Personality:         "Ambitious",
		MediaHandling:       "Evasive",
		Nationality:         "England",
		NationalityISO:      "ENG",
		NationalityFIFACode: "ENG",
		AttributeMasked:     true,
		Attributes: map[string]string{
			"Corners":          "15",
			"Crossing":         "14",
			"Dribbling":        "18",
			"Finishing":        "16",
			"First Touch":      "17",
			"Free Kick Taking": "13",
			"Heading":          "12",
			"Long Shots":       "15",
			"Long Throws":      "8",
			"Marking":          "10",
			"Passing":          "19",
			"Penalty Taking":   "14",
			"Tackling":         "9",
			"Technique":        "18",
		},
		NumericAttributes: map[string]int{
			"Corners":          15,
			"Crossing":         14,
			"Dribbling":        18,
			"Finishing":        16,
			"First Touch":      17,
			"Free Kick Taking": 13,
			"Heading":          12,
			"Long Shots":       15,
			"Long Throws":      8,
			"Marking":          10,
			"Passing":          19,
			"Penalty Taking":   14,
			"Tackling":         9,
			"Technique":        18,
		},
		PerformanceStatsNumeric: map[string]float64{
			"Goals":             12.0,
			"Assists":           18.0,
			"Key Passes":        3.2,
			"Shots per Game":    2.8,
			"Pass Completion":   89.5,
			"Dribbles per Game": 4.1,
			"Tackles per Game":  1.2,
			"Interceptions":     0.8,
		},
		PerformancePercentiles: map[string]map[string]float64{
			"Attacking": {
				"Goals":            85.2,
				"Assists":          92.7,
				"Key Passes":       88.1,
				"Shots per Game":   79.3,
				"Expected Goals":   82.6,
				"Expected Assists": 90.4,
			},
			"Passing": {
				"Pass Completion":    94.1,
				"Long Pass Accuracy": 87.3,
				"Through Balls":      91.8,
				"Crosses":            76.2,
			},
			"Dribbling": {
				"Dribbles per Game":    89.7,
				"Dribble Success Rate": 85.4,
				"Take-ons":             88.2,
			},
			"Defending": {
				"Tackles per Game": 45.2,
				"Interceptions":    38.7,
				"Clearances":       25.1,
			},
		},
		ParsedPositions: []string{"CAM", "CM", "AMR", "AML"},
		ShortPositions:  []string{"CAM", "CM"},
		PositionGroups:  []string{"Midfielder", "Attacking Midfielder"},
		PAC:             82,
		SHO:             78,
		PAS:             91,
		DRI:             88,
		DEF:             45,
		PHY:             72,
		GK:              15,
		DIV:             12,
		HAN:             8,
		REF:             10,
		KIC:             65,
		SPD:             85,
		POS:             88,
		Overall:         84,
		BestRoleOverall: "Advanced Playmaker (Support)",
		RoleSpecificOveralls: []RoleOverallScore{
			{RoleName: "Advanced Playmaker (Support)", Score: 89},
			{RoleName: "Advanced Playmaker (Attack)", Score: 87},
			{RoleName: "Attacking Midfielder (Support)", Score: 86},
			{RoleName: "Attacking Midfielder (Attack)", Score: 85},
			{RoleName: "Central Midfielder (Support)", Score: 82},
			{RoleName: "Central Midfielder (Attack)", Score: 80},
			{RoleName: "Wide Midfielder (Support)", Score: 78},
			{RoleName: "Wide Midfielder (Attack)", Score: 76},
		},
		TransferValueAmount: 50000000,
		WageAmount:          200000,
	}

	// Convert to protobuf
	protoPlayer, err := original.ToProto(ctx)
	if err != nil {
		t.Fatalf("Failed to convert comprehensive player to protobuf: %v", err)
	}

	// Convert back from protobuf
	converted, err := PlayerFromProto(ctx, protoPlayer)
	if err != nil {
		t.Fatalf("Failed to convert comprehensive protobuf player back: %v", err)
	}

	// Comprehensive field verification
	if converted.UID != original.UID {
		t.Errorf("UID mismatch: got %v, want %v", converted.UID, original.UID)
	}
	if converted.Name != original.Name {
		t.Errorf("Name mismatch: got %v, want %v", converted.Name, original.Name)
	}
	if converted.Position != original.Position {
		t.Errorf("Position mismatch: got %v, want %v", converted.Position, original.Position)
	}
	if converted.AttributeMasked != original.AttributeMasked {
		t.Errorf("AttributeMasked mismatch: got %v, want %v", converted.AttributeMasked, original.AttributeMasked)
	}

	// Verify all attributes
	if !reflect.DeepEqual(converted.Attributes, original.Attributes) {
		t.Errorf("Attributes mismatch: got %v, want %v", converted.Attributes, original.Attributes)
	}
	if !reflect.DeepEqual(converted.NumericAttributes, original.NumericAttributes) {
		t.Errorf("NumericAttributes mismatch: got %v, want %v", converted.NumericAttributes, original.NumericAttributes)
	}
	if !reflect.DeepEqual(converted.PerformanceStatsNumeric, original.PerformanceStatsNumeric) {
		t.Errorf("PerformanceStatsNumeric mismatch: got %v, want %v", converted.PerformanceStatsNumeric, original.PerformanceStatsNumeric)
	}

	// Verify nested performance percentiles
	if !reflect.DeepEqual(converted.PerformancePercentiles, original.PerformancePercentiles) {
		t.Errorf("PerformancePercentiles mismatch: got %v, want %v", converted.PerformancePercentiles, original.PerformancePercentiles)
	}

	// Verify slices
	if !reflect.DeepEqual(converted.ParsedPositions, original.ParsedPositions) {
		t.Errorf("ParsedPositions mismatch: got %v, want %v", converted.ParsedPositions, original.ParsedPositions)
	}

	// Verify all individual stats
	statsToCheck := []struct {
		name      string
		got, want int
	}{
		{"PAC", converted.PAC, original.PAC},
		{"SHO", converted.SHO, original.SHO},
		{"PAS", converted.PAS, original.PAS},
		{"DRI", converted.DRI, original.DRI},
		{"DEF", converted.DEF, original.DEF},
		{"PHY", converted.PHY, original.PHY},
		{"Overall", converted.Overall, original.Overall},
	}

	for _, stat := range statsToCheck {
		if stat.got != stat.want {
			t.Errorf("%s mismatch: got %v, want %v", stat.name, stat.got, stat.want)
		}
	}

	// Verify role-specific overalls
	if len(converted.RoleSpecificOveralls) != len(original.RoleSpecificOveralls) {
		t.Errorf("RoleSpecificOveralls length mismatch: got %v, want %v",
			len(converted.RoleSpecificOveralls), len(original.RoleSpecificOveralls))
	}

	for i, role := range converted.RoleSpecificOveralls {
		if i >= len(original.RoleSpecificOveralls) {
			break
		}
		originalRole := original.RoleSpecificOveralls[i]
		if role.RoleName != originalRole.RoleName {
			t.Errorf("Role %d name mismatch: got %v, want %v", i, role.RoleName, originalRole.RoleName)
		}
		if role.Score != originalRole.Score {
			t.Errorf("Role %d score mismatch: got %v, want %v", i, role.Score, originalRole.Score)
		}
	}
}
