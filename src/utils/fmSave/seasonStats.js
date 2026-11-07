// Current-season performance statistics from the unlisted frame after `non_pl_hist_ls`.
// The record walk is translated from fm27-scout-code (MIT, Copyright (c) 2026 FM27 Scout
// contributors), crates/fm27-core/src/season_stats.rs. See ./THIRD_PARTY.md.
// Changes from upstream: pressures attempted/completed are read (u8 @121/@123, which
// upstream says FM27 lacks), goalkeeper clearances @95 are kept, and only the Overall slot
// is decoded. Column formulas and display rules were measured against an in-game FM27
// export (fm27.csv), see .scratch/fm27-save-extras/issues/04-locate-performance-stats.md.
//
// One record per player, keyed by pindex + 1: eight own-team slots (one per competition
// kind), a 20-byte block, then six slots for each other team the player appeared for.
// A slot is either 0 (absent) or a 155-byte line starting `01 07`.

import { parseError, readU32 } from './bytes.js'

const SECTION_HEADER = [0x03, 0x01, 0x74, 0x61, 0x64, 0x2e, 0x0f, 0x00] // "\x03\x01tad." schema 15
const RECORD_PRESENT = 1
const RECORD_END = 0
const RECORD_HEADER_BYTES = 16
const OWN_SLOTS = 8
const OTHER_TEAM_SLOTS = 6
const BLOCK_BYTES = 20
const MAX_OTHER_TEAMS = 16
const RATING_ITEM_BYTES = 8
const MAX_RATING_ITEMS = 256
const FORM_ENTRY_BYTES = 3
const MAX_FORM_ENTRIES = 16
const FOOTER_BYTES = 16
const LINE_BYTES = 155
const OVERALL_SLOT = 5 // league + cup + continental, what FM's player views show
export const SEASON_STATS_REGION = 'non_pl_hist_ls'

// Line fields: [name, offset, width]. Width 1 = u8, 2 = u16, -2 = i16. Some words mean
// different things for goalkeepers (noted); both readings are kept and picked at export.
const LINE_FIELDS = [
  ['ratingSum', 2, 2],
  ['minutes', 4, 2],
  ['starts', 6, 1],
  ['subs', 7, 1],
  ['rated', 8, 1],
  ['goals', 9, 1],
  ['assists', 10, 1],
  ['conceded', 11, 1],
  ['cleanSheets', 17, 1],
  ['passesAttempted', 27, 2],
  ['passesCompleted', 29, 2],
  ['tacklesAttempted', 31, 2],
  ['tacklesCompleted', 33, 2],
  ['savesHeld', 35, 2], // goalkeepers
  ['headersWonOrParried', 37, 2], // outfield: headers won; goalkeepers: saves parried
  ['dribbles', 39, 2],
  ['foulsAgainst', 43, 2],
  ['shots', 45, 2],
  ['shotsOnTarget', 47, 2],
  ['distanceTenths', 65, 2],
  ['savesTipped', 69, 2], // goalkeepers
  ['keyPasses', 73, 2],
  ['keyTackles', 75, 2],
  ['interceptions', 79, 2],
  ['chancesCreated', 81, 2],
  ['blocksOrFaced', 83, 2], // outfield: blocks; goalkeepers: shots on target faced
  ['possessionLost', 87, 2],
  ['possessionWon', 89, 2],
  ['xgHundredths', 91, 2],
  ['xaHundredths', 93, 2],
  ['clearances', 95, 2],
  ['xgpHundredths', 113, -2],
  ['progressivePasses', 119, 2],
  ['pressuresAttempted', 121, 1],
  ['pressuresCompleted', 123, 1],
  ['npxgHundredths', 127, 2],
  ['crossesAttempted', 131, 2],
  ['crossesCompleted', 133, 2],
]

