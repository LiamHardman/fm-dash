package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func colourSetRequest(t *testing.T, method, scale string, document ColourSetDocument) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	colourSetsHandler(response, httptest.NewRequest(method, "/api/colour-sets/"+scale, bytes.NewReader(body)))
	return response
}

func testColourDocument(scale string, count int) ColourSetDocument {
	minimum, maximum := 0, 100
	if scale == "fm" {
		minimum, maximum = 1, 20
	}
	bands := make([]ColourBand, count)
	for i := range bands {
		bands[i] = ColourBand{Min: minimum + i*(maximum-minimum+1)/count, Color: "#123456"}
	}
	return ColourSetDocument{Version: 1, ActiveSetID: "custom", Sets: []ColourSet{{ID: "custom", Name: "Custom", Bands: bands}}}
}

func TestColourSetsPersistenceAndIndependentScales(t *testing.T) {
	previous := configStorage
	t.Cleanup(func() { configStorage = previous })
	dir := t.TempDir()
	configStorage = &localConfigStorage{dir: dir}
	response := colourSetRequest(t, http.MethodGet, "fm", ColourSetDocument{})
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"sets":[]`) {
		t.Fatalf("unexpected default: %d %s", response.Code, response.Body.String())
	}
	for _, scale := range []string{"fm", "fifa"} {
		for _, count := range []int{2, 20} {
			document := testColourDocument(scale, count)
			response = colourSetRequest(t, http.MethodPut, scale, document)
			if response.Code != http.StatusOK {
				t.Fatalf("save %s/%d: %d %s", scale, count, response.Code, response.Body.String())
			}
		}
	}
	// A new storage instance sees both documents, including the active selection.
	configStorage = &localConfigStorage{dir: dir}
	// Dataset expiry must never expire colour sets or list them as Datasets.
	datasets := &LocalFileStorage{datasetDir: dir}
	if err := datasets.CleanupOldDatasets(0, nil); err != nil {
		t.Fatal(err)
	}
	ids, err := datasets.List()
	if err != nil || len(ids) != 0 {
		t.Fatalf("colour sets appeared in Dataset listing: %v, %v", ids, err)
	}
	for _, scale := range []string{"fm", "fifa"} {
		response = colourSetRequest(t, http.MethodGet, scale, ColourSetDocument{})
		var got ColourSetDocument
		if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if response.Code != http.StatusOK || len(got.Sets) != 1 || len(got.Sets[0].Bands) != 20 || got.ActiveSetID != "custom" {
			t.Fatalf("read %s: %s", scale, response.Body.String())
		}
	}
	// Resetting FM must leave FIFA intact.
	response = colourSetRequest(t, http.MethodPut, "fm", ColourSetDocument{Version: 1, Sets: []ColourSet{}})
	if response.Code != http.StatusOK {
		t.Fatal(response.Body.String())
	}
	response = colourSetRequest(t, http.MethodGet, "fifa", ColourSetDocument{})
	if !strings.Contains(response.Body.String(), `"activeSetId":"custom"`) {
		t.Fatalf("FM update changed FIFA: %s", response.Body.String())
	}
}

func TestColourSetsRejectInvalidDocumentsWithoutOverwriting(t *testing.T) {
	previous := configStorage
	configStorage = &localConfigStorage{dir: t.TempDir()}
	t.Cleanup(func() { configStorage = previous })
	valid := testColourDocument("fm", 2)
	colourSetRequest(t, http.MethodPut, "fm", valid)
	cases := map[string]func(*ColourSetDocument){
		"empty bands":         func(d *ColourSetDocument) { d.Sets[0].Bands = nil },
		"too many bands":      func(d *ColourSetDocument) { d.Sets[0].Bands = testColourDocument("fifa", 21).Sets[0].Bands },
		"missing first value": func(d *ColourSetDocument) { d.Sets[0].Bands[0].Min = 2 },
		"overlap":             func(d *ColourSetDocument) { d.Sets[0].Bands[1].Min = 1 },
		"out of range":        func(d *ColourSetDocument) { d.Sets[0].Bands[1].Min = 21 },
		"unsafe colour":       func(d *ColourSetDocument) { d.Sets[0].Bands[0].Color = "red; background:url(x)" },
		"missing active set":  func(d *ColourSetDocument) { d.ActiveSetID = "missing" },
		"duplicate IDs":       func(d *ColourSetDocument) { d.Sets = append(d.Sets, d.Sets[0]) },
		"empty name":          func(d *ColourSetDocument) { d.Sets[0].Name = "   " },
		"unknown version":     func(d *ColourSetDocument) { d.Version = 2 },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			document := testColourDocument("fm", 2)
			mutate(&document)
			response := colourSetRequest(t, http.MethodPut, "fm", document)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("got %d: %s", response.Code, response.Body.String())
			}
		})
	}
	response := colourSetRequest(t, http.MethodGet, "fm", ColourSetDocument{})
	expected, _ := json.Marshal(valid)
	if strings.TrimSpace(response.Body.String()) != string(expected) {
		t.Fatalf("invalid save overwrote valid settings: %s", response.Body.String())
	}
}

type failingColourStorage struct{ missing bool }

func (s failingColourStorage) RetrieveConfig(_ string) ([]byte, error) {
	if s.missing {
		return nil, os.ErrNotExist
	}
	return nil, fmt.Errorf("storage offline")
}
func (s failingColourStorage) StoreConfig(_ string, _ []byte) error {
	return fmt.Errorf("storage offline")
}
func (s failingColourStorage) DeleteConfig(_ string) error { return nil }

func TestColourSetsStorageFailuresAndInvalidRequests(t *testing.T) {
	previous := configStorage
	t.Cleanup(func() { configStorage = previous })
	configStorage = failingColourStorage{}
	for _, method := range []string{http.MethodGet, http.MethodPut} {
		response := colourSetRequest(t, method, "fm", testColourDocument("fm", 2))
		if response.Code != http.StatusInternalServerError {
			t.Fatalf("storage failure returned %d", response.Code)
		}
	}
	configStorage = &localConfigStorage{dir: t.TempDir()}
	for _, body := range []string{
		`null`,
		`{"version":1,"sets":[],"activeSetId":""} {}`,
		`{"version":1,"sets":[],"extra":true}`,
		`{"version":1,"sets":[{"id":"x","name":"X","bands":[{"min":1.5,"color":"#ffffff"}]}]}`,
		strings.Repeat(" ", 2*1024*1024) + `{}`,
	} {
		response := httptest.NewRecorder()
		colourSetsHandler(response, httptest.NewRequest(http.MethodPut, "/api/colour-sets/fm", strings.NewReader(body)))
		if response.Code != http.StatusBadRequest {
			t.Fatalf("invalid JSON accepted: %d", response.Code)
		}
	}
}
