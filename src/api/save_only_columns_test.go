package main

import (
	"context"
	"testing"
)

// A minimal FM-style row: identity, a few attributes, then the save-only columns.
var saveOnlyBaseHeaders = []string{
	"Name", "Position", "Age", "Club",
	"Cor", "Cro", "Dri", "Fin", "Fir", "Pas", "Tec", "Ant", "Dec", "Det", "OtB", "Pos",
	"Acc", "Pac", "Sta", "Str", "Wor", "Tea",
}
var saveOnlyBaseCells = []string{
	"Bruno Costa", "M (C)", "26", "Arsenal",
	"12", "13", "14", "11", "15", "16", "15", "14", "15", "16", "13", "12",
	"14", "13", "15", "12", "16", "15",
}

func saveOnlyRow(t *testing.T, extraHeaders, extraCells []string) Player {
	t.Helper()
	headers := append(append([]string{}, saveOnlyBaseHeaders...), extraHeaders...)
	cells := append(append([]string{}, saveOnlyBaseCells...), extraCells...)
	player, err := parseCellsToPlayer(cells, headers)
	if err != nil {
		t.Fatal(err)
	}
	EnhancePlayerWithCalculations(&player)
	return player
}

var allSaveOnlyHeaders = append([]string{"CA", "PA"}, append(append([]string{}, HiddenAttributeHeaders...), "Inf")...)

func allSaveOnlyCells() []string {
	cells := []string{"152", "171"}
	for i := range HiddenAttributeHeaders {
		cells = append(cells, []string{"3", "17", "12", "5", "9", "14", "18", "4", "11", "15", "16", "13", "7"}[i])
	}
	return append(cells, "Inj, Loa")
}

func TestSaveOnlyColumnsFillNewFields(t *testing.T) {
	player := saveOnlyRow(t, allSaveOnlyHeaders, allSaveOnlyCells())
	if player.SaveCA != 152 || player.PA != 171 {
		t.Errorf("SaveCA/PA = %d/%d, want 152/171", player.SaveCA, player.PA)
	}
	if player.Inf != "Inj, Loa" {
		t.Errorf("Inf = %q", player.Inf)
	}
	if len(player.HiddenAttributes) != 13 || player.HiddenAttributes["Consistency"] != 3 ||
		player.HiddenAttributes["Temperament"] != 7 {
		t.Errorf("HiddenAttributes = %v", player.HiddenAttributes)
	}
	for _, key := range append([]string{"CA", "PA", "Inf"}, HiddenAttributeHeaders...) {
		if _, ok := player.NumericAttributes[key]; ok {
			t.Errorf("%s leaked into NumericAttributes", key)
		}
		if _, ok := player.Attributes[key]; ok {
			t.Errorf("%s leaked into Attributes", key)
		}
	}
}

func TestSaveOnlyColumnsAbsentFromOrdinaryImports(t *testing.T) {
	player := saveOnlyRow(t, nil, nil)
	if player.SaveCA != 0 || player.PA != 0 || player.Inf != "" || player.HiddenAttributes != nil {
		t.Errorf("ordinary import got save-only values: %d %d %q %v", player.SaveCA, player.PA, player.Inf, player.HiddenAttributes)
	}
}

func TestSaveOnlyColumnsDropOutOfRangeValues(t *testing.T) {
	player := saveOnlyRow(t,
		[]string{"CA", "PA", "Consistency", "Dirtiness", "Ambition", "Inf"},
		[]string{"201", "-5", "21", "0", "x", "  Wnt  "})
	if player.SaveCA != 0 || player.PA != 0 {
		t.Errorf("SaveCA/PA = %d/%d, want 0/0", player.SaveCA, player.PA)
	}
	if len(player.HiddenAttributes) != 0 {
		t.Errorf("HiddenAttributes = %v, want none", player.HiddenAttributes)
	}
	if player.Inf != "Wnt" {
		t.Errorf("Inf = %q, want trimmed Wnt", player.Inf)
	}
}

func TestHiddenAttributesDoNotChangeRatings(t *testing.T) {
	without := saveOnlyRow(t, nil, nil)
	hiddenOnly := saveOnlyRow(t, HiddenAttributeHeaders, allSaveOnlyCells()[2:15])
	if len(hiddenOnly.HiddenAttributes) != 13 {
		t.Fatalf("hidden attributes not parsed: %v", hiddenOnly.HiddenAttributes)
	}
	if without.Overall != hiddenOnly.Overall || without.MBR != hiddenOnly.MBR || without.CA != hiddenOnly.CA {
		t.Errorf("ratings changed: overall %d/%d, MBR %d/%d, CA %d/%d",
			without.Overall, hiddenOnly.Overall, without.MBR, hiddenOnly.MBR, without.CA, hiddenOnly.CA)
	}
	if len(without.RoleSpecificOveralls) != len(hiddenOnly.RoleSpecificOveralls) {
		t.Fatal("role overall count changed")
	}
	for i := range without.RoleSpecificOveralls {
		if without.RoleSpecificOveralls[i] != hiddenOnly.RoleSpecificOveralls[i] {
			t.Errorf("role %v changed to %v", without.RoleSpecificOveralls[i], hiddenOnly.RoleSpecificOveralls[i])
		}
	}
}

func TestSaveOnlyFieldsProtoRoundTrip(t *testing.T) {
	player := saveOnlyRow(t, allSaveOnlyHeaders, allSaveOnlyCells())
	protoPlayer, err := player.ToProto(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	back, err := PlayerFromProto(context.Background(), protoPlayer)
	if err != nil {
		t.Fatal(err)
	}
	if back.SaveCA != 152 || back.PA != 171 || back.Inf != "Inj, Loa" || len(back.HiddenAttributes) != 13 ||
		back.HiddenAttributes["Ambition"] != 18 {
		t.Errorf("round trip lost save-only fields: %d %d %q %v", back.SaveCA, back.PA, back.Inf, back.HiddenAttributes)
	}

	empty := saveOnlyRow(t, nil, nil)
	protoEmpty, _ := empty.ToProto(context.Background())
	backEmpty, _ := PlayerFromProto(context.Background(), protoEmpty)
	if backEmpty.SaveCA != 0 || backEmpty.HiddenAttributes != nil || backEmpty.Inf != "" {
		t.Errorf("an ordinary player gained save-only fields: %+v", backEmpty.HiddenAttributes)
	}
}

func TestSaveOnlyFieldsSurviveDeepCopies(t *testing.T) {
	player := saveOnlyRow(t, allSaveOnlyHeaders, allSaveOnlyCells())
	for name, copies := range map[string][]Player{
		"FastDeepCopyPlayers": FastDeepCopyPlayers([]Player{player}),
		"deepCopyPlayers":     deepCopyPlayers([]Player{player}),
	} {
		c := copies[0]
		if c.SaveCA != 152 || c.PA != 171 || c.Inf != "Inj, Loa" || len(c.HiddenAttributes) != 13 {
			t.Errorf("%s lost save-only fields", name)
			continue
		}
		c.HiddenAttributes["Consistency"] = 20
		if player.HiddenAttributes["Consistency"] != 3 {
			t.Errorf("%s shares the HiddenAttributes map with the original", name)
		}
	}
}
