package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"sync"
)

const (
	defaultPlayerCapacity    = 1024            // Initial capacity for player slices
	defaultAttributeCapacity = 64              // Initial capacity for attribute maps
	defaultCellCapacity      = 64              // Initial capacity for cell slices during parsing
	overallScalingFactor     = 5.85            // Used for scaling role-specific attribute averages (1-20) to 0-99
	maxTokenBufferSize       = 2 * 1024 * 1024 // 2MB, for bufio.NewReaderSize

	// Factors for combining role-based and category-based overalls for outfielders
	roleSpecificOverallFactor  = 0.8 // 80% weight for role-specific calculation
	categoryBasedOverallFactor = 0.2 // 20% weight for category-based calculation
)

// Global variables for attribute and role weights.
// These are populated at startup from JSON files or defaults.
var (
	attributeWeights             map[string]map[string]int // For PAC, SHO etc. from individual attributes
	roleSpecificOverallWeights   map[string]map[string]int // For role-specific overall from individual attributes
	muAttributeWeights           sync.RWMutex
	muRoleSpecificOverallWeights sync.RWMutex
	precomputedRoleWeights       map[string][]struct { // Optimized lookup for role weights
		RoleName string
		Weights  map[string]int
	}
	muPrecomputedRoleWeights sync.RWMutex

	// Default/Fallback weights for calculating a general overall based on FIFA stat categories for outfielders
	fifaCategoryOverallWeights = map[string]int{
		"PHY": 20, "PAC": 20, "PAS": 15, "DEF": 15, "DRI": 15, "SHO": 15, // Sums to 100
	}

	// Position-specific weights for FIFA stat categories for outfielders
	attackerFifaCategoryWeights = map[string]int{
		"SHO": 30, "PAC": 25, "DRI": 20, "PHY": 15, "PAS": 10, "DEF": 0, // Sums to 100
	}
	midfielderFifaCategoryWeights = map[string]int{
		"PAS": 30, "PHY": 20, "PAC": 20, "DRI": 15, "DEF": 10, "SHO": 5, // Sums to 100
	}
	defenderFifaCategoryWeights = map[string]int{
		"DEF": 30, "PHY": 25, "PAC": 20, "PAS": 15, "DRI": 5, "SHO": 5, // Sums to 100
	}

	// Metrics collection toggle
	metricsEnabled bool
)

// Default attribute weights if JSON loading fails or file is missing.
var defaultAttributeWeightsGo = map[string]map[string]int{
	"PAC": {"Acc": 8, "Pac": 8, "Agi": 5},
	"SHO": {"Fin": 8, "Lon": 6, "Pen": 4, "Hea": 5, "Cmp": 6, "Tec": 5, "Ant": 4, "Dec": 4, "Fla": 3},
	"PAS": {"Pas": 8, "Cro": 6, "Fre": 4, "Vis": 7, "Tec": 5, "Tea": 4, "Dec": 4, "Cor": 3, "Fir": 4, "OtB": 3},
	"DRI": {"Dri": 8, "Fir": 7, "Tec": 6, "Fla": 5, "Cmp": 4, "OtB": 3},
	"DEF": {"Mar": 8, "Tck": 8, "Hea": 6, "Ant": 7, "Cnt": 6, "Pos": 7, "Dec": 5, "Cmp": 4, "Bra": 5, "Agg": 4, "Wor": 4},
	"PHY": {"Str": 8, "Sta": 7, "Nat": 6, "Jum": 5, "Bal": 4, "Agg": 5, "Bra": 4, "Wor": 4},
	"GK":  {"Han": 20, "Ref": 20, "Cmd": 15, "Aer": 15, "1v1": 10, "Kic": 5, "TRO": 5, "Com": 3, "Thr": 3, "Ecc": 1},
	"DIV": {"Aer": 8, "Ref": 7, "Agi": 6, "1v1": 7, "Han": 5},
	"HAN": {"Han": 10, "Cmd": 7, "Cmp": 5, "Cnt": 4},
	"REF": {"Ref": 10, "Ant": 6, "Cnt": 5, "1v1": 5},
	"KIC": {"Kic": 8, "Thr": 6, "Tec": 5, "Vis": 4, "Pas": 3},
	"SPD": {"Acc": 8, "Pac": 8, "TRO": 6},
	"POS": {"Pos": 8, "Cmd": 7, "Ant": 6, "Dec": 5, "TRO": 4, "Cnt": 4, "Com": 3},
}

