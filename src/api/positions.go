package main

import (
	"sort"
	"strings"
	"sync"
)

// Cache for position group lookups to avoid repeated map iterations
var positionGroupCache = make(map[string][]string)
var positionGroupMutex sync.RWMutex

// positionRoleMapGo maps short position codes (from HTML) to more descriptive role names.
var positionRoleMapGo = map[string]string{
	"GK":  "Goalkeeper",
	"SW":  "Sweeper",
	"DC":  "Defender (Centre)",
	"DR":  "Defender (Right)",
	"DL":  "Defender (Left)",
	"WBR": "Wing-Back (Right)",
	"WBL": "Wing-Back (Left)",
	"DM":  "Defensive Midfielder (Centre)", // DM usually implies central
	"MC":  "Midfielder (Centre)",
	"MR":  "Midfielder (Right)",
	"ML":  "Midfielder (Left)",
	"AMC": "Attacking Midfielder (Centre)",
	"AMR": "Attacking Midfielder (Right)",
	"AML": "Attacking Midfielder (Left)",
	"ST":  "Striker (Centre)", // ST usually implies central
}

// standardizedPositionNameMapGo maps descriptive role names to fully standardized position names.
var standardizedPositionNameMapGo = map[string]string{
	"Goalkeeper":                    "Goalkeeper",
	"Sweeper":                       "Sweeper",
	"Defender (Centre)":             "Centre Back",
	"Defender (Right)":              "Right Back",
	"Defender (Left)":               "Left Back",
	"Wing-Back (Right)":             "Right Wing-Back",
	"Wing-Back (Left)":              "Left Wing-Back",
	"Defensive Midfielder (Centre)": "Centre Defensive Midfielder",
	"Midfielder (Centre)":           "Centre Midfielder",
	"Midfielder (Right)":            "Right Midfielder",
	"Midfielder (Left)":             "Left Midfielder",
	"Attacking Midfielder (Centre)": "Centre Attacking Midfielder",
	"Attacking Midfielder (Right)":  "Right Attacking Midfielder",
	"Attacking Midfielder (Left)":   "Left Attacking Midfielder",
	"Striker (Centre)":              "Striker",
}

// positionGroupsGo maps broad position categories to the standardized position names they encompass.
var positionGroupsGo = map[string][]string{
	"Goalkeepers": {"Goalkeeper"},
	"Defenders":   {"Sweeper", "Right Back", "Left Back", "Centre Back"},
	"Wing-Backs":  {"Right Wing-Back", "Left Wing-Back"},
	"Midfielders": {"Centre Defensive Midfielder", "Right Midfielder", "Left Midfielder", "Centre Midfielder", "Centre Attacking Midfielder", "Right Attacking Midfielder", "Left Attacking Midfielder"},
	"Attackers":   {"Striker"},
}

// parsedPositionToBaseRoleKeyGo maps standardized full position names back to their primary short codes (e.g., "GK", "DC").
// This is useful for looking up role-specific weights.
var parsedPositionToBaseRoleKeyGo = map[string]string{
	"Goalkeeper":                  "GK",
	"Sweeper":                     "SW",
	"Right Back":                  "DR",
	"Left Back":                   "DL",
	"Centre Back":                 "DC",
	"Right Wing-Back":             "WBR",
	"Left Wing-Back":              "WBL",
	"Centre Defensive Midfielder": "DM",
	"Right Midfielder":            "MR",
	"Left Midfielder":             "ML",
	"Centre Midfielder":           "MC",
	"Right Attacking Midfielder":  "AMR",
	"Left Attacking Midfielder":   "AML",
	"Centre Attacking Midfielder": "AMC",
	"Striker":                     "ST",
}

// ShortPositionDisplayOrder defines the preferred order for displaying short position codes.
var ShortPositionDisplayOrder = []string{
	"GK", "SW", "DR", "DC", "DL", "WBR", "WBL", "DM", "MR", "MC", "ML", "AMR", "AMC", "AML", "ST",
}

