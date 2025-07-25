/**
 * Player utility functions for data processing and analysis
 */

/**
 * Get the division of a player, handling different field names
 * @param {Object} player - The player object
 * @returns {string} The player's division or 'N/A' if not found
 */
export const getPlayerDivision = player => {
  return player.division || player.Division || 'N/A'
}

/**
 * Convert a value to a numeric value, handling various formats
 * @param {any} val - The value to convert
 * @returns {number|null} The numeric value or null if invalid
 */
export const getNumericValue = val => {
  if (val === undefined || val === null || val === '-' || val === '') return null
  const cleaned = String(val).replace(/,/g, '').replace(/%/g, '')
  const num = parseFloat(cleaned)
  return Number.isNaN(num) ? null : num
}

/**
 * Get a player's positions as an array
 * @param {Object} player - The player object
 * @returns {Array<string>} Array of player positions
 */
export const getPlayerPositions = player => {
  return player.shortPositions || player.short_positions || []
}

/**
 * Check if a player matches a specific position
 * @param {Object} player - The player object
 * @param {string} position - The position to check
 * @returns {boolean} True if player matches the position
 */
export const playerMatchesPosition = (player, position) => {
  const positions = getPlayerPositions(player)
  return positions.includes(position)
}

/**
 * Get a player's overall rating
 * @param {Object} player - The player object
 * @returns {number} The player's overall rating
 */
export const getPlayerOverall = player => {
  return player.Overall || player.overall || 0
}

/**
 * Get a player's age as a number
 * @param {Object} player - The player object
 * @returns {number} The player's age
 */
export const getPlayerAge = player => {
  return parseInt(player.age || player.Age || 0, 10)
}

/**
 * Get a player's transfer value as a number
 * @param {Object} player - The player object
 * @returns {number} The player's transfer value
 */
export const getPlayerTransferValue = player => {
  return parseInt(player.transferValueAmount || player.TransferValueAmount || 0, 10)
}

/**
 * Get a player's wage as a number
 * @param {Object} player - The player object
 * @returns {number} The player's wage
 */
export const getPlayerWage = player => {
  return parseInt(player.wageAmount || player.WageAmount || 0, 10)
} 