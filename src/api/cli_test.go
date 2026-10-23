package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

func runTestCLI(t *testing.T, input string, args ...string) (int, cliEnvelope, []byte) {
	t.Helper()
	var out, logs bytes.Buffer
	code := RunCLI(context.Background(), args, strings.NewReader(input), &out, &logs)
	var envelope cliEnvelope
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
		t.Fatalf("invalid stdout (code %d): %s; stderr: %s", code, out.String(), logs.String())
	}
	return code, envelope, out.Bytes()
}

func TestCLIParsingAndCorpus(t *testing.T) {
	workspace := t.TempDir()
	code, e, _ := runTestCLI(t, "", "parse", "players", "testdata/cli/players.csv", "--strict", "--workers", "2", "--workspace", workspace)
	if code != 0 {
		t.Fatalf("parse: %+v", e)
	}
	data, _ := json.Marshal(e.Data)
	var parsed ParsedPlayers
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}
	if len(parsed.Players) != 2 || parsed.Players[0].Name != "Fernández" || parsed.Players[0].Overall == 0 || parsed.Counts.Seen != 2 {
		t.Fatalf("wrong parse: %+v", parsed)
	}
	entries, err := os.ReadDir(workspace)
	if err != nil || len(entries) != 0 {
		t.Fatalf("direct parse must not persist state: %v %v", entries, err)
	}
	for _, command := range []string{"validate", "bench"} {
		code, e, _ = runTestCLI(t, "", "parse", command, "--manifest", "testdata/cli/corpus.json", "--iterations", "2", "--workers", "2")
		if code != 0 {
			t.Fatalf("%s failed: %+v", command, e)
		}
	}
}

func TestCLIParseDiagnosticsAndCSVQuoting(t *testing.T) {
	input := "\ufeffUID;Name;Position;Age;Wage\n1;\"Name; With Delimiter\";M (C);24;£1K\n2;;GK;21;£2K\n3;\"Two\nLines\";GK;20;£2K\n"
	for _, strict := range []bool{false, true} {
		args := []string{"parse", "players", "-", "--input-format", "csv", "--stage", "parsed", "--workers", "3"}
		if strict {
			args = append(args, "--strict")
		}
		code, e, _ := runTestCLI(t, input, args...)
		if strict && code != 3 || !strict && code != 0 {
			t.Fatalf("strict=%v: code %d %+v", strict, code, e)
		}
		data, _ := json.Marshal(e.Data)
		var p ParsedPlayers
		_ = json.Unmarshal(data, &p)
		if p.Counts.Seen != 3 || p.Counts.Accepted != 2 || p.Counts.Rejected != 1 || p.Players[0].Name != "Name; With Delimiter" || p.Players[1].Name != "Two\nLines" || p.Players[0].Overall != 0 {
			t.Fatalf("unexpected CSV parse: %+v", p)
		}
	}
	code, _, _ := runTestCLI(t, "Name;Position;Age\n\"unterminated", "parse", "players", "-", "--input-format", "csv")
	if code != 3 {
		t.Fatalf("malformed CSV should fail, code %d", code)
	}
}

