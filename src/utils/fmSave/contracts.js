// Decoding the wage and end date of a player's in-effect contract.
// Translated from fm27-scout-code (MIT, Copyright (c) 2026 FM27 Scout contributors),
// crates/fm27-core/src/contracts.rs. Only the chain records, tails and in-effect
// selection are ported; clauses, loan blocks and fallback dates are not (FM-Dash's
// export has no columns for them). See ./THIRD_PARTY.md.
//
// Dates are ISO strings, which compare correctly as strings.

import { findBytes, readU32 } from './bytes.js'
import { decodeDate } from './layout.js'

const TAG = [0x01, 0x00, 0x6c, 0x07]
const CHAIN_WINDOW_START_OFFSET = -30
const CHAIN_WINDOW_END_OFFSET = -30
const SELECTOR_OFFSET = 5
const START_OFFSET = -19
const TEAM_ID_OFFSET = 9
const WAGE_OFFSET = 17
const TAIL_BASE_OFFSET = 66
const TAIL_STEP_BYTES = 29
const TAIL_MAX_EVENT_COUNT = 200
const TAIL_SENTINEL_BYTE_OFFSET = 3
const TAIL_SENTINEL_BYTE_VALUE = 3
const TAIL_SENTINEL_WORD_OFFSET = 4
const TAIL_SENTINEL_WORD_VALUE = 4
const TAIL_END_OFFSET = 28
const TAIL_EVENT_COUNT_OFFSET = 42
const TAIL_STRUCT_END = 46
// Listing state in the tail (FM-Dash's own decode): +39 bit flags, +40 a listing code.
const TAIL_LISTING_FLAGS_OFFSET = 39
const TAIL_LISTING_CODE_OFFSET = 40

export class ContractDecoder {
  /**
   * @param {import('./clubs.js').ClubIndex} clubs
   * @param {string} clock ISO date
   */
  constructor(clubs, clock) {
    this.clubs = clubs
    this.clock = clock
  }

  /**
   * @returns {{clubUid?: number, wage?: number, start?: string, end?: string,
   *   listingFlags?: number, listingCode?: number} | undefined}
   */
  decode(gameDb, recordOffset, recordWindowEnd, isLastRecord, pindex, playerClubUid) {
    const records = this.findChainRecords(
      gameDb,
      recordOffset,
      recordWindowEnd,
      isLastRecord,
      pindex
    )
    if (!records) return undefined
    const selected = selectInEffect(records, playerClubUid, this.clock)
    if (!selected) return {}
    return {
      clubUid: selected.clubUid,
      wage: selected.wage,
      start: selected.start,
      end: selected.end,
      listingFlags: selected.listingFlags,
      listingCode: selected.listingCode,
    }
  }

  findChainRecords(gameDb, recordOffset, recordWindowEnd, isLastRecord, pindex) {
    const searchStart = Math.max(0, recordOffset + CHAIN_WINDOW_START_OFFSET)
    const searchEnd = isLastRecord
      ? recordWindowEnd
      : Math.max(searchStart, recordWindowEnd + CHAIN_WINDOW_END_OFFSET)
    if (searchStart >= searchEnd) return undefined
    const v = (pindex + 1) >>> 0
    const selector = new Uint8Array([v & 0xff, (v >>> 8) & 0xff, (v >>> 16) & 0xff, v >>> 24])
    const searchLimit = Math.min(searchEnd + SELECTOR_OFFSET, gameDb.length)
    const records = []
    let hit = findBytes(gameDb, selector, searchStart + SELECTOR_OFFSET, searchLimit)
    while (hit !== -1) {
      const tagOffset = hit - SELECTOR_OFFSET
      if (tagOffset >= searchStart && tagOffset < searchEnd) {
        if (hasTag(gameDb, tagOffset) || this.isDateMarkedRecord(gameDb, tagOffset)) {
          records.push(this.decodeChainRecord(gameDb, tagOffset))
        }
      }
      hit = findBytes(gameDb, selector, hit + 1, searchLimit)
    }
    return records.length ? records : undefined
  }