// ShortPositionOrderMap provides a quick lookup for sorting short positions.
var ShortPositionOrderMap = func() map[string]int {
	m := make(map[string]int)
	for i, pos := range ShortPositionDisplayOrder {
		m[pos] = i
	}
	return m
}()

// GetShortPositionKeyFromRoleName extracts the base position key (e.g., "DC")
// from a full role name (e.g., "DC - Central Defender - Defend").
// This is used in config.go during precomputation of role weights.
func GetShortPositionKeyFromRoleName(roleFullName string) string {
	parts := strings.SplitN(roleFullName, " - ", 2)
	if len(parts) > 0 {
		return strings.TrimSpace(parts[0])
	}
	return "" // Or handle error/log if appropriate
}

// ParsePlayerPositionsGo processes a raw position string (e.g., "D/M (RLC), AM (C), ST (C)")
// and returns a slice of standardized, full position names (e.g., ["Right Back", "Centre Midfielder"]).
func ParsePlayerPositionsGo(positionStr string) []string {
	if positionStr == "" {
		return []string{}
	}
	finalPositionsSet := make(map[string]struct{})
	mainParts := strings.Split(positionStr, ",")

	for _, part := range mainParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		var rolesStringSegment string
		var explicitSidesArray []string

		// Check for side indicators like (RLC), (R), (L), (C)
		sideMatchEnd := strings.LastIndex(part, ")")
		sideMatchStart := strings.LastIndex(part, "(")

		if sideMatchEnd == len(part)-1 && sideMatchStart > 0 && sideMatchStart < sideMatchEnd {
			rolesStringSegment = strings.TrimSpace(part[:sideMatchStart])
			sidesStr := part[sideMatchStart+1 : sideMatchEnd]
			for _, r := range sidesStr { // R, L, C
				explicitSidesArray = append(explicitSidesArray, string(r))
			}
		} else {
			rolesStringSegment = part // No explicit side indicators, or malformed
		}

		// Roles like D/M/AM or just D, M, AM, ST, DM, WB, GK
		individualRoleKeys := strings.Split(rolesStringSegment, "/")

		for _, roleKey := range individualRoleKeys {
			roleKey = strings.TrimSpace(roleKey)
			if roleKey == "" {
				continue
			}

			sidesToUse := explicitSidesArray
			if len(sidesToUse) == 0 { // Infer sides if not explicit
				switch roleKey {
				case "D", "M", "AM", "ST", "DM", "SW": // These are typically central if not specified
					sidesToUse = []string{"C"}
				case "WB": // Wing-backs need a side, but this case might be ambiguous without (R/L)
					// If WB is alone, it's hard to infer. Let's assume it might appear as WBR/WBL directly
					// or as WB (RL). If just "WB", it's underspecified.
					// For now, if no explicit sides, we might not be able to map "WB" alone.
					// However, the original code implies WB (C) is not a valid mapping.
					// Let's assume WB must have explicit sides or come as WBR/WBL.
					// If sidesToUse is empty here for WB, it won't match positionRoleMapGo.
					continue // Skip if WB has no explicit side
				case "GK":
					sidesToUse = []string{""} // GK has no side
				default:
					// If it's already a full key like DR, DL, MC, STC (though STC isn't standard here)
					// we can try a direct lookup.
					// This part of the logic relies on positionRoleMapGo having direct keys.
					sidesToUse = []string{""} // Treat as a direct key attempt
				}
			}

			for _, sideKey := range sidesToUse {
				var mapLookupKey string
				if sideKey == "" { // For GK or direct keys like "DC"
					mapLookupKey = roleKey
				} else {
					// Construct keys like D R, D L, D C, M R, M L, M C, AM R, AM L, AM C
					// Or DM C, WB R, WB L, ST C, SW C
					switch {
					case (roleKey == "D" || roleKey == "M" || roleKey == "AM") && (sideKey == "R" || sideKey == "L" || sideKey == "C"):
						mapLookupKey = roleKey + sideKey
					case roleKey == "DM" && sideKey == "C":
						mapLookupKey = "DM" // DM is implicitly DM C in positionRoleMapGo
					case roleKey == "WB" && (sideKey == "R" || sideKey == "L"):
						mapLookupKey = "WB" + sideKey
					case roleKey == "ST" && sideKey == "C":
						mapLookupKey = "ST" // ST is implicitly ST C
					case roleKey == "SW" && sideKey == "C":
						mapLookupKey = "SW" // SW is implicitly SW C
					default:
						// Fallback for simple roleKey + sideKey if it's a direct match in positionRoleMapGo
						// e.g. if positionRoleMapGo had "DR", "DL" etc.
						mapLookupKey = roleKey + sideKey
					}
				}
				// Ensure we use the correct keys for positionRoleMapGo
				// e.g. D C -> DC, M R -> MR
				switch mapLookupKey {
				case "STC": // ST (C) becomes ST
					mapLookupKey = "ST"
				case "DMC": // DM (C) becomes DM
					mapLookupKey = "DM"
				}
				// Valid keys like DR, DL, DC, MR, ML, MC, AR, AL, AC (though AR/AL/AC not in map)
				// mapLookupKey is already properly set above

				roleFullName, roleExists := positionRoleMapGo[mapLookupKey]
				if roleExists {
					standardizedName, stdOk := standardizedPositionNameMapGo[roleFullName]
					if stdOk {
						finalPositionsSet[standardizedName] = struct{}{}
					} else {
						// This case should ideally not happen if maps are consistent
						// but as a fallback, if roleFullName is already a standardized name.
						isAlreadyStandard := false
						for _, groupPositions := range positionGroupsGo {
							for _, p := range groupPositions {
								if p == roleFullName {
									isAlreadyStandard = true
									break
								}
							}
							if isAlreadyStandard {
								break
							}
						}
						if isAlreadyStandard {
							finalPositionsSet[roleFullName] = struct{}{}
						}
					}
				}
			}
		}
	}

	finalPositions := make([]string, 0, len(finalPositionsSet))
	for pos := range finalPositionsSet {
		finalPositions = append(finalPositions, pos)
	}
	sort.Strings(finalPositions) // Ensure consistent order
	return finalPositions
}

