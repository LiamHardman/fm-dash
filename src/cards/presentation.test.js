import { describe, expect, it } from 'vitest'
import { mapPlayerCard } from '../utils/playerCardMapper'
import { renderPlayerCard } from './forged-crest.mjs'

describe('card presentation regressions', () => {
  it('anchors the rating and primary position in the same centred block', () => {
    const svg = renderPlayerCard({ player: { rating: 83, primaryPosition: 'CM' } })
    const scene = new DOMParser().parseFromString(svg, 'image/svg+xml')
    const block = scene.querySelector('[aria-label="Rating and primary position"]')
    expect(block.getAttribute('style')).toContain('letter-spacing:0')
    const text = [...block.querySelectorAll('text')]
    expect(text.map((node) => node.getAttribute('x'))).toEqual(['208', '208'])
    expect(text.map((node) => node.getAttribute('text-anchor'))).toEqual(['middle', 'middle'])
  })
  it('centres cutout and photo portraits on the standard card face', () => {
    for (const portraitMode of ['cutout', 'photo']) {
      const svg = renderPlayerCard({
        player: { name: 'Sander Maas', rating: 83, primaryPosition: 'CM' },
        portrait: '/face.png',
        portraitMode,
      })
      const scene = new DOMParser().parseFromString(svg, 'image/svg+xml')
      const portrait = scene.querySelector('g[clip-path] image')
      expect(Number(portrait.getAttribute('x')) + Number(portrait.getAttribute('width')) / 2).toBe(
        512
      )
      expect(portrait.getAttribute('preserveAspectRatio')).toContain('meet')
    }
  })

  it('keeps identity marks below the stats and inside the content panel', () => {
    for (const edition of ['gold', 'icon']) {
      const svg = renderPlayerCard({
        edition,
        player: { nation: { src: '/flag.png' }, club: { src: '/club.png' } },
      })
      const scene = new DOMParser().parseFromString(svg, 'image/svg+xml')
      for (const image of scene.querySelectorAll('g[role="img"] image')) {
        const y = Number(image.getAttribute('y'))
        expect(y).toBeGreaterThanOrEqual(1190)
        expect(y + Number(image.getAttribute('height'))).toBeLessThanOrEqual(1335)
      }
    }
  })
  it('leaves the Icon artwork transparent instead of adding a grey portrait panel', () => {
    const svg = renderPlayerCard({ player: { name: 'Player' }, edition: 'icon' })
    expect(svg).not.toContain('fill="#d7d2c4"')
  })

  it('shows supporting labels and values large enough for a 240px card', () => {
    const svg = renderPlayerCard({
      player: { name: 'Player', age: 21, transferValue: '£24.5M', weeklyWage: '£38K' },
      showSupportingInfo: true,
    })
    const scene = new DOMParser().parseFromString(svg, 'image/svg+xml')
    const footer = scene.querySelector('[aria-label="Supporting information"]')
    for (const text of footer.querySelectorAll('text')) {
      expect((Number(text.getAttribute('font-size')) * 240) / 1100).toBeGreaterThanOrEqual(10)
    }
  })

  it('includes actual flag and club images in the gallery example', async () => {
    const { CARD_STUDIO_PLAYER } = await import('../utils/cardStudioPlayer')
    const card = mapPlayerCard(CARD_STUDIO_PLAYER, {
      clubImageUrl: CARD_STUDIO_PLAYER.clubImageUrl,
    })
    expect(card.nation.src).toBeTruthy()
    expect(card.club.src).toBeTruthy()
  })
})