func TestCLIUsageOutputAndCancellation(t *testing.T) {
	for _, args := range [][]string{{"unknown"}, {"parse", "players", "-"}, {"parse", "players", "x.csv", "--workers", "0"}, {"version", "--timeout", "0s"}, {"version", "--format", "invalid"}, {"version", "--missing"}} {
		code, _, _ := runTestCLI(t, "", args...)
		if code == 0 {
			t.Fatalf("expected error: %v", args)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var out, logs bytes.Buffer
	code := RunCLI(ctx, []string{"parse", "players", "testdata/cli/players.csv"}, strings.NewReader(""), &out, &logs)
	if code != 130 {
		t.Fatalf("cancel code=%d output=%s", code, out.String())
	}
	path := filepath.Join(t.TempDir(), "result.json")
	if err := os.WriteFile(path, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	code = RunCLI(context.Background(), []string{"version", "--output", path}, strings.NewReader(""), &out, &logs)
	if code != 6 {
		t.Fatalf("overwrite without flag should fail: %d", code)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "keep" {
		t.Fatal("existing output was changed")
	}
	code = RunCLI(context.Background(), []string{"version", "--output", path, "--overwrite"}, strings.NewReader(""), &out, &logs)
	if code != 0 {
		t.Fatalf("explicit overwrite failed: %s", logs.String())
	}
}

func TestCLIDiffRejectsChangedValuesAndMissingRows(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "base.json")
	candidate := filepath.Join(dir, "candidate.json")
	for _, row := range []string{"changed", "removed", "reordered"} {
		a := map[string]any{"players": []any{map[string]any{"uid": 1, "name": "one"}, map[string]any{"uid": 2, "name": "two"}}, "sourceRows": []int{2, 3}, "timings": map[string]any{"totalMs": 1}}
		b := map[string]any{"players": []any{map[string]any{"uid": 1, "name": "one"}, map[string]any{"uid": 2, "name": "two"}}, "sourceRows": []int{2, 3}, "timings": map[string]any{"totalMs": 99}}
		switch row {
		case "changed":
			b["players"].([]any)[0].(map[string]any)["name"] = "wrong"
		case "removed":
			b["players"] = b["players"].([]any)[:1]
		case "reordered":
			b["players"] = []any{b["players"].([]any)[1], b["players"].([]any)[0]}
		}
		for path, value := range map[string]any{base: a, candidate: b} {
			data, _ := json.Marshal(value)
			if err := os.WriteFile(path, data, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		code, e, _ := runTestCLI(t, "", "parse", "diff", "--baseline", base, "--candidate", candidate)
		if row == "reordered" && code != 0 || row != "reordered" && code != 4 {
			t.Fatalf("%s code=%d %+v", row, code, e)
		}
	}
}

func TestCLIBenchmarkAccuracyAndTimingGates(t *testing.T) {
	dir := t.TempDir()
	fixture, _ := filepath.Abs("testdata/cli/players.csv")
	manifest := filepath.Join(dir, "corpus.json")
	baseline := filepath.Join(dir, "baseline.json")
	writeManifest := func(count int) {
		data, _ := json.Marshal(map[string]any{"version": 1, "cases": []any{map[string]any{"name": "players", "path": fixture, "expectedCount": count}}})
		if err := os.WriteFile(manifest, data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	writeManifest(99)
	code, _, _ := runTestCLI(t, "", "parse", "bench", "--manifest", manifest, "--iterations", "1", "--warmup", "0")
	if code != 4 {
		t.Fatalf("incorrect row count passed: %d", code)
	}
	writeManifest(2)
	code, e, _ := runTestCLI(t, "", "parse", "bench", "--manifest", manifest, "--iterations", "1", "--warmup", "0")
	if code != 0 {
		t.Fatalf("benchmark: %+v", e)
	}
	data, _ := json.Marshal(e.Data)
	var report cliEvaluation
	_ = json.Unmarshal(data, &report)
	report.Cases[0].MedianMS = 1e-12
	data, _ = json.Marshal(map[string]any{"data": report})
	if err := os.WriteFile(baseline, data, 0o600); err != nil {
		t.Fatal(err)
	}
	code, _, _ = runTestCLI(t, "", "parse", "bench", "--manifest", manifest, "--iterations", "1", "--warmup", "0", "--baseline", baseline, "--max-regression", "0")
	if code != 4 {
		t.Fatalf("timing gate should fail: %d", code)
	}
}

func TestCLIWorkspaceLifecycleAndRevision(t *testing.T) {
	dir := t.TempDir()
	code, e, _ := runTestCLI(t, "", "datasets", "import", "testdata/cli/players.csv", "--workspace", dir, "--name", "First")
	if code != 0 {
		t.Fatalf("import: %+v", e)
	}
	data, _ := json.Marshal(e.Data)
	var dataset cliDataset
	_ = json.Unmarshal(data, &dataset)
	for _, command := range [][]string{{"datasets", "show"}, {"players", "list"}, {"squad", "best-xi"}, {"teams", "list"}, {"nations", "list"}} {
		args := append(command, "--workspace", dir, "--dataset", dataset.ID)
		code, e, _ = runTestCLI(t, "", args...)
		if code != 0 {
			t.Fatalf("%v: %+v", args, e)
		}
	}
	code, e, _ = runTestCLI(t, "", "shortlists", "list", "--workspace", dir)
	if code != 0 {
		t.Fatal(e)
	}
	doc := e.Data.(map[string]any)
	revision := doc["revision"].(string)
	body := `{"version":1,"lists":[{"id":"test","name":"Test","items":[],"currency":"GBP","wagePeriod":"week","scenarios":[]}]}`
	code, e, _ = runTestCLI(t, "", "shortlists", "set", "--workspace", dir, "--id", revision, "--body", body)
	if code != 0 {
		t.Fatalf("set: %+v", e)
	}
	code, _, _ = runTestCLI(t, "", "shortlists", "set", "--workspace", dir, "--id", revision, "--body", body)
	if code != 7 {
		t.Fatalf("stale revision should fail: %d", code)
	}
	code, _, _ = runTestCLI(t, "", "datasets", "delete", "--workspace", dir, "--dataset", dataset.ID)
	if code != 0 {
		t.Fatalf("delete: %d", code)
	}
	code, _, _ = runTestCLI(t, "", "datasets", "show", "--workspace", dir, "--dataset", dataset.ID)
	if code != 5 {
		t.Fatalf("deleted Dataset still available: %d", code)
	}
}

func TestCLIAPIRevisionAndErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/fail" {
			w.WriteHeader(412)
			_, _ = w.Write([]byte(`{"error":"conflict"}`))
			return
		}
		if r.Header.Get("If-Match") != "revision" || r.Header.Get("Accept") != "application/json" {
			t.Error("missing request headers")
		}
		w.Header().Set("ETag", "new-revision")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	code, e, _ := runTestCLI(t, "", "api", "/api/config", "--base-url", server.URL, "--id", "revision")
	if code != 0 {
		t.Fatalf("API: %+v", e)
	}
	code, _, _ = runTestCLI(t, "", "api", "/api/fail", "--base-url", server.URL)
	if code != 7 {
		t.Fatalf("API conflict code=%d", code)
	}
}

func TestSquadMatchesGUIFixture(t *testing.T) {
	data, err := os.ReadFile("testdata/cli/squad.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Players  []Player         `json:"players"`
		Slots    []FormationSlot  `json:"slots"`
		Expected SquadComposition `json:"expected"`
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	actual := ComputeSquadComposition(fixture.Players, fixture.Slots)
	if actual.Average != fixture.Expected.Average {
		t.Fatalf("average %d != GUI %d", actual.Average, fixture.Expected.Average)
	}
	for id, expected := range fixture.Expected.Slots {
		got := actual.Slots[id]
		if len(got) != len(expected) {
			t.Fatalf("slot %s length differs", id)
		}
		for i := range got {
			if got[i].Player.UID != expected[i].Player.UID || got[i].Overall != expected[i].Overall || got[i].Exact != expected[i].Exact {
				t.Fatalf("slot %s differs: %+v != %+v", id, got[i], expected[i])
			}
		}
	}
}

func TestParserWorkersAreDeterministicAndCancel(t *testing.T) {
	input := "UID;Name;Position;Age;Wage\n"
	for i := 1; i <= 150; i++ {
		input += fmt.Sprintf("%d;Player %d;M (C);24;£1K\n", i, i)
	}
	var baseline any
	for _, workers := range []int{1, 2, 8} {
		p, err := ParsePlayers(context.Background(), strings.NewReader(input), ParseOptions{"csv", "complete", workers, true})
		if err != nil {
			t.Fatal(err)
		}
		semantic, err := cliSemantic(p)
		if err != nil {
			t.Fatal(err)
		}
		if baseline == nil {
			baseline = semantic
		} else if !reflect.DeepEqual(baseline, semantic) {
			t.Fatalf("semantic output depends on worker scheduling: %+v", cliDifferences(baseline, semantic))
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	_, err := ParsePlayers(ctx, strings.NewReader(input), ParseOptions{"csv", "complete", 4, true})
	if err == nil {
		t.Fatal("expired context should fail")
	}
}

func TestCLIConfigurationProfilesAndRecipes(t *testing.T) {
	w := t.TempDir()
	code, e, _ := runTestCLI(t, "", "config", "set", "--workspace", w, "--body", `{"useScaledRatings":false}`)
	if code != 0 {
		t.Fatal(e)
	}
	code, e, _ = runTestCLI(t, "", "config", "show", "--workspace", w)
	if code != 0 || e.Data.(map[string]any)["useScaledRatings"] != false {
		t.Fatalf("config was not persisted: %+v", e)
	}
	code, e, _ = runTestCLI(t, "", "weights", "set", "--workspace", w, "--body", `{"id":"test","name":"Test","weights":{"PAC":{"Acc":100}}}`)
	if code != 0 {
		t.Fatal(e)
	}
	code, e, _ = runTestCLI(t, "", "weights", "activate", "--workspace", w, "--id", "test")
	if code != 0 {
		t.Fatal(e)
	}
	code, e, _ = runTestCLI(t, "", "config", "show", "--workspace", w)
	if code != 0 {
		t.Fatal(e)
	}
	weights := e.Data.(map[string]any)["attributeWeights"].(map[string]any)
	if weights["PAC"].(map[string]any)["Acc"] != float64(100) {
		t.Fatal("active weights were not persisted")
	}
	code, e, _ = runTestCLI(t, "", "recipes", "encode", "--body", `{"name":"Fernández","minOverall":75}`)
	if code != 0 {
		t.Fatal(e)
	}
	recipe := e.Data.(map[string]any)["recipe"].(string)
	code, e, _ = runTestCLI(t, "", "recipes", "decode", "--query", recipe)
	if code != 0 || e.Data.(map[string]any)["name"] != "Fernández" {
		t.Fatal(e)
	}
	code, _, _ = runTestCLI(t, "", "recipes", "encode", "--body", `{"datasetId":"private"}`)
	if code != 3 {
		t.Fatalf("Dataset identity in recipe accepted: %d", code)
	}
}

func TestCLIJSONLinesAndUpload(t *testing.T) {
	var out, logs bytes.Buffer
	code := RunCLI(context.Background(), []string{"parse", "players", "testdata/cli/players.csv", "--format", "jsonl"}, strings.NewReader(""), &out, &logs)
	if code != 0 {
		t.Fatalf("JSONL: %s", out.String())
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 4 {
		t.Fatalf("expected header, two players and summary: %d", len(lines))
	}
	for i, line := range lines {
		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatal(err)
		}
		if i == 0 && record["type"] != "header" || i == 3 && record["type"] != "summary" {
			t.Fatal("invalid record framing")
		}
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/upload" {
			t.Error("wrong upload request")
		}
		f, h, err := r.FormFile("playerFile")
		if err != nil {
			t.Error(err)
			return
		}
		defer func() { _ = f.Close() }()
		if h.Filename != "players.csv" {
			t.Error("filename lost")
		}
		_, _ = w.Write([]byte(`{"datasetId":"remote"}`))
	}))
	defer server.Close()
	code, e, _ := runTestCLI(t, "", "datasets", "upload", "testdata/cli/players.csv", "--base-url", server.URL)
	if code != 0 {
		t.Fatal(e)
	}
}

func TestCLIWorkspaceConcurrentMutationsAndDeadline(t *testing.T) {
	w, _ := cliOpenWorkspace(t.TempDir())
	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := w.mutate(ctx, func() error {
				path := filepath.Join(w.dir, "counter")
				data, _ := os.ReadFile(path)
				return atomicConfigWrite(path, append(data, 'x'))
			})
			if err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	data, _ := os.ReadFile(filepath.Join(w.dir, "counter"))
	if len(data) != 8 {
		t.Fatalf("lost writes: %q", data)
	}
	lock, err := os.OpenFile(filepath.Join(w.dir, "workspace.lock"), os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = lock.Close() }()
	if err = lockConfigFile(lock); err != nil {
		t.Fatal(err)
	}
	defer unlockConfigFile(lock)
	timeout, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
	defer cancel()
	if err = w.mutate(timeout, func() error { t.Error("must not mutate after deadline"); return nil }); err != context.DeadlineExceeded {
		t.Fatalf("lock wait ignored deadline: %v", err)
	}
}

func TestScenarioMatchesGUIPlanningContract(t *testing.T) {
	signing, sale := ShortlistPlayerRef{"a", 1}, ShortlistPlayerRef{"b", 2}
	result := EvaluateRecruitmentScenario(RecruitmentScenario{FeeBudget: 100, WageBudget: 10, Signings: []ShortlistPlayerRef{signing}, Sales: []ScenarioSale{{PlayerRef: sale, Fee: 30, Wage: 5}}}, []ShortlistItem{{PlayerRef: signing, TargetFee: 80, TargetWage: 12}}, []Player{{UID: 2, Position: "GK"}}, func(ShortlistPlayerRef) *Player { return &Player{Position: "ST"} })
	if result.FeeRemaining != 50 || result.WageRemaining != 3 || len(result.Coverage) != 2 || result.Coverage[0].Position != "GK" || result.Coverage[0].Count != 0 || result.Coverage[1].Position != "ST" || result.Coverage[1].Count != 1 {
		t.Fatalf("differs from profilePlanning.test.js: %+v", result)
	}
}

func TestNativeSaveCoverageAndTruncationErrors(t *testing.T) {
	content := lengthPrefixedString("Fernández")
	input := buildFakeFMSave(t, content)
	inspection, err := InspectFMSave(context.Background(), bytes.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if inspection.ScannedBytes != len(content) || inspection.Truncated || inspection.Mode != "strings" || len(inspection.Fields) != 1 {
		t.Fatalf("wrong native-save coverage: %+v", inspection)
	}
	_, err = InspectFMSave(context.Background(), bytes.NewReader(input[:len(input)-3]))
	if err == nil {
		t.Fatal("truncated compressed input was accepted")
	}
	input = buildFakeFMSave(t, make([]byte, fmSaveScanWindow+20))
	inspection, err = InspectFMSave(context.Background(), bytes.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if !inspection.Truncated || inspection.ScannedBytes != fmSaveScanWindow {
		t.Fatal("scan bound was not reported")
	}
}

func TestRepeatedPercentileCalculationPreservesActualRanks(t *testing.T) {
	makePlayers := func() []Player {
		return []Player{{Name: "one", PerformanceStatsNumeric: map[string]float64{"Gls/90": 1}, PositionGroups: []string{"Midfielders"}, ShortPositions: []string{"MC"}}, {Name: "two", PerformanceStatsNumeric: map[string]float64{"Gls/90": 3}, PositionGroups: []string{"Midfielders"}, ShortPositions: []string{"MC"}}}
	}
	a, b := makePlayers(), makePlayers()
	CalculatePlayerPerformancePercentiles(a)
	CalculatePlayerPerformancePercentiles(b)
	if a[0].PerformancePercentiles["Global"]["Gls/90"] >= a[1].PerformancePercentiles["Global"]["Gls/90"] {
		t.Fatal("different values must produce different ranks")
	}
	if !reflect.DeepEqual(a[0].PerformancePercentiles, b[0].PerformancePercentiles) || !reflect.DeepEqual(a[1].PerformancePercentiles, b[1].PerformancePercentiles) {
		t.Fatal("repeated calculation changed ranks")
	}
}

func TestCLIExportCategoriesAndProfilePreview(t *testing.T) {
	w := t.TempDir()
	code, e, _ := runTestCLI(t, "", "datasets", "import", "testdata/cli/players.csv", "--workspace", w)
	if code != 0 {
		t.Fatal(e)
	}
	id := e.Data.(map[string]any)["datasetId"].(string)
	code, e, _ = runTestCLI(t, "", "datasets", "export", "--workspace", w, "--dataset", id, "--body", `{"options":{"basicInfo":true}}`)
	if code != 0 {
		t.Fatal(e)
	}
	players := e.Data.(map[string]any)["players"].([]any)
	if _, hasWage := players[0].(map[string]any)["wage"]; hasWage {
		t.Fatal("unselected contract category was exported")
	}
	code, e, _ = runTestCLI(t, "", "profile", "export", "--workspace", w)
	if code != 0 {
		t.Fatal(e)
	}
	backup := e.Data.(map[string]any)
	backup["savedSearches"] = map[string]any{"version": 1, "searches": []any{map[string]any{"id": "new", "name": "New", "filters": map[string]any{"minOverall": 75}}}}
	data, _ := json.Marshal(backup)
	code, e, _ = runTestCLI(t, "", "profile", "restore", "--workspace", w, "--body", string(data), "--dry-run")
	if code != 0 {
		t.Fatal(e)
	}
	code, e, _ = runTestCLI(t, "", "saved-searches", "list", "--workspace", w)
	if code != 0 {
		t.Fatal(e)
	}
	document := e.Data.(map[string]any)["document"].(map[string]any)
	if len(document["searches"].([]any)) != 0 {
		t.Fatal("preview wrote profile data")
	}
	for _, query := range []string{`{"minUnknown":1}`, `{"ageRange":{"min":"bad"}}`, `{"minOverall":"bad"}`} {
		_, err := cliFilterPlayers(nil, mustFilter(t, query))
		if err == nil {
			t.Fatalf("invalid filter accepted: %s", query)
		}
	}
}

func mustFilter(t *testing.T, query string) map[string]any {
	t.Helper()
	var f map[string]any
	if err := cliDecodeJSON([]byte(query), &f); err != nil {
		t.Fatal(err)
	}
	return f
}
