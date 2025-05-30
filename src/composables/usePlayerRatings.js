import { computed } from 'vue'

export function usePlayerRatings() {
  // Unified rating color system for 1-100 scale
  const getUnifiedRatingClass = (rating, maxScale = 100) => {
    if (!rating && rating !== 0) return 'rating-na'

    const numValue = Number.parseInt(rating, 10)
    if (Number.isNaN(numValue)) return 'rating-na'

    // Convert to percentage if different scale
    const percentage = maxScale === 100 ? numValue : (numValue / maxScale) * 100

    if (percentage >= 90) return 'rating-90'
    if (percentage >= 85) return 'rating-85'
    if (percentage >= 80) return 'rating-80'
    if (percentage >= 75) return 'rating-75'
    if (percentage >= 70) return 'rating-70'
    if (percentage >= 65) return 'rating-65'
    if (percentage >= 60) return 'rating-60'
    if (percentage >= 55) return 'rating-55'
    if (percentage >= 50) return 'rating-50'
    return 'rating-below-50'
  }

  // Alternative rating system for 1-20 scale (FM attributes)
  const getFMRatingClass = rating => {
    if (!rating && rating !== 0) return 'rating-na'

    const numValue = Number.parseInt(rating, 10)
    if (Number.isNaN(numValue)) return 'rating-na'

    if (numValue >= 18) return 'rating-tier-6'
    if (numValue >= 15) return 'rating-tier-5'
    if (numValue >= 12) return 'rating-tier-4'
    if (numValue >= 9) return 'rating-tier-3'
    if (numValue >= 6) return 'rating-tier-2'
    if (numValue >= 3) return 'rating-tier-1'
    return 'rating-tier-0'
  }

  // Attribute mapping from short names to full names
  const attributeFullNameMap = {
    COR: 'Corners',
    CRO: 'Crossing',
    DRI: 'Dribbling',
    FIN: 'Finishing',
    FRE: 'Free Kicks',
    HEA: 'Heading',
    LON: 'Long Shots',
    LTH: 'Long Throws',
    MAR: 'Marking',
    PAS: 'Passing',
    PEN: 'Penalty Taking',
    TAC: 'Tackling',
    TEC: 'Technique',
    AGG: 'Aggression',
    ANT: 'Anticipation',
    BRA: 'Bravery',
    COM: 'Composure',
    CON: 'Concentration',
    DEC: 'Decisions',
    DET: 'Determination',
    FLA: 'Flair',
    LEA: 'Leadership',
    OTB: 'Off The Ball',
    POS: 'Positioning',
    TEA: 'Teamwork',
    VIS: 'Vision',
    WOR: 'Work Rate',
    ACC: 'Acceleration',
    AGI: 'Agility',
    BAL: 'Balance',
    JUM: 'Jumping Reach',
    NAT: 'Natural Fitness',
    PAC: 'Pace',
    STA: 'Stamina',
    STR: 'Strength',
    // FIFA stats
    DIV: 'Diving',
    HAN: 'Handling',
    KIC: 'Kicking',
    REF: 'Reflexes',
    SPD: 'Speed (GK)',
    PHY: 'Physical',
    DEF: 'Defending',
    SHO: 'Shooting'
  }

  // Attribute groupings for better organization
  const attributeGroups = computed(() => ({
    technical: {
      name: 'Technical',
      attrs: [
        'COR',
        'CRO',
        'DRI',
        'FIN',
        'FRE',
        'HEA',
        'LON',
        'LTH',
        'MAR',
        'PAS',
        'PEN',
        'TAC',
        'TEC'
      ]
    },
    mental: {
      name: 'Mental',
      attrs: [
        'AGG',
        'ANT',
        'BRA',
        'COM',
        'CON',
        'DEC',
        'DET',
        'FLA',
        'LEA',
        'OTB',
        'POS',
        'TEA',
        'VIS',
        'WOR'
      ]
    },
    physical: {
      name: 'Physical',
      attrs: ['ACC', 'AGI', 'BAL', 'JUM', 'NAT', 'PAC', 'STA', 'STR']
    },
    goalkeeper: {
      name: 'Goalkeeper',
      attrs: ['DIV', 'HAN', 'KIC', 'REF', 'SPD']
    },
    fifa: {
      name: 'FIFA',
      attrs: ['PAC', 'SHO', 'PAS', 'DRI', 'DEF', 'PHY']
    }
  }))

  // Calculate weighted average rating
  const calculateWeightedRating = (attributes, weights) => {
    let totalWeight = 0
    let weightedSum = 0

    for (const [attr, weight] of Object.entries(weights)) {
      if (attributes[attr] !== undefined && attributes[attr] !== null) {
        weightedSum += attributes[attr] * weight
        totalWeight += weight
      }
    }

    return totalWeight > 0 ? Math.round(weightedSum / totalWeight) : 0
  }

  // Get rating color based on value
  const getRatingColor = (rating, maxScale = 100) => {
    const ratingClass = getUnifiedRatingClass(rating, maxScale)
    const colorMap = {
      'rating-90': '#00ff00', // Bright green
      'rating-85': '#7fff00', // Chartreuse
      'rating-80': '#ffff00', // Yellow
      'rating-75': '#ffd700', // Gold
      'rating-70': '#ffa500', // Orange
      'rating-65': '#ff6347', // Tomato
      'rating-60': '#ff4500', // Orange red
      'rating-55': '#ff0000', // Red
      'rating-50': '#dc143c', // Crimson
      'rating-below-50': '#8b0000', // Dark red
      'rating-na': '#808080' // Gray
    }
    return colorMap[ratingClass] || '#808080'
  }

  // Format rating for display
  const formatRating = (rating, maxScale = 100, showScale = false) => {
    if (!rating && rating !== 0) return '-'

    const numValue = Number.parseInt(rating, 10)
    if (Number.isNaN(numValue)) return '-'

    if (showScale) {
      return `${numValue}/${maxScale}`
    }
    return numValue.toString()
  }

  return {
    getUnifiedRatingClass,
    getFMRatingClass,
    attributeFullNameMap,
    attributeGroups,
    calculateWeightedRating,
    getRatingColor,
    formatRating
  }
}
