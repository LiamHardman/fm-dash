package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

// roleWeightsJSON is a verbatim copy of
// src/api/public/role_specific_overall_weights.json (see role_weights.json
// in this directory). Keep the two files in sync manually -- there is no
// build-time link between the two Go modules.
//
//go:embed role_weights.json
var roleWeightsJSON []byte

// roleStylePosition returns the text before the FIRST " - " in roleName,
// trimmed. Mirrors src/api/positions.go's GetShortPositionKeyFromRoleName.
func roleStylePosition(roleName string) string {
	parts := strings.SplitN(roleName, " - ", 2)
	if len(parts) > 0 {
		return strings.TrimSpace(parts[0])
	}
	return ""
}

// roleStyleCategory returns roleName with its trailing " - <Duty>" removed.
// Mirrors src/api/calculations.go's roleStyleCategory exactly: strips
// everything after the LAST " - "; falls back to the full name if no " - "
// is present at all (never hit by the real 143 roles, kept for parity).
func roleStyleCategory(roleName string) string {
	idx := strings.LastIndex(roleName, " - ")
	if idx == -1 {
		return roleName
	}
	return roleName[:idx]
}

// parseEmbeddedRoleWeights parses the embedded role_weights.json copy into
// map[roleName]map[attributeCode]weight, matching the shape src/api's
// config.go loads role_specific_overall_weights.json into.
func parseEmbeddedRoleWeights() (map[string]map[string]int, error) {
	var parsed map[string]map[string]int
	if err := json.Unmarshal(roleWeightsJSON, &parsed); err != nil {
		return nil, fmt.Errorf("parsing embedded role_weights.json: %w", err)
	}
	return parsed, nil
}
