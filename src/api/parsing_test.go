package main

import (
	"testing"
)

func TestParseMonetaryValueGo(t *testing.T) {
	tests := []struct {
		name            string
		rawValue        string
		expectedDisplay string
		expectedNumeric int64
		expectedSymbol  string
	}{
		{
			name:            "simple dollar amount with M",
			rawValue:        "$1.5M",
			expectedDisplay: "$1.5M",
			expectedNumeric: 1500000,
			expectedSymbol:  "$",
		},
		{
			name:            "pound amount with K",
			rawValue:        "£25K",
			expectedDisplay: "£25K",
			expectedNumeric: 25000,
			expectedSymbol:  "£",
		},
		{
			name:            "euro amount",
			rawValue:        "€500K",
			expectedDisplay: "€500K",
			expectedNumeric: 500000,
			expectedSymbol:  "€",
		},
		{
			name:            "wage with per week",
			rawValue:        "£25K p/w",
			expectedDisplay: "£25K p/w",
			expectedNumeric: 25000,
			expectedSymbol:  "£",
		},
		{
			name:            "wage with alternative per week",
			rawValue:        "$50K /w",
			expectedDisplay: "$50K /w",
			expectedNumeric: 50000,
			expectedSymbol:  "$",
		},
		{
			name:            "range value - takes higher",
			rawValue:        "£10M - £15M",
			expectedDisplay: "£10M - £15M",
			expectedNumeric: 15000000,
			expectedSymbol:  "£",
		},
		{
			name:            "brazilian real",
			rawValue:        "R$2.5M",
			expectedDisplay: "R$2.5M",
			expectedNumeric: 2500000,
			expectedSymbol:  "R$",
		},
		{
			name:            "swiss franc",
			rawValue:        "CHF1.2M",
			expectedDisplay: "CHF1.2M",
			expectedNumeric: 1200000,
			expectedSymbol:  "CHF",
		},
		{
			name:            "australian dollar",
			rawValue:        "A$800K",
			expectedDisplay: "A$800K",
			expectedNumeric: 800000,
			expectedSymbol:  "A$",
		},
		{
			name:            "canadian dollar",
			rawValue:        "CA$1M",
			expectedDisplay: "CA$1M",
			expectedNumeric: 1000000,
			expectedSymbol:  "CA$",
		},
		{
			name:            "mexican peso",
			rawValue:        "Mex$5M",
			expectedDisplay: "Mex$5M",
			expectedNumeric: 5000000,
			expectedSymbol:  "Mex$",
		},
		{
			name:            "krona",
			rawValue:        "kr500K",
			expectedDisplay: "kr500K",
			expectedNumeric: 500000,
			expectedSymbol:  "kr",
		},
		{
			name:            "polish zloty",
			rawValue:        "zł2M",
			expectedDisplay: "zł2M",
			expectedNumeric: 2000000,
			expectedSymbol:  "zł",
		},
		{
			name:            "south african rand",
			rawValue:        "R1.5M",
			expectedDisplay: "R1.5M",
			expectedNumeric: 1500000,
			expectedSymbol:  "R",
		},
		{
			name:            "yen",
			rawValue:        "¥100M",
			expectedDisplay: "¥100M",
			expectedNumeric: 100000000,
			expectedSymbol:  "¥",
		},
		{
			name:            "rupee",
			rawValue:        "₹50M",
			expectedDisplay: "₹50M",
			expectedNumeric: 50000000,
			expectedSymbol:  "₹",
		},
		{
			name:            "ruble",
			rawValue:        "₽25M",
			expectedDisplay: "₽25M",
			expectedNumeric: 25000000,
			expectedSymbol:  "₽",
		},
		{
			name:            "turkish lira",
			rawValue:        "₺10M",
			expectedDisplay: "₺10M",
			expectedNumeric: 10000000,
			expectedSymbol:  "₺",
		},
		{
			name:            "korean won",
			rawValue:        "₩1000M",
			expectedDisplay: "₩1000M",
			expectedNumeric: 1000000000,
			expectedSymbol:  "₩",
		},
		{
			name:            "amount with commas",
			rawValue:        "$1,500,000",
			expectedDisplay: "$1,500,000",
			expectedNumeric: 1500000,
			expectedSymbol:  "$",
		},
		{
			name:            "decimal amount",
			rawValue:        "£2.75M",
			expectedDisplay: "£2.75M",
			expectedNumeric: 2750000,
			expectedSymbol:  "£",
		},
		{
			name:            "no currency symbol - defaults to dollar",
			rawValue:        "1.5M",
			expectedDisplay: "1.5M",
			expectedNumeric: 1500000,
			expectedSymbol:  "$",
		},
		{
			name:            "empty string",
			rawValue:        "",
			expectedDisplay: "",
			expectedNumeric: 0,
			expectedSymbol:  "",
		},
		{
			name:            "invalid format",
			rawValue:        "invalid",
			expectedDisplay: "invalid",
			expectedNumeric: 0,
			expectedSymbol:  "",
		},
		{
			name:            "zero value",
			rawValue:        "$0",
			expectedDisplay: "$0",
			expectedNumeric: 0,
			expectedSymbol:  "$",
		},
		{
			name:            "amount with spaces",
			rawValue:        " £ 25 K ",
			expectedDisplay: "£ 25 K",
			expectedNumeric: 25000,
			expectedSymbol:  "£",
		},
		{
			name:            "very large amount",
			rawValue:        "$999M",
			expectedDisplay: "$999M",
			expectedNumeric: 999000000,
			expectedSymbol:  "$",
		},
		{
			name:            "small K amount",
			rawValue:        "€1K",
			expectedDisplay: "€1K",
			expectedNumeric: 1000,
			expectedSymbol:  "€",
		},
		{
			name:            "fractional K amount",
			rawValue:        "$2.5K",
			expectedDisplay: "$2.5K",
			expectedNumeric: 2500,
			expectedSymbol:  "$",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			display, numeric, symbol := ParseMonetaryValueGo(tt.rawValue)

			if display != tt.expectedDisplay {
				t.Errorf("ParseMonetaryValueGo(%s) display = %s; want %s", tt.rawValue, display, tt.expectedDisplay)
			}

			if numeric != tt.expectedNumeric {
				t.Errorf("ParseMonetaryValueGo(%s) numeric = %d; want %d", tt.rawValue, numeric, tt.expectedNumeric)
			}

			if symbol != tt.expectedSymbol {
				t.Errorf("ParseMonetaryValueGo(%s) symbol = %s; want %s", tt.rawValue, symbol, tt.expectedSymbol)
			}
		})
	}
}

