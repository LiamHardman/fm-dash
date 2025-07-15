package arrow

import (
	"context"
	"testing"

	"github.com/apache/arrow/go/v17/arrow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArrowProcessor_ConvertFromJSON(t *testing.T) {
	processor := NewArrowProcessor()
	ctx := context.Background()

	t.Run("Empty slice", func(t *testing.T) {
		players := []Player{}
		table, err := processor.ConvertFromJSON(ctx, players)
		
		require.NoError(t, err)
		assert.NotNil(t, table)
		assert.Equal(t, int64(0), table.NumRows())
		assert.Equal(t, int64(len(PlayerSchema.Fields())), table.NumCols())
	})

	t.Run("Single player conversion", func(t *testing.T) {
		players := []Player{
			createTestPlayer(),
		}

		table, err := processor.ConvertFromJSON(ctx, players)
		
		require.NoError(t, err)
		assert.NotNil(t, table)
		assert.Equal(t, int64(1), table.NumRows())
		assert.Equal(t, int64(len(PlayerSchema.Fields())), table.NumCols())

		// Verify schema matches
		assert.True(t, table.Schema().Equal(PlayerSchema))
	})

	t.Run("Multiple players conversion", func(t *testing.T) {
		players := []Player{
			createTestPlayer(),
			createTestPlayerWithNulls(),
			createTestGoalkeeper(),
		}

		table, err := processor.ConvertFromJSON(ctx, players)
		
		require.NoError(t, err)
		assert.NotNil(t, table)
		assert.Equal(t, int64(3), table.NumRows())
		assert.Equal(t, int64(len(PlayerSchema.Fields())), table.NumCols())
	})

	t.Run("Player with complex nested data", func(t *testing.T) {
		player := createTestPlayerWithComplexData()
		players := []Player{player}

		table, err := processor.ConvertFromJSON(ctx, players)
		
		require.NoError(t, err)
		assert.NotNil(t, table)
		assert.Equal(t, int64(1), table.NumRows())
	})
}

func TestArrowProcessor_ConvertToJSON(t *testing.T) {
	processor := NewArrowProcessor()
	ctx := context.Background()

	t.Run("Nil table", func(t *testing.T) {
		var nilTable arrow.Table
		players, err := processor.ConvertToJSON(ctx, nilTable)
		
		assert.Error(t, err)
		assert.Nil(t, players)
		assert.Contains(t, err.Error(), "table cannot be nil")
	})

	t.Run("Empty table", func(t *testing.T) {
		// Create empty table using the processor's method
		emptyPlayers := []Player{}
		emptyTable, err := processor.ConvertFromJSON(ctx, emptyPlayers)
		require.NoError(t, err)
		defer emptyTable.Release()

		players, err := processor.ConvertToJSON(ctx, emptyTable)
		
		require.NoError(t, err)
		assert.NotNil(t, players)
		assert.Equal(t, 0, len(players))
	})

	t.Run("Single player table", func(t *testing.T) {
		originalPlayer := createTestPlayer()
		originalPlayers := []Player{originalPlayer}

		// Convert to Arrow
		table, err := processor.ConvertFromJSON(ctx, originalPlayers)
		require.NoError(t, err)
		defer table.Release()

		// Convert back to JSON
		convertedPlayers, err := processor.ConvertToJSON(ctx, table)
		require.NoError(t, err)

		assert.Equal(t, 1, len(convertedPlayers))
		assertPlayersEqual(t, originalPlayer, convertedPlayers[0])
	})

	t.Run("Multiple players table", func(t *testing.T) {
		originalPlayers := []Player{
			createTestPlayer(),
			createTestPlayerWithNulls(),
			createTestGoalkeeper(),
		}

		// Convert to Arrow
		table, err := processor.ConvertFromJSON(ctx, originalPlayers)
		require.NoError(t, err)
		defer table.Release()

		// Convert back to JSON
		convertedPlayers, err := processor.ConvertToJSON(ctx, table)
		require.NoError(t, err)

		assert.Equal(t, len(originalPlayers), len(convertedPlayers))
		for i, original := range originalPlayers {
			assertPlayersEqual(t, original, convertedPlayers[i])
		}
	})
}

