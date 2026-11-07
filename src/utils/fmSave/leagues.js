// Competition index (stage table + editor-database-id pairs) and best-effort club leagues.
// Translated from fm27-scout-code (MIT, Copyright (c) 2026 FM27 Scout contributors),
// crates/fm27-core/src/{competitions,leagues,reference}.rs. See ./THIRD_PARTY.md.
//
// The save stores no competition names; they come from the bundled table keyed on the
// editor database id. A club's league comes from the match log (FM-Dash's own decode; see
// scanMatchLogLeagues), else - for the top division of seven nations - from the same
// competition id appearing in all three season windows of the club record. Other clubs
// keep no league rather than a guessed one.

import { findBytes, MISSING_U32, readU32 } from './bytes.js'
import { COMPETITION_NAMES } from './competition-names.js'
import { COMPETITION_SHORT_NAMES } from './competition-short-names.js'

const ROW_BYTES = 33
const PREVIOUS_STAGE_ID_OFFSET = 0
const STAGE_ID_OFFSET = 4
const STAGE_ID_COPY_OFFSET = 8
const ZERO_BYTE_OFFSET = 12
const COMPETITION_ID_OFFSET = 13
const STAGE_ID_MAX_EXCLUSIVE = 200_000
const COMPETITION_ID_LIMIT = 1_000_000
const SEARCH_BYTES = 2_000_000
const CHAIN_ROWS = 200
const RESYNCHRONISATION_BYTES = 4_096

const PAIR_MARKER = new Uint8Array([...new Array(16).fill(0xff), 0x01])
const RECORD_OFFSET_FROM_MARKER = 30
const ENTITY_ID_RANGE = [1, 199_999]
const DATABASE_ID_MAX = 0x7fffffff
const CONSTANT_BYTES = [
  [-8, 7],
  [-6, 0],
  [-4, 7],
  [-1, 255],
  [12, 0],
  [15, 1],
  [22, 0],
  [23, 0],
  [24, 0],
  [25, 0],
  [26, 1],
  [32, 2],
  [40, 3],
  [55, 6],
  [60, 7],
]

const LEAGUE_WINDOWS = [
  [350, 500],
  [600, 750],
  [1100, 1250],
]

/** @returns {Array<{id:number, databaseId?:number, name?:string}>} */
export function readCompetitions(gameDb) {
  const competitionIds = readStageTable(gameDb)
  const databaseIds = locateDatabaseIds(gameDb)
  const competitionsByDatabaseId = new Map()
  for (const id of competitionIds) {
    const databaseId = databaseIds.get(id)
    if (databaseId === undefined) continue
    if (!competitionsByDatabaseId.has(databaseId)) competitionsByDatabaseId.set(databaseId, [])
    competitionsByDatabaseId.get(databaseId).push(id)
  }
  const contested = new Set(
    [...competitionsByDatabaseId.values()].filter((ids) => ids.length > 1).flat()
  )
  return [...competitionIds]
    .sort((a, b) => a - b)
    .map((id) => {
      const databaseId = contested.has(id) ? undefined : databaseIds.get(id)
      return {
        id,
        databaseId,
        name: databaseId === undefined ? undefined : COMPETITION_NAMES[databaseId],
      }
    })
}

function readStageTable(gameDb) {
  const tableStart = findTableStart(gameDb)
  if (tableStart === undefined) throw new Error('no stage table was found in game_db')
  const competitionIds = new Set()
  const seenStageIds = new Set()
  for (const rowOffset of walkRows(gameDb, tableStart)) {
    const stageId = readU32(gameDb, rowOffset + STAGE_ID_OFFSET)
    if (seenStageIds.has(stageId)) throw new Error(`stage id ${stageId} appears in two stage rows`)
    seenStageIds.add(stageId)
    const competitionId = readU32(gameDb, rowOffset + COMPETITION_ID_OFFSET)
    if (
      competitionId === undefined ||
      competitionId === MISSING_U32 ||
      competitionId >= COMPETITION_ID_LIMIT
    )
      continue
    competitionIds.add(competitionId)
  }
  return competitionIds
}

