package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPlayerReferencesResolveAcrossDatasets(t *testing.T) {
	previous := storage
	storage = CreateInMemoryStorage()
	t.Cleanup(func() { storage = previous })
	ids := []string{"ce27fb2f-5d0f-4a09-b11c-591dcbf25d91", "ce27fb2f-5d0f-4a09-b11c-591dcbf25d92", "ce27fb2f-5d0f-4a09-b11c-591dcbf25d93"}
	for i, id := range ids[:2] {
		if err := storage.Store(id, DatasetData{Players: []Player{{UID: int64(i + 1), Name: id}}, CurrencySymbol: "GBP"}); err != nil {
			t.Fatal(err)
		}
	}
	refs := []ShortlistPlayerRef{{ids[0], 1}, {ids[1], 2}, {ids[2], 3}, {ids[0], 999}}
	body, _ := json.Marshal(map[string]any{"references": refs})
	w := httptest.NewRecorder()
	playerReferencesHandler(w, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(body))))
	var result struct {
		Items []resolvedPlayerReference `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if w.Code != 200 || len(result.Items) != 4 {
		t.Fatalf("unexpected response %d %s", w.Code, w.Body.String())
	}
	for i, state := range []string{"available", "available", "expired", "missing"} {
		if result.Items[i].State != state {
			t.Fatalf("item %d state=%s", i, result.Items[i].State)
		}
	}
}

func TestProfileBackupValidationRejectsSnapshotsAndInvalidUnits(t *testing.T) {
	base := `{"version":1,"shortlists":{"version":1,"lists":[{"id":"a","name":"A","currency":"GBP","wagePeriod":"week","items":[]}]},"savedSearches":{"version":1,"searches":[]},"progressionProjects":{"version":1,"projects":[]}}`
	for _, tc := range []struct {
		body   string
		status int
	}{
		{base, 200},
		{strings.Replace(base, `"items":[]`, `"items":[],"players":[{"name":"Private snapshot"}]`, 1), 400},
		{strings.Replace(base, `"GBP"`, `"INVALID"`, 1), 400},
		{strings.Replace(base, `"week"`, `"day"`, 1), 400},
		{base + base, 400},
	} {
		w := httptest.NewRecorder()
		profileBackupValidateHandler(w, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.body)))
		if w.Code != tc.status {
			t.Fatalf("got %d want %d: %s", w.Code, tc.status, w.Body.String())
		}
	}
}

func TestScenarioValidationRejectsConflictingReferences(t *testing.T) {
	ref := ShortlistPlayerRef{"ce27fb2f-5d0f-4a09-b11c-591dcbf25d91", 1}
	s := RecruitmentScenario{ID: "one", Name: "One", DatasetID: ref.DatasetID, Signings: []ShortlistPlayerRef{ref}}
	if !validateRecruitmentScenarios([]RecruitmentScenario{s}) {
		t.Fatal("valid scenario rejected")
	}
	s.Sales = []ScenarioSale{{PlayerRef: ref, Fee: 10}}
	if validateRecruitmentScenarios([]RecruitmentScenario{s}) {
		t.Fatal("cannot sign and sell the same Player reference")
	}
}
