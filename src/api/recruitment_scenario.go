package main

import (
	"github.com/google/uuid"
	"math"
	"strings"
)

type RecruitmentScenario struct {
	ID         string               `json:"id"`
	Name       string               `json:"name"`
	DatasetID  string               `json:"datasetId"`
	FeeBudget  float64              `json:"feeBudget"`
	WageBudget float64              `json:"wageBudget"`
	Signings   []ShortlistPlayerRef `json:"signings"`
	Sales      []ScenarioSale       `json:"sales"`
}
type ScenarioSale struct {
	PlayerRef ShortlistPlayerRef `json:"playerRef"`
	Fee       float64            `json:"fee"`
	Wage      float64            `json:"wage"`
}

func validPlanningCurrency(value string) bool {
	return strings.Contains("|USD|GBP|EUR|AUD|CAD|JPY|BRL|CHF|", "|"+value+"|") && len(value) == 3
}
func validPlanningAmount(value float64) bool {
	return value >= 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}
func validPlayerRef(ref ShortlistPlayerRef) bool {
	_, err := uuid.Parse(ref.DatasetID)
	return err == nil && ref.PlayerUID > 0
}
func validateRecruitmentScenarios(scenarios []RecruitmentScenario) bool {
	if len(scenarios) > 50 {
		return false
	}
	seen := map[string]bool{}
	for _, scenario := range scenarios {
		if scenario.ID == "" || seen[scenario.ID] || len(scenario.Name) == 0 || len(scenario.Name) > 100 || !validPlanningAmount(scenario.FeeBudget) || !validPlanningAmount(scenario.WageBudget) || len(scenario.Signings) > 100 || len(scenario.Sales) > 100 {
			return false
		}
		seen[scenario.ID] = true
		if _, err := uuid.Parse(scenario.DatasetID); err != nil {
			return false
		}
		refs := map[ShortlistPlayerRef]bool{}
		for _, ref := range scenario.Signings {
			if !validPlayerRef(ref) || refs[ref] {
				return false
			}
			refs[ref] = true
		}
		for _, sale := range scenario.Sales {
			if !validPlayerRef(sale.PlayerRef) || sale.PlayerRef.DatasetID != scenario.DatasetID || refs[sale.PlayerRef] || !validPlanningAmount(sale.Fee) || !validPlanningAmount(sale.Wage) {
				return false
			}
			refs[sale.PlayerRef] = true
		}
	}
	return true
}
