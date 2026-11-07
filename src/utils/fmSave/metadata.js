// Save metadata: game/build recognition and the in-game clock.
// Translated from fm27-scout-code (MIT, Copyright (c) 2026 FM27 Scout contributors),
// crates/fm27-core/src/metadata.rs. See ./THIRD_PARTY.md.

import { decodeUtf8, MISSING_U32, readU32, SaveParseError } from './bytes.js'
import { decodeDate, GAME } from './layout.js'

/**
 * @param {import('./container.js').SaveFile} save
 * @returns {Promise<{game: string, build: string, dbVersion: string, gameDate?: string,
 *   managedClubUid?: number}>}
 */
export async function readMetadata(save) {
  const summary = await save.readSection('save_game_summary')
  const build = findVersion(summary)
  if (!build) {
    throw new SaveParseError('unsupported', 'no game version found in save_game_summary')
  }
  if (!build.startsWith('27.')) {
    throw new SaveParseError(
      'unsupported',
      `save was written by build ${build}; only Football Manager 27 saves are supported`
    )
  }

  const gameInfo = await save.readSection('game_info')
  const dbVersion = readLengthString(gameInfo, 8, 64)?.[0] ?? 'unknown'
  const tail = readSummaryTail(summary, build)
  const gameDate = readGameInfoDate(gameInfo) ?? tail?.clock
  return { game: GAME, build, dbVersion, gameDate, managedClubUid: tail?.managedClubUid }
}

/**
 * The summary after the build string: division names, the manager's name, the managed
 * club's short name, then a u32 club uid (FM-Dash's own decode, verified on one save)
 * and the clock. With several humans this is the manager who saved.
 */
export function readSummaryTail(summary, build) {
  let read = readLengthString(summary, 8, 65_536)
  if (!read) return undefined
  let offset = read[1]
  read = readLengthString(summary, offset, 32)
  if (!read || read[0] !== build) return undefined
  offset = read[1]
  const divisionCount = readU32(summary, offset)
  if (divisionCount === undefined || divisionCount > 4096) return undefined
  offset += 4
  for (let i = 0; i < divisionCount; i++) {
    read = readLengthString(summary, offset, 1024)
    if (!read) return undefined
    offset = read[1]
  }
  read = readLengthString(summary, offset + 8, 1024)
  if (!read) return undefined
  read = readLengthString(summary, read[1], 1024)
  if (!read) return undefined
  const uid = readU32(summary, read[1])
  return {
    managedClubUid: uid && uid !== MISSING_U32 ? uid : undefined,
    clock: decodeDate(summary, read[1] + 4),
  }
}

function readGameInfoDate(gameInfo) {
  const version = readLengthString(gameInfo, 8, 64)
  if (!version) return undefined
  const versionEnd = version[1]
  const countAt = versionEnd + 165
  const count = readU32(gameInfo, countAt)
  if (count === undefined || count > 256) return undefined
  const stringsStart = countAt + 4
  let cursor = stringsStart
  for (let i = 0; i < count; i++) {
    const read = readLengthString(gameInfo, cursor, 4096)
    if (!read) return undefined
    cursor = read[1]
  }
  return decodeDate(gameInfo, versionEnd + (cursor - stringsStart) + 174)
}

const VERSION_PATTERN = /^\d+\.\d+\.\d+\+\d+$/

export function findVersion(summary) {
  for (let offset = 0; offset < summary.length - 4; offset++) {
    const length = readU32(summary, offset)
    if (length === undefined || length < 5 || length > 32) continue
    const text = decodeUtf8(summary, offset + 4, offset + 4 + length)
    if (text !== undefined && VERSION_PATTERN.test(text)) return text
  }
  return undefined
}

function readLengthString(bytes, offset, maxLen) {
  const length = readU32(bytes, offset)
  if (length === undefined || length > maxLen) return undefined
  const start = offset + 4
  const text = decodeUtf8(bytes, start, start + length)
  return text === undefined ? undefined : [text, start + length]
}
