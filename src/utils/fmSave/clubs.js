// Club records and team lists in `game_db`.
// Translated from fm27-scout-code (MIT, Copyright (c) 2026 FM27 Scout contributors),
// crates/fm27-core/src/clubs.rs. The club-status table (reputation, last league
// position) is not ported: FM-Dash's export has no column for it. See ./THIRD_PARTY.md.
//
// Full saves are bounded by the counted club table, which also supplies the initial
// club (uid 0) and exact record ends. Small fragments fall back to the anchor scan.

import {
  decodeUtf8,
  findBytes,
  MISSING_U32,
  parseError,
  readU8,
  readU16,
  readU32,
} from './bytes.js'

const ANCHOR = new Uint8Array([0xff, 0xff, 0xff, 0xff])
const ANCHOR_OFFSET = 17
const CLUB_INDEX_OFFSET = 0
const UID_OFFSET = 4
const UID_COPY_OFFSET = 8
const ZERO_BYTE_OFFSET = 12
const NATION_OFFSET = 13
const FA_NATION_OFFSET = 21
const NATION_COPY_OFFSET = 25
const LONG_NAME_OFFSET = 39
const NATION_MAX = 999
const MAX_NAME_BYTES = 255
const STOP_GAP_BYTES = 64 * 1024
const NULL_DATE_TRIPLE = new Uint8Array([1, 0, 0x6c, 0x07, 1, 0, 0x6c, 0x07, 1, 0, 0x6c, 0x07])
const TEAM_COUNTS_OFFSET = 37
const FIRST_ENTRY_BYTES = 25
const FIRST_ENTRY_LEAD = 0x02
const SECOND_ENTRY_BYTES = 12
const SECOND_ENTRY_LEAD = 0x01
const FILLER_BYTES = 9
const TEAM_COUNT_RANGE = [1, 8]
const TEAM_ID_RANGE = [1, 3_000_000]
const AFFILIATE_COUNT_MAX = 8
const FLOAT_ANCHOR = new Uint8Array([0x00, 0x00, 0x80, 0x3f]) // 1.0f32 LE
const FLOAT_ANCHOR_OFFSET = 33
const MINIMUM_FLOAT_ANCHOR_RECORD_OFFSET = 40

// [countBytes, leadBytes] steps (tail bytes are always 0 in the original).
const AFTER_NAMES = [
  [0, 8],
  [1, 4],
]
const BASE_FINANCE_BYTES = 45
const FULL_FINANCE = [
  [0, 66],
  [0, 235],
  [1, 8],
  [1, 8],
  [0, 235],
  [1, 8],
  [1, 8],
  [0, 235],
  [1, 8],
  [1, 8],
  [0, 235],
  [1, 8],
  [1, 8],
  [0, 67],
  [1, 230],
  [4, 49],
  [0, 71],
  [1, 19],
  [0, 67],
]
const AFTER_FINANCE = [
  [1, 25],
  [1, 12],
  [0, 9],
  [1, 4],
  [1, 4],
  [2, 4],
  [2, 4],
  [2, 4],
  [1, 16],
  [2, 28],
]

const inRange = (value, [lo, hi]) => value >= lo && value <= hi

