package main

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestParserSourceFidelityStatistics(t *testing.T) {
	columns := []struct {
		source, canonical, value string
		numeric                  float64
	}{
		{"Pass Completion Percentage", "Pas %", "94%", 94},
		{"Save Percentage", "Sv %", "81%", 81},
		{"Passes Attempted per 90", "Ps A/90", "58.3", 58.3},
		{"Tackle Completion Percentage", "Tck R", "78", 78},
		{"Tackles Completed per 90", "Tck/90", "0.6", 0.6},
		{"Crosses Completed Ratio", "Cr C/A", "36", 36},
		{"Chances Created per 90", "Ch C/90", "0.5", 0.5},
		{"Crosses Completed per 90", "Cr C/90", "0.3", 0.3},
		{"Minutes", "Mins", "4,665", 4665},
		{"Headers Won per 90", "Hdrs W/90", "3.2", 3.2},
		{"Crosses Attempted per 90", "CRS A/90", "0.8", 0.8},
		{"Dribbles per 90", "Drb/90", "1.4", 1.4},
		{"Key Passes per 90", "K Ps/90", "2.2", 2.2},
		{"Key Tackles per 90", "K Tck/90", "0.2", 0.2},
		{"Clearances per 90", "Clr/90", "1.0", 1},
		{"Possession Won per 90", "Poss Won/90", "14.2", 14.2},
		{"Interceptions per 90", "Int/90", "2.5", 2.5},
		{"Shots on Target per 90", "ShT/90", "1.5", 1.5},
		{"Possession Lost per 90", "Poss Lost/90", "15.7", 15.7},
		{"Progressive Passes per 90", "Pr passes/90", "3.4", 3.4},
		{"Passes Completed per 90", "Ps C/90", "44.6", 44.6},
		{"Saves per 90", "Saves/90", "3.1", 3.1},
		{"Sprints/90", "Sprints/90", "2.14", 2.14},
		{"K Hdrs/90", "K Hdrs/90", "0.09", 0.09},
		{"Aer A/90", "Aer A/90", "0.64", 0.64},
	}
	for _, format := range []string{"csv", "html"} {
		for _, workers := range []int{1, 4} {
			t.Run(fmt.Sprintf("%s/workers=%d", format, workers), func(t *testing.T) {
				headers := []string{"UID", "Name", "Position", "Age", "Wage"}
				values := []string{"1", "Example", "M (C)", "24", "£1K p/w"}
				for _, c := range columns {
					header := c.source
					if format == "html" {
						header = c.canonical
					}
					headers = append(headers, header)
					values = append(values, c.value)
				}
				input := strings.Join(headers, ";") + "\n" + strings.Join(values, ";") + "\n"
				if format == "html" {
					input = "<table><tr><th>" + strings.Join(headers, "</th><th>") + "</th></tr><tr><td>" + strings.Join(values, "</td><td>") + "</td></tr></table>"
				}
				p, err := ParsePlayers(context.Background(), strings.NewReader(input), ParseOptions{format, "complete", workers, true})
				if err != nil {
					t.Fatal(err)
				}
				player := &p.Players[0]
				protoPlayer, err := player.ToProto(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				for _, c := range columns {
					if got, ok := player.PerformanceStatsNumeric[c.canonical]; !ok || got != c.numeric {
						t.Errorf("%s: expected numeric %v, got %v (present=%v)", c.source, c.numeric, got, ok)
					}
					if got := p.RawAttributes[0][c.canonical]; got != c.value {
						t.Errorf("raw %s: %q != %q", c.source, got, c.value)
					}
					if got, ok := player.PerformancePercentiles["Global"][c.canonical]; !ok || got != 50 {
						t.Errorf("percentile %s: %v (present=%v)", c.source, got, ok)
					}
					if got := protoPlayer.EssentialAttributes[c.canonical]; got != c.value {
						t.Errorf("protobuf raw %s: %q != %q", c.source, got, c.value)
					}
					if got, ok := protoPlayer.PerformanceStatsNumeric[c.canonical]; !ok || got != c.numeric {
						t.Errorf("protobuf numeric %s: %v (present=%v)", c.source, got, ok)
					}
				}
			})
		}
	}
}

func TestParserPreservesPlayerMetadataAtEveryStage(t *testing.T) {
	input := "UID;Name;Position;Age;Club;Division;Based In\n1;First;M (C);24;Example FC;League A;Country A\n2;Second;M (C);24;Example FC;League B;Country B\n"
	for _, stage := range []string{"parsed", "enhanced", "complete"} {
		p, err := ParsePlayers(context.Background(), strings.NewReader(input), ParseOptions{"csv", stage, 2, true})
		if err != nil {
			t.Fatal(err)
		}
		if p.Players[1].Division != "League B" || p.Players[1].BasedIn != "Country B" {
			t.Errorf("%s rewrites source metadata: %+v", stage, p.Players[1])
		}
	}
}

func TestParserPreservesDistinctRatings(t *testing.T) {
	for _, input := range []string{
		"UID;Name;Position;Age;Rating;Average Rating Club\n1;Example;M (C);24;7.10;6.20\n",
		"UID;Name;Position;Age;Average Rating Club;Rating\n1;Example;M (C);24;6.20;7.10\n",
	} {
		p, err := ParsePlayers(context.Background(), strings.NewReader(input), ParseOptions{"csv", "complete", 2, true})
		if err != nil {
			t.Fatal(err)
		}
		for key, want := range map[string]float64{"Av Rat": 7.1, "Average Rating Club": 6.2} {
			if got, ok := p.Players[0].PerformanceStatsNumeric[key]; !ok || got != want {
				t.Errorf("%s: %v (present=%v), want %v", key, got, ok, want)
			}
		}
		if p.RawAttributes[0]["Av Rat"] != "7.10" || p.RawAttributes[0]["Average Rating Club"] != "6.20" {
			t.Errorf("ratings lost: %+v", p.RawAttributes[0])
		}
	}
}

func TestMonetaryValuesRoundToWholeCurrencyUnits(t *testing.T) {
	for input, want := range map[string]int64{"£4.1M": 4100000, "£8.2M": 8200000, "£1M - £4.1M": 4100000, "£4.1K p/w": 4100, "£1.49": 1, "£1.50": 2} {
		_, got, _ := ParseMonetaryValueGo(input)
		if got != want {
			t.Errorf("%s: %d != %d", input, got, want)
		}
	}
}

func TestTeamAggregationPreservesSourceMetadata(t *testing.T) {
	players := []Player{
		{Name: "Loan", Club: "Example", Division: "Other", BasedIn: "Elsewhere"},
		{Name: "First", Club: "Example", Division: "Home", BasedIn: "Here"},
		{Name: "Second", Club: "Example", Division: "Home", BasedIn: "Here"},
	}
	before := make([]Player, len(players))
	copy(before, players)
	teams := processTeamsData(players, "Home", "Here")
	if len(teams) != 1 || teams[0].PlayerCount != 3 {
		t.Fatalf("club aggregation split loan rows: %+v", teams)
	}
	if teams[0].Players[0].Division != "Other" {
		t.Error("team player response lost source Division")
	}
	leagues := CreateConcurrentLeagueProcessor(1).ProcessLeaguesAsync(context.Background(), players)
	if len(leagues) != 1 || leagues[0].Name != "Home" || leagues[0].PlayerCount != 3 {
		t.Errorf("league aggregation split loan rows: %+v", leagues)
	}
	if !reflect.DeepEqual(before, players) {
		t.Error("aggregation mutated source players")
	}
}

func TestStoragePreservesPlayerSourceMetadata(t *testing.T) {
	originalStorage := storage
	storage = CreateMockStorage()
	t.Cleanup(func() { storage = originalStorage })
	input := "UID;Name;Position;Age;Club;Division;Based In\n1;First;M (C);24;Example FC;League A;Country A\n2;Second;M (C);24;Example FC;League B;Country B\n"
	p, err := ParsePlayers(context.Background(), strings.NewReader(input), ParseOptions{"csv", "complete", 1, true})
	if err != nil {
		t.Fatal(err)
	}
	const datasetID = "source-metadata-regression"
	t.Cleanup(func() {
		storeMutex.Lock()
		delete(playerDataStore, datasetID)
		storeMutex.Unlock()
	})
	SetPlayerData(datasetID, p.Players, "£")
	stored, _, err := RetrieveDataset(datasetID)
	if err != nil {
		t.Fatal(err)
	}
	served, _, ok := GetPlayerData(datasetID)
	if !ok {
		t.Fatal("stored Dataset is unavailable")
	}
	for _, players := range [][]Player{stored, served} {
		if players[1].Division != "League B" || players[1].BasedIn != "Country B" {
			t.Error("storage rewrote source metadata")
		}
	}
}
