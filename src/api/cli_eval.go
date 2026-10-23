package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
	"time"
)

type cliCorpus struct {
	Version int             `json:"version"`
	Cases   []cliCorpusCase `json:"cases"`
}
type cliCorpusCase struct {
	Name                string           `json:"name"`
	Path                string           `json:"path"`
	SHA256              string           `json:"sha256"`
	Format              string           `json:"format"`
	Stage               string           `json:"stage"`
	ExpectedCount       *int             `json:"expectedCount"`
	ExpectedRejected    int              `json:"expectedRejected"`
	ExpectedErrors      int              `json:"expectedErrors"`
	ExpectedDiagnostics []string         `json:"expectedDiagnostics,omitempty"`
	Expected            []cliExpectation `json:"expected,omitempty"`
	Golden              string           `json:"golden,omitempty"`
}
type cliExpectation struct {
	Path  string `json:"path"` // JSON pointer, e.g. /players/0/name
	Value any    `json:"value"`
}
type cliDifference struct {
	Path     string `json:"path"`
	Kind     string `json:"kind"`
	Expected any    `json:"expected,omitempty"`
	Actual   any    `json:"actual,omitempty"`
}

func cliSemantic(value any) (any, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var root any
	if err = cliDecodeJSON(data, &root); err != nil {
		return nil, err
	}
	if m, ok := root.(map[string]any); ok {
		if _, envelope := m["schemaVersion"]; envelope {
			if m["status"] == "error" {
				return nil, cliFail(3, "cannot compare a failed command result")
			}
			root = m["data"]
			m, _ = root.(map[string]any)
		}
		delete(m, "timings")
		// Worker count controls execution, not semantic results.
		if options, ok := m["options"].(map[string]any); ok {
			delete(options, "workers")
		}
		if players, ok := m["players"].([]any); ok {
			rows, _ := m["sourceRows"].([]any)
			attrs, _ := m["rawAttributes"].([]any)
			counts := map[string]int{}
			for _, v := range players {
				p, _ := v.(map[string]any)
				counts[fmt.Sprint(p["uid"])]++
			}
			records := map[string]any{}
			for i, v := range players {
				p, _ := v.(map[string]any)
				uid := fmt.Sprint(p["uid"])
				key := "uid:" + uid
				if uid == "0" || counts[uid] != 1 || strings.HasPrefix(uid, "-") {
					if len(rows) != len(players) {
						return nil, cliFail(3, "ambiguous player identity without aligned sourceRows")
					}
					key = "row:" + fmt.Sprint(rows[i])
				}
				record := map[string]any{"player": v}
				if len(attrs) == len(players) {
					record["rawAttributes"] = attrs[i]
				}
				records[key] = record
			}
			m["players"] = records
			delete(m, "sourceRows")
			delete(m, "rawAttributes")
		}
	}
	return root, nil
}

func cliDifferences(expected, actual any) []cliDifference {
	changes := []cliDifference{}
	var visit func(string, any, any)
	visit = func(path string, a, b any) {
		if len(changes) >= 200 {
			return
		}
		am, aok := a.(map[string]any)
		bm, bok := b.(map[string]any)
		if aok && bok {
			keys := map[string]bool{}
			for k := range am {
				keys[k] = true
			}
			for k := range bm {
				keys[k] = true
			}
			ordered := make([]string, 0, len(keys))
			for k := range keys {
				ordered = append(ordered, k)
			}
			sort.Strings(ordered)
			for _, k := range ordered {
				av, ae := am[k]
				bv, be := bm[k]
				p := path + "/" + strings.ReplaceAll(strings.ReplaceAll(k, "~", "~0"), "/", "~1")
				if !ae {
					changes = append(changes, cliDifference{p, "added", nil, bv})
				} else if !be {
					changes = append(changes, cliDifference{p, "removed", av, nil})
				} else {
					visit(p, av, bv)
				}
				if len(changes) >= 200 {
					break
				}
			}
			return
		}
		as, aok := a.([]any)
		bs, bok := b.([]any)
		if aok && bok && len(as) == len(bs) {
			for i := range as {
				visit(path+"/"+strconv.Itoa(i), as[i], bs[i])
			}
			return
		}
		if !reflect.DeepEqual(a, b) {
			changes = append(changes, cliDifference{path, "changed", a, b})
		}
	}
	visit("", expected, actual)
	return changes
}

