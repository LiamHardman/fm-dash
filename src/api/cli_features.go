package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

//go:embed public/cli_weight_presets.json
var cliWeightPresets []byte

type cliWeightProfile struct {
	ID          string                    `json:"id"`
	Name        string                    `json:"name"`
	Description string                    `json:"description,omitempty"`
	Weights     map[string]map[string]int `json:"weights"`
}

func cliWeights(ctx context.Context, command string, o cliOptions, stdin io.Reader, w *cliWorkspace) (any, error) {
	var presets []cliWeightProfile
	if err := json.Unmarshal(cliWeightPresets, &presets); err != nil {
		return nil, err
	}
	type library struct {
		Version  int                `json:"version"`
		Profiles []cliWeightProfile `json:"profiles"`
		ActiveID string             `json:"activeId"`
	}
	var output any
	err := w.mutate(ctx, func() error {
		current := library{Version: 1, Profiles: []cliWeightProfile{}}
		path := filepath.Join(w.dir, "weights.json")
		data, err := os.ReadFile(path)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if err == nil {
			if err = cliDecodeJSON(data, &current); err != nil {
				return err
			}
		}
		if command == "weights list" {
			output = map[string]any{"custom": current, "presets": presets}
			return nil
		}
		if command == "weights set" {
			data, err := cliBody(ctx, o, stdin)
			if err != nil {
				return err
			}
			var p cliWeightProfile
			if err = cliDecodeJSON(data, &p); err != nil {
				return err
			}
			if p.ID == "" {
				p.ID = uuid.NewString()
			}
			if strings.TrimSpace(p.Name) == "" || len(p.Name) > 100 || strings.HasPrefix(p.ID, "preset-") {
				return cliFail(3, "profile requires a name and a non-preset ID")
			}
			if err = SetAttributeWeights(p.Weights); err != nil {
				return cliFail(3, "weights: %v", err)
			}
			found := false
			for i := range current.Profiles {
				if current.Profiles[i].ID == p.ID {
					current.Profiles[i] = p
					found = true
				}
			}
			if !found {
				current.Profiles = append(current.Profiles, p)
			}
			if current.ActiveID == p.ID {
				if err = cliPersistCalculationConfig(w); err != nil {
					return err
				}
			}
			output = p
		} else {
			if err = cliRequired(o.id, "--id"); err != nil {
				return err
			}
			if command == "weights delete" {
				kept := []cliWeightProfile{}
				found := false
				for _, p := range current.Profiles {
					if p.ID == o.id {
						found = true
					} else {
						kept = append(kept, p)
					}
				}
				if !found {
					return cliFail(5, "custom weight profile is unavailable")
				}
				current.Profiles = kept
				if current.ActiveID == o.id {
					return cliFail(2, "activate defaults or another profile before deleting the active profile")
				}
				output = map[string]any{"deleted": o.id}
			} else {
				var weights map[string]map[string]int
				if o.id == "default" {
					_ = json.Unmarshal(mustCLIDefault("public/attribute_weights.json"), &weights)
				} else {
					for _, p := range append(presets, current.Profiles...) {
						if p.ID == o.id {
							weights = p.Weights
						}
					}
				}
				if weights == nil {
					return cliFail(5, "weight profile is unavailable")
				}
				if err = SetAttributeWeights(weights); err != nil {
					return err
				}
				current.ActiveID = o.id
				if err = cliPersistCalculationConfig(w); err != nil {
					return err
				}
				output = map[string]any{"activeId": o.id, "weights": GetAttributeWeightsSnapshot()}
			}
		}
		data, err = json.Marshal(current)
		if err != nil {
			return err
		}
		return atomicConfigWrite(path, data)
	})
	return output, err
}

func mustCLIDefault(name string) []byte { data, _ := cliDefaults.ReadFile(name); return data }
func cliPersistCalculationConfig(w *cliWorkspace) error {
	data, err := json.Marshal(map[string]any{"attributeWeights": GetAttributeWeightsSnapshot(), "useScaledRatings": GetUseScaledRatings()})
	if err != nil {
		return err
	}
	return atomicConfigWrite(filepath.Join(w.dir, "config.json"), data)
}

func cliPreset(name string, players []Player) (map[string]any, error) {
	switch name {
	case "":
		return map[string]any{}, nil
	case "first-team-ready":
		return map[string]any{"ageRange": map[string]any{"min": 23, "max": 30}, "minOverall": 75}, nil
	case "high-potential-wonderkid":
		return map[string]any{"ageRange": map[string]any{"min": 15, "max": 21}, "minOverall": 75}, nil
	case "affordable-target":
		maximum := int64(0)
		for _, p := range players {
			maximum = max(maximum, p.TransferValueAmount)
		}
		if maximum == 0 {
			maximum = 100000000
		}
		return map[string]any{"transferValueRangeLocal": map[string]any{"min": 0, "max": max(int64(1000000), (maximum+2)/4), "userSet": true}}, nil
	case "free-agent-depth":
		return map[string]any{"transferValueRangeLocal": map[string]any{"min": 0, "max": 0, "userSet": true}}, nil
	default:
		return nil, cliFail(2, "unknown Starter preset %q", name)
	}
}

