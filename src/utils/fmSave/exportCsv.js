// Player rows decoded from an FM27 save -> a semicolon CSV shaped like FM's own
// "Export to CSV" view, so FM-Dash's existing /api/upload pipeline (csvHeaderNormalize ->
// parseCellsToPlayer) ingests it unchanged.
//
// Current-season stats come from the season-stats frame (seasonStats.js) and use FM's own
// cell formats; cells FM shows as zero only because the advanced block is missing are blank,
// which the Go parser stores as missing. Personality and Media Handling are derived from the
// hidden personality bytes (personality.js). Media Description is absent (FM generates it at
// runtime, it is not in the save). Division is blank wherever the league can't be derived.

import { infCell } from './inf.js'
import { NATION_FIFA_CODES } from './nation-codes.js'
import { personalityLabels } from './personality.js'
import { SEASON_STAT_HEADERS, seasonStatCells } from './seasonStats.js'

// fm27-scout-code attribute key -> FM's CSV header (src/api/parsing.go csvHeaderNormalize).
export const ATTRIBUTE_HEADERS = {
  corners: 'Corners',
  crossing: 'Crossing',
  dribbling: 'Dribbling',
  finishing: 'Finishing',
  first_touch: 'First Touch',
  free_kick_taking: 'Free Kick Taking',
  heading: 'Heading',
  long_shots: 'Long Shots',
  long_throws: 'Long Throws',
  marking: 'Marking',
  passing: 'Passing',
  penalty_taking: 'Penalty Taking',
  tackling: 'Tackling',
  technique: 'Technique',
  aggression: 'Aggression',
  anticipation: 'Anticipation',
  bravery: 'Bravery',
  composure: 'Composure',
  concentration: 'Concentration',
  decisions: 'Decisions',
  determination: 'Determination',
  flair: 'Flair',
  leadership: 'Leadership',
  off_the_ball: 'Off The Ball',
  positioning: 'Positioning',
  teamwork: 'Team Work',
  vision: 'Vision',
  work_rate: 'Work Rate',
  acceleration: 'Acceleration',
  agility: 'Agility',
  balance: 'Balance',
  jumping_reach: 'Jumping Reach',
  natural_fitness: 'Natural Fitness',
  pace: 'Pace',
  stamina: 'Stamina',
  strength: 'Strength',
  aerial_reach: 'Aerial Reach',
  command_of_area: 'Command Of Area',
  communication: 'Communication',
  eccentricity: 'Eccentricity',
  handling: 'Handling',
  kicking: 'Kicking',
  one_on_ones: 'One On Ones',
  punching: 'Punching',
  reflexes: 'Reflexes',
  rushing_out: 'Rushing Out (Tendency)',
  throwing: 'Throwing',
}

const IDENTITY_HEADERS = [
  // Nationality precedes Nation: parseCellsToPlayer overwrites Nationality on its own
  // column, but only fills it from Nation's FIFA code while it is still empty.
  'Player',
  'Nationality',
  'Nation',
  'Club',
  'Division',
  'Position',
  'Based In',
  'Age',
  'DoB',
  'Transfer Value',
  'Wage',
  'Unique ID',
  'Left Foot',
  'Right Foot',
  'Personality',
  'Media Handling',
]

// Save-only columns: true CA/PA, the 5 hidden and 8 personality attributes (1-20), and the
// Inf flags. Go parses them into SaveCA, PA, HiddenAttributes and Inf, never into the
// attributes that feed ratings. [header, read(player)]
const SAVE_ONLY_COLUMNS = [
  ['CA', (p) => (p.ca > 0 ? p.ca : undefined)],
  ['PA', (p) => (p.pa > 0 ? p.pa : undefined)],
  ['Consistency', (p) => p.attributes.consistency],
  ['Dirtiness', (p) => p.attributes.dirtiness],
  ['Important Matches', (p) => p.attributes.important_matches],
  ['Injury Proneness', (p) => p.attributes.injury_proneness],
  ['Versatility', (p) => p.attributes.versatility],
  ['Adaptability', (p) => p.personality?.adaptability],
  ['Ambition', (p) => p.personality?.ambition],
  ['Controversy', (p) => p.personality?.controversy],
  ['Loyalty', (p) => p.personality?.loyalty],
  ['Pressure', (p) => p.personality?.pressure],
  ['Professionalism', (p) => p.personality?.professionalism],
  ['Sportsmanship', (p) => p.personality?.sportsmanship],
  ['Temperament', (p) => p.personality?.temperament],
  ['Inf', (p) => infCell(p)],
]
export const SAVE_ONLY_HEADERS = SAVE_ONLY_COLUMNS.map(([header]) => header)

