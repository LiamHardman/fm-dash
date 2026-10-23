package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"
)

// cliExport applies the same category switches as ExportOptionsDialog. With no
// explicit switches the existing complete backend export remains available.
func cliExport(ctx context.Context, players []Player, d *cliDataset, o cliOptions, stdin io.Reader) (any, error) {
	if o.body == "" {
		if o.format == "csv" {
			return []byte(generateCSVContent(players)), nil
		}
		var result any
		err := json.Unmarshal(generateJSONContent(players, d.Result.Currency, d.ID), &result)
		return result, err
	}
	data, err := cliBody(ctx, o, stdin)
	if err != nil {
		return nil, err
	}
	var options struct {
		Options map[string]bool `json:"options"`
		Preset  string          `json:"preset,omitempty"`
		Format  string          `json:"format,omitempty"`
	}
	if err = cliDecodeJSON(data, &options); err != nil {
		return nil, err
	}
	allowed := map[string]bool{"basicInfo": true, "fifahStats": true, "fmStats": true, "roleRatings": true, "performancePercentiles": true, "contractInfo": true, "personalInfo": true}
	enabled := false
	for key, on := range options.Options {
		if !allowed[key] {
			return nil, cliFail(2, "unknown export category %q", key)
		}
		enabled = enabled || on
	}
	if !enabled {
		return nil, cliFail(2, "select at least one export category")
	}
	rows := []map[string]any{}
	for _, p := range players {
		row := map[string]any{}
		if options.Options["basicInfo"] {
			row = map[string]any{"name": p.Name, "age": p.Age, "nationality": p.Nationality, "nationalityIso": p.NationalityISO, "nationality_fifa_code": p.NationalityFIFACode, "club": p.Club, "division": p.Division, "position": p.Position, "shortPositions": p.ShortPositions, "parsedPositions": p.ParsedPositions, "positionGroups": p.PositionGroups}
		}
		if options.Options["fifahStats"] {
			for key, value := range map[string]int{"Overall": p.Overall, "PAC": p.PAC, "SHO": p.SHO, "PAS": p.PAS, "DRI": p.DRI, "DEF": p.DEF, "PHY": p.PHY, "GK": p.GK} {
				row[key] = value
			}
		}
		if options.Options["fmStats"] {
			row["attributes"] = p.Attributes
		}
		if options.Options["roleRatings"] {
			row["roleSpecificOveralls"] = p.RoleSpecificOveralls
		}
		if options.Options["performancePercentiles"] {
			row["performancePercentiles"] = p.PerformancePercentiles
		}
		if options.Options["contractInfo"] {
			row["transferValue"] = p.TransferValue
			row["transferValueAmount"] = p.TransferValueAmount
			row["wage"] = p.Wage
			row["wageAmount"] = p.WageAmount
		}
		if options.Options["personalInfo"] {
			row["personality"] = p.Personality
			row["media_handling"] = p.MediaHandling
			row["attributeMasked"] = p.AttributeMasked
			for _, key := range []string{"foot", "height", "weight"} {
				if value, ok := p.Attributes[key]; ok {
					row[key] = value
				}
			}
		}
		rows = append(rows, row)
	}
	if o.format != "csv" {
		return map[string]any{"metadata": map[string]any{"exportDate": time.Now().UTC(), "datasetId": d.ID, "currencySymbol": d.Result.Currency, "totalPlayers": len(rows), "categories": options.Options}, "players": rows}, nil
	}
	// Nested values are JSON in individual CSV cells, preserving every selected
	// value and avoiding lossy stringification of maps or role arrays.
	columns := map[string]bool{}
	for _, row := range rows {
		for key := range row {
			columns[key] = true
		}
	}
	keys := cliSortedKeys(columns)
	sort.Strings(keys)
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	if err = writer.Write(keys); err != nil {
		return nil, err
	}
	for _, row := range rows {
		record := []string{}
		for _, key := range keys {
			value := row[key]
			switch v := value.(type) {
			case string:
				record = append(record, v)
			case nil:
				record = append(record, "")
			case int, int64, float64, bool:
				record = append(record, fmt.Sprint(v))
			default:
				encoded, e := json.Marshal(v)
				if e != nil {
					return nil, e
				}
				record = append(record, string(encoded))
			}
		}
		if err = writer.Write(record); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	return buffer.Bytes(), writer.Error()
}