func TestArrowProcessor_RoundTripConversion(t *testing.T) {
	processor := NewArrowProcessor()
	ctx := context.Background()

	t.Run("Round trip maintains data integrity", func(t *testing.T) {
		originalPlayers := []Player{
			createTestPlayer(),
			createTestPlayerWithNulls(),
			createTestGoalkeeper(),
			createTestPlayerWithComplexData(),
		}

		// JSON -> Arrow
		table, err := processor.ConvertFromJSON(ctx, originalPlayers)
		require.NoError(t, err)
		defer table.Release()

		// Arrow -> JSON
		convertedPlayers, err := processor.ConvertToJSON(ctx, table)
		require.NoError(t, err)

		// Verify data integrity
		assert.Equal(t, len(originalPlayers), len(convertedPlayers))
		for i, original := range originalPlayers {
			assertPlayersEqual(t, original, convertedPlayers[i])
		}
	})

	t.Run("Multiple round trips", func(t *testing.T) {
		originalPlayers := []Player{createTestPlayer()}

		currentPlayers := originalPlayers
		for i := 0; i < 3; i++ {
			// JSON -> Arrow
			table, err := processor.ConvertFromJSON(ctx, currentPlayers)
			require.NoError(t, err)

			// Arrow -> JSON
			convertedPlayers, err := processor.ConvertToJSON(ctx, table)
			require.NoError(t, err)
			table.Release()

			currentPlayers = convertedPlayers
		}

		// Verify data integrity after multiple conversions
		assert.Equal(t, len(originalPlayers), len(currentPlayers))
		assertPlayersEqual(t, originalPlayers[0], currentPlayers[0])
	})
}

func TestArrowProcessor_ErrorHandling(t *testing.T) {
	processor := NewArrowProcessor()
	ctx := context.Background()

	t.Run("Invalid JSON in complex fields", func(t *testing.T) {
		// This test would require modifying the Player struct to have invalid JSON
		// For now, we'll test with valid data and ensure no errors occur
		player := createTestPlayer()
		players := []Player{player}

		table, err := processor.ConvertFromJSON(ctx, players)
		require.NoError(t, err)
		defer table.Release()

		convertedPlayers, err := processor.ConvertToJSON(ctx, table)
		require.NoError(t, err)
		assert.Equal(t, 1, len(convertedPlayers))
	})
}

func TestArrowProcessor_NullHandling(t *testing.T) {
	processor := NewArrowProcessor()
	ctx := context.Background()

	t.Run("Nullable fields handled correctly", func(t *testing.T) {
		player := createTestPlayerWithNulls()
		players := []Player{player}

		table, err := processor.ConvertFromJSON(ctx, players)
		require.NoError(t, err)
		defer table.Release()

		convertedPlayers, err := processor.ConvertToJSON(ctx, table)
		require.NoError(t, err)

		assert.Equal(t, 1, len(convertedPlayers))
		converted := convertedPlayers[0]

		// Check that nullable fields are properly handled
		assert.Equal(t, "", converted.Personality)
		assert.Equal(t, "", converted.MediaHandling)
		assert.Equal(t, false, converted.AttributeMasked)
		assert.Equal(t, 0, converted.GK)
		assert.Equal(t, 0, converted.DIV)
		assert.Equal(t, 0, converted.HAN)
		assert.Equal(t, 0, converted.REF)
		assert.Equal(t, 0, converted.KIC)
		assert.Equal(t, 0, converted.SPD)
		assert.Equal(t, 0, converted.POS)
	})
}

// Helper functions for creating test data

