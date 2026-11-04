// The three name pools stored in `game_db`: first names, surnames and common names.
// Translated from fm27-scout-code (MIT, Copyright (c) 2026 FM27 Scout contributors),
// crates/fm27-core/src/names.rs. See ./THIRD_PARTY.md.

import { decodeUtf8, findBytes, parseError, readU32 } from './bytes.js'
import { NAME_POOL_PREFIX, NAME_POOL_SIGNATURE } from './layout.js'

const MAX_NAME_BYTES = 1024
const MAX_SIGNATURE_SCALE = 1024

export class PoolIndex {
  /** @param {Uint32Array} offsets offset of each entry's u32 length word */
  constructor(offsets) {
    this.offsets = offsets
    this.count = offsets.length
  }

  nameAt(bytes, index) {
    if (index === undefined || index === 0xffffffff || index >= this.count) return undefined
    const offset = this.offsets[index]
    const length = readU32(bytes, offset)
    if (!length) return undefined
    return decodeUtf8(bytes, offset + 4, offset + 4 + length)
  }
}

export class NamePools {
  constructor(firstNames, surnames, commonNames, endOffset) {
    this.firstNames = firstNames
    this.surnames = surnames
    this.commonNames = commonNames
    this.endOffset = endOffset
  }

  static locate(gameDb) {
    let found
    let firstError
    for (const at of signatureCandidates(gameDb)) {
      try {
        const pools = NamePools.readAfterSignature(gameDb, at + NAME_POOL_SIGNATURE.length)
        if (found) throw parseError('name pools are ambiguous: more than one valid signature')
        found = pools
      } catch (error) {
        if (error.message.startsWith('name pools are ambiguous')) throw error
        firstError ??= error
      }
    }
    if (found) return found
    throw firstError ?? parseError('name pools were not found')
  }

  static readAfterSignature(gameDb, start) {
    const cursor = { at: start }
    const firstNames = readPool(gameDb, cursor, 'first names')
    const surnames = readPool(gameDb, cursor, 'surnames')
    const commonNames = readPool(gameDb, cursor, 'common names')
    return new NamePools(firstNames, surnames, commonNames, cursor.at)
  }

  resolveCommon(bytes, id) {
    return (
      this.commonNames.nameAt(bytes, id) ??
      this.firstNames.nameAt(bytes, id) ??
      this.surnames.nameAt(bytes, id)
    )
  }
}

function signatureCandidates(gameDb) {
  const candidates = []
  let at = findBytes(gameDb, NAME_POOL_PREFIX, 0)
  while (at !== -1) {
    if (at + NAME_POOL_SIGNATURE.length <= gameDb.length && signatureMatches(gameDb, at)) {
      candidates.push(at)
    }
    at = findBytes(gameDb, NAME_POOL_PREFIX, at + 1)
  }
  return candidates
}

function signatureMatches(bytes, at) {
  let exact = true
  for (let i = 0; i < NAME_POOL_SIGNATURE.length; i++) {
    if (bytes[at + i] !== NAME_POOL_SIGNATURE[i]) {
      exact = false
      break
    }
  }
  if (exact) return true
  const words = Array.from({ length: 6 }, (_, i) => readU32(bytes, at + i * 4))
  const expected = [46421, 0, 1024, 256, 2048, 0]
  if (words[0] !== expected[0] || words[5] !== expected[5]) return false
  if (words[3] === 0 || words[3] % expected[3] !== 0) return false
  const scale = words[3] / expected[3]
  if (scale === 0 || scale > MAX_SIGNATURE_SCALE) return false
  return (
    words[2] === expected[2] * scale &&
    words[4] === expected[4] * scale &&
    (words[1] === 0 || (words[1] & 0xffffff00) === 0)
  )
}

function readPool(gameDb, cursor, label) {
  const count = readU32(gameDb, cursor.at)
  if (count === undefined) throw parseError(`${label} pool count is outside the section`)
  let position = cursor.at + 4
  if (position > gameDb.length) throw parseError(`${label} pool runs past the section`)
  if (count * 8 > gameDb.length - position) {
    throw parseError(`${label} pool claims ${count} entries, more than the section can hold`)
  }
  const offsets = new Uint32Array(count)
  for (let index = 0; index < count; index++) {
    const entryId = readU32(gameDb, position)
    const nameLength = readU32(gameDb, position + 4)
    if (entryId === undefined || nameLength === undefined) {
      throw parseError(`${label} pool runs past the section`)
    }
    if (entryId !== index) throw parseError(`${label} pool entry ${index} has id ${entryId}`)
    if (nameLength > MAX_NAME_BYTES) {
      throw parseError(
        `${label} pool entry ${index} is ${nameLength} bytes, over the ${MAX_NAME_BYTES} cap`
      )
    }
    offsets[index] = position + 4
    position += 8 + nameLength
    if (position > gameDb.length) throw parseError(`${label} pool runs past the section`)
  }
  cursor.at = position
  return new PoolIndex(offsets)
}
