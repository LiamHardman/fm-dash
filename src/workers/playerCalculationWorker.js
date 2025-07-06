// Worker for expensive player calculations
// This runs in a separate thread to avoid blocking the main UI thread

const gkStatMapping = {
  PAC: 'SPD',
  SHO: 'KIC',
  PAS: 'KIC',
  DRI: 'HAN',
  DEF: 'SPD',
  PHY: 'POS'
}

// Position sort order for position-based sorting
const positionSortOrder = [
  'GK',
  'DR',
  'DL',
  'DC',
  'WBR',
  'WBL',
  'DM',
  'MC',
  'MR',
  'ML',
  'AMR',
  'AMC',
  'AML',
  'ST'
]

/**
 * Get player value with GK mapping applied
 */
function getPlayerValue(player, fieldKey, columnName = null, isGoalkeeperView = false) {
  if (!isGoalkeeperView && player.position && player.position.includes('GK')) {
    const mappedStat = gkStatMapping[columnName || fieldKey]
    if (mappedStat && player[mappedStat] !== undefined) {
      return player[mappedStat]
    }
  }

  return player[fieldKey]
}

/**
 * Calculate position index for sorting (expensive string processing)
 */
function getPositionIndex(positionString) {
  if (!positionString || typeof positionString !== 'string') {
    return positionSortOrder.length + 2
  }

  let processedString = positionString.toUpperCase()
  processedString = processedString.replace(/\\bST\\s*\\(C\\)/g, 'ST')
  processedString = processedString.replace(/\\bM\\s*\\(C\\)/g, 'MC')
  processedString = processedString.replace(/\\bAM\\s*\\(C\\)/g, 'AMC')
  processedString = processedString.replace(/\\bDM\\s*\\(C\\)/g, 'DM')
  processedString = processedString.replace(/\\bD\\s*\\(C\\)/g, 'DC')
  processedString = processedString.replace(/\\bGK\\s*\\(C\\)/g, 'GK')

  let minFoundIndex = positionSortOrder.length
  const sideMatch = processedString.match(/\\(([^)]+)\\)$/)
  let mainPart = processedString
  const sidesSpecified = []

  if (sideMatch?.[1]) {
    mainPart = processedString.substring(0, sideMatch.index).trim()
    const sideSpec = sideMatch[1]
    if (sideSpec.includes('R')) sidesSpecified.push('R')
    if (sideSpec.includes('L')) sidesSpecified.push('L')
  }

  mainPart = mainPart.replace(/\\s*\\(.*?\\)\\s*/g, '').trim()
  const basePositionCodes = mainPart
    .split(/[,/]/)
    .map(p => p.trim())
    .filter(p => p.length > 0)
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
    rolesToEvaluate.add(processedString.replace(/\\s*\\(.*?\\)\\s*/g, '').trim())
  }
  if (rolesToEvaluate.size === 0) return positionSortOrder.length + 1

  for (const role of rolesToEvaluate) {
    const index = positionSortOrder.indexOf(role)
    if (index !== -1 && index < minFoundIndex) {
      minFoundIndex = index
    }
  }
  return minFoundIndex === positionSortOrder.length ? positionSortOrder.length + 1 : minFoundIndex
}

/**
 * Optimized fast sorting for large datasets
 * Uses native Array.sort() for maximum performance in Web Worker
 */
function fastSortPlayers(players, fieldKey, direction, sortField, isGoalkeeperView) {
  return players.sort((a, b) => {
    let vA = getPlayerValue(a, fieldKey, sortField, isGoalkeeperView)
    let vB = getPlayerValue(b, fieldKey, sortField, isGoalkeeperView)
    const aIsNull = vA === null || vA === undefined
    const bIsNull = vB === null || vB === undefined

    if (aIsNull && bIsNull) return 0
    if (aIsNull) return direction === 'asc' ? 1 : -1
    if (bIsNull) return direction === 'asc' ? -1 : 1

    if (fieldKey === 'position') {
      const indexA = getPositionIndex(vA)
      const indexB = getPositionIndex(vB)
      return direction === 'asc' ? indexA - indexB : indexB - indexA
    }
    if (typeof vA === 'number' && typeof vB === 'number') {
      return direction === 'asc' ? vA - vB : vB - vA
    }
    if (typeof vA === 'string' && typeof vB === 'string') {
      vA = vA.toLowerCase()
      vB = vB.toLowerCase()
      if (vA < vB) return direction === 'asc' ? -1 : 1
      if (vA > vB) return direction === 'asc' ? 1 : -1
      return 0
    }
    return 0
  })
}

/**
 * Custom sort function for players
 */
