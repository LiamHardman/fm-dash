// Card tier catalog and stat-boost formula for the Card Generator feature.
// Boost formula ported from src/pages/TeamViewPage.vue's createPrestigePlayer/
// getUpgradedStat so both places share one source of truth. Tier list and
// boost numbers per .scratch/card-generator/issues/01-card-type-boost-tier-catalog.md.
// Design class names are best-effort picks from src/styles/card-designs/ —
// see that ticket for the full reasoning behind each tier's boost/floor.

export const OUTFIELD_STAT_KEYS = ['pac', 'dri', 'sho', 'def', 'pas', 'phy']
export const GOALKEEPER_STAT_KEYS = ['div', 'ref', 'han', 'spd', 'kic', 'pos']

export const OUTFIELD_STAT_LABELS = {
  pac: 'PAC',
  dri: 'DRI',
  sho: 'SHO',
  def: 'DEF',
  pas: 'PAS',
  phy: 'PHY',
}
export const GOALKEEPER_STAT_LABELS = {
  div: 'DIV',
  ref: 'REF',
  han: 'HAN',
  spd: 'SPD',
  kic: 'KIC',
  pos: 'POS',
}

// Tier | boost | floor(minOverall) | rare toggle | design override (null = let PlayerCards decide)
export const CARD_TIERS = [
  {
    key: 'squad-player',
    label: 'Squad Player',
    boost: 0,
    minOverall: 0,
    rareToggle: false,
    design: null,
  },
  { key: 'bronze', label: 'Bronze', boost: 0, minOverall: 0, rareToggle: true, design: null },
  { key: 'silver', label: 'Silver', boost: 0, minOverall: 0, rareToggle: true, design: null },
  { key: 'gold', label: 'Gold', boost: 0, minOverall: 0, rareToggle: true, design: null },
  {
    key: 'bargain',
    label: 'Bargain',
    boost: 0,
    minOverall: 0,
    rareToggle: false,
    design: 'card-design--pale-geometric',
  },
  {
    key: 'wonderkid',
    label: 'Wonderkid',
    boost: 3,
    minOverall: 0,
    rareToggle: false,
    design: 'card-design--amber-facets',
  },
  {
    key: 'totw',
    label: 'Team of the Week',
    boost: 4,
    minOverall: 0,
    rareToggle: false,
    design: 'card-design--parchment-frame',
  },
  {
    key: 'motm',
    label: 'Man of the Match',
    boost: 5,
    minOverall: 0,
    rareToggle: false,
    design: 'card-design--match-night-flare',
  },
  {
    key: 'hero',
    label: 'Hero',
    boost: 8,
    minOverall: 0,
    rareToggle: false,
    design: 'card-design--badge-poster',
  },
  {
    key: 'tots',
    label: 'Team of the Season',
    boost: 9,
    minOverall: 82,
    rareToggle: false,
    design: 'card-design--tots-crown-aurora',
  },
  {
    key: 'icon',
    label: 'Icon',
    boost: 10,
    minOverall: 85,
    rareToggle: false,
    design: 'card-design--ivory-museum',
  },
]

const STAT_RELEVANCE_BY_POSITION = {
  GK: { div: 1.18, han: 1.08, kic: 0.72, ref: 1.2, spd: 0.55, pos: 1.12 },
  LB: { pac: 0.9, dri: 0.65, sho: 0.22, def: 1.2, pas: 0.76, phy: 1.05 },
  RB: { pac: 0.9, dri: 0.65, sho: 0.22, def: 1.2, pas: 0.76, phy: 1.05 },
  CB: { pac: 0.58, dri: 0.28, sho: 0.18, def: 1.24, pas: 0.64, phy: 1.14 },
  CDM: { pac: 0.5, dri: 0.68, sho: 0.36, def: 1.12, pas: 1.1, phy: 0.96 },
  CM: { pac: 0.48, dri: 0.9, sho: 0.58, def: 0.68, pas: 1.18, phy: 0.78 },
  CAM: { pac: 0.72, dri: 1.12, sho: 0.95, def: 0.24, pas: 1.16, phy: 0.52 },
  LW: { pac: 1.12, dri: 1.14, sho: 0.82, def: 0.22, pas: 0.82, phy: 0.52 },
  LM: { pac: 1.02, dri: 1.04, sho: 0.62, def: 0.48, pas: 0.92, phy: 0.62 },
  ST: { pac: 0.84, dri: 0.72, sho: 1.22, def: 0.16, pas: 0.58, phy: 0.92 },
  RW: { pac: 1.12, dri: 1.14, sho: 0.82, def: 0.22, pas: 0.82, phy: 0.52 },
  RM: { pac: 1.02, dri: 1.04, sho: 0.62, def: 0.48, pas: 0.92, phy: 0.62 },
}

