const gkFields = { pac: 'div', sho: 'han', pas: 'kic', dri: 'ref', def: 'spd', phy: 'pos' }
const positions = [
  'GK',
  'DR',
  'DC',
  'DL',
  'WBR',
  'WBL',
  'DM',
  'MC',
  'MR',
  'ML',
  'AMR',
  'AMC',
  'AML',
  'ST',
]

export function playerSortValue(player, field, goalkeeper = false) {
  const mapped = gkFields[field.toLowerCase()]
  if (mapped && (goalkeeper || String(player.position || '').includes('GK'))) {
    const value = player[mapped] ?? player[mapped.toUpperCase()]
    if (value !== undefined) return value
  }
  return player[field]
}

export function positionRank(value) {
  if (typeof value !== 'string' || !value) return positions.length + 2
  const upper = value.toUpperCase()
  const matches = new Set()
  for (const part of upper.split(',')) {
    const match = part.trim().match(/^([A-Z/\s]+?)(?:\s*\(([RLC]+)\))?$/)
    if (!match) continue
    for (const base of match[1].split('/').map((code) => code.trim())) {
      if (!match[2]) {
        matches.add(base)
        continue
      }
      for (const side of match[2]) {
        matches.add(['GK', 'ST', 'DM'].includes(base) && side === 'C' ? base : base + side)
      }
    }
  }
  const indices = [...matches].map((p) => positions.indexOf(p)).filter((i) => i >= 0)
  return indices.length ? Math.min(...indices) : positions.length + 1
}

export function comparePlayers(a, b, spec) {
  let left = playerSortValue(a, spec.field, spec.goalkeeper)
  let right = playerSortValue(b, spec.field, spec.goalkeeper)
  const sign = spec.direction === 'asc' ? 1 : -1
  if (left == null && right == null) return 0
  if (left == null) return sign
  if (right == null) return -sign
  if (spec.field === 'position') return sign * (positionRank(left) - positionRank(right))
  if (typeof left === 'number' && typeof right === 'number') return sign * (left - right)
  left = String(left).toLowerCase()
  right = String(right).toLowerCase()
  return sign * (left < right ? -1 : left > right ? 1 : 0)
}

// Table sort fields are scalars. Nested attributes, percentiles, and player
// metadata stay on the UI thread; a worker needs only the sortable projection.
export function projectSortRows(players) {
  return players.map((player) =>
    Object.fromEntries(
      Object.entries(player).filter(
        ([key, value]) =>
          !['__proto__', 'constructor', 'prototype'].includes(key) &&
          (value == null || ['string', 'number', 'boolean'].includes(typeof value))
      )
    )
  )
}
export function sortPlayerIndices(rows, spec) {
  return rows
    .map((_, index) => index)
    .sort((a, b) => comparePlayers(rows[a], rows[b], spec) || a - b)
}
