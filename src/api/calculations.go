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

	// Scale to approximately 0-100. The factor 5.3 was used in the original code.
	// (20 * 5.3 = 106, 1 * 5.3 = 5.3). Clamping to 0-99.
	// A common approach is ((avg - 1) / 19) * 99 to map 1-20 to 0-99.
	// Let's use a slightly adjusted scaling to better fit 0-99.
	// If average is 1, score should be low. If 20, score should be 99.
	// Scaled = ( (CurrentValue - MinValue) / (MaxValue - MinValue) ) * ScaleRange + ScaleMin
	// Here, MinValue=1, MaxValue=20 for attributes. ScaleRange=99, ScaleMin=0.
	// Scaled = ( (weightedAverage - 1) / (20 - 1) ) * 99

	scaledScore := ((weightedAverage - 1.0) / 19.0) * 99.0
	finalScore := int(math.Round(scaledScore))

	return Clamp(finalScore, 0, 99)
}

// CalculateOverallForRoleGo calculates a player's suitability for a specific role.
// playerNumericAttributes are 1-20. roleSpecificAttrWeights define importance.
// The result is scaled by overallScalingFactor (e.g., 5.85) and clamped to 0-99.
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
				validAttributeValue := math.Max(1, math.Min(20, float64(attributeValue))) // Ensure 1-20 range
				weightedAttributeSum += validAttributeValue * float64(weightForAttribute)
				totalApplicableWeightsSum += float64(weightForAttribute)
			}
		}
	}

	if totalApplicableWeightsSum == 0 {
		return 0 // No relevant attributes for this role or all weights were zero
	}

	rawPositionalOverall := weightedAttributeSum / totalApplicableWeightsSum // This is on a 1-20 scale

	// Scale to 0-99 using the overallScalingFactor
	// Original: int(math.Min(99, math.Round(rawPositionalOverall*overallScalingFactor)))
	// Let's use the same ((X-1)/19)*99 logic for consistency if overallScalingFactor is meant to achieve this.
	// If overallScalingFactor = 5.85, then 1*5.85 = 5.85, 20*5.85 = 117.
	// This factor seems designed to map 1-20 to a wider range then clamp.
	// Let's stick to the original scaling method for this function if it's distinct.
	scaledScore := rawPositionalOverall * overallScalingFactor
	finalScore := int(math.Round(scaledScore))

	return Clamp(finalScore, 0, 99)
}

// CalculateCategoryBasedOverall calculates a general overall score based on FIFA stat categories (PHY, SHO, etc.).
// The input FIFA stat categories (player.PHY, player.SHO) are already on a 0-100 (clamped 0-99) scale.
// The categoryWeights define the importance of each category for this specific overall type.
// The result of this function will also be on a 0-99 scale.
func CalculateCategoryBasedOverall(player *Player, categoryWeights map[string]int) int {
	categories := make(map[string]int)
	categories["PHY"] = player.PHY
	categories["SHO"] = player.SHO
	categories["PAS"] = player.PAS
	categories["DRI"] = player.DRI
	categories["DEF"] = player.DEF
	categories["MEN"] = player.MEN
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

	return Clamp(calculatedOverall, 0, 99)
}
