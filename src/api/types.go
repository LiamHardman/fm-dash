package main

// --- START: Struct Definitions ---

// RoleOverallScore stores the calculated overall score for a player in a specific role.
type RoleOverallScore struct {
	RoleName string `json:"roleName"`
	Score    int    `json:"score"`
}

// Player holds all the information and calculated statistics for a football player.
type Player struct {
	UID                     string                        `json:"uid"`                    // Unique identifier for the player
	Name                    string                        `json:"name"`
	Position                string                        `json:"position"` // Raw position string from HTML
	Age                     string                        `json:"age"`
	Club                    string                        `json:"club"`
	Division                string                        `json:"division"`
	TransferValue           string                        `json:"transfer_value"` // Original string representation
	Wage                    string                        `json:"wage"`           // Original string representation
	Personality             string                        `json:"personality,omitempty"`
	MediaHandling           string                        `json:"media_handling,omitempty"`
	Nationality             string                        `json:"nationality"`            // Full nationality name
	NationalityISO          string                        `json:"nationality_iso"`        // ISO 3166-1 alpha-2 code (or similar)
	NationalityFIFACode     string                        `json:"nationality_fifa_code"`  // FIFA country code
	Attributes              map[string]string             `json:"attributes"`             // Raw attributes from HTML
	NumericAttributes       map[string]int                `json:"-"`                      // Parsed numeric attributes (1-20 scale)
	PerformanceStatsNumeric map[string]float64            `json:"-"`                      // Parsed numeric performance stats
	PerformancePercentiles  map[string]map[string]float64 `json:"performancePercentiles"` // Percentiles for performance stats
	ParsedPositions         []string                      `json:"parsedPositions"`        // Standardized long position names
	ShortPositions          []string                      `json:"shortPositions"`         // Standardized short position codes (e.g., DC, ST)
	PositionGroups          []string                      `json:"positionGroups"`         // Broad groups like "Defenders", "Midfielders"
	PAC                     int                           `json:"PAC"`                    // Calculated Pace category score (0-99)
	SHO                     int                           `json:"SHO"`                    // Calculated Shooting category score (0-99)
	PAS                     int                           `json:"PAS"`                    // Calculated Passing category score (0-99)
	DRI                     int                           `json:"DRI"`                    // Calculated Dribbling category score (0-99)
	DEF                     int                           `json:"DEF"`                    // Calculated Defending category score (0-99)
	PHY                     int                           `json:"PHY"`                    // Calculated Physical category score (0-99)
	GK                      int                           `json:"GK,omitempty"`           // Calculated Goalkeeping category score (0-99)
	DIV                     int                           `json:"DIV,omitempty"`          // Calculated Diving category score (0-99)
	HAN                     int                           `json:"HAN,omitempty"`          // Calculated Handling category score (0-99)
	REF                     int                           `json:"REF,omitempty"`          // Calculated Reflexes category score (0-99)
	KIC                     int                           `json:"KIC,omitempty"`          // Calculated Kicking category score (0-99)
	SPD                     int                           `json:"SPD,omitempty"`          // Calculated Speed category score (0-99)
	POS                     int                           `json:"POS,omitempty"`          // Calculated Positioning category score (0-99)
	Overall                 int                           `json:"Overall"`                // Blended overall score (0-99)
	RoleSpecificOveralls    []RoleOverallScore            `json:"roleSpecificOveralls"`   // Overall scores for specific roles
	TransferValueAmount     int64                         `json:"transferValueAmount"`    // Numeric transfer value
	WageAmount              int64                         `json:"wageAmount"`             // Numeric wage amount
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
