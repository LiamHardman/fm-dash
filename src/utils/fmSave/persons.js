// Decoding a player's person block: names, birth date, nations and personality.
// The block is found by its eight-byte personality run.
// Translated from fm27-scout-code (MIT, Copyright (c) 2026 FM27 Scout contributors),
// crates/fm27-core/src/persons.rs. Traits and home-grown club names are not ported
// (FM-Dash's export has no column for them). See ./THIRD_PARTY.md.

import { decodeUtf8, MISSING_U32, readU8, readU16, readU32 } from './bytes.js'
import { ageAt, decodeDate, EARLIEST_YEAR, PersonLayout as P } from './layout.js'

const SECOND_NATION_KEY = (P.SECOND_NATION_PAIR[1] << 8) | P.SECOND_NATION_PAIR[0]

export class PersonDecoder {
  /**
   * @param {import('./names.js').NamePools} names
   * @param {string} clock ISO date of the save's in-game clock
   */
  constructor(names, clock) {
    this.names = names
    this.clock = clock
    this.clockYear = Number(clock.slice(0, 4))
  }

  /** Decode the unique person block in [windowStart, windowEnd), or undefined. */
  decode(gameDb, windowStart, windowEnd) {
    const searchStart = windowStart + P.PERSONALITY_OFFSET_FROM_BIRTH - 1
    if (searchStart >= windowEnd) return undefined
    const end = Math.min(windowEnd, gameDb.length)
    let selected
    for (let zero = searchStart; zero + 1 + P.PERSONALITY_COUNT <= end; zero++) {
      if (gameDb[zero] !== 0) continue
      let personalityRun = true
      for (let i = zero + 1; i < zero + 1 + P.PERSONALITY_COUNT; i++) {
        if (gameDb[i] < 1 || gameDb[i] > 20) {
          personalityRun = false
          break
        }
      }
      if (!personalityRun) continue
      const birthOffset = zero - (P.PERSONALITY_OFFSET_FROM_BIRTH - 1)
      const person = this.tryCandidate(gameDb, birthOffset, end)
      if (person) {
        if (selected) return undefined // FM27 requires a unique candidate in the window.
        selected = person
      }
    }
    return selected
  }

  tryCandidate(gameDb, birthOffset, windowEnd) {
    const birthWord = readU32(gameDb, birthOffset)
    if (birthWord === undefined) return undefined
    const packed = birthWord & 0xffff
    const year = birthWord >>> 16
    if (packed >> 9 !== 0) return undefined
    const dayOfYear = packed & 0x1ff
    if (dayOfYear < 1 || dayOfYear > 366) return undefined
    if (year < EARLIEST_YEAR || year > this.clockYear) return undefined

    const zeroStart = birthOffset + P.DATE_ZERO_BYTES_OFFSET_FROM_BIRTH
    if (zeroStart + P.DATE_ZERO_BYTES_COUNT > gameDb.length) return undefined
    for (let i = zeroStart; i < zeroStart + P.DATE_ZERO_BYTES_COUNT; i++) {
      if (gameDb[i] !== 0) return undefined
    }

    const legalNameLength = legalNameLengthAt(gameDb, birthOffset)
    if (legalNameLength === undefined) return undefined
    const blockStart = birthOffset - (P.LEGAL_NAME_OFFSET_FROM_BLOCK_START + legalNameLength)
    if (blockStart - 8 < 0) return undefined
    const firstNameId = readU32(gameDb, blockStart)
    const surnameId = readU32(gameDb, blockStart + 5)
    const commonNameId = readU32(gameDb, blockStart + 10)
    if (
      readU8(gameDb, blockStart + 4) !== 0 ||
      readU8(gameDb, blockStart + 9) !== 0 ||
      readU8(gameDb, blockStart + 14) !== 0
    ) {
      return undefined
    }
    for (const id of [firstNameId, surnameId, commonNameId]) {
      if (id === undefined) return undefined
      if (id !== MISSING_U32 && id >= P.NAME_ID_LIMIT) return undefined
    }

    const nationId = readU16(gameDb, birthOffset + P.NATION_ID_OFFSET_FROM_BIRTH)
    const personalityAt = birthOffset + P.PERSONALITY_OFFSET_FROM_BIRTH
    if (nationId === undefined || personalityAt + P.PERSONALITY_COUNT > gameDb.length) {
      return undefined
    }
    const personality = Array.from(
      gameDb.subarray(personalityAt, personalityAt + P.PERSONALITY_COUNT)
    )

    const present = readU8(gameDb, birthOffset + P.RELATION_PRESENT_OFFSET_FROM_BIRTH)
    const count = readU8(gameDb, birthOffset + P.RELATION_COUNT_OFFSET_FROM_BIRTH)
    if (present === undefined || count === undefined) return undefined
    const secondNationIds = []
    if (present !== 0 && count !== 0) {
      const entriesStart = birthOffset + P.RELATION_ENTRIES_OFFSET_FROM_BIRTH
      if (entriesStart + P.RELATION_ENTRY_BYTES * count > windowEnd) return undefined
      for (let entry = 0; entry < count; entry++) {
        const at = entriesStart + entry * P.RELATION_ENTRY_BYTES
        const referenced = readU32(gameDb, at)
        const kindRole = readU16(gameDb, at + 10)
        const qualifier = readU8(gameDb, at + 12)
        if (referenced === undefined || kindRole === undefined || qualifier === undefined) {
          return undefined
        }
        if (
          kindRole === SECOND_NATION_KEY &&
          P.SECOND_NATION_QUALIFIERS.includes(qualifier) &&
          !secondNationIds.includes(referenced)
        ) {
          secondNationIds.push(referenced)
        }
      }
    }

    const firstName = this.names.firstNames.nameAt(gameDb, firstNameId)
    const lastName = this.names.surnames.nameAt(gameDb, surnameId)
    const commonName =
      commonNameId === MISSING_U32 ? undefined : this.names.resolveCommon(gameDb, commonNameId)
    const legalName =
      legalNameLength === 0
        ? undefined
        : decodeUtf8(
            gameDb,
            blockStart + P.LEGAL_NAME_OFFSET_FROM_BLOCK_START,
            blockStart + P.LEGAL_NAME_OFFSET_FROM_BLOCK_START + legalNameLength
          )
    const fullName = firstName && lastName ? `${firstName} ${lastName}` : undefined
    const name = commonName ?? fullName ?? firstName ?? lastName ?? legalName

    const birthDate = decodeDate(gameDb, birthOffset)
    return {
      name,
      firstName,
      lastName,
      commonName,
      fullName,
      birthDate,
      age: birthDate ? ageAt(birthDate, this.clock) : undefined,
      nationId,
      secondNationIds,
      personality,
    }
  }
}

function legalNameLengthAt(gameDb, birthOffset) {
  if (birthOffset - 4 < 0) return undefined
  const zeroWord = readU32(gameDb, birthOffset - 4)
  if (zeroWord === undefined) return undefined
  if (zeroWord === 0) return 0
  for (let length = P.LEGAL_NAME_LENGTH_MIN; length <= P.LEGAL_NAME_LENGTH_MAX; length++) {
    const at = birthOffset - 4 - length
    if (at < 0) return undefined
    if (readU32(gameDb, at) === length) return length
  }
  return undefined
}
