package main

import "sort"

// GetFirstNCells returns the first N cells from a slice of strings.
// Useful for logging or error messages.
func GetFirstNCells(slice []string, n int) []string {
	if n < 0 {
		n = 0
	}
	if n > len(slice) {
		n = len(slice)
	}
	return slice[:n]
}

// BToMb converts bytes to megabytes.
func BToMb(b uint64) float64 {
	return float64(b) / 1024 / 1024
}

// GetMapKeys extracts and sorts keys from a map[string]map[string]float64.
// Useful for debugging or consistent logging of map keys.
// This specific signature was for player.PerformancePercentiles.
func GetMapKeys(m map[string]map[string]float64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys) // Sort for consistent order
	return keys
}

// GetMapKeysStringFloat provides a more generic way to get keys if the inner map value type is known.
// This version is for map[string]float64, as used in PerformancePercentiles["Global"]
func GetMapKeysStringFloat(m map[string]float64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Clamp ensures a value is within a min and max range.
func Clamp(value, minVal, maxVal int) int {
	if value < minVal {
		return minVal
	}
	if value > maxVal {
		return maxVal
	}
	return value
}
