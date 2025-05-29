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
func CalculateOverallForRoleGo(playerNumericAttributes, roleSpecificAttrWeights map[string]int) int {
	if len(roleSpecificAttrWeights) == 0 {
		return 0
	}

	var weightedAttributeSum float64
	var totalApplicableWeightsSum float64

	// Optimized loop: reduce math operations and casting
	for attrKey, weightForAttribute := range roleSpecificAttrWeights {
		if attributeValue, exists := playerNumericAttributes[attrKey]; exists && attributeValue > 0 {
			// Fast path: assume attributes are already in valid 1-20 range
			// Only clamp if outside expected range (rare case)
			var validValue float64
			if attributeValue >= 1 && attributeValue <= 20 {
				validValue = float64(attributeValue) // Fast path - no math.Max/Min needed
			} else {
				validValue = math.Max(1, math.Min(20, float64(attributeValue))) // Slow path - clamp
			}

			weightFloat := float64(weightForAttribute)
			weightedAttributeSum += validValue * weightFloat
			totalApplicableWeightsSum += weightFloat
		}
	}

	if totalApplicableWeightsSum == 0 {
		return 0
	}

	// Combined scaling and rounding operation
	scaledScore := (weightedAttributeSum / totalApplicableWeightsSum) * overallScalingFactor
	finalScore := int(scaledScore + 0.5) // Faster than math.Round for positive numbers

	// Clamp result to 0-99 range
	if finalScore > 99 {
		return 99
	} else if finalScore < 0 {
		return 0
	}
	return finalScore
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
