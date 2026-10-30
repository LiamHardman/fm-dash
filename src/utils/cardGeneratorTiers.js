// Card tier catalog and stat-boost formula for the Card Generator feature.
// Boost formula ported from src/pages/TeamViewPage.vue's createPrestigePlayer/
// getUpgradedStat so both places share one source of truth. Tier list and
// boost numbers per .scratch/card-generator/issues/01-card-type-boost-tier-catalog.md.
import { EDITION_MANIFEST } from '../cards/editions.mjs'
import { isCardGoalkeeper, mapPlayerCard } from './playerCardMapper'

export const OUTFIELD_STAT_KEYS = ['pac', 'sho', 'pas', 'dri', 'def', 'phy']
export const GOALKEEPER_STAT_KEYS = ['div', 'han', 'kic', 'ref', 'spd', 'pos']

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

// Presentation identifiers and labels come from Cardlab; boost policy stays in FMDash.
const BOOSTS = { wonderkid: 3, totw: 4, motm: 5, hero: 8, tots: 9, icon: 10 }
export const CARD_TIERS = Object.entries(EDITION_MANIFEST).map(([key, edition]) => ({
  key,
  label: edition.label,
  boost: BOOSTS[key] || 0,
  minOverall: key === 'icon' ? 85 : key === 'tots' ? 82 : 0,
  rareToggle: edition.alias || Object.hasOwn(edition.variants, 'rare'),
}))

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

function clampNumber(value, min, max) {
  return Math.min(Math.max(value, min), max)
}

function getDisplayPosition(player) {
  const position = mapPlayerCard(player).primaryPosition
  return { RWB: 'RB', LWB: 'LB' }[position] || position || 'CM'
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
  return isCardGoalkeeper(player)
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
  const isGK = isGoalkeeper(player)
  const displayPosition = isGK ? 'GK' : getDisplayPosition(player)
  const targetOverall = Math.min(99, Math.max(baseOverall + tier.boost, tier.minOverall))
  const relevance = STAT_RELEVANCE_BY_POSITION[displayPosition] || STAT_RELEVANCE_BY_POSITION.CM
  const statKeys = isGK ? GOALKEEPER_STAT_KEYS : OUTFIELD_STAT_KEYS
  const positionBonus = isGK ? 0 : 1

  const stats = { overall: targetOverall }
  for (const key of statKeys) {
    const raw = player?.[key] ?? player?.[key.toUpperCase()]
    if (raw !== undefined) {
      stats[key] =
        tier.boost === 0 && tier.minOverall === 0
          ? raw
          : getUpgradedStat(
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
