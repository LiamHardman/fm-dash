package main

import (
	"reflect"
	"sort"
	"testing"
)

func TestGetShortPositionKeyFromRoleName(t *testing.T) {
	tests := []struct {
		name         string
		roleFullName string
		expected     string
	}{
		{
			name:         "standard role format",
			roleFullName: "DC - Central Defender - Defend",
			expected:     "DC",
		},
		{
			name:         "role with single dash",
			roleFullName: "GK - Goalkeeper",
			expected:     "GK",
		},
		{
			name:         "role without dashes",
			roleFullName: "ST",
			expected:     "ST",
		},
		{
			name:         "empty string",
			roleFullName: "",
			expected:     "",
		},
		{
			name:         "role with spaces around key",
			roleFullName: " MC - Central Midfielder - Support",
			expected:     "MC",
		},
		{
			name:         "complex role name",
			roleFullName: "AMC - Attacking Midfielder Centre - Attack",
			expected:     "AMC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetShortPositionKeyFromRoleName(tt.roleFullName)
			if result != tt.expected {
				t.Errorf("GetShortPositionKeyFromRoleName(%s) = %s; want %s", tt.roleFullName, result, tt.expected)
			}
		})
	}
}

func TestParsePlayerPositionsGo(t *testing.T) {
	tests := []struct {
		name        string
		positionStr string
		expected    []string
	}{
		{
			name:        "empty string",
			positionStr: "",
			expected:    []string{},
		},
		{
			name:        "single goalkeeper position",
			positionStr: "GK",
			expected:    []string{"Goalkeeper"},
		},
		{
			name:        "defender with explicit sides",
			positionStr: "D (RLC)",
			expected:    []string{"Right Back", "Left Back", "Centre Back"},
		},
		{
			name:        "midfielder centre",
			positionStr: "M (C)",
			expected:    []string{"Centre Midfielder"},
		},
		{
			name:        "attacking midfielder right",
			positionStr: "AM (R)",
			expected:    []string{"Right Attacking Midfielder"},
		},
		{
			name:        "striker centre",
			positionStr: "ST (C)",
			expected:    []string{"Striker"},
		},
		{
			name:        "wing-back right",
			positionStr: "WB (R)",
			expected:    []string{"Right Wing-Back"},
		},
		{
			name:        "defensive midfielder",
			positionStr: "DM",
			expected:    []string{"Centre Defensive Midfielder"},
		},
		{
			name:        "multiple positions with comma",
			positionStr: "D (C), M (C)",
			expected:    []string{"Centre Back", "Centre Midfielder"},
		},
		{
			name:        "complex multi-position",
			positionStr: "D/M (RLC), AM (C), ST (C)",
			expected:    []string{"Centre Attacking Midfielder", "Centre Back", "Centre Midfielder", "Left Back", "Left Midfielder", "Right Back", "Right Midfielder", "Striker"},
		},
		{
			name:        "sweeper",
			positionStr: "SW",
			expected:    []string{"Sweeper"},
		},
		{
			name:        "wing-back both sides",
			positionStr: "WB (RL)",
			expected:    []string{"Left Wing-Back", "Right Wing-Back"},
		},
		{
			name:        "malformed position - no closing bracket",
			positionStr: "D (R",
			expected:    []string{},
		},
		{
			name:        "position with extra spaces",
			positionStr: " D (C) , M (R) ",
			expected:    []string{"Centre Back", "Right Midfielder"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParsePlayerPositionsGo(tt.positionStr)

			// Sort both slices for comparison since order might vary
			sort.Strings(result)
			sort.Strings(tt.expected)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParsePlayerPositionsGo(%s) = %v; want %v", tt.positionStr, result, tt.expected)
			}
		})
	}
}

