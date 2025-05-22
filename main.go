package main

import (
	"encoding/json"
	"errors"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid" // For generating unique IDs
	"golang.org/x/net/html"
)

const (
	defaultPlayerCapacity    = 1024
	defaultAttributeCapacity = 64
	defaultCellCapacity      = 64
	overallScalingFactor     = 5.85 // As used in Vue
)

// --- START: Struct Definitions ---

// RoleOverallScore stores a calculated overall score for a specific role.
type RoleOverallScore struct {
	RoleName string `json:"roleName"`
	Score    int    `json:"score"`
}

// Player struct now includes calculated stats and overall.
type Player struct {
	Name                 string             `json:"name"`
	Position             string             `json:"position"` // Original position string
	Age                  string             `json:"age"`
	Club                 string             `json:"club"`
	TransferValue        string             `json:"transfer_value"`
	Wage                 string             `json:"wage"`
	Personality          string             `json:"personality,omitempty"`    // New field
	MediaHandling        string             `json:"media_handling,omitempty"` // New field
	Nationality          string             `json:"nationality"`              // Full country name
	NationalityISO       string             `json:"nationality_iso"`          // 2-letter ISO code for flags
	NationalityFIFACode  string             `json:"nationality_fifa_code"`    // Original 3-letter FIFA code
	Attributes           map[string]string  `json:"attributes"`               // Raw string attributes
	NumericAttributes    map[string]int     `json:"-"`                        // Parsed numeric attributes (internal use)
	ParsedPositions      []string           `json:"parsedPositions"`          // Standardized positions
	PositionGroups       []string           `json:"positionGroups"`           // General groups like "Defenders", "Midfielders"
	PHY                  int                `json:"PHY"`                      // Calculated Physical stat
	SHO                  int                `json:"SHO"`                      // Calculated Shooting stat
	PAS                  int                `json:"PAS"`                      // Calculated Passing stat
	DRI                  int                `json:"DRI"`                      // Calculated Dribbling stat
	DEF                  int                `json:"DEF"`                      // Calculated Defending stat
	MEN                  int                `json:"MEN"`                      // Calculated Mental stat
	GK                   int                `json:"GK,omitempty"`             // Calculated Goalkeeping stat (NEW)
	Overall              int                `json:"Overall"`                  // Best overall rating
	RoleSpecificOveralls []RoleOverallScore `json:"roleSpecificOveralls"`     // All calculated role overalls
	TransferValueAmount  int64              `json:"transferValueAmount"`      // Numeric transfer value for sorting
	WageAmount           int64              `json:"wageAmount"`               // Numeric wage for sorting
}

// PlayerParseResult is used for concurrent processing.
type PlayerParseResult struct {
	Player Player
	Err    error
}

// UploadResponse defines the structure of the JSON response after a successful upload.
type UploadResponse struct {
	DatasetID string   `json:"datasetId"`
	Message   string   `json:"message"`
	Players   []Player `json:"players,omitempty"` // Optionally return players directly too, or just ID
}

// --- END: Struct Definitions ---

// --- START: In-Memory Store for Player Data ---
var (
	playerDataStore = make(map[string][]Player)
	storeMutex      sync.RWMutex
)

// --- END: In-Memory Store for Player Data ---

// --- START: Weight Data and Loading ---
var (
	attributeWeights             map[string]map[string]int
	roleSpecificOverallWeights   map[string]map[string]int
	muAttributeWeights           sync.RWMutex
	muRoleSpecificOverallWeights sync.RWMutex
)

// Default weights (mirrors frontend defaults)
var defaultAttributeWeightsGo = map[string]map[string]int{
	"PHY": {"Acc": 7, "Pac": 6, "Str": 5, "Sta": 4, "Nat": 3, "Bal": 2, "Jum": 1},
	"SHO": {"Fin": 7, "OtB": 6, "Cmp": 5, "Tec": 4, "Hea": 3, "Lon": 2, "Pen": 1},
	"PAS": {"Pas": 7, "Vis": 6, "Tec": 5, "Cro": 4, "Fre": 3, "Cor": 2, "L Th": 1},
	"DRI": {"Dri": 6, "Fir": 5, "Tec": 4, "Agi": 3, "Bal": 2, "Fla": 1},
	"DEF": {"Tck": 6, "Mar": 5, "Hea": 4, "Pos": 3, "Cnt": 2, "Ant": 1},
	"MEN": {"Wor": 11, "Dec": 10, "Tea": 9, "Det": 8, "Bra": 7, "Ldr": 6, "Vis": 5, "Agg": 4, "OtB": 3, "Pos": 2, "Ant": 1},
	"GK":  {"Han": 20, "Ref": 20, "Cmd": 15, "Aer": 15, "1v1": 10, "Kic": 5, "TRO": 5, "Com": 3, "Thr": 3, "Ecc": 1},
}

