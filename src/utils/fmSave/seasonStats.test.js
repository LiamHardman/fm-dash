import { zstdCompressSync } from 'node:zlib'
import { describe, expect, it } from 'vitest'
import { bufferSource, SaveFile } from './container.js'
import {
  advancedBlockMissing,
  fixed2,
  readOverallLines,
  readSeasonStatsBody,
  SEASON_STAT_HEADERS,
  seasonStatCells,
} from './seasonStats.js'
import { buildSave } from './testing/buildSave.js'

const HEADER = [0x03, 0x01, 0x74, 0x61, 0x64, 0x2e, 0x0f, 0x00]
const le32 = (v) => [v & 0xff, (v >>> 8) & 0xff, (v >>> 16) & 0xff, v >>> 24]

// A 155-byte line: fields are { offset: [value, width] } with width 1 or 2 (LE).
function line(fields = {}) {
  const bytes = new Uint8Array(155)
  bytes[0] = 0x01
  bytes[1] = 0x07
  for (const [offset, [value, width]] of Object.entries(fields)) {
    const at = Number(offset)
    bytes[at] = value & 0xff
    if (width === 2) bytes[at + 1] = (value >> 8) & 0xff
  }
  return [...bytes]
}
const slots = (count, filled) =>
  Array.from({ length: count }, (_, i) => (filled[i] ? filled[i] : [0])).flat()

/** One record: own = {slotIndex: lineBytes}, others = [{teamId, lines: {slotIndex: lineBytes}}]. */
function record(key, { own = {}, others = [], ratingItems = 0, form = 0 } = {}) {
  return [
    1,
    ...le32(key),
    ...new Array(16).fill(0xaa),
    ...slots(8, own),
    ...new Array(20).fill(0),
    ...le32(others.length),
    ...others.flatMap(({ teamId, lines }) => [0, ...le32(teamId), ...slots(6, lines)]),
    ...le32(ratingItems),
    ...new Array(ratingItems * 8).fill(7),
    form,
    ...new Array(form * 3).fill(3),
  ]
}
const section = (...records) =>
  new Uint8Array([...HEADER, ...records.flat(), 0, ...new Array(16).fill(0)])

// A realistic outfield line (offsets as in seasonStats.js LINE_FIELDS).
const outfield = (extra = {}) =>
  line({
    2: [7000, 2], // rating sum: 700.0 over 100 rated -> 7.00
    4: [1800, 2], // minutes
    6: [18, 1], // starts
    7: [2, 1], // subs
    8: [100, 1], // rated apps
    9: [4, 1], // goals
    10: [2, 1], // assists
    27: [800, 2], // passes attempted
    29: [700, 2], // passes completed
    31: [30, 2], // tackles attempted
    33: [20, 2], // tackles completed
    37: [10, 2], // headers won
    45: [40, 2], // shots
    47: [20, 2], // shots on target
    65: [2000, 2], // distance tenths
    83: [5, 2], // blocks
    87: [100, 2], // possession lost
    91: [450, 2], // xG hundredths
    95: [12, 2], // clearances
    121: [60, 1], // pressures attempted
    123: [20, 1], // pressures completed
    131: [30, 2], // crosses attempted
    133: [7, 2], // crosses completed
    ...extra,
  })

const cellMap = (cells) => Object.fromEntries(SEASON_STAT_HEADERS.map((h, i) => [h, cells[i]]))

