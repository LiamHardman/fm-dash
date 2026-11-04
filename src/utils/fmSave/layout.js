// FM27 layout constants and date/attribute helpers.
// Translated from fm27-scout-code (MIT, Copyright (c) 2026 FM27 Scout contributors),
// crates/fm27-core/src/layout.rs. See ./THIRD_PARTY.md.

import { readU16 } from './bytes.js'

export const GAME = 'FM27'
export const VERIFIED_BUILDS = ['27.0.0+2486695', '27.0.0+2493426']
export const EARLIEST_YEAR = 1901
export const LATEST_YEAR = 2200

export const POSITION_NAMES = [
  'GK',
  'SW',
  'DL',
  'DC',
  'DR',
  'DM',
  'ML',
  'MC',
  'MR',
  'AML',
  'AMC',
  'AMR',
  'STC',
  'WBL',
  'WBR',
]

// The 52 stored attributes, in save order (feet are stored between technique and flair
// and are excluded here; see storedAttributeIndex).
export const ATTRIBUTE_KEYS = [
  'crossing',
  'dribbling',
  'finishing',
  'heading',
  'long_shots',
  'marking',
  'off_the_ball',
  'passing',
  'penalty_taking',
  'tackling',
  'vision',
  'handling',
  'aerial_reach',
  'command_of_area',
  'communication',
  'kicking',
  'throwing',
  'anticipation',
  'decisions',
  'one_on_ones',
  'positioning',
  'reflexes',
  'first_touch',
  'technique',
  'flair',
  'corners',
  'teamwork',
  'work_rate',
  'long_throws',
  'eccentricity',
  'rushing_out',
  'punching',
  'acceleration',
  'free_kick_taking',
  'strength',
  'stamina',
  'pace',
  'jumping_reach',
  'leadership',
  'dirtiness',
  'balance',
  'bravery',
  'consistency',
  'aggression',
  'agility',
  'important_matches',
  'injury_proneness',
  'versatility',
  'natural_fitness',
  'determination',
  'composure',
  'concentration',
]

export const HIDDEN_ATTRIBUTE_KEYS = [
  'dirtiness',
  'consistency',
  'important_matches',
  'injury_proneness',
  'versatility',
]

export const PERSONALITY_KEYS = [
  'adaptability',
  'ambition',
  'loyalty',
  'pressure',
  'professionalism',
  'sportsmanship',
  'temperament',
  'controversy',
]

export const PlayerLayout = Object.freeze({
  MARKER: [0x01, 0x00, 0x6c, 0x07],
  MARKER_OFFSET: 102,
  PINDEX_OFFSET: -19,
  UID_OFFSET: -15,
  UID_COPY_OFFSET: -11,
  OBJECT_KIND_OFFSET: -7,
  HOME_REPUTATION_OFFSET: -6,
  CURRENT_REPUTATION_OFFSET: -4,
  WORLD_REPUTATION_OFFSET: -2,
  CA_OFFSET: 0,
  PA_OFFSET: 2,
  REPUTATION_BUCKET_OFFSET: 8,
  TEAM_ID_OFFSET: 16,
  RATINGS_OFFSET: 24,
  RATINGS_COUNT: 15,
  ATTRIBUTES_OFFSET: 39,
  ATTRIBUTE_COUNT: 54,
  LEFT_FOOT_INDEX: 24,
  RIGHT_FOOT_INDEX: 25,
  TRANSFER_VALUE_OFFSET: 93,
  TRANSFER_VALUE_PLACEHOLDER: 300_000_000,
  CLUB_JOIN_DATE_OFFSET: 98,
  MATCH_SHARPNESS_OFFSET: 106,
  CONDITION_OFFSET: 110,
  HEIGHT_OFFSET: 121,
  DECODE_EXTENT: 122,
  ALLOCATED_UID_MINIMUM: 2_000_000_000,
})

export const PersonLayout = Object.freeze({
  WINDOW_START_OFFSET: 100,
  PERSONALITY_OFFSET_FROM_BIRTH: 22,
  PERSONALITY_COUNT: 8,
  DATE_ZERO_BYTES_OFFSET_FROM_BIRTH: 15,
  DATE_ZERO_BYTES_COUNT: 6,
  NATION_ID_OFFSET_FROM_BIRTH: 14,
  RELATION_PRESENT_OFFSET_FROM_BIRTH: 38,
  RELATION_COUNT_OFFSET_FROM_BIRTH: 39,
  RELATION_ENTRIES_OFFSET_FROM_BIRTH: 40,
  RELATION_ENTRY_BYTES: 16,
  LEGAL_NAME_LENGTH_MIN: 2,
  LEGAL_NAME_LENGTH_MAX: 80,
  LEGAL_NAME_OFFSET_FROM_BLOCK_START: 19,
  TRAIT_BITS_OFFSET_FROM_BLOCK_START: -8,
  NAME_ID_LIMIT: 1 << 23,
  SECOND_NATION_PAIR: [8, 9],
  HOME_GROWN_NATION_PAIR: [8, 70],
  HOME_GROWN_CLUB_PAIR: [1, 72],
  SECOND_NATION_QUALIFIERS: [15, 100],
  RELATION_SENTINEL_VALUE: 0xff,
})

const NAME_POOL_WORDS = [46421, 0, 1024, 256, 2048, 0]
export const NAME_POOL_SIGNATURE = (() => {
  const out = new Uint8Array(24)
  const view = new DataView(out.buffer)
  NAME_POOL_WORDS.forEach((word, i) => view.setUint32(i * 4, word, true))
  return out
})()
export const NAME_POOL_PREFIX = new Uint8Array([0x55, 0xb5, 0x00, 0x00, 0x00]) // 46421 = 0xB555

const isLeapYear = (year) => (year % 4 === 0 && year % 100 !== 0) || year % 400 === 0

/**
 * Decode a stored game date. `offset` points at the low 16-bit day/slot word; the
 * low 9 bits are the 1-based day of the year, the next u16 is the year.
 * Returns an ISO `YYYY-MM-DD` string, or undefined (Rust `NaiveDate::from_yo_opt`).
 */
export function decodeDate(bytes, offset) {
  const packed = readU16(bytes, offset)
  const year = readU16(bytes, offset + 2)
  if (packed === undefined || year === undefined) return undefined
  const dayOfYear = packed & 0x1ff
  if (year < EARLIEST_YEAR || year > LATEST_YEAR) return undefined
  if (dayOfYear < 1 || dayOfYear > (isLeapYear(year) ? 366 : 365)) return undefined
  const date = new Date(Date.UTC(year, 0, dayOfYear))
  return date.toISOString().slice(0, 10)
}

/** Whole years between an ISO birth date and an ISO clock date. */
export function ageAt(birth, clock) {
  const [by, bm, bd] = birth.split('-').map(Number)
  const [cy, cm, cd] = clock.split('-').map(Number)
  let age = cy - by
  if (cm < bm || (cm === bm && cd < bd)) age -= 1
  return age
}

/** Raw 1..100 attribute → the 1..20 scale the game shows. */
export function scaleAttribute(raw) {
  return Math.max(1, Math.floor((raw + 2) / 5))
}

/** Byte index of an attribute inside the stored 54-byte slice (feet sit at 24 and 25). */
export function storedAttributeIndex(index) {
  return index < PlayerLayout.LEFT_FOOT_INDEX ? index : index + 2
}
