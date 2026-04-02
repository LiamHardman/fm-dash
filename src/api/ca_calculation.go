package main

import "math"

// casAttrOrder defines the exact sequence of FM attribute keys used by the fm21-cas weight tables.
// This mirrors the order in Player.setAttrsWeight() from the fm21-cas Android app:
//   - Indices 0–10:  Goalkeeping (Aer, Cmd, Com, Ecc, Han, Kic, 1v1, Pun, Ref, TRO, Thr)
//   - Indices 11–24: Technical   (Cor, Cro, Dri, Fin, Fir, Fre, Hea, Lon, L Th, Mar, Pas, Pen, Tck, Tec)
//   - Indices 25–38: Mental      (Agg, Ant, Bra, Cmp, Cnt, Dec, Det, Fla, Ldr, OtB, Pos, Tea, Vis, Wor)
//   - Indices 39–47: Physical    (Acc, Agi, Bal, Jum, Nat, Pac, Sta, Str)
//
// Note: The original app included a 49th "Weaker Foot" attribute which is randomised
// based on preferred foot. fmdash does not expose this data, so it is omitted.
// All weight tables below contain 48 values, one per attribute above.
var casAttrOrder = []string{
	// Goalkeeping (0–10)
	"Aer", "Cmd", "Com", "Ecc", "Han", "Kic", "1v1", "Pun", "Ref", "TRO", "Thr",
	// Technical (11–24)
	"Cor", "Cro", "Dri", "Fin", "Fir", "Fre", "Hea", "Lon", "L Th", "Mar", "Pas", "Pen", "Tck", "Tec",
	// Mental (25–38)
	"Agg", "Ant", "Bra", "Cmp", "Cnt", "Dec", "Det", "Fla", "Ldr", "OtB", "Pos", "Tea", "Vis", "Wor",
	// Physical (39–47)
	"Acc", "Agi", "Bal", "Jum", "Nat", "Pac", "Sta", "Str",
}

// casPositionWeights maps each fmdash short position code to the fm21-cas weight array.
// Values are taken directly from the Player.setAttrsWeight() switch statement in fm21-cas.
// Zero-weight entries represent attributes that are irrelevant for that position; the
// formula replaces them with 1 (hidden attribute floor) during calculation.
//
// Each entry has 48 floats ordered by casAttrOrder above.
var casPositionWeights = map[string][]float64{
	// POS_GK
	"GK": {
		6, 6, 5, 0, 8, 5, 4, 0, 8, 0, 3, // Goalkeeping
		0, 0, 0, 0, 1, 0, 1, 0, 0, 0, 3, 0, 0, 1, // Technical
		0, 3, 6, 2, 6, 10, 0, 0, 2, 0, 5, 2, 1, 1, // Mental
		6, 8, 2, 1, 0, 3, 1, 4, // Physical (no Str mapped; last slot is Str)
	},
	// POS_DRL (Right/Left Defender) — maps to both DR and DL
	"DR": {
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 2, 1, 1, 3, 1, 2, 1, 1, 3, 2, 1, 4, 2,
		0, 3, 2, 2, 4, 7, 0, 0, 1, 1, 4, 2, 2, 2,
		7, 6, 2, 2, 0, 5, 6, 4,
	},
	"DL": {
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 2, 1, 1, 3, 1, 2, 1, 1, 3, 2, 1, 4, 2,
		0, 3, 2, 2, 4, 7, 0, 0, 1, 1, 4, 2, 2, 2,
		7, 6, 2, 2, 0, 5, 6, 4,
	},
	// POS_DC
	"DC": {
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 1, 1, 1, 2, 1, 5, 1, 1, 8, 2, 1, 5, 1,
		0, 5, 2, 2, 4, 10, 0, 0, 2, 1, 8, 1, 1, 2,
		6, 6, 2, 6, 0, 5, 3, 6,
	},
	// POS_WBRL (Wing-Back Right/Left)
	"WBR": {
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 3, 2, 1, 3, 1, 1, 1, 1, 2, 3, 1, 3, 3,
		0, 3, 1, 2, 3, 5, 0, 0, 1, 2, 3, 2, 2, 2,
		8, 5, 2, 1, 0, 6, 7, 4,
	},
	"WBL": {
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 3, 2, 1, 3, 1, 1, 1, 1, 2, 3, 1, 3, 3,
		0, 3, 1, 2, 3, 5, 0, 0, 1, 2, 3, 2, 2, 2,
		8, 5, 2, 1, 0, 6, 7, 4,
	},
	// POS_DM
	"DM": {
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 1, 2, 2, 4, 1, 1, 3, 1, 3, 4, 1, 7, 3,
		0, 5, 1, 2, 3, 8, 0, 0, 1, 1, 5, 2, 4, 4,
		6, 6, 2, 1, 0, 4, 4, 5,
	},
	// POS_MRL (Right/Left Midfielder)
	"MR": {
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 5, 3, 2, 4, 1, 1, 2, 1, 1, 3, 1, 2, 4,
		0, 3, 1, 3, 2, 5, 0, 0, 1, 2, 1, 2, 3, 3,
		8, 6, 2, 1, 0, 6, 5, 3,
	},
	"ML": {
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 5, 3, 2, 4, 1, 1, 2, 1, 1, 3, 1, 2, 4,
		0, 3, 1, 3, 2, 5, 0, 0, 1, 2, 1, 2, 3, 3,
		8, 6, 2, 1, 0, 6, 5, 3,
	},
	// POS_MC
	"MC": {
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 1, 2, 2, 6, 1, 1, 3, 1, 3, 6, 1, 3, 4,
		0, 3, 1, 3, 2, 7, 0, 0, 1, 3, 3, 2, 6, 3,
		6, 6, 2, 1, 0, 5, 6, 4,
	},
	// POS_AMRL (Attacking Midfielder Right/Left)
	"AMR": {
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 5, 5, 2, 5, 1, 1, 2, 1, 1, 2, 1, 2, 4,
		0, 3, 1, 3, 2, 5, 0, 0, 1, 2, 1, 2, 3, 3,
		10, 6, 2, 1, 0, 10, 7, 3,
	},
	"AML": {
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 5, 5, 2, 5, 1, 1, 2, 1, 1, 2, 1, 2, 4,
		0, 3, 1, 3, 2, 5, 0, 0, 1, 2, 1, 2, 3, 3,
		10, 6, 2, 1, 0, 10, 7, 3,
	},
	// POS_AMC
	"AMC": {
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 1, 3, 3, 5, 1, 1, 3, 1, 1, 4, 1, 2, 5,
		0, 3, 1, 3, 2, 6, 0, 0, 1, 3, 2, 2, 6, 3,
		9, 6, 2, 1, 0, 7, 6, 3,
	},
	// POS_STC — maps to "ST" in fmdash
	"ST": {
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 2, 5, 8, 6, 1, 6, 2, 1, 1, 2, 1, 1, 4,
		0, 5, 1, 6, 2, 5, 0, 0, 1, 6, 2, 1, 2, 2,
		10, 6, 2, 5, 0, 7, 6, 6,
	},
	// Sweeper — uses DC weights as a reasonable approximation (not in original app)
	"SW": {
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 1, 1, 1, 2, 1, 5, 1, 1, 8, 2, 1, 5, 1,
		0, 5, 2, 2, 4, 10, 0, 0, 2, 1, 8, 1, 1, 2,
		6, 6, 2, 6, 0, 5, 3, 6,
	},
}

