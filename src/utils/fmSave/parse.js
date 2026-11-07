// End-to-end parsing of one FM27 save into player rows.
// Translated from fm27-scout-code (MIT, Copyright (c) 2026 FM27 Scout contributors),
// crates/fm27-core/src/parse.rs — the player path only (no finances, stadiums,
// injuries, suspensions or roles). See ./THIRD_PARTY.md.

import { MISSING_U32, parseError, readI16, readU8, readU16, readU32 } from './bytes.js'
import { ClubIndex, clubDisplayName } from './clubs.js'
import { SaveFile } from './container.js'
import { ContractDecoder } from './contracts.js'
import { readInjuredPindexes } from './injuries.js'
import {
  ATTRIBUTE_KEYS,
  decodeDate,
  PlayerLayout as L,
  PERSONALITY_KEYS,
  POSITION_NAMES,
  scaleAttribute,
  storedAttributeIndex,
} from './layout.js'
import { resolveClubLeagues } from './leagues.js'
import { readMetadata } from './metadata.js'
import { NamePools } from './names.js'
import { NationIndex } from './nations.js'
import { PersonDecoder } from './persons.js'
import { closingPersonHeader, locatePlayerRecords } from './playerScan.js'
import { readOverallLines, readSeasonStatsBody } from './seasonStats.js'

const DEFAULT_CLOCK = '2026-07-01'

/**
 * Parse a save from a container source (see container.js bufferSource/blobSource).
 * @param {{size:number, read(offset:number, length:number):Promise<Uint8Array>}} source
 * @param {(progress: {stage: string, current?: number, total?: number}) => void} [onProgress]
 */
export async function parseSave(source, onProgress = () => {}) {
  const started = performance.now()
  onProgress({ stage: 'opening' })
  const save = await SaveFile.open(source)
  const metadata = await readMetadata(save)
  onProgress({ stage: 'decompressing' })
  let gameDb = await save.readSection('game_db')
  const clock = metadata.gameDate ?? DEFAULT_CLOCK
  const result = parseGameDb(gameDb, clock, onProgress)
  await attachInjuries(save, result.players, clock, result.report.warnings)
  // Drop game_db (~227 MB) before the stats frame (~73 MB) is decompressed.
  gameDb = undefined
  onProgress({ stage: 'stats' })
  Object.assign(
    result.report,
    await attachSeasonStats(save, result.players, result.report.warnings, onProgress)
  )
  return {
    metadata: { ...metadata, saveName: save.saveName },
    ...result,
    report: { ...result.report, elapsedMs: Math.round(performance.now() - started) },
  }
}

/**
 * Mark players injured on the game date (`player.injured`). A missing or unreadable
 * injury_manager adds a warning and leaves every player uninjured.
 */
export async function attachInjuries(save, players, clock, warnings) {
  try {
    const injured = readInjuredPindexes(await save.readSection('injury_manager'), clock)
    for (const player of players) player.injured = injured.has(player.pindex)
  } catch (error) {
    warnings.push(`injuries unavailable: ${error.message}`)
  }
}

/**
 * Join each player's current-season Overall line. Afterwards `player.seasonStats` is the
 * summed line, or null when the player has no Overall line (all zeros). When the stats
 * can't be read it stays undefined (blank cells), a warning is added, and the import goes on.
 */
export async function attachSeasonStats(save, players, warnings, onProgress = () => {}) {
  const started = performance.now()
  let statsFrameBytes = 0
  let withStats = 0
  try {
    let body = await readSeasonStatsBody(save)
    statsFrameBytes = body.length
    onProgress({ stage: 'stats', current: 1, total: 2 })
    const overall = readOverallLines(body)
    body = undefined
    for (const player of players) {
      const line = overall.get(player.pindex + 1)
      player.seasonStats = line ?? null
      if (line) withStats++
    }
  } catch (error) {
    for (const player of players) player.seasonStats = undefined
    warnings.push(`season stats unavailable: ${error.message}`)
  }
  return { withStats, statsMs: Math.round(performance.now() - started), statsFrameBytes }
}

/**
 * Decode every player in a decompressed game_db.
 * Club, league and contract joins are best-effort: when one cannot be read, players
 * still decode with those columns blank and the report says so.
 */
