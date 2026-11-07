// Manual-gated check against a real FM27 save, which can't be committed (≈210 MB).
// Run with:  FM27_SAVE=fm27save.fm npx vitest run src/utils/fmSave/realSave
// Thresholds are set from the user's fm27save.fm (build 27.0.0+2493426, clock 2036-07-01),
// loosened so another FM27 save of a similar shape also passes. The acceptance checks
// against fm27.csv (an in-game export of 143 players from that save) only run when the
// CSV is present.

import { closeSync, existsSync, fstatSync, openSync, readFileSync, readSync } from 'node:fs'
import { gzipSync } from 'node:zlib'
import { afterAll, beforeAll, describe, expect, it } from 'vitest'
import { SAVE_ONLY_HEADERS } from './exportCsv.js'
import { exportSave } from './exportSave.js'
import { SEASON_STAT_HEADERS } from './seasonStats.js'
import { readCsv } from './testing/csv.js'

const savePath = process.env.FM27_SAVE
const groundTruthPath = process.env.FM27_CSV ?? 'fm27.csv'

describe.skipIf(!savePath)('real FM27 save', () => {
  let result
  let exported
  let fd

  beforeAll(async () => {
    fd = openSync(savePath, 'r')
    const source = {
      size: fstatSync(fd).size,
      read: async (offset, length) => {
        const buffer = new Uint8Array(length)
        readSync(fd, buffer, 0, length, offset)
        return buffer
      },
    }
    // Peak memory per stage: every stage boundary is a sample (the work between is synchronous).
    const peaks = {}
    let stage = 'start'
    const sample = () => {
      const rss = process.memoryUsage().rss
      peaks[stage] = Math.max(peaks[stage] ?? 0, rss)
    }
    result = await exportSave(source, (progress) => {
      sample()
      stage = progress.stage
      sample()
    })
    sample()
    console.log(
      'rss MB at stage boundaries',
      JSON.stringify(
        Object.fromEntries(Object.entries(peaks).map(([k, v]) => [k, Math.round(v / 1e6)]))
      )
    )
    exported = readCsv(result.csv, ';')
    console.log('summary', JSON.stringify(result.summary))
    console.log('report', JSON.stringify(result.parsed.report))
  }, 240_000)

  afterAll(() => {
    if (fd !== undefined) closeSync(fd)
  })

  it('decodes players with names, attributes, clubs and wages', () => {
    const { metadata, players, report } = result.parsed
    const share = (predicate) => players.filter(predicate).length / players.length

    expect(metadata.build).toMatch(/^27\./)
    expect(report.players).toBeGreaterThan(10_000)
    expect(report.warnings).toEqual([])
    expect(share((p) => p.name)).toBeGreaterThan(0.98)
    expect(share((p) => p.age >= 14 && p.age <= 50)).toBeGreaterThan(0.98)
    expect(share((p) => p.nationName)).toBeGreaterThan(0.95)
    expect(share((p) => p.uniqueId)).toBeGreaterThan(0.99)
    expect(share((p) => p.clubName)).toBeGreaterThan(0.8)
    expect(share((p) => p.wage !== undefined)).toBeGreaterThan(0.5)
    for (const player of players.slice(0, 2000)) {
      for (const value of Object.values(player.attributes)) {
        expect(value).toBeGreaterThanOrEqual(1)
        expect(value).toBeLessThanOrEqual(20)
      }
    }
  })

  it('keeps the port leftovers within their known size (Tier 0 regression guard)', () => {
    const { players } = result.parsed
    const nameless = players.filter((p) => !p.name).length
    const withoutUniqueId = players.filter((p) => !p.uniqueId).length
    console.log('nameless', nameless, 'without Unique ID', withoutUniqueId)
    expect(nameless).toBeLessThanOrEqual(5)
    expect(withoutUniqueId).toBeLessThanOrEqual(10)
    const ids = players.map((p) => p.uniqueId).filter(Boolean)
    expect(new Set(ids).size).toBe(ids.length)
  })

  describe.skipIf(!existsSync(groundTruthPath))('against fm27.csv', () => {
    let truth
    let joined

    beforeAll(() => {
      truth = readCsv(readFileSync(groundTruthPath, 'utf8'), ',')
      const byUniqueId = new Map(exported.map((row) => [row['Unique ID'], row]))
      joined = truth.map((row) => [row, byUniqueId.get(row['Unique ID'])])
    })

    /** Count rows where `pick(truth)` equals `pick(export)`, logging the misses. */
    const agreement = (column, pick = (row) => row[column]) => {
      let same = 0
      const misses = []
      for (const [want, got] of joined) {
        if (got && pick(want) === pick(got)) same++
        else misses.push([want.Player, pick(want), got && pick(got)])
      }
      console.log(`${column}: ${same}/${joined.length}`, misses.slice(0, 8))
      return same
    }

    it('joins every row by Unique ID with matching name, DoB, Club and Based In', () => {
      expect(truth.length).toBe(143)
      expect(joined.filter(([, got]) => got).length).toBe(143)
      expect(agreement('Player')).toBe(143)
      expect(agreement('DoB', (row) => row.DoB.split(' ')[0])).toBe(143)
      expect(agreement('Club')).toBe(143)
      expect(agreement('Based In')).toBe(143)
    })

    it('derives Division for every row', () => {
      expect(agreement('Division')).toBe(143)
      const { players, report } = result.parsed
      const strong = players.filter((p) => p.name && p.ca >= 100)
      const covered = strong.filter((p) => p.leagueName).length
      console.log(
        `Division: clubs ${report.divisionClubs}, players ${report.withDivision}, ` +
          `CA>=100 coverage ${covered}/${strong.length} (${((100 * covered) / strong.length).toFixed(1)}%), ` +
          `${report.divisionMs} ms`
      )
    })

    it('derives the Inf flags for every row', () => {
      const counts = {}
      for (const [want] of joined)
        counts[want.Inf || 'blank'] = (counts[want.Inf || 'blank'] ?? 0) + 1
      console.log('fm27.csv Inf', JSON.stringify(counts))
      expect(agreement('Inf')).toBe(143)
      const flagged = exported.filter((row) => row.Inf)
      const tally = {}
      for (const row of flagged)
        for (const code of row.Inf.split(', ')) tally[code] = (tally[code] ?? 0) + 1
      console.log('export Inf codes', JSON.stringify(tally))
    })

    it('derives Media Handling for every row', () => {
      expect(agreement('Media Handling')).toBe(143)
    })

    it('derives Personality as read in-game for 14 players', () => {
      // In-game Personality labels read by the user on this save (ticket 07).
      const readings = [
        ['Lyan Pereira', 'Atlético de Madrid', 'Unambitious'],
        ['Kein Takeishi', 'Bayern München', 'Unambitious'],
        ['Mateo Garrido', 'Alavés', 'Leader'],
        ['Diego Mascardi', 'Hellas Verona', 'Perfectionist'],
        ['Carlos Núñez', 'Napoli', 'Model Professional'],
        ['Ribrayshen Robertha', 'Bayern München', 'Professional'],
        ['Pau Cubarsí', 'Barcelona', 'Model Citizen'],
        ['Martín Falcioni', 'Huracán', 'Temperamental'],
        ['Dejan Causevic', 'Dinamo Zagreb', 'Low Self Belief'],
        ['Klebinho', 'Botafogo', 'Realist'],
        ['Lulinha', 'Palmeiras', 'Mercenary'],
        ['Marcos Martins', 'Bahia', 'Fickle'],
        ['Diogo Marcondes', 'Porto', 'Balanced'],
        ['Basílio', 'Benfica', 'Light-Hearted'],
      ]
      let same = 0
      for (const [name, club, label] of readings) {
        const rows = exported.filter((row) => row.Player === name && row.Club === club)
        if (rows.length === 1 && rows[0].Personality === label) same++
        else
          console.log(
            'Personality miss',
            name,
            club,
            label,
            rows.map((r) => r.Personality)
          )
      }
      console.log(`Personality: ${same}/${readings.length}`)
      expect(same).toBe(readings.length)
    })

    it('derives the 42 current-season stat columns', () => {
      const approximate = new Set(['Rating', 'Average Rating Club'])
      const isZero = (cell) => /^0(\.0+)?%?$/.test(cell)
      const results = SEASON_STAT_HEADERS.map((column) => {
        let same = 0
        let blank = 0
        const misses = []
        for (const [want, got] of joined) {
          const ours = got?.[column] ?? ''
          const theirs = want[column]
          let ok
          if (ours === '') {
            // Blank = missing advanced block; the in-game export prints zeros there.
            blank++
            ok = isZero(theirs)
          } else if (approximate.has(column) && theirs !== '-' && ours !== '-') {
            ok = Math.abs(Number(ours) - Number(theirs)) <= 0.02 + 1e-9
          } else ok = ours === theirs
          if (ok) same++
          else misses.push(`${want.Player}: csv ${theirs} ours ${ours}`)
        }
        return { column, same, blank, misses }
      })
      for (const r of results) {
        console.log(
          `${r.column.padEnd(22)} ${r.same}/${joined.length} blank ${r.blank}`,
          r.misses.slice(0, 3).join(' | ')
        )
      }
      const exact = results.filter((r) => !approximate.has(r.column) && r.same === 143).length
      console.log('exact stat columns', exact, '/ 40; withStats', result.parsed.report.withStats)
      for (const r of results) expect(r.same, r.column).toBe(143)
    })
  })

  it('detects the managed club', () => {
    console.log(
      'managedClub',
      JSON.stringify(result.summary.managedClub),
      'uid',
      result.parsed.metadata.managedClubUid
    )
    expect(result.summary.managedClub?.club).toBe('PFK Montana')
    expect(result.summary.managedClub?.division).toBeTruthy()
  })

  it('carries the save-only columns and stays under the 18 MB upload guard', () => {
    const header = result.csv.slice(0, result.csv.indexOf('\n')).split(';')
    expect(header.slice(-16)).toEqual(SAVE_ONLY_HEADERS)
    const withSaveCa = exported.filter((row) => row.CA).length
    const withHidden = exported.filter((row) => row.Consistency && row.Temperament).length
    console.log(
      'rows with CA',
      withSaveCa,
      'with hidden attributes',
      withHidden,
      'of',
      exported.length
    )
    expect(withSaveCa / exported.length).toBeGreaterThan(0.99)
    const bytes = gzipSync(result.csv).length
    console.log('export gzipped MB', (bytes / 1e6).toFixed(2))
    expect(bytes).toBeLessThan(18_000_000)
  })
})
