package main

import (
	"log"
	"math"
)

// CalculateFifaStatGo calculates a FIFA-style category stat (e.g., PHY, SHO) from individual attributes.
// The playerNumericAttributes map should contain attributes on a 1-20 scale.
// The result is scaled to be approximately 0-100 (clamped at 99).
func CalculateFifaStatGo(playerNumericAttributes map[string]int, categoryName string) int {
	muAttributeWeights.RLock()
	// Prefer loaded attributeWeights, fallback to defaultAttributeWeightsGo if the first is nil or category is missing
	var currentCategoryWeightsSource map[string]map[string]int
	if attributeWeights != nil {
		currentCategoryWeightsSource = attributeWeights
	} else {
		log.Printf("Warning: global attributeWeights is nil in CalculateFifaStatGo. Using default for %s.", categoryName)
		currentCategoryWeightsSource = defaultAttributeWeightsGo // Fallback to compiled-in defaults
	}
	muAttributeWeights.RUnlock()

	categoryAttributeWeights, ok := currentCategoryWeightsSource[categoryName]
	if !ok {
		// If category not in primary source, try the compiled-in default as a further fallback
		categoryAttributeWeights, ok = defaultAttributeWeightsGo[categoryName]
		if !ok {
			log.Printf("Error: Default attribute weights for category '%s' also not found. Returning 0.", categoryName)
			return 0
		}
		log.Printf("Warning: Category '%s' not found in loaded attribute weights, using compiled-in default.", categoryName)
	}

	var weightedSum float64
	var totalWeightOfPresentAttributes float64

	for attrName, attrWeight := range categoryAttributeWeights {
		attrValue, exists := playerNumericAttributes[attrName]
		if exists {
			// Ensure attribute values are within the expected 1-20 range for calculation
			if attrValue >= 1 && attrValue <= 20 {
				weightedSum += float64(attrValue * attrWeight)
				totalWeightOfPresentAttributes += float64(attrWeight)
			}
			// Values outside 1-20 (e.g., 0 if parsing failed) are ignored for this attribute's contribution
		}
	}

	if totalWeightOfPresentAttributes == 0 {
		return 0 // Avoid division by zero; no relevant attributes found or all had zero weight
	}

	weightedAverage := weightedSum / totalWeightOfPresentAttributes // This average is on a 1-20 scale

	// Revert to original scaling method: Scale to approx 0-100 using factor 5.3
	// Original: return int(math.Round(weightedAverage * 5.3))
	finalScore := int(math.Round(weightedAverage * 5.3))

	return Clamp(finalScore, 0, 99) // Clamp from utils.go
}

// CalculateOverallForRoleGo calculates a player's suitability for a specific role.
// playerNumericAttributes are 1-20. roleSpecificAttrWeights define importance.
// The result is scaled by overallScalingFactor (e.g., 5.85 from config.go) and clamped to 0-99.
func CalculateOverallForRoleGo(playerNumericAttributes map[string]int, roleSpecificAttrWeights map[string]int) int {
	if len(roleSpecificAttrWeights) == 0 {
		// log.Printf("Warning: No weights provided for role calculation. Returning 0.")
		return 0
	}

	var weightedAttributeSum float64
	var totalApplicableWeightsSum float64

	for attrKey, weightForAttribute := range roleSpecificAttrWeights {
		attributeValue, exists := playerNumericAttributes[attrKey]
		if exists {
			// Clamp attribute value to 1-20 for calculation, though they should already be.
			// Using 0 for missing/invalid attributes is also an option.
			// Original code used math.Max(0, math.Min(20, float64(attributeValue)))
			// Let's assume attributes are already valid 1-20 from parsing, or 0 if invalid.
			// We should only consider attributes > 0 for positive contribution.
			if attributeValue > 0 { // Consider only attributes with a positive value
				// The original calculation used math.Max(0, math.Min(20, float64(attributeValue)))
				// which means a 0 attribute would contribute 0.
				// If we only consider >0, then an attribute of 0 from parsing (e.g. failed Atoi) won't contribute.
				// This seems fine. If an attribute is truly 0, it shouldn't contribute.
				// If it's 1-20, it will be used as is.
				validAttributeValue := math.Max(1, math.Min(20, float64(attributeValue))) // Ensure 1-20 range for calculation
				weightedAttributeSum += validAttributeValue * float64(weightForAttribute)
				totalApplicableWeightsSum += float64(weightForAttribute)
			}
		}
	}

	if totalApplicableWeightsSum == 0 {
		return 0 // No relevant attributes for this role or all weights were zero
	}

	rawPositionalOverall := weightedAttributeSum / totalApplicableWeightsSum // This is on a 1-20 scale

	// Scale to 0-99 using the overallScalingFactor from config.go
	// Original: int(math.Min(99, math.Round(rawPositionalOverall*overallScalingFactor)))
	scaledScore := rawPositionalOverall * overallScalingFactor // overallScalingFactor is 5.85
	finalScore := int(math.Round(scaledScore))

	return Clamp(finalScore, 0, 99) // Clamp from utils.go
}

// CalculateCategoryBasedOverall calculates a general overall score based on FIFA stat categories (PAC, SHO, etc.).
// The input FIFA stat categories (player.PAC, player.SHO) are already on a 0-100 (clamped 0-99) scale.
// The categoryWeights define the importance of each category for this specific overall type.
// The result of this function will also be on a 0-99 scale.
func CalculateCategoryBasedOverall(player *Player, categoryWeights map[string]int) int {
	categories := make(map[string]int)
	categories["PAC"] = player.PAC
	categories["SHO"] = player.SHO
	categories["PAS"] = player.PAS
	categories["DRI"] = player.DRI
	categories["DEF"] = player.DEF
	categories["PHY"] = player.PHY
	// GK stat is not typically used for outfielders in this type of calculation.

	var weightedSum float64
	var totalWeight float64

	for catName, catWeight := range categoryWeights {
		catValue, exists := categories[catName] // catValue is 0-99
		if exists {
			weightedSum += float64(catValue * catWeight)
			totalWeight += float64(catWeight)
		}
	}

	if totalWeight == 0 {
		return 0 // Avoid division by zero
	}

	calculatedOverall := int(math.Round(weightedSum / totalWeight)) // Result is on 0-99 scale

	return Clamp(calculatedOverall, 0, 99) // Clamp from utils.go
}
