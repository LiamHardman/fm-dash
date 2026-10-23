package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func cliWorkflow(ctx context.Context, command string, args []string, o cliOptions, stdin io.Reader, e *cliEnvelope) (any, error) {
	w, err := cliOpenWorkspace(o.workspace)
	if err != nil {
		return nil, err
	}
	if command == "presets list" {
		return []string{"first-team-ready", "high-potential-wonderkid", "affordable-target", "free-agent-depth"}, nil
	}
	if strings.HasPrefix(command, "weights ") {
		return cliWeights(ctx, command, o, stdin, w)
	}
	if command == "saved-searches run" || command == "saved-searches explain" {
		return cliSavedSearch(ctx, command, o, w)
	}
	if command == "scenarios evaluate" {
		return cliScenario(ctx, o, w)
	}
	if strings.HasPrefix(command, "profile ") || strings.HasPrefix(command, "recipes ") || command == "shortlists list" || command == "shortlists set" || command == "saved-searches list" || command == "saved-searches set" || command == "progression projects" || command == "progression set" {
		return cliProfile(ctx, command, o, stdin, w)
	}
	if command == "references resolve" {
		return cliReferences(ctx, o, stdin, w)
	}
	if command == "datasets import" {
		return cliImport(ctx, args, o, stdin, w, e)
	}
	if command == "datasets list" {
		return w.listDatasets(ctx)
	}
	if command == "progression analyze" {
		return cliProgression(ctx, o, stdin, w)
	}
	if command == "config show" {
		return map[string]any{"attributeWeights": GetAttributeWeightsSnapshot(), "roleWeights": roleSpecificOverallWeights, "useScaledRatings": GetUseScaledRatings(), "datasetRetentionDays": int(getRetentionPeriod().Hours() / 24)}, nil
	}
	if command == "config set" || command == "config validate" {
		return cliConfig(ctx, command, o, stdin, w)
	}
	if command == "status" {
		datasets, err := w.listDatasets(ctx)
		return map[string]any{"workspace": w.dir, "datasets": datasets, "runtime": cliBuildInfo()}, err
	}
	if err = cliRequired(o.dataset, "--dataset"); err != nil {
		return nil, err
	}
	if command == "datasets delete" {
		path, err := w.datasetPath(o.dataset)
		if err != nil {
			return nil, err
		}
		err = w.mutate(ctx, func() error {
			err := os.Remove(path)
			if os.IsNotExist(err) {
				return cliFail(5, "Dataset is unavailable")
			}
			return err
		})
		return map[string]any{"deleted": o.dataset}, err
	}
	d, err := w.readDataset(ctx, o.dataset)
	if err != nil {
		return nil, err
	}
	players := d.Result.Players
	switch command {
	case "datasets show":
		return d, nil
	case "datasets share":
		base, err := url.Parse(o.baseURL)
		if err != nil || base.Host == "" || (base.Scheme != "http" && base.Scheme != "https") {
			return nil, cliFail(2, "base-url must be HTTP(S)")
		}
		return map[string]any{"datasetId": d.ID, "url": base.ResolveReference(&url.URL{Path: "/dataset/" + d.ID}).String(), "requiresDatasetOnBackend": true}, nil
	case "datasets rename":
		if err = cliRequired(o.name, "--name"); err != nil {
			return nil, err
		}
		err = w.mutate(ctx, func() error {
			fresh, e := w.readDataset(ctx, o.dataset)
			if e != nil {
				return e
			}
			fresh.Label = o.name
			d = fresh
			return w.writeDataset(d)
		})
		return d, err
	case "ratings recalculate":
		err = w.mutate(ctx, func() error {
			fresh, e := w.readDataset(ctx, o.dataset)
			if e != nil {
				return e
			}
			d = fresh
			for i := range d.Result.Players {
				if err := ctx.Err(); err != nil {
					return err
				}
				EnhancePlayerWithCalculations(&d.Result.Players[i])
			}
			NormalizeMBRScoresRelativeToMax(d.Result.Players)
			CalculatePlayerPerformancePercentiles(d.Result.Players)
			d.Result.Options.Stage = "complete"
			if err := ctx.Err(); err != nil {
				return err
			}
			return w.writeDataset(d)
		})
		return d, err
	case "managed-club get", "managed-club set", "managed-club clear":
		path := filepath.Join(w.dir, "profile", "managed-club", d.ID+".json")
		if command == "managed-club get" {
			data, err := os.ReadFile(path)
			if os.IsNotExist(err) {
				return nil, nil
			}
			if err != nil {
				return nil, err
			}
			var club ManagedTeam
			err = json.Unmarshal(data, &club)
			return club, err
		}
		var club ManagedTeam
		if command == "managed-club set" {
			club.Club = o.name
			if o.body != "" {
				data, err := cliBody(ctx, o, stdin)
				if err != nil {
					return nil, err
				}
				if err = cliDecodeJSON(data, &club); err != nil {
					return nil, err
				}
			}
			found := false
			for _, p := range players {
				if p.Club == club.Club && (club.Division == "" || p.Division == club.Division) {
					found = true
					if club.Division == "" {
						club.Division = p.Division
					}
					break
				}
			}
			if !found {
				return nil, cliFail(3, "Managed club is not present in this Dataset")
			}
		}
		err = w.mutate(ctx, func() error {
			if command == "managed-club clear" {
				err := os.Remove(path)
				if os.IsNotExist(err) {
					return nil
				}
				return err
			}
			if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
				return err
			}
			data, _ := json.Marshal(club)
			return atomicConfigWrite(path, data)
		})
		return club, err
	case "players show", "players compare":
		if err = cliRequired(o.id, "--id"); err != nil {
			return nil, err
		}
		selected := []Player{}
		for _, s := range strings.Split(o.id, ",") {
			id, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
			if err != nil || id <= 0 {
				return nil, cliFail(2, "invalid player UID")
			}
			matches := 0
			for _, p := range players {
				if p.UID == id {
					selected = append(selected, p)
					matches++
				}
			}
			if matches != 1 {
				return nil, cliFail(5, "player UID %d is unavailable or ambiguous", id)
			}
		}
		if command == "players show" {
			if len(selected) != 1 {
				return nil, cliFail(2, "players show requires one UID")
			}
			return selected[0], nil
		}
		return selected, nil
	case "teams list", "teams top":
		if o.name != "" {
			subset := []Player{}
			for _, p := range players {
				if p.Club == o.name {
					subset = append(subset, p)
				}
			}
			players = subset
		}
		return processTopTeamsData(players, o.limit), nil
	case "nations list":
		return processNationsSummaryData(players), nil
	case "nations top-players":
		return selectTopNationPlayers(players, o.name, o.limit), nil
	case "leagues list":
		return CreateConcurrentLeagueProcessor(max(1, o.workers)).ProcessLeaguesAsync(ctx, players), nil
	case "divisions list":
		names := map[string]bool{}
		for _, p := range playersWithClubMetadata(players) {
			names[p.Division] = true
		}
		return cliSortedKeys(names), nil
	case "squad best-xi", "squad depth":
		return cliSquad(players, o)
	case "recruitment recommend":
		return cliRecommend(ctx, players, o, stdin)
	case "recruitment upgrades":
		data, err := cliBody(ctx, o, stdin)
		if err != nil {
			return nil, err
		}
		var req UpgradeFinderRequest
		if err = cliDecodeJSON(data, &req); err != nil {
			return nil, err
		}
		return OptimizedFindPlayerUpgrades(players, filterByClub(players, req.Team), req), nil
	}
	filters, err := cliPreset(o.preset, players)
	if err != nil {
		return nil, err
	}
	if o.query != "" && command != "players search" {
		var overrides map[string]any
		if err = cliDecodeJSON([]byte(o.query), &overrides); err != nil {
			return nil, err
		}
		for key, value := range overrides {
			filters[key] = value
		}
	}
	if command == "recruitment free-agents" {
		free := []Player{}
		for _, p := range players {
			club := strings.ToLower(p.Club)
			if club == "" || club == "-" || strings.Contains(club, "free") || strings.Contains(club, "unattached") {
				free = append(free, p)
			}
		}
		players = free
		filters["sortField"] = "overall"
	}
	if command == "recruitment wonderkids" {
		filters["ageRange"] = map[string]any{"min": json.Number("15"), "max": json.Number("21")}
		filters["sortField"] = "overall"
	}
	selected, err := cliFilterPlayers(players, filters)
	if err != nil {
		return nil, err
	}
	switch command {
	case "analysis save":
		return SaveAnalysis(selected), nil
	case "analysis performance":
		return PerformanceAnalysis(selected), nil
	case "players list":
		return cliPaginate(selected, o)
	case "players search":
		return performSearch(players, o.query), nil
	case "datasets export":
		return cliExport(ctx, selected, d, o, stdin)
	case "recruitment bargains":
		return processBargainHunter(selected, 0, 0, 0, 100, 0), nil
	case "recruitment free-agents", "recruitment wonderkids":
		if command == "recruitment wonderkids" {
			counts := map[string]int{}
			top := []Player{}
			for _, p := range selected {
				if counts[p.Age] < 10 {
					top = append(top, p)
					counts[p.Age]++
				}
			}
			selected = top
		}
		return cliPaginate(selected, o)
	default:
		return nil, cliFail(2, "workflow %s is not implemented", command)
	}
}

