/**
 * Player utility functions for data processing and analysis
 */

/**
 * Get the division of a player, handling different field names
 * @param {Object} player - The player object
 * @returns {string} The player's division or 'N/A' if not found
 */
export const getPlayerDivision = (player) => {
  return player.division || player.Division || 'N/A'
}

/**
 * Convert a value to a numeric value, handling various formats
 * @param {any} val - The value to convert
 * @returns {number|null} The numeric value or null if invalid
 */
export const getNumericValue = (val) => {
  if (val === undefined || val === null || val === '-' || val === '') return null
  const cleaned = String(val).replace(/,/g, '').replace(/%/g, '')
  const num = Number.parseFloat(cleaned)
  return Number.isNaN(num) ? null : num
}

/**
 * Get a player's positions as an array
 * @param {Object} player - The player object
 * @returns {Array<string>} Array of player positions
 */
export const getPlayerPositions = (player) => {
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
export const getPlayerOverall = (player) => {
  return player.overall || 0
}

/**
 * Compute a FIFA-category-based overall using position-group weights
 * When a player is GK, uses GK stat; otherwise uses PAC/SHO/PAS/DRI/DEF/PHY with weights by position group
 * @param {Object} player - Player object with FIFA stats (PAC, SHO, PAS, DRI, DEF, PHY) or GK
 * @returns {number}
 */
export const getFifaWeightedOverallByPosition = (player) => {
  if (!player) return 0
  const isGK = player.position?.includes('GK') || player.gk > 0
  if (isGK) {
    // Use primary GK category if present
    const parts = [
      player.gk,
      player.div,
      player.han,
      player.ref,
      player.spd,
      player.pos,
    ]
    const valid = parts.filter((v) => typeof v === 'number' && !Number.isNaN(v))
    if (valid.length === 0) return 0
    // Simple average for GK categories
    return Math.round(valid.reduce((a, b) => a + b, 0) / valid.length)
  }

  const PAC = player.pac ?? 0
  const SHO = player.sho ?? 0
  const PAS = player.pas ?? 0
  const DRI = player.dri ?? 0
  const DEF = player.def ?? 0
  const PHY = player.phy ?? 0

  const posStrRaw = String(player.position || '').toUpperCase()
  // Collect candidate base positions from short positions and position string (best-of evaluation)
  const allowedBase = new Set([
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
  ])
  const shortPositions = Array.isArray(player.shortPositions)
    ? player.shortPositions
    : Array.isArray(player.short_positions)
      ? player.short_positions
      : []
  const candidatesFromShort = shortPositions
    .map((p) => String(p).toUpperCase())
    .filter((p) => allowedBase.has(p))
  // Extract compact tokens (e.g., "MC", "AMC") directly
  const compactTokenRegex = /\b(GK|DR|DC|DL|WBR|WBL|DM|MR|MC|ML|AMR|AMC|AML|ST)\b/g
  const compactMatches = posStrRaw.match(compactTokenRegex) || []
  // Extract spaced tokens like "M (C)", "AM (R)", "D (L)", "WB (R)" and normalize them
  const spacedTokenRegex = /\b(AM|DM|WB|[DMWA])\s*\(\s*([CRL])\s*\)\b/g
  const spacedMatchesNormalized = []
  for (const match of posStrRaw.matchAll(spacedTokenRegex)) {
    const group = match[1]
    const side = match[2]
    let normalized = ''
    switch (group) {
      case 'D':
        normalized = side === 'C' ? 'DC' : side === 'R' ? 'DR' : 'DL'
        break
      case 'M':
        normalized = side === 'C' ? 'MC' : side === 'R' ? 'MR' : 'ML'
        break
      case 'AM':
        normalized = side === 'C' ? 'AMC' : side === 'R' ? 'AMR' : 'AML'
        break
      case 'WB':
        normalized = side === 'R' ? 'WBR' : 'WBL'
        break
      case 'W':
        normalized = side === 'R' ? 'AMR' : 'AML' // Map wide mids to AMR/AML for weighting purposes
        break
      default:
        normalized = ''
    }
    if (normalized && allowedBase.has(normalized)) {
      spacedMatchesNormalized.push(normalized)
    }
  }
  const matchesFromStr = [...compactMatches, ...spacedMatchesNormalized]
  const candidateBasePositions = [
    ...new Set([...candidatesFromShort, ...matchesFromStr].filter((p) => p !== 'GK')),
  ]

  const positionArchetypeWeights = {
    DR: [
      { PAC: 35, DRI: 20, PAS: 18, DEF: 17, PHY: 5, SHO: 5 },
      { PAC: 30, DEF: 25, PAS: 20, DRI: 15, PHY: 5, SHO: 5 },
      { DEF: 30, PAC: 25, PHY: 20, PAS: 15, DRI: 5, SHO: 5 },
    ],
    DL: [
      { PAC: 35, DRI: 20, PAS: 18, DEF: 17, PHY: 5, SHO: 5 },
      { PAC: 30, DEF: 25, PAS: 20, DRI: 15, PHY: 5, SHO: 5 },
      { DEF: 30, PAC: 25, PHY: 20, PAS: 15, DRI: 5, SHO: 5 },
    ],
    WBR: [
      { PAC: 35, PAS: 20, DRI: 20, SHO: 8, DEF: 12, PHY: 5 },
      { PAC: 28, DEF: 30, PAS: 12, DRI: 15, PHY: 10, SHO: 5 },
    ],
    WBL: [
      { PAC: 35, PAS: 20, DRI: 20, SHO: 8, DEF: 12, PHY: 5 },
      { PAC: 28, DEF: 30, PAS: 12, DRI: 15, PHY: 10, SHO: 5 },
    ],
    DC: [
      { DEF: 30, PHY: 20, PAC: 15, PAS: 25, DRI: 5, SHO: 5 },
      { DEF: 40, PHY: 30, PAC: 15, PAS: 5, DRI: 5, SHO: 5 },
      { DEF: 35, PHY: 20, PAC: 20, PAS: 15, DRI: 5, SHO: 5 },
    ],
    DM: [
      { PAS: 35, DRI: 15, DEF: 20, PHY: 15, PAC: 10, SHO: 5 },
      { DEF: 30, PAS: 15, PHY: 25, PAC: 10, DRI: 10, SHO: 10 },
      { DEF: 30, PAS: 15, PHY: 25, PAC: 15, DRI: 10, SHO: 5 },
    ],
    MR: [
      { PAC: 35, DRI: 25, PAS: 15, SHO: 15, PHY: 5, DEF: 5 },
      { PAC: 25, DRI: 25, PAS: 30, SHO: 10, PHY: 5, DEF: 5 },
      { PAC: 20, DRI: 15, PAS: 20, DEF: 20, PHY: 20, SHO: 5 },
    ],
    ML: [
      { PAC: 35, DRI: 25, PAS: 15, SHO: 15, PHY: 5, DEF: 5 },
      { PAC: 25, DRI: 25, PAS: 30, SHO: 10, PHY: 5, DEF: 5 },
      { PAC: 20, DRI: 15, PAS: 20, DEF: 20, PHY: 20, SHO: 5 },
    ],
    MC: [
      { PAS: 35, DRI: 20, PAC: 10, DEF: 15, PHY: 10, SHO: 10 },
      { PAS: 20, DRI: 15, PAC: 20, DEF: 15, PHY: 20, SHO: 10 },
      { PAS: 15, DRI: 10, PAC: 15, DEF: 30, PHY: 20, SHO: 10 },
      // Creator-focused MC archetype to better reflect playmakers
      { PAS: 47, DRI: 32, PAC: 12, PHY: 5, DEF: 2, SHO: 2 },
    ],
    AMR: [
      { PAC: 30, DRI: 25, PAS: 20, SHO: 20, PHY: 3, DEF: 2 },
      { PAC: 35, DRI: 25, PAS: 20, SHO: 15, PHY: 3, DEF: 2 },
      { PAC: 25, DRI: 25, PAS: 30, SHO: 15, PHY: 3, DEF: 2 },
    ],
    AML: [
      { PAC: 30, DRI: 25, PAS: 20, SHO: 20, PHY: 3, DEF: 2 },
      { PAC: 35, DRI: 25, PAS: 20, SHO: 15, PHY: 3, DEF: 2 },
      { PAC: 25, DRI: 25, PAS: 30, SHO: 15, PHY: 3, DEF: 2 },
    ],
    AMC: [
      { PAS: 30, DRI: 25, PAC: 10, SHO: 25, PHY: 5, DEF: 5 },
      { PAS: 20, DRI: 20, PAC: 15, SHO: 35, PHY: 5, DEF: 5 },
      { PAS: 25, DRI: 20, PAC: 20, SHO: 20, PHY: 10, DEF: 5 },
      // Creator-focused AMC archetype: de-emphasize SHO, boost PAS/DRI/PAC
      { PAS: 37, DRI: 28, PAC: 18, SHO: 10, PHY: 5, DEF: 2 },
    ],
    ST: [
      { SHO: 40, PAC: 30, DRI: 15, PHY: 10, PAS: 3, DEF: 2 },
      { SHO: 30, PAC: 15, DRI: 10, PHY: 35, PAS: 5, DEF: 5 },
      { SHO: 25, PAC: 30, DRI: 15, PHY: 20, PAS: 5, DEF: 5 },
      { SHO: 35, PAC: 25, DRI: 20, PHY: 15, PAS: 3, DEF: 2 },
      { PAS: 25, DRI: 25, SHO: 20, PAC: 20, PHY: 8, DEF: 2 },
    ],
  }

  // Generic per-position baseline if no archetypes present
  const perPositionWeights = {
    DR: { PAC: 30, DEF: 30, PAS: 15, DRI: 10, PHY: 10, SHO: 5 },
    DL: { PAC: 30, DEF: 30, PAS: 15, DRI: 10, PHY: 10, SHO: 5 },
    WBR: { PAC: 35, DEF: 20, PAS: 15, DRI: 15, PHY: 10, SHO: 5 },
    WBL: { PAC: 35, DEF: 20, PAS: 15, DRI: 15, PHY: 10, SHO: 5 },
    DC: { DEF: 35, PHY: 25, PAC: 20, PAS: 10, DRI: 5, SHO: 5 },
    DM: { DEF: 25, PAS: 25, PHY: 20, PAC: 15, DRI: 10, SHO: 5 },
    MR: { PAC: 30, DRI: 25, PAS: 20, SHO: 10, PHY: 10, DEF: 5 },
    ML: { PAC: 30, DRI: 25, PAS: 20, SHO: 10, PHY: 10, DEF: 5 },
    MC: { PAS: 30, DRI: 15, PHY: 15, PAC: 15, DEF: 15, SHO: 10 },
    AMR: { PAC: 30, DRI: 25, PAS: 20, SHO: 15, PHY: 5, DEF: 5 },
    AML: { PAC: 30, DRI: 25, PAS: 20, SHO: 15, PHY: 5, DEF: 5 },
    AMC: { PAS: 25, DRI: 25, PAC: 20, SHO: 20, PHY: 5, DEF: 5 },
    ST: { SHO: 35, PAC: 30, DRI: 15, PHY: 10, PAS: 5, DEF: 5 },
  }

  // Generic fallback groups if base position missing
  const genericGroups = {
    ATT: { SHO: 25, PAC: 35, DRI: 20, PHY: 15, PAS: 5, DEF: 0 },
    MID: { PAS: 25, PHY: 15, PAC: 30, DRI: 15, DEF: 10, SHO: 5 },
    DEF: { DEF: 25, PHY: 20, PAC: 30, PAS: 15, DRI: 5, SHO: 5 },
    ANY: { PHY: 15, PAC: 30, PAS: 15, DEF: 15, DRI: 15, SHO: 10 },
  }

  // Helper to score a single weights object
  const scoreForWeights = (weights) => {
    const weightedSum =
      SHO * (weights.SHO || 0) +
      PAC * (weights.PAC || 0) +
      DRI * (weights.DRI || 0) +
      PHY * (weights.PHY || 0) +
      PAS * (weights.PAS || 0) +
      DEF * (weights.DEF || 0)
    const total =
      (weights.SHO || 0) +
      (weights.PAC || 0) +
      (weights.DRI || 0) +
      (weights.PHY || 0) +
      (weights.PAS || 0) +
      (weights.DEF || 0)
    if (!total) return 0
    return Math.round(weightedSum / total)
  }

  // If we have at least one candidate base position, evaluate all and return the best
  if (candidateBasePositions.length > 0) {
    let bestAcrossPositions = 0
    for (const basePos of candidateBasePositions) {
      const archetypes = positionArchetypeWeights[basePos]
      if (archetypes && archetypes.length > 0) {
        let bestForThisPos = 0
        for (const weights of archetypes) {
          const score = scoreForWeights(weights)
          if (score > bestForThisPos) bestForThisPos = score
        }
        if (bestForThisPos > bestAcrossPositions) bestAcrossPositions = bestForThisPos
        continue
      }
      const weights = perPositionWeights[basePos]
      if (weights) {
        const score = scoreForWeights(weights)
        if (score > bestAcrossPositions) bestAcrossPositions = score
        continue
      }
      // Generic fallback by group if we somehow have an unrecognized basePos
      const isDef = /^(DR|DC|DL|WBR|WBL|DM)$/.test(basePos)
      const isMid = /^(MR|MC|ML|AMR|AMC|AML)$/.test(basePos)
      const isAtt = /^(ST)$/.test(basePos)
      const generic = isAtt
        ? genericGroups.ATT
        : isMid
          ? genericGroups.MID
          : isDef
            ? genericGroups.DEF
            : genericGroups.ANY
      const score = scoreForWeights(generic)
      if (score > bestAcrossPositions) bestAcrossPositions = score
    }
    return bestAcrossPositions
  }

  // No specific base positions identified: infer via generic groups from the raw position string
  let weights = null
  const isDef = /\bD[CRL]\b|\bWB[RL]\b|\bDM\b/.test(posStrRaw)
  const isMid = /\bM[CRL]\b|\bAM[CRL]\b/.test(posStrRaw)
  const isAtt = /\bST\b|\bAM[CRL]\b|\bW[BRL]\b/.test(posStrRaw)
  if (isAtt) weights = genericGroups.ATT
  else if (isMid) weights = genericGroups.MID
  else if (isDef) weights = genericGroups.DEF
  else weights = genericGroups.ANY
  return scoreForWeights(weights)
}

/**
 * Derive short position codes (e.g., DR, DC, MC, AMC) from an FM-style position string.
 * Mirrors backend ParsePlayerPositionsGo behavior closely for consistency.
 * @param {string} positionStr
 * @returns {string[]} Unique short codes
 */
export const deriveShortPositionsFromPositionString = (positionStr) => {
  if (!positionStr || typeof positionStr !== 'string') return []

  const results = new Set()
  const allowedBase = new Set([
    'GK',
    'SW',
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
  ])

  const mainParts = positionStr
    .toUpperCase()
    .split(',')
    .map((p) => p.trim())
    .filter(Boolean)
  for (const part of mainParts) {
    // Extract roles segment and sides from parentheses
    const endIdx = part.lastIndexOf(')')
    const startIdx = part.lastIndexOf('(')
    let rolesSegment = part
    let explicitSides = []
    if (endIdx === part.length - 1 && startIdx > 0 && startIdx < endIdx) {
      rolesSegment = part.slice(0, startIdx).trim()
      const sides = part.slice(startIdx + 1, endIdx)
      explicitSides = sides.split('').filter((c) => c === 'R' || c === 'L' || c === 'C')
    }

    const roleKeys = rolesSegment
      .split('/')
      .map((k) => k.trim())
      .filter(Boolean)
    for (const roleKey of roleKeys) {
      // If already a direct short code like DR, MC, AML, add and continue
      if (allowedBase.has(roleKey)) {
        results.add(roleKey)
        continue
      }

      // Determine sides to use
      let sidesToUse = explicitSides
      if (sidesToUse.length === 0) {
        switch (roleKey) {
          case 'D':
          case 'M':
          case 'AM':
          case 'DM':
          case 'ST':
          case 'SW':
            sidesToUse = ['C']
            break
          case 'WB':
            // WB needs explicit side; skip if not provided
            sidesToUse = []
            break
          case 'GK':
            sidesToUse = ['']
            break
          default:
            sidesToUse = ['']
            break
        }
      }

      for (const side of sidesToUse) {
        let short = ''
        if (side === '') {
          short = roleKey
        } else {
          switch (roleKey) {
            case 'D':
              short = side === 'C' ? 'DC' : side === 'R' ? 'DR' : 'DL'
              break
            case 'M':
              short = side === 'C' ? 'MC' : side === 'R' ? 'MR' : 'ML'
              break
            case 'AM':
              short = side === 'C' ? 'AMC' : side === 'R' ? 'AMR' : 'AML'
              break
            case 'DM':
              short = 'DM'
              break
            case 'WB':
              if (side === 'R') short = 'WBR'
              if (side === 'L') short = 'WBL'
              break
            case 'ST':
              short = 'ST'
              break
            case 'SW':
              short = 'SW'
              break
            case 'GK':
              short = 'GK'
              break
            default:
              short = roleKey + side
          }
        }
        if (short && allowedBase.has(short)) {
          results.add(short)
        }
      }
    }
  }

  return Array.from(results)
}

/**
 * Compute dataset-driven linear calibration to align FIFA-based overalls
 * to the baseline role-based overall scale.
 * Returns { scale, offset }. Use: calibrated = scale * fifa + offset
 * Falls back to {1, 0} if insufficient data.
 * @param {Array<Object>} players
 */
export const computeFifaOverallCalibration = (players) => {
  if (!Array.isArray(players) || players.length === 0) return { scale: 1, offset: 0 }

  const pairs = []
  for (const p of players) {
    const baseline =
      typeof p.Overall === 'number' ? p.Overall : typeof p.overall === 'number' ? p.overall : null
    if (baseline === null || baseline <= 0) continue
    const fifa = getFifaWeightedOverallByPosition(p)
    if (typeof fifa === 'number' && !Number.isNaN(fifa) && fifa > 0) {
      pairs.push({ fifa, baseline })
    }
  }

  if (pairs.length < 10) {
    return { scale: 1, offset: 0 }
  }

  // Compute means
  let sumF = 0
  let sumR = 0
  for (const { fifa, baseline } of pairs) {
    sumF += fifa
    sumR += baseline
  }
  const n = pairs.length
  const meanF = sumF / n
  const meanR = sumR / n

  // Compute standard deviations
  let varF = 0
  let varR = 0
  for (const { fifa, baseline } of pairs) {
    varF += (fifa - meanF) * (fifa - meanF)
    varR += (baseline - meanR) * (baseline - meanR)
  }
  const sdF = Math.sqrt(varF / n)
  const sdR = Math.sqrt(varR / n)

  const scale = sdF > 1e-6 ? sdR / sdF : 1
  const offset = meanR - scale * meanF
  return { scale, offset }
}

/**
 * Get a player's age as a number
 * @param {Object} player - The player object
 * @returns {number} The player's age
 */
export const getPlayerAge = (player) => {
  return Number.parseInt(player.age || player.Age || 0, 10)
}

/**
 * Get a player's transfer value as a number
 * @param {Object} player - The player object
 * @returns {number} The player's transfer value
 */
export const getPlayerTransferValue = (player) => {
  return Number.parseInt(player.transferValueAmount || player.TransferValueAmount || 0, 10)
}

/**
 * Get a player's wage as a number
 * @param {Object} player - The player object
 * @returns {number} The player's wage
 */
export const getPlayerWage = (player) => {
  return Number.parseInt(player.wageAmount || player.WageAmount || 0, 10)
}
