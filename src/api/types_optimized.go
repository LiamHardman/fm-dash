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
	PAC int16 `json:"pac"`
	SHO int16 `json:"sho"`
	PAS int16 `json:"pas"`
	DRI int16 `json:"dri"`
	DEF int16 `json:"def"`
	PHY int16 `json:"phy"`

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

// ConvertToOptimized converts a regular Player to OptimizedPlayer
func ConvertToOptimized(player *Player) *OptimizedPlayer {
	opt := &OptimizedPlayer{
		TransferValueAmount:  player.TransferValueAmount,
		WageAmount:           player.WageAmount,
		Overall:              int32(clampInt32(player.Overall)),
		UID:                  player.UID,
		Name:                 player.Name,
		Club:                 safeClubInterning.SafeIntern(player.Club),
		Position:             safePositionInterning.SafeIntern(player.Position),
		Division:             safeDivisionInterning.SafeIntern(player.Division),
		Nationality:          safeNationalityInterning.SafeIntern(player.Nationality),
		NationalityISO:       safeNationalityInterning.SafeIntern(player.NationalityISO),
		NationalityFIFACode:  safeNationalityInterning.SafeIntern(player.NationalityFIFACode),
		BestRoleOverall:      safePositionInterning.SafeIntern(player.BestRoleOverall),
		PAC:                  int16(clampInt16(player.PAC)),
		SHO:                  int16(clampInt16(player.SHO)),
		PAS:                  int16(clampInt16(player.PAS)),
		DRI:                  int16(clampInt16(player.DRI)),
		DEF:                  int16(clampInt16(player.DEF)),
		PHY:                  int16(clampInt16(player.PHY)),
		GK:                   int16(clampInt16(player.GK)),
		DIV:                  int16(clampInt16(player.DIV)),
		HAN:                  int16(clampInt16(player.HAN)),
		REF:                  int16(clampInt16(player.REF)),
		KIC:                  int16(clampInt16(player.KIC)),
		SPD:                  int16(clampInt16(player.SPD)),
		POS:                  int16(clampInt16(player.POS)),
		AttributeMasked:      player.AttributeMasked,
		RoleSpecificOveralls: player.RoleSpecificOveralls,
	}

	// Parse age from string to int
	if age, err := strconv.Atoi(player.Age); err == nil {
		opt.Age = int32(clampInt32(age))
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
			Personality:             safePersonalityInterning.SafeIntern(player.Personality),
			MediaHandling:           safePersonalityInterning.SafeIntern(player.MediaHandling),
			PerformanceStatsNumeric: player.PerformanceStatsNumeric,
			PerformancePercentiles:  player.PerformancePercentiles,
			CustomAttributes:        player.Attributes,
		}
	}

	return opt
}

// Helper method to copy positions with length tracking
func (op *OptimizedPlayer) copyPositions(src, dst []string, count *int8) {
	*count = 0
	for i, pos := range src {
		if i >= len(dst) {
			break // Array is full
		}
		dst[i] = safePositionInterning.SafeIntern(pos)
		*count++
	}
}

