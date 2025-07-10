package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Log level constants
//
// Usage:
//   - Set LOG_LEVEL environment variable to: DEBUG, INFO, WARN, or CRITICAL
//   - Default is INFO level
//   - Use LogDebug(), LogInfo(), LogWarn(), LogCritical() functions for leveled logging
//   - Messages below the minimum log level will be filtered out
const (
	LogLevelDebug    = 0
	LogLevelInfo     = 1
	LogLevelWarn     = 2
	LogLevelCritical = 3
)

// logLevelNames maps log levels to their string representations
var logLevelNames = map[int]string{
	LogLevelDebug:    "DEBUG",
	LogLevelInfo:     "INFO",
	LogLevelWarn:     "WARN",
	LogLevelCritical: "CRITICAL",
}

// parseLogLevel converts a string log level to its integer constant
func parseLogLevel(level string) int {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return LogLevelDebug
	case "INFO":
		return LogLevelInfo
	case "WARN", "WARNING":
		return LogLevelWarn
	case "CRITICAL", "CRIT", "ERROR":
		return LogLevelCritical
	default:
		return LogLevelInfo // Default to info if unrecognized
	}
}

const (
	// Default capacity constants for better performance
	defaultPlayerCapacity    = 2048            // Increased from 1024 for better pre-allocation
	defaultAttributeCapacity = 80              // Increased from 64 for FM attributes
	defaultCellCapacity      = 80              // Increased from 64 for table cells
	overallScalingFactor     = 5.85            // Used for scaling role-specific attribute averages (1-20) to 0-99
	maxTokenBufferSize       = 4 * 1024 * 1024 // Increased from 2MB to 4MB for larger files

	// General limits
	// maxFileSize           = 50 * 1024 * 1024 // 50MB file size limit
	// maxPlayers            = 200000           // Maximum number of players to process
	// maxDatasetRetentionMB = 500              // Maximum size for datasets to keep in memory

	// Timing constants for async calculation optimization
	// asyncProcessingTimeout    = 30 * time.Second // Timeout for async operations
	// percentileThreshold       = 300              // Player count threshold for async percentiles
	// cacheSize                 = 1000             // Maximum cache entries
	// workerPoolSize            = 5                // Number of workers for background tasks
	// minPlayersForPercentiles  = 10               // Minimum players required for meaningful percentiles
)

// Global variables for attribute and role weights.
// These are populated at startup from JSON files or defaults.
var (
	attributeWeights             map[string]map[string]int // For PAC, SHO etc. from individual attributes
	roleSpecificOverallWeights   map[string]map[string]int // For role-specific overall from individual attributes
	muAttributeWeights           sync.RWMutex
	muRoleSpecificOverallWeights sync.RWMutex
	precomputedRoleWeights       map[string][]struct { // Optimized lookup for role weights
		RoleName string
		Weights  map[string]int
	}
	muPrecomputedRoleWeights sync.RWMutex

	// Default/Fallback weights for calculating a general overall based on FIFA stat categories for outfielders
	// fifaCategoryOverallWeights = map[string]int{
	// 	"PHY": 15, "PAC": 30, "PAS": 15, "DEF": 15, "DRI": 15, "SHO": 10, // Sums to 100
	// }

	// Position-specific weights for FIFA stat categories for outfielders
	// attackerFifaCategoryWeights = map[string]int{
	// 	"SHO": 25, "PAC": 35, "DRI": 20, "PHY": 15, "PAS": 5, "DEF": 0, // Sums to 100
	// }
	// midfielderFifaCategoryWeights = map[string]int{
	// 	"PAS": 25, "PHY": 15, "PAC": 30, "DRI": 15, "DEF": 10, "SHO": 5, // Sums to 100
	// }
	// defenderFifaCategoryWeights = map[string]int{
	// 	"DEF": 25, "PHY": 20, "PAC": 30, "PAS": 15, "DRI": 5, "SHO": 5, // Sums to 100
	// }

	// Metrics collection toggle
	metricsEnabled bool

	// Rating calculation method toggle
	useScaledRatings   bool = true // Default to new scaled ratings
	muUseScaledRatings sync.RWMutex

	// Logging configuration
	logAllRequests bool = false        // Default to only logging non-200 responses
	minLogLevel    int  = LogLevelInfo // Default to info level
	muLogLevel     sync.RWMutex
)

