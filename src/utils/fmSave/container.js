// Read-only access to the FM27 save container.
// Translated from fm27-scout-code (MIT, Copyright (c) 2026 FM27 Scout contributors), crates/fm27-core/src/container.rs
// (read path only; the rewrite/editor half is not ported). See ./THIRD_PARTY.md.
//
// A save is a 26-byte header, a run of zstd frames, and a trailer that holds a
// compressed directory listing the named frames. Offsets in the directory are
// relative to the end of the header (FRAME_BASE).
//
// The container reads through a `source` = { size: number, read(offset, length): Promise<Uint8Array> },
// so a browser File can be sliced without loading the whole save.

import { decompress } from 'fzstd'
import { corrupt, decodeUtf8, readU32, readU64, SaveParseError } from './bytes.js'

const FILE_MAGIC = [0x02, 0x01, 0x66, 0x6d, 0x66, 0x2e] // "\x02\x01fmf."
export const HEADER_SIZE = 26
const TRAILER_POINTER_AT = 9
const TRAILER_HEADER_SIZE = 9
const FRAME_BASE = 26
export const DEFAULT_FRAME_CAP = 1024 * 1024 * 1024
const DEFAULT_TOTAL_CAP = 4 * 1024 * 1024 * 1024
const MAX_TRAILER_COMPRESSED = 64 * 1024 * 1024
const MAX_TRAILER_DECOMPRESSED = 64 * 1024 * 1024
const MAX_SAVE_NAME_BYTES = 4096
const MAX_ENTRY_NAME_BYTES = 256
const MAX_DIRECTORY_ENTRIES = 100_000
const ENTRY_FIXED_TAIL_BYTES = 40

const startsWithMagic = (bytes) => FILE_MAGIC.every((byte, i) => bytes[i] === byte)
const isAlnum = (b) =>
  (b >= 0x30 && b <= 0x39) || (b >= 0x41 && b <= 0x5a) || (b >= 0x61 && b <= 0x7a)
const isNameByte = (b) => isAlnum(b) || b === 0x5f

export function isSection(entry) {
  return entry.extension === '.dat' || entry.extension === '.cmt'
}

/** Wrap a Uint8Array (whole file in memory) as a container source. */
export function bufferSource(bytes) {
  return {
    size: bytes.length,
    read: async (offset, length) => bytes.subarray(offset, offset + length),
  }
}

/** Wrap a browser File/Blob as a container source. */
export function blobSource(blob) {
  return {
    size: blob.size,
    read: async (offset, length) =>
      new Uint8Array(await blob.slice(offset, offset + length).arrayBuffer()),
  }
}

export class SaveFile {
  constructor({ source, saveName, trailerOffset, directory, sections }) {
    this.source = source
    this.saveName = saveName
    this.trailerOffset = trailerOffset
    this.directory = directory
    this.sections = sections // Map<name, entry>
  }

  static async open(source) {
    const fileSize = source.size
    if (fileSize < HEADER_SIZE) {
      throw corrupt('the file is too short to be an FM save')
    }
    const header = await source.read(0, HEADER_SIZE)
    if (!startsWithMagic(header)) {
      throw new SaveParseError('not_fm_save', 'this file is not a Football Manager save')
    }

    const trailerRelative = readU64(header, 9)
    const trailerOffset = TRAILER_POINTER_AT + trailerRelative
    if (trailerOffset < HEADER_SIZE || trailerOffset + TRAILER_HEADER_SIZE >= fileSize) {
      throw corrupt('the save directory is missing')
    }
    const trailerLen = fileSize - trailerOffset
    if (trailerLen - TRAILER_HEADER_SIZE > MAX_TRAILER_COMPRESSED) {
      throw corrupt('the save directory is implausibly large')
    }
    const trailer = await source.read(trailerOffset, trailerLen)
    if (!startsWithMagic(trailer)) {
      throw corrupt('the save directory is damaged')
    }

    const payload = decompressFrame(
      trailer.subarray(TRAILER_HEADER_SIZE),
      undefined,
      MAX_TRAILER_DECOMPRESSED,
      'the save directory'
    )
    const directory = parseDirectory(payload)
    validateAllEntries(directory, trailerOffset)

    const sections = new Map()
    for (const entry of directory.topEntries) {
      if (!isSection(entry)) continue
      if (sections.has(entry.name)) {
        throw corrupt(`the save directory lists section "${entry.name}" twice`)
      }
      sections.set(entry.name, entry)
    }
    return new SaveFile({
      source,
      saveName: directory.saveName,
      trailerOffset,
      directory,
      sections,
    })
  }

