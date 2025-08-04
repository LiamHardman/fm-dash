package main

import (
	"context"
	"fmt"
	"sync"
	"time"
	"unsafe"

	"go.opentelemetry.io/otel/attribute"
)

// AdvancedMemoryOptimizer provides aggressive memory optimization
type AdvancedMemoryOptimizer struct {
	useOptimizedStructs bool
	useBitPacking       bool
	useStringPooling    bool
	useMemoryMapping    bool

	// String pools for common values
	clubPool        *StringPool
	positionPool    *StringPool
	nationalityPool *StringPool

	// Conversion stats
	conversions struct {
		regular   int64
		optimized int64
		byteSaved int64
		mu        sync.RWMutex
	}
}

// StringPool manages a pool of frequently used strings
type StringPool struct {
	pool map[string]string
	mu   sync.RWMutex
}

// NewStringPool creates a new string pool
func NewStringPool() *StringPool {
	return &StringPool{
		pool: make(map[string]string),
	}
}

// Intern returns an interned version of the string
func (sp *StringPool) Intern(s string) string {
	if s == "" {
		return ""
	}

	sp.mu.RLock()
	if interned, exists := sp.pool[s]; exists {
		sp.mu.RUnlock()
		return interned
	}
	sp.mu.RUnlock()

	sp.mu.Lock()
	defer sp.mu.Unlock()

	// Double-check after acquiring write lock
	if interned, exists := sp.pool[s]; exists {
		return interned
	}

	// Store the string in the pool
	sp.pool[s] = s
	return s
}

// GlobalAdvancedOptimizer is the global instance
var GlobalAdvancedOptimizer = &AdvancedMemoryOptimizer{
	useOptimizedStructs: true,
	useBitPacking:       true,
	useStringPooling:    true,
	useMemoryMapping:    false, // Enable for very large datasets

	clubPool:        NewStringPool(),
	positionPool:    NewStringPool(),
	nationalityPool: NewStringPool(),
}

// OptimizePlayersAdvanced applies aggressive memory optimizations
func (amo *AdvancedMemoryOptimizer) OptimizePlayersAdvanced(ctx context.Context, players []Player) ([]OptimizedPlayer, int64) {
	ctx, span := StartSpan(ctx, "memory.optimize_advanced")
	defer span.End()

	startTime := time.Now()
	originalSize := amo.estimatePlayerSliceMemory(players)

	optimized := make([]OptimizedPlayer, len(players))

	for i := range players {
		optimized[i] = *amo.ConvertToOptimizedAdvanced(&players[i])
	}

	optimizedSize := amo.estimateOptimizedSliceMemory(optimized)
	saved := originalSize - optimizedSize

	// Update stats
	amo.conversions.mu.Lock()
	amo.conversions.regular += int64(len(players))
	amo.conversions.optimized += int64(len(optimized))
	amo.conversions.byteSaved += saved
	amo.conversions.mu.Unlock()

	SetSpanAttributes(ctx,
		attribute.Int("players.count", len(players)),
		attribute.Int64("memory.original_bytes", originalSize),
		attribute.Int64("memory.optimized_bytes", optimizedSize),
		attribute.Int64("memory.saved_bytes", saved),
		attribute.Float64("memory.reduction_percent", float64(saved)/float64(originalSize)*100),
		attribute.Int64("optimization.duration_ms", time.Since(startTime).Milliseconds()),
	)

	logInfo(ctx, "Advanced optimization completed",
		"player_count", len(players),
		"original_size_mb", float64(originalSize)/1024/1024,
		"optimized_size_mb", float64(optimizedSize)/1024/1024,
		"saved_mb", float64(saved)/1024/1024,
		"reduction_percent", float64(saved)/float64(originalSize)*100,
		"per_player_bytes", optimizedSize/int64(len(optimized)))

	return optimized, saved
}

