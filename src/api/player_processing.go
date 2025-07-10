package main

import (
	"errors"
	"fmt"
	"log"
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

// fastParseInt provides optimized integer parsing with fast paths
func fastParseInt(s string) (int, error) {
	if s == "" {
		return 0, errors.New("empty string")
	}

	// Fast path for single digit numbers (common for FM attributes 1-20)
	if len(s) == 1 {
		if s[0] >= '0' && s[0] <= '9' {
			return int(s[0] - '0'), nil
		}
		return 0, errors.New("invalid character")
	}

	// Fast path for two digit numbers (also common for FM attributes)
	if len(s) == 2 {
		if s[0] >= '1' && s[0] <= '9' && s[1] >= '0' && s[1] <= '9' {
			return int(s[0]-'0')*10 + int(s[1]-'0'), nil
		}
	}

	// Fallback to standard parsing for larger numbers
	return strconv.Atoi(s)
}

// fastParseFloat provides optimized float parsing for performance stats
func fastParseFloat(s string) (float64, error) {
	if s == "" {
		return 0, errors.New("empty string")
	}

	// Fast path for simple integers that don't need float parsing
	if !strings.Contains(s, ".") && !strings.Contains(s, "e") && !strings.Contains(s, "E") {
		if val, err := fastParseInt(s); err == nil {
			return float64(val), nil
		}
	}

	return strconv.ParseFloat(s, 64)
}

// parseCellsToPlayer converts a slice of string cells (a row from the HTML table)
// into a Player struct, based on the provided headers.
func parseCellsToPlayer(cells, headers []string) (Player, error) {
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
		case "UID", "uid", "Uid", "ID", "id", "Id", "Player ID", "PlayerId", "player_id", "unique_id", "UniqueId":
			player.UID = cellValue
			isAnAttributeField = false
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
			// Handle masked Natural Fitness attribute first
			if cellValue == "-" {
				// This is a masked Natural Fitness attribute, store it as an attribute
				player.Attributes[headerNameClean] = cellValue
				// Don't overwrite nationality if it's already been set
				isAnAttributeField = true
			} else {
				// Attempt to parse as int or range (for Natural Fitness attribute) using optimized parsing
				valInt, err := fastParseInt(cellValue)
				isValidRange := strings.Contains(cellValue, "-") && len(strings.Split(cellValue, "-")) == 2

				if (err == nil && valInt >= 1 && valInt <= 20) || isValidRange {
					// This is Natural Fitness (numeric attribute or range like "16-18")
					player.Attributes[headerNameClean] = cellValue // Store "Nat" as an attribute
					// Don't overwrite nationality if it's already been set
					isAnAttributeField = true
				} else {
					// Only process as Nationality if we haven't already set nationality from a previous "Nat" column
					if player.Nationality == "" && player.NationalityFIFACode == "" {
						// Assume it's Nationality (FIFA code)
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
					} else {
						// Nationality already set, this must be a non-numeric Natural Fitness value
						// Store as attribute but don't overwrite nationality
						player.Attributes[headerNameClean] = cellValue
						isAnAttributeField = true
					}
				}
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
				if headerNameClean != "" && cellValue != "" {
					// Include all attributes, even those with "-" (masked) values
					// The frontend will handle displaying masked attributes appropriately
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
	// Ensure maps are initialized (though parseCellsToPlayer should do this)
	if player.NumericAttributes == nil {
		player.NumericAttributes = make(map[string]int, len(player.Attributes))
	}
	if player.PerformanceStatsNumeric == nil {
		player.PerformanceStatsNumeric = make(map[string]float64, len(PerformanceStatKeys))
	}

	// Apply string interning optimization for memory efficiency (after all map operations)
	// Use enhanced version with compression for better memory savings
	if memOptConfig.UseStringInterning {
		EnhancedOptimizePlayerStrings(player)
	}

	// Note: OptimizedPlayer conversion disabled here as it's inefficient to convert back immediately
	// OptimizedPlayer should be used in storage and processing layers, not during enhancement
	// The conversion happens in storage layers where memory benefits are actually realized
	if memOptConfig.UseOptimizedStructs {
		// This optimization is handled at the storage level for maximum memory benefit
		// Converting here and back would negate the memory savings
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
			// Check for masked attributes like "10-15" or completely masked "-"
			switch {
			case valStr == "-":
				// Completely masked attribute
				player.NumericAttributes[key] = 0 // Default to 0 for completely masked
				player.AttributeMasked = true
			case strings.Contains(valStr, "-"):
				parts := strings.Split(valStr, "-")
				if len(parts) == 2 {
					val1, err1 := fastParseInt(strings.TrimSpace(parts[0]))
					val2, err2 := fastParseInt(strings.TrimSpace(parts[1]))

					if err1 == nil && err2 == nil {
						player.NumericAttributes[key] = (val1 + val2) / 2
						player.AttributeMasked = true
					} else {
						player.NumericAttributes[key] = 0 // Default on parsing error
					}
				} else {
					player.NumericAttributes[key] = 0 // Default if split is not two parts
				}
			default:
				valInt, err := fastParseInt(valStr)
				if err == nil {
					player.NumericAttributes[key] = valInt
				} else {
					player.NumericAttributes[key] = 0 // Default to 0 if not a valid number
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

					// Handle specific cases for different performance stats
					var parsedValue float64
					var err error

					switch key {
					case "Mins":
						// Handle comma-separated minutes like "1,980"
						minsCleaned := strings.ReplaceAll(statStrCleaned, ",", "")
						parsedValue, err = fastParseFloat(minsCleaned)
					case "Apps":
						// Handle appearances with substitutes like "22 (4)"
						if strings.Contains(statStrCleaned, "(") {
							// Extract main appearances and substitute appearances
							parts := strings.Split(statStrCleaned, " (")
							if len(parts) == 2 {
								mainApps, err1 := fastParseFloat(strings.TrimSpace(parts[0]))
								subAppsStr := strings.TrimRight(strings.TrimSpace(parts[1]), ")")
								subApps, err2 := fastParseFloat(subAppsStr)
								if err1 == nil && err2 == nil {
									parsedValue = mainApps + subApps
									err = nil
								} else {
									err = fmt.Errorf("failed to parse appearances")
								}
							} else {
								err = fmt.Errorf("invalid appearances format")
							}
						} else {
							// Simple number
							parsedValue, err = fastParseFloat(statStrCleaned)
						}
					default:
						// Default parsing - remove commas for any other stats that might have them
						statStrCleaned = strings.ReplaceAll(statStrCleaned, ",", "")
						parsedValue, err = fastParseFloat(statStrCleaned)
					}

					if err == nil {
						player.PerformanceStatsNumeric[key] = parsedValue
					}
					// Don't store anything for unparseable stats - absence from map indicates missing data
				}
				// Don't store anything for missing stats ("-") - absence from map indicates missing data
			}
			// If it's neither a known numeric attribute nor a performance stat, it remains in player.Attributes as a string.
		}
	}

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
		if GetUseScaledRatings() {
			player.GK = CalculateFifaStatGo(player.NumericAttributes, "GK")
			player.DIV = CalculateFifaStatGo(player.NumericAttributes, "DIV")
			player.HAN = CalculateFifaStatGo(player.NumericAttributes, "HAN")
			player.REF = CalculateFifaStatGo(player.NumericAttributes, "REF")
			player.KIC = CalculateFifaStatGo(player.NumericAttributes, "KIC")
			player.SPD = CalculateFifaStatGo(player.NumericAttributes, "SPD")
			player.POS = CalculateFifaStatGo(player.NumericAttributes, "POS")
		} else {
			player.GK = CalculateFifaStatGoLinear(player.NumericAttributes, "GK")
			player.DIV = CalculateFifaStatGoLinear(player.NumericAttributes, "DIV")
			player.HAN = CalculateFifaStatGoLinear(player.NumericAttributes, "HAN")
			player.REF = CalculateFifaStatGoLinear(player.NumericAttributes, "REF")
			player.KIC = CalculateFifaStatGoLinear(player.NumericAttributes, "KIC")
			player.SPD = CalculateFifaStatGoLinear(player.NumericAttributes, "SPD")
			player.POS = CalculateFifaStatGoLinear(player.NumericAttributes, "POS")
		}
		// Set outfield stats to 0 for goalkeepers
		player.PAC = 0
		player.SHO = 0
		player.PAS = 0
		player.DRI = 0
		player.DEF = 0
		player.PHY = 0
	} else {
		// Outfield players get outfield stats
		if GetUseScaledRatings() {
			player.PAC = CalculateFifaStatGo(player.NumericAttributes, "PAC") // CalculateFifaStatGo from calculations.go
			player.SHO = CalculateFifaStatGo(player.NumericAttributes, "SHO")
			player.PAS = CalculateFifaStatGo(player.NumericAttributes, "PAS")
			player.DRI = CalculateFifaStatGo(player.NumericAttributes, "DRI")
			player.DEF = CalculateFifaStatGo(player.NumericAttributes, "DEF")
			player.PHY = CalculateFifaStatGo(player.NumericAttributes, "PHY")
		} else {
			player.PAC = CalculateFifaStatGoLinear(player.NumericAttributes, "PAC")
			player.SHO = CalculateFifaStatGoLinear(player.NumericAttributes, "SHO")
			player.PAS = CalculateFifaStatGoLinear(player.NumericAttributes, "PAS")
			player.DRI = CalculateFifaStatGoLinear(player.NumericAttributes, "DRI")
			player.DEF = CalculateFifaStatGoLinear(player.NumericAttributes, "DEF")
			player.PHY = CalculateFifaStatGoLinear(player.NumericAttributes, "PHY")
		}
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
	bestRoleName := ""
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
			// Only log this warning once to avoid spam
			log.Printf("Warning: precomputedRoleWeights is empty. Falling back to iterating roleSpecificOverallWeights (SLOW PATH). Check init logs.")
		} else {
			shouldUseFallback = false
		}
		muRoleSpecificOverallWeights.RUnlock()
	}

	switch {
	case shouldUseFallback:
		muRoleSpecificOverallWeights.RLock()
		fallbackWeights := roleSpecificOverallWeights
		muRoleSpecificOverallWeights.RUnlock()

		// Get matchingRoles slice from pool
		matchingRoles := roleSlicePool.Get().([]struct {
			name    string
			weights map[string]int
		})
		matchingRoles = matchingRoles[:0] // Reset length but keep capacity
		//nolint:staticcheck // SA6002: slice type is appropriate for sync.Pool
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
					matchingRoles = append(matchingRoles, struct {
						name    string
						weights map[string]int
					}{name: roleFullName, weights: specificWeights})
					processedRoleNames[roleFullName] = struct{}{}
				}
			}
		}

		// Batch process all matching roles
		for _, role := range matchingRoles {
			var overallForThisRole int
			if GetUseScaledRatings() {
				overallForThisRole = CalculateOverallForRoleGo(player.NumericAttributes, role.weights)
			} else {
				overallForThisRole = CalculateOverallForRoleGoLinear(player.NumericAttributes, role.weights)
			}
			calculatedRoleOveralls = append(calculatedRoleOveralls, RoleOverallScore{RoleName: role.name, Score: overallForThisRole})
			if overallForThisRole > maxRoleBasedOverall {
				maxRoleBasedOverall = overallForThisRole
				bestRoleName = role.name
			}
		}

		if len(calculatedRoleOveralls) == 0 && len(player.ShortPositions) > 0 {
			log.Printf("Fallback Warning: Player '%s' with ShortPositions %v found no matching roles in fallback roleSpecificOverallWeights. Role-based overall will be 0.", player.Name, player.ShortPositions)
		}

	case len(player.ShortPositions) > 0:
		foundAnyRoleMatch := false

		// Get allApplicableRoles slice from pool
		allApplicableRoles := roleSlicePool.Get().([]struct {
			name    string
			weights map[string]int
		})
		allApplicableRoles = allApplicableRoles[:0] // Reset length but keep capacity
		//nolint:staticcheck // SA6002: slice type is appropriate for sync.Pool
		defer roleSlicePool.Put(allApplicableRoles)

		// Collect all applicable roles from all positions in one pass
		for _, shortKey := range player.ShortPositions {
			if applicableRoles, found := currentPrecomputedWeights[shortKey]; found {
				foundAnyRoleMatch = true
				for _, roleData := range applicableRoles {
					if _, alreadyProcessed := processedRoleNames[roleData.RoleName]; !alreadyProcessed {
						allApplicableRoles = append(allApplicableRoles, struct {
							name    string
							weights map[string]int
						}{name: roleData.RoleName, weights: roleData.Weights})
						processedRoleNames[roleData.RoleName] = struct{}{}
					}
				}
			}
		}

		// Batch process all applicable roles
		for _, role := range allApplicableRoles {
			var overallForThisRole int
			if GetUseScaledRatings() {
				overallForThisRole = CalculateOverallForRoleGo(player.NumericAttributes, role.weights)
			} else {
				overallForThisRole = CalculateOverallForRoleGoLinear(player.NumericAttributes, role.weights)
			}
			calculatedRoleOveralls = append(calculatedRoleOveralls, RoleOverallScore{RoleName: role.name, Score: overallForThisRole})
			if overallForThisRole > maxRoleBasedOverall {
				maxRoleBasedOverall = overallForThisRole
				bestRoleName = role.name
			}
		}

		if !foundAnyRoleMatch && len(player.ShortPositions) > 0 {
			log.Printf("Warning: Player '%s' with ShortPositions %v found no matching roles in precomputedRoleWeights. Role-based overall will be 0.", player.Name, player.ShortPositions)
		}
	default:
		// This case means player has no short positions, so maxRoleBasedOverall will naturally be 0.
	}

	player.RoleSpecificOveralls = calculatedRoleOveralls
	sort.Slice(player.RoleSpecificOveralls, func(i, j int) bool {
		if player.RoleSpecificOveralls[i].Score != player.RoleSpecificOveralls[j].Score {
			return player.RoleSpecificOveralls[i].Score > player.RoleSpecificOveralls[j].Score
		}
		return player.RoleSpecificOveralls[i].RoleName < player.RoleSpecificOveralls[j].RoleName
	})

	// Set the best role name
	player.BestRoleOverall = bestRoleName

	// Calculate overall as the mean of the top 7 role ratings instead of using all role ratings
	meanRoleBasedOverall := 0
	if len(calculatedRoleOveralls) > 0 {
		// Sort the role overalls by score (highest first) to get the top ratings
		sort.Slice(calculatedRoleOveralls, func(i, j int) bool {
			return calculatedRoleOveralls[i].Score > calculatedRoleOveralls[j].Score
		})

		// Take the top 7 role ratings (or all of them if there are fewer than 7)
		topRoleCount := len(calculatedRoleOveralls)
		if topRoleCount > 7 {
			topRoleCount = 7
		}

		totalRoleOveralls := 0
		for i := 0; i < topRoleCount; i++ {
			totalRoleOveralls += calculatedRoleOveralls[i].Score
		}
		meanRoleBasedOverall = totalRoleOveralls / topRoleCount
	}

	// Set Overall to the mean of the top 7 role-specific scores
	// This provides a more balanced representation of the player's abilities focused on their best roles
	player.Overall = meanRoleBasedOverall

	// Note: We changed from using the mean of all role-specific overall scores
	// to using the mean of the top 7 role-specific overall scores
	// --- END: Overall Calculation ---

	// Check ALL attributes for masking and set the AttributeMasked flag
	// This is more comprehensive than only checking technical attributes above
	player.AttributeMasked = false

	// Define FM attribute keys that should be checked for masking
	fmAttributeKeys := map[string]bool{
		// Physical
		"Acc": true, "Pac": true, "Str": true, "Sta": true, "Nat": true, "Bal": true, "Jum": true, "Agi": true,
		// Mental
		"Agg": true, "Ant": true, "Bra": true, "Cmp": true, "Cnt": true, "Dec": true, "Det": true, "Fla": true,
		"Ldr": true, "OtB": true, "Pos": true, "Tea": true, "Vis": true, "Wor": true,
		// Technical
		"Cor": true, "Cro": true, "Dri": true, "Fin": true, "Fir": true, "Fre": true, "Hea": true, "Lon": true,
		"L Th": true, "Mar": true, "Pas": true, "Pen": true, "Tck": true, "Tec": true,
		// Goalkeeping
		"Aer": true, "Cmd": true, "Com": true, "Ecc": true, "Han": true, "Kic": true, "1v1": true, "Ref": true,
		"TRO": true, "Thr": true, "Pun": true,
		// Other potential FM attributes
		"Left Foot": true, "Right Foot": true,
	}

	for key, value := range player.Attributes {
		// Only check FM attributes, not performance stats
		if !fmAttributeKeys[key] {
			continue
		}

		// Check for completely masked attributes (-)
		if value == "-" {
			player.AttributeMasked = true
			break
		}
		// Check for range attributes (e.g., "10-15", "12-18")
		if strings.Contains(value, "-") {
			parts := strings.Split(value, "-")
			if len(parts) == 2 {
				// Verify both parts are numeric to ensure it's a range, not something like "Real Madrid-B"
				_, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
				_, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
				if err1 == nil && err2 == nil {
					player.AttributeMasked = true
					break
				}
			}
		}
	}
}