func cliPaginate(players []Player, o cliOptions) (any, error) {
	if o.offset < 0 || o.limit < 0 {
		return nil, cliFail(2, "offset and limit must be nonnegative")
	}
	start := min(o.offset, len(players))
	end := len(players)
	if o.limit > 0 {
		end = min(end, start+o.limit)
	}
	return map[string]any{"players": players[start:end], "total": len(players), "offset": start, "limit": o.limit}, nil
}

// Filter keys follow the GUI's Saved search criteria. Unsupported criteria fail
// explicitly rather than silently broadening an agent's candidate pool.
func cliFilterPlayers(players []Player, filters map[string]any) ([]Player, error) {
	knownMinimum := map[string]bool{"Overall": true, "Salary": true, "PAC": true, "SHO": true, "PAS": true, "DRI": true, "DEF": true, "PHY": true, "GK": true, "DIV": true, "HAN": true, "REF": true, "KIC": true, "SPD": true, "POS": true, "MBR": true, "CA": true, "MEN": true}
	canonicalAttributes := map[string]string{}
	for key := range fmAttributeKeys {
		normalized := strings.ReplaceAll(key, " ", "")
		knownMinimum[normalized] = true
		canonicalAttributes[normalized] = key
	}
	for _, key := range PerformanceStatKeys {
		knownMinimum[key] = true
	}
	allowed := map[string]bool{"name": true, "club": true, "position": true, "role": true, "nationality": true, "division": true, "continentNationalities": true, "mediaHandling": true, "personality": true, "ageRange": true, "transferValueRangeLocal": true, "minSalary": true, "maxSalary": true, "minOverall": true, "maxOverall": true, "sortField": true, "sortDir": true}
	for key := range filters {
		if strings.HasPrefix(key, "min") && !knownMinimum[strings.TrimPrefix(key, "min")] {
			return nil, cliFail(2, "unsupported minimum filter %q", key)
		}
		if !allowed[key] && !strings.HasPrefix(key, "min") {
			return nil, cliFail(2, "unsupported filter %q", key)
		}
		if strings.HasPrefix(key, "min") || key == "maxOverall" || key == "maxSalary" {
			value, err := strconv.ParseFloat(fmt.Sprint(filters[key]), 64)
			if err != nil || math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
				return nil, cliFail(2, "filter %s must be a nonnegative finite number", key)
			}
		}
	}
	for _, key := range []string{"ageRange", "transferValueRangeLocal"} {
		if raw, exists := filters[key]; exists {
			r, ok := raw.(map[string]any)
			if !ok {
				return nil, cliFail(2, "%s must be an object", key)
			}
			for field, v := range r {
				if field == "userSet" {
					if _, ok := v.(bool); !ok {
						return nil, cliFail(2, "userSet must be boolean")
					}
					continue
				}
				if field != "min" && field != "max" {
					return nil, cliFail(2, "unknown %s field %q", key, field)
				}
				n, err := strconv.ParseFloat(fmt.Sprint(v), 64)
				if err != nil || math.IsNaN(n) || math.IsInf(n, 0) || n < 0 {
					return nil, cliFail(2, "invalid numeric range bound")
				}
			}
		}
	}
	text := func(key string) string { v, _ := filters[key].(string); return v }
	number := func(v any) float64 { f, _ := strconv.ParseFloat(fmt.Sprint(v), 64); return f }
	match := func(value string, wanted any) bool {
		if wanted == nil {
			return true
		}
		if values, ok := wanted.([]any); ok {
			if len(values) == 0 {
				return true
			}
			for _, v := range values {
				if strings.EqualFold(value, fmt.Sprint(v)) {
					return true
				}
			}
			return false
		}
		s, _ := wanted.(string)
		return s == "" || strings.EqualFold(value, s)
	}
	result := []Player{}
	for _, p := range players {
		if text("name") != "" && !strings.Contains(strings.ToLower(p.Name), strings.ToLower(text("name"))) {
			continue
		}
		if !match(p.Club, filters["club"]) || !match(p.Division, filters["division"]) || !match(p.Nationality, filters["nationality"]) || (p.Personality != "" && !match(p.Personality, filters["personality"])) {
			continue
		}
		if wanted, ok := filters["mediaHandling"].([]any); ok && len(wanted) > 0 && p.MediaHandling != "" {
			found := false
			for _, entry := range strings.Split(p.MediaHandling, ",") {
				if match(strings.TrimSpace(entry), wanted) {
					found = true
				}
			}
			if !found {
				continue
			}
		}
		if pos := filters["position"]; pos != nil {
			positions, ok := pos.([]any)
			if !ok {
				return nil, cliFail(2, "position must be an array")
			}
			found := len(positions) == 0
			for _, value := range positions {
				for _, short := range p.ShortPositions {
					if short == fmt.Sprint(value) {
						found = true
					}
				}
				if strings.Contains(p.Position, fmt.Sprint(value)) {
					found = true
				}
			}
			if !found {
				continue
			}
		}
		if role := text("role"); role != "" {
			found := false
			for _, r := range p.RoleSpecificOveralls {
				if r.RoleName == role {
					found = true
				}
			}
			if !found {
				continue
			}
		}
		if ageRange, ok := filters["ageRange"].(map[string]any); ok {
			age, err := strconv.Atoi(p.Age)
			if err != nil {
				continue
			}
			if v, exists := ageRange["min"]; exists && float64(age) < number(v) {
				continue
			}
			if v, exists := ageRange["max"]; exists && float64(age) > number(v) {
				continue
			}
		}
		if r, ok := filters["transferValueRangeLocal"].(map[string]any); ok {
			if v, exists := r["min"]; exists && float64(p.TransferValueAmount) < number(v) {
				continue
			}
			if v, exists := r["max"]; exists && float64(p.TransferValueAmount) > number(v) {
				continue
			}
		}
		if float64(p.Overall) < number(filters["minOverall"]) || float64(p.WageAmount) < number(filters["minSalary"]) {
			continue
		}
		if v, exists := filters["maxOverall"]; exists && float64(p.Overall) > number(v) {
			continue
		}
		if v, exists := filters["maxSalary"]; exists && float64(p.WageAmount) > number(v) {
			continue
		}
		pass := true
		for key, v := range filters {
			if !strings.HasPrefix(key, "min") || key == "minOverall" || key == "minSalary" {
				continue
			}
			attr := strings.TrimPrefix(key, "min")
			if len(attr) > 0 {
				attr = strings.ToUpper(attr[:1]) + attr[1:]
			}
			value, known := p.NumericAttributes[attr]
			if raw, ok := canonicalAttributes[attr]; ok {
				value, known = p.NumericAttributes[raw], true
				if text, exists := p.Attributes[raw]; exists {
					first := strings.SplitN(text, "-", 2)[0]
					value, _ = strconv.Atoi(strings.TrimSpace(first))
				}
			}
			if stat, ok := p.PerformanceStatsNumeric[attr]; ok {
				value, known = int(stat), true
			}
			for raw, rawValue := range p.NumericAttributes {
				if strings.ReplaceAll(raw, " ", "") == attr {
					value, known = rawValue, true
				}
			}
			categories := map[string]int{"PAC": p.PAC, "SHO": p.SHO, "PAS": p.PAS, "DRI": p.DRI, "DEF": p.DEF, "PHY": p.PHY, "GK": p.GK, "DIV": p.DIV, "HAN": p.HAN, "REF": p.REF, "KIC": p.KIC, "SPD": p.SPD, "POS": p.POS, "MBR": p.MBR, "CA": p.CA}
			if category, ok := categories[attr]; ok {
				value, known = category, true
			}
			if !known && !fmAttributeKeys[attr] && attr != "MEN" && number(v) > 0 {
				return nil, cliFail(2, "unknown numeric attribute filter %q", key)
			}
			if float64(value) < number(v) {
				pass = false
				break
			}
		}
		if pass {
			result = append(result, p)
		}
	}
	if field := text("sortField"); field != "" {
		sort.SliceStable(result, func(i, j int) bool {
			var a, b float64
			switch field {
			case "overall":
				a, b = float64(result[i].Overall), float64(result[j].Overall)
			case "age":
				a, _ = strconv.ParseFloat(result[i].Age, 64)
				b, _ = strconv.ParseFloat(result[j].Age, 64)
			case "value":
				a, b = float64(result[i].TransferValueAmount), float64(result[j].TransferValueAmount)
			default:
				a, b = float64(result[i].NumericAttributes[field]), float64(result[j].NumericAttributes[field])
			}
			if text("sortDir") == "asc" {
				return a < b
			}
			return a > b
		})
	}
	return result, nil
}

