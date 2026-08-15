package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makePlayer(uid int64, name, age string, overall int, valueAmount, wageAmount int64) Player {
	return Player{
		UID:                 uid,
		Name:                name,
		Position:            "CM",
		Age:                 age,
		Club:                "Test FC",
		Division:            "Test Division",
		Overall:             overall,
		TransferValueAmount: valueAmount,
		WageAmount:          wageAmount,
		NumericAttributes:   map[string]int{"Passing": 12},
	}
}

func postProgressionAnalyze(t *testing.T, body ProgressionAnalyzeRequest, query string) *httptest.ResponseRecorder {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/progression/analyze"+query, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	progressionAnalyzeHandler(rec, req)
	return rec
}

func TestProgressionAnalyze_IntersectionOverlapping(t *testing.T) {
	ds1, ds2 := "prog-test-overlap-1", "prog-test-overlap-2"
	SetPlayerData(ds1, []Player{
		makePlayer(1, "Alice", "20", 60, 1000000, 5000),
		makePlayer(2, "Bob", "22", 65, 2000000, 6000),
	}, "£")
	SetPlayerData(ds2, []Player{
		makePlayer(1, "Alice", "22", 65, 1500000, 6000),
		makePlayer(3, "Charlie", "24", 70, 3000000, 7000),
	}, "£")
	defer DeleteDataset(ds1)
	defer DeleteDataset(ds2)

	rec := postProgressionAnalyze(t, ProgressionAnalyzeRequest{DatasetIDs: []string{ds1, ds2}}, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp ProgressionAnalyzeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Players) != 1 || resp.Players[0].UID != 1 {
		t.Fatalf("expected only UID 1 in intersection, got %+v", resp.Players)
	}
}

func TestProgressionAnalyze_IntersectionDisjoint(t *testing.T) {
	ds1, ds2 := "prog-test-disjoint-1", "prog-test-disjoint-2"
	SetPlayerData(ds1, []Player{makePlayer(1, "Alice", "20", 60, 1000000, 5000)}, "£")
	SetPlayerData(ds2, []Player{makePlayer(2, "Bob", "22", 65, 2000000, 6000)}, "£")
	defer DeleteDataset(ds1)
	defer DeleteDataset(ds2)

	rec := postProgressionAnalyze(t, ProgressionAnalyzeRequest{DatasetIDs: []string{ds1, ds2}}, "")
	var resp ProgressionAnalyzeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.EmptyIntersection {
		t.Fatalf("expected emptyIntersection=true, got %+v", resp)
	}
}

func TestProgressionAnalyze_IntersectionFullyIdenticalThreeWay(t *testing.T) {
	ds1, ds2, ds3 := "prog-test-identical-1", "prog-test-identical-2", "prog-test-identical-3"
	SetPlayerData(ds1, []Player{
		makePlayer(1, "Alice", "18", 55, 500000, 3000),
		makePlayer(2, "Bob", "19", 58, 600000, 3200),
	}, "£")
	SetPlayerData(ds2, []Player{
		makePlayer(1, "Alice", "19", 60, 700000, 3500),
		makePlayer(2, "Bob", "20", 62, 800000, 3700),
	}, "£")
	SetPlayerData(ds3, []Player{
		makePlayer(1, "Alice", "20", 65, 900000, 4000),
		makePlayer(2, "Bob", "21", 66, 1000000, 4200),
	}, "£")
	defer DeleteDataset(ds1)
	defer DeleteDataset(ds2)
	defer DeleteDataset(ds3)

	rec := postProgressionAnalyze(t, ProgressionAnalyzeRequest{DatasetIDs: []string{ds1, ds2, ds3}}, "")
	var resp ProgressionAnalyzeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Players) != 2 {
		t.Fatalf("expected 2 intersected players, got %d", len(resp.Players))
	}
	for _, p := range resp.Players {
		if len(p.Snapshots) != 3 {
			t.Fatalf("expected 3 snapshots per player, got %d", len(p.Snapshots))
		}
	}
}

func TestProgressionAnalyze_OrderingDistinctAges(t *testing.T) {
	older, younger := "prog-test-order-older", "prog-test-order-younger"
	// older dataset has the higher mean age even though it's listed second in the request
	SetPlayerData(older, []Player{makePlayer(1, "Alice", "30", 70, 1000000, 5000)}, "£")
	SetPlayerData(younger, []Player{makePlayer(1, "Alice", "20", 60, 800000, 4000)}, "£")
	defer DeleteDataset(older)
	defer DeleteDataset(younger)

	rec := postProgressionAnalyze(t, ProgressionAnalyzeRequest{DatasetIDs: []string{older, younger}}, "")
	var resp ProgressionAnalyzeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Order) != 2 || resp.Order[0] != younger || resp.Order[1] != older {
		t.Fatalf("expected order [younger, older], got %v", resp.Order)
	}
	// first snapshot should be the younger (earlier) one
	if resp.Players[0].Snapshots[0].Overall != 60 {
		t.Fatalf("expected first snapshot overall 60, got %d", resp.Players[0].Snapshots[0].Overall)
	}
}

func TestProgressionAnalyze_OrderingTiedMeansAmbiguous(t *testing.T) {
	ds1, ds2 := "prog-test-tie-1", "prog-test-tie-2"
	SetPlayerData(ds1, []Player{makePlayer(1, "Alice", "25", 70, 1000000, 5000)}, "£")
	SetPlayerData(ds2, []Player{makePlayer(1, "Alice", "25", 72, 1100000, 5200)}, "£")
	defer DeleteDataset(ds1)
	defer DeleteDataset(ds2)

	rec := postProgressionAnalyze(t, ProgressionAnalyzeRequest{DatasetIDs: []string{ds1, ds2}}, "")
	var resp ProgressionAnalyzeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.OrderAmbiguous {
		t.Fatalf("expected orderAmbiguous=true for tied means, got %+v", resp)
	}
	if len(resp.AmbiguousDatasetIDs) != 2 {
		t.Fatalf("expected both datasets flagged ambiguous, got %v", resp.AmbiguousDatasetIDs)
	}
}

