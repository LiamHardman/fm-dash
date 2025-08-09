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
  return player.Overall || player.overall || 0
}

/**
 * Compute a FIFA-category-based overall using position-group weights
 * When a player is GK, uses GK stat; otherwise uses PAC/SHO/PAS/DRI/DEF/PHY with weights by position group
 * @param {Object} player - Player object with FIFA stats (PAC, SHO, PAS, DRI, DEF, PHY) or GK
 * @returns {number}
 */
export const getFifaWeightedOverallByPosition = (player) => {
  if (!player) return 0
  const isGK = (player.position && player.position.includes('GK')) || player.GK > 0 || player.gk > 0
  if (isGK) {
    // GK Top-k boosted aggregator over GK categories
    const gkStats = [
      player.DIV ?? player.div ?? 0,
      player.HAN ?? player.han ?? 0,
      player.REF ?? player.ref ?? 0,
      player.SPD ?? player.spd ?? 0,
      player.POS ?? player.pos ?? 0,
      player.KIC ?? player.kic ?? 0,
    ]
    const validCount = gkStats.filter((v) => typeof v === 'number' && !Number.isNaN(v)).length
    if (validCount === 0) return 0

    const sortedGk = gkStats.slice().sort((a, b) => b - a)
    const kGK = 3
    const lambdaGK = 0.6
    const topkAvgGK =
      sortedGk.slice(0, kGK).reduce((s, v) => s + v, 0) / Math.max(1, Math.min(kGK, gkStats.length))
    const meanAllGK = gkStats.reduce((s, v) => s + v, 0) / gkStats.length
    const gkScore = Math.round(lambdaGK * topkAvgGK + (1 - lambdaGK) * meanAllGK)
    return Math.max(0, Math.min(99, gkScore))
  }

  const PAC = player.PAC ?? player.pac ?? 0
  const SHO = player.SHO ?? player.sho ?? 0
  const PAS = player.PAS ?? player.pas ?? 0
  const DRI = player.DRI ?? player.dri ?? 0
  const DEF = player.DEF ?? player.def ?? 0
  const PHY = player.PHY ?? player.phy ?? 0

  const posStrRaw = String(player.position || '').toUpperCase()
  // Extract first base position code, ignoring sides in parentheses
  const basePosMatch = posStrRaw.match(/\b(GK|DR|DC|DL|WBR|WBL|DM|MR|MC|ML|AMR|AMC|AML|ST)\b/)
  const basePos = basePosMatch ? basePosMatch[1] : ''

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

  const archetypes = positionArchetypeWeights[basePos]

  // Top-k boosted aggregator parameters per position (defaults)
  const topKParams = {
    ST: { k: 2, lambda: 0.75 },
    AMC: { k: 2, lambda: 0.7 },
    AMR: { k: 2, lambda: 0.7 },
    AML: { k: 2, lambda: 0.7 },
    MR: { k: 2, lambda: 0.65 },
    ML: { k: 2, lambda: 0.65 },
    MC: { k: 3, lambda: 0.6 },
    DM: { k: 3, lambda: 0.55 },
    DR: { k: 3, lambda: 0.55 },
    DL: { k: 3, lambda: 0.55 },
    WBR: { k: 3, lambda: 0.55 },
    WBL: { k: 3, lambda: 0.55 },
    DC: { k: 2, lambda: 0.55 },
  }
  const { k, lambda } = topKParams[basePos] || { k: 3, lambda: 0.6 }
  const outfieldStats = [PAC, SHO, PAS, DRI, DEF, PHY]
  const sortedOutfield = outfieldStats.slice().sort((a, b) => b - a)
  const topK = Math.max(1, Math.min(k, outfieldStats.length))
  const topkAvg = sortedOutfield.slice(0, topK).reduce((s, v) => s + v, 0) / topK
  if (!archetypes || archetypes.length === 0) {
    let weights = perPositionWeights[basePos]
    if (!weights) {
      const isDef = /\bD[CRL]\b|\bWB[RL]\b|\bDM\b/.test(posStrRaw)
      const isMid = /\bM[CRL]\b|\bAM[CRL]\b/.test(posStrRaw)
      const isAtt = /\bST\b|\bAM[CRL]\b|\bW[BRL]\b/.test(posStrRaw)
      if (isAtt) weights = genericGroups.ATT
      else if (isMid) weights = genericGroups.MID
      else if (isDef) weights = genericGroups.DEF
      else weights = genericGroups.ANY
    }
    const weightedSum =
      SHO * (weights.SHO || 0) +
      PAC * (weights.PAC || 0) +
      DRI * (weights.DRI || 0) +
      PHY * (weights.PHY || 0) +
      PAS * (weights.PAS || 0) +
      DEF * (weights.DEF || 0)
    const total = Object.values(weights).reduce((a, b) => a + b, 0) || 1
    const weightedMean = weightedSum / total
    const blended = Math.round(lambda * topkAvg + (1 - lambda) * weightedMean)
    return Math.max(0, Math.min(99, blended))
  }

  let best = 0
  for (const weights of archetypes) {
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
    if (total <= 0) continue
    const weightedMean = weightedSum / total
    const score = Math.round(lambda * topkAvg + (1 - lambda) * weightedMean)
    if (score > best) best = score
  }
  return Math.max(0, Math.min(99, best))
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