func cliConfig(ctx context.Context, command string, o cliOptions, stdin io.Reader, w *cliWorkspace) (any, error) {
	data, err := cliBody(ctx, o, stdin)
	if err != nil {
		return nil, err
	}
	var config struct {
		UseScaledRatings *bool                     `json:"useScaledRatings,omitempty"`
		AttributeWeights map[string]map[string]int `json:"attributeWeights,omitempty"`
	}
	if err = cliDecodeJSON(data, &config); err != nil {
		return nil, err
	}
	if config.UseScaledRatings == nil && len(config.AttributeWeights) == 0 {
		return nil, cliFail(3, "configuration is empty")
	}
	if config.AttributeWeights != nil {
		if err = SetAttributeWeights(config.AttributeWeights); err != nil {
			return nil, cliFail(3, "%v", err)
		}
	}
	if config.UseScaledRatings != nil {
		SetUseScaledRatings(*config.UseScaledRatings)
	}
	resolved := map[string]any{"attributeWeights": GetAttributeWeightsSnapshot(), "useScaledRatings": GetUseScaledRatings()}
	if command == "config set" {
		err = w.mutate(ctx, func() error {
			data, err := json.Marshal(resolved)
			if err != nil {
				return err
			}
			return atomicConfigWrite(filepath.Join(w.dir, "config.json"), data)
		})
	}
	return resolved, err
}

