import { readFileSync, writeFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import { clubDisplayName } from './clubs.js'
import {
  CSV_HEADERS,
  formatMoney,
  formatPositions,
  playersToCsv,
  SAVE_ONLY_HEADERS,
} from './exportCsv.js'
import { ATTRIBUTE_KEYS, POSITION_NAMES } from './layout.js'
import { NATION_FIFA_CODES } from './nation-codes.js'
import { NationIndex, nationName } from './nations.js'
import { statLine } from './seasonStats.js'

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
  // Ambition 18, loyalty 9: Ambitious. Temperament 5: Volatile; loyalty < 11: Media-friendly.
  personality: {
    adaptability: 10,
    ambition: 18,
    loyalty: 9,
    pressure: 10,
    professionalism: 12,
    sportsmanship: 12,
    temperament: 5,
    controversy: 10,
  },
  gender: 0,
  ca: 152,
  pa: 171,
  seasonStats: statLine({
    minutes: 1800,
    starts: 20,
    rated: 20,
    ratingSum: 1400,
    goals: 5,
    passesAttempted: 900,
    passesCompleted: 810,
    distanceTenths: 2000,
    possessionLost: 90,
  }),
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
    // Refresh it after an intended export change with UPDATE_FMSAVE_FIXTURE=1.
    const fixturePath = 'src/api/testdata/fmsave/save_export.csv'
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
        // Advanced block missing (distance and possession lost 0): those cells are blank.
        seasonStats: statLine({ minutes: 900, starts: 10, rated: 0, conceded: 9 }),
      }),
    ])
    if (process.env.UPDATE_FMSAVE_FIXTURE) writeFileSync(fixturePath, csv)
    expect(csv).toBe(readFileSync(fixturePath, 'utf8'))
  })
})

describe('Tier 0 port fixes', () => {
  const cellsOf = (player) => {
    const [header, row] = playersToCsv([player]).csv.trim().split('\n')
    return Object.fromEntries(header.split(';').map((h, i) => [h, row.split(';')[i]]))
  }

  it('Club prefers the short name and falls back to the long one', () => {
    expect(
      clubDisplayName({ name: 'Profesionalen Futbolen Klub Montana', shortName: 'PFK Montana' })
    ).toBe('PFK Montana')
    expect(clubDisplayName({ name: 'Arsenal FC', shortName: '' })).toBe('Arsenal FC')
    expect(clubDisplayName({ name: 'Arsenal FC' })).toBe('Arsenal FC')
    expect(clubDisplayName(undefined)).toBeUndefined()
    expect(cellsOf(samplePlayer({ clubName: 'PFK Montana' })).Club).toBe('PFK Montana')
  })

  it('uses FM short nation names, which reach Based In and the Nationality fallback', () => {
    expect(nationName(799)).toBe('Türkiye')
    expect(nationName(390)).toBe('U.S.A.')
    expect(nationName(53)).toBe('DR Congo')
    expect(nationName(49)).toBe('Congo')
    expect(nationName(765)).toBe('England') // not overridden
    // parse.js routes Based In and Nationality through NationIndex.nameFor -> nationName.
    const nations = new NationIndex([
      { nationId: 0, databaseId: 785, name: nationName(785) },
      { nationId: 1, databaseId: 17, name: nationName(17) },
    ])
    expect(nations.nameFor(0)).toBe('N.Ireland')
    const cells = cellsOf(
      samplePlayer({
        nationDatabaseId: 999_999, // no FIFA code -> Nationality carries the name
        nationName: nations.nameFor(1),
        clubNationName: nations.nameFor(0),
      })
    )
    expect(cells.Nationality).toBe('Eq. Guinea')
    expect(cells['Based In']).toBe('N.Ireland')
  })

  it('codes Congo (49) as CGO and DR Congo (53) as COD', () => {
    expect(NATION_FIFA_CODES[49]).toBe('CGO')
    expect(NATION_FIFA_CODES[53]).toBe('COD')
    expect(cellsOf(samplePlayer({ nationDatabaseId: 53 })).Nation).toBe('COD')
  })
})

describe('save-only columns', () => {
  const cellsOf = (player) => {
    const [header, row] = playersToCsv([player]).csv.trim().split('\n')
    return Object.fromEntries(header.split(';').map((h, i) => [h, row.split(';')[i]]))
  }

  it('appends CA, PA, the 13 hidden attributes and Inf, in order, at the end', () => {
    expect(SAVE_ONLY_HEADERS).toEqual([
      'CA',
      'PA',
      'Consistency',
      'Dirtiness',
      'Important Matches',
      'Injury Proneness',
      'Versatility',
      'Adaptability',
      'Ambition',
      'Controversy',
      'Loyalty',
      'Pressure',
      'Professionalism',
      'Sportsmanship',
      'Temperament',
      'Inf',
    ])
    expect(CSV_HEADERS.slice(-16)).toEqual(SAVE_ONLY_HEADERS)
  })

  it('fills the values from the port', () => {
    const cells = cellsOf(
      samplePlayer({
        attributes: { ...samplePlayer().attributes, consistency: 14, injury_proneness: 3 },
        personality: { ...samplePlayer().personality, ambition: 18, temperament: 5 },
        injured: true,
      })
    )
    expect(cells).toMatchObject({
      CA: '152',
      PA: '171',
      Consistency: '14',
      'Injury Proneness': '3',
      Ambition: '18',
      Temperament: '5',
      Inf: 'Inj',
    })
  })

  it('blanks missing values', () => {
    const cells = cellsOf(samplePlayer({ ca: 0, pa: undefined, personality: undefined }))
    for (const header of ['CA', 'PA', 'Adaptability', 'Temperament', 'Inf']) {
      expect(cells[header], header).toBe('')
    }
  })
})
