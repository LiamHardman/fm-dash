// Save-local nation ids and the editor-database ids used by the bundled name table.
// Translated from fm27-scout-code (MIT, Copyright (c) 2026 FM27 Scout contributors),
// crates/fm27-core/src/nations.rs. See ./THIRD_PARTY.md.
//
// A save stores no nation names. Nation fields in player and club records hold an index
// into the save's nation table, whose closing header pairs that index with the editor
// database id. The bundled table maps the database id to an English name.

import { findBytes, parseError, readU8, readU16, readU32 } from './bytes.js'
import { NATION_NAMES } from './nation-names.js'

const NATION_ROW_COUNT = 251
const NATION_TABLE_HEAD = new Uint8Array([0xfb, 0x00, 0x00, 0x00, 0x00])
const MAX_STRING_BYTES = 1024
const MAX_ENTRIES = 100_000

export function nationName(databaseId) {
  return NATION_NAMES[databaseId]
}

export class NationIndex {
  constructor(nations = []) {
    this.nations = nations
    this.byId = new Map(nations.map((nation) => [nation.nationId, nation]))
  }

  static read(gameDb) {
    const candidates = []
    let head = findBytes(gameDb, NATION_TABLE_HEAD, 0)
    while (head !== -1) {
      const nations = walkTable(gameDb, head)
      if (nations) candidates.push(nations)
      head = findBytes(gameDb, NATION_TABLE_HEAD, head + 1)
    }
    if (candidates.length === 0) throw parseError('the nation table was not found')
    // Rust: stable sort by length, then pop() → the last of the longest candidates.
    let best = candidates[0]
    for (const candidate of candidates) if (candidate.length >= best.length) best = candidate
    return new NationIndex(best)
  }

  nameFor(nationId) {
    return this.byId.get(nationId)?.name
  }

  databaseIdFor(nationId) {
    return this.byId.get(nationId)?.databaseId
  }
}

function walkTable(gameDb, head) {
  if (head + 5 > gameDb.length) return undefined
  if (readU32(gameDb, head) !== NATION_ROW_COUNT) return undefined
  let cursor = head + 5
  const nations = []
  for (let index = 0; index < NATION_ROW_COUNT; index++) {
    const recordStart = cursor
    if (readU8(gameDb, cursor) !== 0) return undefined
    cursor = walkRecord(gameDb, cursor + 1)
    if (cursor === undefined) return undefined
    const storedIndex = readU32(gameDb, cursor)
    const databaseId = readU32(gameDb, cursor + 4)
    const databaseCopy = readU32(gameDb, cursor + 8)
    if (databaseCopy === undefined) return undefined
    if (storedIndex !== index || databaseId !== databaseCopy || databaseId === 0) return undefined
    cursor += 12
    if (recordStart === cursor) return undefined
    nations.push({ nationId: index, databaseId, name: nationName(databaseId) })
  }
  return nations
}

function walkRecord(gameDb, start) {
  let cursor = start + 4
  cursor = skipCount(gameDb, cursor, 2, 4)
  if (cursor === undefined) return undefined
  cursor = skipCount(gameDb, cursor, 4, 4)
  if (cursor === undefined) return undefined
  const count = readU16(gameDb, cursor)
  if (count === undefined || count > MAX_ENTRIES) return undefined
  cursor += 2
  for (let i = 0; i < count; i++) {
    cursor += 1
    for (let s = 0; s < 4; s++) {
      const length = readU32(gameDb, cursor)
      if (length === undefined || length > MAX_STRING_BYTES) return undefined
      cursor += 4 + length
      if (cursor > gameDb.length) return undefined
    }
    cursor += 20
  }
  cursor += 48
  for (const [countBytes, entryBytes] of [
    [4, 25],
    [4, 12],
    [4, 14],
  ]) {
    cursor = skipCount(gameDb, cursor, countBytes, entryBytes)
    if (cursor === undefined) return undefined
  }
  cursor += 24
  for (const [countBytes, entryBytes] of [
    [1, 5],
    [4, 28],
    [4, 28],
    [4, 12],
  ]) {
    cursor = skipCount(gameDb, cursor, countBytes, entryBytes)
    if (cursor === undefined) return undefined
  }
  cursor += 54
  return cursor <= gameDb.length ? cursor : undefined
}

function skipCount(gameDb, cursor, countBytes, entryBytes) {
  const count =
    countBytes === 1
      ? readU8(gameDb, cursor)
      : countBytes === 2
        ? readU16(gameDb, cursor)
        : readU32(gameDb, cursor)
  if (count === undefined || count > MAX_ENTRIES) return undefined
  const end = cursor + countBytes + count * entryBytes
  return end <= gameDb.length ? end : undefined
}