func cliPointer(value any, path string) (any, bool) {
	if path == "" {
		return value, true
	}
	if !strings.HasPrefix(path, "/") {
		return nil, false
	}
	for _, key := range strings.Split(path[1:], "/") {
		key = strings.ReplaceAll(strings.ReplaceAll(key, "~1", "/"), "~0", "~")
		switch v := value.(type) {
		case map[string]any:
			var ok bool
			value, ok = v[key]
			if !ok {
				return nil, false
			}
		case []any:
			i, err := strconv.Atoi(key)
			if err != nil || i < 0 || i >= len(v) {
				return nil, false
			}
			value = v[i]
		default:
			return nil, false
		}
	}
	return value, true
}

func cliDiff(ctx context.Context, o cliOptions, stdin io.Reader) (any, error) {
	if o.baseline == "" || o.candidate == "" {
		return nil, cliFail(2, "diff requires --baseline and --candidate")
	}
	var a, b any
	if err := cliReadJSON(ctx, o.baseline, stdin, o.maxBytes, &a); err != nil {
		return nil, err
	}
	if err := cliReadJSON(ctx, o.candidate, stdin, o.maxBytes, &b); err != nil {
		return nil, err
	}
	checkHash := func(value any) string {
		m, _ := value.(map[string]any)
		p, _ := m["provenance"].(map[string]any)
		s, _ := p["inputSha256"].(string)
		return s
	}
	if ah, bh := checkHash(a), checkHash(b); ah != "" && bh != "" && ah != bh {
		return nil, cliFail(2, "input hashes differ; semantic parser comparison requires the same fixture")
	}
	av, err := cliSemantic(a)
	if err != nil {
		return nil, err
	}
	bv, err := cliSemantic(b)
	if err != nil {
		return nil, err
	}
	changes := cliDifferences(av, bv)
	report := map[string]any{"equal": len(changes) == 0, "differences": changes, "detailLimit": 200, "excluded": []string{"envelope metadata", "timings", "options.workers"}, "numericTolerance": 0}
	if len(changes) > 0 {
		return report, cliFail(4, "semantic output differs")
	}
	return report, nil
}

type cliBenchSample struct {
	TotalMS        float64      `json:"totalMs"`
	ReadMS         float64      `json:"readMs"`
	SerializeMS    float64      `json:"serializeMs"`
	Timings        ParseTimings `json:"parseTimings"`
	AllocatedBytes uint64       `json:"allocatedBytes"`
	Allocations    uint64       `json:"allocations"`
	HeapBytes      uint64       `json:"heapBytes"`
}
type cliCaseReport struct {
	Name           string           `json:"name"`
	InputSHA256    string           `json:"inputSha256"`
	InputBytes     int              `json:"inputBytes"`
	Format         string           `json:"format"`
	Stage          string           `json:"stage"`
	Count          int              `json:"count"`
	Passed         bool             `json:"passed"`
	Failures       []string         `json:"failures"`
	Samples        []cliBenchSample `json:"samples,omitempty"`
	MedianMS       float64          `json:"medianMs,omitempty"`
	P95MS          float64          `json:"p95Ms,omitempty"`
	RowsPerSecond  float64          `json:"rowsPerSecond,omitempty"`
	BytesPerSecond float64          `json:"bytesPerSecond,omitempty"`
}
type cliEvaluation struct {
	Version             int             `json:"version"`
	Passed              bool            `json:"passed"`
	Workers             int             `json:"workers"`
	ConfigurationSHA256 string          `json:"configurationSha256"`
	Environment         map[string]any  `json:"environment"`
	CacheMode           string          `json:"cacheMode"`
	Cases               []cliCaseReport `json:"cases"`
}

