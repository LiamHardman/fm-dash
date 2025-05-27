package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	defaultPlayerCapacity    = 1024            // Initial capacity for player slices
	defaultAttributeCapacity = 64              // Initial capacity for attribute maps
	defaultCellCapacity      = 64              // Initial capacity for cell slices during parsing
	overallScalingFactor     = 5.85            // Used for scaling role-specific attribute averages (1-20) to 0-99
	maxTokenBufferSize       = 2 * 1024 * 1024 // 2MB, for bufio.NewReaderSize

	// Factors for combining role-based and category-based overalls for outfielders
	roleSpecificOverallFactor  = 0.8 // 80% weight for role-specific calculation
	categoryBasedOverallFactor = 0.2 // 20% weight for category-based calculation
)

// Global variables for attribute and role weights.
// These are populated at startup from JSON files or defaults.
var (
	attributeWeights             map[string]map[string]int // For PHY, SHO etc. from individual attributes
	roleSpecificOverallWeights   map[string]map[string]int // For role-specific overall from individual attributes
	muAttributeWeights           sync.RWMutex
	muRoleSpecificOverallWeights sync.RWMutex
	precomputedRoleWeights       map[string][]struct { // Optimized lookup for role weights
		RoleName string
		Weights  map[string]int
	}
	muPrecomputedRoleWeights sync.RWMutex

	// Default/Fallback weights for calculating a general overall based on FIFA stat categories for outfielders
	fifaCategoryOverallWeights = map[string]int{
		"PHY": 25, "MEN": 25, "PAS": 15, "DEF": 15, "DRI": 10, "SHO": 10, // Sums to 100
	}

	// Position-specific weights for FIFA stat categories for outfielders
	attackerFifaCategoryWeights = map[string]int{
		"SHO": 30, "PHY": 25, "DRI": 20, "MEN": 15, "PAS": 10, "DEF": 0, // Sums to 100
	}
	midfielderFifaCategoryWeights = map[string]int{
		"PAS": 30, "MEN": 25, "PHY": 20, "DRI": 15, "DEF": 5, "SHO": 5, // Sums to 100
	}
	defenderFifaCategoryWeights = map[string]int{
		"DEF": 30, "PHY": 30, "MEN": 20, "PAS": 15, "DRI": 5, "SHO": 0, // Sums to 100
	}

	// Metrics collection toggle
	metricsEnabled bool

	// OpenTelemetry metrics
	meter               metric.Meter
	uploadDuration      metric.Float64Histogram
	uploadRowsPerSecond metric.Float64Gauge
	uploadFileSize      metric.Float64Gauge
	uploadPlayersProcessed metric.Float64Gauge
	uploadMemoryUsage   metric.Float64Gauge
	uploadsTotal        metric.Int64Counter
	uploadWorkers       metric.Float64Gauge
)

// Default attribute weights if JSON loading fails or file is missing.
var defaultAttributeWeightsGo = map[string]map[string]int{
	"PHY": {"Acc": 7, "Pac": 6, "Str": 5, "Sta": 4, "Nat": 3, "Bal": 2, "Jum": 1},
	"SHO": {"Fin": 7, "OtB": 6, "Cmp": 5, "Tec": 4, "Hea": 3, "Lon": 2, "Pen": 1},
	"PAS": {"Pas": 7, "Vis": 6, "Tec": 5, "Cro": 4, "Fre": 3, "Cor": 2, "L Th": 1},
	"DRI": {"Dri": 6, "Fir": 5, "Tec": 4, "Agi": 3, "Bal": 2, "Fla": 1},
	"DEF": {"Tck": 6, "Mar": 5, "Hea": 4, "Pos": 3, "Cnt": 2, "Ant": 1},
	"MEN": {"Wor": 11, "Dec": 10, "Tea": 9, "Det": 8, "Bra": 7, "Ldr": 6, "Vis": 5, "Agg": 4, "OtB": 3, "Pos": 2, "Ant": 1},
	"GK":  {"Han": 20, "Ref": 20, "Cmd": 15, "Aer": 15, "1v1": 10, "Kic": 5, "TRO": 5, "Com": 3, "Thr": 3, "Ecc": 1},
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
		"Acc": 90, "Agi": 60, "Bal": 35, "Jum": 65, "Nat": 10, "Pac": 90, "Sta": 30, "Str": 50,
	},
	"DC - Central Defender - Defend": {
		"Cor": 10, "Cro": 10, "Dri": 30, "Fin": 10, "Fir": 30, "Fre": 5, "Hea": 60, "Lon": 0, "L Th": 0, "Mar": 70, "Pas": 40, "Pen": 0, "Tck": 70, "Tec": 30,
		"Agg": 60, "Ant": 65, "Bra": 50, "Cmp": 80, "Cnt": 65, "Dec": 65, "Det": 20, "Fla": 10, "Ldr": 10, "OtB": 10, "Pos": 65, "Tea": 20, "Vis": 30, "Wor": 60,
		"Acc": 60, "Agi": 30, "Bal": 30, "Jum": 65, "Nat": 10, "Pac": 70, "Sta": 40, "Str": 60,
	},
}

