package main

import "sync"

// --- START: Struct Definitions ---

// RoleOverallScore stores the calculated overall score for a player in a specific role.
type RoleOverallScore struct {
	RoleName string `json:"roleName"`
	Score    int    `json:"score"`
}

// Player holds all the information and calculated statistics for a football player.
type Player struct {
	UID                     int64                         `json:"uid"`
	Name                    string                        `json:"name"`
	Position                string                        `json:"position"`
	Age                     string                        `json:"age"`
	Club                    string                        `json:"club"`
	Division                string                        `json:"division"`
	BasedIn                 string                        `json:"basedIn,omitempty"`
	TransferValue           string                        `json:"transfer_value"`
	Wage                    string                        `json:"wage"`
	Personality             string                        `json:"personality,omitempty"`
	MediaHandling           string                        `json:"media_handling,omitempty"`
	Nationality             string                        `json:"nationality"`
	NationalityISO          string                        `json:"nationalityIso"`
	NationalityFIFACode     string                        `json:"nationality_fifa_code"`
	AttributeMasked         bool                          `json:"attributeMasked,omitempty"`
	Attributes              map[string]string             `json:"attributes"`
	NumericAttributes       map[string]int                `json:"numericAttributes"`
	PerformanceStatsNumeric map[string]float64            `json:"performanceStatsNumeric"`
	PerformancePercentiles  map[string]map[string]float64 `json:"performancePercentiles"`
	ParsedPositions         []string                      `json:"parsedPositions"`
	ShortPositions          []string                      `json:"shortPositions"`
	PositionGroups          []string                      `json:"positionGroups"`
	PAC                     int                           `json:"pac"`
	SHO                     int                           `json:"sho"`
	PAS                     int                           `json:"pas"`
	DRI                     int                           `json:"dri"`
	DEF                     int                           `json:"def"`
	PHY                     int                           `json:"phy"`
	GK                      int                           `json:"gk,omitempty"`
	DIV                     int                           `json:"div,omitempty"`
	HAN                     int                           `json:"han,omitempty"`
	REF                     int                           `json:"ref,omitempty"`
	KIC                     int                           `json:"kic,omitempty"`
	SPD                     int                           `json:"spd,omitempty"`
	POS                     int                           `json:"pos,omitempty"`
	Overall              int                `json:"overall"`
	TotalStats           int                `json:"totalStats"`
	MBR                  int                `json:"mbr"`
	CA                   int                `json:"ca"` // fm21-cas Current Ability estimate (0–200 scale)
	BestRoleOverall      string             `json:"bestRoleOverall"`
	RoleSpecificOveralls []RoleOverallScore `json:"roleSpecificOveralls"`
	TransferValueAmount  int64              `json:"transferValueAmount"`
	WageAmount           int64              `json:"wageAmount"`

	// Mutex to protect concurrent access to maps
	mu sync.RWMutex `json:"-"`
}

// PlayerParseResult is used by worker goroutines to return a parsed player or an error.
type PlayerParseResult struct {
	Player Player
	Err    error
}

// UploadResponse is the JSON response sent after a successful file upload and parse.
type UploadResponse struct {
	DatasetID              string   `json:"datasetId"`
	Message                string   `json:"message"`
	DetectedCurrencySymbol string   `json:"detectedCurrencySymbol,omitempty"`
	Players                []Player `json:"players,omitempty"` // Return processed players
	Roles                  []string `json:"roles,omitempty"`   // Return available roles
	PercentilesReady       bool     `json:"percentilesReady"`  // Indicate if percentiles are ready
	PlayerCount            int      `json:"playerCount"`       // Total player count
}

// PlayerDataWithCurrency is the JSON response for fetching player data, including the currency.
type PlayerDataWithCurrency struct {
	Players        []Player `json:"players"`
	CurrencySymbol string   `json:"currencySymbol"`
}

// --- END: Struct Definitions ---
