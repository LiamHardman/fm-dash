package main

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
	TransferValue           string                        `json:"transfer_value"`
	Wage                    string                        `json:"wage"`
	Personality             string                        `json:"personality,omitempty"`
	MediaHandling           string                        `json:"media_handling,omitempty"`
	Nationality             string                        `json:"nationality"`
	NationalityISO          string                        `json:"nationality_iso"`
	NationalityFIFACode     string                        `json:"nationality_fifa_code"`
	AttributeMasked         bool                          `json:"attributeMasked,omitempty"`
	Attributes              map[string]string             `json:"attributes"`
	NumericAttributes       map[string]int                `json:"numericAttributes"`
	PerformanceStatsNumeric map[string]float64            `json:"performanceStatsNumeric"`
	PerformancePercentiles  map[string]map[string]float64 `json:"performancePercentiles"`
	ParsedPositions         []string                      `json:"parsedPositions"`
	ShortPositions          []string                      `json:"shortPositions"`
	PositionGroups          []string                      `json:"positionGroups"`
	PAC                     int                           `json:"PAC"`
	SHO                     int                           `json:"SHO"`
	PAS                     int                           `json:"PAS"`
	DRI                     int                           `json:"DRI"`
	DEF                     int                           `json:"DEF"`
	PHY                     int                           `json:"PHY"`
	GK                      int                           `json:"GK,omitempty"`
	DIV                     int                           `json:"DIV,omitempty"`
	HAN                     int                           `json:"HAN,omitempty"`
	REF                     int                           `json:"REF,omitempty"`
	KIC                     int                           `json:"KIC,omitempty"`
	SPD                     int                           `json:"SPD,omitempty"`
	POS                     int                           `json:"POS,omitempty"`
	Overall                 int                           `json:"Overall"`
	BestRoleOverall         string                        `json:"bestRoleOverall"`
	RoleSpecificOveralls    []RoleOverallScore            `json:"roleSpecificOveralls"`
	TransferValueAmount     int64                         `json:"transferValueAmount"`
	WageAmount              int64                         `json:"wageAmount"`
}

// PlayerParseResult is used by worker goroutines to return a parsed player or an error.
type PlayerParseResult struct {
	Player Player
	Err    error
}

// UploadResponse is the JSON response sent after a successful file upload and parse.
type UploadResponse struct {
	DatasetID              string `json:"datasetId"`
	Message                string `json:"message"`
	DetectedCurrencySymbol string `json:"detectedCurrencySymbol,omitempty"`
}

// PlayerDataWithCurrency is the JSON response for fetching player data, including the currency.
type PlayerDataWithCurrency struct {
	Players        []Player `json:"players"`
	CurrencySymbol string   `json:"currencySymbol"`
}

// --- END: Struct Definitions ---