// Helper method to convert attributes from map to arrays
func (op *OptimizedPlayer) convertAttributes(attrs map[string]int) {
	for name, value := range attrs {
		val := int8(clampInt8(value)) // FM attributes are 1-20

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
	size += len(op.Name) + len(op.Club) + len(op.Position) +
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
func EstimateMemorySavings(originalCount int) (originalMB, optimizedMB, savingsPercent float64) {
	// Rough estimates based on analysis
	originalBytesPerPlayer := 1800
	optimizedBytesPerPlayer := 600 // Estimated with optimizations

	originalMB = float64(originalCount*originalBytesPerPlayer) / 1024 / 1024
	optimizedMB = float64(originalCount*optimizedBytesPerPlayer) / 1024 / 1024
	savingsPercent = (originalMB - optimizedMB) / originalMB * 100

	return
}

// ConvertFromOptimized converts an OptimizedPlayer back to regular Player
func ConvertFromOptimized(opt *OptimizedPlayer) *Player {
	player := &Player{
		UID:                     opt.UID,
		Name:                    opt.Name,
		Club:                    opt.Club,
		Position:                opt.Position,
		Division:                opt.Division,
		Nationality:             opt.Nationality,
		NationalityISO:          opt.NationalityISO,
		NationalityFIFACode:     opt.NationalityFIFACode,
		BestRoleOverall:         opt.BestRoleOverall,
		Overall:                 int(opt.Overall),
		Age:                     strconv.Itoa(int(opt.Age)),
		PAC:                     int(opt.PAC),
		SHO:                     int(opt.SHO),
		PAS:                     int(opt.PAS),
		DRI:                     int(opt.DRI),
		DEF:                     int(opt.DEF),
		PHY:                     int(opt.PHY),
		GK:                      int(opt.GK),
		DIV:                     int(opt.DIV),
		HAN:                     int(opt.HAN),
		REF:                     int(opt.REF),
		KIC:                     int(opt.KIC),
		SPD:                     int(opt.SPD),
		POS:                     int(opt.POS),
		AttributeMasked:         opt.AttributeMasked,
		TransferValueAmount:     opt.TransferValueAmount,
		WageAmount:              opt.WageAmount,
		RoleSpecificOveralls:    opt.RoleSpecificOveralls,
		Attributes:              make(map[string]string),
		NumericAttributes:       make(map[string]int),
		PerformanceStatsNumeric: make(map[string]float64),
		PerformancePercentiles:  make(map[string]map[string]float64),
	}

	// Convert arrays back to maps
	opt.convertArraysToMaps(player.NumericAttributes)

	// Copy positions back
	for i := int8(0); i < opt.PositionCount && i < 4; i++ {
		player.ParsedPositions = append(player.ParsedPositions, opt.ParsedPositions[i])
	}
	for i := int8(0); i < opt.ShortPositionCount && i < 4; i++ {
		player.ShortPositions = append(player.ShortPositions, opt.ShortPositions[i])
	}
	for i := int8(0); i < opt.PositionGroupsCount && i < 2; i++ {
		player.PositionGroups = append(player.PositionGroups, opt.PositionGroups[i])
	}

	// Copy extended data if present
	if opt.Extended != nil {
		player.TransferValue = opt.Extended.TransferValue
		player.Wage = opt.Extended.Wage
		player.Personality = opt.Extended.Personality
		player.MediaHandling = opt.Extended.MediaHandling
		player.PerformanceStatsNumeric = opt.Extended.PerformanceStatsNumeric
		player.PerformancePercentiles = opt.Extended.PerformancePercentiles
		for k, v := range opt.Extended.CustomAttributes {
			player.Attributes[k] = v
		}
	}

	return player
}

// convertArraysToMaps converts the attribute arrays back to maps for compatibility
func (op *OptimizedPlayer) convertArraysToMaps(attrs map[string]int) {
	// Convert technical attributes
	for name, idx := range TechnicalAttrIndices {
		if idx < len(op.TechnicalAttributes) && op.TechnicalAttributes[idx] != 0 {
			attrs[name] = int(op.TechnicalAttributes[idx])
		}
	}

	// Convert mental attributes
	for name, idx := range MentalAttrIndices {
		if idx < len(op.MentalAttributes) && op.MentalAttributes[idx] != 0 {
			attrs[name] = int(op.MentalAttributes[idx])
		}
	}

	// Convert physical attributes
	for name, idx := range PhysicalAttrIndices {
		if idx < len(op.PhysicalAttributes) && op.PhysicalAttributes[idx] != 0 {
			attrs[name] = int(op.PhysicalAttributes[idx])
		}
	}

	// Convert goalkeeper attributes
	for name, idx := range GoalkeeperAttrIndices {
		if idx < len(op.GoalkeeperAttributes) && op.GoalkeeperAttributes[idx] != 0 {
			attrs[name] = int(op.GoalkeeperAttributes[idx])
		}
	}
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