func cliProgression(ctx context.Context, o cliOptions, stdin io.Reader, w *cliWorkspace) (any, error) {
	data, err := cliBody(ctx, o, stdin)
	if err != nil {
		return nil, err
	}
	var req ProgressionAnalyzeRequest
	if err = cliDecodeJSON(data, &req); err != nil {
		return nil, err
	}
	if len(req.DatasetIDs) < 2 {
		return nil, cliFail(3, "at least two Datasets are required")
	}
	datasets := map[string][]Player{}
	currency := ""
	for _, id := range req.DatasetIDs {
		if _, ok := datasets[id]; ok {
			return nil, cliFail(3, "duplicate Dataset identifier")
		}
		d, err := w.readDataset(ctx, id)
		if err != nil {
			return nil, err
		}
		datasets[id] = d.Result.Players
		if currency == "" {
			currency = d.Result.Currency
		}
	}
	order := req.Order
	if len(order) > 0 {
		seen := map[string]bool{}
		if len(order) != len(req.DatasetIDs) {
			return nil, cliFail(3, "order must contain every Dataset exactly once")
		}
		for _, id := range order {
			if _, ok := datasets[id]; !ok || seen[id] {
				return nil, cliFail(3, "invalid snapshot order")
			}
			seen[id] = true
		}
	} else {
		var ambiguous bool
		var ids []string
		order, ambiguous, ids = determineSnapshotOrder(req.DatasetIDs, datasets)
		if ambiguous {
			return ProgressionAnalyzeResponse{OrderAmbiguous: true, AmbiguousDatasetIDs: ids}, cliFail(3, "snapshot order is ambiguous; supply order")
		}
	}
	byID := map[string]map[int64]Player{}
	for _, id := range order {
		index := map[int64]Player{}
		duplicates := map[int64]bool{}
		for _, p := range datasets[id] {
			if _, exists := index[p.UID]; exists {
				duplicates[p.UID] = true
			}
			index[p.UID] = p
		}
		for uid := range duplicates {
			delete(index, uid)
		}
		delete(index, 0)
		byID[id] = index
	}
	players := []ProgressionPlayer{}
	for uid := range byID[order[0]] {
		snapshots := []Player{}
		for _, id := range order {
			p, exists := byID[id][uid]
			if !exists {
				break
			}
			snapshots = append(snapshots, p)
		}
		if len(snapshots) != len(order) {
			continue
		}
		p := snapshots[len(snapshots)-1]
		players = append(players, ProgressionPlayer{uid, p.Name, p.Position, p.Club, snapshots})
	}
	sort.Slice(players, func(i, j int) bool {
		if players[i].Name == players[j].Name {
			return players[i].UID < players[j].UID
		}
		return players[i].Name < players[j].Name
	})
	return ProgressionAnalyzeResponse{Order: order, Players: players, EmptyIntersection: len(players) == 0, CurrencySymbol: currency}, nil
}