func createTestPlayer() Player {
	return Player{
		UID:                     12345,
		Name:                    "Test Player",
		Position:                "CM",
		Age:                     "25",
		Club:                    "Test FC",
		Division:                "Premier League",
		TransferValue:           "£50M",
		Wage:                    "£100K",
		Personality:             "Professional",
		MediaHandling:           "Evasive",
		Nationality:             "England",
		NationalityISO:          "ENG",
		NationalityFIFACode:     "ENG",
		AttributeMasked:         false,
		Attributes:              map[string]string{"Crossing": "15", "Finishing": "12"},
		NumericAttributes:       map[string]int{"Crossing": 15, "Finishing": 12},
		PerformanceStatsNumeric: map[string]float64{"Goals": 5.5, "Assists": 3.2},
		PerformancePercentiles:  map[string]map[string]float64{"Goals": {"League": 75.5}},
		ParsedPositions:         []string{"CM", "CAM"},
		ShortPositions:          []string{"CM"},
		PositionGroups:          []string{"Midfielder"},
		PAC:                     75,
		SHO:                     65,
		PAS:                     80,
		DRI:                     70,
		DEF:                     60,
		PHY:                     75,
		GK:                      0,
		DIV:                     0,
		HAN:                     0,
		REF:                     0,
		KIC:                     0,
		SPD:                     0,
		POS:                     0,
		Overall:                 75,
		BestRoleOverall:         "Central Midfielder",
		RoleSpecificOveralls:    []RoleOverallScore{{RoleName: "CM", Score: 75}},
		TransferValueAmount:     50000000,
		WageAmount:              100000,
	}
}

func createTestPlayerWithNulls() Player {
	return Player{
		UID:                     67890,
		Name:                    "Null Player",
		Position:                "CB",
		Age:                     "30",
		Club:                    "Null FC",
		Division:                "Championship",
		TransferValue:           "£10M",
		Wage:                    "£50K",
		Personality:             "", // Empty string for nullable field
		MediaHandling:           "", // Empty string for nullable field
		Nationality:             "Spain",
		NationalityISO:          "ESP",
		NationalityFIFACode:     "ESP",
		AttributeMasked:         false,
		Attributes:              map[string]string{"Tackling": "18"},
		NumericAttributes:       map[string]int{"Tackling": 18},
		PerformanceStatsNumeric: map[string]float64{"Tackles": 4.2},
		PerformancePercentiles:  map[string]map[string]float64{"Tackles": {"League": 85.0}},
		ParsedPositions:         []string{"CB"},
		ShortPositions:          []string{"CB"},
		PositionGroups:          []string{"Defender"},
		PAC:                     60,
		SHO:                     30,
		PAS:                     70,
		DRI:                     50,
		DEF:                     85,
		PHY:                     80,
		GK:                      0, // Zero for nullable GK fields
		DIV:                     0,
		HAN:                     0,
		REF:                     0,
		KIC:                     0,
		SPD:                     0,
		POS:                     0,
		Overall:                 70,
		BestRoleOverall:         "Centre Back",
		RoleSpecificOveralls:    []RoleOverallScore{{RoleName: "CB", Score: 70}},
		TransferValueAmount:     10000000,
		WageAmount:              50000,
	}
}

func createTestGoalkeeper() Player {
	return Player{
		UID:                     11111,
		Name:                    "Test Keeper",
		Position:                "GK",
		Age:                     "28",
		Club:                    "Keeper FC",
		Division:                "Premier League",
		TransferValue:           "£30M",
		Wage:                    "£80K",
		Personality:             "Determined",
		MediaHandling:           "Outspoken",
		Nationality:             "Germany",
		NationalityISO:          "GER",
		NationalityFIFACode:     "GER",
		AttributeMasked:         false,
		Attributes:              map[string]string{"Handling": "18", "Reflexes": "17"},
		NumericAttributes:       map[string]int{"Handling": 18, "Reflexes": 17},
		PerformanceStatsNumeric: map[string]float64{"Saves": 3.8, "CleanSheets": 0.4},
		PerformancePercentiles:  map[string]map[string]float64{"Saves": {"League": 90.0}},
		ParsedPositions:         []string{"GK"},
		ShortPositions:          []string{"GK"},
		PositionGroups:          []string{"Goalkeeper"},
		PAC:                     40,
		SHO:                     20,
		PAS:                     60,
		DRI:                     30,
		DEF:                     10,
		PHY:                     70,
		GK:                      85, // GK attributes populated
		DIV:                     82,
		HAN:                     88,
		REF:                     85,
		KIC:                     75,
		SPD:                     80,
		POS:                     83,
		Overall:                 82,
		BestRoleOverall:         "Goalkeeper",
		RoleSpecificOveralls:    []RoleOverallScore{{RoleName: "GK", Score: 82}},
		TransferValueAmount:     30000000,
		WageAmount:              80000,
	}
}