/** Decode the fields FM-Dash uses from one 155-byte line. */
export function decodeLine(body, offset) {
  const line = {}
  for (const [name, at, width] of LINE_FIELDS) {
    const p = offset + at
    const u16 = body[p] | (body[p + 1] << 8)
    line[name] = width === 1 ? body[p] : width === 2 ? u16 : (u16 << 16) >> 16
  }
  return line
}

/** Field-by-field sum of lines (a player's own Overall line plus other-team Overall lines). */
export function sumLines(lines) {
  if (lines.length === 1) return lines[0]
  const total = {}
  for (const [name] of LINE_FIELDS) total[name] = lines.reduce((sum, line) => sum + line[name], 0)
  return total
}

function readSlot(body, position) {
  const byte = body[position]
  if (byte === undefined) throw parseError('a season-statistics slot runs past the section')
  if (byte === 0) return [undefined, position + 1]
  if (position + LINE_BYTES > body.length || byte !== 0x01 || body[position + 1] !== 0x07) {
    throw parseError(`the season-statistics slot at byte ${position} is neither absent nor a line`)
  }
  return [position, position + LINE_BYTES]
}

function readCount(body, position, maximum, what) {
  const count = readU32(body, position)
  if (count === undefined) throw parseError(`a season-statistics ${what} count is truncated`)
  if (count > maximum) throw parseError(`a season-statistics record counts ${count} ${what}`)
  return count
}

export function isSeasonStatsBody(body) {
  return SECTION_HEADER.every((byte, i) => body[i] === byte)
}

/**
 * Walk the section and return each player's summed Overall line, keyed by pindex + 1.
 * Every count, slot and the footer are checked, so a moved layout fails closed.
 * @returns {Map<number, object>}
 */
export function readOverallLines(body) {
  if (!isSeasonStatsBody(body)) {
    throw parseError('the season-statistics section does not start with its header')
  }
  const overall = new Map()
  let position = SECTION_HEADER.length
  for (;;) {
    if (position >= body.length) {
      throw parseError('the season-statistics record list runs past the section')
    }
    const marker = body[position]
    if (marker === RECORD_END) break
    if (marker !== RECORD_PRESENT || position + 1 + 4 + RECORD_HEADER_BYTES > body.length) {
      throw parseError(`the season-statistics record at byte ${position} does not start as one`)
    }
    const key = readU32(body, position + 1)
    if (overall.has(key)) throw parseError(`the season-statistics records repeat player key ${key}`)
    position += 1 + 4 + RECORD_HEADER_BYTES
    const lines = []
    for (let slot = 0; slot < OWN_SLOTS; slot++) {
      const [line, next] = readSlot(body, position)
      if (slot === OVERALL_SLOT && line !== undefined) lines.push(decodeLine(body, line))
      position = next
    }
    position += BLOCK_BYTES
    const otherTeams = readCount(body, position, MAX_OTHER_TEAMS, 'other teams')
    position += 4
    for (let team = 0; team < otherTeams; team++) {
      position += 1 + 4 // lead byte + team id
      for (let slot = 0; slot < OTHER_TEAM_SLOTS; slot++) {
        const [line, next] = readSlot(body, position)
        if (slot === OVERALL_SLOT && line !== undefined) lines.push(decodeLine(body, line))
        position = next
      }
    }
    const ratingItems = readCount(body, position, MAX_RATING_ITEMS, 'rating items')
    position += 4 + ratingItems * RATING_ITEM_BYTES
    const formEntries = body[position]
    if (formEntries === undefined) throw parseError('a season-statistics form list is truncated')
    if (formEntries > MAX_FORM_ENTRIES) {
      throw parseError(`a season-statistics form list counts ${formEntries} entries`)
    }
    position += 1 + formEntries * FORM_ENTRY_BYTES
    overall.set(key, lines.length ? sumLines(lines) : undefined)
  }
  const footerEnd = position + 1 + FOOTER_BYTES
  if (footerEnd !== body.length) {
    throw parseError(
      `the season-statistics list ends ${body.length - footerEnd} bytes away from its footer`
    )
  }
  return overall
}

