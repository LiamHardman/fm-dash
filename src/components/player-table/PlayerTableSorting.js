import { nextTick } from 'vue'
import { getPlayerValue, getPositionIndex } from './PlayerTableUtils'

const MAX_DISPLAY_PLAYERS = 1000

/**
 * Custom sort function for complex sorting logic
 * @param {Object} a - First player
 * @param {Object} b - Second player
 * @param {String} field - Field to sort by
 * @param {String} dir - Sort direction
 * @param {String} sortField - Sort field name
 * @param {Boolean} isGoalkeeperView - Whether in goalkeeper view
 * @returns {Number} Sort result
 */
export function customSortFn(a, b, field, dir, sortField, isGoalkeeperView = false) {
  // Use getPlayerValue to ensure GK stat mapping is applied for sorting
  let vA = getPlayerValue(a, field, sortField, isGoalkeeperView)
  let vB = getPlayerValue(b, field, sortField, isGoalkeeperView)
  const aIsNull = vA === null || vA === undefined
  const bIsNull = vB === null || vB === undefined

  if (aIsNull && bIsNull) return 0
  if (aIsNull) return dir === 'asc' ? 1 : -1
  if (bIsNull) return dir === 'asc' ? -1 : 1

  if (field === 'position') {
    const indexA = getPositionIndex(vA)
    const indexB = getPositionIndex(vB)
    return dir === 'asc' ? indexA - indexB : indexB - indexA
  }
  if (typeof vA === 'number' && typeof vB === 'number') {
    return dir === 'asc' ? vA - vB : vB - vA
  }
  if (typeof vA === 'string' && typeof vB === 'string') {
    vA = vA.toLowerCase()
    vB = vB.toLowerCase()
    if (vA < vB) return dir === 'asc' ? -1 : 1
    if (vA > vB) return dir === 'asc' ? 1 : -1
    return 0
  }
  return 0
}

/**
 * Async sorting for large datasets using web workers
 * @param {Array} players - Players array
 * @param {String} fieldKey - Field key to sort by
 * @param {String} direction - Sort direction
 * @param {String} sortField - Sort field name
 * @param {Boolean} isGoalkeeperView - Whether in goalkeeper view
 * @param {Function} customSortFn - Custom sort function
 * @param {String} sortKey - Sort key for caching
 * @param {Object} sortController - Sort controller for cancellation
 * @param {Object} sortingUtils - Sorting utilities (sortLargeArray, sortPlayersWorker)
 * @returns {Promise<Object>} Sorting result with data and metadata
 */
export async function triggerAsyncSort(
  players,
  fieldKey,
  direction,
  sortField,
  isGoalkeeperView,
  customSortFn,
  sortKey,
  sortController,
  sortingUtils
) {
  // Prevent concurrent async sorting operations
  if (sortController.isAsyncSorting) {
    return
  }

  // Cancel any previous sort operation
  if (sortController.currentSortController) {
    sortController.currentSortController.cancelled = true
  }

  // Create new controller for this sort operation
  const controller = { cancelled: false }
  sortController.currentSortController = controller

  // Set async sorting flag AFTER ensuring cache stability
  sortController.isAsyncSorting = true

  try {
    let fullSortedList
    const playerCount = players.length

    // Use passed sorting utilities
    const { sortLargeArray, sortPlayersWorker } = sortingUtils

    // Tiered sorting strategy based on dataset size
    if (playerCount >= 2000) {
      try {
        fullSortedList = await sortPlayersWorker(
          [...players],
          fieldKey,
          direction,
          sortField,
          isGoalkeeperView
        )
      } catch (_workerError) {
        // Fallback to main thread if Web Worker fails
        fullSortedList = await sortLargeArray([...players], fieldKey, direction, customSortFn, 2000)
      }
    } else {
      fullSortedList = await sortLargeArray([...players], fieldKey, direction, customSortFn, 2000)
    }

    // Check if this sort was cancelled
    if (controller.cancelled) {
      return { cancelled: true }
    }

    let result
    let sliced = false

    if (fullSortedList.length > MAX_DISPLAY_PLAYERS) {
      sliced = true
      result = fullSortedList.slice(0, MAX_DISPLAY_PLAYERS)
    } else {
      result = fullSortedList
    }

    return {
      result,
      sliced,
      totalCount: fullSortedList.length,
      cancelled: false,
    }
  } catch (_error) {
    if (controller.cancelled) {
      return { cancelled: true }
    }

    // Show user-friendly error notification
    const errorMessage = 'Sorting was interrupted. Showing unsorted data.'

    // Fallback to direct assignment if async sorting fails
    const fallbackResult = [...players].slice(0, MAX_DISPLAY_PLAYERS)

    return {
      result: fallbackResult,
      sliced: players.length > MAX_DISPLAY_PLAYERS,
      totalCount: players.length,
      errorMessage,
      cancelled: false,
    }
  } finally {
    // Always clear the sorting flag and controller, regardless of outcome
    sortController.isAsyncSorting = false
    sortController.currentSortController = null
  }
}

/**
 * Synchronous sorting for small datasets
 * @param {Array} players - Players array
 * @param {String} fieldKey - Field key to sort by
 * @param {String} direction - Sort direction
 * @param {String} sortField - Sort field name
 * @param {Boolean} isGoalkeeperView - Whether in goalkeeper view
 * @returns {Object} Sorting result with data and metadata
 */
export function syncSort(players, fieldKey, direction, sortField, isGoalkeeperView) {
  const playersToSort = [...players]
  const fullSortedList = playersToSort.sort((a, b) =>
    customSortFn(a, b, fieldKey, direction, sortField, isGoalkeeperView)
  )

  let result
  let sliced = false

  if (fullSortedList.length > MAX_DISPLAY_PLAYERS) {
    sliced = true
    result = fullSortedList.slice(0, MAX_DISPLAY_PLAYERS)
  } else {
    result = fullSortedList
  }

  return {
    result,
    sliced,
    totalCount: fullSortedList.length,
  }
}

/**
 * Gets the sort field key for a column
 * @param {Array} currentColumns - Current columns configuration
 * @param {String} colName - Column name
 * @returns {String} Sort field key
 */
export function getSortFieldKey(currentColumns, colName) {
  const colDef = currentColumns.find((c) => c.name === colName)
  return colDef?.sortField || colDef?.field || colName
}

/**
 * Determines if sorting should be async based on dataset size
 * @param {Number} playerCount - Number of players
 * @returns {Boolean} Whether to use async sorting
 */
export function shouldUseAsyncSort(playerCount) {
  return playerCount > 500
}

export { MAX_DISPLAY_PLAYERS }