// Default role specific weights.
var defaultRoleSpecificOverallWeightsGo = map[string]map[string]int{
	"DC - BPD":     {"Cor": 5, "Cro": 1, "Dri": 40, "Fin": 10, "Fir": 35, "Fre": 10, "Hea": 55, "Lon": 10, "Tea": 20, "L Th": 0, "Mar": 55, "Pas": 55, "Pen": 10, "Tck": 40, "Tec": 35, "Agg": 40, "Ant": 50, "Bra": 30, "Cmp": 80, "Cnt": 50, "Dec": 50, "Det": 20, "Fla": 10, "Ldr": 10, "OtB": 10, "Pos": 55, "Vis": 50, "Wor": 55, "Acc": 90, "Agi": 60, "Bal": 35, "Jum": 65, "Nat": 10, "Pac": 90, "Sta": 30, "Str": 50},
	"ST - AF - At": {"Cor": 5, "Cro": 5, "Dri": 75, "Fin": 80, "Fir": 50, "Fre": 5, "Hea": 25, "Lon": 25, "L Th": 1, "Mar": 1, "Pas": 40, "Pen": 20, "Tck": 5, "Tec": 65, "Agg": 50, "Ant": 50, "Bra": 20, "Cmp": 35, "Cnt": 5, "Dec": 45, "Det": 20, "Fla": 25, "Ldr": 10, "OtB": 45, "Pos": 5, "Tea": 10, "Vis": 20, "Wor": 60, "Acc": 100, "Agi": 30, "Bal": 50, "Jum": 20, "Nat": 10, "Pac": 70, "Sta": 65, "Str": 25},
	"GK - GK - De": {
		"Aer": 80, "Cmd": 75, "Com": 50, "Ecc": 10, "Han": 90, "Kic": 40, "1v1": 80, "Ref": 90, "TRO": 30, "Thr": 40,
		"Ant": 60, "Cmp": 60, "Cnt": 70, "Dec": 70, "Pos": 75,
		"Agi": 50, "Jum": 60, "Str": 50, "Acc": 30, "Pac": 30,
		"Det": 50, "Ldr": 40, "Bra": 60, "Wor": 40, "Tea": 40,
	},
	"GK - SK - De": {
		"Aer": 75, "Cmd": 70, "Com": 55, "Ecc": 20, "Han": 85, "Kic": 60, "1v1": 75, "Ref": 85, "TRO": 60, "Thr": 50,
		"Ant": 65, "Cmp": 65, "Cnt": 65, "Dec": 75, "Pos": 70,
		"Acc": 50, "Agi": 55, "Jum": 55, "Pac": 50, "Str": 45,
		"Fir": 40, "Pas": 40, "Tec": 30, "Vis": 30,
		"Det": 50, "Ldr": 40, "Bra": 60, "Wor": 40, "Tea": 40,
	},
	"GK - SK - Su": {
		"Aer": 70, "Cmd": 65, "Com": 60, "Ecc": 30, "Han": 80, "Kic": 75, "1v1": 70, "Ref": 80, "TRO": 75, "Thr": 65,
		"Ant": 70, "Cmp": 70, "Cnt": 60, "Dec": 80, "Pos": 65, "Vis": 50,
		"Acc": 60, "Agi": 60, "Jum": 50, "Pac": 60, "Str": 40,
		"Fir": 60, "Pas": 60, "Tec": 50,
		"Det": 50, "Ldr": 40, "Bra": 50, "Wor": 50, "Tea": 50, "Fla": 20, "OtB": 30,
	},
	"GK - SK - At": {
		"Aer": 65, "Cmd": 60, "Com": 65, "Ecc": 40, "Han": 75, "Kic": 85, "1v1": 65, "Ref": 75, "TRO": 85, "Thr": 75,
		"Ant": 75, "Cmp": 75, "Cnt": 55, "Dec": 85, "Pos": 60, "Vis": 65, "Fla": 40, "OtB": 40,
		"Acc": 70, "Agi": 65, "Jum": 45, "Pac": 70, "Str": 35,
		"Fir": 70, "Pas": 70, "Tec": 60,
		"Det": 50, "Ldr": 40, "Bra": 40, "Wor": 50, "Tea": 50,
	},
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
	log.Printf("Successfully loaded weights from %s", filePath)
	return weights, nil
}