function rowDecodes(gameDb, rowOffset) {
  if (rowOffset < 0 || rowOffset + ROW_BYTES > gameDb.length) return false
  if (gameDb[rowOffset + ZERO_BYTE_OFFSET] !== 0) return false
  const stageId = readU32(gameDb, rowOffset + STAGE_ID_OFFSET)
  if (readU32(gameDb, rowOffset + STAGE_ID_COPY_OFFSET) !== stageId) return false
  if (!(stageId > 0 && stageId < STAGE_ID_MAX_EXCLUSIVE)) return false
  const previous = readU32(gameDb, rowOffset + PREVIOUS_STAGE_ID_OFFSET)
  return previous === stageId - 1 || previous === MISSING_U32
}

function chainStartsHere(gameDb, rowOffset) {
  for (let step = 0; step < CHAIN_ROWS; step++) {
    if (!rowDecodes(gameDb, rowOffset + step * ROW_BYTES)) return false
  }
  return true
}

function chainIdSteps(gameDb, rowOffset) {
  let previous
  let steps = 0
  for (let step = 0; step < CHAIN_ROWS; step++) {
    const stageId = readU32(gameDb, rowOffset + step * ROW_BYTES + STAGE_ID_OFFSET)
    if (stageId === undefined) break
    if (previous !== undefined && stageId === previous + 1) steps++
    previous = stageId
  }
  return steps
}

function findTableStart(gameDb) {
  const length = gameDb.length
  let searchedFrom = length
  let windowBytes = SEARCH_BYTES
  while (searchedFrom > 0) {
    const windowStart = Math.max(0, length - windowBytes)
    const start = findTableStartBetween(gameDb, windowStart, searchedFrom)
    if (start !== undefined) return start
    searchedFrom = windowStart
    windowBytes *= 2
  }
  return undefined
}

function findTableStartBetween(gameDb, firstCandidate, candidatesEnd) {
  const length = gameDb.length
  const lastCandidate = Math.min(candidatesEnd, Math.max(0, length - ROW_BYTES) + 1)
  for (let candidate = firstCandidate; candidate < lastCandidate; candidate++) {
    if (gameDb[candidate + ZERO_BYTE_OFFSET] !== 0) continue
    if (!rowDecodes(gameDb, candidate) || !chainStartsHere(gameDb, candidate)) continue
    const chainEnd = Math.min(candidate + ROW_BYTES, length)
    let tableStart = candidate
    let bestSteps = chainIdSteps(gameDb, candidate)
    for (let offset = candidate; offset < chainEnd; offset++) {
      if (!chainStartsHere(gameDb, offset)) continue
      const steps = chainIdSteps(gameDb, offset)
      if (steps > bestSteps) {
        bestSteps = steps
        tableStart = offset
      }
    }
    while (tableStart >= ROW_BYTES && rowDecodes(gameDb, tableStart - ROW_BYTES))
      tableStart -= ROW_BYTES
    return tableStart
  }
  return undefined
}

function walkRows(gameDb, tableStart) {
  const length = gameDb.length
  const rows = []
  let rowOffset = tableStart
  while (rowOffset + ROW_BYTES <= length) {
    if (rowDecodes(gameDb, rowOffset)) {
      rows.push(rowOffset)
      rowOffset += ROW_BYTES
      continue
    }
    const furthest = Math.min(rowOffset + RESYNCHRONISATION_BYTES, Math.max(0, length - ROW_BYTES))
    let next
    for (let candidate = rowOffset + 1; candidate <= furthest; candidate++) {
      if (gameDb[candidate + ZERO_BYTE_OFFSET] === 0 && rowDecodes(gameDb, candidate)) {
        next = candidate
        break
      }
    }
    if (next === undefined) break
    rowOffset = next
  }
  return rows
}

