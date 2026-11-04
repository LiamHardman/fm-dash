package main

import (
	"bufio"
	"context"
	"os"
	"strings"
	"testing"
)

func TestSniffCSVDelimiter(t *testing.T) {
	cases := map[string]rune{
		"Name;Position;Age\nA;B;1\n":                    ';',
		"Inf,Player,Nation\n,A,ENG\n":                   ',',
		"\"Club, The\";Name;Age\n":                      ';', // quoted comma ignored
		"Name;Position\n\"x,y,z,w\";1\n":                ';', // only the header line counts
		"Name\n":                                        ';', // no delimiter: legacy default
		"Name,Position;Age\n":                           ';', // tie: legacy default
		"Player,\"Position\",Age,\"Club; FC\",Wage\n":   ',',
		"Player,Club,Age\r\nA,\"Lyon, FC\",20\r\n":      ',',
		"Player;Club;Age;Wage;Value\nA,B,C,D,E,F,G,H\n": ';',
	}
	for input, want := range cases {
		if got := sniffCSVDelimiter(bufio.NewReader(strings.NewReader(input))); got != want {
			t.Errorf("sniffCSVDelimiter(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestSplitPositionAndRole(t *testing.T) {
	cases := []struct{ in, position, role string }{
		{"D (L) Full-Back", "D (L)", "Full-Back"},
		{"GK Goalkeeper", "GK", "Goalkeeper"},
		{"DM Defensive Midfielder", "DM", "Defensive Midfielder"},
		{"DM, M (C) Central Midfielder", "DM, M (C)", "Central Midfielder"},
		{"M (C), AM (RLC), ST (C) Inside Forward", "M (C), AM (RLC), ST (C)", "Inside Forward"},
		{"D/WB (R), DM, M/AM (R) Inside Wing-Back", "D/WB (R), DM, M/AM (R)", "Inside Wing-Back"},
		{"D (C) No-Nonsense Centre-Back", "D (C)", "No-Nonsense Centre-Back"},
		// Classic FM exports carry no role and must pass through untouched.
		{"D (RC), DM", "D (RC), DM", ""},
		{"GK", "GK", ""},
		{"ST (C)", "ST (C)", ""},
		{"", "", ""},
		{"Unrecognised", "Unrecognised", ""},
	}
	for _, c := range cases {
		position, role := SplitPositionAndRole(c.in)
		if position != c.position || role != c.role {
			t.Errorf("SplitPositionAndRole(%q) = (%q, %q), want (%q, %q)", c.in, position, role, c.position, c.role)
		}
	}
}

func parseFM27File(t *testing.T, path string) *ParsedPlayers {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	parsed, err := ParsePlayers(context.Background(), file, ParseOptions{Format: "csv", Workers: 4})
	if err != nil {
		t.Fatalf("ParsePlayers: %v", err)
	}
	return parsed
}

func TestFM27ModCSVParses(t *testing.T) {
	parsed := parseFM27File(t, "testdata/fm27/mod_export.csv")

	if missing := missingRequiredExportHeaders(parsed.Headers); len(missing) != 0 {
		t.Fatalf("FM27 mod export missing required headers after normalization: %v", missing)
	}
	if parsed.Counts.Errors != 0 || parsed.Counts.Accepted != 5 {
		t.Fatalf("counts = %+v, diagnostics = %+v", parsed.Counts, parsed.Diagnostics)
	}
	if parsed.Currency != "£" {
		t.Errorf("currency = %q, want £", parsed.Currency)
	}

	byName := map[string]*Player{}
	for i := range parsed.Players {
		byName[parsed.Players[i].Name] = &parsed.Players[i]
	}

	fuentes := byName["Joseph Fuentes"]
	if fuentes == nil {
		t.Fatal("Joseph Fuentes not parsed")
	}
	if fuentes.Position != "D (L)" || fuentes.Attributes["Role"] != "Full-Back" {
		t.Errorf("position/role = %q/%q", fuentes.Position, fuentes.Attributes["Role"])
	}
	if fuentes.UID != 2002198107 || fuentes.Club != "CD Universitario (Panama)" || fuentes.BasedIn != "Panama" {
		t.Errorf("identity fields wrong: uid=%d club=%q basedIn=%q", fuentes.UID, fuentes.Club, fuentes.BasedIn)
	}
	if fuentes.NationalityFIFACode != "PAN" {
		t.Errorf("nationality code = %q, want PAN", fuentes.NationalityFIFACode)
	}
	// Natural Fitness is the second "Nat" column and must not clobber nationality.
	if fuentes.NumericAttributes["Nat"] != 5 || fuentes.NumericAttributes["Ant"] != 11 || fuentes.NumericAttributes["Tec"] != 13 {
		t.Errorf("attributes: Nat=%d Ant=%d Tec=%d", fuentes.NumericAttributes["Nat"], fuentes.NumericAttributes["Ant"], fuentes.NumericAttributes["Tec"])
	}
	if fuentes.WageAmount != 55 || fuentes.TransferValueAmount != 26000 {
		t.Errorf("money: wage=%d value=%d", fuentes.WageAmount, fuentes.TransferValueAmount)
	}
	if fuentes.PerformanceStatsNumeric["Av Rat"] != 6.76 || fuentes.PerformanceStatsNumeric["Mins"] != 3960 || fuentes.PerformanceStatsNumeric["K Ps/90"] != 0.77 {
		t.Errorf("performance stats: %+v", fuentes.PerformanceStatsNumeric)
	}
	if len(fuentes.ShortPositions) == 0 {
		t.Error("expected parsed short positions for D (L)")
	}

	keeper := byName["Prince Chisanga"]
	if keeper == nil || keeper.Position != "GK" {
		t.Fatalf("goalkeeper not parsed with stripped role: %+v", keeper)
	}

	multi := byName["Lindell Sween"]
	if multi == nil || multi.Position != "M (C), AM (RLC), ST (C)" || multi.Attributes["Role"] != "Inside Forward" {
		t.Fatalf("quoted multi-position cell not parsed: %+v", multi)
	}
}

// TestFM27ModCSVFullExport exercises the real export when it is present in the
// repository root (it is not committed).
func TestFM27ModCSVFullExport(t *testing.T) {
	const path = "../../fm27.csv"
	if _, err := os.Stat(path); err != nil {
		t.Skip("fm27.csv not present")
	}
	parsed := parseFM27File(t, path)
	if parsed.Counts.Errors != 0 || parsed.Counts.Rejected != 0 {
		t.Fatalf("counts = %+v, first diagnostics = %+v", parsed.Counts, parsed.Diagnostics[:min(5, len(parsed.Diagnostics))])
	}
	for i := range parsed.Players {
		p := &parsed.Players[i]
		if len(p.ShortPositions) == 0 {
			t.Errorf("%s: position %q produced no short positions", p.Name, p.Position)
		}
	}
}
