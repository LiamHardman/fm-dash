import { describe, expect, it } from 'vitest'
import { FAMILY_PRESENTATIONS, resolveEdition } from '../cards/editions.mjs'
import { renderPlayerCard } from '../cards/forged-crest.mjs'
import { applyTierBoost, CARD_TIERS, statKeysFor } from './cardGeneratorTiers'
import {
  cardPositions,
  formatCardCurrency,
  mapPlayerCard,
  selectPlayerCardEdition,
} from './playerCardMapper'

const player = {
  name: 'Tomás de Albuquerque',
  Overall: 87,
  position: 'AM (RLC), M (RL)',
  pac: 91,
  sho: 83,
  pas: 88,
  dri: 92,
  def: 0,
  phy: 74,
  age: 21,
  transferValueAmount: 24500000,
  wageAmount: 38000,
}

describe('FMDash card mapping', () => {
  it('preserves actual values, zero, full names and supplied position order', () => {
    expect(mapPlayerCard(player)).toMatchObject({
      name: player.name,
      rating: 87,
      primaryPosition: 'RW',
      alternatePositions: ['LW', 'CAM', 'RM', 'LM'],
      attributes: { PAC: 91, SHO: 83, PAS: 88, DRI: 92, DEF: 0, PHY: 74 },
      transferValue: '£24.5M',
      weeklyWage: '£38K',
    })
    expect(cardPositions('D/WB (RL), DM, M (C)')).toEqual(['RB', 'LB', 'RWB', 'LWB', 'CDM', 'CM'])
  })

  it('uses role ratings and pitch overrides without losing natural alternates', () => {
    const mapped = mapPlayerCard(
      { ...player, roleSpecificOveralls: { 'AMC - Attacking Midfielder': 0, 'AMR - Winger': 92 } },
      { selectedRole: 'AMC', positionOverride: 'LW' }
    )
    expect(mapped).toMatchObject({
      rating: 0,
      primaryPosition: 'LW',
      alternatePositions: ['RW', 'CAM', 'RM', 'LM'],
    })
  })

  it('recognises goalkeepers without a position string and keeps all six keys attached', () => {
    const keeper = {
      short_positions: ['GK'],
      Overall: 0,
      div: 99,
      han: 0,
      kic: 31,
      ref: 82,
      spd: 44,
      pos: 73,
    }
    expect(mapPlayerCard(keeper)).toMatchObject({
      goalkeeper: true,
      primaryPosition: 'GK',
      rating: 0,
      attributes: { DIV: 99, HAN: 0, KIC: 31, REF: 82, SPD: 44, POS: 73 },
    })
    expect(statKeysFor(keeper)).toEqual(['div', 'han', 'kic', 'ref', 'spd', 'pos'])
  })

  it('does not invent missing stats, ratings, identities or Playstyle+ eligibility', () => {
    expect(mapPlayerCard({})).toMatchObject({
      rating: undefined,
      nation: undefined,
      club: undefined,
      playstylePlus: undefined,
    })
    expect(mapPlayerCard({ ...player, playstylePlus: 'unknown' }).playstylePlus).toBeUndefined()
    expect(mapPlayerCard(player, { playstylePlus: 'technical' }).playstylePlus).toBe('technical')
    expect(
      mapPlayerCard({ ...player, playstylePlus: 'technical' }, { playstylePlus: '' }).playstylePlus
    ).toBeUndefined()
  })

  it('preserves abbreviated currencies and formats zero and ranges', () => {
    expect(formatCardCurrency('£127M - £381M')).toBe('£381M')
    expect(formatCardCurrency('€38K p/w')).toBe('€38K')
    expect(formatCardCurrency(0)).toBe('£0')
    expect(formatCardCurrency('N/A')).toBeUndefined()
  })

  it('keeps host metal thresholds and rarity but allows explicit generator choices', () => {
    for (const [rating, baseEdition] of [
      [64, 'bronze'],
      [65, 'silver'],
      [75, 'gold'],
    ]) {
      const p = { ...player, Overall: rating }
      expect(selectPlayerCardEdition(p, mapPlayerCard(p))).toMatchObject({
        edition: 'squad-player',
        baseEdition,
        rare: true,
      })
    }
    expect(
      selectPlayerCardEdition(player, mapPlayerCard(player), { edition: 'bronze', rare: false })
    ).toMatchObject({ edition: 'bronze', rare: false })
    expect(
      selectPlayerCardEdition(player, mapPlayerCard(player), { edition: 'icon', rare: true })
    ).toMatchObject({ edition: 'icon', rare: false })
    expect(selectPlayerCardEdition({}, mapPlayerCard({})).rare).toBe(false)
  })

  it('renders every presentation with real mapped data and correct stat order', () => {
    for (const presentation of FAMILY_PRESENTATIONS) {
      const svg = renderPlayerCard({
        ...presentation,
        player: mapPlayerCard(player),
        idPrefix: 'test-player',
        showSupportingInfo: true,
      })
      expect(svg).toContain(resolveEdition(presentation).artwork)
      expect(svg).toContain('ALBUQUERQUE')
      expect(svg).toContain('£24.5M')
      expect(svg.indexOf('>PAC<')).toBeLessThan(svg.indexOf('>SHO<'))
      expect(svg.indexOf('>SHO<')).toBeLessThan(svg.indexOf('>PAS<'))
    }
  })

  it('keeps base stats unchanged and applies special boosts independently of the artwork', () => {
    for (const tier of CARD_TIERS.filter((tier) => tier.boost === 0)) {
      expect(applyTierBoost(player, tier)).toEqual({
        overall: 87,
        pac: 91,
        sho: 83,
        pas: 88,
        dri: 92,
        def: 0,
        phy: 74,
      })
    }
    expect(
      applyTierBoost(
        player,
        CARD_TIERS.find((tier) => tier.key === 'icon')
      ).overall
    ).toBe(97)
  })
})
