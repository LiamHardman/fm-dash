package main

import (
	"context"
	"io"
	"math"
)

type SigningCriteria struct {
	Club             string   `json:"club"`
	Positions        []string `json:"positions"`
	MinOverall       *int     `json:"minOverall,omitempty"`
	MaxOverall       *int     `json:"maxOverall,omitempty"`
	MaxTransferValue *int64   `json:"maxTransferValue,omitempty"`
}

// RecommendSignings implements RecommendSigningDialog's squad-relative defaults.
func RecommendSignings(players []Player, criteria SigningCriteria) (any, error) {
	if len(criteria.Positions) == 0 || criteria.Club == "" {
		return nil, cliFail(3, "club and at least one position are required")
	}
	squad := filterByClub(players, criteria.Club)
	names := map[string]bool{}
	match := func(p Player) bool {
		for _, pos := range p.ShortPositions {
			for _, wanted := range criteria.Positions {
				if pos == wanted {
					return true
				}
			}
		}
		return false
	}
	lowest, total, count, maxValue := 101, 0, 0, int64(0)
	for _, p := range squad {
		names[p.Name] = true
		maxValue = max(maxValue, p.TransferValueAmount)
		if p.Overall > 0 {
			total += p.Overall
			count++
			if match(p) {
				lowest = min(lowest, p.Overall)
			}
		}
	}
	average := 60
	if count > 0 {
		average = int(math.Floor(float64(total)/float64(count) + .5))
	}
	minimum := max(0, average-10)
	if lowest <= 100 {
		minimum = max(0, lowest-3) / 5 * 5
	}
	maximum := 100
	if criteria.MinOverall != nil {
		minimum = *criteria.MinOverall
	}
	if criteria.MaxOverall != nil {
		maximum = *criteria.MaxOverall
	}
	if criteria.MaxTransferValue != nil {
		maxValue = *criteria.MaxTransferValue
	}
	if minimum < 0 || maximum < minimum || maxValue < 0 {
		return nil, cliFail(3, "invalid rating or transfer value bounds")
	}
	result := []Player{}
	for _, p := range players {
		if !match(p) || names[p.Name] || p.Overall < minimum || p.Overall > maximum || p.TransferValueAmount > maxValue {
			continue
		}
		result = append(result, p)
	}
	return map[string]any{"players": result, "criteria": map[string]any{"club": criteria.Club, "positions": criteria.Positions, "minOverall": minimum, "maxOverall": maximum, "maxTransferValue": maxValue}}, nil
}

func cliRecommend(ctx context.Context, players []Player, o cliOptions, stdin io.Reader) (any, error) {
	data, err := cliBody(ctx, o, stdin)
	if err != nil {
		return nil, err
	}
	var criteria SigningCriteria
	if err = cliDecodeJSON(data, &criteria); err != nil {
		return nil, err
	}
	return RecommendSignings(players, criteria)
}
