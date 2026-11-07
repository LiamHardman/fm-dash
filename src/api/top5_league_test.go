package main

import "testing"

func TestIsTop5League(t *testing.T) {
	cases := []struct {
		division, basedIn string
		want              bool
	}{
		// FM27's own exports (and save imports) print Spain's top flight as "LaLiga".
		{"LaLiga", "Spain", true},
		{"LaLiga", "", true},
		{"La Liga", "Spain", true},
		{"Primera División", "", true},
		{"Premier League", "England", true},
		{"Premier League", "Spain", false},
		{"Ligue 1 McDonald's", "France", true},
		{"Série A", "Brazil", false},
		{"LaLiga Hypermotion", "Spain", false},
	}
	for _, c := range cases {
		if got := IsTop5League(c.division, c.basedIn); got != c.want {
			t.Errorf("IsTop5League(%q, %q) = %v, want %v", c.division, c.basedIn, got, c.want)
		}
	}
}
