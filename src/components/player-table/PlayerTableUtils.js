import { memoize } from '../../composables/useMemoization'
import { formatCurrency } from '../../utils/currencyUtils'

// GK stat mapping for both display and sorting consistency
const gkStatMapping = {
  pac: 'div', // Diving -> Pace
  sho: 'han', // Handling -> Shooting
  pas: 'kic', // Kicking -> Passing
  dri: 'ref', // Reflexes -> Dribbling
  def: 'spd', // Speed -> Defending
  phy: 'pos', // Positioning -> Physical
}

/**
 * Gets the player value for a field, handling GK stat mapping
 * @param {Object} player - Player object
 * @param {String} fieldKey - Field key
 * @param {String} columnName - Column name
 * @param {Boolean} isGoalkeeperView - Whether in goalkeeper view
 * @returns {*} Player value for the field
 */
export function getPlayerValue(player, fieldKey, _columnName = null, isGoalkeeperView = false) {
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

/**
 * Memoized version of getPlayerValue for non-Overall fields only
 */
export const getPlayerValueMemoized = memoize(
  (player, fieldKey, columnName = null, isGoalkeeperView = false, _cacheGeneration = 0) => {
    return getPlayerValue(player, fieldKey, columnName, isGoalkeeperView)
  },
  {
    maxSize: 1000,
    keyGenerator: (player, fieldKey, columnName, isGoalkeeperView, cacheGeneration) => {
      // Try to use the player's UID for cache key
      let playerUID = player.UID || player.uid

      // If no UID available or UID is empty, create a composite unique key
      if (!playerUID || playerUID === '') {
        playerUID = `${player.name || 'unknown'}-${player.club || 'unknown'}-${player.age || 'unknown'}-${player.position || 'unknown'}`
      }

      return `gen${cacheGeneration}-${playerUID}-${fieldKey}-${columnName || ''}-${isGoalkeeperView}`
    },
    cacheKey: 'playerValue',
  }
)

/**
 * Gets the display value for a player and column
 * @param {Object} player - Player object
 * @param {Object} col - Column definition
 * @param {Boolean} isGoalkeeperView - Whether in goalkeeper view
 * @param {Number} cacheGeneration - Cache generation number
 * @returns {*} Display value
 */
export function getDisplayValue(player, col, isGoalkeeperView = false, cacheGeneration = 0) {
  // For Overall field, always use non-memoized version to ensure reactivity
  if (col.field === 'Overall' || col.name === 'Overall') {
    return getPlayerValue(player, col.field, col.name, isGoalkeeperView)
  }
  // For other fields, use memoized version for performance
  return getPlayerValueMemoized(player, col.field, col.name, isGoalkeeperView, cacheGeneration)
}

/**
 * Memoized rating class calculation (called frequently in table rendering)
 */
export const getUnifiedRatingClass = memoize(
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

/**
 * Gets the money class for a numeric amount
 * @param {Number} numericAmount - Numeric amount
 * @returns {String} CSS class
 */
export function getMoneyClass(numericAmount) {
  if (numericAmount === null || numericAmount === undefined) return 'money-na'
  return 'money-uniform'
}

/**
 * Gets the value score class
 * @param {Number} valueScore - Value score
 * @returns {String} CSS class
 */
export function getValueScoreClass(valueScore) {
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

/**
 * Formats display currency
 * @param {Number} numericAmount - Numeric amount
 * @param {String} currencySymbol - Currency symbol
 * @param {String} originalDisplayValue - Original display value
 * @returns {String} Formatted currency
 */
export function formatDisplayCurrency(numericAmount, currencySymbol, originalDisplayValue) {
  return formatCurrency(numericAmount, currencySymbol, originalDisplayValue)
}

/**
 * Handles flag error by hiding the flag and showing placeholder
 * @param {Event} event - Error event
 */
export function onFlagError(event) {
  if (event.target) event.target.style.display = 'none'
  const placeholderIcon = event.target.nextElementSibling
  if (placeholderIcon?.classList.contains('q-icon')) {
    placeholderIcon.style.display = 'inline-flex'
  }
}

/**
 * Position sort order for consistent sorting
 */
export const positionSortOrder = [
  'GK',
  'DR',
  'DC',
  'DL',
  'WBR',
  'WBL',
  'DM',
  'MR',
  'MC',
  'ML',
  'AMR',
  'AMC',
  'AML',
  'ST',
]

/**
 * Memoized position index calculation (expensive string processing)
 */
export const getPositionIndex = memoize(
  (positionString) => {
    if (!positionString || typeof positionString !== 'string') {
      return positionSortOrder.length + 2 // Place invalid/empty last
    }
    let processedString = positionString.toUpperCase()
    processedString = processedString.replace(/\bST\s*\(C\)/g, 'ST')
    processedString = processedString.replace(/\bM\s*\(C\)/g, 'MC')
    processedString = processedString.replace(/\bAM\s*\(C\)/g, 'AMC')
    processedString = processedString.replace(/\bDM\s*\(C\)/g, 'DM')
    processedString = processedString.replace(/\bD\s*\(C\)/g, 'DC')
    processedString = processedString.replace(/\bGK\s*\(C\)/g, 'GK')

    let minFoundIndex = positionSortOrder.length
    const sideMatch = processedString.match(/\(([^)]+)\)$/)
    let mainPart = processedString
    const sidesSpecified = []

    if (sideMatch?.[1]) {
      mainPart = processedString.substring(0, sideMatch.index).trim()
      const sideSpec = sideMatch[1]
      if (sideSpec.includes('R')) sidesSpecified.push('R')
      if (sideSpec.includes('L')) sidesSpecified.push('L')
    }

    mainPart = mainPart.replace(/\s*\(.*?\)\s*/g, '').trim()
    const basePositionCodes = mainPart
      .split(/[,/]/)
      .map((p) => p.trim())
      .filter((p) => p.length > 0)
    const rolesToEvaluate = new Set()

    for (const baseCode of basePositionCodes) {
      if (sidesSpecified.length > 0) {
        for (const side of sidesSpecified) {
          rolesToEvaluate.add(baseCode + side)
        }
      }
      rolesToEvaluate.add(baseCode)
    }

    if (rolesToEvaluate.size === 0 && positionString.trim() !== '') {
      rolesToEvaluate.add(processedString.replace(/\s*\(.*?\)\s*/g, '').trim())
    }
    if (rolesToEvaluate.size === 0) return positionSortOrder.length + 1

    for (const role of rolesToEvaluate) {
      const index = positionSortOrder.indexOf(role)
      if (index !== -1 && index < minFoundIndex) {
        minFoundIndex = index
      }
    }
    return minFoundIndex === positionSortOrder.length ? positionSortOrder.length + 1 : minFoundIndex
  },
  {
    maxSize: 100, // Cache position calculations
    keyGenerator: (positionString) => positionString,
    cacheKey: 'positionIndex',
  }
)