export function parseGameDb(gameDb, clock, onProgress = () => {}) {
  const warnings = []
  onProgress({ stage: 'names' })
  const names = NamePools.locate(gameDb)
  let nations
  try {
    nations = NationIndex.read(gameDb)
  } catch {
    nations = new NationIndex()
  }
  onProgress({ stage: 'scanning' })
  const index = locatePlayerRecords(gameDb, names.endOffset)
  onProgress({ stage: 'clubs' })
  let clubs
  try {
    clubs = ClubIndex.read(gameDb)
    for (const club of clubs.clubs) club.nationName = nations.nameFor(club.nationId)
  } catch (error) {
    clubs = undefined
    warnings.push(`club table unavailable: ${error.message}`)
  }
  let division = { divisionClubs: 0, divisionMs: 0 }
  if (clubs) {
    try {
      const { divisionClubs, divisionMs } = resolveClubLeagues(gameDb, clubs, nations, clock)
      division = { divisionClubs, divisionMs }
    } catch (error) {
      warnings.push(`club leagues unavailable: ${error.message}`)
    }
  }
  const contracts = new ContractDecoder(clubs, clock)
  const personDecoder = new PersonDecoder(names, clock)

  const players = []
  let withPerson = 0
  let withClub = 0
  const total = index.records.length
  for (let position = 0; position < total; position++) {
    if (position % 5000 === 0) onProgress({ stage: 'decoding', current: position, total })
    const record = index.records[position]
    const offset = record.offset
    const windowEnd = index.windowEnd(position, gameDb.length)

    const teamIdRaw = readU32(gameDb, offset + L.TEAM_ID_OFFSET) ?? MISSING_U32
    const teamId = teamIdRaw === MISSING_U32 ? undefined : teamIdRaw
    const club = teamId === undefined ? undefined : clubs?.teamFields(teamId)
    if (club) withClub++

    const closing = closingPersonHeader(gameDb, record.pindex, offset + L.DECODE_EXTENT, windowEnd)
    const person = personDecoder.decode(gameDb, offset + 100, closing ? closing[0] : windowEnd)
    if (person) withPerson++

    const contract = contracts.decode(
      gameDb,
      offset,
      windowEnd,
      position + 1 === total,
      record.pindex,
      club?.uid
    )

    const attributeStart = offset + L.ATTRIBUTES_OFFSET
    const attributes = {}
    ATTRIBUTE_KEYS.forEach((key, i) => {
      attributes[key] = scaleAttribute(
        readU8(gameDb, attributeStart + storedAttributeIndex(i)) ?? 0
      )
    })
    const positions = {}
    POSITION_NAMES.forEach((name, i) => {
      positions[name] = readU8(gameDb, offset + L.RATINGS_OFFSET + i) ?? 1
    })

    const transferValueRaw = readU32(gameDb, offset + L.TRANSFER_VALUE_OFFSET) ?? MISSING_U32
    const transferValue =
      transferValueRaw === MISSING_U32 ||
      transferValueRaw === L.TRANSFER_VALUE_PLACEHOLDER ||
      transferValueRaw === 0
        ? undefined
        : transferValueRaw
    const rawPa = readI16(gameDb, offset + L.PA_OFFSET) ?? -1

    players.push({
      uid: record.uid,
      pindex: record.pindex, // internal: the season-stats key is pindex + 1; not exported
      uniqueId: closing?.[1],
      name: person?.name,
      firstName: person?.firstName,
      lastName: person?.lastName,
      commonName: person?.commonName,
      birthDate: person?.birthDate,
      age: person?.age,
      nationId: person?.nationId,
      nationName: person ? nations.nameFor(person.nationId) : undefined,
      nationDatabaseId: person ? nations.databaseIdFor(person.nationId) : undefined,
      secondNationIds: person?.secondNationIds ?? [],
      personality: person
        ? Object.fromEntries(PERSONALITY_KEYS.map((key, i) => [key, person.personality[i]]))
        : undefined,
      heightCm: readU8(gameDb, offset + L.HEIGHT_OFFSET) ?? 0,
      ca: readU16(gameDb, offset + L.CA_OFFSET) ?? 0,
      pa: rawPa < 0 ? undefined : rawPa,
      teamId,
      clubUid: club?.uid,
      clubName: clubDisplayName(club),
      clubLongName: club?.name,
      clubNationName: club?.nationName,
      leagueName: club?.leagueName,
      // internal: the Division's competition database id (scripts/fm27-division-names.mjs)
      leagueDatabaseId: club?.leagueCompetition?.databaseId,
      gender: club?.gender,
      clubJoinDate: decodeDate(gameDb, offset + L.CLUB_JOIN_DATE_OFFSET),
      attributes,
      leftFoot: scaleAttribute(readU8(gameDb, attributeStart + L.LEFT_FOOT_INDEX) ?? 0),
      rightFoot: scaleAttribute(readU8(gameDb, attributeStart + L.RIGHT_FOOT_INDEX) ?? 0),
      positions,
      transferValue,
      wage: contract?.wage,
      contractEnd: contract?.end,
      listingFlags: contract?.listingFlags,
      listingCode: contract?.listingCode,
    })
  }

  if (players.length === 0) throw parseError('no players decoded')
  onProgress({ stage: 'done', current: players.length, total: players.length })
  return {
    players,
    clubs, // the ClubIndex (undefined when unreadable), for the managed-club lookup
    report: {
      players: players.length,
      markerless: index.markerless,
      withPerson,
      withClub,
      withContract: players.filter((player) => player.wage !== undefined).length,
      withDivision: players.filter((player) => player.leagueName).length,
      ...division,
      warnings,
    },
  }
}
