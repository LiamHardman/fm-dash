package main

import (
	"encoding/json"
	"errors"
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
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid" // For generating unique IDs
	"golang.org/x/net/html"
)

const (
	defaultPlayerCapacity    = 1024
	defaultAttributeCapacity = 64
	defaultCellCapacity      = 64
	overallScalingFactor     = 5.85
)

// --- START: Struct Definitions ---

type RoleOverallScore struct {
	RoleName string `json:"roleName"` // Will store keys like "DC - Ball Playing Defender"
	Score    int    `json:"score"`
}

type Player struct {
	Name                   string                        `json:"name"`
	Position               string                        `json:"position"`
	Age                    string                        `json:"age"`
	Club                   string                        `json:"club"`
	TransferValue          string                        `json:"transfer_value"`
	Wage                   string                        `json:"wage"`
	Personality            string                        `json:"personality,omitempty"`
	MediaHandling          string                        `json:"media_handling,omitempty"`
	Nationality            string                        `json:"nationality"`
	NationalityISO         string                        `json:"nationality_iso"`
	NationalityFIFACode    string                        `json:"nationality_fifa_code"`
	Attributes             map[string]string             `json:"attributes"`
	NumericAttributes      map[string]int                `json:"-"`
	PerformancePercentiles map[string]map[string]float64 `json:"performancePercentiles"`
	ParsedPositions        []string                      `json:"parsedPositions"`
	ShortPositions         []string                      `json:"shortPositions"`
	PositionGroups         []string                      `json:"positionGroups"`
	PHY                    int                           `json:"PHY"`
	SHO                    int                           `json:"SHO"`
	PAS                    int                           `json:"PAS"`
	DRI                    int                           `json:"DRI"`
	DEF                    int                           `json:"DEF"`
	MEN                    int                           `json:"MEN"`
	GK                     int                           `json:"GK,omitempty"`
	Overall                int                           `json:"Overall"`
	RoleSpecificOveralls   []RoleOverallScore            `json:"roleSpecificOveralls"`
	TransferValueAmount    int64                         `json:"transferValueAmount"`
	WageAmount             int64                         `json:"wageAmount"`
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
	roleSpecificOverallWeights   map[string]map[string]int // Keys will be like "DC - Ball Playing Defender"
	muAttributeWeights           sync.RWMutex
	muRoleSpecificOverallWeights sync.RWMutex
)

var performanceStatKeys = []string{
	"Asts/90", "Av Rat", "Blk/90", "Ch C/90", "Clr/90", "Cr C/90", "Drb/90",
	"xA/90", "xG/90", "Gls/90", "Hdrs W/90", "Int/90", "K Ps/90", "Ps C/90",
	"Shot/90", "Tck/90", "Poss Won/90", "ShT/90", "Pres C/90", "Poss Lost/90",
	"Pr passes/90", "Conv %", "Tck R", "Pas %", "Cr C/A",
}

var positionGroupsForPercentiles = []string{"Goalkeepers", "Defenders", "Midfielders", "Attackers"}

var defaultAttributeWeightsGo = map[string]map[string]int{
	"PHY": {"Acc": 7, "Pac": 6, "Str": 5, "Sta": 4, "Nat": 3, "Bal": 2, "Jum": 1},
	"SHO": {"Fin": 7, "OtB": 6, "Cmp": 5, "Tec": 4, "Hea": 3, "Lon": 2, "Pen": 1},
	"PAS": {"Pas": 7, "Vis": 6, "Tec": 5, "Cro": 4, "Fre": 3, "Cor": 2, "L Th": 1},
	"DRI": {"Dri": 6, "Fir": 5, "Tec": 4, "Agi": 3, "Bal": 2, "Fla": 1},
	"DEF": {"Tck": 6, "Mar": 5, "Hea": 4, "Pos": 3, "Cnt": 2, "Ant": 1},
	"MEN": {"Wor": 11, "Dec": 10, "Tea": 9, "Det": 8, "Bra": 7, "Ldr": 6, "Vis": 5, "Agg": 4, "OtB": 3, "Pos": 2, "Ant": 1},
	"GK":  {"Han": 20, "Ref": 20, "Cmd": 15, "Aer": 15, "1v1": 10, "Kic": 5, "TRO": 5, "Com": 3, "Thr": 3, "Ecc": 1},
}