  clubForTeam(teamId) {
    const hit = this.clubs?.fieldingTeamToClub.get(teamId)
    return hit && this.clubs.nameByUid.has(hit[0]) ? hit[0] : undefined
  }

  decodeChainRecord(gameDb, tagOffset) {
    const teamId = readU32(gameDb, tagOffset + TEAM_ID_OFFSET) ?? 0
    const wage = readU32(gameDb, tagOffset + WAGE_OFFSET) ?? 0
    const startAt = tagOffset + START_OFFSET
    const start = startAt >= 0 ? decodeDate(gameDb, startAt) : undefined
    const tail = locateTail(gameDb, tagOffset)
    return {
      clubUid: this.clubForTeam(teamId),
      teamId,
      wage,
      start,
      end: tail === undefined ? undefined : decodeDate(gameDb, tail + TAIL_END_OFFSET),
      hasTerms: tail !== undefined,
      listingFlags: tail === undefined ? undefined : gameDb[tail + TAIL_LISTING_FLAGS_OFFSET],
      listingCode: tail === undefined ? undefined : gameDb[tail + TAIL_LISTING_CODE_OFFSET],
    }
  }

  isDateMarkedRecord(gameDb, tagOffset) {
    if (decodeDate(gameDb, tagOffset) === undefined) return false
    const startAt = tagOffset + START_OFFSET
    if (startAt < 0 || decodeDate(gameDb, startAt) === undefined) return false
    return this.clubForTeam(readU32(gameDb, tagOffset + TEAM_ID_OFFSET) ?? 0) !== undefined
  }
}

function hasTag(gameDb, at) {
  if (at + TAG.length > gameDb.length) return false
  for (let i = 0; i < TAG.length; i++) if (gameDb[at + i] !== TAG[i]) return false
  return true
}

function locateTail(gameDb, tagOffset) {
  const first = tagOffset - TAIL_BASE_OFFSET
  if (first < 0) return undefined
  if (tailMatches(gameDb, first, 0)) return first
  for (let eventCount = 1; eventCount <= TAIL_MAX_EVENT_COUNT; eventCount++) {
    const candidate = first - TAIL_STEP_BYTES * eventCount
    if (candidate < 0) return undefined
    if (tailMatches(gameDb, candidate, eventCount)) return candidate
  }
  return undefined
}

function tailMatches(gameDb, tail, expectedEventCount) {
  return (
    tail + TAIL_STRUCT_END <= gameDb.length &&
    gameDb[tail] === 0 &&
    gameDb[tail + 1] === 0 &&
    gameDb[tail + TAIL_SENTINEL_BYTE_OFFSET] === TAIL_SENTINEL_BYTE_VALUE &&
    readU32(gameDb, tail + TAIL_SENTINEL_WORD_OFFSET) === TAIL_SENTINEL_WORD_VALUE &&
    readU32(gameDb, tail + TAIL_EVENT_COUNT_OFFSET) === expectedEventCount
  )
}

// Compare [start, end] keys lexicographically (ISO dates compare as strings).
const keyGreater = (a, b) => a[0] > b[0] || (a[0] === b[0] && a[1] > b[1])

/** The contract in effect on `clock`; prefers the player's own club. */
export function selectInEffect(records, playerClubUid, clock) {
  let latest
  let atOwnClub
  let earliestRunning
  for (const record of records) {
    if (!record.hasTerms || !record.start || record.start > clock) continue
    const end = record.end ?? record.start
    const key = [record.start, end > record.start ? end : record.start]
    if (!latest || keyGreater(key, latest.key)) latest = { record, key }
    if (record.end && record.end < clock) continue
    if (playerClubUid !== undefined && record.clubUid === playerClubUid) {
      if (!atOwnClub || keyGreater(key, atOwnClub.key)) atOwnClub = { record, key }
    } else if (!earliestRunning || keyGreater(earliestRunning.key, key)) {
      earliestRunning = { record, key }
    }
  }
  return (atOwnClub ?? latest ?? earliestRunning)?.record
}
