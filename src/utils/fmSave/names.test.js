import { describe, expect, it } from 'vitest'
import { ageAt, decodeDate, NAME_POOL_SIGNATURE, scaleAttribute } from './layout.js'
import { findVersion } from './metadata.js'
import { NamePools } from './names.js'

const enc = new TextEncoder()
const u32 = (value) => {
  const out = new Uint8Array(4)
  new DataView(out.buffer).setUint32(0, value, true)
  return [...out]
}
const pool = (names) => [
  ...u32(names.length),
  ...names.flatMap((name, id) => [
    ...u32(id),
    ...u32(enc.encode(name).length),
    ...enc.encode(name),
  ]),
]
const poolsBlock = (first, sur, common) => [
  ...NAME_POOL_SIGNATURE,
  ...pool(first),
  ...pool(sur),
  ...pool(common),
]

describe('NamePools', () => {
  it('reads the three pools after the signature', () => {
    const db = new Uint8Array([9, 9, 9, ...poolsBlock(['Ana', 'Bo'], ['Silva'], ['Pelé']), 7])
    const pools = NamePools.locate(db)
    expect(pools.firstNames.nameAt(db, 1)).toBe('Bo')
    expect(pools.surnames.nameAt(db, 0)).toBe('Silva')
    expect(pools.commonNames.nameAt(db, 0)).toBe('Pelé')
    expect(pools.firstNames.nameAt(db, 2)).toBeUndefined()
    expect(pools.firstNames.nameAt(db, 0xffffffff)).toBeUndefined()
    expect(pools.endOffset).toBe(db.length - 1)
    expect(pools.resolveCommon(db, 1)).toBe('Bo') // falls back to first names
  })

  it('refuses two valid signatures', () => {
    const block = poolsBlock(['A'], ['B'], ['C'])
    expect(() => NamePools.locate(new Uint8Array([...block, ...block]))).toThrow(/ambiguous/)
  })

  it('reports when no pools exist', () => {
    expect(() => NamePools.locate(new Uint8Array(100))).toThrow(/not found/)
  })

  it('rejects a pool whose ids are out of order', () => {
    const bad = [...NAME_POOL_SIGNATURE, ...u32(1), ...u32(5), ...u32(1), 65]
    expect(() => NamePools.locate(new Uint8Array(bad))).toThrow(/has id 5/)
  })
})

describe('layout helpers', () => {
  it('decodes day-of-year dates', () => {
    // day 32 of 2036 = 1 Feb; high bits of the packed word are slot flags and ignored
    const bytes = new Uint8Array([32, 0x02, ...u32(2036).slice(0, 2)])
    expect(decodeDate(bytes, 0)).toBe('2036-02-01')
    expect(decodeDate(new Uint8Array([0, 0, 0xec, 0x07]), 0)).toBeUndefined() // day 0
    // 0x076c = 1900 is the save's null date and falls outside 1901..2200
    expect(decodeDate(new Uint8Array([1, 0, 0x6c, 0x07]), 0)).toBeUndefined()
  })

  it('computes age on the clock date', () => {
    expect(ageAt('2010-07-02', '2036-07-01')).toBe(25)
    expect(ageAt('2010-07-01', '2036-07-01')).toBe(26)
  })

  it('scales raw attributes to 1-20', () => {
    expect(scaleAttribute(0)).toBe(1)
    expect(scaleAttribute(100)).toBe(20)
    expect(scaleAttribute(73)).toBe(15)
  })

  it('finds a build string in the summary', () => {
    const text = enc.encode('27.0.0+2493426')
    const bytes = new Uint8Array([1, 2, 3, ...u32(text.length), ...text, 0])
    expect(findVersion(bytes)).toBe('27.0.0+2493426')
    expect(findVersion(new Uint8Array(40))).toBeUndefined()
  })
})