// ConvertToOptimizedAdvanced converts a Player to OptimizedPlayer with aggressive optimizations
func (amo *AdvancedMemoryOptimizer) ConvertToOptimizedAdvanced(player *Player) *OptimizedPlayer {
	opt := &OptimizedPlayer{
		TransferValueAmount: player.TransferValueAmount,
		WageAmount:          player.WageAmount,
		Overall:             int32(clampInt32(player.Overall)),
		UID:                 player.UID,
		Name:                player.Name, // Keep name as-is (unique)

		// Aggressive string interning for common values
		Club:                amo.clubPool.Intern(player.Club),
		Position:            amo.positionPool.Intern(player.Position),
		Division:            amo.clubPool.Intern(player.Division), // Clubs and divisions overlap
		Nationality:         amo.nationalityPool.Intern(player.Nationality),
		NationalityISO:      amo.nationalityPool.Intern(player.NationalityISO),
		NationalityFIFACode: amo.nationalityPool.Intern(player.NationalityFIFACode),
		BestRoleOverall:     amo.positionPool.Intern(player.BestRoleOverall),

		// FIFA stats as int16
		PAC:        int16(clampInt16(player.PAC)),
		SHO:        int16(clampInt16(player.SHO)),
		PAS:        int16(clampInt16(player.PAS)),
		DRI:        int16(clampInt16(player.DRI)),
		DEF:        int16(clampInt16(player.DEF)),
		PHY:        int16(clampInt16(player.PHY)),
		TotalStats: int16(clampInt16(player.TotalStats)),
		MBR:        int16(clampInt16(player.MBR)),

		// GK stats
		GK:  int16(clampInt16(player.GK)),
		DIV: int16(clampInt16(player.DIV)),
		HAN: int16(clampInt16(player.HAN)),
		REF: int16(clampInt16(player.REF)),
		KIC: int16(clampInt16(player.KIC)),
		SPD: int16(clampInt16(player.SPD)),
		POS: int16(clampInt16(player.POS)),

		AttributeMasked:      player.AttributeMasked,
		RoleSpecificOveralls: player.RoleSpecificOveralls,
	}

	// Parse age safely
	if age, err := safeParseInt(player.Age); err == nil {
		opt.Age = int32(clampInt32(age))
	}

	// Pack FM attributes into compact arrays
	amo.packAttributes(player, opt)

	// Pack positions into fixed arrays
	amo.packPositions(player, opt)

	// Create extended data only if needed
	if amo.hasExtendedData(player) {
		opt.Extended = &PlayerExtendedData{
			TransferValue:           player.TransferValue,
			Wage:                    player.Wage,
			Personality:             amo.clubPool.Intern(player.Personality),
			MediaHandling:           amo.clubPool.Intern(player.MediaHandling),
			PerformanceStatsNumeric: player.PerformanceStatsNumeric,
			PerformancePercentiles:  player.PerformancePercentiles,
		}
	}

	return opt
}

// packAttributes converts map[string]int to compact arrays
func (amo *AdvancedMemoryOptimizer) packAttributes(player *Player, opt *OptimizedPlayer) {
	// Technical attributes mapping
	technicalAttrs := []string{"Cor", "Cro", "Dri", "Fin", "Fir", "Fre", "Hea", "Lon", "L Th", "Mar", "Pas", "Pen", "Tck", "Tec", "Thr", "Fla", "Ant", "Cmp", "Cnt", "OtB"}
	for i, attr := range technicalAttrs {
		if i < len(opt.TechnicalAttributes) {
			if val, exists := player.NumericAttributes[attr]; exists {
				opt.TechnicalAttributes[i] = int8(clampInt8(val))
			}
		}
	}

	// Mental attributes
	mentalAttrs := []string{"Agg", "Ant", "Bra", "Cmp", "Cnt", "Dec", "Det", "Fla", "Ldr", "OtB", "Pos", "Tea", "Vis", "Wor", "Bra"}
	for i, attr := range mentalAttrs {
		if i < len(opt.MentalAttributes) {
			if val, exists := player.NumericAttributes[attr]; exists {
				opt.MentalAttributes[i] = int8(clampInt8(val))
			}
		}
	}

	// Physical attributes
	physicalAttrs := []string{"Acc", "Agi", "Bal", "Jum", "Nat", "Pac", "Sta", "Str"}
	for i, attr := range physicalAttrs {
		if i < len(opt.PhysicalAttributes) {
			if val, exists := player.NumericAttributes[attr]; exists {
				opt.PhysicalAttributes[i] = int8(clampInt8(val))
			}
		}
	}

	// Goalkeeper attributes
	gkAttrs := []string{"Aer", "Cmd", "Com", "Ecc", "Han", "Kic", "1v1", "Pun", "Ref", "TRO"}
	for i, attr := range gkAttrs {
		if i < len(opt.GoalkeeperAttributes) {
			if val, exists := player.NumericAttributes[attr]; exists {
				opt.GoalkeeperAttributes[i] = int8(clampInt8(val))
			}
		}
	}
}

// packPositions converts slices to fixed arrays
func (amo *AdvancedMemoryOptimizer) packPositions(player *Player, opt *OptimizedPlayer) {
	// Pack parsed positions
	opt.PositionCount = int8(min(len(player.ParsedPositions), len(opt.ParsedPositions)))
	for i := 0; i < int(opt.PositionCount); i++ {
		opt.ParsedPositions[i] = amo.positionPool.Intern(player.ParsedPositions[i])
	}

	// Pack short positions
	opt.ShortPositionCount = int8(min(len(player.ShortPositions), len(opt.ShortPositions)))
	for i := 0; i < int(opt.ShortPositionCount); i++ {
		opt.ShortPositions[i] = amo.positionPool.Intern(player.ShortPositions[i])
	}

	// Pack position groups
	opt.PositionGroupsCount = int8(min(len(player.PositionGroups), len(opt.PositionGroups)))
	for i := 0; i < int(opt.PositionGroupsCount); i++ {
		opt.PositionGroups[i] = amo.positionPool.Intern(player.PositionGroups[i])
	}

	// Check if goalkeeper
	for _, group := range player.PositionGroups {
		if group == "Goalkeeper" {
			opt.IsGoalkeeper = true
			break
		}
	}
}

