package main

import (
	"math"
	"strconv"
	"strings"
)

// Global variables for attribute weights
var globalAttributeWeights map[string]map[string]int

// getDefaultWeightsForCategory returns the default weights for a given category
func getDefaultWeightsForCategory(categoryName string) map[string]int {
	if weights, exists := defaultAttributeWeightsGo[categoryName]; exists {
		return weights
	}
	return nil
}

// calculateWeightedAverageLinear calculates a weighted average using linear scaling
func calculateWeightedAverageLinear(playerNumericAttributes, categoryAttributeWeights map[string]int) (result int) {
	if len(categoryAttributeWeights) == 0 {
		result = 0
		return
	}

	var weightedAttributeSum float64
	var totalApplicableWeightsSum float64

	for attrKey, weightForAttribute := range categoryAttributeWeights {
		attributeValue, exists := playerNumericAttributes[attrKey]
		if !exists || attributeValue < 1 || attributeValue > 20 {
			continue
		}

		validValue := float64(attributeValue)
		weightFloat := float64(weightForAttribute)
		weightedAttributeSum += validValue * weightFloat
		totalApplicableWeightsSum += weightFloat
	}

	if totalApplicableWeightsSum == 0 {
		result = 0
		return
	}

	// Apply linear scaling
	scaledScore := (weightedAttributeSum / totalApplicableWeightsSum) * 5.3
	result = int(scaledScore + 0.5)
	result = Clamp(result, 0, 99)
	return
}

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
	}

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

// applyNonLinearScaling applies a non-linear scaling curve to compress lower ratings
// while keeping higher ratings (75+) relatively unchanged.
// This creates a natural scaling that:
// - Keeps players rated 75+ roughly the same
// - Progressively lowers players below 75
// - Makes players at 50 or below significantly lower
func applyNonLinearScaling(linearRating float64) int {
	if math.IsNaN(linearRating) || math.IsInf(linearRating, -1) {
		return 0
	}
	if math.IsInf(linearRating, 1) {
		return 99
	}

	// Use lookup table for integer values
	if linearRating >= 0 && linearRating < 100 && linearRating == float64(int(linearRating)) {
		return nonLinearScalingLookup[int(linearRating)]
	}

	// Fallback to computation for non-integer values
	return computeNonLinearScaling(linearRating)
}

// calculateWeightedAverage calculates the weighted average for a set of player attributes.
func calculateWeightedAverage(playerNumericAttributes, categoryAttributeWeights map[string]int) float64 {
	var weightedSum, totalWeightOfPresentAttributes int64

	for attrName, attrWeight := range categoryAttributeWeights {
		attrValue, exists := playerNumericAttributes[attrName]
		if exists && attrValue >= 1 && attrValue <= 20 {
			// Use int64 for calculations to prevent overflow.
			weightedSum += int64(attrValue * attrWeight)
			totalWeightOfPresentAttributes += int64(attrWeight)
		}
	}

	if totalWeightOfPresentAttributes == 0 {
		return 0.0 // Avoid division by zero.
	}

	return float64(weightedSum) / float64(totalWeightOfPresentAttributes)
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
		LogWarn("Warning: global attributeWeights is nil in CalculateFifaStatGo. Using default for %s.", categoryName)
		currentCategoryWeightsSource = defaultAttributeWeightsGo // Fallback to compiled-in defaults
	}
	muAttributeWeights.RUnlock()

	// Special handling for "PAS" category to calculate based on three methods and take the highest.
	if categoryName == "PAS" {
		// Method 1: Standard
		weights1, ok1 := currentCategoryWeightsSource["PAS_standard"]
		score1 := 0
		if ok1 {
			avg1 := calculateWeightedAverage(playerNumericAttributes, weights1)
			score1 = applyNonLinearScaling(avg1 * 5.3)
		}

		// Method 2: No Set Pieces
		weights2, ok2 := currentCategoryWeightsSource["PAS_no_set_pieces"]
		score2 := 0
		if ok2 {
			avg2 := calculateWeightedAverage(playerNumericAttributes, weights2)
			score2 = applyNonLinearScaling(avg2 * 5.3)
		}

		// Method 3: No Off The Ball
		weights3, ok3 := currentCategoryWeightsSource["PAS_no_off_ball"]
		score3 := 0
		if ok3 {
			avg3 := calculateWeightedAverage(playerNumericAttributes, weights3)
			score3 = applyNonLinearScaling(avg3 * 5.3)
		}

		// Determine the maximum score from the three methods.
		maxScore := score1
		if score2 > maxScore {
			maxScore = score2
		}
		if score3 > maxScore {
			maxScore = score3
		}
		return Clamp(maxScore, 0, 99)
	}

	categoryAttributeWeights, ok := currentCategoryWeightsSource[categoryName]
	if !ok {
		// If category not in primary source, try the compiled-in default as a further fallback
		categoryAttributeWeights, ok = defaultAttributeWeightsGo[categoryName]
		if !ok {
			LogWarn("Error: Default attribute weights for category '%s' also not found. Returning 0.", categoryName)
			return 0
		}
		LogWarn("Warning: Category '%s' not found in loaded attribute weights, using compiled-in default.", categoryName)
	}

	// Default calculation for all other categories.
	weightedAverage := calculateWeightedAverage(playerNumericAttributes, categoryAttributeWeights)
	if weightedAverage == 0 {
		return 0
	}

	// Apply original linear scaling first to get to ~0-100 scale
	linearScore := weightedAverage * 5.3

	// Apply non-linear scaling to compress lower ratings
	finalScore := applyNonLinearScaling(linearScore)

	return Clamp(finalScore, 0, 99)
}

