package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"golang.org/x/net/html"
)

// ParseOptions describes computation, independently of HTTP, storage and the GUI.
type ParseOptions struct {
	Format  string `json:"format"`
	Stage   string `json:"stage"`
	Workers int    `json:"workers"`
	Strict  bool   `json:"strict"`
}

type ParseDiagnostic struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Row      int    `json:"row,omitempty"`
	Column   string `json:"column,omitempty"`
	Message  string `json:"message"`
}

type ParseCounts struct {
	Seen     int `json:"seen"`
	Accepted int `json:"accepted"`
	Rejected int `json:"rejected"`
	Warnings int `json:"warnings"`
	Errors   int `json:"errors"`
}

type ParseTimings struct {
	DecodeMS  float64 `json:"decodeMs"`
	EnhanceMS float64 `json:"enhanceMs"`
	DatasetMS float64 `json:"datasetMs"`
	TotalMS   float64 `json:"totalMs"`
}

type ParsedPlayers struct {
	Headers              []string            `json:"headers"`
	Players              []Player            `json:"players"`
	SourceRows           []int               `json:"sourceRows"`
	RawAttributes        []map[string]string `json:"rawAttributes"`
	Currency             string              `json:"currencySymbol"`
	Options              ParseOptions        `json:"options"`
	Counts               ParseCounts         `json:"counts"`
	Diagnostics          []ParseDiagnostic   `json:"diagnostics"`
	DiagnosticsTruncated bool                `json:"diagnosticsTruncated"`
	Timings              ParseTimings        `json:"timings"`
}

var errStrictParse = errors.New("strict parsing rejected input; inspect diagnostics")

func (p *ParsedPlayers) diagnose(code, severity string, row int, column, message string) {
	if len(message) > 1024 {
		runes := []rune(message)
		message = string(runes[:min(len(runes), 256)]) + "…"
	}
	if severity == "error" {
		p.Counts.Errors++
	} else {
		p.Counts.Warnings++
	}
	if len(p.Diagnostics) < 200 {
		p.Diagnostics = append(p.Diagnostics, ParseDiagnostic{code, severity, row, column, message})
	} else {
		p.DiagnosticsTruncated = true
	}
}

// contextReader cooperates with bounded local reads and decompression. A caller
// supplying a blocking network reader is responsible for setting its deadline.
type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}

// scanExport is the sole CSV/HTML row decoder used by both transports and legacy
// parser entry points. emit receives immutable cells and one-based table rows.
func scanExport(ctx context.Context, input io.Reader, format string, emit func(int, []string, bool) error) error {
	r := bufio.NewReader(contextReader{ctx, input})
	if prefix, _ := r.Peek(3); string(prefix) == "\xef\xbb\xbf" {
		_, _ = r.Discard(3)
	}
	if format == "csv" {
		decoder := csv.NewReader(r)
		decoder.Comma = ';'
		decoder.FieldsPerRecord = -1
		row := 0
		for {
			cells, err := decoder.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("CSV record %d: %w", row+1, err)
			}
			if len(cells) == 1 && strings.TrimSpace(cells[0]) == "" {
				continue
			}
			row++
			if err = emit(row, cells, row == 1); err != nil {
				return err
			}
		}
		if row == 0 {
			return errors.New("no table headers found")
		}
		return nil
	}
	if format != "html" {
		return fmt.Errorf("unsupported player export format %q", format)
	}
	z := html.NewTokenizer(r)
	tableDepth, row := 0, 0
	inRow, inCell, headerSeen := false, false, false
	var cells []string
	var cell strings.Builder
	flushCell := func() {
		if inCell {
			cells = append(cells, strings.TrimSpace(cell.String()))
			inCell = false
			cell.Reset()
		}
	}
	flushRow := func() error {
		flushCell()
		if !inRow {
			return nil
		}
		inRow = false
		if len(cells) == 0 {
			return nil
		}
		row++
		err := emit(row, cells, !headerSeen)
		headerSeen = true
		return err
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		kind := z.Next()
		token := z.Token()
		switch kind {
		case html.ErrorToken:
			if z.Err() != io.EOF {
				return z.Err()
			}
			if err := flushRow(); err != nil {
				return err
			}
			if !headerSeen {
				return errors.New("no table headers found")
			}
			return nil
		case html.StartTagToken:
			switch token.Data {
			case "table":
				tableDepth++
			case "tr":
				if tableDepth == 1 {
					if err := flushRow(); err != nil {
						return err
					}
					cells = nil
					inRow = true
				}
			case "td", "th":
				if inRow && tableDepth == 1 {
					flushCell()
					inCell = true
				}
			case "br":
				if inCell {
					cell.WriteByte(' ')
				}
			}
		case html.TextToken:
			if inCell {
				cell.WriteString(token.Data)
			}
		case html.EndTagToken:
			switch token.Data {
			case "td", "th":
				if tableDepth == 1 {
					flushCell()
				}
			case "tr":
				if tableDepth == 1 {
					if err := flushRow(); err != nil {
						return err
					}
				}
			case "table":
				if tableDepth == 1 {
					if err := flushRow(); err != nil {
						return err
					}
					if headerSeen {
						return nil
					}
				}
				if tableDepth > 0 {
					tableDepth--
				}
			}
		}
	}
}

