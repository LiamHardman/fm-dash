package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "net/http/pprof"

	"github.com/google/uuid"
	"golang.org/x/net/html"
)

const (
	defaultPlayerCapacity    = 1024
	defaultAttributeCapacity = 64
	defaultCellCapacity      = 64
	overallScalingFactor     = 5.85
	maxTokenBufferSize       = 2 * 1024 * 1024 // 2MB
)

// --- START: Struct Definitions ---

type RoleOverallScore struct {
	RoleName string `json:"roleName"`
	Score    int    `json:"score"`
}

type Player struct {
	Name                    string                        `json:"name"`
	Position                string                        `json:"position"`
	Age                     string                        `json:"age"`
	Club                    string                        `json:"club"`
	TransferValue           string                        `json:"transfer_value"`
	Wage                    string                        `json:"wage"`
	Personality             string                        `json:"personality,omitempty"`
	MediaHandling           string                        `json:"media_handling,omitempty"`
	Nationality             string                        `json:"nationality"`
	NationalityISO          string                        `json:"nationality_iso"`
	NationalityFIFACode     string                        `json:"nationality_fifa_code"`
	Attributes              map[string]string             `json:"attributes"`
	NumericAttributes       map[string]int                `json:"-"` // Not serialized, used for calculations
	PerformanceStatsNumeric map[string]float64            `json:"-"` // For pre-parsed numeric performance stats
	PerformancePercentiles  map[string]map[string]float64 `json:"performancePercentiles"`
	ParsedPositions         []string                      `json:"parsedPositions"`
	ShortPositions          []string                      `json:"shortPositions"`
	PositionGroups          []string                      `json:"positionGroups"` // Broad groups like "Defenders", "Midfielders"
	PHY                     int                           `json:"PHY"`
	SHO                     int                           `json:"SHO"`
	PAS                     int                           `json:"PAS"`
	DRI                     int                           `json:"DRI"`
	DEF                     int                           `json:"DEF"`
	MEN                     int                           `json:"MEN"`
	GK                      int                           `json:"GK,omitempty"`
	Overall                 int                           `json:"Overall"`
	RoleSpecificOveralls    []RoleOverallScore            `json:"roleSpecificOveralls"`
	TransferValueAmount     int64                         `json:"transferValueAmount"`
	WageAmount              int64                         `json:"wageAmount"`
}

type PlayerParseResult struct {
	Player Player
	Err    error
}

type UploadResponse struct {
	DatasetID              string `json:"datasetId"`
	Message                string `json:"message"`
	DetectedCurrencySymbol string `json:"detectedCurrencySymbol,omitempty"`
}

type PlayerDataWithCurrency struct {
	Players        []Player `json:"players"`
	CurrencySymbol string   `json:"currencySymbol"`
}

// --- END: Struct Definitions ---

var (
	playerDataStore = make(map[string]struct {
		Players        []Player
		CurrencySymbol string
	})
	storeMutex                   sync.RWMutex
	attributeWeights             map[string]map[string]int
	roleSpecificOverallWeights   map[string]map[string]int
	muAttributeWeights           sync.RWMutex
	muRoleSpecificOverallWeights sync.RWMutex
	precomputedRoleWeights       map[string][]struct {
		RoleName string
		Weights  map[string]int
	}
	muPrecomputedRoleWeights sync.RWMutex
)

var performanceStatKeys = []string{
	"Asts/90", "Av Rat", "Blk/90", "Ch C/90", "Clr/90", "Cr C/90", "Drb/90",
	"xA/90", "xG/90", "Gls/90", "Hdrs W/90", "Int/90", "K Ps/90", "Ps C/90",
	"Shot/90", "Tck/90", "Poss Won/90", "ShT/90", "Pres C/90", "Poss Lost/90",
	"Pr passes/90", "Conv %", "Tck R", "Pas %", "Cr C/A",
}

// Existing broad position groups for percentiles
var positionGroupsForPercentiles = []string{"Goalkeepers", "Defenders", "Midfielders", "Attackers"}

// NEW: Detailed position groups for more granular percentile comparisons
// These keys will be used in the Player.PerformancePercentiles map
var detailedPositionGroupsForPercentiles = map[string][]string{
	"Full-backs":                      {"DR", "DL"},
	"Centre-backs":                    {"DC"},
	"Wing-backs":                      {"WBR", "WBL"},
	"Defensive Midfielders":           {"DM"},
	"Central Midfielders":             {"MC"},
	"Wide Midfielders":                {"MR", "ML"},
	"Attacking Midfielders (Central)": {"AMC"},
	"Wingers":                         {"AMR", "AML"},
	"Strikers":                        {"ST"},
	// Note: "Goalkeepers" is already handled by the broad positionGroupsForPercentiles.
	// If a specific "Goalkeepers (Detailed)" is needed, it would map to {"GK"}.
	// For now, the existing "Goalkeepers" group (mapping to "Goalkeeper" full name) should suffice.
}

var defaultAttributeWeightsGo = map[string]map[string]int{
	"PHY": {"Acc": 7, "Pac": 6, "Str": 5, "Sta": 4, "Nat": 3, "Bal": 2, "Jum": 1},
	"SHO": {"Fin": 7, "OtB": 6, "Cmp": 5, "Tec": 4, "Hea": 3, "Lon": 2, "Pen": 1},
	"PAS": {"Pas": 7, "Vis": 6, "Tec": 5, "Cro": 4, "Fre": 3, "Cor": 2, "L Th": 1},
	"DRI": {"Dri": 6, "Fir": 5, "Tec": 4, "Agi": 3, "Bal": 2, "Fla": 1},
	"DEF": {"Tck": 6, "Mar": 5, "Hea": 4, "Pos": 3, "Cnt": 2, "Ant": 1},
	"MEN": {"Wor": 11, "Dec": 10, "Tea": 9, "Det": 8, "Bra": 7, "Ldr": 6, "Vis": 5, "Agg": 4, "OtB": 3, "Pos": 2, "Ant": 1},
	"GK":  {"Han": 20, "Ref": 20, "Cmd": 15, "Aer": 15, "1v1": 10, "Kic": 5, "TRO": 5, "Com": 3, "Thr": 3, "Ecc": 1},
}

