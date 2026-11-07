// Locating player records in `game_db`.
// Translated from fm27-scout-code (MIT, Copyright (c) 2026 FM27 Scout contributors),
// crates/fm27-core/src/player_scan.rs. See ./THIRD_PARTY.md.
//
// Two scans are merged: a fast scan for the `01 00 6c 07` marker, and a completeness
// scan for the 15 rating bytes followed by 54 attribute bytes.

import { findBytes, MISSING_U32, parseError, readI16, readU8, readU16, readU32 } from './bytes.js'
import { PlayerLayout as L } from './layout.js'

const MARKER = new Uint8Array(L.MARKER)

/**
 * @returns {{ records: Array<{offset:number,pindex:number,uid:number}>, markerless: number,
 *             windowEnd(position:number, length:number): number }}
 */
export function locatePlayerRecords(gameDb, namePoolsEnd) {
  const markerRecords = scanMarkers(gameDb)
  const markerOffsets = new Set(markerRecords.map((record) => record.offset))
  const completeness = scanCompleteness(gameDb, namePoolsEnd, markerOffsets)

  const merged = markerRecords.concat(completeness).sort((a, b) => a.offset - b.offset)
  const records = []
  for (const record of merged) {
    if (records.length && records[records.length - 1].offset === record.offset) continue
    records.push(record)
  }
  if (records.length === 0) throw parseError('no player records found')

  const pindexes = new Set()
  const uids = new Set()
  for (const record of records) {
    if (pindexes.has(record.pindex)) {
      throw parseError(`player pindex ${record.pindex} appears in two records`)
    }
    if (uids.has(record.uid)) throw parseError(`player uid ${record.uid} appears in two records`)
    pindexes.add(record.pindex)
    uids.add(record.uid)
  }

  return {
    records,
    markerless: completeness.length,
    windowEnd: (position, length) => records[position + 1]?.offset ?? length,
  }
}

function scanMarkers(gameDb) {
  const records = []
  let markerAt = findBytes(gameDb, MARKER, 0)
  while (markerAt !== -1) {
    const recordOffset = markerAt - L.MARKER_OFFSET
    if (recordOffset >= 0) {
      const accepted = acceptCandidate(gameDb, recordOffset, true)
      if (accepted) {
        if (recordOffset + L.DECODE_EXTENT > gameDb.length) {
          throw parseError('a player record runs past game_db')
        }
        records.push({ offset: recordOffset, pindex: accepted[0], uid: accepted[1] })
      }
    }
    markerAt = findBytes(gameDb, MARKER, markerAt + 1)
  }
  return records
}

// O(n): count consecutive bytes in 1..20; every true record's 15 rating bytes trigger a
// check of the 54 attribute bytes that follow.
function scanCompleteness(gameDb, namePoolsEnd, markerOffsets) {
  const records = []
  const length = gameDb.length
  let run = 0
  for (let index = Math.min(namePoolsEnd, length); index < length; index++) {
    const byte = gameDb[index]
    if (byte >= 1 && byte <= 20) {
      run++
    } else {
      run = 0
      continue
    }
    if (run < L.RATINGS_COUNT) continue
    const ratingsStart = index + 1 - L.RATINGS_COUNT
    const recordOffset = ratingsStart - L.RATINGS_OFFSET
    if (recordOffset < 0 || markerOffsets.has(recordOffset)) continue
    const attrsStart = ratingsStart + L.RATINGS_COUNT
    const attrsEnd = attrsStart + L.ATTRIBUTE_COUNT
    if (attrsEnd > length) break
    let allInRange = true
    for (let i = attrsStart; i < attrsEnd; i++) {
      const value = gameDb[i]
      if (value < 1 || value > 125) {
        allInRange = false
        break
      }
    }
    if (!allInRange) continue
    const accepted = acceptCandidate(gameDb, recordOffset, false)
    if (accepted) records.push({ offset: recordOffset, pindex: accepted[0], uid: accepted[1] })
  }
  return records
}

function acceptCandidate(gameDb, recordOffset, checkPattern) {
  const pindexAt = recordOffset + L.PINDEX_OFFSET
  if (pindexAt < 0) return undefined
  const attributesStart = recordOffset + L.ATTRIBUTES_OFFSET
  const attributesEnd = attributesStart + L.ATTRIBUTE_COUNT
  if (attributesEnd > gameDb.length) return undefined

  const uid = readU32(gameDb, recordOffset + L.UID_OFFSET)
  const uidCopy = readU32(gameDb, recordOffset + L.UID_COPY_OFFSET)
  if (!uid || uid === MISSING_U32 || !uidCopy || uidCopy === MISSING_U32) return undefined
  const kind = readU8(gameDb, recordOffset + L.OBJECT_KIND_OFFSET)
  const kindOk = kind === 2 || kind === 3
  if (uid !== uidCopy && (uid < L.ALLOCATED_UID_MINIMUM || !kindOk)) return undefined

  const ca = readU16(gameDb, recordOffset + L.CA_OFFSET)
  const pa = readI16(gameDb, recordOffset + L.PA_OFFSET)
  const bucket = readU8(gameDb, recordOffset + L.REPUTATION_BUCKET_OFFSET)
  let highest = 0
  for (let i = attributesStart; i < attributesEnd; i++) if (gameDb[i] > highest) highest = gameDb[i]

  const edited = ca > 200 || pa > 200 || highest > 100
  if (edited && (uid !== uidCopy || !kindOk)) return undefined
  const abilityMax = edited ? 255 : 200
  if (ca > abilityMax || pa > abilityMax || pa < -10) return undefined
  if (bucket > 200) return undefined

  if (checkPattern) {
    const ratingsStart = recordOffset + L.RATINGS_OFFSET
    for (let i = ratingsStart; i < ratingsStart + L.RATINGS_COUNT; i++) {
      if (gameDb[i] < 1 || gameDb[i] > 20) return undefined
    }
    for (let i = attributesStart; i < attributesEnd; i++) {
      if (gameDb[i] < 1 || gameDb[i] > 125) return undefined
    }
  }

  const pindex = readU32(gameDb, pindexAt)
  return pindex === undefined ? undefined : [pindex, uid]
}

/**
 * The closing person header after a player record pairs `pindex + 1` with the
 * player's unique id (twice). The hit wins when exactly one distinct uid appears in the
 * window; repeats of that uid are fine, two different uids are ambiguous.
 * Adapted from `closing_person_header` in parse.rs, minus its uid-ordering bounds:
 * unique ids are not ordered by record uid, and the bounds rejected 7.9% of players.
 */
export function closingPersonHeader(gameDb, personId, searchStart, searchEnd) {
  if (personId === MISSING_U32 || searchStart > searchEnd || searchEnd > gameDb.length) {
    return undefined
  }
  const v = (personId + 1) >>> 0
  const needle = new Uint8Array([v & 0xff, (v >>> 8) & 0xff, (v >>> 16) & 0xff, v >>> 24])
  let found
  let from = searchStart
  for (;;) {
    const hit = findBytes(gameDb, needle, from, searchEnd)
    if (hit === -1) break
    from = hit + 1
    const uid = readU32(gameDb, hit + 4)
    const uidCopy = readU32(gameDb, hit + 8)
    if (uid === undefined || uidCopy === undefined) continue
    if (uid !== uidCopy || uid === 0 || uid === MISSING_U32) continue
    if (!found) found = [hit, uid]
    else if (found[1] !== uid) return undefined
  }
  return found
}