function locateDatabaseIds(gameDb) {
  const values = new Map()
  const conflicts = new Set()
  for (
    let markerAt = findBytes(gameDb, PAIR_MARKER, 0);
    markerAt !== -1;
    markerAt = findBytes(gameDb, PAIR_MARKER, markerAt + 1)
  ) {
    const recordStart = markerAt + RECORD_OFFSET_FROM_MARKER
    const entityId = readU32(gameDb, recordStart)
    const databaseId = readU32(gameDb, recordStart + 4)
    const databaseIdCopy = readU32(gameDb, recordStart + 8)
    if (databaseIdCopy === undefined || databaseId !== databaseIdCopy) continue
    if (entityId < ENTITY_ID_RANGE[0] || entityId > ENTITY_ID_RANGE[1]) continue
    if (databaseId < 1 || databaseId > DATABASE_ID_MAX) continue
    const constantsHold = CONSTANT_BYTES.every(([offset, value]) => {
      const at = recordStart + offset
      return at >= 0 && at < gameDb.length && gameDb[at] === value
    })
    if (!constantsHold) continue
    const existing = values.get(entityId)
    if (existing === undefined) values.set(entityId, databaseId)
    else if (existing !== databaseId) conflicts.add(entityId)
  }
  for (const entityId of conflicts) values.delete(entityId)
  return values
}

function isTopLeague(nationDatabaseId, name) {
  switch (nationDatabaseId) {
    case 765:
      return name === 'Premier League' // England
    case 776:
      return name === 'Italian Serie A' // Italy
    case 796:
      return name === 'Spanish First Division' // Spain
    case 771:
      return name === 'Bundesliga' // Germany
    case 769:
      return name.startsWith('Ligue 1') // France
    case 784:
      return name === 'VriendenLoterij Eredivisie' || name === 'Eredivisie' // Netherlands
    case 788:
      return name === 'Portuguese Premier League' // Portugal
    default:
      return false
  }
}

function appearsInAllWindows(gameDb, span, competitionId) {
  const recordLength = Math.max(0, span.recordEnd - span.recordStart)
  let hits = 0
  for (const [windowStart, windowEnd] of LEAGUE_WINDOWS) {
    const start = Math.min(windowStart, recordLength)
    const end = Math.min(windowEnd, recordLength)
    if (start >= end) continue
    for (let offset = start; offset < Math.max(0, end - 3); offset++) {
      if (readU32(gameDb, span.recordStart + offset) === competitionId) {
        hits++
        break
      }
    }
  }
  return hits === LEAGUE_WINDOWS.length
}

// --- Match-log Division -------------------------------------------------------------
// game_db logs every match a team played as `01 | u16 day (& 0x1FF) | u16 year | u32 team
// | u32 competition`, followed by `00 01 03 02` (full log) or `0C 00 01` (compact log).
// A club's Division is the league of its first team's latest logged match.

const MATCH_LOG_ENTRY_BYTES = 13
const DAY_MASK = 0x1ff
// Competitions whose bundled name marks them as something other than a league. Plain
// "Super" is not excluded: 3F Superliga, the Swiss/Turkish/Greek Super Leagues are leagues.
const NOT_LEAGUE =
  /Friendly|Cup|Copa|Champions League|Europa League|Conference League|Libertadores|Sudamericana|Recopa|State|Shield|Trophy|International|Qualif|Nations|World|Charity|Olympic|U1\d|U2\d|Playoff|Play-off/i
// A competition counts for a club when seen this often for it, or for this many clubs of
// its nation; single hits are byte-pattern noise.
const MIN_CLUB_HITS = 2
const MIN_NATION_CLUBS = 3

export function isLeagueName(name) {
  return Boolean(name) && !NOT_LEAGUE.test(name)
}

const isMatchLogTail = (gameDb, at) =>
  (gameDb[at] === 0x00 &&
    gameDb[at + 1] === 0x01 &&
    gameDb[at + 2] === 0x03 &&
    gameDb[at + 3] === 0x02) ||
  (gameDb[at] === 0x0c && gameDb[at + 1] === 0x00 && gameDb[at + 2] === 0x01)

/**
 * Each club's latest league from the match log.
 * @param {import('./clubs.js').ClubIndex} clubs
 * @param {Map<number, {id:number, databaseId?:number, name?:string}>} competitionsById
 * @param {number} gameYear
 * @returns {Map<number, object>} club uid -> competition
 */