// Default attribute weights if JSON loading fails or file is missing.
var defaultAttributeWeightsGo = map[string]map[string]int{
	"PAC": {"Acc": 12, "Pac": 12, "Agi": 5},
	"SHO": {"Fin": 8, "Lon": 6, "Pen": 4, "Hea": 5, "Cmp": 6, "Tec": 5, "Ant": 4, "Dec": 4, "Fla": 3},
	"PAS": {"Pas": 8, "Cro": 6, "Fre": 4, "Vis": 7, "Tec": 5, "Tea": 4, "Dec": 4, "Cor": 3, "Fir": 4, "OtB": 3},
	"DRI": {"Dri": 8, "Fir": 7, "Tec": 6, "Fla": 5, "Cmp": 4, "OtB": 3},
	"DEF": {"Mar": 8, "Tck": 8, "Hea": 6, "Ant": 7, "Cnt": 6, "Pos": 7, "Dec": 5, "Cmp": 4, "Bra": 5, "Agg": 4, "Wor": 4},
	"PHY": {"Str": 8, "Sta": 7, "Nat": 6, "Jum": 5, "Bal": 4, "Agg": 5, "Bra": 4, "Wor": 4},
	"GK":  {"Han": 20, "Ref": 20, "Cmd": 15, "Aer": 15, "1v1": 10, "Kic": 5, "TRO": 5, "Com": 3, "Thr": 3, "Ecc": 1},
	"DIV": {"Aer": 8, "Ref": 7, "Agi": 6, "1v1": 7, "Han": 5},
	"HAN": {"Han": 10, "Cmd": 7, "Cmp": 5, "Cnt": 4},
	"REF": {"Ref": 10, "Ant": 6, "Cnt": 5, "1v1": 5},
	"KIC": {"Kic": 8, "Thr": 6, "Tec": 5, "Vis": 4, "Pas": 3},
	"SPD": {"Acc": 8, "Pac": 8, "TRO": 6},
	"POS": {"Pos": 8, "Cmd": 7, "Ant": 6, "Dec": 5, "TRO": 4, "Cnt": 4, "Com": 3},
}

// Default role-specific overall weights if JSON loading fails or file is missing.
// "Generic" roles have been removed from this default map.
var defaultRoleSpecificOverallWeightsGo = map[string]map[string]int{
	"GK - Goalkeeper - Defend": {"Han": 90, "Ref": 90, "Aer": 80, "Cmd": 75, "1v1": 80, "Cnt": 70, "Dec": 70, "Pos": 75, "Ant": 60, "Cmp": 60, "Bra": 60, "Com": 50, "Kic": 40, "Thr": 40, "TRO": 30, "Det": 50, "Ldr": 40, "Wor": 40, "Tea": 40, "Agi": 50, "Jum": 60, "Str": 50, "Acc": 30, "Pac": 30, "Ecc": 10},
	// Add other non-generic default roles here if you have them.
	// For example, if you had a "DC - Ball Playing Defender - Defend" as a default, it would remain.
	// Example (assuming it was a default, keeping it for structure):
	"DC - Ball Playing Defender - Defend": {
		"Cor": 5, "Cro": 1, "Dri": 40, "Fin": 10, "Fir": 35, "Fre": 10, "Hea": 55, "Lon": 10, "Tea": 20, "L Th": 0, "Mar": 55, "Pas": 55, "Pen": 10, "Tck": 40, "Tec": 35,
		"Agg": 40, "Ant": 50, "Bra": 30, "Cmp": 80, "Cnt": 50, "Dec": 50, "Det": 20, "Fla": 10, "Ldr": 10, "OtB": 10, "Pos": 55, "Vis": 50, "Wor": 55,
		"Acc": 100, "Agi": 60, "Bal": 35, "Jum": 65, "Nat": 10, "Pac": 100, "Sta": 30, "Str": 50,
	},
	"DC - Central Defender - Defend": {
		"Cor": 10, "Cro": 10, "Dri": 30, "Fin": 10, "Fir": 30, "Fre": 5, "Hea": 60, "Lon": 0, "L Th": 0, "Mar": 70, "Pas": 40, "Pen": 0, "Tck": 70, "Tec": 30,
		"Agg": 60, "Ant": 65, "Bra": 50, "Cmp": 80, "Cnt": 65, "Dec": 65, "Det": 20, "Fla": 10, "Ldr": 10, "OtB": 10, "Pos": 65, "Tea": 20, "Vis": 30, "Wor": 60,
		"Acc": 80, "Agi": 30, "Bal": 30, "Jum": 65, "Nat": 10, "Pac": 90, "Sta": 40, "Str": 60,
	},
}

