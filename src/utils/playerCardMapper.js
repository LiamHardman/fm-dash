import { EDITION_MANIFEST } from '../cards/editions.mjs'
import { PLAYSTYLE_PLUS_CATALOG } from '../cards/playstyle-plus-catalog.mjs'
import { getFifaWeightedOverallByPosition } from './playerUtils'

const POSITIONS = {
  GK: 'GK',
  DC: 'CB',
  DR: 'RB',
  DL: 'LB',
  WBR: 'RWB',
  WBL: 'LWB',
  DM: 'CDM',
  DMC: 'CDM',
  MC: 'CM',
  MR: 'RM',
  ML: 'LM',
  AMC: 'CAM',
  AMR: 'RW',
  AML: 'LW',
  STC: 'ST',
  ST: 'ST',
}

export function cardPositions(value) {
  const positions = []
  const text = String(value || '')
    .trim()
    .toUpperCase()
  // FM exports can group roles and sides: D/WB (RL), AM (RLC).
  const expanded = text.replace(/\b([A-Z]+(?:\/[A-Z]+)*)\s*\(([RLC]+)\)/g, (_, roles, sides) => {
    return roles
      .split('/')
      .flatMap((role) =>
        [...sides].map((side) => {
          if (role === 'DM' || role === 'ST') return role
          return role + side
        })
      )
      .join(',')
  })
  for (const token of expanded.split(/[,/;]+/)) {
    const position = token.trim()
    if (position) positions.push(POSITIONS[position] || position)
  }
  return [...new Set(positions)]
}

function rolesFor(player) {
  const roles = player.roleSpecificOveralls || player.role_specific_overalls || {}
  return Array.isArray(roles)
    ? roles
    : Object.entries(roles).map(([roleName, score]) => ({ roleName, score }))
}

export function isCardGoalkeeper(player) {
  return (
    [...(player.shortPositions || player.short_positions || []), player.position].some((position) =>
      /\bGK\b|goalkeeper/i.test(String(position || ''))
    ) ||
    player.positionGroups?.includes('Goalkeepers') ||
    player.parsedPositions?.includes('Goalkeeper') ||
    false
  )
}

export function formatCardCurrency(value, symbol = '£') {
  if (value == null || value === '' || value === 'N/A' || value === '-') return undefined
  if (typeof value === 'string') {
    const upper = value
      .split(' - ')
      .at(-1)
      .trim()
      .replace(/\s*(p\/w|per week|\/wk)$/i, '')
      .trim()
    // Preserve already abbreviated values, including their actual currency.
    if (/[KMB]$/i.test(upper)) return /[£$€¥]/.test(upper) ? upper : symbol + upper
    value = upper.replace(/[£$€¥,\s]/g, '')
  }
  const amount = Number(value)
  if (!Number.isFinite(amount)) return undefined
  const divisor = amount >= 1e9 ? 1e9 : amount >= 1e6 ? 1e6 : amount >= 1e3 ? 1e3 : 1
  return `${symbol}${Number((amount / divisor).toFixed(1))}${{ 1000000000: 'B', 1000000: 'M', 1000: 'K', 1: '' }[divisor]}`
}

// The host owns ratings, FM position translation, rarity and financial formatting.
export function mapPlayerCard(player = {}, options = {}) {
  const roles = rolesFor(player)
  const selected =
    options.selectedRole &&
    (roles.find((role) => role.roleName === options.selectedRole) ||
      roles.find((role) => role.roleName?.startsWith(options.selectedRole.split(' - ')[0])))
  const bestRole = roles.reduce(
    (best, role) => (Number(role.score) > Number(best?.score ?? -1) ? role : best),
    null
  )
  const naturalPositions = [
    ...new Set([
      ...(player.shortPositions || player.short_positions || []).flatMap(cardPositions),
      ...cardPositions(player.position),
    ]),
  ]
  const rolePosition = (
    selected?.roleName ||
    bestRole?.roleName ||
    player.best_role_overall ||
    ''
  ).split(' - ')[0]
  const primaryPosition =
    cardPositions(options.positionOverride || rolePosition)[0] || naturalPositions[0] || ''
  const goalkeeper = isCardGoalkeeper(player)
  const explicitOverall = player.Overall ?? player.overall
  const rating =
    selected?.score ??
    explicitOverall ??
    (options.useStatSummaryForOverall ? getFifaWeightedOverallByPosition(player) : undefined)
  const labels = goalkeeper
    ? ['DIV', 'HAN', 'KIC', 'REF', 'SPD', 'POS']
    : ['PAC', 'SHO', 'PAS', 'DRI', 'DEF', 'PHY']
  const attributes = Object.fromEntries(
    labels.map((label) => [label, player[label.toLowerCase()] ?? player[label]])
  )
  const iso = player.nationalityIso || player.nationality_iso
  const nationLabel = player.nationality || player.nation || iso || player.nationality_fifa_code
  const playstyle = options.playstylePlus ?? player.playstylePlus ?? player.playstyle_plus
  return {
    name: player.name || '',
    rating,
    primaryPosition,
    // The renderer supports eight alternates; FM players can exceed that.
    alternatePositions: naturalPositions
      .filter((position) => position !== primaryPosition)
      .slice(0, 8),
    goalkeeper,
    attributes,
    playstylePlus: Object.hasOwn(PLAYSTYLE_PLUS_CATALOG, playstyle) ? playstyle : undefined,
    nation:
      nationLabel || options.nationFlagUrl
        ? {
            label: nationLabel || '',
            code: player.nationality_fifa_code || iso || '',
            src:
              options.nationFlagUrl ||
              player.nationFlagUrl ||
              (iso ? `https://flagcdn.com/w80/${iso.toLowerCase()}.png` : ''),
          }
        : undefined,
    club:
      player.club && player.club !== '-'
        ? {
            label: player.club,
            code: player.club
              .split(/\s+/)
              .map((word) => word[0])
              .join('')
              .slice(0, 3)
              .toUpperCase(),
            src: options.clubImageUrl || player.clubImageUrl || '',
          }
        : options.clubImageUrl
          ? { label: 'Club', code: '', src: options.clubImageUrl }
          : undefined,
    age: player.age,
    transferValue: formatCardCurrency(
      player.transferValueAmount ?? player.transfer_value ?? player.value ?? player.market_value,
      options.currencySymbol
    ),
    weeklyWage: formatCardCurrency(
      player.wageAmount ?? player.wage ?? player.salary ?? player.weekly_wage,
      options.currencySymbol
    ),
  }
}

export function selectPlayerCardEdition(player, card, options = {}) {
  const baseEdition =
    Number(card.rating) >= 75 ? 'gold' : Number(card.rating) >= 65 ? 'silver' : 'bronze'
  const supplied =
    options.edition ||
    player.cardEdition ||
    player.cardType ||
    player.card_type ||
    (player.isIcon ? 'icon' : '')
  const edition = Object.hasOwn(EDITION_MANIFEST, supplied) ? supplied : 'squad-player'
  const stats = Object.values(card.attributes)
    .filter((value) => value != null && Number.isFinite(Number(value)))
    .map(Number)
  const rating = Number(card.rating)
  const automaticRare =
    Number.isFinite(rating) &&
    (rating >= 85 ||
      stats.filter((value) => value >= rating - 1).length >= 2 ||
      stats.filter((value) => value >= rating - 4).length >= 4)
  const supportsRare = edition === 'squad-player' || ['bronze', 'silver', 'gold'].includes(edition)
  return {
    edition,
    baseEdition: options.baseEdition || baseEdition,
    rare: supportsRare && (options.rare ?? automaticRare),
  }
}
