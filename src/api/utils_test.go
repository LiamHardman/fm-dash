package main

import (
	"reflect"
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
			name:     "normal case - get first 3 elements",
			slice:    []string{"a", "b", "c", "d", "e"},
			n:        3,
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "n equals slice length",
			slice:    []string{"a", "b", "c"},
			n:        3,
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "n greater than slice length",
			slice:    []string{"a", "b"},
			n:        5,
			expected: []string{"a", "b"},
		},
		{
			name:     "n is zero",
			slice:    []string{"a", "b", "c"},
			n:        0,
			expected: []string{},
		},
		{
			name:     "n is negative",
			slice:    []string{"a", "b", "c"},
			n:        -1,
			expected: []string{},
		},
		{
			name:     "empty slice",
			slice:    []string{},
			n:        3,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFirstNCells(tt.slice, tt.n)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("GetFirstNCells(%v, %d) = %v; want %v", tt.slice, tt.n, result, tt.expected)
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
			name:     "zero bytes",
			bytes:    0,
			expected: 0.0,
		},
		{
			name:     "1 MB in bytes",
			bytes:    1024 * 1024,
			expected: 1.0,
		},
		{
			name:     "2.5 MB in bytes",
			bytes:    2621440, // 2.5 * 1024 * 1024
			expected: 2.5,
		},
		{
			name:     "fractional MB",
			bytes:    512 * 1024, // 0.5 MB
			expected: 0.5,
		},
		{
			name:     "large value",
			bytes:    1073741824, // 1024 MB = 1 GB
			expected: 1024.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BToMb(tt.bytes)
			if result != tt.expected {
				t.Errorf("BToMb(%d) = %f; want %f", tt.bytes, result, tt.expected)
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
			name: "normal case with multiple keys",
			input: map[string]map[string]float64{
				"c": {"value": 1.0},
				"a": {"value": 2.0},
				"b": {"value": 3.0},
			},
			expected: []string{"a", "b", "c"}, // sorted
		},
		{
			name:     "empty map",
			input:    map[string]map[string]float64{},
			expected: []string{},
		},
		{
			name: "single key",
			input: map[string]map[string]float64{
				"single": {"value": 1.0},
			},
			expected: []string{"single"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMapKeys(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("GetMapKeys(%v) = %v; want %v", tt.input, result, tt.expected)
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
			name: "normal case with multiple keys",
			input: map[string]float64{
				"c": 1.0,
				"a": 2.0,
				"b": 3.0,
			},
			expected: []string{"a", "b", "c"}, // sorted
		},
		{
			name:     "empty map",
			input:    map[string]float64{},
			expected: []string{},
		},
		{
			name: "single key",
			input: map[string]float64{
				"single": 1.0,
			},
			expected: []string{"single"},
		},
		{
			name: "keys with numbers and special chars",
			input: map[string]float64{
				"key_3": 1.0,
				"key_1": 2.0,
				"key_2": 3.0,
			},
			expected: []string{"key_1", "key_2", "key_3"}, // sorted
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMapKeysStringFloat(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("GetMapKeysStringFloat(%v) = %v; want %v", tt.input, result, tt.expected)
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
			name:     "value within range",
			value:    5,
			minVal:   1,
			maxVal:   10,
			expected: 5,
		},
		{
			name:     "value below minimum",
			value:    -5,
			minVal:   1,
			maxVal:   10,
			expected: 1,
		},
		{
			name:     "value above maximum",
			value:    15,
			minVal:   1,
			maxVal:   10,
			expected: 10,
		},
		{
			name:     "value equals minimum",
			value:    1,
			minVal:   1,
			maxVal:   10,
			expected: 1,
		},
		{
			name:     "value equals maximum",
			value:    10,
			minVal:   1,
			maxVal:   10,
			expected: 10,
		},
		{
			name:     "zero values",
			value:    0,
			minVal:   0,
			maxVal:   0,
			expected: 0,
		},
		{
			name:     "negative range",
			value:    -5,
			minVal:   -10,
			maxVal:   -1,
			expected: -5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Clamp(tt.value, tt.minVal, tt.maxVal)
			if result != tt.expected {
				t.Errorf("Clamp(%d, %d, %d) = %d; want %d", tt.value, tt.minVal, tt.maxVal, result, tt.expected)
			}
		})
	}
}
