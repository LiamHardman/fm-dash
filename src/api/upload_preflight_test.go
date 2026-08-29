package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"os"
)

func TestUploadPreflightAcceptsProvidedMoneyballExport(t *testing.T) {
	content, err := os.ReadFile("../../moneyball_export_20260827_120808.csv")
	if err != nil { t.Skipf("provided export is not available in this checkout: %v", err) }
	format, headers, err := extractExportHeaders("moneyball_export_20260827_120808.csv", content)
	if err != nil { t.Fatalf("extract headers: %v", err) }
	if format != "csv" { t.Fatalf("expected csv, got %q", format) }
	if missing := missingRequiredExportHeaders(headers); len(missing) > 0 { t.Fatalf("provided export is missing required columns: %v", missing) }
}

func TestUploadPreflightAcceptsCompleteCSVProfile(t *testing.T) {
	response := runUploadPreflight(t, "players.csv", strings.Join(requiredExportHeaders, ";")+"\n")

	if !response.Valid {
		t.Fatalf("expected complete profile to be valid, missing %v", response.MissingColumns)
	}
	if response.Format != "csv" {
		t.Fatalf("expected csv format, got %q", response.Format)
	}
}

func TestUploadPreflightAcceptsCompleteHTMLProfile(t *testing.T) {
	headers := "<th>" + strings.Join(requiredExportHeaders, "</th><th>") + "</th>"
	response := runUploadPreflight(t, "players.html", "<table><thead><tr>"+headers+"</tr></thead></table>")

	if !response.Valid {
		t.Fatalf("expected complete profile to be valid, missing %v", response.MissingColumns)
	}
	if response.Format != "html" {
		t.Fatalf("expected html format, got %q", response.Format)
	}
}

func TestUploadPreflightReportsMissingColumns(t *testing.T) {
	response := runUploadPreflight(t, "players.csv", "Name;Position;Age;Acc\n")

	if response.Valid {
		t.Fatal("expected incomplete profile to be invalid")
	}
	if !containsPreflightColumn(response.MissingColumns, "Pac") {
		t.Fatalf("expected missing columns to include Pac, got %v", response.MissingColumns)
	}
	if containsPreflightColumn(response.MissingColumns, "Name") {
		t.Fatalf("did not expect Name to be missing, got %v", response.MissingColumns)
	}
}

func runUploadPreflight(t *testing.T, filename, content string) UploadPreflightResponse {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("playerFile", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/upload/preflight", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()

	uploadPreflightHandler(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response UploadPreflightResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return response
}

func containsPreflightColumn(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