describe('readOverallLines', () => {
  it('walks records, keeps the Overall slot and sums other-team Overall lines', () => {
    const body = section(
      record(11, { own: { 1: outfield(), 5: outfield() }, ratingItems: 2, form: 3 }),
      record(12, {
        own: { 5: line({ 4: [100, 2], 9: [1, 1] }) },
        others: [
          { teamId: 900, lines: { 0: line({ 4: [999, 2] }), 5: line({ 4: [50, 2], 9: [2, 1] }) } },
        ],
      }),
      record(13, { own: { 1: outfield() } }) // no Overall line
    )
    const overall = readOverallLines(body)
    expect([...overall.keys()]).toEqual([11, 12, 13])
    expect(overall.get(11)).toMatchObject({ minutes: 1800, goals: 4, pressuresAttempted: 60 })
    expect(overall.get(12)).toMatchObject({ minutes: 150, goals: 3 }) // 100+50, not the slot-0 999
    expect(overall.get(13)).toBeUndefined()
  })

  it('reads xGP as a signed word', () => {
    const body = section(record(1, { own: { 5: line({ 4: [900, 2], 113: [0xffce, 2] }) } }))
    expect(readOverallLines(body).get(1).xgpHundredths).toBe(-50)
  })

  it('fails closed on a bad footer, a bad slot, a repeated key or a wrong header', () => {
    const good = section(record(1))
    expect(() => readOverallLines(good.subarray(0, good.length - 1))).toThrow(/footer/)
    const badSlot = section(record(1, { own: { 2: [0x01, 0x09, ...new Array(153).fill(0)] } }))
    expect(() => readOverallLines(badSlot)).toThrow(/neither absent nor a line/)
    expect(() => readOverallLines(section(record(1), record(1)))).toThrow(/repeat player key/)
    expect(() => readOverallLines(new Uint8Array([1, 2, 3, 4, 5, 6, 7, 8, 0]))).toThrow(/header/)
    const tooManyTeams = section(
      record(1, { others: new Array(17).fill({ teamId: 1, lines: {} }) })
    )
    expect(() => readOverallLines(tooManyTeams)).toThrow(/other teams/)
  })
})

describe('readSeasonStatsBody', () => {
  it('picks the stats frame among decoy tad. frames in the gap after non_pl_hist_ls', async () => {
    const stats = section(record(5, { own: { 5: outfield() } }))
    const decoySchema4 = new Uint8Array([0x03, 0x01, 0x74, 0x61, 0x64, 0x2e, 0x04, 0x00, 1, 2, 3])
    const big = new Uint8Array(200_000).map((_, i) => (i * 7919) % 256) // larger, not stats
    const frames = [decoySchema4, big, stats].map((b) => new Uint8Array(zstdCompressSync(b)))
    const save = await SaveFile.open(
      bufferSource(
        buildSave([
          ['game_db', new Uint8Array([1])],
          ['non_pl_hist_ls', new Uint8Array([2]), frames],
          ['later', new Uint8Array([3])],
        ])
      )
    )
    const body = await readSeasonStatsBody(save)
    expect(body).toEqual(stats)
  })

  it('fails when no frame holds the section', async () => {
    const save = await SaveFile.open(
      bufferSource(
        buildSave([
          [
            'non_pl_hist_ls',
            new Uint8Array([2]),
            [new Uint8Array(zstdCompressSync(new Uint8Array(9)))],
          ],
        ])
      )
    )
    await expect(readSeasonStatsBody(save)).rejects.toThrow(/holds the season-statistics/)
  })
})

describe('fixed2', () => {
  it('rounds exact ties half to even and everything else correctly', () => {
    expect(fixed2(0.125)).toBe('0.12')
    expect(fixed2(0.375)).toBe('0.38')
    expect(fixed2(3.375)).toBe('3.38')
    expect(fixed2(-0.125)).toBe('-0.12')
    // 54.825 is not exact in binary (the double is just above it); FM printed 54.83 here.
    expect(fixed2((2193 * 90) / 3600)).toBe('54.83')
    expect(fixed2(0.025)).toBe('0.03') // 0.025 is slightly above the tie in binary
    expect(fixed2(1 / 3)).toBe('0.33')
  })
})