var defaultRoleSpecificOverallWeightsGo = map[string]map[string]int{
	"DC - Generic Defender":    {"Mar": 80, "Hea": 50, "Tck": 50, "Pos": 80, "Str": 60, "Pac": 50, "Acc": 60, "Jum": 60, "Cnt": 40, "Cmp": 20, "Bra": 20, "Ant": 50, "Fir": 20, "Pas": 20, "Tec": 10, "Wor": 20, "Ldr": 20, "Dec": 10, "Vis": 10, "OtB": 10, "Agi": 60, "Bal": 20, "Sta": 30, "Cor": 10, "Cro": 10, "Dri": 10, "Fin": 10, "Fre": 10, "Lon": 10, "L Th": 10, "Pen": 10, "Agg": 0, "Det": 0, "Fla": 0, "Nat": 0},
	"ST - Generic Striker":     {"Fin": 80, "Fir": 60, "OtB": 60, "Cmp": 60, "Hea": 60, "Acc": 100, "Pac": 70, "Str": 60, "Jum": 50, "Tec": 40, "Ant": 50, "Dec": 50, "Dri": 50, "Wor": 20, "Sta": 60, "Cor": 10, "Cro": 20, "Fre": 10, "Lon": 20, "L Th": 10, "Mar": 10, "Pas": 20, "Pen": 10, "Tck": 10, "Agg": 0, "Bra": 10, "Cnt": 20, "Det": 0, "Fla": 0, "Ldr": 10, "Pos": 20, "Tea": 10, "Vis": 20, "Agi": 60, "Bal": 20, "Nat": 0},
	"GK - Goalkeeper - Defend": {"Han": 90, "Ref": 90, "Aer": 80, "Cmd": 75, "1v1": 80, "Cnt": 70, "Dec": 70, "Pos": 75, "Ant": 60, "Cmp": 60, "Bra": 60, "Com": 50, "Kic": 40, "Thr": 40, "TRO": 30, "Det": 50, "Ldr": 40, "Wor": 40, "Tea": 40, "Agi": 50, "Jum": 60, "Str": 50, "Acc": 30, "Pac": 30, "Ecc": 10},
}