// hasExtendedData checks if the player needs extended data
func (amo *AdvancedMemoryOptimizer) hasExtendedData(player *Player) bool {
	return len(player.PerformanceStatsNumeric) > 0 ||
		len(player.PerformancePercentiles) > 0 ||
		player.Personality != "" ||
		player.MediaHandling != ""
}

// estimatePlayerSliceMemory estimates memory usage of regular players
func (amo *AdvancedMemoryOptimizer) estimatePlayerSliceMemory(players []Player) int64 {
	if len(players) == 0 {
		return 0
	}

	sampleSize := min(len(players), 10)
	totalSize := int64(0)

	for i := 0; i < sampleSize; i++ {
		totalSize += amo.estimatePlayerMemory(&players[i])
	}

	avgSize := totalSize / int64(sampleSize)
	return avgSize * int64(len(players))
}

// estimateOptimizedSliceMemory estimates memory usage of optimized players
func (amo *AdvancedMemoryOptimizer) estimateOptimizedSliceMemory(players []OptimizedPlayer) int64 {
	if len(players) == 0 {
		return 0
	}

	baseSize := int64(unsafe.Sizeof(OptimizedPlayer{}))
	totalSize := baseSize * int64(len(players))

	// Add extended data size
	for _, player := range players {
		if player.Extended != nil {
			totalSize += int64(unsafe.Sizeof(*player.Extended))
			// Add map sizes
			totalSize += int64(len(player.Extended.PerformanceStatsNumeric) * 32)
			totalSize += int64(len(player.Extended.PerformancePercentiles) * 64)
		}

		// Add role overalls
		totalSize += int64(len(player.RoleSpecificOveralls)) * int64(unsafe.Sizeof(RoleOverallScore{}))
	}

	return totalSize
}

// estimatePlayerMemory estimates memory usage of a single regular player
func (amo *AdvancedMemoryOptimizer) estimatePlayerMemory(player *Player) int64 {
	size := int64(unsafe.Sizeof(*player))

	// Add string sizes
	size += int64(len(player.Name) + len(player.Position) + len(player.Age) +
		len(player.Club) + len(player.Division) + len(player.TransferValue) + len(player.Wage) +
		len(player.Personality) + len(player.MediaHandling) + len(player.Nationality) +
		len(player.NationalityISO) + len(player.NationalityFIFACode) + len(player.BestRoleOverall))

	// Add map overhead
	size += int64(len(player.Attributes) * 32)
	size += int64(len(player.NumericAttributes) * 24)
	size += int64(len(player.PerformanceStatsNumeric) * 24)

	// Add nested maps
	for _, innerMap := range player.PerformancePercentiles {
		size += int64(len(innerMap) * 24)
	}
	size += int64(len(player.PerformancePercentiles) * 32)

	// Add slices
	for _, pos := range player.ParsedPositions {
		size += int64(len(pos))
	}
	for _, pos := range player.ShortPositions {
		size += int64(len(pos))
	}
	for _, group := range player.PositionGroups {
		size += int64(len(group))
	}

	size += int64(len(player.RoleSpecificOveralls)) * int64(unsafe.Sizeof(RoleOverallScore{}))

	return size
}

// GetOptimizationStats returns optimization statistics
func (amo *AdvancedMemoryOptimizer) GetOptimizationStats() map[string]interface{} {
	amo.conversions.mu.RLock()
	defer amo.conversions.mu.RUnlock()

	stats := map[string]interface{}{
		"regular_players_processed": amo.conversions.regular,
		"optimized_players_created": amo.conversions.optimized,
		"total_bytes_saved":         amo.conversions.byteSaved,
		"total_mb_saved":            float64(amo.conversions.byteSaved) / 1024 / 1024,
		"use_optimized_structs":     amo.useOptimizedStructs,
		"use_bit_packing":           amo.useBitPacking,
		"use_string_pooling":        amo.useStringPooling,
		"use_memory_mapping":        amo.useMemoryMapping,
	}

	if amo.conversions.regular > 0 {
		stats["avg_bytes_saved_per_player"] = amo.conversions.byteSaved / amo.conversions.regular
		stats["memory_reduction_percent"] = float64(amo.conversions.byteSaved) / (float64(amo.conversions.regular) * 65536) * 100 // Assume 64KB per player
	}

	// String pool stats
	amo.clubPool.mu.RLock()
	stats["club_pool_size"] = len(amo.clubPool.pool)
	amo.clubPool.mu.RUnlock()

	amo.positionPool.mu.RLock()
	stats["position_pool_size"] = len(amo.positionPool.pool)
	amo.positionPool.mu.RUnlock()

	amo.nationalityPool.mu.RLock()
	stats["nationality_pool_size"] = len(amo.nationalityPool.pool)
	amo.nationalityPool.mu.RUnlock()

	return stats
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func safeParseInt(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	// Simple integer parsing
	result := 0
	for _, char := range s {
		if char >= '0' && char <= '9' {
			result = result*10 + int(char-'0')
		} else {
			return 0, fmt.Errorf("invalid character")
		}
	}
	return result, nil
}