func TestParseMonetaryValueGoEdgeCases(t *testing.T) {
	// Test edge cases and boundary conditions
	tests := []struct {
		name        string
		rawValue    string
		description string
	}{
		{
			name:        "multiple currency symbols",
			rawValue:    "$€100M",
			description: "Should handle multiple symbols gracefully",
		},
		{
			name:        "currency symbol at end",
			rawValue:    "100M$",
			description: "Should detect symbol even if at end",
		},
		{
			name:        "mixed case suffixes",
			rawValue:    "$100m",
			description: "Should handle lowercase m suffix",
		},
		{
			name:        "mixed case suffixes K",
			rawValue:    "£50k",
			description: "Should handle lowercase k suffix",
		},
		{
			name:        "no multiplier",
			rawValue:    "$1000",
			description: "Should handle amounts without K/M suffix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			display, numeric, symbol := ParseMonetaryValueGo(tt.rawValue)

			// Basic validation - should not panic and should return reasonable values
			if display == "" && tt.rawValue != "" {
				t.Errorf("ParseMonetaryValueGo(%s) returned empty display for non-empty input", tt.rawValue)
			}

			if numeric < 0 {
				t.Errorf("ParseMonetaryValueGo(%s) returned negative numeric value: %d", tt.rawValue, numeric)
			}

			// Log the results for manual inspection
			t.Logf("ParseMonetaryValueGo(%s) = display:%s, numeric:%d, symbol:%s",
				tt.rawValue, display, numeric, symbol)
		})
	}
}

// Benchmark the monetary parsing function
func BenchmarkParseMonetaryValueGo(b *testing.B) {
	testValues := []string{
		"$1.5M",
		"£25K p/w",
		"€500K",
		"R$2.5M",
		"CHF1.2M",
		"¥100M",
		"$1,500,000",
		"£10M - £15M",
	}

	for i := 0; i < b.N; i++ {
		for _, value := range testValues {
			ParseMonetaryValueGo(value)
		}
	}
}