  sectionEntry(name) {
    const entry = this.sections.get(name)
    if (!entry) throw corrupt(`save has no "${name}" section`)
    return entry
  }

  async readSection(name) {
    const entry = this.sectionEntry(name)
    if (entry.decompressedSize > DEFAULT_FRAME_CAP) {
      throw corrupt(`section "${name}" is larger than the per-frame cap`)
    }
    const compressed = await this.source.read(entry.frameOffset, entry.compressedSize)
    return decompressFrame(
      compressed,
      entry.decompressedSize,
      DEFAULT_FRAME_CAP,
      `section "${name}"`
    )
  }

  totalDeclaredBytes() {
    return this.directory.topEntries.reduce((sum, entry) => sum + entry.decompressedSize, 0)
  }

  sectionNames() {
    return [...this.sections.keys()].sort()
  }
}

function readLengthPrefixedString(bytes, offset, maxLen) {
  const len = readU32(bytes, offset)
  if (len === undefined || len > maxLen) return undefined
  const start = offset + 4
  const text = decodeUtf8(bytes, start, start + len)
  return text === undefined ? undefined : [text, start + len]
}

function parseDirectory(payload) {
  const named = readLengthPrefixedString(payload, 0, MAX_SAVE_NAME_BYTES)
  if (!named) throw corrupt('the save directory name is unreadable')
  let [saveName, cursor] = named
  const reserved = readU32(payload, cursor)
  if (reserved === undefined) throw corrupt('the save directory reserved word is unreadable')
  cursor += 4

  const topEntries = []
  for (;;) {
    const parsed = parseDirectoryEntry(payload, cursor)
    if (!parsed) break
    topEntries.push(parsed[0])
    cursor = parsed[1]
    if (topEntries.length > MAX_DIRECTORY_ENTRIES) {
      throw corrupt('the save directory has too many entries')
    }
  }

  const groups = []
  while (cursor < payload.length) {
    if (payload.subarray(cursor).every((byte) => byte === 0)) break
    const [group, next] = parseDirectoryGroup(payload, cursor)
    groups.push(group)
    cursor = next
  }

  const anySection =
    topEntries.some(isSection) || groups.some((group) => group.entries.some(isSection))
  if (!anySection) throw corrupt('the save directory lists no sections')
  topEntries.sort((a, b) => a.frameOffset - b.frameOffset)
  return { saveName, reserved, topEntries, groups }
}

function parseDirectoryGroup(payload, start) {
  const version = readU32(payload, start)
  if (version === undefined) throw corrupt('a directory group header is truncated')
  let cursor = start + 4
  const nameLen = readU32(payload, cursor)
  if (nameLen === undefined) throw corrupt('a directory group name is truncated')
  if (nameLen < 1 || nameLen > MAX_ENTRY_NAME_BYTES) {
    throw corrupt('a directory group name has an invalid length')
  }
  const nameEnd = cursor + 4 + nameLen
  if (nameEnd > payload.length) throw corrupt('a directory group name is truncated')
  const nameBytes = payload.subarray(cursor + 4, nameEnd)
  if (!nameBytes.every(isNameByte)) throw corrupt('a directory group name is not valid')
  const name = decodeUtf8(payload, cursor + 4, nameEnd)
  cursor = nameEnd

  const count = readU32(payload, cursor)
  if (count === undefined) throw corrupt('a directory group entry count is truncated')
  cursor += 4
  if (count > MAX_DIRECTORY_ENTRIES) throw corrupt('a directory group has too many entries')

  const entries = []
  for (let index = 0; index < count; index++) {
    const parsed = parseDirectoryEntry(payload, cursor)
    if (!parsed) throw corrupt(`directory group "${name}" entry ${index} is unreadable`)
    entries.push(parsed[0])
    cursor = parsed[1]
  }
  return [{ version, name, entries }, cursor]
}

