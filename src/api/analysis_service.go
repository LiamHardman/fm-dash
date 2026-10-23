package main

import (
	"math"
	"sort"
)

type SaveAnalysisPoint struct {
	UID        int64  `json:"uid"`
	Name       string `json:"name"`
	CA         int    `json:"ca"`
	CAScaled   int    `json:"caScaled"`
	Overall    int    `json:"overall"`
	Difference int    `json:"difference"`
	Age        string `json:"age"`
	Value      int64  `json:"value"`
}

// SaveAnalysis returns the numbers behind SaveAnalysisPage's scatter, outlier
// lists and role distribution, including the same CA-to-99 scaling and rounding.
func SaveAnalysis(players []Player) any {
	points := []SaveAnalysisPoint{}
	counts := map[string]int{}
	roleOrder := []string{}
	for _, p := range players {
		ca := int(math.Floor(float64(p.CA)*99/200 + .5))
		points = append(points, SaveAnalysisPoint{p.UID, p.Name, p.CA, ca, p.Overall, p.Overall - ca, p.Age, p.TransferValueAmount})
		role := p.BestRoleOverall
		if role == "" {
			role = "Unknown"
		}
		if counts[role] == 0 {
			roleOrder = append(roleOrder, role)
		}
		counts[role]++
	}
	top := append([]SaveAnalysisPoint{}, points...)
	bottom := append([]SaveAnalysisPoint{}, points...)
	sort.SliceStable(top, func(i, j int) bool { return top[i].Difference > top[j].Difference })
	sort.SliceStable(bottom, func(i, j int) bool { return bottom[i].Difference < bottom[j].Difference })
	sort.SliceStable(roleOrder, func(i, j int) bool { return counts[roleOrder[i]] > counts[roleOrder[j]] })
	roles := []map[string]any{}
	for _, role := range roleOrder {
		share := float64(0)
		if len(players) > 0 {
			share = math.Floor(float64(counts[role])*1000/float64(len(players))+.5) / 10
		}
		roles = append(roles, map[string]any{"role": role, "count": counts[role], "share": share})
	}
	return map[string]any{"players": points, "overratedOutliers": top[:min(15, len(top))], "underratedOutliers": bottom[:min(15, len(bottom))], "roleDistribution": roles}
}

func PerformanceAnalysis(players []Player) any {
	top := map[string][]Player{}
	for _, key := range PerformanceStatKeys {
		candidates := []Player{}
		for _, p := range players {
			if value, ok := p.PerformanceStatsNumeric[key]; ok && !math.IsNaN(value) {
				candidates = append(candidates, p)
			}
		}
		sort.SliceStable(candidates, func(i, j int) bool {
			return candidates[i].PerformanceStatsNumeric[key] > candidates[j].PerformanceStatsNumeric[key]
		})
		top[key] = candidates[:min(10, len(candidates))]
	}
	ratings := make([]int, len(players))
	for i, p := range players {
		ratings[i] = p.Overall
	}
	sort.Sort(sort.Reverse(sort.IntSlice(ratings)))
	threshold := 0
	if len(ratings) > 0 {
		threshold = ratings[min(249, len(ratings)-1)]
	}
	minutes := []float64{}
	for _, p := range players {
		if p.Overall >= threshold {
			minutes = append(minutes, p.PerformanceStatsNumeric["Mins"])
		}
	}
	sort.Sort(sort.Reverse(sort.Float64Slice(minutes)))
	minuteThreshold := float64(0)
	if len(minutes) > 0 {
		minuteThreshold = math.Floor(minutes[min(99, len(minutes)-1)]/50+.5) * 50
	}
	return map[string]any{"players": players, "topPlayersByStat": top, "recommendedThresholds": map[string]any{"overall": threshold, "minutes": minuteThreshold}, "distributions": BuildPercentileDistributions(players, DivisionFilterAll, "")}
}