// PerformanceStatKeys lists the column headers for player performance statistics.
var PerformanceStatKeys = []string{
	"Asts/90", "Av Rat", "Blk/90", "Ch C/90", "Clr/90", "Cr C/90", "Drb/90",
	"xA/90", "xG/90", "Gls/90", "Hdrs W/90", "Int/90", "K Ps/90", "Ps C/90",
	"Shot/90", "Tck/90", "Poss Won/90", "ShT/90", "Pres C/90", "Poss Lost/90",
	"Pr passes/90", "Conv %", "Tck R", "Pas %", "Cr C/A",
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
		log.Printf("Warning: Could not read %s: %v. Using default weights.", filePath, err)
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

// init is automatically called when the package is loaded.
// It loads attribute and role weights from JSON files or uses defaults.
func init() {
	var errAttr, errRole error

	// Load attribute weights for FIFA stats (PHY, SHO, etc.)
	loadedAttrWeights, errAttr := loadJSONWeights(filepath.Join("public", "attribute_weights.json"), defaultAttributeWeightsGo)
	muAttributeWeights.Lock()
	if errAttr != nil {
		log.Printf("Using default attribute_weights due to error: %v. Default attribute_weights has %d entries.", errAttr, len(defaultAttributeWeightsGo))
		attributeWeights = make(map[string]map[string]int)
		for k, v := range defaultAttributeWeightsGo { // Deep copy
			innerMap := make(map[string]int)
			for ik, iv := range v {
				innerMap[ik] = iv
			}
			attributeWeights[k] = innerMap
		}
	} else {
		attributeWeights = loadedAttrWeights
	}
	muAttributeWeights.Unlock()

	// Load role-specific overall weights
	loadedRoleWeights, errRole := loadJSONWeights(filepath.Join("public", "role_specific_overall_weights.json"), defaultRoleSpecificOverallWeightsGo)
	muRoleSpecificOverallWeights.Lock()
	if errRole != nil {
		log.Printf("Using default role_specific_overall_weights due to error: %v. Default role_specific_overall_weights has %d entries.", errRole, len(defaultRoleSpecificOverallWeightsGo))
		roleSpecificOverallWeights = make(map[string]map[string]int)
		for k, v := range defaultRoleSpecificOverallWeightsGo { // Deep copy
			innerMap := make(map[string]int)
			for ik, iv := range v {
				innerMap[ik] = iv
			}
			roleSpecificOverallWeights[k] = innerMap
		}
	} else {
		roleSpecificOverallWeights = loadedRoleWeights
	}
	muRoleSpecificOverallWeights.Unlock()

	// Precompute role weights for faster lookup during player processing
	muPrecomputedRoleWeights.Lock()
	precomputedRoleWeights = make(map[string][]struct {
		RoleName string
		Weights  map[string]int
	})
	// Use the just loaded (or defaulted) roleSpecificOverallWeights for precomputation
	sourceWeightsToPrecompute := roleSpecificOverallWeights
	for roleFullName, weights := range sourceWeightsToPrecompute {
		// Extract short position key (e.g., "DC" from "DC - Central Defender - Defend")
		// This logic might need adjustment based on the exact format of roleFullName in your JSON
		shortKey := GetShortPositionKeyFromRoleName(roleFullName) // Assumes GetShortPositionKeyFromRoleName is defined, e.g., in positions.go

		if shortKey != "" {
			copiedWeights := make(map[string]int, len(weights)) // Deep copy weights
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

	// Initialize metrics toggle from environment variable
	metricsEnabled = os.Getenv("ENABLE_METRICS") == "true"
	log.Printf("Metrics collection enabled: %v", metricsEnabled)

	// Initialize OpenTelemetry metrics
	if metricsEnabled {
		initMetrics()
	}
}

// initMetrics initializes OpenTelemetry metrics instruments
func initMetrics() {
	meter = otel.Meter("v2fmdash-api")

	var err error
	uploadDuration, err = meter.Float64Histogram(
		"fm24_upload_duration_seconds",
		metric.WithDescription("Time taken to process uploads"),
		metric.WithUnit("s"),
	)
	if err != nil {
		log.Printf("Failed to create upload duration histogram: %v", err)
	}

	uploadRowsPerSecond, err = meter.Float64Gauge(
		"fm24_upload_rows_per_second",
		metric.WithDescription("Rows processed per second in last upload"),
	)
	if err != nil {
		log.Printf("Failed to create upload rows per second gauge: %v", err)
	}

	uploadFileSize, err = meter.Float64Gauge(
		"fm24_upload_file_size_bytes",
		metric.WithDescription("Size of last uploaded file in bytes"),
		metric.WithUnit("By"),
	)
	if err != nil {
		log.Printf("Failed to create upload file size gauge: %v", err)
	}

	uploadPlayersProcessed, err = meter.Float64Gauge(
		"fm24_upload_players_processed",
		metric.WithDescription("Number of players processed in last upload"),
	)
	if err != nil {
		log.Printf("Failed to create upload players processed gauge: %v", err)
	}

	uploadMemoryUsage, err = meter.Float64Gauge(
		"fm24_upload_memory_usage_mb",
		metric.WithDescription("Memory usage during last upload in MB"),
		metric.WithUnit("MB"),
	)
	if err != nil {
		log.Printf("Failed to create upload memory usage gauge: %v", err)
	}

	uploadsTotal, err = meter.Int64Counter(
		"fm24_uploads_total",
		metric.WithDescription("Total number of file uploads processed"),
	)
	if err != nil {
		log.Printf("Failed to create uploads total counter: %v", err)
	}

	uploadWorkers, err = meter.Float64Gauge(
		"fm24_upload_workers",
		metric.WithDescription("Number of workers used in last upload"),
	)
	if err != nil {
		log.Printf("Failed to create upload workers gauge: %v", err)
	}

	log.Printf("OpenTelemetry metrics initialized successfully")
}

// recordUploadMetrics stores metrics for a completed upload if metrics are enabled.
func recordUploadMetrics(filename string, fileSizeBytes int64, totalDuration, parseDuration time.Duration, 
	playersProcessed int, memoryAllocMB float64, numWorkers, numGoroutines int) {
	if !metricsEnabled {
		return
	}

	ctx := context.Background()
	rowsPerSecond := 0.0
	if parseDuration.Seconds() > 0 {
		rowsPerSecond = float64(playersProcessed) / parseDuration.Seconds()
	}

	// Record to OpenTelemetry metrics
	if uploadDuration != nil {
		uploadDuration.Record(ctx, totalDuration.Seconds(), metric.WithAttributes(
			attribute.String("type", "total"),
		))
		uploadDuration.Record(ctx, parseDuration.Seconds(), metric.WithAttributes(
			attribute.String("type", "parse"),
		))
	}
	
	if uploadRowsPerSecond != nil {
		uploadRowsPerSecond.Record(ctx, rowsPerSecond)
	}
	
	if uploadFileSize != nil {
		uploadFileSize.Record(ctx, float64(fileSizeBytes))
	}
	
	if uploadPlayersProcessed != nil {
		uploadPlayersProcessed.Record(ctx, float64(playersProcessed))
	}
	
	if uploadMemoryUsage != nil {
		uploadMemoryUsage.Record(ctx, memoryAllocMB)
	}
	
	if uploadWorkers != nil {
		uploadWorkers.Record(ctx, float64(numWorkers))
	}
	
	if uploadsTotal != nil {
		uploadsTotal.Add(ctx, 1)
	}
}
