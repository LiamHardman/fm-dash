// Test-only builder for synthetic FM27 save containers.
// Mirrors fm27-scout-code's `build_save` test helper (container.rs, MIT). See ../THIRD_PARTY.md.
// Node-only: uses zlib's zstd compressor (Node >= 22.15).

import { zstdCompressSync } from 'node:zlib'

const MAGIC = [0x02, 0x01, 0x66, 0x6d, 0x66, 0x2e]
const HEADER_SIZE = 26
const TRAILER_POINTER_AT = 9

const u32 = (value) => {
  const out = new Uint8Array(4)
  new DataView(out.buffer).setUint32(0, value, true)
  return out
}
const u64 = (value) => {
  const out = new Uint8Array(8)
  new DataView(out.buffer).setBigUint64(0, BigInt(value), true)
  return out
}
const text = (value) => new TextEncoder().encode(value)
const concat = (parts) => {
  const out = new Uint8Array(parts.reduce((n, p) => n + p.length, 0))
  let at = 0
  for (const part of parts) {
    out.set(part, at)
    at += part.length
  }
  return out
}
const lengthPrefixed = (value) => concat([u32(text(value).length), text(value)])

function encodeEntry(entry) {
  return concat([
    lengthPrefixed(entry.name),
    lengthPrefixed(entry.extension),
    u64(entry.frameOffset - HEADER_SIZE),
    u64(entry.compressedSize),
    u64(entry.decompressedSize),
    u64(0),
    u64(0),
  ])
}

/**
 * @param {Array<[string, Uint8Array, Uint8Array[]?]>} top top-level sections; an optional third
 *   element is raw bytes (e.g. zstd frames) placed after the section's frame but not listed
 * @param {Array<[string, Uint8Array]>} group sections in the `rgman` group
 * @returns {Uint8Array} a complete .fm file
 */
export function buildSave(top, group = [], saveName = 'test save') {
  const frames = []
  let frameBytes = 0
  const append = ([name, data, unlisted = []], entries) => {
    const compressed = new Uint8Array(zstdCompressSync(data))
    entries.push({
      name,
      extension: '.dat',
      frameOffset: HEADER_SIZE + frameBytes,
      compressedSize: compressed.length,
      decompressedSize: data.length,
    })
    frames.push(compressed)
    frameBytes += compressed.length
    for (const raw of unlisted) {
      frames.push(raw)
      frameBytes += raw.length
    }
  }
  const topEntries = []
  const groupEntries = []
  for (const section of top) append(section, topEntries)
  for (const section of group) append(section, groupEntries)

  const directory = concat([
    lengthPrefixed(saveName),
    u32(top.length),
    ...topEntries.map(encodeEntry),
    u32(1),
    lengthPrefixed('rgman'),
    u32(groupEntries.length),
    ...groupEntries.map(encodeEntry),
    new Uint8Array(4),
  ])

  const header = new Uint8Array(HEADER_SIZE)
  header.set(MAGIC, 0)
  header.set([0x08, 0, 0], 6)
  header.set(u32(17), 17)
  header[25] = 3
  const trailerOffset = HEADER_SIZE + frameBytes
  header.set(u64(trailerOffset - TRAILER_POINTER_AT), 9)

  return concat([
    header,
    ...frames,
    new Uint8Array([...MAGIC, 0x08, 0, 0]),
    new Uint8Array(zstdCompressSync(directory)),
  ])
}
