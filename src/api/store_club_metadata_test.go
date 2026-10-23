package main

import "testing"

func TestStoredDatasetClubMetadata(t *testing.T) {
	originalStorage := storage
	originalCacheDisabled := disableInMemoryDatasetCache
	t.Cleanup(func() {
		storage = originalStorage
		disableInMemoryDatasetCache = originalCacheDisabled
	})

	for _, cacheDisabled := range []bool{false, true} {
		name := "cache_enabled"
		if cacheDisabled {
			name = "cache_disabled"
		}
		t.Run(name, func(t *testing.T) {
			disableInMemoryDatasetCache = cacheDisabled
			storage = CreateInMemoryStorage()
			datasetID := "club-metadata-" + name
			t.Cleanup(func() {
				storeMutex.Lock()
				delete(playerDataStore, datasetID)
				storeMutex.Unlock()
			})
			// Player metadata must survive retrieval, caching and lightweight readers.
			// Club aggregation separately derives the majority from the whole squad.
			players := []Player{
				{UID: 1, Name: "Adil Aouchiche", Club: "Schalke 04", Division: "Championnat Inferieures", BasedIn: "France"},
				{UID: 2, Name: "Squad player A", Club: "Schalke 04", Division: "2. Bundesliga", BasedIn: "Germany"},
				{UID: 3, Name: "Squad player B", Club: "Schalke 04", Division: "2. Bundesliga", BasedIn: "Germany"},
			}
			for i := range players {
				players[i].NumericAttributes = map[string]int{"Pac": 12}
				players[i].PerformanceStatsNumeric = map[string]float64{}
			}
			if err := storage.Store(datasetID, DatasetData{Players: players, CurrencySymbol: "GBP"}); err != nil {
				t.Fatal(err)
			}
			loaded, _, err := RetrieveDataset(datasetID)
			if err != nil {
				t.Fatal(err)
			}
			for _, player := range loaded {
				expected := players[player.UID-1]
				if player.Division != expected.Division || player.BasedIn != expected.BasedIn {
					t.Errorf("dataset retrieval: %s has division %q and country %q", player.Name, player.Division, player.BasedIn)
				}
			}
			for _, read := range []struct {
				name string
				get  func(string) ([]Player, string, bool)
			}{
				{"upgrade finder before player table", GetPlayerDataForUpgradeFinder},
				{"search indexing before player table", GetPlayerDataForIndexing},
				{"player table first load", GetPlayerData},
				{"player table repeat load", GetPlayerData},
				{"upgrade finder", GetPlayerDataForUpgradeFinder},
				{"search indexing", GetPlayerDataForIndexing},
			} {
				got, currency, ok := read.get(datasetID)
				if !ok || len(got) != len(players) || currency != "GBP" {
					t.Fatalf("%s: dataset not retrieved intact", read.name)
				}
				for _, player := range got {
					expected := players[player.UID-1]
					if player.Division != expected.Division {
						t.Errorf("%s: %s at %s has division %q, want %q", read.name, player.Name, player.Club, player.Division, expected.Division)
					}
					// The lightweight search index intentionally omits BasedIn.
					if player.BasedIn != "" && player.BasedIn != expected.BasedIn {
						t.Errorf("%s: %s has country %q, want %s", read.name, player.Name, player.BasedIn, expected.BasedIn)
					}
				}
			}
			teams := processTeamsData(loaded, "2. Bundesliga", "Germany")
			if len(teams) != 1 || teams[0].PlayerCount != 3 {
				t.Fatalf("club summary should group the whole squad: %+v", teams)
			}
			if players[0].Division != "Championnat Inferieures" || players[0].BasedIn != "France" {
				t.Error("loading mutated the storage-owned player slice")
			}
		})
	}
}