func loadJSONWeights(filePath string, defaultWeights map[string]map[string]int) (map[string]map[string]int, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Warning: Could not read %s: %v. Using default weights.", filePath, err)
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

func init() {
	var errAttr, errRole error
	loadedAttrWeights, errAttr := loadJSONWeights(filepath.Join("public", "attribute_weights.json"), defaultAttributeWeightsGo)
	if errAttr != nil {
		log.Printf("Using default attribute_weights due to error: %v. Default attribute_weights has %d entries.", errAttr, len(defaultAttributeWeightsGo))
		attributeWeights = make(map[string]map[string]int)
		for k, v := range defaultAttributeWeightsGo {
			innerMap := make(map[string]int)
			for ik, iv := range v {
				innerMap[ik] = iv
			}
			attributeWeights[k] = innerMap
		}
	} else {
		attributeWeights = loadedAttrWeights
	}

	loadedRoleWeights, errRole := loadJSONWeights(filepath.Join("public", "role_specific_overall_weights.json"), defaultRoleSpecificOverallWeightsGo)
	if errRole != nil {
		log.Printf("Using default role_specific_overall_weights due to error: %v. Default role_specific_overall_weights has %d entries.", errRole, len(defaultRoleSpecificOverallWeightsGo))
		roleSpecificOverallWeights = make(map[string]map[string]int)
		for k, v := range defaultRoleSpecificOverallWeightsGo {
			innerMap := make(map[string]int)
			for ik, iv := range v {
				innerMap[ik] = iv
			}
			roleSpecificOverallWeights[k] = innerMap
		}
	} else {
		roleSpecificOverallWeights = loadedRoleWeights
	}

	muPrecomputedRoleWeights.Lock()
	precomputedRoleWeights = make(map[string][]struct {
		RoleName string
		Weights  map[string]int
	})
	sourceWeightsToPrecompute := roleSpecificOverallWeights
	for roleFullName, weights := range sourceWeightsToPrecompute {
		parts := strings.SplitN(roleFullName, " - ", 2)
		if len(parts) > 0 {
			shortKey := strings.TrimSpace(parts[0])
			copiedWeights := make(map[string]int, len(weights))
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
}

var (
	positionRoleMapGo = map[string]string{
		"GK": "Goalkeeper", "SW": "Sweeper", "DC": "Defender (Centre)", "DR": "Defender (Right)", "DL": "Defender (Left)",
		"WBR": "Wing-Back (Right)", "WBL": "Wing-Back (Left)", "DM": "Defensive Midfielder (Centre)", "MC": "Midfielder (Centre)",
		"MR": "Midfielder (Right)", "ML": "Midfielder (Left)", "AMC": "Attacking Midfielder (Centre)",
		"AMR": "Attacking Midfielder (Right)", "AML": "Attacking Midfielder (Left)", "ST": "Striker (Centre)",
	}
	standardizedPositionNameMapGo = map[string]string{
		"Goalkeeper": "Goalkeeper", "Sweeper": "Sweeper",
		"Defender (Centre)": "Centre Back", "Defender (Right)": "Right Back", "Defender (Left)": "Left Back",
		"Wing-Back (Right)": "Right Wing-Back", "Wing-Back (Left)": "Left Wing-Back",
		"Defensive Midfielder (Centre)": "Centre Defensive Midfielder",
		"Midfielder (Centre)":           "Centre Midfielder", "Midfielder (Right)": "Right Midfielder", "Midfielder (Left)": "Left Midfielder",
		"Attacking Midfielder (Centre)": "Centre Attacking Midfielder", "Attacking Midfielder (Right)": "Right Attacking Midfielder", "Attacking Midfielder (Left)": "Left Attacking Midfielder",
		"Striker (Centre)": "Striker",
	}
	positionGroupsGo = map[string][]string{ // Broad groups, player belongs if any of their standardized long names match
		"Goalkeepers": {"Goalkeeper"},
		"Defenders":   {"Sweeper", "Right Back", "Left Back", "Centre Back"}, // Includes Full-backs and Centre-backs
		"Wing-Backs":  {"Right Wing-Back", "Left Wing-Back"},                 // This is specific
		"Midfielders": {"Centre Defensive Midfielder", "Right Midfielder", "Left Midfielder", "Centre Midfielder", "Centre Attacking Midfielder", "Right Attacking Midfielder", "Left Attacking Midfielder"},
		"Attackers":   {"Striker"}, // Broad, might include wingers if they are parsed as "Attacking Midfielder (Right/Left)"
	}
	parsedPositionToBaseRoleKeyGo = map[string]string{
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
	nullString = ""
)

var shortPositionDisplayOrder = []string{
	"GK", "SW", "DR", "DC", "DL", "WBR", "WBL", "DM", "MR", "MC", "ML", "AMR", "AMC", "AML", "ST",
}
var shortPositionOrderMap = func() map[string]int {
	m := make(map[string]int)
	for i, pos := range shortPositionDisplayOrder {
		m[pos] = i
	}
	return m
}()

func parsePlayerPositionsGo(positionStr string) []string {
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

		sideMatchEnd := strings.LastIndex(part, ")")
		sideMatchStart := strings.LastIndex(part, "(")

		if sideMatchEnd == len(part)-1 && sideMatchStart > 0 && sideMatchStart < sideMatchEnd {
			rolesStringSegment = strings.TrimSpace(part[:sideMatchStart])
			sidesStr := part[sideMatchStart+1 : sideMatchEnd]
			for _, r := range sidesStr {
				explicitSidesArray = append(explicitSidesArray, string(r))
			}
		} else {
			rolesStringSegment = part
		}

		individualRoleKeys := strings.Split(rolesStringSegment, "/")

		for _, roleKey := range individualRoleKeys {
			roleKey = strings.TrimSpace(roleKey)
			if roleKey == "" {
				continue
			}

			sidesToUse := explicitSidesArray
			if len(sidesToUse) == 0 {
				switch roleKey {
				case "D", "M", "AM", "ST", "DM", "WB", "SW":
					sidesToUse = []string{"C"}
				case "GK":
					sidesToUse = []string{""}
				default:
					sidesToUse = []string{""}
				}
			}

			for _, sideKey := range sidesToUse {
				var mapLookupKey string
				if sideKey == "" {
					mapLookupKey = roleKey
				} else {
					if (roleKey == "D" || roleKey == "M" || roleKey == "AM") && sideKey == "C" {
						mapLookupKey = roleKey + "C"
					} else if roleKey == "DM" && sideKey == "C" {
						mapLookupKey = "DM"
					} else if roleKey == "WB" && (sideKey == "R" || sideKey == "L") {
						mapLookupKey = "WB" + sideKey
					} else if roleKey == "ST" && sideKey == "C" {
						mapLookupKey = "ST"
					} else if roleKey == "SW" && sideKey == "C" {
						mapLookupKey = "SW"
					} else {
						mapLookupKey = roleKey + sideKey
					}
				}

				roleFullName, roleExists := positionRoleMapGo[mapLookupKey]
				if roleExists {
					standardizedName, stdOk := standardizedPositionNameMapGo[roleFullName]
					if stdOk {
						finalPositionsSet[standardizedName] = struct{}{}
					} else {
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
	sort.Strings(finalPositions)
	return finalPositions
}

func getPlayerPositionGroupsGo(parsedPositionsArray []string) []string {
	groupsSet := make(map[string]struct{})
	if len(parsedPositionsArray) == 0 {
		return []string{}
	}
	for _, pos := range parsedPositionsArray {
		for groupName, groupPositions := range positionGroupsGo {
			for _, p := range groupPositions {
				if p == pos {
					groupsSet[groupName] = struct{}{}
					break
				}
			}
		}
	}
	groups := make([]string, 0, len(groupsSet))
	for group := range groupsSet {
		groups = append(groups, group)
	}
	sort.Strings(groups)
	return groups
}

func calculateFifaStatGo(playerNumericAttributes map[string]int, categoryName string) int {
	muAttributeWeights.RLock()
	var currentCategoryWeightsSource map[string]map[string]int
	if attributeWeights != nil {
		currentCategoryWeightsSource = attributeWeights
	} else {
		log.Printf("Warning: attributeWeights is nil in calculateFifaStatGo. Using default for %s.", categoryName)
		currentCategoryWeightsSource = defaultAttributeWeightsGo
	}
	muAttributeWeights.RUnlock()

	categoryAttributeWeights, ok := currentCategoryWeightsSource[categoryName]
	if !ok {
		categoryAttributeWeights, ok = defaultAttributeWeightsGo[categoryName]
		if !ok {
			log.Printf("Error: Default attribute weights for category '%s' also not found. Returning 0.", categoryName)
			return 0
		}
	}

	var weightedSum float64
	var totalWeightOfPresentAttributes float64

	for attrName, attrWeight := range categoryAttributeWeights {
		attrValue, exists := playerNumericAttributes[attrName]
		if exists {
			if attrValue >= 1 && attrValue <= 20 {
				weightedSum += float64(attrValue * attrWeight)
				totalWeightOfPresentAttributes += float64(attrWeight)
			}
		}
	}

	if totalWeightOfPresentAttributes == 0 {
		return 0
	}
	weightedAverage := weightedSum / totalWeightOfPresentAttributes
	return int(math.Round(weightedAverage * 5.3))
}

func calculateOverallForRoleGo(playerNumericAttributes map[string]int, roleSpecificAttrWeights map[string]int) int {
	if len(roleSpecificAttrWeights) == 0 {
		return 0
	}
	var weightedAttributeSum float64
	var totalApplicableWeightsSum float64
	for attrKey, weightForAttribute := range roleSpecificAttrWeights {
		attributeValue, exists := playerNumericAttributes[attrKey]
		if exists {
			validAttributeValue := math.Max(0, math.Min(20, float64(attributeValue)))
			weightedAttributeSum += validAttributeValue * float64(weightForAttribute)
			totalApplicableWeightsSum += float64(weightForAttribute)
		}
	}
	if totalApplicableWeightsSum == 0 {
		return 0
	}
	rawPositionalOverall := weightedAttributeSum / totalApplicableWeightsSum
	return int(math.Min(99, math.Round(rawPositionalOverall*overallScalingFactor)))
}

var fifaCountryCodes = map[string]string{
	"ENG": "England", "GER": "Germany", "ESP": "Spain", "ITA": "Italy", "FRA": "France", "NED": "Netherlands", "POR": "Portugal", "BEL": "Belgium", "ARG": "Argentina", "BRA": "Brazil", "URU": "Uruguay", "COL": "Colombia", "CHI": "Chile", "MEX": "Mexico", "USA": "United States", "CAN": "Canada", "JPN": "Japan", "KOR": "South Korea", "AUS": "Australia", "CRO": "Croatia", "SUI": "Switzerland", "SWE": "Sweden", "NOR": "Norway", "DEN": "Denmark", "POL": "Poland", "AUT": "Austria", "TUR": "Turkey", "RUS": "Russia", "UKR": "Ukraine", "SRB": "Serbia", "SCO": "Scotland", "WAL": "Wales", "NIR": "Northern Ireland", "IRL": "Republic of Ireland", "CZE": "Czech Republic", "SVK": "Slovakia", "HUN": "Hungary", "ROU": "Romania", "GRE": "Greece", "EGY": "Egypt", "NGA": "Nigeria", "SEN": "Senegal", "CIV": "Ivory Coast", "GHA": "Ghana", "CMR": "Cameroon", "MAR": "Morocco", "ALG": "Algeria", "TUN": "Tunisia",
}
var fifaToISO2 = map[string]string{
	"ENG": "gb-eng", "GER": "de", "ESP": "es", "ITA": "it", "FRA": "fr", "NED": "nl", "POR": "pt", "BEL": "be", "ARG": "ar", "BRA": "br", "URU": "uy", "COL": "co", "CHI": "cl", "MEX": "mx", "USA": "us", "CAN": "ca", "JPN": "jp", "KOR": "kr", "AUS": "au", "CRO": "hr", "SUI": "ch", "SWE": "se", "NOR": "no", "DEN": "dk", "POL": "pl", "AUT": "at", "TUR": "tr", "RUS": "ru", "UKR": "ua", "SRB": "rs", "SCO": "gb-sct", "WAL": "gb-wls", "NIR": "gb-nir", "IRL": "ie", "CZE": "cz", "SVK": "sk", "HUN": "hu", "ROU": "ro", "GRE": "gr", "EGY": "eg", "NGA": "ng", "SEN": "sn", "CIV": "ci", "GHA": "gh", "CMR": "cm", "MAR": "ma", "ALG": "dz", "TUN": "tn",
}
var currencySymbolRegex = regexp.MustCompile(`([€£$])`)

func parseMonetaryValueGo(rawValue string) (originalDisplay string, numericValue int64, detectedSymbol string) {
	cleanedValue := strings.TrimSpace(rawValue)
	originalDisplay = cleanedValue

	matches := currencySymbolRegex.FindStringSubmatch(cleanedValue)
	if len(matches) > 1 {
		detectedSymbol = matches[1]
	} else {
		detectedSymbol = ""
	}

	if strings.Contains(cleanedValue, " - ") {
		parts := strings.Split(cleanedValue, " - ")
		if len(parts) == 2 {
			cleanedValue = strings.TrimSpace(parts[1])
			symbolMatchesRange := currencySymbolRegex.FindStringSubmatch(cleanedValue)
			if len(symbolMatchesRange) > 1 {
				detectedSymbol = symbolMatchesRange[1]
			}
		}
	}

	cleanedValue = strings.ReplaceAll(cleanedValue, "€", "")
	cleanedValue = strings.ReplaceAll(cleanedValue, "£", "")
	cleanedValue = strings.ReplaceAll(cleanedValue, "$", "")
	cleanedValue = strings.TrimSpace(strings.ReplaceAll(cleanedValue, "p/w", ""))
	cleanedValue = strings.TrimSpace(strings.ReplaceAll(cleanedValue, "/w", ""))

	multiplier := int64(1)
	if strings.HasSuffix(cleanedValue, "M") || strings.HasSuffix(cleanedValue, "m") {
		multiplier = 1000000
		cleanedValue = strings.TrimRight(cleanedValue, "Mm")
	} else if strings.HasSuffix(cleanedValue, "K") || strings.HasSuffix(cleanedValue, "k") {
		multiplier = 1000
		cleanedValue = strings.TrimRight(cleanedValue, "Kk")
	}

	cleanedValue = strings.ReplaceAll(cleanedValue, ",", "")

	valFloat, err := strconv.ParseFloat(strings.TrimSpace(cleanedValue), 64)
	if err == nil {
		numericValue = int64(valFloat * float64(multiplier))
	} else {
		numericValue = 0
	}

	return originalDisplay, numericValue, detectedSymbol
}

func calculatePercentileValue(value float64, sortedValues []float64) float64 {
	if len(sortedValues) == 0 {
		return -1
	}
	if len(sortedValues) == 1 && sortedValues[0] == value {
		return 50.0
	}

	countSmaller := sort.SearchFloat64s(sortedValues, value)
	endRangeIndex := sort.Search(len(sortedValues), func(i int) bool { return sortedValues[i] > value })
	countEqual := endRangeIndex - countSmaller

	if countSmaller >= len(sortedValues) || (countSmaller < len(sortedValues) && sortedValues[countSmaller] != value) {
		countEqual = 0
	}

	percentile := (float64(countSmaller) + (0.5 * float64(countEqual))) / float64(len(sortedValues)) * 100.0
	return math.Round(percentile)
}

func calculatePlayerPerformancePercentiles(players []Player) {
	if len(players) == 0 {
		return
	}

	// Initialize PerformancePercentiles maps for all players
	for i := range players {
		if players[i].PerformancePercentiles == nil {
			players[i].PerformancePercentiles = make(map[string]map[string]float64)
		}
		if players[i].PerformancePercentiles["Global"] == nil {
			players[i].PerformancePercentiles["Global"] = make(map[string]float64)
		}
	}

	// --- Global Percentiles ---
	for _, statKey := range performanceStatKeys {
		allStatValues := make([]float64, 0, len(players))
		for i := range players {
			val, ok := players[i].PerformanceStatsNumeric[statKey]
			if ok && !math.IsNaN(val) {
				allStatValues = append(allStatValues, val)
			}
		}

		if len(allStatValues) == 0 {
			for i := range players {
				players[i].PerformancePercentiles["Global"][statKey] = -1
			}
			continue
		}
		sort.Float64s(allStatValues)

		for i := range players {
			val, ok := players[i].PerformanceStatsNumeric[statKey]
			if ok && !math.IsNaN(val) {
				players[i].PerformancePercentiles["Global"][statKey] = calculatePercentileValue(val, allStatValues)
			} else {
				players[i].PerformancePercentiles["Global"][statKey] = -1
			}
		}
	}

	// --- Broad Positional Group Percentiles (Optimized) ---
	// This uses player.PositionGroups (e.g., "Defenders", "Midfielders")
	groupStatValueLists := make(map[string]map[string][]float64) // groupName -> statKey -> []values

	for _, groupName := range positionGroupsForPercentiles {
		groupStatValueLists[groupName] = make(map[string][]float64)
		for _, statKey := range performanceStatKeys {
			groupStatValueLists[groupName][statKey] = make([]float64, 0, len(players)/len(positionGroupsForPercentiles))
		}
	}

	for i := range players {
		player := &players[i]
		for _, pg := range player.PositionGroups { // pg is like "Defenders", "Midfielders"
			if _, ok := groupStatValueLists[pg]; ok {
				for _, statKey := range performanceStatKeys {
					val, statOk := player.PerformanceStatsNumeric[statKey]
					if statOk && !math.IsNaN(val) {
						groupStatValueLists[pg][statKey] = append(groupStatValueLists[pg][statKey], val)
					}
				}
			}
		}
	}

	for _, groupName := range positionGroupsForPercentiles {
		for i := range players {
			if players[i].PerformancePercentiles[groupName] == nil {
				players[i].PerformancePercentiles[groupName] = make(map[string]float64)
			}
		}

		for _, statKey := range performanceStatKeys {
			groupValues := groupStatValueLists[groupName][statKey]

			if len(groupValues) == 0 {
				for i := range players {
					isPlayerInGroup := false
					for _, pg := range players[i].PositionGroups {
						if pg == groupName {
							isPlayerInGroup = true
							break
						}
					}
					if isPlayerInGroup {
						players[i].PerformancePercentiles[groupName][statKey] = -1
					}
				}
				continue
			}
			sort.Float64s(groupValues)

			for i := range players {
				player := &players[i]
				isPlayerInGroup := false
				for _, pg := range player.PositionGroups {
					if pg == groupName {
						isPlayerInGroup = true
						break
					}
				}

				if isPlayerInGroup {
					val, statOk := player.PerformanceStatsNumeric[statKey]
					if statOk && !math.IsNaN(val) {
						player.PerformancePercentiles[groupName][statKey] = calculatePercentileValue(val, groupValues)
					} else {
						player.PerformancePercentiles[groupName][statKey] = -1
					}
				}
			}
		}
	}

	// --- DETAILED Positional Group Percentiles (NEW) ---
	// This uses player.ShortPositions (e.g., "DC", "ST", "AMR")
	for detailedGroupName, shortPositionsInGroup := range detailedPositionGroupsForPercentiles {
		// Ensure PerformancePercentiles map for this detailed group exists for all players
		for i := range players {
			if players[i].PerformancePercentiles[detailedGroupName] == nil {
				players[i].PerformancePercentiles[detailedGroupName] = make(map[string]float64)
			}
		}

		// Prepare data structures for this specific detailed group
		currentDetailedGroupStatValues := make(map[string][]float64) // statKey -> []values
		for _, statKey := range performanceStatKeys {
			currentDetailedGroupStatValues[statKey] = make([]float64, 0, len(players)/10) // Estimate
		}

		// Populate stat value lists for this detailed group
		playerIndicesInDetailedGroup := []int{} // Store indices of players in this group

		for i := range players {
			player := &players[i]
			isPlayerInThisDetailedGroup := false
			for _, playerShortPos := range player.ShortPositions {
				for _, requiredShortPos := range shortPositionsInGroup {
					if playerShortPos == requiredShortPos {
						isPlayerInThisDetailedGroup = true
						break
					}
				}
				if isPlayerInThisDetailedGroup {
					break
				}
			}

			if isPlayerInThisDetailedGroup {
				playerIndicesInDetailedGroup = append(playerIndicesInDetailedGroup, i)
				for _, statKey := range performanceStatKeys {
					val, statOk := player.PerformanceStatsNumeric[statKey]
					if statOk && !math.IsNaN(val) {
						currentDetailedGroupStatValues[statKey] = append(currentDetailedGroupStatValues[statKey], val)
					}
				}
			}
		}

		// Calculate and assign percentiles for this detailed group
		for _, statKey := range performanceStatKeys {
			statValuesForCalc := currentDetailedGroupStatValues[statKey]

			if len(statValuesForCalc) == 0 {
				// If no values for this stat in this group, assign -1 to players belonging to this group
				for _, playerIndex := range playerIndicesInDetailedGroup {
					players[playerIndex].PerformancePercentiles[detailedGroupName][statKey] = -1
				}
				continue
			}
			sort.Float64s(statValuesForCalc) // Sort the collected values for this stat and group

			for _, playerIndex := range playerIndicesInDetailedGroup {
				player := &players[playerIndex]
				val, statOk := player.PerformanceStatsNumeric[statKey]
				if statOk && !math.IsNaN(val) {
					player.PerformancePercentiles[detailedGroupName][statKey] = calculatePercentileValue(val, statValuesForCalc)
				} else {
					player.PerformancePercentiles[detailedGroupName][statKey] = -1
				}
			}
		}
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}
	startTime := time.Now()

	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB
		http.Error(w, "Error parsing multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("playerFile")
	if err != nil {
		http.Error(w, "Error retrieving the file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileSize := handler.Size
	log.Printf("Uploaded File: %s (Size: %d bytes)", handler.Filename, fileSize)

	parseStartTime := time.Now()

	bufferedReader := bufio.NewReaderSize(file, maxTokenBufferSize)
	tokenizer := html.NewTokenizer(bufferedReader)

	var headers []string
	playersList := make([]Player, 0, defaultPlayerCapacity)
	var processingError error

	numWorkers := runtime.NumCPU()
	if numWorkers == 0 {
		numWorkers = 1
	}
	const rowBufferMultiplier = 10
	rowCellsChan := make(chan []string, numWorkers*rowBufferMultiplier)
	resultsChan := make(chan PlayerParseResult, numWorkers*rowBufferMultiplier)
	var wg sync.WaitGroup

	var currentCells []string
	inHeaderRow := false
	inDataRow := false
	inTable := false
	inTHead := false
	inTBody := false
	var cellBuilder strings.Builder

	var headersSnapshot []string
	workersStarted := false

	doneConsumingResults := make(chan struct{})
	go func() {
		defer close(doneConsumingResults)
		for result := range resultsChan {
			if result.Err == nil {
				playersList = append(playersList, result.Player)
			} else {
				log.Printf("Skipping row due to error from worker: %v", result.Err)
			}
		}
		log.Println("Finished collecting results from resultsChan.")
	}()

tokenLoop:
	for {
		tt := tokenizer.Next()
		token := tokenizer.Token()

		switch tt {
		case html.ErrorToken:
			err := tokenizer.Err()
			if err == io.EOF {
				if inDataRow && len(currentCells) > 0 && workersStarted {
					cellsCopy := make([]string, len(currentCells))
					copy(cellsCopy, currentCells)
					rowCellsChan <- cellsCopy
				}
				break tokenLoop
			}
			log.Printf("HTML tokenization error: %v", err)
			processingError = errors.New("Error tokenizing HTML: " + err.Error())
			break tokenLoop
		case html.StartTagToken:
			switch token.Data {
			case "table":
				inTable = true
			case "thead":
				if inTable {
					inTHead = true
				}
			case "tbody":
				if inTable {
					inTBody = true
					if !workersStarted && len(headers) > 0 {
						headersSnapshot = make([]string, len(headers))
						copy(headersSnapshot, headers)
						log.Printf("Headers found (tbody start), launching %d workers: %v", numWorkers, headersSnapshot)
						wg.Add(numWorkers)
						for i := 0; i < numWorkers; i++ {
							go playerParserWorker(i, rowCellsChan, resultsChan, &wg, headersSnapshot)
						}
						workersStarted = true
					}
				}
			case "tr":
				currentCells = make([]string, 0, defaultCellCapacity)
				if inTHead || (inTable && !inTBody && !workersStarted) {
					inHeaderRow = true
				} else if inTBody || (inTable && !inTHead && workersStarted) {
					inDataRow = true
				} else if inTable && !inTHead && !inTBody && workersStarted {
					inDataRow = true
				}
			case "th":
				if inHeaderRow {
					cellBuilder.Reset()
				} else if inDataRow {
					cellBuilder.Reset()
				}
			case "td":
				if inDataRow {
					cellBuilder.Reset()
				}
			}
		case html.TextToken:
			if inHeaderRow || inDataRow {
				cellBuilder.WriteString(token.Data)
			}
		case html.EndTagToken:
			switch token.Data {
			case "th":
				if inHeaderRow {
					headers = append(headers, strings.TrimSpace(cellBuilder.String()))
					cellBuilder.Reset()
				} else if inDataRow {
					currentCells = append(currentCells, strings.TrimSpace(cellBuilder.String()))
					cellBuilder.Reset()
				}
			case "td":
				if inDataRow {
					currentCells = append(currentCells, strings.TrimSpace(cellBuilder.String()))
					cellBuilder.Reset()
				}
			case "tr":
				if inHeaderRow {
					if len(headers) > 0 && !workersStarted {
						headersSnapshot = make([]string, len(headers))
						copy(headersSnapshot, headers)
						log.Printf("Headers found (end of header tr), launching %d workers: %v", numWorkers, headersSnapshot)
						wg.Add(numWorkers)
						for i := 0; i < numWorkers; i++ {
							go playerParserWorker(i, rowCellsChan, resultsChan, &wg, headersSnapshot)
						}
						workersStarted = true
					}
					inHeaderRow = false
				} else if inDataRow {
					if len(currentCells) > 0 && workersStarted {
						cellsCopy := make([]string, len(currentCells))
						copy(cellsCopy, currentCells)
						rowCellsChan <- cellsCopy
					}
					inDataRow = false
				}
			case "thead":
				inTHead = false
				if len(headers) > 0 && !workersStarted {
					headersSnapshot = make([]string, len(headers))
					copy(headersSnapshot, headers)
					log.Printf("Headers found (thead end), launching %d workers: %v", numWorkers, headersSnapshot)
					wg.Add(numWorkers)
					for i := 0; i < numWorkers; i++ {
						go playerParserWorker(i, rowCellsChan, resultsChan, &wg, headersSnapshot)
					}
					workersStarted = true
				}
			case "tbody":
				inTBody = false
			case "table":
				inTable = false
			}
		}
	}
	close(rowCellsChan)
	log.Println("Row cells channel closed (tokenizer finished).")

	if processingError != nil {
		http.Error(w, processingError.Error(), http.StatusInternalServerError)
		if workersStarted {
			wg.Wait()
		}
		close(resultsChan)
		<-doneConsumingResults
		return
	}

	if !workersStarted && len(headers) > 0 {
		headersSnapshot = make([]string, len(headers))
		copy(headersSnapshot, headers)
		log.Printf("Headers found (fallback after token loop), launching %d workers: %v", numWorkers, headersSnapshot)
		wg.Add(numWorkers)
		for i := 0; i < numWorkers; i++ {
			go playerParserWorker(i, rowCellsChan, resultsChan, &wg, headersSnapshot)
		}
		workersStarted = true
	}

	if !workersStarted {
		log.Println("Critical: Workers were not started (no headers found or other parsing issue).")
		close(resultsChan)
		if len(headers) == 0 {
			http.Error(w, "Could not parse table headers, no data processed.", http.StatusInternalServerError)
			<-doneConsumingResults
			return
		}
	}

	log.Println("Waiting for workers to finish...")
	wg.Wait()
	log.Println("All workers have completed (wg.Wait() returned).")

	close(resultsChan)
	log.Println("ResultsChan closed after all workers finished.")

	<-doneConsumingResults
	log.Println("Results consumer goroutine finished processing all items.")

	finalDatasetCurrencySymbol := "$"
	if len(playersList) > 0 {
		var foundSymbol bool
		for _, p := range playersList {
			_, _, tvSymbol := parseMonetaryValueGo(p.TransferValue)
			if tvSymbol != "" {
				finalDatasetCurrencySymbol = tvSymbol
				foundSymbol = true
				break
			}
			_, _, wSymbol := parseMonetaryValueGo(p.Wage)
			if wSymbol != "" {
				finalDatasetCurrencySymbol = wSymbol
				foundSymbol = true
				break
			}
		}
		if !foundSymbol {
			log.Println("No currency symbol detected from parsed players, using default '$'.")
		}
	}

	if len(playersList) > 0 {
		log.Println("Calculating player performance percentiles...")
		calculatePlayerPerformancePercentiles(playersList) // Pass by value, modifications happen on elements
		log.Println("Finished calculating percentiles.")
	}

	parseDuration := time.Since(parseStartTime)
	datasetID := uuid.New().String()

	storeMutex.Lock()
	playerDataStore[datasetID] = struct {
		Players        []Player
		CurrencySymbol string
	}{Players: playersList, CurrencySymbol: finalDatasetCurrencySymbol}
	storeMutex.Unlock()

	log.Printf("Stored %d players with DatasetID: %s. Detected Currency: %s", len(playersList), datasetID, finalDatasetCurrencySymbol)
	if len(playersList) > 0 {
		log.Printf("DEBUG: Sample Player 1 after all processing: Name='%s', Overall=%d, ParsedPositions=%v, ShortPositions=%v, PositionGroups=%v", playersList[0].Name, playersList[0].Overall, playersList[0].ParsedPositions, playersList[0].ShortPositions, playersList[0].PositionGroups)
		if pp, ok := playersList[0].PerformancePercentiles["Global"]; ok && len(pp) > 0 {
			log.Printf("DEBUG: Sample Player 1 Performance Percentile Keys (Global and Detailed): %v", getMapKeys(playersList[0].PerformancePercentiles))
		}
	} else {
		log.Println("No players were successfully parsed or list is empty.")
	}

	response := UploadResponse{DatasetID: datasetID, Message: "File uploaded and parsed successfully.", DetectedCurrencySymbol: finalDatasetCurrencySymbol}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding JSON: "+err.Error(), http.StatusInternalServerError)
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	rowsPerSecond := 0.0
	if parseDuration.Seconds() > 0 {
		rowsPerSecond = float64(len(playersList)) / parseDuration.Seconds()
	}
	totalDuration := time.Since(startTime)
	log.Printf("--- Perf Metrics (Streaming) --- File: %s, Size: %d KB, Total Time: %v, Parse Time: %v, Parsed Players: %d, Rows/Sec: %.2f, MemAlloc: %.2f MiB, Workers: %d, Goroutines: %d ---",
		handler.Filename, fileSize/1024, totalDuration, parseDuration, len(playersList), rowsPerSecond, bToMb(memStats.Alloc), numWorkers, runtime.NumGoroutine())
}

func playerParserWorker(workerID int, rowCellsChan <-chan []string, resultsChan chan<- PlayerParseResult, wg *sync.WaitGroup, headers []string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Worker %d PANICKED: %v", workerID, r)
		}
		wg.Done()
	}()
	if len(headers) == 0 {
		log.Printf("Worker %d started with NO headers. Draining rowCellsChan and exiting.", workerID)
		for range rowCellsChan {
		}
		return
	}
	for cells := range rowCellsChan {
		player, err := parseCellsToPlayer(cells, headers)
		if err == nil {
			enhancePlayerWithCalculations(&player) // Pass pointer to modify original player
		}
		resultsChan <- PlayerParseResult{Player: player, Err: err}
	}
}

func getMapKeys(m map[string]map[string]float64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func playerDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/players/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		http.Error(w, "Dataset ID is missing", http.StatusBadRequest)
		return
	}
	datasetID := pathParts[0]

	storeMutex.RLock()
	data, found := playerDataStore[datasetID]
	storeMutex.RUnlock()

	if !found {
		log.Printf("Player data not found for DatasetID: %s", datasetID)
		http.Error(w, "Player data not found for the given ID. It might have expired or the ID is incorrect.", http.StatusNotFound)
		return
	}

	response := PlayerDataWithCurrency{Players: data.Players, CurrencySymbol: data.CurrencySymbol}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON for dataset %s: %v", datasetID, err)
	}
}

func enhancePlayerWithCalculations(player *Player) {
	if player.NumericAttributes == nil {
		player.NumericAttributes = make(map[string]int, len(player.Attributes))
	}
	if player.PerformanceStatsNumeric == nil {
		player.PerformanceStatsNumeric = make(map[string]float64, len(performanceStatKeys))
	}

	for key, valStr := range player.Attributes {
		switch key {
		case "Acc", "Pac", "Str", "Sta", "Nat", "Bal", "Jum",
			"Fin", "OtB", "Cmp", "Tec", "Hea", "Lon", "Pen",
			"Pas", "Vis", "Cro", "Fre", "Cor", "L Th",
			"Dri", "Fir", "Agi", "Fla",
			"Tck", "Mar", "Pos", "Cnt", "Ant",
			"Wor", "Dec", "Tea", "Det", "Bra", "Ldr", "Agg",
			"Han", "Ref", "Cmd", "Aer", "1v1", "Kic", "TRO", "Com", "Thr", "Ecc", "Pun":
			valInt, err := strconv.Atoi(valStr)
			if err == nil {
				player.NumericAttributes[key] = valInt
			} else {
				player.NumericAttributes[key] = 0
			}
		default:
		}
	}

	for _, statKey := range performanceStatKeys {
		statStr, ok := player.Attributes[statKey]
		if ok && statStr != "-" && statStr != "" {
			statStrCleaned := strings.ReplaceAll(statStr, "%", "")
			if val, err := strconv.ParseFloat(statStrCleaned, 64); err == nil {
				player.PerformanceStatsNumeric[statKey] = val
			} else {
				player.PerformanceStatsNumeric[statKey] = math.NaN()
			}
		} else {
			player.PerformanceStatsNumeric[statKey] = math.NaN()
		}
	}

	player.ParsedPositions = parsePlayerPositionsGo(player.Position)
	player.PositionGroups = getPlayerPositionGroupsGo(player.ParsedPositions) // For broad groups

	shortPosSet := make(map[string]struct{})
	for _, pPos := range player.ParsedPositions {
		if shortKey, ok := parsedPositionToBaseRoleKeyGo[pPos]; ok && shortKey != nullString {
			shortPosSet[shortKey] = struct{}{}
		} else if pPos == "Goalkeeper" {
			shortPosSet["GK"] = struct{}{}
		}
	}
	player.ShortPositions = make([]string, 0, len(shortPosSet))
	for sp := range shortPosSet {
		player.ShortPositions = append(player.ShortPositions, sp)
	}
	sort.Slice(player.ShortPositions, func(i, j int) bool {
		orderI, okI := shortPositionOrderMap[player.ShortPositions[i]]
		orderJ, okJ := shortPositionOrderMap[player.ShortPositions[j]]
		if !okI {
			orderI = len(shortPositionDisplayOrder) + i
		}
		if !okJ {
			orderJ = len(shortPositionDisplayOrder) + j
		}
		return orderI < orderJ
	})

	player.PHY = calculateFifaStatGo(player.NumericAttributes, "PHY")
	player.SHO = calculateFifaStatGo(player.NumericAttributes, "SHO")
	player.PAS = calculateFifaStatGo(player.NumericAttributes, "PAS")
	player.DRI = calculateFifaStatGo(player.NumericAttributes, "DRI")
	player.DEF = calculateFifaStatGo(player.NumericAttributes, "DEF")
	player.MEN = calculateFifaStatGo(player.NumericAttributes, "MEN")

	isGoalkeeper := false
	for _, posGroup := range player.PositionGroups { // Check broad group
		if posGroup == "Goalkeepers" {
			isGoalkeeper = true
			break
		}
	}
	if isGoalkeeper {
		player.GK = calculateFifaStatGo(player.NumericAttributes, "GK")
	} else {
		player.GK = 0
	}

	maxOverall := 0
	calculatedRoleOveralls := make([]RoleOverallScore, 0, 5)

	muPrecomputedRoleWeights.RLock()
	currentPrecomputedWeights := precomputedRoleWeights
	muPrecomputedRoleWeights.RUnlock()

	if len(currentPrecomputedWeights) == 0 && len(roleSpecificOverallWeights) > 0 {
		log.Printf("Warning: precomputedRoleWeights is empty for player '%s'. Falling back to iterating roleSpecificOverallWeights (SLOW PATH).", player.Name)
		muRoleSpecificOverallWeights.RLock()
		fallbackWeights := roleSpecificOverallWeights
		muRoleSpecificOverallWeights.RUnlock()

		if len(player.ParsedPositions) > 0 {
			foundAnyRoleMatch := false
			processedRoleNames := make(map[string]struct{})

			for _, parsedPos := range player.ParsedPositions {
				shortKey, ok := parsedPositionToBaseRoleKeyGo[parsedPos]
				if !ok || shortKey == nullString {
					if parsedPos == "Goalkeeper" {
						shortKey = "GK"
					} else {
						continue
					}
				}

				for roleKeyInJsonLoop, specificWeightsLoop := range fallbackWeights {
					if strings.HasPrefix(roleKeyInJsonLoop, shortKey+" - ") || (shortKey == "GK" && roleKeyInJsonLoop == "GK - Goalkeeper - Defend") {
						if _, alreadyProcessed := processedRoleNames[roleKeyInJsonLoop]; alreadyProcessed {
							continue
						}
						foundAnyRoleMatch = true
						overallForThisRole := calculateOverallForRoleGo(player.NumericAttributes, specificWeightsLoop)
						calculatedRoleOveralls = append(calculatedRoleOveralls, RoleOverallScore{RoleName: roleKeyInJsonLoop, Score: overallForThisRole})
						if overallForThisRole > maxOverall {
							maxOverall = overallForThisRole
						}
						processedRoleNames[roleKeyInJsonLoop] = struct{}{}
					}
				}
			}
			if !foundAnyRoleMatch {
				log.Printf("Fallback Warning: Player '%s' with ParsedPositions %v found no matching roles in fallback roleSpecificOverallWeights. MaxOverall will be 0.", player.Name, player.ParsedPositions)
			}
		}
	} else if len(player.ParsedPositions) > 0 {
		foundAnyRoleMatch := false
		processedRoleNames := make(map[string]struct{})

		for _, parsedPos := range player.ParsedPositions {
			shortKey, ok := parsedPositionToBaseRoleKeyGo[parsedPos]
			if !ok || shortKey == nullString {
				if parsedPos == "Goalkeeper" {
					shortKey = "GK"
				} else {
					continue
				}
			}

			if applicableRoles, found := currentPrecomputedWeights[shortKey]; found {
				for _, roleData := range applicableRoles {
					if _, alreadyProcessed := processedRoleNames[roleData.RoleName]; alreadyProcessed {
						continue
					}
					foundAnyRoleMatch = true
					overallForThisRole := calculateOverallForRoleGo(player.NumericAttributes, roleData.Weights)
					calculatedRoleOveralls = append(calculatedRoleOveralls, RoleOverallScore{RoleName: roleData.RoleName, Score: overallForThisRole})
					if overallForThisRole > maxOverall {
						maxOverall = overallForThisRole
					}
					processedRoleNames[roleData.RoleName] = struct{}{}
				}
			}
		}
		if !foundAnyRoleMatch {
			log.Printf("Warning: Player '%s' with ParsedPositions %v (ShortPositions: %v) found no matching roles in precomputedRoleWeights. MaxOverall will be 0.", player.Name, player.ParsedPositions, player.ShortPositions)
		}
	} else {
		log.Printf("Warning: Player '%s' has no ParsedPositions. MaxOverall will be 0.", player.Name)
	}

	player.Overall = maxOverall
	player.RoleSpecificOveralls = calculatedRoleOveralls
	sort.Slice(player.RoleSpecificOveralls, func(i, j int) bool {
		if player.RoleSpecificOveralls[i].Score != player.RoleSpecificOveralls[j].Score {
			return player.RoleSpecificOveralls[i].Score > player.RoleSpecificOveralls[j].Score
		}
		return player.RoleSpecificOveralls[i].RoleName < player.RoleSpecificOveralls[j].RoleName
	})
}

func parseCellsToPlayer(cells []string, headers []string) (Player, error) {
	if len(headers) == 0 {
		return Player{}, errors.New("cannot process row: headers are empty")
	}

	if len(cells) < len(headers) {
		diff := len(headers) - len(cells)
		cells = append(cells, make([]string, diff)...)
	}

	player := Player{
		Attributes:              make(map[string]string, defaultAttributeCapacity),
		PerformanceStatsNumeric: make(map[string]float64, len(performanceStatKeys)),
		PerformancePercentiles:  make(map[string]map[string]float64),
	}

	knownNonAttributeHeaders := map[string]bool{
		"Inf": true, "Rec": true,
	}
	foundName := false

	for i, headerNameClean := range headers {
		cellValue := ""
		if i < len(cells) {
			cellValue = strings.TrimSpace(cells[i])
		}

		isAnAttributeField := true

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
		case "Transfer Value":
			player.TransferValue, player.TransferValueAmount, _ = parseMonetaryValueGo(cellValue)
			isAnAttributeField = false
		case "Wage":
			player.Wage, player.WageAmount, _ = parseMonetaryValueGo(cellValue)
			isAnAttributeField = false
		case "Personality":
			player.Personality = cellValue
			isAnAttributeField = false
		case "Media Handling":
			player.MediaHandling = cellValue
			isAnAttributeField = false
		case "Nat":
			valInt, err := strconv.Atoi(cellValue)
			if err == nil && valInt >= 1 && valInt <= 20 {
				player.Attributes[headerNameClean] = cellValue
			} else {
				fifaCode := strings.ToUpper(cellValue)
				player.NationalityFIFACode = fifaCode
				if fullName, ok := fifaCountryCodes[fifaCode]; ok {
					player.Nationality = fullName
				} else {
					player.Nationality = cellValue
				}
				if isoCode, ok := fifaToISO2[fifaCode]; ok {
					player.NationalityISO = isoCode
				} else {
					if len(fifaCode) >= 2 {
						player.NationalityISO = strings.ToLower(fifaCode[:2])
					} else {
						player.NationalityISO = strings.ToLower(fifaCode)
					}
				}
				isAnAttributeField = false
			}
		case "Left Foot", "Right Foot":
			isAnAttributeField = false
		default:
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
		isPotentiallyMeaningfulRow := false
		for _, cellContent := range cells {
			if strings.TrimSpace(cellContent) != "" {
				isPotentiallyMeaningfulRow = true
				break
			}
		}
		if isPotentiallyMeaningfulRow {
			return Player{}, errors.New("skipped row: 'Name' field is missing or empty, but other data present. First few cells: " + strings.Join(getFirstNCells(cells, 5), ", "))
		}
		return Player{}, errors.New("skipped row: 'Name' field missing and row appears empty or is likely a non-player row")
	}

	return player, nil
}

func getFirstNCells(slice []string, n int) []string {
	if n < 0 {
		n = 0
	}
	if n > len(slice) {
		n = len(slice)
	}
	return slice[:n]
}
func bToMb(b uint64) float64 { return float64(b) / 1024 / 1024 }

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(".", "index.html"))
	})

	fsPublic := http.FileServer(http.Dir("./public"))
	http.Handle("/public/", http.StripPrefix("/public/", fsPublic))

	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/api/players/", playerDataHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8091"
	}

	log.Printf("Server starting on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
