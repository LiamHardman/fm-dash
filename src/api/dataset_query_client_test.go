package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPlayerToMaterializeRow(t *testing.T) {
	p := Player{
		UID:                 1234567890,
		Name:                "Test Player",
		Position:            "M (C)",
		Age:                 "24",
		Club:                "Test FC",
		Division:            "Test Division",
		BasedIn:             "England",
		TransferValue:       "£1M - £2M",
		TransferValueAmount: 1_500_000,
		Wage:                "£10K p/w",
		WageAmount:          10_000,
		Personality:         "Ambitious",
		MediaHandling:       "Evasive",
		Nationality:         "England",
		NationalityISO:      "gb-eng",
		NationalityFIFACode: "ENG",
		AttributeMasked:     true,
		NumericAttributes:   map[string]int{"Pas": 15},
		PerformanceStatsNumeric: map[string]float64{
			"Gls/90": 0.5,
		},
		ParsedPositions: []string{"M (C)"},
		ShortPositions:  []string{"MC"},
		PositionGroups:  []string{"Midfielders"},
	}

	row := playerToMaterializeRow(p)

	if row.PlayerID != p.UID {
		t.Errorf("PlayerID = %d, want %d", row.PlayerID, p.UID)
	}
	if row.Name != p.Name {
		t.Errorf("Name = %q, want %q", row.Name, p.Name)
	}
	if row.Position != p.Position {
		t.Errorf("Position = %q, want %q", row.Position, p.Position)
	}
	if row.Age != p.Age {
		t.Errorf("Age = %q, want %q", row.Age, p.Age)
	}
	if row.Club != p.Club {
		t.Errorf("Club = %q, want %q", row.Club, p.Club)
	}
	if row.Division != p.Division {
		t.Errorf("Division = %q, want %q", row.Division, p.Division)
	}
	if row.BasedIn != p.BasedIn {
		t.Errorf("BasedIn = %q, want %q", row.BasedIn, p.BasedIn)
	}
	if row.TransferValue != p.TransferValue {
		t.Errorf("TransferValue = %q, want %q", row.TransferValue, p.TransferValue)
	}
	if row.TransferValueAmount != p.TransferValueAmount {
		t.Errorf("TransferValueAmount = %d, want %d", row.TransferValueAmount, p.TransferValueAmount)
	}
	if row.Wage != p.Wage {
		t.Errorf("Wage = %q, want %q", row.Wage, p.Wage)
	}
	if row.WageAmount != p.WageAmount {
		t.Errorf("WageAmount = %d, want %d", row.WageAmount, p.WageAmount)
	}
	if row.Personality != p.Personality {
		t.Errorf("Personality = %q, want %q", row.Personality, p.Personality)
	}
	if row.MediaHandling != p.MediaHandling {
		t.Errorf("MediaHandling = %q, want %q", row.MediaHandling, p.MediaHandling)
	}
	if row.Nationality != p.Nationality {
		t.Errorf("Nationality = %q, want %q", row.Nationality, p.Nationality)
	}
	if row.NationalityISO != p.NationalityISO {
		t.Errorf("NationalityISO = %q, want %q", row.NationalityISO, p.NationalityISO)
	}
	if row.NationalityFIFACode != p.NationalityFIFACode {
		t.Errorf("NationalityFIFACode = %q, want %q", row.NationalityFIFACode, p.NationalityFIFACode)
	}
	if row.AttributeMasked != p.AttributeMasked {
		t.Errorf("AttributeMasked = %v, want %v", row.AttributeMasked, p.AttributeMasked)
	}
	if row.NumericAttributes["Pas"] != 15 {
		t.Errorf("NumericAttributes[Pas] = %d, want 15", row.NumericAttributes["Pas"])
	}
	if row.PerformanceStatsNumeric["Gls/90"] != 0.5 {
		t.Errorf("PerformanceStatsNumeric[Gls/90] = %v, want 0.5", row.PerformanceStatsNumeric["Gls/90"])
	}
	if len(row.ParsedPositions) != 1 || row.ParsedPositions[0] != "M (C)" {
		t.Errorf("ParsedPositions = %v, want [M (C)]", row.ParsedPositions)
	}
	if len(row.ShortPositions) != 1 || row.ShortPositions[0] != "MC" {
		t.Errorf("ShortPositions = %v, want [MC]", row.ShortPositions)
	}
	if len(row.PositionGroups) != 1 || row.PositionGroups[0] != "Midfielders" {
		t.Errorf("PositionGroups = %v, want [Midfielders]", row.PositionGroups)
	}
}

func TestMaterializeDatasetNoopWhenURLUnset(t *testing.T) {
	original := datasetQueryServiceURL
	datasetQueryServiceURL = ""
	defer func() { datasetQueryServiceURL = original }()

	err := materializeDataset(context.Background(), "dataset-1", []Player{{UID: 1, Name: "P"}})
	if err != nil {
		t.Fatalf("materializeDataset returned an error when URL is unset (should be a no-op): %v", err)
	}
}

func TestMaterializeDatasetSuccessPostsExpectedPayload(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody materializeRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decoding request body: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	original := datasetQueryServiceURL
	datasetQueryServiceURL = server.URL
	defer func() { datasetQueryServiceURL = original }()

	players := []Player{
		{UID: 1, Name: "Alice", NumericAttributes: map[string]int{"Pas": 10}},
		{UID: 2, Name: "Bob", NumericAttributes: map[string]int{"Tck": 12}},
	}

	if err := materializeDataset(context.Background(), "dataset-xyz", players); err != nil {
		t.Fatalf("materializeDataset failed: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/internal/materialize/dataset-xyz" {
		t.Errorf("path = %q, want /internal/materialize/dataset-xyz", gotPath)
	}
	if len(gotBody.Players) != 2 {
		t.Fatalf("got %d players, want 2", len(gotBody.Players))
	}
	if gotBody.Players[0].PlayerID != 1 || gotBody.Players[0].Name != "Alice" {
		t.Errorf("player 0 = %+v, want PlayerID=1 Name=Alice", gotBody.Players[0])
	}
	if gotBody.Players[1].PlayerID != 2 || gotBody.Players[1].Name != "Bob" {
		t.Errorf("player 1 = %+v, want PlayerID=2 Name=Bob", gotBody.Players[1])
	}
}

func TestMaterializeDatasetFailurePropagatesError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	original := datasetQueryServiceURL
	datasetQueryServiceURL = server.URL
	defer func() { datasetQueryServiceURL = original }()

	err := materializeDataset(context.Background(), "dataset-1", []Player{{UID: 1, Name: "P"}})
	if err == nil {
		t.Fatal("materializeDataset returned nil error, want a non-nil error on a 500 response")
	}
}
