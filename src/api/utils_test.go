package main

import (
	"testing"
)

func TestGetFirstNCells(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		n        int
		expected []string
	}{
		{
			name:     "Normal case - get first 3 elements",
			slice:    []string{"a", "b", "c", "d", "e"},
			n:        3,
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "N greater than slice length",
			slice:    []string{"a", "b"},
			n:        5,
			expected: []string{"a", "b"},
		},
		{
			name:     "N equals slice length",
			slice:    []string{"x", "y", "z"},
			n:        3,
			expected: []string{"x", "y", "z"},
		},
		{
			name:     "N is zero",
			slice:    []string{"a", "b", "c"},
			n:        0,
			expected: []string{},
		},
		{
			name:     "N is negative",
			slice:    []string{"a", "b", "c"},
			n:        -1,
			expected: []string{},
		},
		{
			name:     "Empty slice",
			slice:    []string{},
			n:        3,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFirstNCells(tt.slice, tt.n)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("Expected %v, got %v", tt.expected, result)
					break
				}
			}
		})
	}
}

func TestBToMb(t *testing.T) {
	tests := []struct {
		name     string
		bytes    uint64
		expected float64
	}{
		{
			name:     "Zero bytes",
			bytes:    0,
			expected: 0.0,
		},
		{
			name:     "1 MB in bytes",
			bytes:    1024 * 1024,
			expected: 1.0,
		},
		{
			name:     "Half MB",
			bytes:    512 * 1024,
			expected: 0.5,
		},
		{
			name:     "2.5 MB",
			bytes:    2560 * 1024,
			expected: 2.5,
		},
		{
			name:     "1 GB in bytes",
			bytes:    1024 * 1024 * 1024,
			expected: 1024.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BToMb(tt.bytes)
			if result != tt.expected {
				t.Errorf("Expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestGetMapKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]map[string]float64
		expected []string
	}{
		{
			name: "Normal map with multiple keys",
			input: map[string]map[string]float64{
				"global": {"stat1": 1.0, "stat2": 2.0},
				"league": {"stat1": 1.5, "stat2": 2.5},
				"alpha":  {"stat1": 0.5, "stat2": 1.5},
			},
			expected: []string{"alpha", "global", "league"}, // Should be sorted
		},
		{
			name:     "Empty map",
			input:    map[string]map[string]float64{},
			expected: []string{},
		},
		{
			name: "Single key map",
			input: map[string]map[string]float64{
				"only": {"stat": 1.0},
			},
			expected: []string{"only"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMapKeys(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("Expected %v, got %v", tt.expected, result)
					break
				}
			}
		})
	}
}

func TestGetMapKeysStringFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]float64
		expected []string
	}{
		{
			name: "Normal map with multiple keys",
			input: map[string]float64{
				"stat3": 3.0,
				"stat1": 1.0,
				"stat2": 2.0,
			},
			expected: []string{"stat1", "stat2", "stat3"}, // Should be sorted
		},
		{
			name:     "Empty map",
			input:    map[string]float64{},
			expected: []string{},
		},
		{
			name: "Single key map",
			input: map[string]float64{
				"only_stat": 1.0,
			},
			expected: []string{"only_stat"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMapKeysStringFloat(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("Expected %v, got %v", tt.expected, result)
					break
				}
			}
		})
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		minVal   int
		maxVal   int
		expected int
	}{
		{
			name:     "Value within range",
			value:    5,
			minVal:   1,
			maxVal:   10,
			expected: 5,
		},
		{
			name:     "Value below minimum",
			value:    -5,
			minVal:   0,
			maxVal:   10,
			expected: 0,
		},
		{
			name:     "Value above maximum",
			value:    15,
			minVal:   0,
			maxVal:   10,
			expected: 10,
		},
		{
			name:     "Value equals minimum",
			value:    0,
			minVal:   0,
			maxVal:   10,
			expected: 0,
		},
		{
			name:     "Value equals maximum",
			value:    10,
			minVal:   0,
			maxVal:   10,
			expected: 10,
		},
		{
			name:     "Negative range",
			value:    -15,
			minVal:   -10,
			maxVal:   -5,
			expected: -10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Clamp(tt.value, tt.minVal, tt.maxVal)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}