function parseDirectoryEntry(payload, cursor) {
  const nameLen = readU32(payload, cursor)
  if (nameLen === undefined || nameLen < 1 || nameLen > MAX_ENTRY_NAME_BYTES) return undefined
  const extensionLenAt = cursor + 4 + nameLen
  const tailAt = extensionLenAt + 8
  if (tailAt + ENTRY_FIXED_TAIL_BYTES > payload.length) return undefined
  if (readU32(payload, extensionLenAt) !== 4) return undefined
  const nameBytes = payload.subarray(cursor + 4, extensionLenAt)
  if (!nameBytes.every(isNameByte)) return undefined
  const ext = payload.subarray(extensionLenAt + 4, tailAt)
  if (ext.length !== 4 || ext[0] !== 0x2e || !ext.subarray(1).every(isAlnum)) return undefined
  return [
    {
      name: decodeUtf8(payload, cursor + 4, extensionLenAt),
      extension: decodeUtf8(payload, extensionLenAt + 4, tailAt),
      frameOffset: FRAME_BASE + readU64(payload, tailAt),
      compressedSize: readU64(payload, tailAt + 8),
      decompressedSize: readU64(payload, tailAt + 16),
      trailingValues: [readU64(payload, tailAt + 24), readU64(payload, tailAt + 32)],
    },
    tailAt + ENTRY_FIXED_TAIL_BYTES,
  ]
}

function validateAllEntries(directory, trailerOffset) {
  const entries = [...directory.topEntries, ...directory.groups.flatMap((group) => group.entries)]
  entries.sort((a, b) => a.frameOffset - b.frameOffset)
  let previousEnd = HEADER_SIZE
  let total = 0
  for (const entry of entries) {
    if (
      entry.compressedSize === 0 ||
      entry.frameOffset < previousEnd ||
      entry.frameOffset + entry.compressedSize > trailerOffset
    ) {
      throw corrupt(`directory entry ${entry.name}${entry.extension} points outside the frame area`)
    }
    if (entry.decompressedSize > DEFAULT_FRAME_CAP) {
      throw corrupt(
        `directory entry ${entry.name}${entry.extension} is larger than the per-frame cap`
      )
    }
    total += entry.decompressedSize
    previousEnd = entry.frameOffset + entry.compressedSize
  }
  if (total > DEFAULT_TOTAL_CAP) {
    throw corrupt('the save declares more data than the total decompressed cap')
  }
}

function decompressFrame(compressed, expectedSize, cap, what) {
  const limit = expectedSize === undefined ? cap : Math.min(expectedSize, cap)
  let decoded
  try {
    // Preallocating the output when the size is known avoids fzstd's grow-and-copy.
    decoded =
      expectedSize === undefined
        ? decompress(compressed)
        : decompress(compressed, new Uint8Array(expectedSize))
  } catch (error) {
    throw corrupt(`${what} could not be decompressed: ${error?.message ?? error}`)
  }
  if (decoded.length > limit) {
    throw corrupt(`${what} decompresses to more than ${limit} bytes`)
  }
  if (expectedSize !== undefined && decoded.length !== expectedSize) {
    throw corrupt(`${what} decompressed to ${decoded.length} bytes, expected ${expectedSize}`)
  }
  return decoded
}