// calculateCASForPosition computes the fm21-cas CA rating for a player at a given position.
// It implements the two-pass weighted amplification formula from the fm21-cas Android app.
// Returns 0 if the position is unknown or no valid attribute data is available.
func calculateCASForPosition(playerNumericAttributes map[string]int, shortPositionKey string) int {
	weights, ok := casPositionWeights[shortPositionKey]
	if !ok || len(weights) != len(casAttrOrder) {
		return 0
	}

	n := len(casAttrOrder)

	// Build point and weight arrays:
	// - Zero-weight attributes (irrelevant for the position) are floored to 1 (hidden attr treatment)
	// - Attributes not present in the player data are set to 1 (minimum)
	points := make([]float64, n)
	effectiveWeights := make([]float64, n)

	for i, attrKey := range casAttrOrder {
		rawWeight := weights[i]
		attrValue := 1 // floor: hidden/irrelevant attrs default to 1
		if v, exists := playerNumericAttributes[attrKey]; exists && v > 0 {
			attrValue = v
		}
		// Clamp to valid FM range [1, 20]
		if attrValue < 1 {
			attrValue = 1
		}
		if attrValue > 20 {
			attrValue = 20
		}

		points[i] = float64(attrValue)

		// Zero-weight positions are treated as weight=1 for the denominator
		// (preserves the spirit of the fm21-cas hidden attribute handling)
		if rawWeight == 0 {
			effectiveWeights[i] = 1.0
		} else {
			effectiveWeights[i] = rawWeight
		}
	}

	// Pass 1: Compute sums
	var sumAttrs, sumWeights float64
	for i := 0; i < n; i++ {
		sumAttrs += points[i]
		sumWeights += effectiveWeights[i]
	}

	if sumWeights == 0 {
		return 0
	}

	// Weighted average of raw points
	avgCaWeighted := sumAttrs / sumWeights

	// Pass 2: Amplified weighted sum
	var sumAttrMod float64
	amplifier := (sumWeights + avgCaWeighted) / sumWeights
	for i := 0; i < n; i++ {
		sumAttrMod += (points[i] * effectiveWeights[i]) * amplifier
	}

	// Final scaling (×10 as per the original formula)
	finalMod := sumAttrMod / sumWeights * amplifier * 10

	result := int(math.Round(finalMod))

	// Clamp to FM CA scale (1–200); values below 1 or above 200 are pathological
	if result < 1 {
		return 1
	}
	if result > 200 {
		return 200
	}
	return result
}

// CalculateCAS computes the Current Ability (CA) estimate for a player using the fm21-cas
// algorithm. It calculates the CA for every position the player can play and returns
// the maximum score, reflecting the player's best positional ability.
//
// Output is on a 0–200 scale, mirroring Football Manager's internal CA range.
func CalculateCAS(player *Player) int {
	if len(player.ShortPositions) == 0 {
		return 0
	}

	player.mu.RLock()
	// Create a safe copy of numeric attributes to avoid concurrent map access
	attrs := make(map[string]int, len(player.NumericAttributes))
	for k, v := range player.NumericAttributes {
		attrs[k] = v
	}
	player.mu.RUnlock()

	maxCA := 0
	for _, shortPos := range player.ShortPositions {
		ca := calculateCASForPosition(attrs, shortPos)
		if ca > maxCA {
			maxCA = ca
		}
	}
	return maxCA
}
