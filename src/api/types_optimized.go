package main


// OptimizedPlayer is a memory-efficient version of the Player struct
// Memory layout is optimized for 64-bit architectures
type OptimizedPlayer struct {
	// === CORE DATA (frequently accessed, cache-friendly) ===
	// Group 64-bit values first for alignment
	TransferValueAmount int64 `json:"transferValueAmount"`
	WageAmount          int64 `json:"wageAmount"`

	// Group 32-bit integers together
	Overall int32 `json:"overall"`
	Age     int32 `json:"age"` // Changed from string to int

	// FIFA stats as int16 (range 0-99, saves 2 bytes per field)
	PAC        int16 `json:"pac"`
	SHO        int16 `json:"sho"`
	PAS        int16 `json:"pas"`
	DRI        int16 `json:"dri"`
	DEF        int16 `json:"def"`
	PHY        int16 `json:"phy"`
	TotalStats int16 `json:"totalStats"`
	MBR        int16 `json:"mbr"`

	// Goalkeeper stats (often zero, packed together)
	GK  int16 `json:"gk"`
	DIV int16 `json:"div"`
	HAN int16 `json:"han"`
	REF int16 `json:"ref"`
	KIC int16 `json:"kic"`
	SPD int16 `json:"spd"`
	POS int16 `json:"pos"`

	// Identifiers and core strings (interned)
	UID                 int64  `json:"uid"`
	Name                string `json:"name"`
	Club                string `json:"club"`                  // Interned
	Position            string `json:"position"`              // Interned
	Division            string `json:"division"`              // Interned
	Nationality         string `json:"nationality"`           // Interned
	NationalityISO      string `json:"nationalityIso"`        // Interned - Changed from nationality_iso
	NationalityFIFACode string `json:"nationality_fifa_code"` // Interned
	BestRoleOverall     string `json:"bestRoleOverall"`       // Interned

	// Compact arrays instead of maps for known attributes (1-20 range)
	// Use int8 for FM attributes (range 1-20, saves significant memory)
	TechnicalAttributes  [20]int8 `json:"technical_attributes"`  // Cor, Cro, Dri, Fin, etc.
	MentalAttributes     [15]int8 `json:"mental_attributes"`     // Agg, Ant, Bra, Cmp, etc.
	PhysicalAttributes   [8]int8  `json:"physical_attributes"`   // Acc, Agi, Bal, Jum, etc.
	GoalkeeperAttributes [10]int8 `json:"goalkeeper_attributes"` // Aer, Cmd, Com, Ecc, etc.

	// Position data (using small slices instead of variable slices)
	ParsedPositions [4]string `json:"parsedPositions"` // Most players have 1-3 positions
	ShortPositions  [4]string `json:"shortPositions"`
	PositionGroups  [2]string `json:"positionGroups"` // Usually just one group

	// Role-specific data
	RoleSpecificOveralls []RoleOverallScore `json:"roleSpecificOveralls"`

	// Flags and small values (grouped at end to minimize padding)
	PositionCount       int8 `json:"-"` // Track actual positions used
	ShortPositionCount  int8 `json:"-"`
	PositionGroupsCount int8 `json:"-"`
	AttributeMasked     bool `json:"attributeMasked,omitempty"`
	IsGoalkeeper        bool `json:"-"` // Cached for quick checks

	// === EXTENDED DATA (pointer to reduce memory when not needed) ===
	Extended *PlayerExtendedData `json:"extended,omitempty"`
}

// PlayerExtendedData contains less frequently accessed data
type PlayerExtendedData struct {
	// Raw string values (kept for compatibility)
	TransferValue string `json:"transfer_value"`
	Wage          string `json:"wage"`
	Personality   string `json:"personality,omitempty"`    // Interned
	MediaHandling string `json:"media_handling,omitempty"` // Interned

	// Performance data (loaded on demand)
	PerformanceStatsNumeric map[string]float64            `json:"performanceStatsNumeric,omitempty"`
	PerformancePercentiles  map[string]map[string]float64 `json:"performancePercentiles,omitempty"`

	// Raw attributes map (for unknown/custom attributes)
	CustomAttributes map[string]string `json:"customAttributes,omitempty"`
}

// AttributeIndices maps attribute names to array indices for fast lookups
var (
	TechnicalAttrIndices = map[string]int{
		// Core technical attributes (expanded to include all common FM attributes)
		"Cor": 0, "Cro": 1, "Dri": 2, "Fin": 3, "Fir": 4, "Fre": 5,
		"Hea": 6, "Lon": 7, "L Th": 8, "Mar": 9, "Pas": 10, "Pen": 11,
		"Tck": 12, "Tec": 13, "Thr": 14,
	}

	MentalAttrIndices = map[string]int{
		// Core mental attributes (expanded for comprehensive coverage)
		"Agg": 0, "Ant": 1, "Bra": 2, "Cmp": 3, "Cnt": 4, "Dec": 5,
		"Det": 6, "Fla": 7, "Ldr": 8, "OtB": 9, "Pos": 10, "Tea": 11,
		"Vis": 12, "Wor": 13,
	}

	PhysicalAttrIndices = map[string]int{
		// All physical attributes
		"Acc": 0, "Agi": 1, "Bal": 2, "Jum": 3, "Nat": 4, "Pac": 5, "Sta": 6, "Str": 7,
	}

	GoalkeeperAttrIndices = map[string]int{
		// Comprehensive goalkeeper attributes
		"Aer": 0, "Cmd": 1, "Com": 2, "Ecc": 3, "Han": 4, "Kic": 5,
		"1v1": 6, "Ref": 7, "TRO": 8, "Pun": 9,
	}
)

// EstimateMemorySavings compares memory usage with original Player struct
func EstimateMemorySavings(originalCount int) (originalMB, optimizedMB, savingsPercent float64) {
	// Rough estimates based on analysis
	originalBytesPerPlayer := 1800
	optimizedBytesPerPlayer := 600 // Estimated with optimizations

	originalMB = float64(originalCount*originalBytesPerPlayer) / 1024 / 1024
	optimizedMB = float64(originalCount*optimizedBytesPerPlayer) / 1024 / 1024
	savingsPercent = (originalMB - optimizedMB) / originalMB * 100

	return
}

// Helper functions to clamp integer values to prevent overflow
func clampInt8(val int) int8 {
	if val > 127 {
		return 127
	}
	if val < -128 {
		return -128
	}
	return int8(val)
}

func clampInt16(val int) int16 {
	if val > 32767 {
		return 32767
	}
	if val < -32768 {
		return -32768
	}
	return int16(val)
}

func clampInt32(val int) int32 {
	if val > 2147483647 {
		return 2147483647
	}
	if val < -2147483648 {
		return -2147483648
	}
	return int32(val)
}