export class ClubIndex {
  static read(gameDb) {
    const [anchorRecords, anchorScanEnd] = scanClubRecords(gameDb)
    if (anchorRecords.length === 0) throw parseError('no club records found')

    let records = readCountedTable(gameDb, anchorRecords)
    if (!records) {
      records = anchorRecords
      records.forEach((record, i) => {
        record.recordEnd = records[i + 1]?.recordStart ?? anchorScanEnd
      })
      const initial = findInitialClub(gameDb, records)
      if (initial) records.unshift(initial)
    }
    if (records.length === 0) throw parseError('no club records found')

    const index = new ClubIndex()
    const uidByIndex = new Map()
    const teamToClub = new Map()
    const affiliateLists = []
    const clubs = []
    for (const record of records) {
      if (uidByIndex.has(record.clubIndex)) {
        throw parseError(`club index ${record.clubIndex} appears twice`)
      }
      uidByIndex.set(record.clubIndex, record.uid)
      if (index.nameByUid.has(record.uid)) throw parseError(`club uid ${record.uid} appears twice`)
      index.nameByUid.set(record.uid, record.name)

      const teamList = readTeamList(
        gameDb,
        record.recordStart,
        record.recordEnd,
        record.teamCountsStart
      )
      affiliateLists.push(teamList.affiliateTeamIds)
      const seen = new Set()
      teamList.teamIds.forEach((teamId, slot) => {
        if (seen.has(teamId))
          throw parseError(`team id ${teamId} is listed twice by club ${record.uid}`)
        seen.add(teamId)
        const other = teamToClub.get(teamId)
        if (other) {
          throw parseError(`team id ${teamId} is stored by clubs ${other[0]} and ${record.uid}`)
        }
        teamToClub.set(teamId, [record.uid, slot])
      })
      index.recordSpans.set(record.uid, {
        recordStart: record.recordStart,
        recordEnd: record.recordEnd,
      })
      clubs.push({
        uid: record.uid,
        uniqueId: record.uniqueId,
        name: record.name,
        shortName: record.shortName,
        nationId: record.nationId,
        faNationId: record.faNationId,
        gender: teamList.gender,
        teams: teamList.teamIds,
      })
    }

    const affiliateTeamToClub = new Map()
    affiliateLists.forEach((ids, position) => {
      const parentUid = records[position].uid
      const ownSlots = clubs[position].teams.length
      ids.forEach((teamId, i) => {
        if (teamToClub.get(teamId)?.[0] === parentUid) return
        if (!affiliateTeamToClub.has(teamId))
          affiliateTeamToClub.set(teamId, [parentUid, ownSlots + i])
      })
    })

    index.clubs = clubs
    index.uidByIndex = uidByIndex
    index.teamToClub = teamToClub
    index.affiliateTeamToClub = affiliateTeamToClub
    index.fieldingTeamToClub = new Map([...teamToClub, ...affiliateTeamToClub])
    index.clubByUid = new Map(clubs.map((club) => [club.uid, club]))
    return index
  }

  constructor() {
    this.clubs = []
    this.nameByUid = new Map()
    this.recordSpans = new Map()
    this.clubByUid = new Map()
    this.teamToClub = new Map()
    this.affiliateTeamToClub = new Map()
    this.fieldingTeamToClub = new Map()
    this.uidByIndex = new Map()
  }

  /** Club fields for a player's team id: `{uid, slot, ...club}` (affiliates resolve to the parent). */
  teamFields(teamId) {
    const hit = this.affiliateTeamToClub.get(teamId) ?? this.teamToClub.get(teamId)
    if (!hit) return undefined
    const club = this.clubByUid.get(hit[0])
    return club ? { ...club, slot: hit[1] } : undefined
  }
}

function readCountedTable(gameDb, anchorRecords) {
  const firstStart = anchorRecords[0].recordStart
  const prefixEnd = Math.min(firstStart + ANCHOR_OFFSET + 4, gameDb.length)
  const anchorDistance = ANCHOR_OFFSET - ZERO_BYTE_OFFSET
  const heads = new Set()
  for (
    let hit = findBytes(gameDb, ANCHOR, 0, prefixEnd);
    hit !== -1;
    hit = findBytes(gameDb, ANCHOR, hit + 1, prefixEnd)
  ) {
    for (const extra of [0, 4]) {
      const body = hit - anchorDistance - extra
      const head = body - 5
      if (head < 0 || head + 5 > gameDb.length || gameDb[head + 4] !== 0) continue
      if (gameDb[body] !== (extra === 0 ? 0 : 12)) continue
      const row = clubTableRecord(gameDb, body)
      if (!row) continue
      const count = readU32(gameDb, head) ?? 0
      const plausible = count >= 1 && count <= 200_000
      if (
        (plausible || row.footerStart === firstStart) &&
        row.uniqueId !== 0 &&
        row.uniqueId !== MISSING_U32
      ) {
        heads.add(head)
      }
    }
  }
  if (heads.size === 0) return undefined

  const tables = []
  for (const head of [...heads].sort((a, b) => a - b)) {
    const rows = walkClubTable(gameDb, head)
    if (!rows || rows.length === 0) continue
    const tableStart = rows[0].bodyStart
    const tableEnd = rows[rows.length - 1].footerStart + 12
    if (anchorRecords.every((r) => tableStart <= r.recordStart && r.recordStart < tableEnd))
      tables.push(rows)
  }
  if (tables.length !== 1) return undefined

  const rows = tables[0]
  const tableEnd = rows[rows.length - 1].footerStart + 12
  const starts = rows.map(
    (row) => row.bodyStart + (gameDb[row.bodyStart] === 12 ? 4 : 0) - ZERO_BYTE_OFFSET
  )
  const records = []
  for (let ordinal = 0; ordinal < rows.length; ordinal++) {
    const start = starts[ordinal]
    if (ordinal > 0 && !matchesAt(gameDb, start + ANCHOR_OFFSET, ANCHOR)) continue
    const uid = ordinal === 0 ? 0 : rows[ordinal - 1].uniqueId + 1
    const recordEnd = starts[ordinal + 1] ?? tableEnd
    const record = acceptTableRecord(
      gameDb,
      start,
      recordEnd,
      rows[ordinal].teamCountsStart,
      ordinal,
      uid,
      rows[ordinal].uniqueId
    )
    if (!record) {
      if (ordinal === 0) return undefined
      continue
    }
    records.push(record)
  }
  return records.length ? records : undefined
}