func TestProgressionAnalyze_OrderingSkipsUnparsableAge(t *testing.T) {
	ds1, ds2 := "prog-test-unparsable-1", "prog-test-unparsable-2"
	// ds1's only player has an unparsable age; mean should be treated as unset -> ambiguous
	// only if another dataset ties with it. Here ds2 has a clean mean, so ds1 alone being
	// unset is fine as long as it doesn't collide with another unset dataset.
	SetPlayerData(ds1, []Player{makePlayer(1, "Alice", "N/A", 70, 1000000, 5000)}, "£")
	SetPlayerData(ds2, []Player{makePlayer(1, "Alice", "25", 72, 1100000, 5200)}, "£")
	defer DeleteDataset(ds1)
	defer DeleteDataset(ds2)

	rec := postProgressionAnalyze(t, ProgressionAnalyzeRequest{DatasetIDs: []string{ds1, ds2}}, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp ProgressionAnalyzeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.OrderAmbiguous {
		t.Fatalf("did not expect ambiguity when only one dataset has an unparsable-age mean, got %+v", resp)
	}
}

func TestProgressionAnalyze_ExplicitOrderOverride(t *testing.T) {
	ds1, ds2 := "prog-test-override-1", "prog-test-override-2"
	SetPlayerData(ds1, []Player{makePlayer(1, "Alice", "25", 70, 1000000, 5000)}, "£")
	SetPlayerData(ds2, []Player{makePlayer(1, "Alice", "25", 72, 1100000, 5200)}, "£")
	defer DeleteDataset(ds1)
	defer DeleteDataset(ds2)

	rec := postProgressionAnalyze(t, ProgressionAnalyzeRequest{
		DatasetIDs: []string{ds1, ds2},
		Order:      []string{ds2, ds1},
	}, "")
	var resp ProgressionAnalyzeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.OrderAmbiguous {
		t.Fatalf("explicit order override should bypass ambiguity detection, got %+v", resp)
	}
	if resp.Players[0].Snapshots[0].Overall != 72 {
		t.Fatalf("expected explicit order to be honored (ds2 first), got overall %d", resp.Players[0].Snapshots[0].Overall)
	}
}

func TestProgressionAnalyze_DeltaSort(t *testing.T) {
	ds1, ds2 := "prog-test-sort-1", "prog-test-sort-2"
	SetPlayerData(ds1, []Player{
		makePlayer(1, "SmallGain", "20", 60, 1000000, 5000),
		makePlayer(2, "BigGain", "20", 60, 1000000, 5000),
	}, "£")
	SetPlayerData(ds2, []Player{
		makePlayer(1, "SmallGain", "21", 62, 1000000, 5000),
		makePlayer(2, "BigGain", "21", 75, 1000000, 5000),
	}, "£")
	defer DeleteDataset(ds1)
	defer DeleteDataset(ds2)

	rec := postProgressionAnalyze(t, ProgressionAnalyzeRequest{DatasetIDs: []string{ds1, ds2}}, "?sortField=Overall&sortDir=desc")
	var resp ProgressionAnalyzeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Players) != 2 || resp.Players[0].Name != "BigGain" {
		t.Fatalf("expected BigGain sorted first by Overall delta desc, got %+v", resp.Players)
	}
}

func TestProgressionAnalyze_EdgeCases(t *testing.T) {
	t.Run("fewer than 2 datasets", func(t *testing.T) {
		rec := postProgressionAnalyze(t, ProgressionAnalyzeRequest{DatasetIDs: []string{"only-one"}}, "")
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("missing dataset returns 404 with the offending id", func(t *testing.T) {
		ds1 := "prog-test-edge-exists"
		SetPlayerData(ds1, []Player{makePlayer(1, "Alice", "20", 60, 1000000, 5000)}, "£")
		defer DeleteDataset(ds1)

		rec := postProgressionAnalyze(t, ProgressionAnalyzeRequest{DatasetIDs: []string{ds1, "does-not-exist"}}, "")
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("wrong method rejected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/progression/analyze", nil)
		rec := httptest.NewRecorder()
		progressionAnalyzeHandler(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected 405, got %d", rec.Code)
		}
	})
}

func TestProgressionDelta(t *testing.T) {
	p := ProgressionPlayer{
		Snapshots: []Player{
			makePlayer(1, "Alice", "20", 60, 1000000, 5000),
			makePlayer(1, "Alice", "22", 75, 1500000, 6000),
		},
	}

	cases := []struct {
		field    string
		expected float64
	}{
		{"Overall", 15},
		{"Value", 500000},
		{"Wage", 1000},
		{"Age", 2},
		{"Passing", 0}, // NumericAttributes fallback, both snapshots set to 12 in makePlayer
	}

	for _, c := range cases {
		got, ok := progressionDelta(p, c.field)
		if !ok {
			t.Fatalf("expected ok=true for field %q", c.field)
		}
		if got != c.expected {
			t.Errorf("field %q: expected delta %v, got %v", c.field, c.expected, got)
		}
	}

	if _, ok := progressionDelta(p, "NotARealField"); ok {
		t.Errorf("expected ok=false for unknown field")
	}
}
