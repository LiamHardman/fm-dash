import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import { CSV_HEADERS, formatMoney, formatPositions, playersToCsv } from './exportCsv.js'
import { ATTRIBUTE_KEYS, POSITION_NAMES } from './layout.js'

const ratings = (overrides) =>
  Object.fromEntries(POSITION_NAMES.map((name) => [name, overrides[name] ?? 1]))

const samplePlayer = (overrides = {}) => ({
  name: 'Bruno Costa',
  nationDatabaseId: 765, // England
  nationName: 'England',
  clubName: 'Arsenal',
  leagueName: 'Premier League',
  clubNationName: 'England',
  age: 26,
  birthDate: '2010-02-01',
  transferValue: 153_400_000,
  wage: 110_000,
  uniqueId: 2_002_330_039,
  leftFoot: 20,
  rightFoot: 6,
  positions: ratings({ DC: 20, DR: 18, DM: 15, MC: 12 }),
  attributes: Object.fromEntries(ATTRIBUTE_KEYS.map((key, i) => [key, (i % 20) + 1])),
  ...overrides,
})

describe('formatPositions', () => {
  it('groups sides in FM order and keeps accomplished positions', () => {
    expect(formatPositions(ratings({ DC: 20, DR: 18, DM: 15, MC: 12 }))).toBe('D (RC), DM')
    expect(formatPositions(ratings({ AMR: 20, AML: 16, STC: 19, MR: 15 }))).toBe(
      'M (R), AM (RL), ST (C)'
    )
    expect(formatPositions(ratings({ GK: 20 }))).toBe('GK')
    expect(formatPositions(ratings({ WBL: 17, DL: 17 }))).toBe('D (L), WB (L)')
  })

  it('falls back to the best-rated position when none is accomplished', () => {
    expect(formatPositions(ratings({ MC: 11, DM: 11, DC: 9 }))).toBe('DM, M (C)')
  })
})

describe('formatMoney', () => {
  it('formats like FM', () => {
    expect(formatMoney(153_400_000)).toBe('£153.4M')
    expect(formatMoney(8_500_000)).toBe('£8.5M')
    expect(formatMoney(631_467)).toBe('£631.5K')
    expect(formatMoney(900)).toBe('£900')
    expect(formatMoney(undefined)).toBe('')
  })
})

describe('playersToCsv', () => {
  it('writes FM export headers and one row per named player', () => {
    const { csv, rows, skipped } = playersToCsv([samplePlayer(), samplePlayer({ name: undefined })])
    const [header, row] = csv.trim().split('\n')
    expect(rows).toBe(1)
    expect(skipped).toBe(1)
    expect(header.split(';')).toEqual(CSV_HEADERS)
    const cells = Object.fromEntries(header.split(';').map((h, i) => [h, row.split(';')[i]]))
    expect(cells).toMatchObject({
      Player: 'Bruno Costa',
      Nation: 'ENG',
      Nationality: '',
      Position: 'D (RC), DM',
      'Transfer Value': '£153.4M',
      Wage: '£110K p/w',
      'Unique ID': '2002330039',
      DoB: '1/2/2010',
      Crossing: '1',
    })
  })

  it('falls back to the full nation name when there is no FIFA code', () => {
    const { csv } = playersToCsv([
      samplePlayer({ nationDatabaseId: 999_999, nationName: 'Zanzibar' }),
    ])
    const [header, row] = csv.trim().split('\n')
    const cells = Object.fromEntries(header.split(';').map((h, i) => [h, row.split(';')[i]]))
    expect(cells.Nation).toBe('')
    expect(cells.Nationality).toBe('Zanzibar')
  })

  it('quotes cells holding the delimiter', () => {
    const { csv } = playersToCsv([samplePlayer({ clubName: 'A;B "C"' })])
    expect(csv).toContain('"A;B ""C"""')
  })

  it('matches the Go round-trip fixture', () => {
    // src/api/fm_save_csv_roundtrip_test.go parses this file with the real backend parser.
    const fixture = readFileSync('src/api/testdata/fmsave/save_export.csv', 'utf8')
    const { csv } = playersToCsv([
      samplePlayer(),
      samplePlayer({
        name: 'Ana Silva',
        nationDatabaseId: 999_999,
        nationName: 'Zanzibar',
        clubName: undefined,
        leagueName: undefined,
        clubNationName: undefined,
        transferValue: undefined,
        wage: undefined,
        positions: ratings({ GK: 20 }),
      }),
    ])
    expect(csv).toBe(fixture)
  })
})