/**
 * Find and decompress the season-statistics frame: gap frames are tried in descending
 * compressed size (the stats frame is the largest on real saves), first header match wins.
 * @param {import('./container.js').SaveFile} save
 */
export async function readSeasonStatsBody(save) {
  const frames = await save.unlistedFrames(SEASON_STATS_REGION)
  frames.sort((a, b) => b.compressedSize - a.compressedSize)
  for (const span of frames) {
    const body = await save.readFrame(span, 'a season-statistics candidate frame')
    if (isSeasonStatsBody(body)) return body
  }
  throw parseError(`no frame after "${SEASON_STATS_REGION}" holds the season-statistics section`)
}

// --- CSV cells -------------------------------------------------------------------------

// fm27.csv's stat headers, in its column order.
export const SEASON_STAT_HEADERS = [
  'Pas %',
  'Sv %',
  'Ps A/90',
  'Asts/90',
  'Tck R',
  'Tck/90',
  'Cr C/A',
  'Fouls Against',
  'Cln/90',
  'Clean Sheets',
  'Shot/90',
  'Ch C/90',
  'Cr C/90',
  'Rating',
  'Apps',
  'Minutes',
  'Hdrs W/90',
  'Crs A/90',
  'Average Rating Club',
  'Drb/90',
  'KP/90',
  'xGP/90',
  'K Tck/90',
  'Clr/90',
  'Blk/90',
  'Poss Won/90',
  'Int/90',
  'Pres A/90',
  'Pres C/90',
  'ShT/90',
  'Poss Lost/90',
  'Pr passes/90',
  'Dist/90',
  'Ps C/90',
  'Con/90',
  'Saves/90',
  'NP-xG/90',
  'xG/90',
  'xA/90',
  'Goals per 90 minutes',
  'Mins/Gl',
  'Conv %',
]

// Cells left empty when FM zeroed the line's advanced block: those zeros are not real.
export const ADVANCED_STAT_HEADERS = new Set([
  'Drb/90',
  'Dist/90',
  'KP/90',
  'K Tck/90',
  'Int/90',
  'Ch C/90',
  'Poss Won/90',
  'Poss Lost/90',
  'xGP/90',
  'Pr passes/90',
  'Pres A/90',
  'Pres C/90',
  'Crs A/90',
  'Cr C/90',
  'Cr C/A',
  'Hdrs W/90',
  'Blk/90',
  'Clr/90',
  'Sv %',
  'Saves/90',
])

const EMPTY_LINE = Object.freeze(Object.fromEntries(LINE_FIELDS.map(([name]) => [name, 0])))

/** A line of zeros with the given fields set (tests and fixtures). */
export const statLine = (fields = {}) => ({ ...EMPTY_LINE, ...fields })

/** FM zeroes the advanced block together; distance and possession lost both 0 detects it. */
export function advancedBlockMissing(line) {
  return line.minutes > 0 && line.distanceTenths === 0 && line.possessionLost === 0
}

const percent = (part, whole) => (whole ? (100 * part) / whole : 0)

/**
 * Two decimals the way FM prints them: an exact tie (0.125) rounds half to even, like C's
 * printf ("0.12"); JS toFixed would round it up. Everything else is toFixed's correct
 * rounding of the exact double. A double sits exactly on a 2-dp tie only when it is an odd
 * multiple of 1/8, and `value * 8` is exact, so the test has no float noise.
 */
export function fixed2(value) {
  const eighths = value * 8
  if (Number.isInteger(eighths) && eighths % 2 !== 0) {
    let hundredths = Math.floor(value * 100)
    if (hundredths % 2 !== 0) hundredths += 1
    return (hundredths / 100).toFixed(2)
  }
  return value.toFixed(2)
}
// FM prints Mins/Gl like .NET's shortest double ("190.529411764706"): 15 significant digits.
const fifteenDigits = (value) => String(Number(value.toPrecision(15)))

