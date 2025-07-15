package main

import (
	"testing"

	"github.com/apache/arrow/go/v18/arrow/memory"
)

func TestNewArrowDatasetManager(t *testing.T) {
	allocator := memory.NewGoAllocator()
	manager := NewArrowDatasetManager(allocator)
	
	if manager == nil {
		t.Fatal("Expected manager to be created, got nil")
	}
	
	if manager.schema == nil {
		t.Error("Expected schema to be initialized")
	}
	
	if manager.allocator != allocator {
		t.Error("Expected allocator to be set correctly")
	}
	
	if manager.stringInterner == nil {
		t.Error("Expected string interner to be initialized")
	}
	
	if manager.recordBatchSize != 10000 {
		t.Errorf("Expected default record batch size 10000, got %d", manager.recordBatchSize)
	}
}

func TestSetRecordBatchSize(t *testing.T) {
	allocator := memory.NewGoAllocator()
	manager := NewArrowDatasetManager(allocator)
	
	manager.SetRecordBatchSize(5000)
	
	if manager.recordBatchSize != 5000 {
		t.Errorf("Expected record batch size 5000, got %d", manager.recordBatchSize)
	}
}

func TestStringInterner(t *testing.T) {
	interner := NewStringInterner()
	
	if interner == nil {
		t.Fatal("Expected interner to be created, got nil")
	}
	
	// Test position interning
	pos1 := interner.InternPosition("GK")
	pos2 := interner.InternPosition("CB")
	pos3 := interner.InternPosition("GK") // Should return same ID
	
	if pos1 == pos2 {
		t.Error("Expected different IDs for different positions")
	}
	
	if pos1 != pos3 {
		t.Error("Expected same ID for same position")
	}
	
	// Test club interning
	club1 := interner.InternClub("Manchester United")
	club2 := interner.InternClub("Liverpool")
	club3 := interner.InternClub("Manchester United") // Should return same ID
	
	if club1 == club2 {
		t.Error("Expected different IDs for different clubs")
	}
	
	if club1 != club3 {
		t.Error("Expected same ID for same club")
	}
	
	// Test nationality interning
	nat1 := interner.InternNationality("England")
	nat2 := interner.InternNationality("Spain")
	nat3 := interner.InternNationality("England") // Should return same ID
	
	if nat1 == nat2 {
		t.Error("Expected different IDs for different nationalities")
	}
	
	if nat1 != nat3 {
		t.Error("Expected same ID for same nationality")
	}
	
	// Test division interning
	div1 := interner.InternDivision("Premier League")
	div2 := interner.InternDivision("Championship")
	div3 := interner.InternDivision("Premier League") // Should return same ID
	
	if div1 == div2 {
		t.Error("Expected different IDs for different divisions")
	}
	
	if div1 != div3 {
		t.Error("Expected same ID for same division")
	}
	
	// Test name interning
	name1 := interner.InternName("John Smith")
	name2 := interner.InternName("Jane Doe")
	name3 := interner.InternName("John Smith") // Should return same ID
	
	if name1 == name2 {
		t.Error("Expected different IDs for different names")
	}
	
	if name1 != name3 {
		t.Error("Expected same ID for same name")
	}
}

func TestStringInternerDictionaries(t *testing.T) {
	interner := NewStringInterner()
	
	// Add some test data
	interner.InternPosition("GK")
	interner.InternPosition("CB")
	interner.InternClub("Manchester United")
	interner.InternNationality("England")
	interner.InternDivision("Premier League")
	interner.InternName("John Smith")
	
	// Test dictionaries
	posDict := interner.GetPositionDictionary()
	if len(posDict) != 2 {
		t.Errorf("Expected 2 positions in dictionary, got %d", len(posDict))
	}
	
	clubDict := interner.GetClubDictionary()
	if len(clubDict) != 1 {
		t.Errorf("Expected 1 club in dictionary, got %d", len(clubDict))
	}
	
	natDict := interner.GetNationalityDictionary()
	if len(natDict) != 1 {
		t.Errorf("Expected 1 nationality in dictionary, got %d", len(natDict))
	}
	
	divDict := interner.GetDivisionDictionary()
	if len(divDict) != 1 {
		t.Errorf("Expected 1 division in dictionary, got %d", len(divDict))
	}
	
	nameDict := interner.GetNameDictionary()
	if len(nameDict) != 1 {
		t.Errorf("Expected 1 name in dictionary, got %d", len(nameDict))
	}
}