// MODIFIED: Default role specific overall weights with SHORT position prefix and FULL role name
var defaultRoleSpecificOverallWeightsGo = map[string]map[string]int{
	"DC - Generic Defender":    {"Mar": 80, "Hea": 50, "Tck": 50, "Pos": 80, "Str": 60, "Pac": 50, "Acc": 60, "Jum": 60, "Cnt": 40, "Cmp": 20, "Bra": 20, "Ant": 50, "Fir": 20, "Pas": 20, "Tec": 10, "Wor": 20, "Ldr": 20, "Dec": 10, "Vis": 10, "OtB": 10, "Agi": 60, "Bal": 20, "Sta": 30, "Cor": 10, "Cro": 10, "Dri": 10, "Fin": 10, "Fre": 10, "Lon": 10, "L Th": 10, "Pen": 10, "Agg": 0, "Det": 0, "Fla": 0, "Nat": 0},
	"ST - Generic Striker":     {"Fin": 80, "Fir": 60, "OtB": 60, "Cmp": 60, "Hea": 60, "Acc": 100, "Pac": 70, "Str": 60, "Jum": 50, "Tec": 40, "Ant": 50, "Dec": 50, "Dri": 50, "Wor": 20, "Sta": 60, "Cor": 10, "Cro": 20, "Fre": 10, "Lon": 20, "L Th": 10, "Mar": 10, "Pas": 20, "Pen": 10, "Tck": 10, "Agg": 0, "Bra": 10, "Cnt": 20, "Det": 0, "Fla": 0, "Ldr": 10, "Pos": 20, "Tea": 10, "Vis": 20, "Agi": 60, "Bal": 20, "Nat": 0},
	"GK - Goalkeeper - Defend": {"Han": 90, "Ref": 90, "Aer": 80, "Cmd": 75, "1v1": 80, "Cnt": 70, "Dec": 70, "Pos": 75, "Ant": 60, "Cmp": 60, "Bra": 60, "Com": 50, "Kic": 40, "Thr": 40, "TRO": 30, "Det": 50, "Ldr": 40, "Wor": 40, "Tea": 40, "Agi": 50, "Jum": 60, "Str": 50, "Acc": 30, "Pac": 30, "Ecc": 10},
}

func loadJSONWeights(filePath string, defaultWeights map[string]map[string]int) (map[string]map[string]int, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Warning: Could not read %s: %v. Using default weights.", filePath, err)
		return defaultWeights, err
	}
	var weights map[string]map[string]int
	if err := json.Unmarshal(data, &weights); err != nil {
		log.Printf("Warning: Could not unmarshal %s: %v. Using default weights.", filePath, err)
		return defaultWeights, err
	}
	if len(weights) == 0 {
		log.Printf("Warning: Weights file %s was loaded but is empty. Using default weights.", filePath)
		return defaultWeights, errors.New("loaded weights file is empty")
	}
	log.Printf("Successfully loaded weights from %s with %d entries.", filePath, len(weights))
	return weights, nil
}