// GetPlayerPositionGroupsGo determines the broad position groups (e.g., "Defenders", "Midfielders")
// a player belongs to based on their standardized parsed positions.
func GetPlayerPositionGroupsGo(parsedPositionsArray []string) []string {
	if len(parsedPositionsArray) == 0 {
		return []string{}
	}

	// Create a cache key from sorted positions
	sort.Strings(parsedPositionsArray) // Ensure consistent key generation
	cacheKey := strings.Join(parsedPositionsArray, "|")

	// Check cache first
	positionGroupMutex.RLock()
	if cached, found := positionGroupCache[cacheKey]; found {
		positionGroupMutex.RUnlock()
		return cached
	}
	positionGroupMutex.RUnlock()

	// Calculate if not in cache
	groupsSet := make(map[string]struct{})
	for _, pos := range parsedPositionsArray { // e.g., pos = "Centre Back"
		for groupName, groupPositions := range positionGroupsGo { // e.g., groupName = "Defenders", groupPositions = ["Sweeper", "Right Back", ...]
			for _, p := range groupPositions {
				if p == pos {
					groupsSet[groupName] = struct{}{}
					break // Found in this group, move to next parsedPos or groupName
				}
			}
		}
	}
	groups := make([]string, 0, len(groupsSet))
	for group := range groupsSet {
		groups = append(groups, group)
	}
	sort.Strings(groups) // Ensure consistent order

	// Cache the result
	positionGroupMutex.Lock()
	positionGroupCache[cacheKey] = groups
	positionGroupMutex.Unlock()

	return groups
}
