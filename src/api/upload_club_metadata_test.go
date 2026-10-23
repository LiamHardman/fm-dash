package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFreshUploadClubMetadata(t *testing.T) {
	headers := []string{"UID", "Name", "Club", "Division", "Based in", "Position", "Age"}
	rows := [][]string{
		{"1", "Adil Aouchiche", "Schalke 04", "Championnat Inferieures", "France", "M (C)", "24"},
		{"2", "Squad player A", "Schalke 04", "2. Bundesliga", "Germany", "M (C)", "25"},
		{"3", "Squad player B", "Schalke 04", "2. Bundesliga", "Germany", "M (C)", "26"},
	}
	csv := strings.Join(headers, ";") + "\n"
	html := "<table><tr><th>" + strings.Join(headers, "</th><th>") + "</th></tr>"
	for _, row := range rows {
		csv += strings.Join(row, ";") + "\n"
		html += "<tr><td>" + strings.Join(row, "</td><td>") + "</td></tr>"
	}
	html += "</table>"

	for _, tc := range []struct{ filename, content string }{
		{"players.csv", csv},
		{"players.html", html},
	} {
		t.Run(tc.filename, func(t *testing.T) {
			originalStorage := storage
			backend := newWaitableStorage()
			storage = backend
			t.Cleanup(func() { storage = originalStorage })

			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			part, err := writer.CreateFormFile("playerFile", tc.filename)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := part.Write([]byte(tc.content)); err != nil {
				t.Fatal(err)
			}
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest(http.MethodPost, "/api/upload", &body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			recorder := httptest.NewRecorder()
			uploadHandler(recorder, req)
			if recorder.Code != http.StatusOK {
				t.Fatalf("upload returned %d: %s", recorder.Code, recorder.Body.String())
			}
			var response UploadResponse
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				removeDuplicateMapping(response.DatasetID)
				storeMutex.Lock()
				delete(playerDataStore, response.DatasetID)
				storeMutex.Unlock()
			})
			// Wait for both the initial persistence and the background percentile save
			// before restoring global storage. The upload response is already captured.
			for i := 0; i < 2; i++ {
				select {
				case <-backend.stored:
				case <-time.After(5 * time.Second):
					t.Fatal("timed out waiting for upload persistence")
				}
			}
			waitForPendingPercentileCalc(response.DatasetID)

			assertSource := func(source string, players []Player) {
				t.Helper()
				if len(players) != len(rows) {
					t.Fatalf("%s: got %d players, want %d", source, len(players), len(rows))
				}
				for _, player := range players {
					expected := rows[player.UID-1]
					if player.Division != expected[3] || player.BasedIn != expected[4] {
						t.Errorf("%s: %s at %s has division %q and country %q", source, player.Name, player.Club, player.Division, player.BasedIn)
					}
				}
			}
			assertSource("first-view upload response", response.Players)
			players, _, found := GetPlayerData(response.DatasetID)
			if !found {
				t.Fatal("uploaded dataset not found")
			}
			assertSource("subsequent dataset load", players)
			teams := processTeamsData(players, "2. Bundesliga", "Germany")
			if len(teams) != 1 || teams[0].PlayerCount != 3 {
				t.Fatalf("club summary should group the whole squad: %+v", teams)
			}
		})
	}
}