function matchesAt(bytes, at, needle) {
  if (at < 0 || at + needle.length > bytes.length) return false
  for (let i = 0; i < needle.length; i++) if (bytes[at + i] !== needle[i]) return false
  return true
}

function walkClubTable(gameDb, head) {
  const count = readU32(gameDb, head)
  if (count === undefined || count < 1 || count > 200_000 || readU8(gameDb, head + 4) !== 0)
    return undefined
  let cursor = head + 5
  const rows = []
  const seen = new Set()
  for (let ordinal = 0; ordinal < count; ordinal++) {
    const row = clubTableRecord(gameDb, cursor)
    if (!row) return undefined
    if (
      readU32(gameDb, row.footerStart) !== ordinal ||
      row.uniqueId === 0 ||
      row.uniqueId === MISSING_U32
    ) {
      return undefined
    }
    if (seen.has(row.uniqueId)) return undefined
    seen.add(row.uniqueId)
    rows.push(row)
    cursor = row.footerStart + 12
  }
  return rows
}

function clubTableRecord(gameDb, body) {
  const kind = gameDb[body]
  if (kind !== 0 && kind !== 12) return undefined
  let cursor = body + 1 + 26 + (kind === 12 ? 4 : 0)
  for (let i = 0; i < 2; i++) {
    const length = readU32(gameDb, cursor)
    if (length === undefined || length > 16_384) return undefined
    cursor += 4 + length
    if (cursor > gameDb.length) return undefined
  }
  cursor = walkSteps(gameDb, cursor, AFTER_NAMES)
  if (cursor === undefined || cursor >= gameDb.length) return undefined
  const finance = gameDb[cursor]
  cursor += 1
  if (finance === 0) cursor += BASE_FINANCE_BYTES
  else if (finance === 1) cursor = walkSteps(gameDb, cursor, FULL_FINANCE)
  else if (finance !== 2) return undefined
  if (cursor === undefined || cursor > gameDb.length) return undefined
  const teamCountsStart = cursor
  cursor = walkSteps(gameDb, cursor, AFTER_FINANCE)
  if (cursor === undefined || cursor + 12 > gameDb.length) return undefined
  return {
    bodyStart: body,
    footerStart: cursor,
    uniqueId: readU32(gameDb, cursor + 4),
    teamCountsStart,
  }
}

function walkSteps(gameDb, start, steps) {
  let position = start
  for (const [countBytes, leadBytes] of steps) {
    if (countBytes > 0) {
      const count =
        countBytes === 1
          ? readU8(gameDb, position)
          : countBytes === 2
            ? readU16(gameDb, position)
            : readU32(gameDb, position)
      if (count === undefined || count > 100_000) return undefined
      position += countBytes + count * leadBytes
    } else {
      position += leadBytes
    }
    if (position > gameDb.length) return undefined
  }
  return position
}

