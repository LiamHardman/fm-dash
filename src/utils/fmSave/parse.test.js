import { describe, expect, it } from 'vitest'
import { NAME_POOL_SIGNATURE } from './layout.js'
import { parseGameDb } from './parse.js'
import { closingPersonHeader } from './playerScan.js'

const enc = new TextEncoder()

class Writer {
  constructor(size) {
    this.bytes = new Uint8Array(size)
    this.view = new DataView(this.bytes.buffer)
  }
  u8(at, v) {
    this.bytes[at] = v
  }
  u16(at, v) {
    this.view.setUint16(at, v, true)
  }
  i16(at, v) {
    this.view.setInt16(at, v, true)
  }
  u32(at, v) {
    this.view.setUint32(at, v, true)
  }
  put(at, arr) {
    this.bytes.set(arr, at)
  }
}

function pools(first, sur, common) {
  const out = [...NAME_POOL_SIGNATURE]
  const u32 = (v) => [v & 255, (v >> 8) & 255, (v >> 16) & 255, v >>> 24]
  for (const list of [first, sur, common]) {
    out.push(...u32(list.length))
    list.forEach((name, id) =>
      out.push(...u32(id), ...u32(enc.encode(name).length), ...enc.encode(name))
    )
  }
  return out
}

/** One synthetic FM27 player record (+ person block + closing header) after the name pools. */
function buildGameDb({ commonName = 0xffffffff, closing = true } = {}) {
  const poolBytes = pools(['Ana', 'Bruno'], ['Silva', 'Costa'], ['Brunão'])
  const w = new Writer(poolBytes.length + 64 + 400)
  w.put(0, poolBytes)
  const R = poolBytes.length + 64
  const pindex = 41
  // record header (negative offsets)
  w.u32(R - 19, pindex)
  w.u32(R - 15, 500) // uid
  w.u32(R - 11, 500) // uid copy
  w.u8(R - 7, 3) // object kind
  // record body
  w.u16(R + 0, 150) // CA
  w.i16(R + 2, 170) // PA
  w.u8(R + 8, 10) // reputation bucket
  w.u32(R + 16, 0xffffffff) // no team
  for (let i = 0; i < 15; i++) w.u8(R + 24 + i, 1)
  w.u8(R + 24 + 3, 20) // DC
  w.u8(R + 24 + 5, 16) // DM
  for (let i = 0; i < 54; i++) w.u8(R + 39 + i, 50) // raw 50 -> 10
  w.u8(R + 39 + 0, 100) // crossing -> 20
  w.u8(R + 39 + 24, 100) // left foot -> 20
  w.u8(R + 39 + 25, 30) // right foot -> 6
  w.u32(R + 93, 1_234_567) // transfer value
  w.put(R + 102, [0x01, 0x00, 0x6c, 0x07]) // marker
  w.u8(R + 121, 188) // height
  // person block: legal name length 0 -> birth = blockStart + 19
  const birth = R + 160
  const block = birth - 19
  w.u32(block, 1) // first name "Bruno"
  w.u32(block + 5, 1) // surname "Costa"
  w.u32(block + 10, commonName)
  w.u32(birth, (2010 << 16) | 32) // 1 Feb 2010
  w.u16(birth + 14, 7) // nation id
  w.put(birth + 22, [11, 12, 13, 14, 15, 16, 17, 18]) // personality
  // closing person header: pindex + 1, then the unique id twice
  if (closing) {
    w.u32(R + 220, pindex + 1)
    w.u32(R + 224, 2_002_330_039)
    w.u32(R + 228, 2_002_330_039)
  }
  return { gameDb: w.bytes, recordOffset: R }
}

describe('parseGameDb', () => {
  it('decodes a synthetic player record end to end', () => {
    const { gameDb } = buildGameDb()
    const { players, report } = parseGameDb(gameDb, '2036-07-01')
    expect(report).toMatchObject({ players: 1, markerless: 0, withPerson: 1 })
    const [p] = players
    expect(p).toMatchObject({
      uid: 500,
      uniqueId: 2_002_330_039,
      name: 'Bruno Costa',
      birthDate: '2010-02-01',
      age: 26,
      nationId: 7,
      ca: 150,
      pa: 170,
      heightCm: 188,
      transferValue: 1_234_567,
      leftFoot: 20,
      rightFoot: 6,
    })
    expect(p.attributes.crossing).toBe(20)
    expect(p.attributes.flair).toBe(10) // first attribute stored after the feet
    expect(p.attributes.concentration).toBe(10)
    expect(p.positions).toMatchObject({ DC: 20, DM: 16, GK: 1 })
    expect(p.personality).toMatchObject({ adaptability: 11, controversy: 18 })
  })

  it('prefers the common name', () => {
    const { gameDb } = buildGameDb({ commonName: 0 })
    expect(parseGameDb(gameDb, '2036-07-01').players[0].name).toBe('Brunão')
  })

  it('leaves uniqueId empty without a closing header', () => {
    const { gameDb } = buildGameDb({ closing: false })
    const [p] = parseGameDb(gameDb, '2036-07-01').players
    expect(p.uniqueId).toBeUndefined()
    expect(p.name).toBe('Bruno Costa')
  })

  it('fails when the save holds no player records', () => {
    expect(() => parseGameDb(new Uint8Array(pools(['A'], ['B'], ['C'])), '2036-07-01')).toThrow(
      /no player records/
    )
  })
})

describe('closingPersonHeader', () => {
  it('only searches inside its window', () => {
    const { gameDb, recordOffset } = buildGameDb()
    const end = recordOffset + 232
    expect(closingPersonHeader(gameDb, 41, recordOffset + 122, end, 500)).toEqual([
      recordOffset + 220,
      2_002_330_039,
    ])
    expect(
      closingPersonHeader(gameDb, 41, recordOffset + 122, recordOffset + 223, 500)
    ).toBeUndefined()
  })
})