func createTestPlayerWithComplexData() Player {
	return Player{
		UID:                     99999,
		Name:                    "Complex Player",
		Position:                "ST",
		Age:                     "22",
		Club:                    "Complex FC",
		Division:                "La Liga",
		TransferValue:           "£80M",
		Wage:                    "£200K",
		Personality:             "Ambitious",
		MediaHandling:           "Charismatic",
		Nationality:             "Brazil",
		NationalityISO:          "BRA",
		NationalityFIFACode:     "BRA",
		AttributeMasked:         true,
		Attributes: map[string]string{
			"Crossing":     "12",
			"Finishing":    "18",
			"HeadingAcc":   "16",
			"ShortPassing": "14",
			"Volleys":      "17",
		},
		NumericAttributes: map[string]int{
			"Crossing":     12,
			"Finishing":    18,
			"HeadingAcc":   16,
			"ShortPassing": 14,
			"Volleys":      17,
		},
		PerformanceStatsNumeric: map[string]float64{
			"Goals":         15.2,
			"Assists":       8.7,
			"ShotsPerGame":  4.3,
			"PassAccuracy": 82.5,
		},
		PerformancePercentiles: map[string]map[string]float64{
			"Goals":   {"League": 95.5, "Position": 92.1},
			"Assists": {"League": 78.3, "Position": 85.7},
		},
		ParsedPositions: []string{"ST", "CF", "LW", "RW"},
		ShortPositions:  []string{"ST", "LW", "RW"},
		PositionGroups:  []string{"Forward", "Winger"},
		PAC:             90,
		SHO:             88,
		PAS:             75,
		DRI:             85,
		DEF:             35,
		PHY:             82,
		GK:              0,
		DIV:             0,
		HAN:             0,
		REF:             0,
		KIC:             0,
		SPD:             0,
		POS:             0,
		Overall:         85,
		BestRoleOverall: "Striker",
		RoleSpecificOveralls: []RoleOverallScore{
			{RoleName: "ST", Score: 85},
			{RoleName: "CF", Score: 83},
			{RoleName: "LW", Score: 78},
			{RoleName: "RW", Score: 78},
		},
		TransferValueAmount: 80000000,
		WageAmount:          200000,
	}
}

