package main

import "testing"

func TestGetMentalityModifierFM27Spellings(t *testing.T) {
	cases := map[string]int{
		"Resilient":          15, // FM27 spelling
		"Resillient":         15, // older FM-Dash spelling, still accepted
		"Low Self Belief":    -35,
		"Low Self-Belief":    -35,
		"Natural Leader":     15,
		"Charismatic Leader": 15,
		"Model Citizen":      20,
		"Balanced":           8,
		"Not A Personality":  0,
	}
	for label, want := range cases {
		if got := getMentalityModifier(label); got != want {
			t.Errorf("getMentalityModifier(%q) = %d, want %d", label, got, want)
		}
	}
}
