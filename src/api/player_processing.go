package main

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Pre-compiled regex patterns for better performance
var (
	ageRegex = regexp.MustCompile(`^\d+$`)
	rangeRegex = regexp.MustCompile(`^\d+-\d+$`)
)

// parseCellsToPlayer converts a slice of string cells (a row from the HTML table)
// into a Player struct, based on the provided headers.
func parseCellsToPlayer(cells, headers []string) (Player, error) {
	if len(headers) == 0 {
		return Player{}, errors.New("cannot process row: headers are empty")
	}

	// Ensure cells slice is at least as long as headers, padding with empty strings if necessary.
	if len(cells) < len(headers) {
		diff := len(headers) - len(cells)
		padding := make([]string, diff)
		cells = append(cells, padding...)
	}

	// Create new maps and slices for each player to avoid race conditions
	attributes := make(map[string]string, defaultAttributeCapacity)
	numericAttributes := make(map[string]int, defaultAttributeCapacity)
	performanceStatsNumeric := make(map[string]float64, len(PerformanceStatKeys))
	performancePercentiles := make(map[string]map[string]float64)
	parsedPositions := make([]string, 0, 8)
	shortPositions := make([]string, 0, 8)
	positionGroups := make([]string, 0, 4)
	roleSpecificOveralls := make([]RoleOverallScore, 0, 16)

	player := Player{
		Attributes:              attributes,
		NumericAttributes:       numericAttributes,
		PerformanceStatsNumeric: performanceStatsNumeric,
		PerformancePercentiles:  performancePercentiles,
		ParsedPositions:         parsedPositions,
		ShortPositions:          shortPositions,
		PositionGroups:          positionGroups,
		RoleSpecificOveralls:    roleSpecificOveralls,
	}

	// Headers that are known not to be player attributes (e.g., "Inf", "Rec" for recommendations).
	knownNonAttributeHeaders := map[string]bool{
		"Inf": true, "Rec": true, "Salary": true,
	}
	foundName := false

	for i, headerNameClean := range headers {
		cellValue := ""
		if i < len(cells) {
			cellValue = strings.TrimSpace(cells[i])
		}

		isAnAttributeField := true

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
			player.TransferValue, player.TransferValueAmount, _ = ParseMonetaryValueGo(cellValue)
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
		case "Nat":
			// Handle masked Natural Fitness attribute first
			if cellValue == "-" {
				player.Attributes[headerNameClean] = cellValue
				isAnAttributeField = true
			} else {
				// Use pre-compiled regex for better performance
				isValidInt := ageRegex.MatchString(cellValue)
				isValidRange := rangeRegex.MatchString(cellValue)

				if isValidInt || isValidRange {
					player.Attributes[headerNameClean] = cellValue
					isAnAttributeField = true
				} else {
					// Only process as Nationality if we haven't already set nationality
					if player.Nationality == "" && player.NationalityFIFACode == "" {
						fifaCode := strings.ToUpper(cellValue)
						player.NationalityFIFACode = fifaCode
						if fullName, ok := FifaCountryCodes[fifaCode]; ok {
							player.Nationality = fullName
						} else {
							player.Nationality = cellValue
						}
						if isoCode, ok := FifaToISO2[fifaCode]; ok {
							player.NationalityISO = isoCode
						} else {
							if len(fifaCode) >= 2 {
								player.NationalityISO = strings.ToLower(fifaCode[:2])
							} else {
								player.NationalityISO = strings.ToLower(fifaCode)
							}
						}
						isAnAttributeField = false
					} else {
						player.Attributes[headerNameClean] = cellValue
						isAnAttributeField = true
					}
				}
			}
		case "Left Foot", "Right Foot":
			if cellValue != "" && cellValue != "-" {
				player.Attributes[headerNameClean] = cellValue
			} else {
				isAnAttributeField = false
			}
		default:
			// If not explicitly handled, it's potentially an attribute or a performance stat.
		}

		if isAnAttributeField {
			if _, isKnownNonAttr := knownNonAttributeHeaders[headerNameClean]; !isKnownNonAttr {
				if headerNameClean != "" && cellValue != "" {
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
			if valStr == "-" {
				// Completely masked attribute
				player.NumericAttributes[key] = 0 // Default to 0 for completely masked
				player.AttributeMasked = true
			} else if strings.Contains(valStr, "-") {
				parts := strings.Split(valStr, "-")
				if len(parts) == 2 {
					val1, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
					val2, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))

					if err1 == nil && err2 == nil {
						player.NumericAttributes[key] = (val1 + val2) / 2
						player.AttributeMasked = true
					} else {
						player.NumericAttributes[key] = 0 // Default on parsing error
					}
				} else {
					player.NumericAttributes[key] = 0 // Default if split is not two parts
				}
			} else {
				valInt, err := strconv.Atoi(valStr)
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
						parsedValue, err = strconv.ParseFloat(minsCleaned, 64)
					case "Apps":
						// Handle appearances with substitutes like "22 (4)"
						if strings.Contains(statStrCleaned, "(") {
							// Extract main appearances and substitute appearances
							parts := strings.Split(statStrCleaned, " (")
							if len(parts) == 2 {
								mainApps, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
								subAppsStr := strings.TrimRight(strings.TrimSpace(parts[1]), ")")
								subApps, err2 := strconv.ParseFloat(subAppsStr, 64)
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
							parsedValue, err = strconv.ParseFloat(statStrCleaned, 64)
						}
					default:
						// Default parsing - remove commas for any other stats that might have them
						statStrCleaned = strings.ReplaceAll(statStrCleaned, ",", "")
						parsedValue, err = strconv.ParseFloat(statStrCleaned, 64)
					}

					if err == nil {
						player.PerformanceStatsNumeric[key] = parsedValue
					} else {
						player.PerformanceStatsNumeric[key] = 0 // Use 0 for unparseable stats
					}
				} else {
					player.PerformanceStatsNumeric[key] = 0 // Use 0 for missing stats ("-")
				}
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

	// Calculate role-specific overalls
	maxRoleBasedOverall := 0
	bestRoleName := ""

	// Get current precomputed role weights with proper synchronization
	muPrecomputedRoleWeights.RLock()
	currentPrecomputedWeights := make(map[string][]struct {
		RoleName string
		Weights  map[string]int
	}, len(precomputedRoleWeights))
	
	// Create a defensive copy to avoid race conditions
	for k, v := range precomputedRoleWeights {
		roleSlice := make([]struct {
			RoleName string
			Weights  map[string]int
		}, len(v))
		for i, role := range v {
			roleSlice[i] = struct {
				RoleName string
				Weights  map[string]int
			}{
				RoleName: role.RoleName,
				Weights:  role.Weights, // This is safe as weights maps are read-only after initialization
			}
		}
		currentPrecomputedWeights[k] = roleSlice
	}
	muPrecomputedRoleWeights.RUnlock()

	// Clear existing role overalls and recalculate
	player.RoleSpecificOveralls = make([]RoleOverallScore, 0)

	// Process all applicable roles for this player
	processedRoleNames := make(map[string]struct{})

	for _, shortKey := range player.ShortPositions {
		if applicableRoles, found := currentPrecomputedWeights[shortKey]; found {
			for _, roleData := range applicableRoles {
				if _, alreadyProcessed := processedRoleNames[roleData.RoleName]; !alreadyProcessed {
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
				if _, alreadyProcessed := processedRoleNames[roleData.RoleName]; !alreadyProcessed {
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