function readNations(gameDb, recordStart) {
  const nationId = readU32(gameDb, recordStart + NATION_OFFSET)
  const faNationId = readU32(gameDb, recordStart + FA_NATION_OFFSET)
  const nationCopy = readU32(gameDb, recordStart + NATION_COPY_OFFSET)
  for (const value of [nationId, faNationId, nationCopy]) {
    if (value === undefined || value > NATION_MAX) return undefined
  }
  return [nationId, faNationId]
}

function acceptTableRecord(
  gameDb,
  recordStart,
  recordEnd,
  teamCountsStart,
  clubIndex,
  uid,
  uniqueId
) {
  const nations = readNations(gameDb, recordStart)
  const names = nations && readClubNames(gameDb, recordStart)
  if (!names) return undefined
  return {
    recordStart,
    recordEnd,
    teamCountsStart,
    clubIndex,
    uid,
    uniqueId,
    nationId: nations[0],
    faNationId: nations[1],
    name: names[0],
    shortName: names[1],
  }
}

function readTeamList(gameDb, recordStart, recordEnd, teamCountsStart) {
  if (teamCountsStart !== undefined && teamCountsStart - TEAM_COUNTS_OFFSET >= 0) {
    const list = parseTeamList(gameDb, teamCountsStart - TEAM_COUNTS_OFFSET, recordEnd)
    if (list) return list
  }
  let best
  const consider = (start, list) => {
    if (list && (!best || start < best[0])) best = [start, list]
  }
  for (
    let at = findBytes(gameDb, NULL_DATE_TRIPLE, recordStart, recordEnd);
    at !== -1;
    at = findBytes(gameDb, NULL_DATE_TRIPLE, at + 1, recordEnd)
  ) {
    consider(at, parseTeamList(gameDb, at, recordEnd))
  }
  const floatFrom = recordStart + MINIMUM_FLOAT_ANCHOR_RECORD_OFFSET
  for (
    let hit = findBytes(gameDb, FLOAT_ANCHOR, floatFrom, recordEnd);
    hit !== -1;
    hit = findBytes(gameDb, FLOAT_ANCHOR, hit + 1, recordEnd)
  ) {
    const listStart = hit - FLOAT_ANCHOR_OFFSET
    if (listStart >= 0) consider(listStart, parseTeamList(gameDb, listStart, recordEnd))
  }
  return best?.[1] ?? { teamIds: [], affiliateTeamIds: [], gender: undefined }
}

function parseTeamList(gameDb, listStart, recordEnd) {
  let cursor = listStart + TEAM_COUNTS_OFFSET
  for (const [entryBytes, lead] of [
    [FIRST_ENTRY_BYTES, FIRST_ENTRY_LEAD],
    [SECOND_ENTRY_BYTES, SECOND_ENTRY_LEAD],
  ]) {
    if (cursor >= recordEnd || cursor >= gameDb.length) return undefined
    const count = gameDb[cursor]
    const entriesStart = cursor + 1
    const entriesEnd = entriesStart + count * entryBytes
    if (entriesEnd > recordEnd) return undefined
    for (let entry = 0; entry < count; entry++) {
      if (gameDb[entriesStart + entry * entryBytes] !== lead) return undefined
    }
    cursor = entriesEnd
  }
  const teamCountAt = cursor + FILLER_BYTES
  if (teamCountAt >= recordEnd || teamCountAt >= gameDb.length) return undefined
  const teamCount = gameDb[teamCountAt]
  if (!inRange(teamCount, TEAM_COUNT_RANGE)) return undefined
  const teamIdsStart = teamCountAt + 1
  const teamIdsEnd = teamIdsStart + teamCount * 4
  if (teamIdsEnd > recordEnd) return undefined
  const teamIds = []
  for (let i = 0; i < teamCount; i++) {
    const teamId = readU32(gameDb, teamIdsStart + i * 4)
    if (teamId === undefined || !inRange(teamId, TEAM_ID_RANGE)) return undefined
    teamIds.push(teamId)
  }
  const affiliateTeamIds = []
  const affiliateCount = readU8(gameDb, teamIdsEnd)
  if (affiliateCount > 0 && affiliateCount <= AFFILIATE_COUNT_MAX) {
    const start = teamIdsEnd + 1
    if (start + affiliateCount * 4 <= recordEnd) {
      for (let i = 0; i < affiliateCount; i++) {
        const teamId = readU32(gameDb, start + i * 4)
        if (teamId !== undefined && inRange(teamId, TEAM_ID_RANGE)) affiliateTeamIds.push(teamId)
      }
    }
  }
  return { teamIds, affiliateTeamIds, gender: readU8(gameDb, teamCountAt - 1) }
}