export function scanMatchLogLeagues(gameDb, clubs, competitionsById, gameYear) {
  const leagueIds = new Set(
    [...competitionsById.values()].filter((c) => isLeagueName(c.name)).map((c) => c.id)
  )
  const view = new DataView(gameDb.buffer, gameDb.byteOffset, gameDb.byteLength)
  const entriesByClub = new Map() // club uid -> [[dayKey, competitionId]]
  const last = gameDb.length - MATCH_LOG_ENTRY_BYTES - 4
  for (let p = 0; p <= last; p++) {
    if (gameDb[p] !== 1) continue
    const year = view.getUint16(p + 3, true)
    if (year < gameYear - 2 || year > gameYear + 1) continue
    const competitionId = view.getUint32(p + 9, true)
    if (!leagueIds.has(competitionId)) continue
    if (!isMatchLogTail(gameDb, p + MATCH_LOG_ENTRY_BYTES)) continue
    const day = view.getUint16(p + 1, true) & DAY_MASK
    if (day < 1 || day > 366) continue
    const hit = clubs.teamToClub.get(view.getUint32(p + 5, true))
    if (!hit || hit[1] !== 0) continue // first teams only: reserves and youth play other leagues
    if (!entriesByClub.has(hit[0])) entriesByClub.set(hit[0], [])
    entriesByClub.get(hit[0]).push([year * 1000 + day, competitionId])
  }

  const nationOf = (uid) => clubs.clubByUid.get(uid)?.nationId
  const nationClubs = new Map() // `${nation}|${competition}` -> Set(club uid)
  for (const [uid, entries] of entriesByClub) {
    for (const [, competitionId] of entries) {
      const key = `${nationOf(uid)}|${competitionId}`
      if (!nationClubs.has(key)) nationClubs.set(key, new Set())
      nationClubs.get(key).add(uid)
    }
  }
  const leagues = new Map()
  for (const [uid, entries] of entriesByClub) {
    const hits = new Map()
    for (const [, competitionId] of entries)
      hits.set(competitionId, (hits.get(competitionId) ?? 0) + 1)
    const counts = (competitionId) =>
      hits.get(competitionId) >= MIN_CLUB_HITS ||
      (nationClubs.get(`${nationOf(uid)}|${competitionId}`)?.size ?? 0) >= MIN_NATION_CLUBS
    let latest
    for (const entry of entries) {
      if (counts(entry[1]) && (!latest || entry[0] > latest[0])) latest = entry
    }
    if (latest) leagues.set(uid, competitionsById.get(latest[1]))
  }
  return leagues
}

/** The Division FM prints for a competition: its short name, else the bundled long name. */
export function divisionName(competition) {
  if (!competition) return undefined
  return COMPETITION_SHORT_NAMES[competition.databaseId] ?? competition.name
}

/**
 * Set `leagueName` (and `leagueCompetition`) on every club whose league can be derived:
 * the match-log scan first, else the three-window top-division rule, else nothing.
 * @param {import('./clubs.js').ClubIndex} clubs
 * @param {import('./nations.js').NationIndex} nations
 * @param {string} clock the game date, YYYY-MM-DD
 * @param {Array} [competitions] the competition index (read from game_db when omitted)
 * @returns {{competitions: Array, divisionClubs: number, divisionMs: number}}
 */
export function resolveClubLeagues(gameDb, clubs, nations, clock, competitions) {
  const started = performance.now()
  competitions ??= readCompetitions(gameDb)
  const competitionsById = new Map(competitions.map((c) => [c.id, c]))
  const gameYear = Number(clock.slice(0, 4))
  const fromLog = scanMatchLogLeagues(gameDb, clubs, competitionsById, gameYear)

  const topLeagueByNation = new Map()
  let divisionClubs = 0
  for (const club of clubs.clubs) {
    let league = fromLog.get(club.uid)
    if (!league) {
      const nationDatabaseId = nations.databaseIdFor(club.nationId)
      if (nationDatabaseId !== undefined && !topLeagueByNation.has(nationDatabaseId)) {
        const ids = competitions.filter((c) => c.name && isTopLeague(nationDatabaseId, c.name))
        topLeagueByNation.set(
          nationDatabaseId,
          ids.length ? ids.reduce((a, b) => (a.id <= b.id ? a : b)) : undefined
        )
      }
      const top = topLeagueByNation.get(nationDatabaseId)
      const span = clubs.recordSpans.get(club.uid)
      if (top && span && appearsInAllWindows(gameDb, span, top.id)) league = top
    }
    const name = divisionName(league)
    if (name) {
      club.leagueName = name
      club.leagueCompetition = league
      divisionClubs++
    }
  }
  return { competitions, divisionClubs, divisionMs: Math.round(performance.now() - started) }
}