const SHORT_TO_DISPLAY_POSITION = {
  GK: 'GK',
  DC: 'CB',
  DR: 'RB',
  DL: 'LB',
  WBR: 'RB',
  WBL: 'LB',
  DM: 'CDM',
  MR: 'RM',
  ML: 'LM',
  MC: 'CM',
  AMR: 'RW',
  AML: 'LW',
  AMC: 'CAM',
  ST: 'ST',
}

function clampNumber(value, min, max) {
  return Math.min(Math.max(value, min), max)
}

function getDisplayPosition(player) {
  const positions = player?.shortPositions || player?.short_positions || []
  for (const pos of positions) {
    const mapped = SHORT_TO_DISPLAY_POSITION[pos]
    if (mapped) return mapped
  }
  return 'CM'
}

function getRelevantStatFloor(targetOverall, relevance) {
  if (relevance >= 1.18) return targetOverall - 2
  if (relevance >= 1) return targetOverall - 7
  if (relevance >= 0.82) return targetOverall - 8
  if (relevance >= 0.62) return targetOverall - 10
  return 0
}

function getUpgradedStat(
  statValue,
  baseOverall,
  targetOverall,
  relevance,
  displayPosition,
  positionBonus
) {
  const numericStat = Number(statValue)
  if (Number.isNaN(numericStat) || numericStat <= 0) return statValue
  const overallBoost = Math.max(0, targetOverall - baseOverall)
  let statBoost = overallBoost * (0.65 + relevance * 0.7) + positionBonus
  if (numericStat >= targetOverall + 2) statBoost *= 0.35
  else if (numericStat >= targetOverall) statBoost *= 0.5
  const floor = getRelevantStatFloor(targetOverall, relevance)
  const isAttacker = ['LW', 'LM', 'ST', 'RW', 'RM'].includes(displayPosition)
  const positionCap = relevance >= 0.62 ? targetOverall + 8 : targetOverall - (isAttacker ? 5 : 8)
  const boosted = Math.max(numericStat + statBoost, floor)
  return Math.round(clampNumber(boosted, numericStat, Math.min(99, positionCap)))
}

export function getBaseOverall(player) {
  return Number(player?.Overall ?? player?.overall ?? 0)
}

export function isGoalkeeper(player) {
  return getDisplayPosition(player) === 'GK'
}

export function statKeysFor(player) {
  return isGoalkeeper(player) ? GOALKEEPER_STAT_KEYS : OUTFIELD_STAT_KEYS
}

export function statLabelsFor(player) {
  return isGoalkeeper(player) ? GOALKEEPER_STAT_LABELS : OUTFIELD_STAT_LABELS
}

// Applies a tier's boost formula to produce a fresh baseline stat block —
// this is what re-running the formula on tier change recomputes.
export function applyTierBoost(player, tier) {
  const baseOverall = getBaseOverall(player)
  const displayPosition = getDisplayPosition(player)
  const isGK = displayPosition === 'GK'
  const targetOverall = Math.min(99, Math.max(baseOverall + tier.boost, tier.minOverall))
  const relevance = STAT_RELEVANCE_BY_POSITION[displayPosition] || STAT_RELEVANCE_BY_POSITION.CM
  const statKeys = isGK ? GOALKEEPER_STAT_KEYS : OUTFIELD_STAT_KEYS
  const positionBonus = isGK ? 0 : 1

  const stats = { overall: targetOverall }
  for (const key of statKeys) {
    const raw = player?.[key]
    if (raw !== undefined) {
      stats[key] = getUpgradedStat(
        raw,
        baseOverall,
        targetOverall,
        relevance[key] || 0.5,
        displayPosition,
        positionBonus
      )
    }
  }
  return stats
}