func cliSavedSearch(ctx context.Context, command string, o cliOptions, w *cliWorkspace) (any, error) {
	backup, _, err := w.readProfile()
	if err != nil {
		return nil, err
	}
	var filters map[string]any
	for _, s := range backup.SavedSearches.Searches {
		if s.ID == o.id {
			if err = cliDecodeJSON(s.Filters, &filters); err != nil {
				return nil, err
			}
			break
		}
	}
	if filters == nil {
		return nil, cliFail(5, "Saved search is unavailable; supply --id")
	}
	d, err := w.readDataset(ctx, o.dataset)
	if err != nil {
		return nil, err
	}
	players, err := cliFilterPlayers(d.Result.Players, filters)
	if err != nil {
		return nil, err
	}
	if command == "saved-searches run" {
		return cliPaginate(players, o)
	}
	evidence := []map[string]any{}
	for _, p := range players {
		if o.id != "" && o.name != "" && fmt.Sprint(p.UID) != o.name {
			continue
		}
		evidence = append(evidence, map[string]any{"playerRef": ShortlistPlayerRef{d.ID, p.UID}, "criteria": filters, "matched": true, "age": p.Age, "overall": p.Overall})
	}
	return map[string]any{"criteria": filters, "matches": evidence}, nil
}

type ScenarioCoverage struct {
	Position string `json:"position"`
	Count    int    `json:"count"`
}
type ScenarioEvaluation struct {
	Fees              float64            `json:"fees"`
	Wages             float64            `json:"wages"`
	FeeRemaining      float64            `json:"feeRemaining"`
	WageRemaining     float64            `json:"wageRemaining"`
	MissingTargets    int                `json:"missingTargets"`
	UnresolvedPlayers int                `json:"unresolvedPlayers"`
	Coverage          []ScenarioCoverage `json:"coverage"`
}

// EvaluateRecruitmentScenario mirrors the GUI's evaluateScenario pure function.
func EvaluateRecruitmentScenario(scenario RecruitmentScenario, targets []ShortlistItem, squad []Player, resolve func(ShortlistPlayerRef) *Player) ScenarioEvaluation {
	result := ScenarioEvaluation{Coverage: []ScenarioCoverage{}}
	indices := map[string]int{}
	adjust := func(p *Player, delta int) {
		if p == nil {
			result.UnresolvedPlayers++
			return
		}
		positions := p.ParsedPositions
		if len(positions) == 0 {
			position := p.Position
			if position == "" {
				position = "Unknown"
			}
			positions = []string{position}
		}
		seen := map[string]bool{}
		for _, position := range positions {
			if seen[position] {
				continue
			}
			seen[position] = true
			i, ok := indices[position]
			if !ok {
				i = len(result.Coverage)
				indices[position] = i
				result.Coverage = append(result.Coverage, ScenarioCoverage{Position: position})
			}
			result.Coverage[i].Count += delta
		}
	}
	for i := range squad {
		adjust(&squad[i], 1)
	}
	for _, ref := range scenario.Signings {
		var target *ShortlistItem
		for i := range targets {
			if targets[i].PlayerRef == ref {
				target = &targets[i]
				break
			}
		}
		if target == nil {
			result.MissingTargets++
			continue
		}
		result.Fees += target.TargetFee
		result.Wages += target.TargetWage
		adjust(resolve(ref), 1)
	}
	for _, sale := range scenario.Sales {
		result.Fees -= sale.Fee
		result.Wages -= sale.Wage
		var player *Player
		for i := range squad {
			if squad[i].UID == sale.PlayerRef.PlayerUID {
				player = &squad[i]
				break
			}
		}
		adjust(player, -1)
	}
	result.FeeRemaining = scenario.FeeBudget - result.Fees
	result.WageRemaining = scenario.WageBudget - result.Wages
	return result
}

func cliScenario(ctx context.Context, o cliOptions, w *cliWorkspace) (any, error) {
	backup, _, err := w.readProfile()
	if err != nil {
		return nil, err
	}
	for _, list := range backup.Shortlists.Lists {
		if o.name != "" && list.ID != o.name {
			continue
		}
		for _, scenario := range list.Scenarios {
			if scenario.ID != o.id {
				continue
			}
			d, err := w.readDataset(ctx, scenario.DatasetID)
			if err != nil {
				return nil, err
			}
			clubPath := filepath.Join(w.dir, "profile", "managed-club", d.ID+".json")
			data, err := os.ReadFile(clubPath)
			if err != nil {
				return nil, cliFail(5, "a Managed club is required to evaluate squad coverage")
			}
			var club ManagedTeam
			if err = json.Unmarshal(data, &club); err != nil {
				return nil, err
			}
			squad := filterByClub(d.Result.Players, club.Club)
			var resolveErr error
			result := EvaluateRecruitmentScenario(scenario, list.Items, squad, func(ref ShortlistPlayerRef) *Player {
				d, err := w.readDataset(ctx, ref.DatasetID)
				if err != nil {
					return nil
				}
				for i := range d.Result.Players {
					if d.Result.Players[i].UID == ref.PlayerUID {
						return &d.Result.Players[i]
					}
				}
				return nil
			})
			return result, resolveErr
		}
	}
	return nil, cliFail(5, "scenario is unavailable; supply --id and optionally --name shortlist ID")
}
