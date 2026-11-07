import { describe, expect, it } from 'vitest'
import { ContractDecoder } from './contracts.js'
import { infCell, infFlags } from './inf.js'
import { readInjuredPindexes } from './injuries.js'

const le16 = (v) => [v & 0xff, (v >> 8) & 0xff]
const le32 = (v) => [v & 0xff, (v >>> 8) & 0xff, (v >>> 16) & 0xff, v >>> 24]
// FM date: day-of-year word, then year.
const date = (year, dayOfYear) => [...le16(dayOfYear), ...le16(year)]

describe('contract tail listing state', () => {
  // A contract chain record with a tail 66 bytes before its tag (no events).
  function chainRecord({ flags, code }) {
    const bytes = new Uint8Array(200)
    const tail = 10
    bytes.set([0, 0, 0, 3], tail)
    bytes.set(le32(4), tail + 4)
    bytes.set(date(2040, 180), tail + 28)
    bytes[tail + 39] = flags
    bytes[tail + 40] = code
    bytes.set(le32(0), tail + 42) // event count
    const tag = tail + 66
    bytes.set([0x01, 0x00, 0x6c, 0x07], tag)
    bytes.set(date(2037, 182), tag - 19) // start
    bytes.set(le32(5000), tag + 17) // wage
    return new ContractDecoder(undefined, '2039-08-08').decodeChainRecord(bytes, tag)
  }

  it('reads the listing flags (+39) and code (+40)', () => {
    const record = chainRecord({ flags: 0b11, code: 5 })
    expect(record.hasTerms).toBe(true)
    expect(record.listingFlags).toBe(3)
    expect(record.listingCode).toBe(5)
    expect(record.wage).toBe(5000)
  })
})

describe('infFlags', () => {
  it('Loa from bit 1', () => {
    expect(infFlags({ listingFlags: 0b10 })).toEqual(['Loa'])
    expect(infFlags({ listingFlags: 0b10, listingCode: 3 })).toEqual(['Loa'])
  })

  it('Wnt only with bit 0 AND code 5', () => {
    expect(infFlags({ listingFlags: 0b01, listingCode: 5 })).toEqual(['Wnt'])
    expect(infFlags({ listingFlags: 0b01, listingCode: 4 })).toEqual([])
    expect(infFlags({ listingFlags: 0, listingCode: 5 })).toEqual([])
  })

  it('joins flags in the order Inj, Loa, Wnt', () => {
    expect(infCell({ injured: true, listingFlags: 0b11, listingCode: 5 })).toBe('Inj, Loa, Wnt')
    expect(infCell({ injured: true })).toBe('Inj')
    expect(infCell({})).toBe('')
  })
})

describe('readInjuredPindexes', () => {
  // A window row: byte 0, recovery date @1, expected return @5, selector (pindex+1) @9, byte 13.
  const row = (returnDate, pindex) => [0, ...date(2039, 200), ...returnDate, ...le32(pindex + 1), 0]
  const section = (array0, array1 = []) =>
    new Uint8Array([
      ...new Array(12).fill(0),
      ...le32(array0.length),
      ...array0.flat(),
      ...le32(array1.length),
      ...array1.flat(),
      ...le32(0), // typed array (stride 13)
      ...le32(0), // history log (stride 15)
      ...le32(0),
      ...le32(0),
      ...le32(0),
      1,
      0,
      0,
      0,
      0,
      0,
      0xff,
      0xff,
    ])

  it('marks players whose expected return is on or after the game date', () => {
    const injured = readInjuredPindexes(
      section(
        [row(date(2039, 230), 41), row(date(2039, 210), 42)], // 2039-08-18; 2039-07-29 (past)
        [row(date(2039, 220), 43)] // 2039-08-08: returns today, still listed
      ),
      '2039-08-08'
    )
    expect([...injured].sort()).toEqual([41, 43])
  })

  it('rejects a section whose counts or tail do not fit', () => {
    const good = section([row(date(2039, 230), 1)])
    expect(() => readInjuredPindexes(good.subarray(0, good.length - 1), '2039-08-08')).toThrow()
    const tooMany = good.slice()
    tooMany.set(le32(9999), 12)
    expect(() => readInjuredPindexes(tooMany, '2039-08-08')).toThrow(/runs past/)
  })
})

describe('attachInjuries', () => {
  it('warns and leaves players uninjured when injury_manager is missing', async () => {
    const { attachInjuries } = await import('./parse.js')
    const players = [{ pindex: 1 }]
    const warnings = []
    const save = {
      readSection: async () => {
        throw new Error('save has no "injury_manager" section')
      },
    }
    await attachInjuries(save, players, '2039-08-08', warnings)
    expect(warnings[0]).toMatch(/^injuries unavailable:/)
    expect(players[0].injured).toBeUndefined()
    expect(infCell(players[0])).toBe('')
  })
})
