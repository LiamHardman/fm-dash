package main

import (
	"bytes"
	"compress/gzip"
	"log"
	"strings"
	"sync"
)

// CompressedStringInterning provides memory-efficient string deduplication with compression
type CompressedStringInterning struct {
	strings map[string]string
	mutex   sync.RWMutex
	stats   struct {
		totalRequests int64
		cacheHits     int64
		memSaved      int64
		compressions  int64
	}
	compressionThreshold int // Minimum string length to consider compression
}

// Enhanced global string intern pools with compression
var (
	enhancedClubInterning        = CreateCompressedStringInterning(50) // Club names can be long
	enhancedPersonalityInterning = CreateCompressedStringInterning(30) // Personality descriptions
)

// CreateCompressedStringInterning creates a new compressed string interning instance
func CreateCompressedStringInterning(compressionThreshold int) *CompressedStringInterning {
	return &CompressedStringInterning{
		strings:              make(map[string]string),
		compressionThreshold: compressionThreshold,
	}
}

// SmartIntern uses compression for long strings and regular interning for short ones
func (csi *CompressedStringInterning) SmartIntern(s string) string {
	if s == "" {
		return s
	}

	csi.mutex.RLock()
	if interned, exists := csi.strings[s]; exists {
		csi.mutex.RUnlock()
		csi.stats.cacheHits++
		return csi.decompressIfNeeded(interned)
	}
	csi.mutex.RUnlock()

	csi.mutex.Lock()
	defer csi.mutex.Unlock()

	// Double-check after acquiring write lock
	if interned, exists := csi.strings[s]; exists {
		csi.stats.cacheHits++
		return csi.decompressIfNeeded(interned)
	}

	// Decide whether to compress
	var stored string
	if len(s) >= csi.compressionThreshold && csi.shouldCompress(s) {
		if compressed, err := csi.compressString(s); err == nil && len(compressed) < len(s) {
			stored = compressed
			csi.stats.compressions++
		} else {
			stored = s // Fall back to uncompressed if compression fails or doesn't help
		}
	} else {
		stored = s
	}

	// Store the string and return original
	csi.strings[s] = stored
	csi.stats.totalRequests++
	csi.stats.memSaved += int64(len(s) - len(stored))
	return s
}

// shouldCompress determines if a string is worth compressing based on content
func (csi *CompressedStringInterning) shouldCompress(s string) bool {
	// Don't compress if mostly unique characters (won't compress well)
	if float64(len(csi.uniqueChars(s)))/float64(len(s)) > 0.8 {
		return false
	}

	// Good candidates: repeated patterns, common words
	return strings.Contains(s, " ") || // Multi-word strings
		strings.Count(s, strings.ToLower(s[:1])) > 1 || // Repeated first character
		len(s) > 100 // Very long strings
}

// uniqueChars counts unique characters in a string
func (csi *CompressedStringInterning) uniqueChars(s string) map[rune]bool {
	unique := make(map[rune]bool)
	for _, r := range s {
		unique[r] = true
	}
	return unique
}

// compressString compresses a string using gzip
func (csi *CompressedStringInterning) compressString(s string) (string, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	if _, err := gz.Write([]byte(s)); err != nil {
		if closeErr := gz.Close(); closeErr != nil {
			log.Printf("Failed to close gzip writer: %v", closeErr)
		}
		return "", err
	}

	if err := gz.Close(); err != nil {
		return "", err
	}

	// Prefix with a marker to indicate this is compressed
	return "\x00GZIP\x00" + buf.String(), nil
}

// decompressIfNeeded decompresses a string if it was compressed
func (csi *CompressedStringInterning) decompressIfNeeded(stored string) string {
	if !strings.HasPrefix(stored, "\x00GZIP\x00") {
		return stored // Not compressed
	}

	compressed := stored[7:] // Remove the marker

	reader, err := gzip.NewReader(strings.NewReader(compressed))
	if err != nil {
		return stored // Return as-is on error
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Failed to close gzip reader: %v", closeErr)
		}
	}()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(reader); err != nil {
		return stored // Return as-is on error
	}

	return buf.String()
}

// EnhancedOptimizePlayerStrings uses the new compressed interning system
func EnhancedOptimizePlayerStrings(player *Player) {
	// Use regular interning for short, frequently repeated strings
	player.Club = safeClubInterning.SafeIntern(player.Club)
	player.Position = safePositionInterning.SafeIntern(player.Position)
	player.Division = safeDivisionInterning.SafeIntern(player.Division)
	player.Nationality = safeNationalityInterning.SafeIntern(player.Nationality)
	player.NationalityISO = safeNationalityInterning.SafeIntern(player.NationalityISO)
	player.NationalityFIFACode = safeNationalityInterning.SafeIntern(player.NationalityFIFACode)

	// Use compressed interning for potentially longer strings
	if player.Personality != "" {
		player.Personality = enhancedPersonalityInterning.SmartIntern(player.Personality)
	}
	if player.MediaHandling != "" {
		player.MediaHandling = enhancedPersonalityInterning.SmartIntern(player.MediaHandling)
	}

	// Use enhanced interning for club names (can be long and have repeated patterns)
	player.Club = enhancedClubInterning.SmartIntern(player.Club)

	// Intern position-related slices
	for i, pos := range player.ParsedPositions {
		player.ParsedPositions[i] = safePositionInterning.SafeIntern(pos)
	}
	for i, pos := range player.ShortPositions {
		player.ShortPositions[i] = safePositionInterning.SafeIntern(pos)
	}
	for i, group := range player.PositionGroups {
		player.PositionGroups[i] = safePositionInterning.SafeIntern(group)
	}

	player.BestRoleOverall = safePositionInterning.SafeIntern(player.BestRoleOverall)
}

// GetEnhancedStringInterningStats returns statistics for enhanced interning pools
func GetEnhancedStringInterningStats() map[string]map[string]int64 {
	stats := GetSafeStringInterningStats() // Get existing stats

	// Add enhanced stats
	enhancedClubInterning.mutex.RLock()
	stats["enhanced_clubs"] = map[string]int64{
		"unique_strings": int64(len(enhancedClubInterning.strings)),
		"total_requests": enhancedClubInterning.stats.totalRequests,
		"cache_hits":     enhancedClubInterning.stats.cacheHits,
		"memory_saved":   enhancedClubInterning.stats.memSaved,
		"compressions":   enhancedClubInterning.stats.compressions,
	}
	enhancedClubInterning.mutex.RUnlock()

	enhancedPersonalityInterning.mutex.RLock()
	stats["enhanced_personality"] = map[string]int64{
		"unique_strings": int64(len(enhancedPersonalityInterning.strings)),
		"total_requests": enhancedPersonalityInterning.stats.totalRequests,
		"cache_hits":     enhancedPersonalityInterning.stats.cacheHits,
		"memory_saved":   enhancedPersonalityInterning.stats.memSaved,
		"compressions":   enhancedPersonalityInterning.stats.compressions,
	}
	enhancedPersonalityInterning.mutex.RUnlock()

	return stats
}