export const CSV_HEADERS = [
  ...IDENTITY_HEADERS,
  ...Object.values(ATTRIBUTE_HEADERS),
  ...SEASON_STAT_HEADERS,
  ...SAVE_ONLY_HEADERS,
]
const BLANK_STATS = SEASON_STAT_HEADERS.map(() => '')

// Positions a player is at least accomplished in (rating 15+, fm27-scout-code's threshold).
const POSITION_THRESHOLD = 15
const SIDE_ORDER = ['R', 'L', 'C']
// FM's position groups, in display order, with the save position names that feed them.
const POSITION_GROUPS = [
  ['GK', { GK: '' }],
  ['SW', { SW: '' }],
  ['D', { DR: 'R', DL: 'L', DC: 'C' }],
  ['WB', { WBR: 'R', WBL: 'L' }],
  ['DM', { DM: '' }],
  ['M', { MR: 'R', ML: 'L', MC: 'C' }],
  ['AM', { AMR: 'R', AML: 'L', AMC: 'C' }],
  ['ST', { STC: 'C' }],
]

/**
 * FM-style position string from the 15 position ratings, e.g. "D (RC), DM, ST (C)".
 * Falls back to the best-rated position(s) so a player never has an empty Position.
 */
export function formatPositions(positions) {
  let qualifies = (name) => positions[name] >= POSITION_THRESHOLD
  if (!Object.keys(positions).some(qualifies)) {
    const best = Math.max(...Object.values(positions))
    qualifies = (name) => positions[name] === best
  }
  const parts = []
  for (const [group, members] of POSITION_GROUPS) {
    const sides = Object.entries(members)
      .filter(([name]) => qualifies(name))
      .map(([, side]) => side)
    if (sides.length === 0) continue
    const lettered = SIDE_ORDER.filter((side) => sides.includes(side)).join('')
    parts.push(lettered ? `${group} (${lettered})` : group)
  }
  return parts.join(', ')
}

const trimmed = (value, decimals) => Number(value.toFixed(decimals)).toString()

/** FM-style money: £159.54M, £631.5K, £900. */
export function formatMoney(amount) {
  if (amount === undefined || amount === null) return ''
  if (amount >= 1_000_000) return `£${trimmed(amount / 1_000_000, 2)}M`
  if (amount >= 1_000) return `£${trimmed(amount / 1_000, 1)}K`
  return `£${amount}`
}

function formatDob(iso) {
  if (!iso) return ''
  const [year, month, day] = iso.split('-').map(Number)
  return `${day}/${month}/${year}`
}

function escapeCell(value) {
  const text = value === undefined || value === null ? '' : String(value)
  return /[;"\r\n]/.test(text) ? `"${text.replaceAll('"', '""')}"` : text
}

/** One CSV row (array of cells, in CSV_HEADERS order). */
export function playerRow(player) {
  const nationCode = NATION_FIFA_CODES[player.nationDatabaseId]
  const position = formatPositions(player.positions)
  const identity = [
    player.name,
    nationCode ? '' : (player.nationName ?? ''),
    nationCode ?? '',
    player.clubName ?? '',
    player.leagueName ?? '',
    position,
    player.clubNationName ?? '',
    player.age ?? '',
    formatDob(player.birthDate),
    formatMoney(player.transferValue),
    player.wage === undefined ? '' : `${formatMoney(player.wage)} p/w`,
    player.uniqueId ?? '',
    player.leftFoot,
    player.rightFoot,
    ...personalityLabels(player),
  ]
  const attributes = Object.keys(ATTRIBUTE_HEADERS).map((key) => player.attributes[key])
  // seasonStats: undefined = stats unreadable (blank), null = no Overall line (zeros).
  const stats =
    player.seasonStats === undefined
      ? BLANK_STATS
      : seasonStatCells(player.seasonStats ?? undefined, position.startsWith('GK'))
  const saveOnly = SAVE_ONLY_COLUMNS.map(([, read]) => read(player) ?? '')
  return [...identity, ...attributes, ...stats, ...saveOnly]
}

/**
 * Build the CSV text. Players without a decoded name are skipped: FM-Dash rejects
 * nameless rows anyway.
 * @returns {{ csv: string, rows: number, skipped: number }}
 */
export function playersToCsv(players) {
  const lines = [CSV_HEADERS.join(';')]
  let skipped = 0
  for (const player of players) {
    if (!player.name) {
      skipped++
      continue
    }
    lines.push(playerRow(player).map(escapeCell).join(';'))
  }
  return { csv: `${lines.join('\n')}\n`, rows: lines.length - 1, skipped }
}