func cliEvaluate(ctx context.Context, command string, o cliOptions, stdin io.Reader) (any, error) {
	if o.manifest == "" {
		return nil, cliFail(2, "--manifest is required")
	}
	if o.iterations < 1 || o.iterations > 10000 || o.warmup < 0 || o.warmup > 1000 {
		return nil, cliFail(2, "iterations must be 1-10000 and warmup 0-1000")
	}
	if math.IsNaN(o.maxRegression) || math.IsInf(o.maxRegression, 0) || o.maxRegression < 0 {
		return nil, cliFail(2, "max-regression must be a finite nonnegative percentage")
	}
	var corpus cliCorpus
	if err := cliReadJSON(ctx, o.manifest, stdin, o.maxBytes, &corpus); err != nil {
		return nil, err
	}
	if corpus.Version != 1 || len(corpus.Cases) == 0 {
		return nil, cliFail(3, "manifest needs version 1 and at least one case")
	}
	root := filepath.Dir(o.manifest)
	config, _ := json.Marshal(map[string]any{"weights": GetAttributeWeightsSnapshot(), "roles": roleSpecificOverallWeights, "scaled": GetUseScaledRatings()})
	report := cliEvaluation{Version: 1, Passed: true, Workers: o.workers, ConfigurationSHA256: cliHash(config), Environment: cliBuildInfo(), CacheMode: "warm-process; fresh result per iteration; OS file cache uncontrolled", Cases: []cliCaseReport{}}
	bench := command == "parse bench"
	var cpu bytes.Buffer
	if bench && o.cpuProfile != "" {
		if err := pprof.StartCPUProfile(&cpu); err != nil {
			return nil, err
		}
		defer pprof.StopCPUProfile()
	}
	names := map[string]bool{}
	for _, c := range corpus.Cases {
		if err := ctx.Err(); err != nil {
			return report, err
		}
		if c.Name == "" || names[c.Name] || c.ExpectedCount == nil || *c.ExpectedCount < 0 {
			return report, cliFail(3, "every corpus case requires a unique name and nonnegative expectedCount")
		}
		names[c.Name] = true
		if c.Format == "" {
			var err error
			c.Format, err = cliInputFormat(c.Path, "")
			if err != nil {
				return report, err
			}
		}
		if c.Stage == "" {
			c.Stage = o.stage
		}
		entry := cliCaseReport{Name: c.Name, Format: c.Format, Stage: c.Stage, Passed: true, Failures: []string{}}
		var semanticBaseline any
		runs := 1
		if bench {
			runs = o.warmup + o.iterations
		}
		for iteration := 0; iteration < runs; iteration++ {
			start := time.Now()
			var before, after runtime.MemStats
			runtime.ReadMemStats(&before)
			fixturePath := c.Path
			if !filepath.IsAbs(fixturePath) {
				fixturePath = filepath.Join(root, fixturePath)
			}
			data, err := cliReadInput(ctx, fixturePath, stdin, o.maxBytes)
			if err != nil {
				return report, err
			}
			readMS := float64(time.Since(start).Nanoseconds()) / 1e6
			entry.InputSHA256 = cliHash(data)
			entry.InputBytes = len(data)
			if c.SHA256 != "" && c.SHA256 != entry.InputSHA256 {
				return report, cliFail(4, "fixture %s hash differs from manifest", c.Name)
			}
			var parsed any
			var timings ParseTimings
			var rejected, parseErrors int
			var diagnosticCodes []string
			if c.Format == "fm" {
				p, e := InspectFMSave(ctx, bytes.NewReader(data))
				parsed, err = p, e
				if p != nil {
					entry.Count = len(p.Fields)
				}
			} else {
				p, e := ParsePlayers(ctx, bytes.NewReader(data), ParseOptions{c.Format, c.Stage, o.workers, false})
				parsed, err = p, e
				if p != nil {
					entry.Count = len(p.Players)
					rejected = p.Counts.Rejected
					parseErrors = p.Counts.Errors
					timings = p.Timings
					for _, d := range p.Diagnostics {
						diagnosticCodes = append(diagnosticCodes, d.Code)
					}
				}
			}
			if err != nil {
				return report, cliFail(3, "case %s: %v", c.Name, err)
			}
			serializeStart := time.Now()
			serialized, err := json.Marshal(parsed)
			if err != nil {
				return report, err
			}
			serializeMS := float64(time.Since(serializeStart).Nanoseconds()) / 1e6
			runtime.ReadMemStats(&after)
			sample := cliBenchSample{TotalMS: float64(max(time.Since(start).Nanoseconds(), 1)) / 1e6, ReadMS: readMS, SerializeMS: serializeMS, Timings: timings, AllocatedBytes: after.TotalAlloc - before.TotalAlloc, Allocations: after.Mallocs - before.Mallocs, HeapBytes: after.HeapAlloc}
			var raw any
			if err = cliDecodeJSON(serialized, &raw); err != nil {
				return report, err
			}
			semantic, err := cliSemantic(raw)
			if err != nil {
				return report, err
			}
			failures := []string{}
			if entry.Count != *c.ExpectedCount {
				failures = append(failures, fmt.Sprintf("expected %d records, got %d", *c.ExpectedCount, entry.Count))
			}
			if rejected != c.ExpectedRejected {
				failures = append(failures, fmt.Sprintf("expected %d rejected rows, got %d", c.ExpectedRejected, rejected))
			}
			if parseErrors != c.ExpectedErrors {
				failures = append(failures, fmt.Sprintf("expected %d parse errors, got %d", c.ExpectedErrors, parseErrors))
			}
			if c.ExpectedDiagnostics != nil {
				expected := append([]string{}, c.ExpectedDiagnostics...)
				actual := append([]string{}, diagnosticCodes...)
				sort.Strings(expected)
				sort.Strings(actual)
				if !reflect.DeepEqual(expected, actual) {
					failures = append(failures, "diagnostic codes differ")
				}
			}
			for _, expect := range c.Expected {
				actual, ok := cliPointer(raw, expect.Path)
				if !ok || !reflect.DeepEqual(expect.Value, actual) {
					failures = append(failures, "expectation failed: "+expect.Path)
				}
			}
			if c.Golden != "" {
				var golden any
				goldenPath := c.Golden
				if !filepath.IsAbs(goldenPath) {
					goldenPath = filepath.Join(root, goldenPath)
				}
				if err = cliReadJSON(ctx, goldenPath, stdin, o.maxBytes, &golden); err != nil {
					return report, err
				}
				gs, e := cliSemantic(golden)
				if e != nil {
					return report, e
				}
				if len(cliDifferences(gs, semantic)) > 0 {
					failures = append(failures, "golden output differs")
				}
			}
			if semanticBaseline == nil {
				semanticBaseline = semantic
			} else if len(cliDifferences(semanticBaseline, semantic)) > 0 {
				failures = append(failures, "semantic output changed between iterations")
			}
			if len(failures) > 0 {
				entry.Passed = false
				entry.Failures = failures
				break
			}
			if bench && iteration >= o.warmup {
				entry.Samples = append(entry.Samples, sample)
			}
		}
		if len(entry.Samples) > 0 {
			values := make([]float64, len(entry.Samples))
			for i, s := range entry.Samples {
				values[i] = s.TotalMS
			}
			sort.Float64s(values)
			middle := len(values) / 2
			entry.MedianMS = values[middle]
			if len(values)%2 == 0 {
				entry.MedianMS = (values[middle-1] + values[middle]) / 2
			}
			entry.P95MS = values[int(math.Ceil(float64(len(values))*.95))-1]
			entry.RowsPerSecond = float64(entry.Count) * 1000 / entry.MedianMS
			entry.BytesPerSecond = float64(entry.InputBytes) * 1000 / entry.MedianMS
		}
		report.Passed = report.Passed && entry.Passed
		report.Cases = append(report.Cases, entry)
	}
	if bench && o.cpuProfile != "" {
		pprof.StopCPUProfile()
		if err := cliWriteFile(o.cpuProfile, cpu.Bytes(), o.overwrite); err != nil {
			return report, cliFail(6, "CPU profile: %v", err)
		}
	}
	if bench && o.heapProfile != "" {
		var heap bytes.Buffer
		if err := pprof.WriteHeapProfile(&heap); err != nil {
			return report, err
		}
		if err := cliWriteFile(o.heapProfile, heap.Bytes(), o.overwrite); err != nil {
			return report, cliFail(6, "heap profile: %v", err)
		}
	}
	if bench && o.baseline != "" {
		var envelope struct {
			Data cliEvaluation `json:"data"`
		}
		data, err := os.ReadFile(o.baseline)
		if err != nil {
			return report, err
		}
		if err = json.Unmarshal(data, &envelope); err != nil {
			return report, err
		}
		baseline := envelope.Data
		if !baseline.Passed || baseline.ConfigurationSHA256 != report.ConfigurationSHA256 || baseline.Workers != report.Workers || baseline.CacheMode != report.CacheMode || len(baseline.Cases) != len(report.Cases) {
			return report, cliFail(2, "benchmark baseline is unsuccessful or incompatible")
		}
		for _, key := range []string{"go", "os", "arch", "cpus", "gomaxprocs", "cpuIdentifier", "gogc", "gomemlimit"} {
			if fmt.Sprint(baseline.Environment[key]) != fmt.Sprint(report.Environment[key]) {
				return report, cliFail(2, "benchmark environment differs: %s", key)
			}
		}
		for i := range report.Cases {
			a, b := &report.Cases[i], baseline.Cases[i]
			if a.Name != b.Name || a.InputSHA256 != b.InputSHA256 || a.Stage != b.Stage || a.Format != b.Format || b.MedianMS <= 0 {
				return report, cliFail(2, "benchmark case baseline is incompatible: %s", a.Name)
			}
			if a.MedianMS > b.MedianMS*(1+o.maxRegression/100) {
				a.Passed = false
				a.Failures = append(a.Failures, "median time exceeds allowed regression")
				report.Passed = false
			}
		}
	}
	if !report.Passed {
		return report, cliFail(4, "corpus evaluation failed")
	}
	return report, nil
}