function customSortPlayers(players, fieldKey, direction, sortField, isGoalkeeperView) {
  return players.sort((a, b) => {
    let vA = getPlayerValue(a, fieldKey, sortField, isGoalkeeperView)
    let vB = getPlayerValue(b, fieldKey, sortField, isGoalkeeperView)
    const aIsNull = vA === null || vA === undefined
    const bIsNull = vB === null || vB === undefined

    if (aIsNull && bIsNull) return 0
    if (aIsNull) return direction === 'asc' ? 1 : -1
    if (bIsNull) return direction === 'asc' ? -1 : 1

    if (fieldKey === 'position') {
      const indexA = getPositionIndex(vA)
      const indexB = getPositionIndex(vB)
      return direction === 'asc' ? indexA - indexB : indexB - indexA
    }
    if (typeof vA === 'number' && typeof vB === 'number') {
      return direction === 'asc' ? vA - vB : vB - vA
    }
    if (typeof vA === 'string' && typeof vB === 'string') {
      vA = vA.toLowerCase()
      vB = vB.toLowerCase()
      if (vA < vB) return direction === 'asc' ? -1 : 1
      if (vA > vB) return direction === 'asc' ? 1 : -1
      return 0
    }
    return 0
  })
}

/**
 * Filter players based on criteria
 */
function filterPlayers(players, filters) {
  return players.filter(player => {
    if (filters.name && !player.name.toLowerCase().includes(filters.name.toLowerCase())) {
      return false
    }

    if (filters.club && player.club !== filters.club) {
      return false
    }

    if (filters.position && !player.position.includes(filters.position)) {
      return false
    }

    if (filters.ageMin !== null && player.age < filters.ageMin) {
      return false
    }
    if (filters.ageMax !== null && player.age > filters.ageMax) {
      return false
    }

    if (
      filters.transferValueMin !== null &&
      player.transferValueAmount < filters.transferValueMin
    ) {
      return false
    }
    if (
      filters.transferValueMax !== null &&
      player.transferValueAmount > filters.transferValueMax
    ) {
      return false
    }

    if (filters.overallMin !== null && player.Overall < filters.overallMin) {
      return false
    }
    if (filters.overallMax !== null && player.Overall > filters.overallMax) {
      return false
    }

    return true
  })
}

/**
 * Calculate rating statistics for a group of players
 */
function calculateRatingStats(players, statKey) {
  const values = players
    .map(p => p[statKey])
    .filter(v => v !== null && v !== undefined && !Number.isNaN(v))
    .sort((a, b) => a - b)

  if (values.length === 0) {
    return { min: 0, max: 0, mean: 0, median: 0, count: 0 }
  }

  const min = values[0]
  const max = values[values.length - 1]
  const sum = values.reduce((acc, val) => acc + val, 0)
  const mean = sum / values.length
  const median =
    values.length % 2 === 0
      ? (values[values.length / 2 - 1] + values[values.length / 2]) / 2
      : values[Math.floor(values.length / 2)]

  return { min, max, mean: Math.round(mean * 100) / 100, median, count: values.length }
}

/**
 * Batch process multiple calculations
 */
function batchProcess(players, operations) {
  const results = {}

  for (const operation of operations) {
    switch (operation.type) {
      case 'sort':
        results[operation.id] = customSortPlayers(
          [...players],
          operation.fieldKey,
          operation.direction,
          operation.sortField,
          operation.isGoalkeeperView
        )
        break

      case 'filter':
        results[operation.id] = filterPlayers(players, operation.filters)
        break

      case 'stats':
        results[operation.id] = calculateRatingStats(players, operation.statKey)
        break

      default:
        results[operation.id] = { error: `Unknown operation type: ${operation.type}` }
    }
  }

  return results
}

// Handle messages from main thread
self.onmessage = e => {
  const { type, data, id } = e.data

  try {
    let result

    switch (type) {
      case 'SORT_PLAYERS':
        result = customSortPlayers(
          data.players,
          data.fieldKey,
          data.direction,
          data.sortField,
          data.isGoalkeeperView
        )
        break

      case 'FAST_SORT_PLAYERS':
        result = fastSortPlayers(
          data.players,
          data.fieldKey,
          data.direction,
          data.sortField,
          data.isGoalkeeperView
        )
        break

      case 'FILTER_PLAYERS':
        result = filterPlayers(data.players, data.filters)
        break

      case 'CALCULATE_STATS':
        result = calculateRatingStats(data.players, data.statKey)
        break

      case 'BATCH_PROCESS':
        result = batchProcess(data.players, data.operations)
        break

      case 'GET_POSITION_INDEX':
        result = getPositionIndex(data.positionString)
        break

      case 'GET_PLAYER_VALUE':
        result = getPlayerValue(data.player, data.fieldKey, data.columnName, data.isGoalkeeperView)
        break

      default:
        throw new Error(`Unknown message type: ${type}`)
    }

    self.postMessage({
      type: 'SUCCESS',
      id,
      result
    })
  } catch (error) {
    self.postMessage({
      type: 'ERROR',
      id,
      error: error.message
    })
  }
}