func TestGetPlayerPositionGroupsGo(t *testing.T) {
	tests := []struct {
		name                 string
		parsedPositionsArray []string
		expected             []string
	}{
		{
			name:                 "empty positions",
			parsedPositionsArray: []string{},
			expected:             []string{},
		},
		{
			name:                 "goalkeeper only",
			parsedPositionsArray: []string{"Goalkeeper"},
			expected:             []string{"Goalkeepers"},
		},
		{
			name:                 "defenders only",
			parsedPositionsArray: []string{"Centre Back", "Right Back"},
			expected:             []string{"Defenders"},
		},
		{
			name:                 "midfielders only",
			parsedPositionsArray: []string{"Centre Midfielder", "Right Midfielder"},
			expected:             []string{"Midfielders"},
		},
		{
			name:                 "attackers only",
			parsedPositionsArray: []string{"Striker"},
			expected:             []string{"Attackers"},
		},
		{
			name:                 "wing-backs only",
			parsedPositionsArray: []string{"Right Wing-Back", "Left Wing-Back"},
			expected:             []string{"Wing-Backs"},
		},
		{
			name:                 "mixed positions",
			parsedPositionsArray: []string{"Centre Back", "Centre Midfielder", "Striker"},
			expected:             []string{"Attackers", "Defenders", "Midfielders"},
		},
		{
			name:                 "all position groups",
			parsedPositionsArray: []string{"Goalkeeper", "Centre Back", "Right Wing-Back", "Centre Midfielder", "Striker"},
			expected:             []string{"Attackers", "Defenders", "Goalkeepers", "Midfielders", "Wing-Backs"},
		},
		{
			name:                 "unknown position",
			parsedPositionsArray: []string{"Unknown Position"},
			expected:             []string{},
		},
		{
			name:                 "mixed known and unknown positions",
			parsedPositionsArray: []string{"Centre Back", "Unknown Position", "Striker"},
			expected:             []string{"Attackers", "Defenders"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPlayerPositionGroupsGo(tt.parsedPositionsArray)

			// Sort both slices for comparison
			sort.Strings(result)
			sort.Strings(tt.expected)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("GetPlayerPositionGroupsGo(%v) = %v; want %v", tt.parsedPositionsArray, result, tt.expected)
			}
		})
	}
}

func TestPositionMappings(t *testing.T) {
	// Test that all position mappings are consistent
	t.Run("position role map completeness", func(t *testing.T) {
		// Verify that all keys in positionRoleMapGo have corresponding entries in standardizedPositionNameMapGo
		for shortCode, roleName := range positionRoleMapGo {
			if _, exists := standardizedPositionNameMapGo[roleName]; !exists {
				t.Errorf("Role name '%s' for short code '%s' not found in standardizedPositionNameMapGo", roleName, shortCode)
			}
		}
	})

	t.Run("standardized position name map completeness", func(t *testing.T) {
		// Verify that all standardized position names are covered in position groups
		allGroupPositions := make(map[string]bool)
		for _, positions := range positionGroupsGo {
			for _, pos := range positions {
				allGroupPositions[pos] = true
			}
		}

		for _, standardizedName := range standardizedPositionNameMapGo {
			if !allGroupPositions[standardizedName] {
				t.Errorf("Standardized position name '%s' not found in any position group", standardizedName)
			}
		}
	})

	t.Run("parsed position to base role key completeness", func(t *testing.T) {
		// Verify that all standardized position names have corresponding base role keys
		for _, standardizedName := range standardizedPositionNameMapGo {
			if _, exists := parsedPositionToBaseRoleKeyGo[standardizedName]; !exists {
				t.Errorf("Standardized position name '%s' not found in parsedPositionToBaseRoleKeyGo", standardizedName)
			}
		}
	})

	t.Run("short position display order completeness", func(t *testing.T) {
		// Verify that all short codes in positionRoleMapGo are in the display order
		for shortCode := range positionRoleMapGo {
			if _, exists := ShortPositionOrderMap[shortCode]; !exists {
				t.Errorf("Short code '%s' not found in ShortPositionDisplayOrder", shortCode)
			}
		}
	})
}

func TestShortPositionOrderMap(t *testing.T) {
	// Test that the order map is correctly generated
	for i, pos := range ShortPositionDisplayOrder {
		if ShortPositionOrderMap[pos] != i {
			t.Errorf("ShortPositionOrderMap[%s] = %d; want %d", pos, ShortPositionOrderMap[pos], i)
		}
	}
}

// Benchmark test for position parsing performance
func BenchmarkParsePlayerPositionsGo(b *testing.B) {
	testPosition := "D/M (RLC), AM (C), ST (C)"

	for i := 0; i < b.N; i++ {
		ParsePlayerPositionsGo(testPosition)
	}
}

func BenchmarkGetPlayerPositionGroupsGo(b *testing.B) {
	testPositions := []string{"Centre Back", "Centre Midfielder", "Striker", "Right Wing-Back"}

	for i := 0; i < b.N; i++ {
		GetPlayerPositionGroupsGo(testPositions)
	}
}
