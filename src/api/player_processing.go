package main

import (
	"errors"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// Memory pools for reducing allocations during role calculations
var (
	roleSlicePool = sync.Pool{
		New: func() interface{} {
			return make([]struct {
				name    string
				weights map[string]int
			}, 0, 24)
		},
	}
	
	processedRoleNamesPool = sync.Pool{
		New: func() interface{} {
			return make(map[string]struct{}, 16)
		},
	}
)

// parseCellsToPlayer converts a slice of string cells (a row from the HTML table)
// into a Player struct, based on the provided headers.
func parseCellsToPlayer(cells []string, headers []string) (Player, error) {
	if len(headers) == 0 {
		return Player{}, errors.New("cannot process row: headers are empty")
	}

	// Ensure cells slice is at least as long as headers, padding with empty strings if necessary.
	if len(cells) < len(headers) {
		diff := len(headers) - len(cells)
		padding := make([]string, diff) // Slice of empty strings
		cells = append(cells, padding...)
	}

	player := Player{
		Attributes:              make(map[string]string, defaultAttributeCapacity),
		NumericAttributes:       make(map[string]int, defaultAttributeCapacity),
		PerformanceStatsNumeric: make(map[string]float64, len(PerformanceStatKeys)),
		PerformancePercentiles:  make(map[string]map[string]float64),
	}

	// Headers that are known not to be player attributes (e.g., "Inf", "Rec" for recommendations).
	knownNonAttributeHeaders := map[string]bool{
		"Inf": true, "Rec": true, "Salary": true, // "Salary" is often "Wage"
	}
	foundName := false

	for i, headerNameClean := range headers {
		cellValue := ""
		if i < len(cells) { // Should always be true due to padding
			cellValue = strings.TrimSpace(cells[i])
		}

		isAnAttributeField := true // Assume it's an attribute unless handled otherwise

		switch headerNameClean {
		case "Name":
			player.Name = cellValue
			if cellValue != "" {
				foundName = true
			}
			isAnAttributeField = false
		case "Position":
			player.Position = cellValue
			isAnAttributeField = false
		case "Age":
			player.Age = cellValue
			isAnAttributeField = false
		case "Club":
			player.Club = cellValue
			isAnAttributeField = false
		case "Division":
			player.Division = cellValue
			isAnAttributeField = false
		case "Transfer Value":
			player.TransferValue, player.TransferValueAmount, _ = ParseMonetaryValueGo(cellValue) // Currency symbol detection handled by caller if needed globally
			isAnAttributeField = false
		case "Wage":
			player.Wage, player.WageAmount, _ = ParseMonetaryValueGo(cellValue)
			isAnAttributeField = false
		case "Personality":
			player.Personality = cellValue
			isAnAttributeField = false
		case "Media Handling":
			player.MediaHandling = cellValue
			isAnAttributeField = false
		case "Nat": // This header is ambiguous: could be Nationality (3-letter code) or Natural Fitness (attribute)
			// Attempt to parse as int (for Natural Fitness attribute)
			valInt, err := strconv.Atoi(cellValue)
			if err == nil && valInt >= 1 && valInt <= 20 { // Valid range for an attribute
				player.Attributes[headerNameClean] = cellValue // Store "Nat" as an attribute
				// Numeric conversion will happen in EnhancePlayerWithCalculations
			} else { // Assume it's Nationality (FIFA code)
				fifaCode := strings.ToUpper(cellValue)
				player.NationalityFIFACode = fifaCode
				if fullName, ok := FifaCountryCodes[fifaCode]; ok {
					player.Nationality = fullName
				} else {
					player.Nationality = cellValue // Fallback to the provided value if code not found
				}
				if isoCode, ok := FifaToISO2[fifaCode]; ok {
					player.NationalityISO = isoCode
				} else {
					// Basic fallback for ISO code if FIFA code is unknown
					if len(fifaCode) >= 2 {
						player.NationalityISO = strings.ToLower(fifaCode[:2])
					} else {
						player.NationalityISO = strings.ToLower(fifaCode)
					}
				}
				isAnAttributeField = false // Processed as Nationality
			}
		case "Left Foot", "Right Foot":
			// These are often descriptive (e.g., "Very Strong") or ratings.
			// Treat as attributes if not empty.
			if cellValue != "" && cellValue != "-" {
				player.Attributes[headerNameClean] = cellValue
			} else {
				isAnAttributeField = false
			}
		default:
			// If not explicitly handled, it's potentially an attribute or a performance stat.
			// Attributes map will store it. Numeric conversion and performance stat parsing
			// will be handled in EnhancePlayerWithCalculations.
		}

		if isAnAttributeField {
			if _, isKnownNonAttr := knownNonAttributeHeaders[headerNameClean]; !isKnownNonAttr {
				if headerNameClean != "" && cellValue != "" && cellValue != "-" {
					player.Attributes[headerNameClean] = cellValue
				}
			}
		}
	}

	if !foundName {
		// Check if the row has any meaningful data at all before discarding
		isPotentiallyMeaningfulRow := false
		for _, cellContent := range cells {
			if strings.TrimSpace(cellContent) != "" {
				isPotentiallyMeaningfulRow = true
				break
			}
		}
		if isPotentiallyMeaningfulRow {
			// Log first few cells for debugging if a non-empty row lacks a name
			return Player{}, errors.New("skipped row: 'Name' field is missing or empty, but other data present. First few cells: " + strings.Join(GetFirstNCells(cells, 5), ", "))
		}
		return Player{}, errors.New("skipped row: 'Name' field missing and row appears empty or is likely a non-player row (e.g., header repetition, spacer)")
	}

	return player, nil
}

// EnhancePlayerWithCalculations populates the calculated fields of a Player struct,
// such as numeric attributes, FIFA stats, overall scores, parsed positions, etc.
// It modifies the player struct pointed to.
func EnhancePlayerWithCalculations(player *Player) {
	// --- START DEBUG LOGGING ---
	// Log raw attributes for a specific player to see what's coming in
	// You can change "Xavi Pachón" to a name you expect or use a counter for the first few players
	isSamplePlayer := player.Name == "Xavi Pachón" // Example, adjust if needed
	if isSamplePlayer {
		log.Printf("DEBUG EnhancePlayer START for %s. Raw Attributes: %v", player.Name, player.Attributes)
	}
	// --- END DEBUG LOGGING ---

	// Ensure maps are initialized (though parseCellsToPlayer should do this)
	if player.NumericAttributes == nil {
		player.NumericAttributes = make(map[string]int, len(player.Attributes))
	}
	if player.PerformanceStatsNumeric == nil {
		player.PerformanceStatsNumeric = make(map[string]float64, len(PerformanceStatKeys))
	}

	// Convert string attributes to numeric and parse performance stats
	for key, valStr := range player.Attributes {
		// Check if it's a known attribute that should be numeric (1-20)
		// This list should match attributes used in calculations.
		isNumericTechnicalAttribute := false
		switch key {
		// Physical
		case "Acc", "Pac", "Str", "Sta", "Nat", "Bal", "Jum", "Agi":
			isNumericTechnicalAttribute = true
		// Mental
		case "Agg", "Ant", "Bra", "Cmp", "Cnt", "Dec", "Det", "Fla", "Ldr", "OtB", "Pos", "Tea", "Vis", "Wor":
			isNumericTechnicalAttribute = true
		// Technical
		case "Cor", "Cro", "Dri", "Fin", "Fir", "Fre", "Hea", "Lon", "L Th", "Mar", "Pas", "Pen", "Tck", "Tec":
			isNumericTechnicalAttribute = true
		// Goalkeeping (handled separately if needed, but good to parse if present)
		case "Aer", "Cmd", "Com", "Ecc", "Han", "Kic", "1v1", "Ref", "TRO", "Thr", "Pun":
			isNumericTechnicalAttribute = true
		}

		if isNumericTechnicalAttribute {
			valInt, err := strconv.Atoi(valStr)
			if err == nil {
				player.NumericAttributes[key] = valInt
			} else {
				player.NumericAttributes[key] = 0 // Default to 0 if not a valid number
				if isSamplePlayer {
					log.Printf("DEBUG EnhancePlayer: Atoi FAILED for attribute key '%s', valStr '%s'. Error: %v. Defaulted to 0.", key, valStr, err)
				}
			}
		} else {
			// If not a technical/mental/physical/GK attribute, check if it's a performance stat
			isPerfStat := false
			for _, perfKey := range PerformanceStatKeys { // PerformanceStatKeys from config.go
				if key == perfKey {
					isPerfStat = true
					break
				}
			}
			if isPerfStat {
				if valStr != "-" && valStr != "" {
					statStrCleaned := strings.ReplaceAll(valStr, "%", "") // Remove percentage sign for parsing
					if val, err := strconv.ParseFloat(statStrCleaned, 64); err == nil {
						player.PerformanceStatsNumeric[key] = val
					} else {
						player.PerformanceStatsNumeric[key] = math.NaN() // Use NaN for unparseable stats
						if isSamplePlayer {
							log.Printf("DEBUG EnhancePlayer: ParseFloat FAILED for performance stat key '%s', valStr '%s'. Error: %v. Defaulted to NaN.", key, valStr, err)
						}
					}
				} else {
					player.PerformanceStatsNumeric[key] = math.NaN() // Use NaN for missing stats ("-")
				}
			}
			// If it's neither a known numeric attribute nor a performance stat, it remains in player.Attributes as a string.
		}
	}

	// --- START DEBUG LOGGING ---
	if isSamplePlayer {
		log.Printf("DEBUG EnhancePlayer MID for %s. NumericAttributes: %v", player.Name, player.NumericAttributes)
		log.Printf("DEBUG EnhancePlayer MID for %s. PerformanceStatsNumeric: %v", player.Name, player.PerformanceStatsNumeric)
	}
	// --- END DEBUG LOGGING ---

	// Parse positions
	player.ParsedPositions = ParsePlayerPositionsGo(player.Position)          // from positions.go
	player.PositionGroups = GetPlayerPositionGroupsGo(player.ParsedPositions) // from positions.go

	// Derive short positions (e.g., DC, ST)
	shortPosSet := make(map[string]struct{})
	for _, pPos := range player.ParsedPositions { // e.g., pPos = "Centre Back"
		if shortKey, ok := parsedPositionToBaseRoleKeyGo[pPos]; ok && shortKey != "" { // parsedPositionToBaseRoleKeyGo from positions.go
			shortPosSet[shortKey] = struct{}{}
		} else if pPos == "Goalkeeper" { // Explicit fallback for GK if map somehow misses it
			shortPosSet["GK"] = struct{}{}
		}
	}
	player.ShortPositions = make([]string, 0, len(shortPosSet))
	for sp := range shortPosSet {
		player.ShortPositions = append(player.ShortPositions, sp)
	}
	// Sort short positions according to display preference
	sort.Slice(player.ShortPositions, func(i, j int) bool {
		orderI, okI := ShortPositionOrderMap[player.ShortPositions[i]] // ShortPositionOrderMap from positions.go
		orderJ, okJ := ShortPositionOrderMap[player.ShortPositions[j]]
		if !okI {
			orderI = len(ShortPositionDisplayOrder) + i // Place unknown ones at the end (ShortPositionDisplayOrder from positions.go)
		}
		if !okJ {
			orderJ = len(ShortPositionDisplayOrder) + j
		}
		return orderI < orderJ
	})

	// Determine if player is a goalkeeper first
	isGoalkeeper := false
	for _, posGroup := range player.PositionGroups {
		if posGroup == "Goalkeepers" {
			isGoalkeeper = true
			break
		}
	}

	// Calculate FIFA-style category stats based on player type
	if isGoalkeeper {
		// Goalkeepers get goalkeeper-specific stats
		player.GK = CalculateFifaStatGo(player.NumericAttributes, "GK")
		player.DIV = CalculateFifaStatGo(player.NumericAttributes, "DIV")
		player.HAN = CalculateFifaStatGo(player.NumericAttributes, "HAN")
		player.REF = CalculateFifaStatGo(player.NumericAttributes, "REF")
		player.KIC = CalculateFifaStatGo(player.NumericAttributes, "KIC")
		player.SPD = CalculateFifaStatGo(player.NumericAttributes, "SPD")
		player.POS = CalculateFifaStatGo(player.NumericAttributes, "POS")
		// Set outfield stats to 0 for goalkeepers
		player.PAC = 0
		player.SHO = 0
		player.PAS = 0
		player.DRI = 0
		player.DEF = 0
		player.PHY = 0
	} else {
		// Outfield players get outfield stats
		player.PAC = CalculateFifaStatGo(player.NumericAttributes, "PAC") // CalculateFifaStatGo from calculations.go
		player.SHO = CalculateFifaStatGo(player.NumericAttributes, "SHO")
		player.PAS = CalculateFifaStatGo(player.NumericAttributes, "PAS")
		player.DRI = CalculateFifaStatGo(player.NumericAttributes, "DRI")
		player.DEF = CalculateFifaStatGo(player.NumericAttributes, "DEF")
		player.PHY = CalculateFifaStatGo(player.NumericAttributes, "PHY")
		// Set goalkeeper stats to 0 for outfield players
		player.GK = 0
		player.DIV = 0
		player.HAN = 0
		player.REF = 0
		player.KIC = 0
		player.SPD = 0
		player.POS = 0
	}

	// --- START: Overall Calculation (Optimized) ---
	maxRoleBasedOverall := 0
	calculatedRoleOveralls := make([]RoleOverallScore, 0, 8) // Increased estimate capacity

	// Get processedRoleNames from pool and ensure it's clean
	processedRoleNames := processedRoleNamesPool.Get().(map[string]struct{})
	defer func() {
		// Clear the map and return to pool
		for k := range processedRoleNames {
			delete(processedRoleNames, k)
		}
		processedRoleNamesPool.Put(processedRoleNames)
	}()

	muPrecomputedRoleWeights.RLock()
	currentPrecomputedWeights := precomputedRoleWeights // Use a local copy of the map pointer (precomputedRoleWeights from config.go)
	muPrecomputedRoleWeights.RUnlock()

	shouldUseFallback := len(currentPrecomputedWeights) == 0
	if shouldUseFallback {
		muRoleSpecificOverallWeights.RLock()
		if len(roleSpecificOverallWeights) > 0 { // roleSpecificOverallWeights from config.go
			log.Printf("Warning: precomputedRoleWeights is empty for player '%s'. Falling back to iterating roleSpecificOverallWeights (SLOW PATH). Check init logs.", player.Name)
		} else {
			shouldUseFallback = false
		}
		muRoleSpecificOverallWeights.RUnlock()
	}

	if shouldUseFallback {
		muRoleSpecificOverallWeights.RLock()
		fallbackWeights := roleSpecificOverallWeights
		muRoleSpecificOverallWeights.RUnlock()

		// Get matchingRoles slice from pool
		matchingRoles := roleSlicePool.Get().([]struct {
			name    string
			weights map[string]int
		})
		matchingRoles = matchingRoles[:0] // Reset length but keep capacity
		defer roleSlicePool.Put(matchingRoles)

		// Collect all matching roles first (avoid redundant string operations)
		for _, shortPosKey := range player.ShortPositions {
			posPrefix := shortPosKey + " - "
			isGK := shortPosKey == "GK"
			
			for roleFullName, specificWeights := range fallbackWeights {
				if _, alreadyProcessed := processedRoleNames[roleFullName]; alreadyProcessed {
					continue
				}
				
				if strings.HasPrefix(roleFullName, posPrefix) || (isGK && roleFullName == "GK - Goalkeeper - Defend") {
					matchingRoles = append(matchingRoles, struct{
						name    string
						weights map[string]int
					}{name: roleFullName, weights: specificWeights})
					processedRoleNames[roleFullName] = struct{}{}
				}
			}
		}

		// Batch process all matching roles
		for _, role := range matchingRoles {
			overallForThisRole := CalculateOverallForRoleGo(player.NumericAttributes, role.weights)
			calculatedRoleOveralls = append(calculatedRoleOveralls, RoleOverallScore{RoleName: role.name, Score: overallForThisRole})
			if overallForThisRole > maxRoleBasedOverall {
				maxRoleBasedOverall = overallForThisRole
			}
		}
		
		if len(calculatedRoleOveralls) == 0 && len(player.ShortPositions) > 0 {
			log.Printf("Fallback Warning: Player '%s' with ShortPositions %v found no matching roles in fallback roleSpecificOverallWeights. MaxRoleBasedOverall will be 0.", player.Name, player.ShortPositions)
		}

	} else if len(player.ShortPositions) > 0 {
		foundAnyRoleMatch := false

		// Get allApplicableRoles slice from pool
		allApplicableRoles := roleSlicePool.Get().([]struct {
			name    string
			weights map[string]int
		})
		allApplicableRoles = allApplicableRoles[:0] // Reset length but keep capacity
		defer roleSlicePool.Put(allApplicableRoles)

		// Collect all applicable roles from all positions in one pass
		for _, shortKey := range player.ShortPositions {
			if applicableRoles, found := currentPrecomputedWeights[shortKey]; found {
				foundAnyRoleMatch = true
				for _, roleData := range applicableRoles {
					if _, alreadyProcessed := processedRoleNames[roleData.RoleName]; !alreadyProcessed {
						allApplicableRoles = append(allApplicableRoles, struct{
							name string
							weights map[string]int
						}{name: roleData.RoleName, weights: roleData.Weights})
						processedRoleNames[roleData.RoleName] = struct{}{}
					}
				}
			}
		}

		// Batch process all applicable roles
		for _, role := range allApplicableRoles {
			overallForThisRole := CalculateOverallForRoleGo(player.NumericAttributes, role.weights)
			calculatedRoleOveralls = append(calculatedRoleOveralls, RoleOverallScore{RoleName: role.name, Score: overallForThisRole})
			if overallForThisRole > maxRoleBasedOverall {
				maxRoleBasedOverall = overallForThisRole
			}
		}
		
		if !foundAnyRoleMatch && len(player.ShortPositions) > 0 {
			log.Printf("Warning: Player '%s' with ShortPositions %v found no matching roles in precomputedRoleWeights. MaxRoleBasedOverall will be 0.", player.Name, player.ShortPositions)
		}
	} else {
		// This case means player has no short positions, so maxRoleBasedOverall will naturally be 0.
		// Log only if it's unexpected or for debugging.
		// log.Printf("Player '%s' has no ShortPositions. MaxRoleBasedOverall will be 0.", player.Name)
	}

	player.RoleSpecificOveralls = calculatedRoleOveralls
	sort.Slice(player.RoleSpecificOveralls, func(i, j int) bool {
		if player.RoleSpecificOveralls[i].Score != player.RoleSpecificOveralls[j].Score {
			return player.RoleSpecificOveralls[i].Score > player.RoleSpecificOveralls[j].Score
		}
		return player.RoleSpecificOveralls[i].RoleName < player.RoleSpecificOveralls[j].RoleName
	})

	if isGoalkeeper {
		player.Overall = maxRoleBasedOverall
	} else {
		var selectedCategoryWeights map[string]int
		playerIsAttacker, playerIsMidfielder, playerIsDefender := false, false, false

		for _, group := range player.PositionGroups {
			switch group {
			case "Attackers":
				playerIsAttacker = true
			case "Midfielders", "Wing-Backs":
				playerIsMidfielder = true
			case "Defenders":
				playerIsDefender = true
			}
		}

		if playerIsAttacker {
			selectedCategoryWeights = attackerFifaCategoryWeights // from config.go
		} else if playerIsMidfielder {
			selectedCategoryWeights = midfielderFifaCategoryWeights // from config.go
		} else if playerIsDefender {
			selectedCategoryWeights = defenderFifaCategoryWeights // from config.go
		} else {
			selectedCategoryWeights = fifaCategoryOverallWeights // from config.go
			if len(player.PositionGroups) > 0 {                  // Log only if player has positions but doesn't fit main groups
				log.Printf("Player %s (%v) using GENERIC category weights for blended overall.", player.Name, player.PositionGroups)
			}
		}

		categoryBasedOverall := CalculateCategoryBasedOverall(player, selectedCategoryWeights)                                                             // from calculations.go
		finalOverall := int(math.Round(float64(maxRoleBasedOverall)*roleSpecificOverallFactor + float64(categoryBasedOverall)*categoryBasedOverallFactor)) // factors from config.go
		player.Overall = Clamp(finalOverall, 0, 99)                                                                                                        // Clamp from utils.go
	}
	// --- END: Overall Calculation ---

	// --- START DEBUG LOGGING ---
	if isSamplePlayer {
		log.Printf("DEBUG EnhancePlayer END for %s. Final Overall: %d, PAC: %d, SHO: %d, PAS: %d, DRI: %d, DEF: %d, PHY: %d, GK: %d, DIV: %d, HAN: %d, REF: %d, KIC: %d, SPD: %d, POS: %d",
			player.Name, player.Overall, player.PAC, player.SHO, player.PAS, player.DRI, player.DEF, player.PHY, player.GK, player.DIV, player.HAN, player.REF, player.KIC, player.SPD, player.POS)
		log.Printf("DEBUG EnhancePlayer END for %s. RoleSpecificOveralls: %v", player.Name, player.RoleSpecificOveralls)
	}
	// --- END DEBUG LOGGING ---
}