// ParsePlayers performs a complete, deterministic parse without persisting a
// Dataset. Result slices are aligned by source row, even with multiple workers.
func ParsePlayers(ctx context.Context, input io.Reader, options ParseOptions) (*ParsedPlayers, error) {
	started := time.Now()
	if options.Stage == "" {
		options.Stage = "complete"
	}
	if options.Workers < 1 || options.Workers > 256 {
		return nil, errors.New("workers must be between 1 and 256")
	}
	if options.Stage != "parsed" && options.Stage != "enhanced" && options.Stage != "complete" {
		return nil, errors.New("stage must be parsed, enhanced or complete")
	}
	result := &ParsedPlayers{Options: options, Players: []Player{}, SourceRows: []int{}, RawAttributes: []map[string]string{}, Diagnostics: []ParseDiagnostic{}, Currency: "$"}
	type job struct {
		row   int
		cells []string
	}
	type decoded struct {
		row    int
		player Player
		err    error
		cells  int
	}
	jobs, outputs := make(chan job, options.Workers*2), make(chan decoded, options.Workers*2)
	var wg sync.WaitGroup
	var records []decoded
	collected := make(chan struct{})
	go func() {
		defer close(collected)
		for v := range outputs {
			records = append(records, v)
		}
	}()
	for i := 0; i < options.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				if ctx.Err() != nil {
					continue
				}
				p, err := parseCellsToPlayer(j.cells, result.Headers)
				outputs <- decoded{j.row, p, err, len(j.cells)}
			}
		}()
	}
	err := scanExport(ctx, input, options.Format, func(row int, cells []string, header bool) error {
		if header {
			result.Headers = normalizeCSVHeaders(cells)
			return nil
		}
		result.Counts.Seen++
		select {
		case jobs <- job{row, cells}:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
	close(jobs)
	wg.Wait()
	close(outputs)
	<-collected
	result.Timings.DecodeMS = float64(time.Since(started).Nanoseconds()) / 1e6
	if err != nil {
		return result, err
	}
	if err = ctx.Err(); err != nil {
		return result, err
	}
	seenHeaders := make(map[string]bool)
	for _, h := range result.Headers {
		if h == "Nat" && seenHeaders[h] {
			result.diagnose("ambiguous_nat_header", "warning", 1, h, "Nat occurs twice; values distinguish nationality from Natural Fitness")
			continue
		}
		if h == "" || seenHeaders[h] {
			result.diagnose("duplicate_or_empty_header", "error", 1, h, "header must be nonempty and unique after normalization")
		}
		seenHeaders[h] = true
	}
	for _, required := range []string{"Name", "Position", "Age"} {
		if !seenHeaders[required] {
			result.diagnose("missing_header", "error", 1, required, "required player column is missing")
		}
	}
	sort.Slice(records, func(i, j int) bool { return records[i].row < records[j].row })
	ids, currencies := make(map[int64]bool), make(map[string]bool)
	for _, v := range records {
		if v.err != nil {
			result.Counts.Rejected++
			result.diagnose("rejected_row", "error", v.row, "", v.err.Error())
			continue
		}
		if v.cells != len(result.Headers) {
			result.diagnose("column_count", "error", v.row, "", "cell count differs from header count")
		}
		if !utf8.ValidString(v.player.Name) {
			result.diagnose("invalid_utf8", "error", v.row, "Name", "name is not valid UTF-8")
		}
		if age, e := strconv.Atoi(v.player.Age); e != nil || age < 1 || age > 120 {
			result.diagnose("invalid_age", "error", v.row, "Age", "age must be an integer between 1 and 120")
		}
		if v.player.Position == "" {
			result.diagnose("missing_position", "error", v.row, "Position", "position is empty")
		}
		if v.player.UID <= 0 || ids[v.player.UID] {
			result.diagnose("invalid_player_reference", "warning", v.row, "UID", "missing, invalid or duplicate UID; use source row identity")
		}
		ids[v.player.UID] = true
		for _, value := range []string{v.player.TransferValue, v.player.Wage} {
			_, _, symbol := ParseMonetaryValueGo(value)
			if symbol != "" {
				if len(currencies) == 0 {
					result.Currency = symbol
				}
				currencies[symbol] = true
			}
		}
		result.Players = append(result.Players, v.player)
		result.RawAttributes = append(result.RawAttributes, v.player.Attributes)
		result.SourceRows = append(result.SourceRows, v.row)
	}
	result.Counts.Accepted = len(result.Players)
	if len(currencies) == 0 {
		result.diagnose("currency_fallback", "warning", 0, "", "no currency detected; defaulting to $")
	}
	if len(currencies) > 1 {
		result.diagnose("mixed_currencies", "warning", 0, "", "multiple currencies found; no conversion performed")
	}
	if options.Stage != "parsed" {
		enhanceStart := time.Now()
		enhanceErrors := make([]error, len(result.Players))
		next := make(chan int)
		for i := 0; i < options.Workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for index := range next {
					if ctx.Err() == nil {
						enhanceErrors[index] = enhancePlayerWithCalculations(&result.Players[index])
					}
				}
			}()
		}
		for index := range result.Players {
			if ctx.Err() != nil {
				break
			}
			next <- index
		}
		close(next)
		wg.Wait()
		keptPlayers, keptRows, keptAttributes := result.Players[:0], result.SourceRows[:0], result.RawAttributes[:0]
		for i, err := range enhanceErrors {
			if err != nil {
				result.Counts.Accepted--
				result.Counts.Rejected++
				result.diagnose("calculation_failed", "error", result.SourceRows[i], "", err.Error())
				continue
			}
			keptPlayers = append(keptPlayers, result.Players[i])
			keptRows = append(keptRows, result.SourceRows[i])
			keptAttributes = append(keptAttributes, result.RawAttributes[i])
		}
		result.Players, result.SourceRows, result.RawAttributes = keptPlayers, keptRows, keptAttributes
		result.Timings.EnhanceMS = float64(time.Since(enhanceStart).Nanoseconds()) / 1e6
	}
	if err = ctx.Err(); err != nil {
		return result, err
	}
	if options.Stage == "complete" {
		datasetStart := time.Now()
		NormalizeMBRScoresRelativeToMax(result.Players)
		if err := CalculatePlayerPerformancePercentilesContext(ctx, result.Players); err != nil {
			return result, err
		}
		result.Timings.DatasetMS = float64(time.Since(datasetStart).Nanoseconds()) / 1e6
	}
	result.Timings.TotalMS = float64(time.Since(started).Nanoseconds()) / 1e6
	if err = ctx.Err(); err != nil {
		return result, err
	}
	if options.Strict && (result.Counts.Errors > 0 || result.Counts.Rejected > 0) {
		return result, errStrictParse
	}
	return result, nil
}