// CalculateFifaStatGoLinear calculates a FIFA-style category stat using linear scaling (legacy method)
func CalculateFifaStatGoLinear(playerNumericAttributes map[string]int, categoryName string) int {
	if globalAttributeWeights == nil {
		LogWarn("Warning: global attributeWeights is nil in CalculateFifaStatGoLinear. Using default for %s.", categoryName)
		// Use default weights for this category
		defaultWeights := getDefaultWeightsForCategory(categoryName)
		if defaultWeights == nil {
			LogWarn("Error: Default attribute weights for category '%s' also not found. Returning 0.", categoryName)
			return 0
		}
		return calculateWeightedAverageLinear(playerNumericAttributes, defaultWeights)
	}

	weights, exists := globalAttributeWeights[categoryName]
	if !exists {
		LogWarn("Warning: Category '%s' not found in loaded attribute weights, using compiled-in default.", categoryName)
		// Use default weights for this category
		defaultWeights := getDefaultWeightsForCategory(categoryName)
		if defaultWeights == nil {
			LogWarn("Error: Default attribute weights for category '%s' also not found. Returning 0.", categoryName)
			return 0
		}
		return calculateWeightedAverageLinear(playerNumericAttributes, defaultWeights)
	}

	return calculateWeightedAverageLinear(playerNumericAttributes, weights)
}

