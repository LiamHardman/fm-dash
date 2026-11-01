import { describe, expect, it } from 'vitest'
import { renderPlayerCard } from '../cards/forged-crest.mjs'
import { mapPlayerCard } from './playerCardMapper'

const mapped = (numericAttributes, position = 'MC', options) =>
  mapPlayerCard({ numericAttributes, position }, options)

describe('FM Playstyle+ selection', () => {
  it.each([
    ['acrobatic', { Agi: 17, Tec: 17, Fin: 17 }],
    ['chip-shot', { Fin: 18, Ant: 15, Vis: 15 }],
    ['dead-ball', { Fre: 17 }],
    ['finesse-shot', { Fin: 17, Tec: 17, Lon: 15 }],
    ['game-changer', { Fin: 17, Fla: 17 }],
    ['low-driven-shot', { Fin: 18, Tec: 18, Lon: 14 }],
    ['power-shot', { Fin: 17, Lon: 17 }],
    ['precision-header', { Hea: 16, Jum: 16 }],
    ['incisive-pass', { Pas: 18 }],
    ['inventive', { Pas: 15, Tec: 15, Fla: 15 }],
    ['long-ball-pass', { Pas: 16, Vis: 16 }],
    ['pinged-pass', { Pas: 16, Tec: 16 }],
    ['tiki-taka', { OtB: 15, Cmp: 15, Pas: 15 }],
    ['whipped-pass', { Cro: 17 }],
    ['first-touch', { Fir: 17 }],
    ['press-proven', { Dri: 15, Agi: 15, Pas: 15, Bal: 15 }],
    ['rapid', { Pac: 18 }],
    ['technical', { Agi: 17, Dri: 17 }],
    ['trickster', { Dri: 17, Fla: 17 }],
    ['aerial-fortress', { Jum: 16, Str: 16 }],
    ['anticipate', { Cmp: 17, Tck: 17 }],
    ['block', { Ant: 16, Pos: 16, Mar: 16 }],
    ['intercept', { Mar: 17, Pos: 17 }],
    ['jockey', { Mar: 17 }],
    ['slide-tackle', { Tck: 17 }],
    ['bruiser', { Str: 17 }, 'D (C)'],
    ['enforcer', { Str: 17 }, 'ST (C)'],
    ['long-throw', { 'L Th': 16 }],
    ['quick-step', { Acc: 18 }],
    ['relentless', { Wor: 17, Sta: 17 }],
    ['rush-out', { TRO: 16, '1v1': 16 }, 'GK'],
    ['cross-claimer', { Pun: 17 }, 'GK'],
    ['deflector', { Aer: 14, Han: 16, Ref: 14 }, 'GK'],
    ['far-reach', { Aer: 17, Jum: 14 }, 'GK'],
    ['far-throw', { Thr: 16 }, 'GK'],
    ['footwork', { '1v1': 15, Ref: 15 }, 'GK'],
    ['pinged-pass', { Pas: 14, Tec: 14 }, 'GK'],
    ['long-ball-pass', { Pas: 14, Vis: 14 }, 'GK'],
  ])('selects %s at its threshold (%j, %s)', (id, attributes, position) => {
    expect(mapped(attributes, position).playstylePlus).toBe(id)
    const below = Object.fromEntries(
      Object.entries(attributes).map(([key, value]) => [key, value - 1])
    )
    expect(mapped(below, position).playstylePlus).not.toBe(id)
  })

  it('compares averages without rounding and enforces individual minima and maxima', () => {
    expect(mapped({ Agi: 20, Tec: 16, Fin: 15 }).playstylePlus).toBe('acrobatic')
    expect(mapped({ Agi: 20, Tec: 16, Fin: 14 }).playstylePlus).not.toBe('acrobatic')
    expect(mapped({ Fin: 20, Tec: 16, Lon: 15 }).playstylePlus).not.toBe('finesse-shot')
    expect(mapped({ Fin: 20, Tec: 17, Lon: 14 }).playstylePlus).not.toBe('low-driven-shot')
    expect(mapped({ Fin: 18, Tec: 18 }).playstylePlus).not.toBe('low-driven-shot')
    expect(mapped({ Wor: 20, Sta: 14 }).playstylePlus).toBeUndefined()
  })

  it('ranks by required totals, keeps ties stable, and always puts Long Throw last', () => {
    const attributes = { Agi: 17, Tec: 17, Fin: 17, Dri: 15, Pas: 15, Bal: 13, Pac: 20 }
    expect(mapped(attributes).playstylePlus).toBe('press-proven') // 60 beats 51 and 18
    expect(mapped({ Pas: 16, Vis: 16, Tec: 16 }).playstylePlus).toBe('long-ball-pass')
    expect(mapped({ Fre: 17, 'L Th': 20 }).playstylePlus).toBe('dead-ball')
  })

  it('restricts defender, striker and goalkeeper rules to natural positions', () => {
    expect(mapped({ Str: 17 }).playstylePlus).toBeUndefined()
    expect(mapped({ Str: 17 }, 'AM (RLC)').playstylePlus).toBeUndefined()
    expect(mapped({ Str: 17 }, 'D/WB (RL)').playstylePlus).toBe('bruiser')
    expect(mapped({ Str: 17 }, 'MC', { positionOverride: 'ST' }).playstylePlus).toBeUndefined()
    expect(mapped({ Thr: 20 }).playstylePlus).toBeUndefined()
    expect(mapped({ Pac: 20, Fin: 20, Lon: 20 }, 'GK').playstylePlus).toBeUndefined()
    expect(mapped({ Pas: 14, Tec: 14 }).playstylePlus).toBeUndefined()
    expect(
      mapPlayerCard({ short_positions: ['GK'], numericAttributes: { Thr: 16 } }).playstylePlus
    ).toBe('far-throw')
  })

  it('uses real FM maps, rejects missing or invalid attributes, and never reads summary stats', () => {
    expect(mapPlayerCard({ attributes: { Fre: '17' } }).playstylePlus).toBe('dead-ball')
    expect(mapPlayerCard({ numeric_attributes: { Fre: 17 } }).playstylePlus).toBe('dead-ball')
    expect(mapPlayerCard({ attributes: { Fre: 17 }, numericAttributes: {} }).playstylePlus).toBe(
      'dead-ball'
    )
    expect(mapPlayerCard({ pac: 18, pas: 18, dri: 18, Fre: 17 }).playstylePlus).toBeUndefined()
    for (const Fre of [undefined, null, '', ' ', '-', '17-20', true, 0, 21, NaN, Infinity]) {
      expect(mapped({ Fre }).playstylePlus).toBeUndefined()
    }
    expect(mapped({ Agi: 20, Dri: undefined }).playstylePlus).toBeUndefined()
  })

  it('honours explicit assignments and None ahead of automatic eligibility', () => {
    const player = { numericAttributes: { Fre: 20 }, playstyle_plus: 'technical' }
    expect(mapPlayerCard(player).playstylePlus).toBe('technical')
    expect(mapPlayerCard(player, { playstylePlus: 'rapid' }).playstylePlus).toBe('rapid')
    expect(mapPlayerCard(player, { playstylePlus: '' }).playstylePlus).toBeUndefined()
  })

  it('renders the selected badge with the specified Playstyle+ spelling and keeper label', () => {
    const svg = renderPlayerCard({ player: mapped({ TRO: 16, '1v1': 16 }, 'GK') })
    expect(svg).toContain('data-playstyle-plus="rush-out"')
    expect(svg).toContain('1v1 Close Down Playstyle+')
    expect(svg).not.toContain('PlayStyle+')
  })
})