describe('seasonStatCells', () => {
  const statsLine = (bytes) => readOverallLines(section(record(1, { own: { 5: bytes } }))).get(1)

  it('formats an outfield line like the in-game export', () => {
    const cells = cellMap(seasonStatCells(statsLine(outfield()), false))
    expect(cells).toMatchObject({
      'Pas %': '88%', // 87.5 rounds
      'Tck R': '67', // 66.67 rounds
      'Cr C/A': '23', // 23.33 floors
      'Conv %': '10%',
      Apps: '18 (2)',
      Minutes: '1800',
      Rating: '7.00',
      'Average Rating Club': '7.00',
      'Ps A/90': '40.00',
      'Goals per 90 minutes': '0.20',
      'Mins/Gl': '450',
      'Hdrs W/90': '0.50',
      'Blk/90': '0.25',
      'Clr/90': '0.60',
      'Pres A/90': '3.00',
      'Pres C/90': '1.00',
      'Dist/90': '10.00',
      'xG/90': '0.23', // 4.5 * 90 / 1800 = 0.225 (not an exact binary tie) -> 0.23
      'Sv %': '0',
      'Saves/90': '0',
    })
  })

  it('uses the goalkeeper readings of the shared words', () => {
    const gk = statsLine(
      outfield({ 35: [30, 2], 37: [20, 2], 69: [10, 2], 83: [79, 2], 95: [18, 2] })
    )
    const cells = cellMap(seasonStatCells(gk, true))
    expect(cells['Sv %']).toBe('75') // floor(60 / 79 = 75.9)
    expect(cells['Saves/90']).toBe('3') // floor(60 * 90 / 1800 = 3.0)
    expect(cells['Hdrs W/90']).toBe('0.00')
    expect(cells['Blk/90']).toBe('0.00')
    expect(cells['Clr/90']).toBe('0.90') // kept for goalkeepers
  })

  it('shows 0.00 per 90 below 90 minutes but keeps counts and percentages', () => {
    const cells = cellMap(
      seasonStatCells(statsLine(outfield({ 4: [89, 2], 6: [0, 1], 7: [1, 1] })), false)
    )
    expect(cells['Ps A/90']).toBe('0.00')
    expect(cells['Goals per 90 minutes']).toBe('0.00')
    expect(cells['Pas %']).toBe('88%')
    expect(cells.Apps).toBe('0 (1)')
    expect(cells['Mins/Gl']).toBe('22.25')
  })

  it('leaves the advanced cells blank when the advanced block is missing', () => {
    const missing = statsLine(outfield({ 65: [0, 2], 87: [0, 2] }))
    expect(advancedBlockMissing(missing)).toBe(true)
    const cells = cellMap(seasonStatCells(missing, true))
    for (const header of ['Dist/90', 'Poss Lost/90', 'Pres A/90', 'Cr C/A', 'Sv %', 'Saves/90']) {
      expect(cells[header], header).toBe('')
    }
    expect(cells['xA/90']).toBe('0.00') // xA is emitted
    expect(cells['Ps A/90']).toBe('40.00')
    expect(advancedBlockMissing(statsLine(outfield({ 65: [0, 2] })))).toBe(false)
  })

  it('prints - without rated apps and all zeros without an Overall line', () => {
    expect(cellMap(seasonStatCells(statsLine(outfield({ 8: [0, 1] })), false)).Rating).toBe('-')
    const none = cellMap(seasonStatCells(undefined, false))
    expect(none).toMatchObject({
      Rating: '-',
      'Average Rating Club': '-',
      Apps: '0',
      Minutes: '0',
      'Pas %': '0%',
      'Ps A/90': '0.00',
      'Mins/Gl': '0',
      'Sv %': '0',
    })
  })

  it('prints Mins/Gl with 15 significant digits', () => {
    const cells = cellMap(seasonStatCells(statsLine(outfield({ 4: [3239, 2], 9: [17, 1] })), false))
    expect(cells['Mins/Gl']).toBe('190.529411764706')
  })

  it('computes a two-team season from the summed lines', () => {
    const body = section(
      record(1, {
        own: { 5: outfield({ 4: [900, 2], 9: [1, 1], 8: [10, 1], 2: [700, 2] }) },
        others: [
          {
            teamId: 3,
            lines: { 5: outfield({ 4: [900, 2], 9: [3, 1], 8: [10, 1], 2: [640, 2] }) },
          },
        ],
      })
    )
    const cells = cellMap(seasonStatCells(readOverallLines(body).get(1), false))
    expect(cells.Minutes).toBe('1800')
    expect(cells['Goals per 90 minutes']).toBe('0.20')
    expect(cells.Rating).toBe('6.70') // (700 + 640) / 10 / 20
  })
})