func TestParseAge(t *testing.T) {
	allocator := memory.NewGoAllocator()
	manager := NewArrowDatasetManager(allocator)
	
	testCases := []struct {
		input    string
		expected int
		hasError bool
	}{
		{"25", 25, false},
		{"25 years", 25, false},
		{"25y", 25, false},
		{"25 (25y 123d)", 25, false},
		{"", 0, true},
		{"abc", 0, true},
		{"30 years old", 30, false},
		{"22y 45d", 22, false},
	}
	
	for _, tc := range testCases {
		result, err := manager.parseAge(tc.input)
		
		if tc.hasError {
			if err == nil {
				t.Errorf("Expected error for input '%s', got nil", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("Expected no error for input '%s', got: %v", tc.input, err)
			}
			if result != tc.expected {
				t.Errorf("Expected age %d for input '%s', got %d", tc.expected, tc.input, result)
			}
		}
	}
}

func TestBuildArrowTableEmptyPlayers(t *testing.T) {
	allocator := memory.NewGoAllocator()
	manager := NewArrowDatasetManager(allocator)
	
	_, err := manager.BuildArrowTable([]Player{})
	if err == nil {
		t.Error("Expected error for empty players slice, got nil")
	}
}

func TestBuildArrowTableSinglePlayer(t *testing.T) {
	allocator := memory.NewGoAllocator()
	manager := NewArrowDatasetManager(allocator)
	
	player := Player{
		UID:                 12345,
		Name:                "Test Player",
		Position:            "GK",
		Age:                 "25",
		Club:                "Test Club",
		Division:            "Test Division",
		Nationality:         "England",
		NationalityISO:      "ENG",
		NationalityFIFACode: "ENG",
		Overall:             85,
		PAC:                 70,
		SHO:                 30,
		PAS:                 85,
		DRI:                 40,
		DEF:                 20,
		PHY:                 80,
		GK:                  90,
		DIV:                 85,
		HAN:                 88,
		REF:                 82,
		KIC:                 75,
		SPD:                 60,
		POS:                 85,
		TransferValueAmount: 50000000,
		WageAmount:          100000,
		AttributeMasked:     false,
		Personality:         "Determined",
		MediaHandling:       "Evasive",
		BestRoleOverall:     "Goalkeeper (Defend)",
		ParsedPositions:     []string{"GK"},
		ShortPositions:      []string{"GK"},
		PositionGroups:      []string{"Goalkeeper"},
		PerformanceStatsNumeric: map[string]float64{
			"goals":           0.1,
			"assists":         0.2,
			"pass_completion": 85.5,
		},
		RoleSpecificOveralls: []RoleOverallScore{
			{RoleName: "Goalkeeper (Defend)", Score: 85},
			{RoleName: "Sweeper Keeper (Support)", Score: 78},
		},
	}
	
	table, err := manager.BuildArrowTable([]Player{player})
	if err != nil {
		t.Fatalf("Expected no error building table, got: %v", err)
	}
	
	if table == nil {
		t.Fatal("Expected table to be created, got nil")
	}
	
	if table.NumRows() != 1 {
		t.Errorf("Expected 1 row, got %d", table.NumRows())
	}
	
	if table.NumCols() != int64(len(manager.schema.GetSchema().Fields())) {
		t.Errorf("Expected %d columns, got %d", len(manager.schema.GetSchema().Fields()), table.NumCols())
	}
	
	// Clean up
	table.Release()
}

func TestBuildArrowTableMultiplePlayers(t *testing.T) {
	allocator := memory.NewGoAllocator()
	manager := NewArrowDatasetManager(allocator)
	
	players := []Player{
		{
			UID:                 1,
			Name:                "Player 1",
			Position:            "GK",
			Age:                 "25",
			Club:                "Club A",
			Division:            "Division 1",
			Nationality:         "England",
			NationalityISO:      "ENG",
			NationalityFIFACode: "ENG",
			Overall:             85,
			PAC:                 70,
			SHO:                 30,
			PAS:                 85,
			DRI:                 40,
			DEF:                 20,
			PHY:                 80,
			TransferValueAmount: 50000000,
			WageAmount:          100000,
			AttributeMasked:     false,
			ParsedPositions:     []string{"GK"},
			ShortPositions:      []string{"GK"},
			PositionGroups:      []string{"Goalkeeper"},
			PerformanceStatsNumeric: map[string]float64{
				"goals":   0.1,
				"assists": 0.2,
			},
			RoleSpecificOveralls: []RoleOverallScore{
				{RoleName: "Goalkeeper (Defend)", Score: 85},
			},
		},
		{
			UID:                 2,
			Name:                "Player 2",
			Position:            "CB",
			Age:                 "28",
			Club:                "Club B",
			Division:            "Division 1",
			Nationality:         "Spain",
			NationalityISO:      "ESP",
			NationalityFIFACode: "ESP",
			Overall:             82,
			PAC:                 65,
			SHO:                 45,
			PAS:                 80,
			DRI:                 70,
			DEF:                 85,
			PHY:                 78,
			TransferValueAmount: 40000000,
			WageAmount:          80000,
			AttributeMasked:     false,
			ParsedPositions:     []string{"CB", "RB"},
			ShortPositions:      []string{"CB", "RB"},
			PositionGroups:      []string{"Defender"},
			PerformanceStatsNumeric: map[string]float64{
				"goals":           2.5,
				"assists":         1.8,
				"tackles_per_game": 4.2,
			},
			RoleSpecificOveralls: []RoleOverallScore{
				{RoleName: "Centre Back (Defend)", Score: 82},
				{RoleName: "Ball Playing Defender (Defend)", Score: 79},
			},
		},
	}
	
	table, err := manager.BuildArrowTable(players)
	if err != nil {
		t.Fatalf("Expected no error building table, got: %v", err)
	}
	
	if table == nil {
		t.Fatal("Expected table to be created, got nil")
	}
	
	if table.NumRows() != 2 {
		t.Errorf("Expected 2 rows, got %d", table.NumRows())
	}
	
	if table.NumCols() != int64(len(manager.schema.GetSchema().Fields())) {
		t.Errorf("Expected %d columns, got %d", len(manager.schema.GetSchema().Fields()), table.NumCols())
	}
	
	// Clean up
	table.Release()
}

func TestConvertPlayersToArrowTable(t *testing.T) {
	allocator := memory.NewGoAllocator()
	
	player := Player{
		UID:                 12345,
		Name:                "Test Player",
		Position:            "GK",
		Age:                 "25",
		Club:                "Test Club",
		Division:            "Test Division",
		Nationality:         "England",
		NationalityISO:      "ENG",
		NationalityFIFACode: "ENG",
		Overall:             85,
		PAC:                 70,
		SHO:                 30,
		PAS:                 85,
		DRI:                 40,
		DEF:                 20,
		PHY:                 80,
		TransferValueAmount: 50000000,
		WageAmount:          100000,
		AttributeMasked:     false,
		ParsedPositions:     []string{"GK"},
		ShortPositions:      []string{"GK"},
		PositionGroups:      []string{"Goalkeeper"},
		PerformanceStatsNumeric: map[string]float64{
			"goals":   0.1,
			"assists": 0.2,
		},
		RoleSpecificOveralls: []RoleOverallScore{
			{RoleName: "Goalkeeper (Defend)", Score: 85},
		},
	}
	
	table, err := ConvertPlayersToArrowTable([]Player{player}, allocator)
	if err != nil {
		t.Fatalf("Expected no error converting players, got: %v", err)
	}
	
	if table == nil {
		t.Fatal("Expected table to be created, got nil")
	}
	
	if table.NumRows() != 1 {
		t.Errorf("Expected 1 row, got %d", table.NumRows())
	}
	
	// Clean up
	table.Release()
}

func TestArrowDatasetManagerRelease(t *testing.T) {
	allocator := memory.NewGoAllocator()
	manager := NewArrowDatasetManager(allocator)
	
	// Initialize builders
	err := manager.initializeBuilders(10)
	if err != nil {
		t.Fatalf("Failed to initialize builders: %v", err)
	}
	
	// Release should not panic
	manager.Release()
	
	// Check that builders are cleared
	if manager.builders != nil {
		t.Error("Expected builders to be nil after release")
	}
	
	if manager.dictBuilders != nil {
		t.Error("Expected dictBuilders to be nil after release")
	}
}

func TestBuildArrowTableWithBatching(t *testing.T) {
	allocator := memory.NewGoAllocator()
	manager := NewArrowDatasetManager(allocator)
	
	// Set small batch size for testing
	manager.SetRecordBatchSize(2)
	
	// Create 5 players to test batching (should create 3 batches: 2, 2, 1)
	players := make([]Player, 5)
	for i := 0; i < 5; i++ {
		players[i] = Player{
			UID:                 int64(i + 1),
			Name:                "Player " + string(rune('A'+i)),
			Position:            "CB",
			Age:                 "25",
			Club:                "Test Club",
			Division:            "Test Division",
			Nationality:         "England",
			NationalityISO:      "ENG",
			NationalityFIFACode: "ENG",
			Overall:             80 + i,
			PAC:                 70,
			SHO:                 30,
			PAS:                 85,
			DRI:                 40,
			DEF:                 20,
			PHY:                 80,
			TransferValueAmount: int64(50000000 + i*1000000),
			WageAmount:          int64(100000 + i*10000),
			AttributeMasked:     false,
			ParsedPositions:     []string{"CB"},
			ShortPositions:      []string{"CB"},
			PositionGroups:      []string{"Defender"},
			PerformanceStatsNumeric: map[string]float64{
				"goals":   float64(i),
				"assists": float64(i) * 0.5,
			},
			RoleSpecificOveralls: []RoleOverallScore{
				{RoleName: "Centre Back (Defend)", Score: 80 + i},
			},
		}
	}
	
	table, err := manager.BuildArrowTable(players)
	if err != nil {
		t.Fatalf("Expected no error building table with batching, got: %v", err)
	}
	
	if table == nil {
		t.Fatal("Expected table to be created, got nil")
	}
	
	if table.NumRows() != 5 {
		t.Errorf("Expected 5 rows, got %d", table.NumRows())
	}
	
	// Clean up
	table.Release()
}

func BenchmarkBuildArrowTableSmall(b *testing.B) {
	allocator := memory.NewGoAllocator()
	
	// Create test data
	players := make([]Player, 100)
	for i := 0; i < 100; i++ {
		players[i] = Player{
			UID:                 int64(i + 1),
			Name:                "Player " + string(rune('A'+(i%26))),
			Position:            "CB",
			Age:                 "25",
			Club:                "Test Club",
			Division:            "Test Division",
			Nationality:         "England",
			NationalityISO:      "ENG",
			NationalityFIFACode: "ENG",
			Overall:             80,
			PAC:                 70,
			SHO:                 30,
			PAS:                 85,
			DRI:                 40,
			DEF:                 20,
			PHY:                 80,
			TransferValueAmount: 50000000,
			WageAmount:          100000,
			AttributeMasked:     false,
			ParsedPositions:     []string{"CB"},
			ShortPositions:      []string{"CB"},
			PositionGroups:      []string{"Defender"},
			PerformanceStatsNumeric: map[string]float64{
				"goals":   1.0,
				"assists": 0.5,
			},
			RoleSpecificOveralls: []RoleOverallScore{
				{RoleName: "Centre Back (Defend)", Score: 80},
			},
		}
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		table, err := ConvertPlayersToArrowTable(players, allocator)
		if err != nil {
			b.Fatalf("Error building table: %v", err)
		}
		table.Release()
	}
}

func BenchmarkStringInterner(b *testing.B) {
	interner := NewStringInterner()
	
	positions := []string{"GK", "CB", "LB", "RB", "CDM", "CM", "CAM", "LW", "RW", "ST"}
	clubs := []string{"Manchester United", "Liverpool", "Arsenal", "Chelsea", "Manchester City"}
	nationalities := []string{"England", "Spain", "France", "Germany", "Italy"}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		interner.InternPosition(positions[i%len(positions)])
		interner.InternClub(clubs[i%len(clubs)])
		interner.InternNationality(nationalities[i%len(nationalities)])
	}
}