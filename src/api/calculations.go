package main

import (
	"log"
	"math"
)

// Pre-computed scaling lookup table for better performance
var nonLinearScalingLookup [100]int

// Initialize the lookup table once at startup
func init() {
	for i := 0; i < 100; i++ {
		nonLinearScalingLookup[i] = computeNonLinearScaling(float64(i))
	}
}

// computeNonLinearScaling is the original function used to build the lookup table
func computeNonLinearScaling(linearRating float64) int {
	// Clamp input to reasonable bounds
	if linearRating <= 0 {
		return 0
	}
	if linearRating >= 99 {
		return 99
	}

	// Define the inflection point where compression starts (around 75)
	inflectionPoint := 75.0

	if linearRating >= inflectionPoint {
		// For ratings 75+, apply minimal compression (keep them roughly the same)
		// Use a gentle curve that preserves most of the original rating
		scaledRating := inflectionPoint + (linearRating-inflectionPoint)*0.95
		return int(math.Round(scaledRating))
	} else {
		// For ratings below 75, apply progressive compression
		// Use a power curve that becomes more aggressive as ratings get lower

		// Normalize to 0-1 scale relative to inflection point
		normalizedRating := linearRating / inflectionPoint

		// Apply power curve: higher exponent = more compression for low values
		// Using exponent 1.8 creates good separation
		compressedNormalized := math.Pow(normalizedRating, 1.8)

		// Scale back to final rating
		scaledRating := compressedNormalized * inflectionPoint

		// Ensure minimum rating progression (avoid clustering too much at bottom)
		if scaledRating < 10 && linearRating > 20 {
			scaledRating = 10 + (linearRating-20)*0.15
		}

		return int(math.Round(scaledRating))
	}
}

// applyNonLinearScaling applies a non-linear scaling curve to compress lower ratings
// while keeping higher ratings (75+) relatively unchanged.
// This creates a natural scaling that:
// - Keeps players rated 75+ roughly the same
// - Progressively lowers players below 75
// - Makes players at 50 or below significantly lower
func applyNonLinearScaling(linearRating float64) int {
	// Use lookup table for integer values
	if linearRating >= 0 && linearRating < 100 && linearRating == float64(int(linearRating)) {
		return nonLinearScalingLookup[int(linearRating)]
	}

	// Fallback to computation for non-integer values
	return computeNonLinearScaling(linearRating)
}

// FastClamp efficiently clamps a value between min and max bounds
func FastClamp(value, minVal, maxVal int) int {
	if value < minVal {
		return minVal
	}
	if value > maxVal {
		return maxVal
	}
	return value
}

// CalculateFifaStatGo calculates a FIFA-style category stat (e.g., PHY, SHO) from individual attributes.
// The playerNumericAttributes map should contain attributes on a 1-20 scale.
// The result is scaled using non-linear scaling and clamped at 99.
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

	var weightedSum, totalWeightOfPresentAttributes int64 // Use int64 to avoid overflow

	for attrName, attrWeight := range categoryAttributeWeights {
		attrValue, exists := playerNumericAttributes[attrName]
		if exists && attrValue >= 1 && attrValue <= 20 {
			// Use int64 arithmetic to prevent overflow
			weightedSum += int64(attrValue * attrWeight)
			totalWeightOfPresentAttributes += int64(attrWeight)
		}
	}

	if totalWeightOfPresentAttributes == 0 {
		return 0 // Avoid division by zero; no relevant attributes found or all had zero weight
	}

	// Calculate weighted average (faster than float division)
	weightedAverage := float64(weightedSum) / float64(totalWeightOfPresentAttributes)

	// Apply original linear scaling first to get to ~0-100 scale
	linearScore := weightedAverage * 5.3

	// Apply non-linear scaling to compress lower ratings
	finalScore := applyNonLinearScaling(linearScore)

	return FastClamp(finalScore, 0, 99)
}

// CalculateFifaStatGoLinear calculates a FIFA-style category stat using linear scaling (legacy method)
func CalculateFifaStatGoLinear(playerNumericAttributes map[string]int, categoryName string) int {
	muAttributeWeights.RLock()
	// Prefer loaded attributeWeights, fallback to defaultAttributeWeightsGo if the first is nil or category is missing
	var currentCategoryWeightsSource map[string]map[string]int
	if attributeWeights != nil {
		currentCategoryWeightsSource = attributeWeights
	} else {
		log.Printf("Warning: global attributeWeights is nil in CalculateFifaStatGoLinear. Using default for %s.", categoryName)
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

	// Apply original linear scaling method: Scale to approx 0-100 using factor 5.3
	finalScore := int(math.Round(weightedAverage * 5.3))

	return Clamp(finalScore, 0, 99) // Clamp from utils.go
}

// CalculateOverallForRoleGoLinear calculates a player's suitability for a specific role using linear scaling (legacy method)
func CalculateOverallForRoleGoLinear(playerNumericAttributes, roleSpecificAttrWeights map[string]int) int {
	if len(roleSpecificAttrWeights) == 0 {
		return 0
	}

	var weightedAttributeSum float64
	var totalApplicableWeightsSum float64

	// Optimized loop: reduce math operations and casting
	for attrKey, weightForAttribute := range roleSpecificAttrWeights {
		attributeValue, exists := playerNumericAttributes[attrKey]
		if !exists || attributeValue <= 0 {
			continue
		}

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

	if totalApplicableWeightsSum == 0 {
		return 0
	}

	// Apply original linear scaling
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

// CalculateOverallForRoleGo calculates a player's suitability for a specific role.
// playerNumericAttributes are 1-20. roleSpecificAttrWeights define importance.
// The result is scaled using non-linear scaling and clamped to 0-99.
func CalculateOverallForRoleGo(playerNumericAttributes, roleSpecificAttrWeights map[string]int) int {
	if len(roleSpecificAttrWeights) == 0 {
		return 0
	}

	var weightedAttributeSum float64
	var totalApplicableWeightsSum float64

	// Optimized loop: reduce math operations and casting
	for attrKey, weightForAttribute := range roleSpecificAttrWeights {
		attributeValue, exists := playerNumericAttributes[attrKey]
		if !exists || attributeValue <= 0 {
			continue
		}

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

	if totalApplicableWeightsSum == 0 {
		return 0
	}

	// Apply original linear scaling first
	linearScore := (weightedAttributeSum / totalApplicableWeightsSum) * overallScalingFactor

	// Apply non-linear scaling to compress lower ratings
	finalScore := applyNonLinearScaling(linearScore)

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