func init() {
	var err error
	attributeWeights, err = loadJSONWeights(filepath.Join("public", "attribute_weights.json"), defaultAttributeWeightsGo)
	if err != nil {
		log.Printf("Using default attribute_weights due to error: %v", err)
	}

	roleSpecificOverallWeights, err = loadJSONWeights(filepath.Join("public", "role_specific_overall_weights.json"), defaultRoleSpecificOverallWeightsGo)
	if err != nil {
		log.Printf("Using default role_specific_overall_weights due to error: %v", err)
	}
}

// --- END: Weight Data and Loading ---

// --- START: Position Parsing Logic (Mirrors Vue logic) ---
var (
	positionRoleMapGo = map[string]string{
		"GK": "Goalkeeper", "SW": "Sweeper", "D": "Defender", "WB": "Wing-Back",
		"DM": "Defensive Midfielder", "M": "Midfielder", "AM": "Attacking Midfielder",
		"ST": "Striker", "F": "Forward",
	}
	positionSideMapGo = map[string]string{
		"R": "Right", "L": "Left", "C": "Centre",
	}
	standardizedPositionNameMapGo = map[string]string{
		"Goalkeeper (Centre)": "Goalkeeper", "Goalkeeper": "Goalkeeper",
		"Sweeper (Centre)": "Sweeper", "Sweeper": "Sweeper",
		"Defender (Right)": "Right Back", "Defender (Left)": "Left Back", "Defender (Centre)": "Centre Back",
		"Wing-Back (Right)": "Right Wing-Back", "Wing-Back (Left)": "Left Wing-Back", "Wing-Back (Centre)": "Centre Wing-Back",
		"Defensive Midfielder (Right)": "Right Defensive Midfielder", "Defensive Midfielder (Left)": "Left Defensive Midfielder", "Defensive Midfielder (Centre)": "Centre Defensive Midfielder",
		"Midfielder (Right)": "Right Midfielder", "Midfielder (Left)": "Left Midfielder", "Midfielder (Centre)": "Centre Midfielder",
		"Attacking Midfielder (Right)": "Right Attacking Midfielder", "Attacking Midfielder (Left)": "Left Attacking Midfielder", "Attacking Midfielder (Centre)": "Centre Attacking Midfielder",
		"Striker (Centre)": "Striker", "Striker": "Striker",
		"Forward (Right)": "Right Forward", "Forward (Left)": "Left Forward", "Forward (Centre)": "Centre Forward",
	}
	positionGroupsGo = map[string][]string{
		"Goalkeepers": {"Goalkeeper"},
		"Defenders":   {"Sweeper", "Right Back", "Left Back", "Centre Back", "Right Wing-Back", "Left Wing-Back", "Centre Wing-Back"},
		"Midfielders": {"Right Defensive Midfielder", "Left Defensive Midfielder", "Centre Defensive Midfielder", "Right Midfielder", "Left Midfielder", "Centre Midfielder", "Right Attacking Midfielder", "Left Attacking Midfielder", "Centre Attacking Midfielder"},
		"Attackers":   {"Striker", "Right Forward", "Left Forward", "Centre Forward"},
	}
	parsedPositionToBaseRoleKeyGo = map[string]string{
		"Goalkeeper":                  "GK",
		"Sweeper":                     "DC",
		"Right Back":                  "DR/L",
		"Left Back":                   "DR/L",
		"Centre Back":                 "DC",
		"Right Wing-Back":             "WBR/L",
		"Left Wing-Back":              "WBR/L",
		"Centre Wing-Back":            "WBR/L",
		"Right Defensive Midfielder":  "DM",
		"Left Defensive Midfielder":   "DM",
		"Centre Defensive Midfielder": "DM",
		"Right Midfielder":            "MR/L",
		"Left Midfielder":             "MR/L",
		"Centre Midfielder":           "MC",
		"Right Attacking Midfielder":  "AMR/L",
		"Left Attacking Midfielder":   "AMR/L",
		"Centre Attacking Midfielder": "AMC",
		"Striker":                     "ST",
		"Right Forward":               "AMR/L",
		"Left Forward":                "AMR/L",
		"Centre Forward":              "ST",
	}
	nullString = ""
)

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

			roleFullName, roleExists := positionRoleMapGo[roleKey]
			if roleExists {
				sidesToIterate := explicitSidesArray
				if len(sidesToIterate) == 0 {
					if roleKey == "GK" {
						sidesToIterate = []string{"C"}
					} else {
						sidesToIterate = []string{"C"}
					}
				}

				for _, sideKey := range sidesToIterate {
					sideFullName, sideExists := positionSideMapGo[sideKey]
					if sideExists {
						detailedName := roleFullName + " (" + sideFullName + ")"
						standardizedName, ok := standardizedPositionNameMapGo[detailedName]
						if ok {
							finalPositionsSet[standardizedName] = struct{}{}
						} else {
							if roleKey == "GK" {
								standardizedNameGK, okGK := standardizedPositionNameMapGo[roleFullName]
								if okGK {
									finalPositionsSet[standardizedNameGK] = struct{}{}
									continue
								}
							}
							finalPositionsSet[detailedName] = struct{}{}
						}
					}
				}
			} else {
				standardizedName, ok := standardizedPositionNameMapGo[roleKey]
				if ok {
					finalPositionsSet[standardizedName] = struct{}{}
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

// --- END: Position Parsing Logic ---

// --- START: Calculation Logic ---

func calculateFifaStatGo(playerNumericAttributes map[string]int, categoryName string) int {
	muAttributeWeights.RLock()
	currentCategoryWeightsSource := attributeWeights
	muAttributeWeights.RUnlock()

	categoryAttributeWeights, ok := currentCategoryWeightsSource[categoryName]
	if !ok {
		log.Printf("Warning: Attribute weights for category '%s' not found. Using defaults for this category.", categoryName)
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

// --- END: Calculation Logic ---

// --- START: FIFA Country Code Maps ---
var fifaCountryCodes = map[string]string{
	"AFG": "Afghanistan", "ALB": "Albania", "ALG": "Algeria", "ASA": "American Samoa", "AND": "Andorra",
	"ANG": "Angola", "AIA": "Anguilla", "ATG": "Antigua and Barbuda", "ARG": "Argentina", "ARM": "Armenia",
	"ARU": "Aruba", "AUS": "Australia", "AUT": "Austria", "AZE": "Azerbaijan", "BAH": "Bahamas",
	"BHR": "Bahrain", "BAN": "Bangladesh", "BRB": "Barbados", "BLR": "Belarus", "BEL": "Belgium",
	"BLZ": "Belize", "BEN": "Benin", "BER": "Bermuda", "BHU": "Bhutan", "BOL": "Bolivia",
	"BIH": "Bosnia and Herzegovina", "BOT": "Botswana", "BRA": "Brazil", "VGB": "British Virgin Islands",
	"BRU": "Brunei Darussalam", "BUL": "Bulgaria", "BFA": "Burkina Faso", "BDI": "Burundi", "CAM": "Cambodia",
	"CMR": "Cameroon", "CAN": "Canada", "CPV": "Cape Verde", "CAY": "Cayman Islands", "CTA": "Central African Republic",
	"CHA": "Chad", "CHI": "Chile", "CHN": "China PR", "TPE": "Chinese Taipei", "COL": "Colombia",
	"COM": "Comoros", "CGO": "Congo", "COD": "DR Congo", "COK": "Cook Islands", "CRC": "Costa Rica",
	"CIV": "Ivory Coast", "CRO": "Croatia", "CUB": "Cuba", "CUW": "Curaçao", "CYP": "Cyprus",
	"CZE": "Czech Republic", "DEN": "Denmark", "DJI": "Djibouti", "DMA": "Dominica", "DOM": "Dominican Republic",
	"ECU": "Ecuador", "EGY": "Egypt", "SLV": "El Salvador", "ENG": "England", "EQG": "Equatorial Guinea",
	"ERI": "Eritrea", "EST": "Estonia", "SWZ": "Eswatini", "ETH": "Ethiopia", "FRO": "Faroe Islands",
	"FIJ": "Fiji", "FIN": "Finland", "FRA": "France", "GAB": "Gabon", "GAM": "Gambia",
	"GEO": "Georgia", "GER": "Germany", "GHA": "Ghana", "GIB": "Gibraltar", "GRE": "Greece",
	"GRN": "Grenada", "GUM": "Guam", "GUA": "Guatemala", "GUI": "Guinea", "GNB": "Guinea-Bissau",
	"GUY": "Guyana", "HAI": "Haiti", "HON": "Honduras", "HKG": "Hong Kong", "HUN": "Hungary",
	"ISL": "Iceland", "IND": "India", "IDN": "Indonesia", "IRN": "Iran", "IRQ": "Iraq",
	"IRL": "Republic of Ireland", "ISR": "Israel", "ITA": "Italy", "JAM": "Jamaica", "JPN": "Japan",
	"JOR": "Jordan", "KAZ": "Kazakhstan", "KEN": "Kenya", "PRK": "North Korea", "KOR": "South Korea",
	"KVX": "Kosovo", "KUW": "Kuwait", "KGZ": "Kyrgyzstan", "LAO": "Laos", "LVA": "Latvia",
	"LBN": "Lebanon", "LES": "Lesotho", "LBR": "Liberia", "LBY": "Libya", "LIE": "Liechtenstein",
	"LTU": "Lithuania", "LUX": "Luxembourg", "MAC": "Macau", "MAD": "Madagascar", "MWI": "Malawi",
	"MAS": "Malaysia", "MDV": "Maldives", "MLI": "Mali", "MLT": "Malta", "MTN": "Mauritania",
	"MRI": "Mauritius", "MEX": "Mexico", "MDA": "Moldova", "MNG": "Mongolia", "MNE": "Montenegro",
	"MSR": "Montserrat", "MAR": "Morocco", "MOZ": "Mozambique", "MYA": "Myanmar", "NAM": "Namibia",
	"NEP": "Nepal", "NED": "Netherlands", "NCL": "New Caledonia", "NZL": "New Zealand", "NCA": "Nicaragua",
	"NIG": "Niger", "NGA": "Nigeria", "MKD": "North Macedonia", "NIR": "Northern Ireland", "NOR": "Norway",
	"OMA": "Oman", "PAK": "Pakistan", "PLE": "Palestine", "PAN": "Panama", "PNG": "Papua New Guinea",
	"PAR": "Paraguay", "PER": "Peru", "PHI": "Philippines", "POL": "Poland", "POR": "Portugal",
	"PUR": "Puerto Rico", "QAT": "Qatar", "ROU": "Romania", "RUS": "Russia", "RWA": "Rwanda",
	"SKN": "St. Kitts and Nevis", "LCA": "St. Lucia", "VIN": "St. Vincent & Grenadines", "SAM": "Samoa",
	"SMR": "San Marino", "STP": "São Tomé e Príncipe", "KSA": "Saudi Arabia", "SCO": "Scotland", "SEN": "Senegal",
	"SRB": "Serbia", "SEY": "Seychelles", "SLE": "Sierra Leone", "SIN": "Singapore", "SVK": "Slovakia",
	"SVN": "Slovenia", "SOL": "Solomon Islands", "SOM": "Somalia", "RSA": "South Africa", "SSD": "South Sudan",
	"ESP": "Spain", "SRI": "Sri Lanka", "SDN": "Sudan", "SUR": "Suriname", "SWE": "Sweden",
	"SUI": "Switzerland", "SYR": "Syria", "TAH": "Tahiti", "TJK": "Tajikistan", "TAN": "Tanzania",
	"THA": "Thailand", "TLS": "Timor-Leste", "TOG": "Togo", "TGA": "Tonga", "TRI": "Trinidad and Tobago",
	"TUN": "Tunisia", "TUR": "Turkey", "TKM": "Turkmenistan", "TCA": "Turks and Caicos Islands",
	"UGA": "Uganda", "UKR": "Ukraine", "UAE": "United Arab Emirates", "USA": "USA", "URU": "Uruguay",
	"VIR": "US Virgin Islands", "UZB": "Uzbekistan", "VAN": "Vanuatu", "VEN": "Venezuela", "VIE": "Vietnam",
	"WAL": "Wales", "YEM": "Yemen", "ZAM": "Zambia", "ZIM": "Zimbabwe",
}

var fifaToISO2 = map[string]string{
	"AFG": "AF", "ALB": "AL", "ALG": "DZ", "ASA": "AS", "AND": "AD", "ANG": "AO", "AIA": "AI",
	"ATG": "AG", "ARG": "AR", "ARM": "AM", "ARU": "AW", "AUS": "AU", "AUT": "AT", "AZE": "AZ",
	"BAH": "BS", "BHR": "BH", "BAN": "BD", "BRB": "BB", "BLR": "BY", "BEL": "BE", "BLZ": "BZ",
	"BEN": "BJ", "BER": "BM", "BHU": "BT", "BOL": "BO", "BIH": "BA", "BOT": "BW", "BRA": "BR",
	"VGB": "VG", "BRU": "BN", "BUL": "BG", "BFA": "BF", "BDI": "BI", "CAM": "KH", "CMR": "CM",
	"CAN": "CA", "CPV": "CV", "CAY": "KY", "CTA": "CF", "CHA": "TD", "CHI": "CL", "CHN": "CN",
	"TPE": "TW", "COL": "CO", "COM": "KM", "CGO": "CG", "COD": "CD", "COK": "CK", "CRC": "CR",
	"CIV": "CI", "CRO": "HR", "CUB": "CU", "CUW": "CW", "CYP": "CY", "CZE": "CZ", "DEN": "DK",
	"DJI": "DJ", "DMA": "DM", "DOM": "DO", "ECU": "EC", "EGY": "EG", "SLV": "SV",
	"ENG": "gb-eng", "EQG": "GQ", "ERI": "ER", "EST": "EE", "SWZ": "SZ", "ETH": "ET", "FRO": "FO",
	"FIJ": "FJ", "FIN": "FI", "FRA": "FR", "GAB": "GA", "GAM": "GM", "GEO": "GE", "GER": "DE",
	"GHA": "GH", "GIB": "GI", "GRE": "GR", "GRN": "GD", "GUM": "GU", "GUA": "GT", "GUI": "GN",
	"GNB": "GW", "GUY": "GY", "HAI": "HT", "HON": "HN", "HKG": "HK", "HUN": "HU", "ISL": "IS",
	"IND": "IN", "IDN": "ID", "IRN": "IR", "IRQ": "IQ", "IRL": "IE", "ISR": "IL", "ITA": "IT",
	"JAM": "JM", "JPN": "JP", "JOR": "JO", "KAZ": "KZ", "KEN": "KE", "PRK": "KP", "KOR": "KR",
	"KVX": "XK", "KUW": "KW", "KGZ": "KG", "LAO": "LA", "LVA": "LV", "LBN": "LB", "LES": "LS",
	"LBR": "LR", "LBY": "LY", "LIE": "LI", "LTU": "LT", "LUX": "LU", "MAC": "MO", "MAD": "MG",
	"MWI": "MW", "MAS": "MY", "MDV": "MV", "MLI": "ML", "MLT": "MT", "MTN": "MR", "MRI": "MU",
	"MEX": "MX", "MDA": "MD", "MNG": "MN", "MNE": "ME", "MSR": "MS", "MAR": "MA", "MOZ": "MZ",
	"MYA": "MM", "NAM": "NA", "NEP": "NP", "NED": "NL", "NCL": "NC", "NZL": "NZ", "NCA": "NI",
	"NIG": "NE", "NGA": "NG", "MKD": "MK", "NIR": "gb-nir", "NOR": "NO", "OMA": "OM", "PAK": "PK",
	"PLE": "PS", "PAN": "PA", "PNG": "PG", "PAR": "PY", "PER": "PE", "PHI": "PH", "POL": "PL",
	"POR": "PT", "PUR": "PR", "QAT": "QA", "ROU": "RO", "RUS": "RU", "RWA": "RW",
	"SKN": "KN", "LCA": "LC", "VIN": "VC", "SAM": "WS", "SMR": "SM", "STP": "ST", "KSA": "SA",
	"SCO": "gb-sct", "SEN": "SN", "SRB": "RS", "SEY": "SC", "SLE": "SL", "SIN": "SG", "SVK": "SK",
	"SVN": "SI", "SOL": "SB", "SOM": "SO", "RSA": "ZA", "SSD": "SS", "ESP": "ES", "SRI": "LK",
	"SDN": "SD", "SUR": "SR", "SWE": "SE", "SUI": "CH", "SYR": "SY", "TAH": "PF",
	"TJK": "TJ", "TAN": "TZ", "THA": "TH", "TLS": "TL", "TOG": "TG", "TGA": "TO", "TRI": "TT",
	"TUN": "TN", "TUR": "TR", "TKM": "TM", "TCA": "TC", "UGA": "UG", "UKR": "UA", "UAE": "AE",
	"USA": "US", "URU": "UY", "VIR": "VI", "UZB": "UZ", "VAN": "VU", "VEN": "VE", "VIE": "VN",
	"WAL": "gb-wls", "YEM": "YE", "ZAM": "ZM", "ZIM": "ZW",
}

// --- END: FIFA Country Code Maps ---

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
			sb.WriteByte(' ')
		} else if c.Type == html.TextNode && c.NextSibling != nil && c.NextSibling.Type == html.ElementNode {
			sb.WriteByte(' ')
		}
	}
	return strings.Join(strings.Fields(sb.String()), " ")
}

func parseMonetaryValueGo(rawValue string) (string, int64) {
	cleanedValue := strings.TrimSpace(rawValue)
	if strings.Contains(cleanedValue, " - ") {
		parts := strings.Split(cleanedValue, " - ")
		if len(parts) == 2 {
			cleanedValue = strings.TrimSpace(parts[1])
		}
	}

	var numericValue int64
	originalStr := cleanedValue

	cleanedValue = strings.ReplaceAll(cleanedValue, "€", "")
	cleanedValue = strings.ReplaceAll(cleanedValue, "£", "")
	cleanedValue = strings.ReplaceAll(cleanedValue, "$", "")
	cleanedValue = strings.TrimSpace(strings.ReplaceAll(cleanedValue, "p/w", ""))

	multiplier := int64(1)
	if strings.HasSuffix(cleanedValue, "M") {
		multiplier = 1000000
		cleanedValue = strings.TrimSuffix(cleanedValue, "M")
	} else if strings.HasSuffix(cleanedValue, "K") {
		multiplier = 1000
		cleanedValue = strings.TrimSuffix(cleanedValue, "K")
	}
	cleanedValue = strings.ReplaceAll(cleanedValue, ",", "")

	valFloat, err := strconv.ParseFloat(cleanedValue, 64)
	if err == nil {
		numericValue = int64(valFloat * float64(multiplier))
	}
	return originalStr, numericValue
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
		if n.Type == html.ElementNode && n.Data == "tr" {
			isHeader := false
			tempHeaders := make([]string, 0, defaultCellCapacity)
			for cell := n.FirstChild; cell != nil; cell = cell.NextSibling {
				if cell.Type == html.ElementNode && cell.Data == "th" {
					isHeader = true
					tempHeaders = append(tempHeaders, getNodeTextOptimized(cell))
				}
			}
			if isHeader && len(tempHeaders) > 0 {
				headers = tempHeaders
				headerRowNode = n
				log.Printf("Parsed Headers: %v", headers)
				return true
			}
		}
		if n.Type == html.ElementNode && (n.Data == "tbody" || n.Data == "table" || n.Data == "thead") {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if findHeaderRow(c) {
					return true
				}
			}
		}
		return false
	}

	if !findHeaderRow(tableNode) {
		firstRow := true
		for tr := tableNode.FirstChild; tr != nil; tr = tr.NextSibling {
			if tr.Type == html.ElementNode && tr.Data == "tbody" {
				for rowNode := tr.FirstChild; rowNode != nil; rowNode = rowNode.NextSibling {
					if rowNode.Type == html.ElementNode && rowNode.Data == "tr" {
						if firstRow {
							headerRowNode = rowNode
							for td := rowNode.FirstChild; td != nil; td = td.NextSibling {
								if td.Type == html.ElementNode && (td.Data == "td" || td.Data == "th") {
									headers = append(headers, getNodeTextOptimized(td))
								}
							}
							log.Printf("Warning: No <th> header row found. Using first row as header: %v", headers)
							firstRow = false
							break
						}
					}
				}
				break
			} else if tr.Type == html.ElementNode && tr.Data == "tr" {
				if firstRow {
					headerRowNode = tr
					for td := tr.FirstChild; td != nil; td = td.NextSibling {
						if td.Type == html.ElementNode && (td.Data == "td" || td.Data == "th") {
							headers = append(headers, getNodeTextOptimized(td))
						}
					}
					log.Printf("Warning: No <th> header row found. Using first row as header (direct tr): %v", headers)
					firstRow = false
					break
				}
			}
			if !firstRow {
				break
			}
		}
	}
	if len(headers) == 0 {
		log.Println("Critical: Headers could not be parsed.")
		http.Error(w, "Could not parse table headers", http.StatusInternalServerError)
		return
	}

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
		log.Println("No data rows found.")
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
		go func() {
			defer wg.Done()
			for rowNode := range rowNodeChan {
				player, err := parseRowToPlayer(rowNode, headersSnapshot)
				if err == nil {
					enhancePlayerWithCalculations(&player)
				}
				resultsChan <- PlayerParseResult{Player: player, Err: err}
			}
		}()
	}

	for _, rowNode := range rowNodesToProcess {
		rowNodeChan <- rowNode
	}
	close(rowNodeChan)

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	for result := range resultsChan {
		if result.Err == nil {
			players = append(players, result.Player)
		} else {
			log.Printf("Skipping row due to error: %v", result.Err)
		}
	}

	parseDuration := time.Since(parseStartTime)

	// Store players in memory and return dataset ID
	datasetID := uuid.New().String()
	storeMutex.Lock()
	playerDataStore[datasetID] = players
	storeMutex.Unlock()

	log.Printf("Stored %d players with DatasetID: %s", len(players), datasetID)

	response := UploadResponse{
		DatasetID: datasetID,
		Message:   "File uploaded and parsed successfully.",
		// Players: players, // Optionally include players in response if needed by client immediately
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // For development
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

// New handler to retrieve player data by dataset ID
func playerDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract datasetID from path, e.g., /api/players/{datasetID}
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/players/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		http.Error(w, "Dataset ID is missing", http.StatusBadRequest)
		return
	}
	datasetID := pathParts[0]

	storeMutex.RLock()
	players, found := playerDataStore[datasetID]
	storeMutex.RUnlock()

	if !found {
		http.Error(w, "Player data not found for the given ID", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // For development
	if err := json.NewEncoder(w).Encode(players); err != nil {
		http.Error(w, "Error encoding JSON: "+err.Error(), http.StatusInternalServerError)
	}
}

// enhancePlayerWithCalculations populates calculated fields.
func enhancePlayerWithCalculations(player *Player) {
	player.NumericAttributes = make(map[string]int, len(player.Attributes))
	for key, valStr := range player.Attributes {
		valInt, err := strconv.Atoi(valStr)
		if err == nil {
			player.NumericAttributes[key] = valInt
		} else {
			player.NumericAttributes[key] = 0
		}
	}

	player.ParsedPositions = parsePlayerPositionsGo(player.Position)
	player.PositionGroups = getPlayerPositionGroupsGo(player.ParsedPositions)

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

	if len(player.ParsedPositions) > 0 {
		uniqueBaseRoleKeysConsidered := make(map[string]struct{})

		for _, parsedPos := range player.ParsedPositions {
			baseRoleKey, ok := parsedPositionToBaseRoleKeyGo[parsedPos]
			if !ok || baseRoleKey == nullString {
				if parsedPos == "Goalkeeper" {
					baseRoleKey = "GK"
				} else {
					continue
				}
			}

			for roleKeyInJson, specificWeights := range currentRoleWeightsSource {
				if strings.HasPrefix(roleKeyInJson, baseRoleKey+" - ") {
					overallForThisRole := calculateOverallForRoleGo(player.NumericAttributes, specificWeights)
					calculatedRoleOveralls = append(calculatedRoleOveralls, RoleOverallScore{RoleName: roleKeyInJson, Score: overallForThisRole})
					if overallForThisRole > maxOverall {
						maxOverall = overallForThisRole
					}
				}
			}
			genericRoleKey := baseRoleKey + " - Generic"
			if specificWeights, exists := currentRoleWeightsSource[genericRoleKey]; exists {
				if _, considered := uniqueBaseRoleKeysConsidered[genericRoleKey]; !considered {
					overallForThisRole := calculateOverallForRoleGo(player.NumericAttributes, specificWeights)
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
		isEmptyRow := true
		for _, cellContent := range cells {
			if strings.TrimSpace(cellContent) != "" {
				isEmptyRow = false
				break
			}
		}
		if isEmptyRow && len(cells) < len(headers)/2 {
			return Player{}, errors.New("skipped row: appears to be an empty or malformed row")
		}
	}

	player := Player{
		Attributes: make(map[string]string, defaultAttributeCapacity),
	}

	knownNonAttributeHeaders := map[string]bool{
		"Inf": true,
	}

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
				player.TransferValue, player.TransferValueAmount = parseMonetaryValueGo(cellValue)
				isAnAttributeField = false
			case "Wage":
				player.Wage, player.WageAmount = parseMonetaryValueGo(cellValue)
				isAnAttributeField = false
			case "Personality":
				player.Personality = cellValue
				isAnAttributeField = false
			case "Media Handling":
				player.MediaHandling = cellValue
				isAnAttributeField = false
			case "Nat":
				fifaCode := strings.ToUpper(cellValue)
				player.NationalityFIFACode = fifaCode

				if player.Nationality == "" {
					if fullName, ok := fifaCountryCodes[fifaCode]; ok {
						player.Nationality = fullName
					} else {
						player.Nationality = cellValue
					}

					if isoCode, ok := fifaToISO2[fifaCode]; ok {
						player.NationalityISO = isoCode
					} else {
						player.NationalityISO = strings.ToLower(cellValue)
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
			return Player{}, errors.New("skipped row: 'Name' field is missing or empty. First few cells: " + strings.Join(getFirstNCells(cells, 5), ", "))
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

func bToMb(b uint64) float64 {
	return float64(b) / 1024 / 1024
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(".", "index.html"))
	})

	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	// Handler for uploading files and storing data
	http.HandleFunc("/upload", uploadHandler)
	// Handler for retrieving stored player data
	http.HandleFunc("/api/players/", playerDataHandler) // Note the trailing slash

	port := "8091"
	log.Printf("Server starting on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