// PerformanceStatKeys lists the column headers for player performance statistics.
var PerformanceStatKeys = []string{
	"Asts/90", "Av Rat", "Blk/90", "Ch C/90", "Clr/90", "Cr C/90", "Drb/90",
	"xA/90", "xG/90", "Gls/90", "Hdrs W/90", "Int/90", "K Ps/90", "Ps C/90",
	"Shot/90", "Tck/90", "Poss Won/90", "ShT/90", "Pres C/90", "Poss Lost/90",
	"Pr passes/90", "Conv %", "Tck R", "Pas %", "Cr C/A",
	// New performance stats
	"Fls", "Apps", "NP-xG/90", "Ps A/90", "Mins", "Clean Sheets", "FA", "CRS A/90",
	// Goalkeeper-specific stats
	"Con/90", "Cln/90", "xGP/90", "Sv %",
}

// PositionGroupsForPercentiles defines broad player categories used for percentile calculations.
var PositionGroupsForPercentiles = []string{"Goalkeepers", "Defenders", "Midfielders", "Attackers"}

// DetailedPositionGroupsForPercentiles maps more specific role groups to their short position codes.
var DetailedPositionGroupsForPercentiles = map[string][]string{
	"Full-backs":                      {"DR", "DL"},
	"Centre-backs":                    {"DC"},
	"Wing-backs":                      {"WBR", "WBL"},
	"Defensive Midfielders":           {"DM"},
	"Central Midfielders":             {"MC"},
	"Wide Midfielders":                {"MR", "ML"},
	"Attacking Midfielders (Central)": {"AMC"},
	"Wingers":                         {"AMR", "AML"},
	"Strikers":                        {"ST"},
}

// loadJSONWeights attempts to load weights from a JSON file.
// If loading fails, it falls back to the provided default weights.
func loadJSONWeights(filePath string, defaultWeights map[string]map[string]int) (map[string]map[string]int, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		LogWarn("Could not read %s: %v. Using default weights.", filePath, err)
		// Return a deep copy of default weights
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
		LogWarn("Could not unmarshal %s: %v. Using default weights.", filePath, err)
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
		LogWarn("Weights file %s was loaded but is empty. Using default weights.", filePath)
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

	LogDebug("Successfully loaded weights from %s with %d entries.", filePath, len(weights))
	return weights, nil
}

// Configuration loading state
var (
	configInitialized = false
	configInitOnce    sync.Once
	configInitError   error
)

// init is called when the package is loaded but only does minimal setup
func init() {
	// Only initialize metrics toggle from environment variable
	metricsEnabled = os.Getenv("ENABLE_METRICS") == "true"
	LogInfo("Metrics collection enabled: %v", metricsEnabled)

	// Initialize logging configuration from environment variable
	logAllRequests = os.Getenv("LOG_ALL_REQUESTS") == "true"
	LogInfo("Log all requests enabled: %v", logAllRequests)

	// Initialize log level from environment variable
	logLevelStr := os.Getenv("LOG_LEVEL")
	if logLevelStr == "" {
		logLevelStr = "INFO" // Default to INFO
	}
	SetMinLogLevel(parseLogLevel(logLevelStr))
	LogInfo("Log level set to: %s", logLevelNames[GetMinLogLevel()])

	// Initialize OpenTelemetry metrics if enabled (lightweight operation)
	if metricsEnabled {
		initMetrics()
		initEnhancedMetrics()
	}
}

// initializeConfigAsync loads configuration files asynchronously
func initializeConfigAsync() {
	defer func() {
		configInitialized = true
		if configInitError != nil {
			LogWarn("Configuration initialization completed with errors: %v", configInitError)
		} else {
			LogDebug("Configuration initialization completed successfully")
		}
	}()

	var wg sync.WaitGroup
	var attrErr, roleErr error

	// Load attribute weights asynchronously
	wg.Add(1)
	go func() {
		defer wg.Done()
		loadedAttrWeights, err := loadJSONWeights(filepath.Join("public", "attribute_weights.json"), defaultAttributeWeightsGo)

		muAttributeWeights.Lock()
		defer muAttributeWeights.Unlock()

		if err != nil {
			LogWarn("Using default attribute_weights due to error: %v. Default attribute_weights has %d entries.", err, len(defaultAttributeWeightsGo))
			attributeWeights = deepCopyWeights(defaultAttributeWeightsGo)
			attrErr = err
		} else {
			attributeWeights = loadedAttrWeights
		}
	}()

	// Load role-specific weights asynchronously
	wg.Add(1)
	go func() {
		defer wg.Done()
		loadedRoleWeights, err := loadJSONWeights(filepath.Join("public", "role_specific_overall_weights.json"), defaultRoleSpecificOverallWeightsGo)

		muRoleSpecificOverallWeights.Lock()
		defer muRoleSpecificOverallWeights.Unlock()

		if err != nil {
			LogWarn("Using default role_specific_overall_weights due to error: %v. Default role_specific_overall_weights has %d entries.", err, len(defaultRoleSpecificOverallWeightsGo))
			roleSpecificOverallWeights = deepCopyWeights(defaultRoleSpecificOverallWeightsGo)
			roleErr = err
		} else {
			roleSpecificOverallWeights = loadedRoleWeights
		}
	}()

	// Wait for both file loads to complete
	wg.Wait()

	// Precompute role weights after both loads complete
	precomputeRoleWeights()

	// Set overall error status
	if attrErr != nil || roleErr != nil {
		configInitError = fmt.Errorf("attribute error: %v, role error: %v", attrErr, roleErr)
	}
}

