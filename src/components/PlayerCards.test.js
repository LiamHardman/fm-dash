import { mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { h, nextTick } from 'vue'
import { cardPngBlob } from '../cards/export-png.mjs'
import { prepareCardImages } from '../cards/images.mjs'
import { fetchFullPlayerStats } from '../services/playerService'
import { useUiStore } from '../stores/uiStore'
import PlayerCards from './PlayerCards.vue'

vi.mock('../services/playerService', () => ({ fetchFullPlayerStats: vi.fn() }))
vi.mock('../cards/images.mjs', () => ({ prepareCardImages: vi.fn(async (images) => images) }))
vi.mock('../cards/export-png.mjs', () => ({ cardPngBlob: vi.fn(async () => new Blob(['png'])) }))

const player = {
  uid: '1',
  name: 'Test Player',
  Overall: 75,
  position: 'AM (R)',
  pac: 91,
  sho: 72,
  pas: 80,
  dri: 84,
  def: 30,
  phy: 62,
}
async function settle() {
  for (let i = 0; i < 8; i++) await nextTick()
}
function card(props = {}) {
  const pinia = createPinia()
  useUiStore(pinia).showLogos = false
  return mount(PlayerCards, { props: { player, ...props }, global: { plugins: [pinia] } })
}
beforeEach(() => vi.clearAllMocks())

describe('shared player card adapter', () => {
  it('loads missing FM attributes even with a nation flag and exports the assigned badge', async () => {
    fetchFullPlayerStats.mockResolvedValueOnce({
      data: { player: { numericAttributes: { Fre: 17 } } },
    })
    const wrapper = card({
      datasetId: 'dataset-1',
      player: { ...player, nationalityIso: 'pt', numericAttributes: {} },
    })
    await settle()
    expect(fetchFullPlayerStats).toHaveBeenCalledWith('dataset-1', '1')
    expect(wrapper.get('[data-playstyle-plus="dead-ball"]').exists()).toBe(true)
    await wrapper.vm.exportPng(2)
    expect(
      cardPngBlob.mock.calls[0][0].querySelector('[data-playstyle-plus="dead-ball"]')
    ).not.toBeNull()
    await wrapper.setProps({ playstylePlus: '' })
    await settle()
    expect(wrapper.find('[data-playstyle-plus]').exists()).toBe(false)
    wrapper.unmount()
  })

  it('renders reactive host values with unique SVG identifiers per instance', async () => {
    const parent = mount(
      { render: () => h('div', [h(PlayerCards, { player }), h(PlayerCards, { player })]) },
      { global: { plugins: [createPinia()] } }
    )
    const [first, second] = parent.findAllComponents(PlayerCards)
    await settle()
    expect(first.find('svg').attributes('aria-labelledby')).not.toBe(
      second.find('svg').attributes('aria-labelledby')
    )
    expect(first.get('[aria-label="PAC: 91"]').exists()).toBe(true)
    parent.unmount()
    const reactive = card()
    await reactive.setProps({ player: { ...player, Overall: 0, pac: 0, name: '<Updated>' } })
    await settle()
    expect(reactive.get('[aria-label="PAC: 0"]').exists()).toBe(true)
    expect(reactive.get('svg title').text()).toContain('<Updated>, 0,')
    expect(reactive.find('Updated').exists()).toBe(false)
    reactive.unmount()
  })

  it('discards stale Dataset detail responses after the player changes', async () => {
    let resolveOld
    fetchFullPlayerStats
      .mockReturnValueOnce(
        new Promise((resolve) => {
          resolveOld = resolve
        })
      )
      .mockResolvedValueOnce({ data: { player: { nationalityIso: 'fr', nationality: 'France' } } })
    const wrapper = card({ datasetId: 'dataset-1' })
    await wrapper.setProps({ player: { ...player, uid: '2', name: 'Second Player' } })
    await settle()
    resolveOld({ data: { player: { nationalityIso: 'pt', nationality: 'Portugal' } } })
    await settle()
    expect(wrapper.get('svg title').text()).toContain('France')
    expect(wrapper.get('svg title').text()).not.toContain('Portugal')
    expect(prepareCardImages).toHaveBeenLastCalledWith(
      expect.objectContaining({ nation: 'https://flagcdn.com/w80/fr.png' })
    )
    wrapper.unmount()
  })

  it('waits for host details and image readiness before exporting the current scene', async () => {
    let resolveDetail
    fetchFullPlayerStats.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveDetail = resolve
      })
    )
    const wrapper = card({ datasetId: 'dataset-1', selectedRole: 'AMR - Winger' })
    await settle()
    const pending = wrapper.vm.exportPng(2)
    await settle()
    expect(cardPngBlob).not.toHaveBeenCalled()
    resolveDetail({
      data: { player: { roleSpecificOveralls: { 'AMR - Winger': 95 }, nationalityIso: 'pt' } },
    })
    await pending
    expect(cardPngBlob).toHaveBeenCalledWith(expect.any(SVGElement), 2)
    expect(wrapper.get('svg title').text()).toContain(', 95,')
    wrapper.unmount()
  })
})
