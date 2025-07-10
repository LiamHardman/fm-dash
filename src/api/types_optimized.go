package main

import (
	"strconv"
	"unsafe"
)

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
	PAC, SHO, PAS, DRI, DEF, PHY int16 `json:"pac,sho,pas,dri,def,phy"`

	// Goalkeeper stats (often zero, packed together)
	GK, DIV, HAN, REF, KIC, SPD, POS int16 `json:"gk,div,han,ref,kic,spd,pos,omitempty"`

	// Identifiers and core strings (interned)
	UID                 string `json:"uid"`
	Name                string `json:"name"`
	Club                string `json:"club"`                  // Interned
	Position            string `json:"position"`              // Interned
	Division            string `json:"division"`              // Interned
	Nationality         string `json:"nationality"`           // Interned
	NationalityISO      string `json:"nationality_iso"`       // Interned
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
		"Cor": 0, "Cro": 1, "Dri": 2, "Fin": 3, "Fir": 4, "Fre": 5,
		"Hea": 6, "Lon": 7, "L Th": 8, "Mar": 9, "Pas": 10, "Pen": 11,
		"Tck": 12, "Tec": 13, "Thr": 14, // Add more as needed
	}

	MentalAttrIndices = map[string]int{
		"Agg": 0, "Ant": 1, "Bra": 2, "Cmp": 3, "Cnt": 4, "Dec": 5,
		"Det": 6, "Fla": 7, "Ldr": 8, "OtB": 9, "Pos": 10, "Tea": 11,
		"Vis": 12, "Wor": 13, // Add more as needed
	}

	PhysicalAttrIndices = map[string]int{
		"Acc": 0, "Agi": 1, "Bal": 2, "Jum": 3, "Nat": 4, "Pac": 5, "Sta": 6, "Str": 7,
	}

	GoalkeeperAttrIndices = map[string]int{
		"Aer": 0, "Cmd": 1, "Com": 2, "Ecc": 3, "Han": 4, "Kic": 5,
		"1v1": 6, "Ref": 7, "TRO": 8, "Pun": 9,
	}
)

// ConvertToOptimized converts a regular Player to OptimizedPlayer
func ConvertToOptimized(player *Player) *OptimizedPlayer {
	opt := &OptimizedPlayer{
		TransferValueAmount:  player.TransferValueAmount,
		WageAmount:           player.WageAmount,
		Overall:              int32(player.Overall),
		UID:                  player.UID,
		Name:                 player.Name,
		Club:                 InternClub(player.Club),
		Position:             InternPosition(player.Position),
		Division:             InternDivision(player.Division),
		Nationality:          InternNationality(player.Nationality),
		NationalityISO:       InternNationality(player.NationalityISO),
		NationalityFIFACode:  InternNationality(player.NationalityFIFACode),
		BestRoleOverall:      InternPosition(player.BestRoleOverall),
		PAC:                  int16(player.PAC),
		SHO:                  int16(player.SHO),
		PAS:                  int16(player.PAS),
		DRI:                  int16(player.DRI),
		DEF:                  int16(player.DEF),
		PHY:                  int16(player.PHY),
		GK:                   int16(player.GK),
		DIV:                  int16(player.DIV),
		HAN:                  int16(player.HAN),
		REF:                  int16(player.REF),
		KIC:                  int16(player.KIC),
		SPD:                  int16(player.SPD),
		POS:                  int16(player.POS),
		AttributeMasked:      player.AttributeMasked,
		RoleSpecificOveralls: player.RoleSpecificOveralls,
	}

	// Parse age from string to int
	if age, err := strconv.Atoi(player.Age); err == nil {
		opt.Age = int32(age)
	}

	// Check if goalkeeper
	for _, group := range player.PositionGroups {
		if group == "Goalkeepers" {
			opt.IsGoalkeeper = true
			break
		}
	}

	// Copy positions with bounds checking
	opt.copyPositions(player.ParsedPositions, opt.ParsedPositions[:], &opt.PositionCount)
	opt.copyPositions(player.ShortPositions, opt.ShortPositions[:], &opt.ShortPositionCount)
	opt.copyPositions(player.PositionGroups, opt.PositionGroups[:], &opt.PositionGroupsCount)

	// Convert attributes from maps to arrays
	opt.convertAttributes(player.NumericAttributes)

	// Create extended data only if needed
	if player.TransferValue != "" || player.Wage != "" ||
		player.Personality != "" || player.MediaHandling != "" ||
		len(player.PerformanceStatsNumeric) > 0 ||
		len(player.PerformancePercentiles) > 0 ||
		len(player.Attributes) > 0 {
		opt.Extended = &PlayerExtendedData{
			TransferValue:           player.TransferValue,
			Wage:                    player.Wage,
			Personality:             InternPersonality(player.Personality),
			MediaHandling:           InternPersonality(player.MediaHandling),
			PerformanceStatsNumeric: player.PerformanceStatsNumeric,
			PerformancePercentiles:  player.PerformancePercentiles,
			CustomAttributes:        player.Attributes,
		}
	}

	return opt
}