func init() {
	var errAttr, errRole error
	attributeWeights, errAttr = loadJSONWeights(filepath.Join("public", "attribute_weights.json"), defaultAttributeWeightsGo)
	if errAttr != nil {
		log.Printf("Using default attribute_weights due to error: %v. Default attribute_weights has %d entries.", errAttr, len(defaultAttributeWeightsGo))
	}
	roleSpecificOverallWeights, errRole = loadJSONWeights(filepath.Join("public", "role_specific_overall_weights.json"), defaultRoleSpecificOverallWeightsGo)
	if errRole != nil {
		log.Printf("Using default role_specific_overall_weights due to error: %v. Default role_specific_overall_weights has %d entries.", errRole, len(defaultRoleSpecificOverallWeightsGo))
	}
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
	positionGroupsGo = map[string][]string{
		"Goalkeepers": {"Goalkeeper"},
		"Defenders":   {"Sweeper", "Right Back", "Left Back", "Centre Back"},
		"Wing-Backs":  {"Right Wing-Back", "Left Wing-Back"},
		"Midfielders": {"Centre Defensive Midfielder", "Right Midfielder", "Left Midfielder", "Centre Midfielder", "Centre Attacking Midfielder", "Right Attacking Midfielder", "Left Attacking Midfielder"},
		"Attackers":   {"Striker"},
	}
	// parsedPositionToBaseRoleKeyGo maps standardized full names (e.g., "Centre Back") to short codes (e.g., "DC")
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
				if roleKey == "D" || roleKey == "M" || roleKey == "AM" || roleKey == "ST" || roleKey == "DM" || roleKey == "WB" || roleKey == "SW" {
					sidesToUse = []string{"C"}
				} else if roleKey == "GK" {
					sidesToUse = []string{""}
				} else {
					sidesToUse = []string{""}
				}
			}

			for _, sideKey := range sidesToUse {
				var mapLookupKey string
				if sideKey == "" {
					mapLookupKey = roleKey
				} else {
					if (roleKey == "D" || roleKey == "M" || roleKey == "AM" || roleKey == "ST" || roleKey == "DM" || roleKey == "WB" || roleKey == "SW") && sideKey == "C" {
						mapLookupKey = roleKey
						if roleKey == "D" {
							mapLookupKey = "DC"
						}
						if roleKey == "M" {
							mapLookupKey = "MC"
						}
						if roleKey == "AM" {
							mapLookupKey = "AMC"
						}
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
	currentCategoryWeightsSource := attributeWeights
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

func getNodeTextOptimized(n *html.Node) string {
	if n == nil {
		return ""
	}
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(getNodeTextOptimized(c))
		if c.Type == html.ElementNode && c.NextSibling != nil {
			lastChar, _ := utf8.DecodeLastRuneInString(sb.String())
			if !unicode.IsSpace(lastChar) && c.NextSibling.Type != html.TextNode {
				sb.WriteByte(' ')
			}
		} else if c.Type == html.TextNode && c.NextSibling != nil && c.NextSibling.Type == html.ElementNode {
			lastChar, _ := utf8.DecodeLastRuneInString(sb.String())
			if !unicode.IsSpace(lastChar) {
				sb.WriteByte(' ')
			}
		}
	}
	return strings.Join(strings.Fields(sb.String()), " ")
}
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
	valFloat, err := strconv.ParseFloat(cleanedValue, 64)
	if err == nil {
		numericValue = int64(valFloat * float64(multiplier))
	} else {
		numericValue = 0
	}
	return originalDisplay, numericValue, detectedSymbol
}

func calculatePlayerPerformancePercentiles(players []Player) {
	if len(players) == 0 {
		return
	}
	for i := range players {
		if players[i].PerformancePercentiles == nil {
			players[i].PerformancePercentiles = make(map[string]map[string]float64)
		}
	}
	for _, statKey := range performanceStatKeys {
		var allStatValues []float64
		for _, p := range players {
			if statStr, ok := p.Attributes[statKey]; ok && statStr != "-" && statStr != "" {
				statStrCleaned := strings.ReplaceAll(statStr, "%", "")
				if val, err := strconv.ParseFloat(statStrCleaned, 64); err == nil {
					allStatValues = append(allStatValues, val)
				}
			}
		}
		if len(allStatValues) == 0 {
			for i := range players {
				if players[i].PerformancePercentiles["Global"] == nil {
					players[i].PerformancePercentiles["Global"] = make(map[string]float64)
				}
				players[i].PerformancePercentiles["Global"][statKey] = -1
			}
			continue
		}
		sort.Float64s(allStatValues)
		for i := range players {
			if players[i].PerformancePercentiles["Global"] == nil {
				players[i].PerformancePercentiles["Global"] = make(map[string]float64)
			}
			currentPlayerStatStr, ok := players[i].Attributes[statKey]
			if !ok || currentPlayerStatStr == "-" || currentPlayerStatStr == "" {
				players[i].PerformancePercentiles["Global"][statKey] = -1
				continue
			}
			currentPlayerStatStrCleaned := strings.ReplaceAll(currentPlayerStatStr, "%", "")
			currentValue, err := strconv.ParseFloat(currentPlayerStatStrCleaned, 64)
			if err != nil {
				players[i].PerformancePercentiles["Global"][statKey] = -1
				continue
			}
			countSmaller, countEqual := 0, 0
			for _, v := range allStatValues {
				if v < currentValue {
					countSmaller++
				} else if v == currentValue {
					countEqual++
				}
			}
			var percentile float64
			if len(allStatValues) == 1 && countEqual == 1 {
				percentile = 50.0
			} else if len(allStatValues) > 0 {
				percentile = (float64(countSmaller) + (0.5 * float64(countEqual))) / float64(len(allStatValues)) * 100.0
			} else {
				percentile = -1
			}
			players[i].PerformancePercentiles["Global"][statKey] = math.Round(percentile)
		}
	}
	for _, groupName := range positionGroupsForPercentiles {
		var groupPlayers []Player
		for _, p := range players {
			isPlayerInGroup := false
			for _, pg := range p.PositionGroups {
				if pg == groupName {
					isPlayerInGroup = true
					break
				}
			}
			if isPlayerInGroup {
				groupPlayers = append(groupPlayers, p)
			}
		}
		if len(groupPlayers) == 0 {
			continue
		}
		for _, statKey := range performanceStatKeys {
			var groupStatValues []float64
			for _, p := range groupPlayers {
				if statStr, ok := p.Attributes[statKey]; ok && statStr != "-" && statStr != "" {
					statStrCleaned := strings.ReplaceAll(statStr, "%", "")
					if val, err := strconv.ParseFloat(statStrCleaned, 64); err == nil {
						groupStatValues = append(groupStatValues, val)
					}
				}
			}
			if len(groupStatValues) == 0 {
				for i := range players {
					isPlayerInCurrentProcessingGroup := false
					for _, pg := range players[i].PositionGroups {
						if pg == groupName {
							isPlayerInCurrentProcessingGroup = true
							break
						}
					}
					if isPlayerInCurrentProcessingGroup {
						if players[i].PerformancePercentiles[groupName] == nil {
							players[i].PerformancePercentiles[groupName] = make(map[string]float64)
						}
						players[i].PerformancePercentiles[groupName][statKey] = -1
					}
				}
				continue
			}
			sort.Float64s(groupStatValues)
			for i := range players {
				isPlayerInCurrentProcessingGroup := false
				for _, pg := range players[i].PositionGroups {
					if pg == groupName {
						isPlayerInCurrentProcessingGroup = true
						break
					}
				}
				if !isPlayerInCurrentProcessingGroup {
					continue
				}
				if players[i].PerformancePercentiles[groupName] == nil {
					players[i].PerformancePercentiles[groupName] = make(map[string]float64)
				}
				currentPlayerStatStr, ok := players[i].Attributes[statKey]
				if !ok || currentPlayerStatStr == "-" || currentPlayerStatStr == "" {
					players[i].PerformancePercentiles[groupName][statKey] = -1
					continue
				}
				currentPlayerStatStrCleaned := strings.ReplaceAll(currentPlayerStatStr, "%", "")
				currentValue, err := strconv.ParseFloat(currentPlayerStatStrCleaned, 64)
				if err != nil {
					players[i].PerformancePercentiles[groupName][statKey] = -1
					continue
				}
				countSmaller, countEqual := 0, 0
				for _, v := range groupStatValues {
					if v < currentValue {
						countSmaller++
					} else if v == currentValue {
						countEqual++
					}
				}
				var percentile float64
				if len(groupStatValues) == 1 && countEqual == 1 {
					percentile = 50.0
				} else if len(groupStatValues) > 0 {
					percentile = (float64(countSmaller) + (0.5 * float64(countEqual))) / float64(len(groupStatValues)) * 100.0
				} else {
					percentile = -1
				}
				players[i].PerformancePercentiles[groupName][statKey] = math.Round(percentile)
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
	if err := r.ParseMultipartForm(10 << 20); err != nil {
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
	doc, err := html.Parse(file)
	if err != nil {
		http.Error(w, "Error parsing HTML: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var tableNode *html.Node
	var findTable func(*html.Node)
	findTable = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			tableNode = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if tableNode != nil {
				return
			}
			findTable(c)
		}
	}
	findTable(doc)
	if tableNode == nil {
		http.Error(w, "No table found in the HTML", http.StatusInternalServerError)
		return
	}
	var headers []string
	var headerRowNode *html.Node
	var findHeaderRow func(n *html.Node) bool
	findHeaderRow = func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "thead" {
			for tr := n.FirstChild; tr != nil; tr = tr.NextSibling {
				if tr.Type == html.ElementNode && tr.Data == "tr" {
					isHeader := false
					tempHeaders := make([]string, 0, defaultCellCapacity)
					for cell := tr.FirstChild; cell != nil; cell = cell.NextSibling {
						if cell.Type == html.ElementNode && cell.Data == "th" {
							isHeader = true
							tempHeaders = append(tempHeaders, getNodeTextOptimized(cell))
						}
					}
					if isHeader && len(tempHeaders) > 0 {
						headers = tempHeaders
						headerRowNode = tr
						return true
					}
				}
			}
		}
		if n.Type == html.ElementNode && (n.Data == "table" || n.Data == "tbody") {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && c.Data == "tr" {
					isHeader := false
					tempHeaders := make([]string, 0, defaultCellCapacity)
					allCellsAreTh := true
					hasCells := false
					for cell := c.FirstChild; cell != nil; cell = cell.NextSibling {
						if cell.Type == html.ElementNode && (cell.Data == "th" || cell.Data == "td") {
							hasCells = true
							if cell.Data == "th" {
								isHeader = true
							} else {
								allCellsAreTh = false
							}
							tempHeaders = append(tempHeaders, getNodeTextOptimized(cell))
						}
					}
					if isHeader && allCellsAreTh && hasCells && len(tempHeaders) > 0 {
						headers = tempHeaders
						headerRowNode = c
						return true
					}
					if isHeader && hasCells && len(tempHeaders) > 0 && len(headers) == 0 {
						headers = tempHeaders
						headerRowNode = c
					}
				}
			}
		}
		if n.Type == html.ElementNode && (n.Data == "tbody" || n.Data == "table") {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if findHeaderRow(c) {
					return true
				}
			}
		}
		return len(headers) > 0
	}
	if !findHeaderRow(tableNode) || len(headers) == 0 {
		firstRow := true
		var firstTrNode *html.Node
		var findFirstTr func(*html.Node)
		findFirstTr = func(n *html.Node) {
			if n.Type == html.ElementNode && n.Data == "tr" {
				firstTrNode = n
				return
			}
			if n.Type == html.ElementNode && (n.Data == "tbody" || n.Data == "table") {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if firstTrNode != nil {
						return
					}
					findFirstTr(c)
				}
			}
		}
		findFirstTr(tableNode)
		if firstTrNode != nil && len(headers) == 0 {
			headerRowNode = firstTrNode
			tempHeaders := make([]string, 0, defaultCellCapacity)
			for td := firstTrNode.FirstChild; td != nil; td = td.NextSibling {
				if td.Type == html.ElementNode && (td.Data == "td" || td.Data == "th") {
					tempHeaders = append(tempHeaders, getNodeTextOptimized(td))
				}
			}
			if len(tempHeaders) > 0 {
				headers = tempHeaders
				log.Printf("Warning: No <th> header row found. Using first <tr> as header: %v", headers)
			}
			firstRow = false
		}
		if firstRow && len(headers) == 0 {
			log.Println("Critical: Headers could not be parsed.")
			http.Error(w, "Could not parse table headers", http.StatusInternalServerError)
			return
		}
	}
	if len(headers) == 0 {
		log.Println("Critical: Headers are empty after all parsing attempts.")
		http.Error(w, "Could not parse table headers (empty)", http.StatusInternalServerError)
		return
	}
	log.Printf("Final Headers Used: %v", headers)
	players := make([]Player, 0, defaultPlayerCapacity)
	rowNodesToProcess := make([]*html.Node, 0, defaultPlayerCapacity)
	var collectDataRows func(*html.Node)
	collectDataRows = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			if n != headerRowNode {
				rowNodesToProcess = append(rowNodesToProcess, n)
			}
		}
		if n.Type == html.ElementNode && (n.Data == "tbody" || n.Data == "table") {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				collectDataRows(c)
			}
		}
	}
	collectDataRows(tableNode)
	numRowsToProcess := len(rowNodesToProcess)
	if numRowsToProcess == 0 {
		log.Println("No data rows found after header skipping.")
	}
	numWorkers := runtime.NumCPU()
	if numRowsToProcess < numWorkers {
		numWorkers = numRowsToProcess
	}
	if numWorkers == 0 && numRowsToProcess > 0 {
		numWorkers = 1
	}
	rowNodeChan := make(chan *html.Node, numRowsToProcess)
	resultsChan := make(chan PlayerParseResult, numRowsToProcess)
	var wg sync.WaitGroup
	headersSnapshot := make([]string, len(headers))
	copy(headersSnapshot, headers)
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for rowNode := range rowNodeChan {
				player, err := parseRowToPlayer(rowNode, headersSnapshot)
				if err == nil {
					enhancePlayerWithCalculations(&player)
				}
				resultsChan <- PlayerParseResult{Player: player, Err: err}
			}
		}(i)
	}
	for _, rowNode := range rowNodesToProcess {
		rowNodeChan <- rowNode
	}
	close(rowNodeChan)
	go func() { wg.Wait(); close(resultsChan) }()
	datasetCurrencySymbol := "$"
	foundDatasetSymbol := false
	for result := range resultsChan {
		if result.Err == nil {
			players = append(players, result.Player)
			if !foundDatasetSymbol {
				_, _, tvSymbol := parseMonetaryValueGo(result.Player.TransferValue)
				if tvSymbol != "" {
					datasetCurrencySymbol = tvSymbol
					foundDatasetSymbol = true
				} else {
					_, _, wSymbol := parseMonetaryValueGo(result.Player.Wage)
					if wSymbol != "" {
						datasetCurrencySymbol = wSymbol
						foundDatasetSymbol = true
					}
				}
			}
		} else {
			log.Printf("Skipping row due to error: %v", result.Err)
		}
	}
	if len(players) > 0 {
		calculatePlayerPerformancePercentiles(players)
	}
	parseDuration := time.Since(parseStartTime)
	datasetID := uuid.New().String()
	storeMutex.Lock()
	playerDataStore[datasetID] = struct {
		Players        []Player
		CurrencySymbol string
	}{Players: players, CurrencySymbol: datasetCurrencySymbol}
	storeMutex.Unlock()
	log.Printf("Stored %d players with DatasetID: %s. Detected Currency: %s", len(players), datasetID, datasetCurrencySymbol)
	if len(players) > 0 {
		log.Printf("DEBUG: Sample Player 1 after all processing: Name='%s', Overall=%d, ParsedPositions=%v, ShortPositions=%v, PositionGroups=%v", players[0].Name, players[0].Overall, players[0].ParsedPositions, players[0].ShortPositions, players[0].PositionGroups)
		if len(players[0].PerformancePercentiles) > 0 {
			log.Printf("DEBUG: Sample Player 1 Performance Percentile Keys: %v", getMapKeys(players[0].PerformancePercentiles))
		}
	}
	response := UploadResponse{DatasetID: datasetID, Message: "File uploaded and parsed successfully.", DetectedCurrencySymbol: datasetCurrencySymbol}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding JSON: "+err.Error(), http.StatusInternalServerError)
	}
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	rowsPerSecond := 0.0
	if parseDuration.Seconds() > 0 {
		rowsPerSecond = float64(len(players)) / parseDuration.Seconds()
	}
	totalDuration := time.Since(startTime)
	log.Printf("--- Perf Metrics --- File: %s, Size: %d KB, Total Time: %v, Parse Time: %v, Rows: %d, Parsed: %d, Rows/Sec: %.2f, MemAlloc: %.2f MiB, Workers: %d, Goroutines: %d ---",
		handler.Filename, fileSize/1024, totalDuration, parseDuration, numRowsToProcess, len(players), rowsPerSecond, bToMb(memStats.Alloc), numWorkers, runtime.NumGoroutine())
}

