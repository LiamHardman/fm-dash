// FM's "Inf" (information) flags, e.g. "Inj, Loa". Carried by save imports and by FM27
// CSV-mod exports that include an Inf column (player.inf), but only shown for save imports.

import { caSourceOf } from './caSource.js'

export const INF_CODE_DESCRIPTIONS = {
  Inj: 'Injured',
  Loa: 'Listed for loan',
  Wnt: 'Transfer interest/listing flag (unconfirmed meaning)',
  Lst: 'Transfer listed',
  FrA: 'Free agent',
  Fut: 'Future transfer agreed',
  PR: 'Pre-contract agreed',
  Sus: 'Suspended',
  U21: 'Under 21',
}
// Codes whose meaning hasn't been confirmed in-game get a "?" marker.
export const UNCONFIRMED_INF_CODES = new Set(['Wnt'])

/** "Inj, Loa" -> ['Inj', 'Loa'] (commas and/or whitespace separate codes). */
export function splitInfCodes(inf) {
  return typeof inf === 'string' ? inf.split(/[,\s]+/).filter(Boolean) : []
}

export function describeInfCode(code) {
  return INF_CODE_DESCRIPTIONS[code] ?? 'FM info flag'
}

export function hasInfFlags(players) {
  return Array.isArray(players) && players.some((player) => splitInfCodes(player?.inf).length > 0)
}

/** Whether the Inf column and filters apply: a save import with at least one flagged player. */
export function infFlagsAvailable(players) {
  return caSourceOf(players) === 'save' && hasInfFlags(players)
}

/** The Inf filters: hide injured players, and/or only players listed for loan. */
export function passesInfFilters(player, { hideInjured = false, loanListedOnly = false } = {}) {
  if (!hideInjured && !loanListedOnly) return true
  const codes = splitInfCodes(player?.inf)
  if (hideInjured && codes.includes('Inj')) return false
  if (loanListedOnly && !codes.includes('Loa')) return false
  return true
}