function findInitialClub(gameDb, records) {
  const firstStart = records[0].recordStart
  for (
    let hit = findBytes(gameDb, ANCHOR, Math.max(0, firstStart - 4096), firstStart);
    hit !== -1;
    hit = findBytes(gameDb, ANCHOR, hit + 1, firstStart)
  ) {
    const start = hit - 5 - 12
    if (start < 0 || readU8(gameDb, start + ZERO_BYTE_OFFSET) !== 0) continue
    const nationId = readU32(gameDb, start + NATION_OFFSET)
    const faNationId = readU32(gameDb, start + FA_NATION_OFFSET)
    if (
      nationId === undefined ||
      faNationId === undefined ||
      nationId > NATION_MAX ||
      faNationId > NATION_MAX
    )
      continue
    const names = readClubNames(gameDb, start)
    if (!names) continue
    return {
      recordStart: start,
      recordEnd: firstStart,
      teamCountsStart: undefined,
      clubIndex: 0,
      uid: 0,
      uniqueId: undefined,
      nationId,
      faNationId,
      name: names[0],
      shortName: names[1],
    }
  }
  return undefined
}

function scanClubRecords(gameDb) {
  const records = []
  let lastAcceptedStart = 0
  for (
    let hit = findBytes(gameDb, ANCHOR, ANCHOR_OFFSET);
    hit !== -1;
    hit = findBytes(gameDb, ANCHOR, hit + 1)
  ) {
    const recordStart = hit - ANCHOR_OFFSET
    if (records.length && recordStart - lastAcceptedStart >= STOP_GAP_BYTES) break
    if (readU8(gameDb, recordStart + ZERO_BYTE_OFFSET) !== 0) continue
    const record = acceptClubRecord(gameDb, recordStart)
    if (record) {
      records.push(record)
      lastAcceptedStart = recordStart
    }
  }
  const scanEnd = records.length
    ? Math.min(lastAcceptedStart + STOP_GAP_BYTES, gameDb.length)
    : gameDb.length
  return [records, scanEnd]
}

function acceptClubRecord(gameDb, recordStart) {
  const nations = readNations(gameDb, recordStart)
  if (!nations) return undefined
  const uid = readU32(gameDb, recordStart + UID_OFFSET)
  if (uid === undefined || readU32(gameDb, recordStart + UID_COPY_OFFSET) !== uid) return undefined
  const clubIndex = readU32(gameDb, recordStart + CLUB_INDEX_OFFSET)
  const names = readClubNames(gameDb, recordStart)
  if (clubIndex === undefined || !names) return undefined
  return {
    recordStart,
    recordEnd: gameDb.length,
    teamCountsStart: undefined,
    clubIndex: clubIndex + 1,
    uid: uid + 1,
    uniqueId: undefined,
    nationId: nations[0],
    faNationId: nations[1],
    name: names[0],
    shortName: names[1],
  }
}

function readClubNames(gameDb, recordStart) {
  const longLengthAt = recordStart + LONG_NAME_OFFSET
  const longLength = readU32(gameDb, longLengthAt)
  if (longLength === undefined || longLength < 1 || longLength > MAX_NAME_BYTES) return undefined
  const longStart = longLengthAt + 4
  const shortLengthAt = longStart + longLength
  const shortLength = readU32(gameDb, shortLengthAt)
  if (shortLength === undefined || shortLength < 1 || shortLength > MAX_NAME_BYTES) return undefined
  const shortStart = shortLengthAt + 4
  const name = decodeUtf8(gameDb, longStart, shortLengthAt)
  const shortName = decodeUtf8(gameDb, shortStart, shortStart + shortLength)
  return name === undefined || shortName === undefined ? undefined : [name, shortName]
}