// Helper method to copy positions with length tracking
func (op *OptimizedPlayer) copyPositions(src []string, dst []string, count *int8) {
	*count = 0
	for i, pos := range src {
		if i >= len(dst) {
			break // Array is full
		}
		dst[i] = InternPosition(pos)
		*count++
	}
}

// Helper method to convert attributes from map to arrays
func (op *OptimizedPlayer) convertAttributes(attrs map[string]int) {
	for name, value := range attrs {
		val := int8(value) // FM attributes are 1-20

		if idx, exists := TechnicalAttrIndices[name]; exists {
			op.TechnicalAttributes[idx] = val
		} else if idx, exists := MentalAttrIndices[name]; exists {
			op.MentalAttributes[idx] = val
		} else if idx, exists := PhysicalAttrIndices[name]; exists {
			op.PhysicalAttributes[idx] = val
		} else if idx, exists := GoalkeeperAttrIndices[name]; exists {
			op.GoalkeeperAttributes[idx] = val
		}
		// Unknown attributes go to Extended.CustomAttributes
	}
}

// GetAttribute gets an attribute value by name (for compatibility)
func (op *OptimizedPlayer) GetAttribute(name string) (int, bool) {
	if idx, exists := TechnicalAttrIndices[name]; exists {
		return int(op.TechnicalAttributes[idx]), true
	}
	if idx, exists := MentalAttrIndices[name]; exists {
		return int(op.MentalAttributes[idx]), true
	}
	if idx, exists := PhysicalAttrIndices[name]; exists {
		return int(op.PhysicalAttributes[idx]), true
	}
	if idx, exists := GoalkeeperAttrIndices[name]; exists {
		return int(op.GoalkeeperAttributes[idx]), true
	}

	// Check extended data
	if op.Extended != nil && op.Extended.CustomAttributes != nil {
		if strVal, exists := op.Extended.CustomAttributes[name]; exists {
			if intVal, err := strconv.Atoi(strVal); err == nil {
				return intVal, true
			}
		}
	}

	return 0, false
}

// GetPositions returns the actual positions (excluding empty slots)
func (op *OptimizedPlayer) GetPositions() []string {
	result := make([]string, 0, op.PositionCount)
	for i := int8(0); i < op.PositionCount && i < 4; i++ {
		if op.ParsedPositions[i] != "" {
			result = append(result, op.ParsedPositions[i])
		}
	}
	return result
}

// MemoryUsage returns estimated memory usage in bytes
func (op *OptimizedPlayer) MemoryUsage() int {
	size := int(unsafe.Sizeof(*op))

	// Add string lengths
	size += len(op.UID) + len(op.Name) + len(op.Club) + len(op.Position) +
		len(op.Division) + len(op.Nationality) + len(op.NationalityISO) +
		len(op.NationalityFIFACode) + len(op.BestRoleOverall)

	// Add position strings
	for i := int8(0); i < op.PositionCount && i < 4; i++ {
		size += len(op.ParsedPositions[i])
	}
	for i := int8(0); i < op.ShortPositionCount && i < 4; i++ {
		size += len(op.ShortPositions[i])
	}
	for i := int8(0); i < op.PositionGroupsCount && i < 2; i++ {
		size += len(op.PositionGroups[i])
	}

	// Add role-specific overalls
	size += len(op.RoleSpecificOveralls) * int(unsafe.Sizeof(RoleOverallScore{}))

	// Add extended data if present
	if op.Extended != nil {
		size += int(unsafe.Sizeof(*op.Extended))
		size += len(op.Extended.TransferValue) + len(op.Extended.Wage) +
			len(op.Extended.Personality) + len(op.Extended.MediaHandling)

		// Rough estimate for maps
		size += len(op.Extended.PerformanceStatsNumeric) * 24 // key + float64
		size += len(op.Extended.PerformancePercentiles) * 48  // nested maps
		size += len(op.Extended.CustomAttributes) * 32        // key + value
	}

	return size
}

// EstimateMemorySavings compares memory usage with original Player struct
func EstimateMemorySavings(originalCount int) (originalMB, optimizedMB float64, savingsPercent float64) {
	// Rough estimates based on analysis
	originalBytesPerPlayer := 1800
	optimizedBytesPerPlayer := 600 // Estimated with optimizations

	originalMB = float64(originalCount*originalBytesPerPlayer) / 1024 / 1024
	optimizedMB = float64(originalCount*optimizedBytesPerPlayer) / 1024 / 1024
	savingsPercent = (originalMB - optimizedMB) / originalMB * 100

	return
}