// Default role-specific overall weights if JSON loading fails or file is missing.
// "Generic" roles have been removed from this default map.
var defaultRoleSpecificOverallWeightsGo = map[string]map[string]int{
	"GK - Goalkeeper - Defend": {"Han": 90, "Ref": 90, "Aer": 80, "Cmd": 75, "1v1": 80, "Cnt": 70, "Dec": 70, "Pos": 75, "Ant": 60, "Cmp": 60, "Bra": 60, "Com": 50, "Kic": 40, "Thr": 40, "TRO": 30, "Det": 50, "Ldr": 40, "Wor": 40, "Tea": 40, "Agi": 50, "Jum": 60, "Str": 50, "Acc": 30, "Pac": 30, "Ecc": 10},
	// Add other non-generic default roles here if you have them.
	// For example, if you had a "DC - Ball Playing Defender - Defend" as a default, it would remain.
	// Example (assuming it was a default, keeping it for structure):
	"DC - Ball Playing Defender - Defend": {
		"Cor": 5, "Cro": 1, "Dri": 40, "Fin": 10, "Fir": 35, "Fre": 10, "Hea": 55, "Lon": 10, "Tea": 20, "L Th": 0, "Mar": 55, "Pas": 55, "Pen": 10, "Tck": 40, "Tec": 35,
		"Agg": 40, "Ant": 50, "Bra": 30, "Cmp": 80, "Cnt": 50, "Dec": 50, "Det": 20, "Fla": 10, "Ldr": 10, "OtB": 10, "Pos": 55, "Vis": 50, "Wor": 55,
		"Acc": 90, "Agi": 60, "Bal": 35, "Jum": 65, "Nat": 10, "Pac": 90, "Sta": 30, "Str": 50,
	},
	"DC - Central Defender - Defend": {
		"Cor": 10, "Cro": 10, "Dri": 30, "Fin": 10, "Fir": 30, "Fre": 5, "Hea": 60, "Lon": 0, "L Th": 0, "Mar": 70, "Pas": 40, "Pen": 0, "Tck": 70, "Tec": 30,
		"Agg": 60, "Ant": 65, "Bra": 50, "Cmp": 80, "Cnt": 65, "Dec": 65, "Det": 20, "Fla": 10, "Ldr": 10, "OtB": 10, "Pos": 65, "Tea": 20, "Vis": 30, "Wor": 60,
		"Acc": 60, "Agi": 30, "Bal": 30, "Jum": 65, "Nat": 10, "Pac": 70, "Sta": 40, "Str": 60,
	},
}

// PerformanceStatKeys lists the column headers for player performance statistics.
var PerformanceStatKeys = []string{
	"Asts/90", "Av Rat", "Blk/90", "Ch C/90", "Clr/90", "Cr C/90", "Drb/90",
	"xA/90", "xG/90", "Gls/90", "Hdrs W/90", "Int/90", "K Ps/90", "Ps C/90",
	"Shot/90", "Tck/90", "Poss Won/90", "ShT/90", "Pres C/90", "Poss Lost/90",
	"Pr passes/90", "Conv %", "Tck R", "Pas %", "Cr C/A",
}

// PositionGroupsForPercentiles defines broad player categories used for percentile calculations.
var PositionGroupsForPercentiles = []string{"Goalkeepers", "Defenders", "Midfielders", "Attackers"}

// DetailedPositionGroupsForPercentiles maps more specific role groups to their short position codes.
var DetailedPositionGroupsForPercentiles = map[string][]string{
	"Full-backs":                      {"DR", "DL"},
	"Centre-backs":                    {"DC"},
	"Wing-backs":                      {"WBR", "WBL"},
	"Defensive Midfielders":           {"DM"},
	"Central Midfielders":             {"MC"},
	"Wide Midfielders":                {"MR", "ML"},
	"Attacking Midfielders (Central)": {"AMC"},
	"Wingers":                         {"AMR", "AML"},
	"Strikers":                        {"ST"},
}