func getMapKeys(m map[string]map[string]float64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
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
		http.Error(w, "Error encoding JSON: "+err.Error(), http.StatusInternalServerError)
	}
}

func enhancePlayerWithCalculations(player *Player) {
	player.NumericAttributes = make(map[string]int, len(player.Attributes))
	for key, valStr := range player.Attributes {
		switch key {
		case "Acc", "Pac", "Str", "Sta", "Nat", "Bal", "Jum", "Fin", "OtB", "Cmp", "Tec",
			"Hea", "Lon", "Pen", "Pas", "Vis", "Cro", "Fre", "Cor", "L Th", "Dri", "Fir",
			"Agi", "Fla", "Tck", "Mar", "Pos", "Cnt", "Ant", "Wor", "Dec", "Tea", "Det",
			"Bra", "Ldr", "Agg", "Han", "Ref", "Cmd", "Aer", "1v1", "Kic", "TRO", "Com", "Thr", "Ecc", "Pun":
			valInt, err := strconv.Atoi(valStr)
			if err == nil {
				player.NumericAttributes[key] = valInt
			} else {
				player.NumericAttributes[key] = 0
			}
		default:
		}
	}
	player.ParsedPositions = parsePlayerPositionsGo(player.Position)
	player.PositionGroups = getPlayerPositionGroupsGo(player.ParsedPositions)

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
	player.GK = calculateFifaStatGo(player.NumericAttributes, "GK")

	maxOverall := 0
	calculatedRoleOveralls := []RoleOverallScore{}
	muRoleSpecificOverallWeights.RLock()
	currentRoleWeightsSource := roleSpecificOverallWeights
	muRoleSpecificOverallWeights.RUnlock()

	if len(currentRoleWeightsSource) == 0 {
		log.Printf("Warning: roleSpecificOverallWeights is empty for player '%s'. Overall will be 0.", player.Name)
	}

	if len(player.ParsedPositions) > 0 {
		uniqueBaseRoleKeysConsidered := make(map[string]struct{}) // To avoid double-counting generic roles
		foundAnyRoleMatch := false

		for _, parsedPos := range player.ParsedPositions { // parsedPos is a full name like "Centre Back"
			shortKey, ok := parsedPositionToBaseRoleKeyGo[parsedPos]
			if !ok || shortKey == nullString {
				if parsedPos == "Goalkeeper" { // Special handling for GK if not in map
					shortKey = "GK"
				} else {
					// log.Printf("Warning: No short key found for parsed position '%s' for player '%s'. Skipping role calculation for this position.", parsedPos, player.Name)
					continue
				}
			}

			// Iterate through all role definitions in the weights map
			for roleKeyInJson, specificWeights := range currentRoleWeightsSource {
				// roleKeyInJson is like "DC - Ball Playing Defender" or "ST - Trequartista - Attack"
				// We need to check if it starts with the player's shortKey + " - "
				if strings.HasPrefix(roleKeyInJson, shortKey+" - ") {
					// Check if it's NOT a generic role for this specific loop, generic roles are handled separately
					if !strings.HasSuffix(roleKeyInJson, " - Generic") { // Avoid matching "DC - Generic" here if shortKey is "DC"
						foundAnyRoleMatch = true
						overallForThisRole := calculateOverallForRoleGo(player.NumericAttributes, specificWeights)
						calculatedRoleOveralls = append(calculatedRoleOveralls, RoleOverallScore{RoleName: roleKeyInJson, Score: overallForThisRole})
						if overallForThisRole > maxOverall {
							maxOverall = overallForThisRole
						}
					}
				}
			}

			// Handle generic role for this shortKey
			genericRoleKey := shortKey + " - Generic" // e.g., "DC - Generic"
			if specificWeights, exists := currentRoleWeightsSource[genericRoleKey]; exists {
				if _, considered := uniqueBaseRoleKeysConsidered[genericRoleKey]; !considered {
					foundAnyRoleMatch = true
					overallForThisRole := calculateOverallForRoleGo(player.NumericAttributes, specificWeights)

					// Check if this generic role score was already added (e.g. from a different parsedPos mapping to the same shortKey)
					alreadyAdded := false
					for _, cro := range calculatedRoleOveralls {
						if cro.RoleName == genericRoleKey {
							alreadyAdded = true
							break
						}
					}
					if !alreadyAdded {
						calculatedRoleOveralls = append(calculatedRoleOveralls, RoleOverallScore{RoleName: genericRoleKey, Score: overallForThisRole})
						if overallForThisRole > maxOverall {
							maxOverall = overallForThisRole
						}
					}
					uniqueBaseRoleKeysConsidered[genericRoleKey] = struct{}{}
				}
			}
		}
		if !foundAnyRoleMatch && len(player.ParsedPositions) > 0 {
			log.Printf("Warning: Player '%s' with ParsedPositions %v (ShortPositions: %v) found no matching roles in roleSpecificOverallWeights. MaxOverall will be 0.", player.Name, player.ParsedPositions, player.ShortPositions)
		}

	} else {
		log.Printf("Warning: Player '%s' has no ParsedPositions. MaxOverall will be 0.", player.Name)
	}

	player.Overall = maxOverall
	player.RoleSpecificOveralls = calculatedRoleOveralls
	sort.Slice(player.RoleSpecificOveralls, func(i, j int) bool {
		return player.RoleSpecificOveralls[i].Score > player.RoleSpecificOveralls[j].Score
	})
}

