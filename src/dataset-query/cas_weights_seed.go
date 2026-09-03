package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

// casWeightsJSON is the map[position]map[attribute]weight extraction of
// casPositionWeights/casAttrOrder from src/api/ca_calculation.go, produced
// by a one-off extraction test run against that Go literal (dump_cas_weights_test.go,
// run once via `go test -run TestDumpCASWeights` and then deleted -- never
// committed). Zero-weight (position, attribute) pairs are omitted --
// calculateCASForPosition excludes them entirely, they are NOT "floored to
// 1" despite a misleading comment in ca_calculation.go. Unlike
// role_weights.json, there is no independent JSON source of truth for this
// data; keep this file in sync manually by re-running the extraction
// program if casPositionWeights/casAttrOrder ever changes.
//
//go:embed cas_weights.json
var casWeightsJSON []byte

// parseEmbeddedCasWeights parses the embedded cas_weights.json copy into
// map[position]map[attribute]weight.
func parseEmbeddedCasWeights() (map[string]map[string]float64, error) {
	var parsed map[string]map[string]float64
	if err := json.Unmarshal(casWeightsJSON, &parsed); err != nil {
		return nil, fmt.Errorf("parsing embedded cas_weights.json: %w", err)
	}
	return parsed, nil
}
