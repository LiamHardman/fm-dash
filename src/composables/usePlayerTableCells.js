import { formatCurrency } from '../utils/currencyUtils'
import { memoize } from './useMemoization'

export function usePlayerTableCells(isGoalkeeperView, currencySymbol, cacheGeneration) {
  // GK stat mapping for both display and sorting consistency
  const gkStatMapping = {
    pac: 'div', // Diving -> Pace
    sho: 'han', // Handling -> Shooting
    pas: 'kic', // Kicking -> Passing
    dri: 'ref', // Reflexes -> Dribbling
    def: 'spd', // Speed -> Defending
    phy: 'pos', // Positioning -> Physical
  }

  // Memoized rating class calculation (called frequently in table rendering)
  const getUnifiedRatingClass = memoize(
    (value, maxScale) => {
      const numValue = Number.parseInt(value, 10)
      if (Number.isNaN(numValue) || value === null || value === undefined || value === '-')
        return 'rating-na'
      const percentage = (numValue / maxScale) * 100
      if (percentage >= 90) return 'rating-tier-6'
      if (percentage >= 80) return 'rating-tier-5'
      if (percentage >= 70) return 'rating-tier-4'
      if (percentage >= 55) return 'rating-tier-3'
      if (percentage >= 40) return 'rating-tier-2'
      return 'rating-tier-1'
    },
    {
      maxSize: 200, // Cache up to 200 different rating calculations
      keyGenerator: (value, maxScale) => `${value}-${maxScale}`,
      cacheKey: 'unifiedRatingClass',
    }
  )

  // Custom rating class for TotalStats with specific ranges
  const getTotalStatsRatingClass = memoize(
    (value) => {
      const numValue = Number.parseInt(value, 10)
      if (Number.isNaN(numValue) || value === null || value === undefined || value === '-')
        return 'rating-na'

      // Custom ranges for TotalStats
      if (numValue >= 520) return 'rating-tier-6' // Elite
      if (numValue >= 470) return 'rating-tier-5' // Very good
      if (numValue >= 430) return 'rating-tier-4' // Good
      if (numValue >= 390) return 'rating-tier-3' // Average
      if (numValue >= 350) return 'rating-tier-2' // Below average
      return 'rating-tier-1' // Poor
    },
    {
      maxSize: 200,
      keyGenerator: (value) => `totalStats-${value}`,
      cacheKey: 'totalStatsRatingClass',
    }
  )

  const getMoneyClass = (numericAmount) => {
    if (numericAmount === null || numericAmount === undefined) return 'money-na'
    return 'money-uniform'
  }

  const getValueScoreClass = (valueScore) => {
    if (valueScore === null || valueScore === undefined) return 'rating-na'
    const score = Number(valueScore)
    if (Number.isNaN(score)) return 'rating-na'

    if (score >= 80) return 'rating-tier-6' // Excellent value - highest tier
    if (score >= 60) return 'rating-tier-5' // Great value
    if (score >= 40) return 'rating-tier-4' // Good value
    if (score >= 20) return 'rating-tier-3' // Fair value
    if (score >= 0) return 'rating-tier-2' // Poor value
    return 'rating-na'
  }

  const onFlagError = (event) => {
    if (event.target) event.target.style.display = 'none'
    const placeholderIcon = event.target.nextElementSibling
    if (placeholderIcon?.classList.contains('q-icon')) {
      placeholderIcon.style.display = 'inline-flex'
    }
  }

  const formatDisplayCurrency = (numericAmount, originalDisplayValue) => {
    return formatCurrency(numericAmount, currencySymbol, originalDisplayValue)
  }

  // Main player value getter for both display and sorting
  const getPlayerValue = (player, fieldKey, _columnName = null) => {
    // For regular view, map GK stats to standard FIFA stats if the player is a goalkeeper
    if (!isGoalkeeperView && player.position && player.position.includes('GK')) {
      const mappedStat = gkStatMapping[fieldKey]
      if (mappedStat && player[mappedStat] !== undefined) {
        return player[mappedStat]
      }
    }

    // For goalkeeper view, all players should show goalkeeper stats
    if (isGoalkeeperView) {
      // Map outfield FIFA stats to goalkeeper stats
      const gkStatMappingReverse = {
        pac: 'div', // Pace -> Diving
        sho: 'han', // Shooting -> Handling
        pas: 'kic', // Passing -> Kicking
        dri: 'ref', // Dribbling -> Reflexes
        def: 'spd', // Defending -> Speed
        phy: 'pos', // Physical -> Positioning
      }
      const mappedStat = gkStatMappingReverse[fieldKey]
      if (mappedStat && player[mappedStat] !== undefined) {
        return player[mappedStat]
      }
    }

    // Default behavior - use the field key
    return player[fieldKey]
  }

  // Memoized version for non-Overall fields only
  const getPlayerValueMemoized = memoize(
    (player, fieldKey, columnName = null) => {
      return getPlayerValue(player, fieldKey, columnName)
    },
    {
      maxSize: 1000,
      keyGenerator: (player, fieldKey, columnName) => {
        // Try to use the player's UID for cache key
        let playerUID = player.UID || player.uid

        // If no UID available or UID is empty, create a composite unique key
        if (!playerUID || playerUID === '') {
          playerUID = `${player.name || 'unknown'}-${player.club || 'unknown'}-${player.age || 'unknown'}-${player.position || 'unknown'}`
        }

        return `gen${cacheGeneration.value}-${playerUID}-${fieldKey}-${columnName || ''}`
      },
      cacheKey: 'playerValue',
    }
  )

  const getDisplayValue = (player, col) => {
    // For Overall field, always use non-memoized version to ensure reactivity
    if (col.field === 'Overall' || col.name === 'Overall') {
      return getPlayerValue(player, col.field, col.name)
    }
    // For other fields, use memoized version for performance
    return getPlayerValueMemoized(player, col.field, col.name)
  }

  const clearCaches = () => {
    getUnifiedRatingClass.clearCache()
    getTotalStatsRatingClass.clearCache()
    getPlayerValueMemoized.clearCache()
  }

  return {
    getUnifiedRatingClass,
    getTotalStatsRatingClass,
    getMoneyClass,
    getValueScoreClass,
    onFlagError,
    formatDisplayCurrency,
    getPlayerValue,
    getDisplayValue,
    clearCaches,
  }
}