/**
 * The 42 stat cells for one player, in SEASON_STAT_HEADERS order.
 * @param {object|undefined} line the summed Overall line (undefined = no Overall line: zeros)
 * @param {boolean} goalkeeper the exported Position starts with GK
 */
export function seasonStatCells(line, goalkeeper) {
  const s = line ?? EMPTY_LINE
  const minutes = s.minutes
  // FM shows 0.00 for every per-90 figure below 90 minutes.
  const per90 = (count) => (minutes >= 90 ? (count * 90) / minutes : 0)
  const p90 = (count) => fixed2(per90(count))
  const saves = s.savesHeld + s.headersWonOrParried + s.savesTipped
  const rating = s.rated ? fixed2(s.ratingSum / 10 / s.rated) : '-'

  const cells = {
    'Pas %': `${Math.round(percent(s.passesCompleted, s.passesAttempted))}%`,
    'Sv %': String(goalkeeper ? Math.floor(percent(saves, s.blocksOrFaced)) : 0),
    'Ps A/90': p90(s.passesAttempted),
    'Asts/90': p90(s.assists),
    'Tck R': String(Math.round(percent(s.tacklesCompleted, s.tacklesAttempted))),
    'Tck/90': p90(s.tacklesCompleted),
    'Cr C/A': String(Math.floor(percent(s.crossesCompleted, s.crossesAttempted))),
    'Fouls Against': String(s.foulsAgainst),
    'Cln/90': p90(s.cleanSheets),
    'Clean Sheets': String(s.cleanSheets),
    'Shot/90': p90(s.shots),
    'Ch C/90': p90(s.chancesCreated),
    'Cr C/90': p90(s.crossesCompleted),
    Rating: rating,
    Apps: s.subs ? `${s.starts} (${s.subs})` : String(s.starts),
    Minutes: String(minutes),
    'Hdrs W/90': p90(goalkeeper ? 0 : s.headersWonOrParried),
    'Crs A/90': p90(s.crossesAttempted),
    'Average Rating Club': rating,
    'Drb/90': p90(s.dribbles),
    'KP/90': p90(s.keyPasses),
    'xGP/90': fixed2(per90(s.xgpHundredths) / 100),
    'K Tck/90': p90(s.keyTackles),
    'Clr/90': p90(s.clearances),
    'Blk/90': p90(goalkeeper ? 0 : s.blocksOrFaced),
    'Poss Won/90': p90(s.possessionWon),
    'Int/90': p90(s.interceptions),
    'Pres A/90': p90(s.pressuresAttempted),
    'Pres C/90': p90(s.pressuresCompleted),
    'ShT/90': p90(s.shotsOnTarget),
    'Poss Lost/90': p90(s.possessionLost),
    'Pr passes/90': p90(s.progressivePasses),
    'Dist/90': fixed2(per90(s.distanceTenths) / 10),
    'Ps C/90': p90(s.passesCompleted),
    'Con/90': p90(s.conceded),
    'Saves/90': String(goalkeeper ? Math.floor(per90(saves)) : 0),
    'NP-xG/90': fixed2(per90(s.npxgHundredths) / 100),
    'xG/90': fixed2(per90(s.xgHundredths) / 100),
    'xA/90': fixed2(per90(s.xaHundredths) / 100),
    'Goals per 90 minutes': p90(s.goals),
    'Mins/Gl': s.goals ? fifteenDigits(minutes / s.goals) : '0',
    'Conv %': `${Math.round(percent(s.goals, s.shots))}%`,
  }
  const blankAdvanced = advancedBlockMissing(s)
  return SEASON_STAT_HEADERS.map((header) =>
    blankAdvanced && ADVANCED_STAT_HEADERS.has(header) ? '' : cells[header]
  )
}