// loadJSONWeights attempts to load weights from a JSON file.
// If loading fails, it falls back to the provided default weights.
func loadJSONWeights(filePath string, defaultWeights map[string]map[string]int) (map[string]map[string]int, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Warning: Could not read %s: %v. Using default weights.", filePath, err)
		// Return a deep copy of default weights
		copiedDefault := make(map[string]map[string]int, len(defaultWeights))
		for k, v := range defaultWeights {
			innerCopy := make(map[string]int, len(v))
			for ik, iv := range v {
				innerCopy[ik] = iv
			}
			copiedDefault[k] = innerCopy
		}
		return copiedDefault, err
	}

	var weights map[string]map[string]int
	if err := json.Unmarshal(data, &weights); err != nil {
		log.Printf("Warning: Could not unmarshal %s: %v. Using default weights.", filePath, err)
		copiedDefault := make(map[string]map[string]int, len(defaultWeights))
		for k, v := range defaultWeights {
			innerCopy := make(map[string]int, len(v))
			for ik, iv := range v {
				innerCopy[ik] = iv
			}
			copiedDefault[k] = innerCopy
		}
		return copiedDefault, err
	}

	if len(weights) == 0 {
		log.Printf("Warning: Weights file %s was loaded but is empty. Using default weights.", filePath)
		copiedDefault := make(map[string]map[string]int, len(defaultWeights))
		for k, v := range defaultWeights {
			innerCopy := make(map[string]int, len(v))
			for ik, iv := range v {
				innerCopy[ik] = iv
			}
			copiedDefault[k] = innerCopy
		}
		return copiedDefault, errors.New("loaded weights file is empty")
	}

	log.Printf("Successfully loaded weights from %s with %d entries.", filePath, len(weights))
	return weights, nil
}

// init is automatically called when the package is loaded.
// It loads attribute and role weights from JSON files or uses defaults.
func init() {
	var errAttr, errRole error

	// Load attribute weights for FIFA stats (PAC, SHO, etc.)
	loadedAttrWeights, errAttr := loadJSONWeights(filepath.Join("public", "attribute_weights.json"), defaultAttributeWeightsGo)
	muAttributeWeights.Lock()
	if errAttr != nil {
		log.Printf("Using default attribute_weights due to error: %v. Default attribute_weights has %d entries.", errAttr, len(defaultAttributeWeightsGo))
		attributeWeights = make(map[string]map[string]int)
		for k, v := range defaultAttributeWeightsGo { // Deep copy
			innerMap := make(map[string]int)
			for ik, iv := range v {
				innerMap[ik] = iv
			}
			attributeWeights[k] = innerMap
		}
	} else {
		attributeWeights = loadedAttrWeights
	}
	muAttributeWeights.Unlock()

	// Load role-specific overall weights
	loadedRoleWeights, errRole := loadJSONWeights(filepath.Join("public", "role_specific_overall_weights.json"), defaultRoleSpecificOverallWeightsGo)
	muRoleSpecificOverallWeights.Lock()
	if errRole != nil {
		log.Printf("Using default role_specific_overall_weights due to error: %v. Default role_specific_overall_weights has %d entries.", errRole, len(defaultRoleSpecificOverallWeightsGo))
		roleSpecificOverallWeights = make(map[string]map[string]int)
		for k, v := range defaultRoleSpecificOverallWeightsGo { // Deep copy
			innerMap := make(map[string]int)
			for ik, iv := range v {
				innerMap[ik] = iv
			}
			roleSpecificOverallWeights[k] = innerMap
		}
	} else {
		roleSpecificOverallWeights = loadedRoleWeights
	}
	muRoleSpecificOverallWeights.Unlock()

	// Precompute role weights for faster lookup during player processing
	muPrecomputedRoleWeights.Lock()
	precomputedRoleWeights = make(map[string][]struct {
		RoleName string
		Weights  map[string]int
	})
	// Use the just loaded (or defaulted) roleSpecificOverallWeights for precomputation
	sourceWeightsToPrecompute := roleSpecificOverallWeights
	for roleFullName, weights := range sourceWeightsToPrecompute {
		// Extract short position key (e.g., "DC" from "DC - Central Defender - Defend")
		// This logic might need adjustment based on the exact format of roleFullName in your JSON
		shortKey := GetShortPositionKeyFromRoleName(roleFullName) // Assumes GetShortPositionKeyFromRoleName is defined, e.g., in positions.go

		if shortKey != "" {
			copiedWeights := make(map[string]int, len(weights)) // Deep copy weights
			for k, v := range weights {
				copiedWeights[k] = v
			}
			precomputedRoleWeights[shortKey] = append(precomputedRoleWeights[shortKey], struct {
				RoleName string
				Weights  map[string]int
			}{RoleName: roleFullName, Weights: copiedWeights})
		}
	}
	muPrecomputedRoleWeights.Unlock()
	log.Printf("Precomputed %d base position keys for role weights.", len(precomputedRoleWeights))

	// Initialize metrics toggle from environment variable
	metricsEnabled = os.Getenv("ENABLE_METRICS") == "true"
	log.Printf("Metrics collection enabled: %v", metricsEnabled)

	// Initialize OpenTelemetry metrics
	if metricsEnabled {
		initMetrics()
	}
}