// CalculateOverallForRoleGoLinear calculates a player's suitability for a specific role using linear scaling (legacy method)
func CalculateOverallForRoleGoLinear(playerNumericAttributes, roleSpecificAttrWeights map[string]int) (result int) {
	if len(roleSpecificAttrWeights) == 0 || len(playerNumericAttributes) == 0 {
		return 0
	}

	var weightedAttributeSum float64
	var totalApplicableWeightsSum float64

	// Caller holds player.mu.RLock(); iterate weights directly, look up each in attributes.
	for attrKey, weightForAttribute := range roleSpecificAttrWeights {
		attributeValue, exists := playerNumericAttributes[attrKey]
		if !exists || attributeValue <= 0 {
			continue
		}

		var validValue float64
		if attributeValue >= 1 && attributeValue <= 20 {
			validValue = float64(attributeValue)
		} else {
			validValue = math.Max(1, math.Min(20, float64(attributeValue)))
		}

		weightedAttributeSum += validValue * float64(weightForAttribute)
		totalApplicableWeightsSum += float64(weightForAttribute)
	}

	if totalApplicableWeightsSum == 0 {
		return 0
	}

	scaledScore := (weightedAttributeSum / totalApplicableWeightsSum) * overallScalingFactor
	finalScore := int(scaledScore + 0.5)

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
func CalculateOverallForRoleGo(playerNumericAttributes, roleSpecificAttrWeights map[string]int) (result int) {
	if len(roleSpecificAttrWeights) == 0 || len(playerNumericAttributes) == 0 {
		return 0
	}

	var weightedAttributeSum float64
	var totalApplicableWeightsSum float64

	// Caller holds player.mu.RLock(); iterate weights directly, look up each in attributes.
	for attrKey, weightForAttribute := range roleSpecificAttrWeights {
		attributeValue, exists := playerNumericAttributes[attrKey]
		if !exists || attributeValue <= 0 {
			continue
		}

		var validValue float64
		if attributeValue >= 1 && attributeValue <= 20 {
			validValue = float64(attributeValue)
		} else {
			validValue = math.Max(1, math.Min(20, float64(attributeValue)))
		}

		weightedAttributeSum += validValue * float64(weightForAttribute)
		totalApplicableWeightsSum += float64(weightForAttribute)
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

// roleStyleCategory strips a role's trailing duty (e.g. "- Defend"/"- Cover"/"- Stopper"/
// "- Support"/"- Attack") to get its style category, e.g. "DC - Ball Playing Defender - Defend"
// becomes "DC - Ball Playing Defender". Role names always have the form "POS - Style - Duty".
func roleStyleCategory(roleName string) string {
	idx := strings.LastIndex(roleName, " - ")
	if idx == -1 {
		return roleName
	}
	return roleName[:idx]
}

// CalculateMeanCategoryOverall aggregates a player's role-specific overalls into a single Overall
// score. Duty variants of the same style (Cover/Defend/Stopper/Support/Attack) are collapsed into
// one category via their best score, and Overall is the mean across every distinct category the
// player qualifies for — not a top-N subset. A top-N subset lets a one-dimensional player hide
// their weak categories behind duplicated duty-variants of their one strong category, which is
// how a pure defensive specialist ends up rated close to an equally-defensive but also
// ball-playing-capable peer.
func CalculateMeanCategoryOverall(roleScores []RoleOverallScore) int {
	if len(roleScores) == 0 {
		return 0
	}

	bestByCategory := make(map[string]int, len(roleScores))
	for _, r := range roleScores {
		category := roleStyleCategory(r.RoleName)
		if existing, ok := bestByCategory[category]; !ok || r.Score > existing {
			bestByCategory[category] = r.Score
		}
	}

	total := 0
	for _, score := range bestByCategory {
		total += score
	}
	return total / len(bestByCategory)
}

// categoryStatValue returns the FIFA stat value for a named category key without allocating a map.
func categoryStatValue(player *Player, catName string) (int, bool) {
	switch catName {
	case "PAC":
		return player.PAC, true
	case "SHO":
		return player.SHO, true
	case "PAS":
		return player.PAS, true
	case "DRI":
		return player.DRI, true
	case "DEF":
		return player.DEF, true
	case "PHY":
		return player.PHY, true
	}
	return 0, false
}

// CalculateCategoryBasedOverall calculates a general overall score based on FIFA stat categories (PAC, SHO, etc.).
func CalculateCategoryBasedOverall(player *Player, categoryWeights map[string]int) int {
	var weightedSum float64
	var totalWeight float64

	for catName, catWeight := range categoryWeights {
		if catValue, exists := categoryStatValue(player, catName); exists {
			weightedSum += float64(catValue * catWeight)
			totalWeight += float64(catWeight)
		}
	}

	if totalWeight == 0 {
		return 0
	}

	return Clamp(int(math.Round(weightedSum/totalWeight)), 0, 99)
}

// CalculateTotalStats calculates the sum of all physical, mental, and technical attributes
func CalculateTotalStats(playerNumericAttributes map[string]int) int {
	total := 0

	// Physical attributes
	physicalAttrs := []string{"Acc", "Agi", "Bal", "Jum", "Nat", "Pac", "Sta", "Str"}
	for _, attr := range physicalAttrs {
		if value, exists := playerNumericAttributes[attr]; exists && value > 0 {
			total += value
		}
	}

	// Mental attributes
	mentalAttrs := []string{"Agg", "Ant", "Bra", "Cmp", "Cnt", "Dec", "Det", "Fla", "Ldr", "OtB", "Pos", "Tea", "Vis", "Wor"}
	for _, attr := range mentalAttrs {
		if value, exists := playerNumericAttributes[attr]; exists && value > 0 {
			total += value
		}
	}

	// Technical attributes
	technicalAttrs := []string{"Cor", "Cro", "Dri", "Fin", "Fir", "Fre", "Hea", "Lon", "L Th", "Mar", "Pas", "Pen", "Tck", "Tec"}
	for _, attr := range technicalAttrs {
		if value, exists := playerNumericAttributes[attr]; exists && value > 0 {
			total += value
		}
	}

	return total
}

// CalculateMoneyballRating calculates the Moneyball Rating (MBR) based on overall, age, mentality, and value score
func CalculateMoneyballRating(player *Player) int {
	// Base rating is the player's overall divided by 3 (reduced from 2)
	baseRating := player.Overall / 3

	// Age modifiers
	ageInt, err := strconv.Atoi(player.Age)
	if err != nil {
		ageInt = 26 // Default age if parsing fails
	}
	ageModifier := getAgeModifier(ageInt)

	// Mentality modifier based on personality
	mentalityModifier := getMentalityModifier(player.Personality)

	// Calculate value score using the same methodology as bargain hunter
	var calculatedValueScore float64
	overall := float64(player.Overall)
	transferValueMillions := float64(player.TransferValueAmount) / 1000000.0

	// Prevent division by zero
	if transferValueMillions == 0 {
		calculatedValueScore = 0
	} else {
		// Calculate value-per-rating ratio (millions per overall point)
		valuePerRating := transferValueMillions / overall

		// Get expected value for this overall rating
		expectedValuePerRating := getExpectedValuePerRating(overall)

		// Calculate how many times more expensive they are than expected
		priceMultiplier := valuePerRating / expectedValuePerRating

		// Apply aggressive penalty for expensive players in the value score itself
		var valuePenalty float64
		switch {
		case priceMultiplier >= 5.0:
			valuePenalty = 0.1 // 90% penalty for extremely expensive players
		case priceMultiplier >= 4.0:
			valuePenalty = 0.2 // 80% penalty for severely expensive players
		case priceMultiplier >= 3.0:
			valuePenalty = 0.3 // 70% penalty for very expensive players
		case priceMultiplier >= 2.5:
			valuePenalty = 0.4 // 60% penalty for expensive players
		case priceMultiplier >= 2.0:
			valuePenalty = 0.5 // 50% penalty for somewhat expensive players
		case priceMultiplier >= 1.5:
			valuePenalty = 0.7 // 30% penalty for slightly expensive players
		case priceMultiplier <= 0.5:
			valuePenalty = 1.5 // 50% bonus for undervalued players
		case priceMultiplier <= 0.7:
			valuePenalty = 1.2 // 20% bonus for somewhat undervalued players
		default:
			valuePenalty = 1.0 // No penalty/bonus for reasonably priced players
		}

		// Use logarithmic scaling to reduce the penalty for expensive but valuable players
		logValuePerRating := math.Log10(valuePerRating + 1) // +1 to avoid log(0)

		// Base efficiency score: higher overall rating and lower value-per-rating is better
		baseEfficiency := overall / (logValuePerRating + 1)

		// Apply tier-based multipliers to maintain some differentiation
		switch {
		case overall >= 80:
			// Elite players (80+) - Expect premium pricing, moderate penalty for cost
			calculatedValueScore = baseEfficiency * 1.2
		case overall >= 70:
			// Quality players (70-79) - Good balance of quality and value
			calculatedValueScore = baseEfficiency * 1.0
		case overall >= 60:
			// Decent players (60-69) - Should be better value for money
			calculatedValueScore = baseEfficiency * 0.9
		case overall >= 55:
			// Budget players (55-59) - Expected to be cheap
			calculatedValueScore = baseEfficiency * 0.8
		default:
			// Youth/development players (<55) - Penalized for poor current ability
			calculatedValueScore = baseEfficiency * 0.6
		}

		// Apply the value penalty/bonus based on pricing
		calculatedValueScore *= valuePenalty

		// Apply bonus for exceptional value scenarios (only for reasonably priced players)
		if priceMultiplier <= 1.5 {
			if valuePerRating < expectedValuePerRating*0.7 { // 30% below expected
				calculatedValueScore *= 1.3 // 30% bonus for exceptional value
			} else if valuePerRating < expectedValuePerRating*0.85 { // 15% below expected
				calculatedValueScore *= 1.15 // 15% bonus for good value
			}
		}
	}

	// Value score contribution (reduced multiplier from 0.75 to 0.5)
	valueScoreContribution := int(calculatedValueScore * 0.5)

	// Calculate transfer value penalty for expensive players
	// This is currency-agnostic and scales automatically
	var transferValuePenalty int
	if transferValueMillions > 0 {
		// Calculate value-per-rating ratio (millions per overall point)
		valuePerRating := transferValueMillions / overall

		// Get expected value for this overall rating
		expectedValuePerRating := getExpectedValuePerRating(overall)

		// Calculate how many times more expensive they are than expected
		priceMultiplier := valuePerRating / expectedValuePerRating

		// Apply penalty based on how overpriced they are
		switch {
		case priceMultiplier >= 5.0:
			// Extremely overpriced (5x or more expected value)
			transferValuePenalty = -40
		case priceMultiplier >= 4.0:
			// Severely overpriced (4x expected value)
			transferValuePenalty = -30
		case priceMultiplier >= 3.0:
			// Very overpriced (3x expected value)
			transferValuePenalty = -20
		case priceMultiplier >= 2.5:
			// Overpriced (2.5x expected value)
			transferValuePenalty = -15
		case priceMultiplier >= 2.0:
			// Somewhat overpriced (2x expected value)
			transferValuePenalty = -10
		case priceMultiplier >= 1.5:
			// Slightly overpriced (1.5x expected value)
			transferValuePenalty = -5
		case priceMultiplier <= 0.5:
			// Undervalued (0.5x or less expected value) - bonus
			transferValuePenalty = 5
		case priceMultiplier <= 0.7:
			// Somewhat undervalued (0.7x expected value) - small bonus
			transferValuePenalty = 2
		default:
			// Reasonably priced (0.7x to 1.5x expected value)
			transferValuePenalty = 0
		}
	} else {
		transferValuePenalty = 0
	}

	// Calculate salary penalty based on overpayment
	salaryPenalty := getSalaryPenalty(player.TransferValueAmount, player.WageAmount)

	// Final calculation
	moneyballRating := baseRating + ageModifier + mentalityModifier + valueScoreContribution + transferValuePenalty + salaryPenalty

	// Store raw MBR for relative scaling (will be normalized later)
	return moneyballRating
}

// getSalaryPenalty calculates a penalty for players who are overpaid relative to their transfer value
func getSalaryPenalty(transferValueAmount, wageAmount int64) int {
	// Skip if no salary data
	if wageAmount == 0 {
		return 0
	}

	// Convert to millions for easier calculation
	transferValueMillions := float64(transferValueAmount) / 1000000.0
	wageThousandsPerWeek := float64(wageAmount) / 1000.0

	// Calculate expected salary based on transfer value
	// General rule: annual salary should be roughly 10-15% of transfer value
	// Weekly salary = (transfer value * 0.12) / 52 weeks
	expectedWeeklySalary := (transferValueMillions * 0.12 * 1000) / 52 // Convert to thousands per week

	// Calculate salary ratio (actual vs expected)
	salaryRatio := wageThousandsPerWeek / expectedWeeklySalary

	// Apply penalties based on overpayment
	switch {
	case salaryRatio >= 3.0:
		// Severely overpaid (3x or more expected salary)
		return -15
	case salaryRatio >= 2.5:
		// Very overpaid (2.5x expected salary)
		return -10
	case salaryRatio >= 2.0:
		// Overpaid (2x expected salary)
		return -7
	case salaryRatio >= 1.5:
		// Somewhat overpaid (1.5x expected salary)
		return -3
	case salaryRatio <= 0.5:
		// Underpaid (0.5x or less expected salary) - bonus
		return 5
	case salaryRatio <= 0.7:
		// Somewhat underpaid (0.7x expected salary) - small bonus
		return 2
	default:
		// Reasonably paid (0.7x to 1.5x expected salary)
		return 0
	}
}

// getAgeModifier returns the age modifier for Moneyball Rating calculation
func getAgeModifier(age int) int {
	switch {
	case age >= 16 && age <= 18:
		return 30
	case age == 19:
		return 25
	case age == 20:
		return 22
	case age == 21:
		return 18
	case age == 22:
		return 15
	case age == 23:
		return 10
	case age == 24:
		return 6
	case age == 25:
		return 3
	case age == 26:
		return -1
	case age == 27:
		return -3
	case age == 28:
		return -5
	case age == 29:
		return -10
	case age == 30:
		return -15
	case age == 31:
		return -20
	default:
		// For ages outside the specified range, apply a gradual decline
		if age > 31 {
			return -20 - (age-31)*2 // Additional -2 per year after 31
		}
		return 0 // Default for ages below 16
	}
}

// getMentalityModifier returns the mentality modifier based on personality
func getMentalityModifier(personality string) int {
	// Elite personalities
	switch personality {
	case "Model Citizen":
		return 20
	case "Model Professional":
		return 20
	}

	// Very Good personalities
	switch personality {
	case "Perfectionist", "Resolute", "Professional", "Fairly Professional",
		"Iron Willed", "Resillient", "Spirited", "Driven", "Determined",
		"Fairly Determined", "Charismatic Leader", "Born Leader", "Leader",
		"Very Ambitious", "Ambitious", "Fairly Ambitious":
		return 15
	}

	// Good personalities
	switch personality {
	case "Balanced", "Light-Hearted", "Jovial", "Very Loyal", "Loyal",
		"Fairly Loyal", "Honest", "Sporting", "Fairly Sporting":
		return 8
	}

	// Okay personalities (neutral)
	switch personality {
	case "Balanced", "Light-Hearted", "Jovial", "Very Loyal", "Loyal",
		"Fairly Loyal", "Honest", "Sporting", "Fairly Sporting":
		return 0
	}

	// Poor personalities
	switch personality {
	case "Fickle", "Mercenary", "Unambitious", "Unsporting", "Realist":
		return -15
	}

	// Very Poor personalities
	switch personality {
	case "Slack", "Casual", "Temperamental", "Spineless", "Low Self-Belief",
		"Easily Discouraged", "Low Determination":
		return -35
	}

	// Default for unknown personalities
	return 0
}

// NormalizeMBRScoresRelativeToMax normalizes all MBR scores relative to the maximum MBR value.
// The player with the highest MBR gets 100, and all others are scaled proportionally.
// For example: if max MBR is 150, a player with MBR 135 gets 90 (135/150 * 100).
func NormalizeMBRScoresRelativeToMax(players []Player) {
	if len(players) == 0 {
		return
	}

	// First pass: Find the maximum MBR value
	maxMBR := players[0].MBR
	for i := range players {
		if players[i].MBR > maxMBR {
			maxMBR = players[i].MBR
		}
	}

	// Handle edge case where max MBR is 0 or negative
	if maxMBR <= 0 {
		// If all players have 0 or negative MBR, set them all to 0
		for i := range players {
			players[i].MBR = 0
		}
		return
	}

	// Second pass: Normalize all MBR scores relative to the maximum
	for i := range players {
		// Calculate the relative score: (player_score / max_score) * 100
		normalizedMBR := int((float64(players[i].MBR) / float64(maxMBR)) * 100.0)

		// Ensure the result is at least 0
		if normalizedMBR < 0 {
			normalizedMBR = 0
		}

		players[i].MBR = normalizedMBR
	}
}