func parseRowToPlayer(tr *html.Node, headers []string) (Player, error) {
	var cells []string
	cellCap := defaultCellCapacity
	if len(headers) > 0 {
		cellCap = len(headers)
	}
	cells = make([]string, 0, cellCap)
	for td := tr.FirstChild; td != nil; td = td.NextSibling {
		if td.Type == html.ElementNode && (td.Data == "td" || td.Data == "th") {
			cells = append(cells, getNodeTextOptimized(td))
		}
	}
	if len(headers) == 0 {
		return Player{}, errors.New("cannot process row: headers are empty")
	}
	if len(cells) == 0 {
		return Player{}, errors.New("skipped row: appears to be an empty row (no cells)")
	}
	if len(cells) < len(headers)/2 {
		isMeaningful := false
		for _, cellContent := range cells {
			if strings.TrimSpace(cellContent) != "" {
				isMeaningful = true
				break
			}
		}
		if !isMeaningful {
			return Player{}, errors.New("skipped row: too few cells and all are empty")
		}
	}
	player := Player{
		Attributes:             make(map[string]string, defaultAttributeCapacity),
		PerformancePercentiles: make(map[string]map[string]float64),
	}
	knownNonAttributeHeaders := map[string]bool{"Inf": true}
	foundName := false
	for i, headerName := range headers {
		if i < len(cells) {
			cellValue := strings.TrimSpace(cells[i])
			isAnAttributeField := true

			switch headerName {
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
				if err == nil && valInt >= 0 && valInt <= 20 {
					// Natural Fitness attribute
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
				if _, isKnownNonAttr := knownNonAttributeHeaders[headerName]; !isKnownNonAttr {
					if headerName != "" && cellValue != "" && cellValue != "-" {
						player.Attributes[headerName] = cellValue
					}
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
		return Player{}, errors.New("skipped row: 'Name' field missing and row appears empty")
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
	port := "8091"
	log.Printf("Server starting on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
