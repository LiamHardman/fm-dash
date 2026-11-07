// Current injuries from the small `injury_manager` section.
// The walk is translated from fm27-scout-code (MIT, Copyright (c) 2026 FM27 Scout
// contributors), crates/fm27-core/src/injuries.rs; only the current-injury windows are
// read (no typed rows, no history log). See ./THIRD_PARTY.md.
//
// The section is an array-of-arrays store: four counted arrays (strides 14, 14, 13, 15),
// three counted lists, then a fixed 8-byte tail. Arrays 0 and 1 hold the current injury
// windows: byte 0, a recovery date @1, the expected return the game shows @5, and the
// person selector (pindex + 1) @9. A count that doesn't fit rejects the whole section.

import { parseError, readU32 } from './bytes.js'
import { decodeDate } from './layout.js'

const ARRAYS_OFFSET = 12
const ARRAY_STRIDES = [14, 14, 13, 15]
const WINDOW_ARRAYS = [0, 1]
const LIST_ENTRY_BYTES = [4, 4, 5]
const TAIL = [1, 0, 0, 0, 0, 0, 0xff, 0xff]
const RETURN_DATE_OFFSET = 5
const SELECTOR_OFFSET = 9

/**
 * The pindexes of players injured on the game date (expected return on or after it).
 * @param {Uint8Array} section the decompressed injury_manager section
 * @param {string} clock ISO game date
 * @returns {Set<number>}
 */
export function readInjuredPindexes(section, clock) {
  let cursor = ARRAYS_OFFSET
  const arrays = []
  for (const stride of ARRAY_STRIDES) {
    const count = readU32(section, cursor)
    if (count === undefined) throw parseError('injury_manager is truncated')
    cursor += 4
    if (count > Math.floor((section.length - cursor) / stride)) {
      throw parseError('an injury_manager array runs past the section')
    }
    arrays.push({ start: cursor, count, stride })
    cursor += count * stride
  }
  for (const entryBytes of LIST_ENTRY_BYTES) {
    const count = readU32(section, cursor)
    if (count === undefined) throw parseError('injury_manager is truncated')
    cursor += 4
    if (count > Math.floor((section.length - cursor) / entryBytes)) {
      throw parseError('an injury_manager list runs past the section')
    }
    cursor += count * entryBytes
  }
  if (cursor + TAIL.length !== section.length || TAIL.some((b, i) => section[cursor + i] !== b)) {
    throw parseError('injury_manager does not end with its tail')
  }

  const injured = new Set()
  for (const index of WINDOW_ARRAYS) {
    const { start, count, stride } = arrays[index]
    for (let row = 0; row < count; row++) {
      const at = start + row * stride
      const expectedReturn = decodeDate(section, at + RETURN_DATE_OFFSET)
      const selector = readU32(section, at + SELECTOR_OFFSET)
      if (!expectedReturn || !selector || expectedReturn < clock) continue
      injured.add(selector - 1)
    }
  }
  return injured
}