// deepCopyWeights creates a deep copy of weight maps
func deepCopyWeights(source map[string]map[string]int) map[string]map[string]int {
	result := make(map[string]map[string]int, len(source))
	for k, v := range source {
		innerMap := make(map[string]int, len(v))
		for ik, iv := range v {
			innerMap[ik] = iv
		}
		result[k] = innerMap
	}
	return result
}

// precomputeRoleWeights computes role weights for faster lookup
func precomputeRoleWeights() {
	muPrecomputedRoleWeights.Lock()
	defer muPrecomputedRoleWeights.Unlock()

	precomputedRoleWeights = make(map[string][]struct {
		RoleName string
		Weights  map[string]int
	})

	muRoleSpecificOverallWeights.RLock()
	sourceWeights := roleSpecificOverallWeights
	muRoleSpecificOverallWeights.RUnlock()

	for roleFullName, weights := range sourceWeights {
		shortKey := GetShortPositionKeyFromRoleName(roleFullName)
		if shortKey != "" {
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
	LogDebug("Precomputed %d base position keys for role weights.", len(precomputedRoleWeights))
}

// EnsureConfigInitialized waits for configuration to be loaded (with timeout)
func EnsureConfigInitialized(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for !configInitialized && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	if !configInitialized {
		return fmt.Errorf("configuration initialization timed out after %v", timeout)
	}

	return configInitError
}

// GetUseScaledRatings returns the current rating calculation method preference
func GetUseScaledRatings() bool {
	muUseScaledRatings.RLock()
	defer muUseScaledRatings.RUnlock()
	return useScaledRatings
}

// SetUseScaledRatings sets the rating calculation method preference
func SetUseScaledRatings(useScaled bool) {
	muUseScaledRatings.Lock()
	defer muUseScaledRatings.Unlock()
	useScaledRatings = useScaled
	LogInfo("Rating calculation method changed to: %s", map[bool]string{true: "scaled", false: "linear"}[useScaled])
}

// GetMinLogLevel returns the current minimum log level
func GetMinLogLevel() int {
	muLogLevel.RLock()
	defer muLogLevel.RUnlock()
	return minLogLevel
}

// SetMinLogLevel sets the minimum log level
func SetMinLogLevel(level int) {
	muLogLevel.Lock()
	defer muLogLevel.Unlock()
	minLogLevel = level
}

// shouldLog checks if a message at the given level should be logged
func shouldLog(level int) bool {
	return level >= GetMinLogLevel()
}

// Log helper functions that respect the minimum log level and use slog for SignOz integration
func LogDebug(format string, args ...interface{}) {
	if shouldLog(LogLevelDebug) {
		// Use slog so logs go to SignOz via OTLP handler
		slog.Debug("[DEBUG] " + fmt.Sprintf(format, args...))
	}
}

func LogInfo(format string, args ...interface{}) {
	if shouldLog(LogLevelInfo) {
		// Use slog so logs go to SignOz via OTLP handler
		slog.Info("[INFO] " + fmt.Sprintf(format, args...))
	}
}

func LogWarn(format string, args ...interface{}) {
	if shouldLog(LogLevelWarn) {
		// Use slog so logs go to SignOz via OTLP handler
		slog.Warn("[WARN] " + fmt.Sprintf(format, args...))
	}
}

func LogCritical(format string, args ...interface{}) {
	if shouldLog(LogLevelCritical) {
		// Use slog so logs go to SignOz via OTLP handler
		slog.Error("[CRITICAL] " + fmt.Sprintf(format, args...))
	}
}
