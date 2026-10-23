package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"
)

func cliJSONLines(e cliEnvelope) ([]byte, error) {
	var out bytes.Buffer
	encoder := json.NewEncoder(&out)
	if p, ok := e.Data.(*ParsedPlayers); ok && p != nil {
		header := map[string]any{"type": "header", "schemaVersion": e.SchemaVersion, "command": e.Command, "headers": p.Headers, "options": p.Options, "provenance": e.Provenance}
		if err := encoder.Encode(header); err != nil {
			return nil, err
		}
		for i, player := range p.Players {
			if err := encoder.Encode(map[string]any{"type": "player", "sourceRow": p.SourceRows[i], "player": player, "rawAttributes": p.RawAttributes[i]}); err != nil {
				return nil, err
			}
		}
		if err := encoder.Encode(map[string]any{"type": "summary", "status": e.Status, "counts": p.Counts, "currencySymbol": p.Currency, "timings": p.Timings, "diagnostics": e.Diagnostics, "diagnosticsTruncated": p.DiagnosticsTruncated}); err != nil {
			return nil, err
		}
	} else {
		if err := encoder.Encode(e); err != nil {
			return nil, err
		}
	}
	return bytes.TrimSuffix(out.Bytes(), []byte("\n")), nil
}

func cliTable(e cliEnvelope) ([]byte, error) {
	encoded, err := json.Marshal(e.Data)
	if err != nil {
		return nil, err
	}
	var value any
	if err = json.Unmarshal(encoded, &value); err != nil {
		return nil, err
	}
	if m, ok := value.(map[string]any); ok {
		if rows, exists := m["players"]; exists {
			value = rows
		}
	}
	var output bytes.Buffer
	w := tabwriter.NewWriter(&output, 0, 4, 2, ' ', 0)
	_, _ = fmt.Fprintf(w, "%s\t%s\n", e.Command, e.Status)
	if rows, ok := value.([]any); ok {
		keys := map[string]bool{}
		for _, row := range rows {
			if m, ok := row.(map[string]any); ok {
				for key, v := range m {
					switch v.(type) {
					case nil, string, float64, bool:
						keys[key] = true
					}
				}
			}
		}
		columns := cliSortedKeys(keys)
		_, _ = fmt.Fprintln(w, strings.Join(columns, "\t"))
		for _, row := range rows {
			m, _ := row.(map[string]any)
			cells := []string{}
			for _, key := range columns {
				cells = append(cells, strings.NewReplacer("\t", " ", "\n", " ", "\r", " ").Replace(fmt.Sprint(m[key])))
			}
			_, _ = fmt.Fprintln(w, strings.Join(cells, "\t"))
		}
	} else if m, ok := value.(map[string]any); ok {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			v, _ := json.Marshal(m[key])
			_, _ = fmt.Fprintf(w, "%s\t%s\n", key, v)
		}
	} else {
		_, _ = fmt.Fprintln(w, string(encoded))
	}
	for _, d := range e.Diagnostics {
		_, _ = fmt.Fprintf(w, "%s\t%s: %s\n", d.Severity, d.Code, d.Message)
	}
	if err = w.Flush(); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(output.Bytes(), []byte("\n")), nil
}
