// Little-endian readers over Uint8Array.
// Translated from fm27-scout-code (MIT, Copyright (c) 2026 FM27 Scout contributors), crates/fm27-core/src/layout.rs.
// See ./THIRD_PARTY.md.
//
// Every reader returns `undefined` when the read would run past the buffer, mirroring
// the Rust `Option` returns, so callers can fall back the same way the original does.

export const MISSING_U32 = 0xffffffff

export function readU8(bytes, offset) {
  return offset >= 0 && offset < bytes.length ? bytes[offset] : undefined
}

export function readU16(bytes, offset) {
  if (offset < 0 || offset + 2 > bytes.length) return undefined
  return bytes[offset] | (bytes[offset + 1] << 8)
}

export function readI16(bytes, offset) {
  const value = readU16(bytes, offset)
  return value === undefined ? undefined : (value << 16) >> 16
}

export function readU32(bytes, offset) {
  if (offset < 0 || offset + 4 > bytes.length) return undefined
  return (
    (bytes[offset] |
      (bytes[offset + 1] << 8) |
      (bytes[offset + 2] << 16) |
      (bytes[offset + 3] << 24)) >>>
    0
  )
}

export function readI32(bytes, offset) {
  const value = readU32(bytes, offset)
  return value === undefined ? undefined : value | 0
}

// u64 as a Number. Save offsets and sizes stay far below 2^53.
export function readU64(bytes, offset) {
  const low = readU32(bytes, offset)
  const high = readU32(bytes, offset + 4)
  if (low === undefined || high === undefined) return undefined
  return high * 0x100000000 + low
}

const utf8 = new TextDecoder('utf-8', { fatal: true })

// Strict UTF-8 decode; returns undefined on invalid input (Rust `str::from_utf8(..).ok()`).
export function decodeUtf8(bytes, start, end) {
  if (start < 0 || end > bytes.length || start > end) return undefined
  try {
    return utf8.decode(bytes.subarray(start, end))
  } catch {
    return undefined
  }
}

// Find `needle` wholly inside haystack[from, end); -1 when absent.
// Always pass `end` for windowed searches: an unbounded miss scans the whole buffer.
export function findBytes(haystack, needle, from = 0, end = haystack.length) {
  const first = needle[0]
  const limit = Math.min(end, haystack.length)
  const last = limit - needle.length
  // A bounded view keeps native indexOf from running past `end` on a miss.
  const view = limit === haystack.length ? haystack : haystack.subarray(0, limit)
  let at = view.indexOf(first, from)
  while (at !== -1 && at <= last) {
    let match = true
    for (let i = 1; i < needle.length; i++) {
      if (haystack[at + i] !== needle[i]) {
        match = false
        break
      }
    }
    if (match) return at
    at = view.indexOf(first, at + 1)
  }
  return -1
}

export class SaveParseError extends Error {
  constructor(kind, message) {
    super(message)
    this.name = 'SaveParseError'
    this.kind = kind // 'not_fm_save' | 'corrupt' | 'unsupported' | 'parse'
  }
}

export const corrupt = (message) => new SaveParseError('corrupt', message)
export const parseError = (message) => new SaveParseError('parse', message)