// RecalculatePlayerRatings recalculates all ratings for a player based on the current calculation method setting
func RecalculatePlayerRatings(player *Player) {
	// Determine if player is a goalkeeper first
	isGoalkeeper := false
	for _, posGroup := range player.PositionGroups {
		if posGroup == "Goalkeepers" {
			isGoalkeeper = true
			break
		}
	}

	// Calculate FIFA-style category stats based on player type and current setting
	if isGoalkeeper {
		// Goalkeepers get goalkeeper-specific stats
		if GetUseScaledRatings() {
			player.GK = CalculateFifaStatGo(player.NumericAttributes, "GK")
			player.DIV = CalculateFifaStatGo(player.NumericAttributes, "DIV")
			player.HAN = CalculateFifaStatGo(player.NumericAttributes, "HAN")
			player.REF = CalculateFifaStatGo(player.NumericAttributes, "REF")
			player.KIC = CalculateFifaStatGo(player.NumericAttributes, "KIC")
			player.SPD = CalculateFifaStatGo(player.NumericAttributes, "SPD")
			player.POS = CalculateFifaStatGo(player.NumericAttributes, "POS")
		} else {
			player.GK = CalculateFifaStatGoLinear(player.NumericAttributes, "GK")
			player.DIV = CalculateFifaStatGoLinear(player.NumericAttributes, "DIV")
			player.HAN = CalculateFifaStatGoLinear(player.NumericAttributes, "HAN")
			player.REF = CalculateFifaStatGoLinear(player.NumericAttributes, "REF")
			player.KIC = CalculateFifaStatGoLinear(player.NumericAttributes, "KIC")
			player.SPD = CalculateFifaStatGoLinear(player.NumericAttributes, "SPD")
			player.POS = CalculateFifaStatGoLinear(player.NumericAttributes, "POS")
		}
		// Set outfield stats to 0 for goalkeepers
		player.PAC = 0
		player.SHO = 0
		player.PAS = 0
		player.DRI = 0
		player.DEF = 0
		player.PHY = 0
	} else {
		// Outfield players get outfield stats
		if GetUseScaledRatings() {
			player.PAC = CalculateFifaStatGo(player.NumericAttributes, "PAC")
			player.SHO = CalculateFifaStatGo(player.NumericAttributes, "SHO")
			player.PAS = CalculateFifaStatGo(player.NumericAttributes, "PAS")
			player.DRI = CalculateFifaStatGo(player.NumericAttributes, "DRI")
			player.DEF = CalculateFifaStatGo(player.NumericAttributes, "DEF")
			player.PHY = CalculateFifaStatGo(player.NumericAttributes, "PHY")
		} else {
			player.PAC = CalculateFifaStatGoLinear(player.NumericAttributes, "PAC")
			player.SHO = CalculateFifaStatGoLinear(player.NumericAttributes, "SHO")
			player.PAS = CalculateFifaStatGoLinear(player.NumericAttributes, "PAS")
			player.DRI = CalculateFifaStatGoLinear(player.NumericAttributes, "DRI")
			player.DEF = CalculateFifaStatGoLinear(player.NumericAttributes, "DEF")
			player.PHY = CalculateFifaStatGoLinear(player.NumericAttributes, "PHY")
		}
		// Set goalkeeper stats to 0 for outfield players
		player.GK = 0
		player.DIV = 0
		player.HAN = 0
		player.REF = 0
		player.KIC = 0
		player.SPD = 0
		player.POS = 0
	}

	// Recalculate role-specific overalls
	maxRoleBasedOverall := 0
	bestRoleName := ""

	// Get current precomputed role weights
	muPrecomputedRoleWeights.RLock()
	currentPrecomputedWeights := precomputedRoleWeights
	muPrecomputedRoleWeights.RUnlock()

	// Clear existing role overalls and recalculate
	player.RoleSpecificOveralls = make([]RoleOverallScore, 0)

	// Process all applicable roles for this player
	processedRoleNames := make(map[string]struct{})

	for _, shortKey := range player.ShortPositions {
		if applicableRoles, found := currentPrecomputedWeights[shortKey]; found {
			for _, roleData := range applicableRoles {
				if _, alreadyProcessed := processedRoleNames[roleData.RoleName]; alreadyProcessed {
					continue
				}

				var overallForThisRole int
				if GetUseScaledRatings() {
					overallForThisRole = CalculateOverallForRoleGo(player.NumericAttributes, roleData.Weights)
				} else {
					overallForThisRole = CalculateOverallForRoleGoLinear(player.NumericAttributes, roleData.Weights)
				}

				player.RoleSpecificOveralls = append(player.RoleSpecificOveralls, RoleOverallScore{
					RoleName: roleData.RoleName,
					Score:    overallForThisRole,
				})

				if overallForThisRole > maxRoleBasedOverall {
					maxRoleBasedOverall = overallForThisRole
					bestRoleName = roleData.RoleName
				}

				processedRoleNames[roleData.RoleName] = struct{}{}
			}
		}
	}

	// Sort role-specific overalls by score (highest first)
	sort.Slice(player.RoleSpecificOveralls, func(i, j int) bool {
		if player.RoleSpecificOveralls[i].Score != player.RoleSpecificOveralls[j].Score {
			return player.RoleSpecificOveralls[i].Score > player.RoleSpecificOveralls[j].Score
		}
		return player.RoleSpecificOveralls[i].RoleName < player.RoleSpecificOveralls[j].RoleName
	})

	// Calculate overall as the mean of the top 7 role ratings instead of using all role ratings
	meanRoleBasedOverall := 0
	if len(player.RoleSpecificOveralls) > 0 {
		// Take the top 7 role ratings (or all of them if there are fewer than 7)
		topRoleCount := len(player.RoleSpecificOveralls)
		if topRoleCount > 7 {
			topRoleCount = 7
		}

		totalRoleOveralls := 0
		for i := 0; i < topRoleCount; i++ {
			totalRoleOveralls += player.RoleSpecificOveralls[i].Score
		}
		meanRoleBasedOverall = totalRoleOveralls / topRoleCount
	}

	// Update overall and best role
	player.BestRoleOverall = bestRoleName
	player.Overall = meanRoleBasedOverall

	// Check ALL attributes for masking and set the AttributeMasked flag
	// This ensures the flag is updated even during recalculations
	player.AttributeMasked = false

	// Define FM attribute keys that should be checked for masking
	fmAttributeKeys := map[string]bool{
		// Physical
		"Acc": true, "Pac": true, "Str": true, "Sta": true, "Nat": true, "Bal": true, "Jum": true, "Agi": true,
		// Mental
		"Agg": true, "Ant": true, "Bra": true, "Cmp": true, "Cnt": true, "Dec": true, "Det": true, "Fla": true,
		"Ldr": true, "OtB": true, "Pos": true, "Tea": true, "Vis": true, "Wor": true,
		// Technical
		"Cor": true, "Cro": true, "Dri": true, "Fin": true, "Fir": true, "Fre": true, "Hea": true, "Lon": true,
		"L Th": true, "Mar": true, "Pas": true, "Pen": true, "Tck": true, "Tec": true,
		// Goalkeeping
		"Aer": true, "Cmd": true, "Com": true, "Ecc": true, "Han": true, "Kic": true, "1v1": true, "Ref": true,
		"TRO": true, "Thr": true, "Pun": true,
		// Other potential FM attributes
		"Left Foot": true, "Right Foot": true,
	}

	for key, value := range player.Attributes {
		// Only check FM attributes, not performance stats
		if !fmAttributeKeys[key] {
			continue
		}

		// Check for completely masked attributes (-)
		if value == "-" {
			player.AttributeMasked = true
			break
		}
		// Check for range attributes (e.g., "10-15", "12-18")
		if strings.Contains(value, "-") {
			parts := strings.Split(value, "-")
			if len(parts) == 2 {
				// Verify both parts are numeric to ensure it's a range, not something like "Real Madrid-B"
				_, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
				_, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
				if err1 == nil && err2 == nil {
					player.AttributeMasked = true
					break
				}
			}
		}
	}
}

// RecalculateAllPlayersRatings recalculates ratings for all players in a slice
func RecalculateAllPlayersRatings(players []Player) []Player {
	recalculatedPlayers := make([]Player, len(players))
	for i := range players {
		// Make a copy to avoid modifying the original
		recalculatedPlayers[i] = players[i]
		RecalculatePlayerRatings(&recalculatedPlayers[i])
	}

	return recalculatedPlayers
}