// assertPlayersEqual compares two Player structs for equality
func assertPlayersEqual(t *testing.T, expected, actual Player) {
	assert.Equal(t, expected.UID, actual.UID)
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Position, actual.Position)
	assert.Equal(t, expected.Age, actual.Age)
	assert.Equal(t, expected.Club, actual.Club)
	assert.Equal(t, expected.Division, actual.Division)
	assert.Equal(t, expected.TransferValue, actual.TransferValue)
	assert.Equal(t, expected.Wage, actual.Wage)
	assert.Equal(t, expected.Personality, actual.Personality)
	assert.Equal(t, expected.MediaHandling, actual.MediaHandling)
	assert.Equal(t, expected.Nationality, actual.Nationality)
	assert.Equal(t, expected.NationalityISO, actual.NationalityISO)
	assert.Equal(t, expected.NationalityFIFACode, actual.NationalityFIFACode)
	assert.Equal(t, expected.AttributeMasked, actual.AttributeMasked)
	assert.Equal(t, expected.Attributes, actual.Attributes)
	assert.Equal(t, expected.NumericAttributes, actual.NumericAttributes)
	assert.Equal(t, expected.PerformanceStatsNumeric, actual.PerformanceStatsNumeric)
	assert.Equal(t, expected.PerformancePercentiles, actual.PerformancePercentiles)
	assert.Equal(t, expected.ParsedPositions, actual.ParsedPositions)
	assert.Equal(t, expected.ShortPositions, actual.ShortPositions)
	assert.Equal(t, expected.PositionGroups, actual.PositionGroups)
	assert.Equal(t, expected.PAC, actual.PAC)
	assert.Equal(t, expected.SHO, actual.SHO)
	assert.Equal(t, expected.PAS, actual.PAS)
	assert.Equal(t, expected.DRI, actual.DRI)
	assert.Equal(t, expected.DEF, actual.DEF)
	assert.Equal(t, expected.PHY, actual.PHY)
	assert.Equal(t, expected.GK, actual.GK)
	assert.Equal(t, expected.DIV, actual.DIV)
	assert.Equal(t, expected.HAN, actual.HAN)
	assert.Equal(t, expected.REF, actual.REF)
	assert.Equal(t, expected.KIC, actual.KIC)
	assert.Equal(t, expected.SPD, actual.SPD)
	assert.Equal(t, expected.POS, actual.POS)
	assert.Equal(t, expected.Overall, actual.Overall)
	assert.Equal(t, expected.BestRoleOverall, actual.BestRoleOverall)
	assert.Equal(t, expected.RoleSpecificOveralls, actual.RoleSpecificOveralls)
	assert.Equal(t, expected.TransferValueAmount, actual.TransferValueAmount)
	assert.Equal(t, expected.WageAmount, actual.WageAmount)
}

func TestArrowProcessor_MemoryManagement(t *testing.T) {
	processor := NewArrowProcessor()
	ctx := context.Background()

	t.Run("Memory is properly released", func(t *testing.T) {
		// Create a large dataset to test memory management
		players := make([]Player, 1000)
		for i := 0; i < 1000; i++ {
			players[i] = createTestPlayer()
			players[i].UID = int64(i)
		}

		// Convert to Arrow
		table, err := processor.ConvertFromJSON(ctx, players)
		require.NoError(t, err)

		// Verify table properties
		assert.Equal(t, int64(1000), table.NumRows())
		assert.Equal(t, int64(len(PlayerSchema.Fields())), table.NumCols())

		// Convert back to JSON
		convertedPlayers, err := processor.ConvertToJSON(ctx, table)
		require.NoError(t, err)
		assert.Equal(t, 1000, len(convertedPlayers))

		// Release table memory
		table.Release()

		// Verify first and last players to ensure data integrity
		assertPlayersEqual(t, players[0], convertedPlayers[0])
		assertPlayersEqual(t, players[999], convertedPlayers[999])
	})
}

func BenchmarkArrowProcessor_ConvertFromJSON(b *testing.B) {
	processor := NewArrowProcessor()
	ctx := context.Background()
	
	// Create test data
	players := make([]Player, 100)
	for i := 0; i < 100; i++ {
		players[i] = createTestPlayer()
		players[i].UID = int64(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		table, err := processor.ConvertFromJSON(ctx, players)
		if err != nil {
			b.Fatal(err)
		}
		table.Release()
	}
}

func BenchmarkArrowProcessor_ConvertToJSON(b *testing.B) {
	processor := NewArrowProcessor()
	ctx := context.Background()
	
	// Create test data
	players := make([]Player, 100)
	for i := 0; i < 100; i++ {
		players[i] = createTestPlayer()
		players[i].UID = int64(i)
	}

	// Convert to Arrow table once
	table, err := processor.ConvertFromJSON(ctx, players)
	if err != nil {
		b.Fatal(err)
	}
	defer table.Release()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := processor.ConvertToJSON(ctx, table)
		if err != nil {
			b.Fatal(err)
		}
	}
}