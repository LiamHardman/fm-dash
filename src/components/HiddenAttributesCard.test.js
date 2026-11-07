import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import { usePlayerStore } from '../stores/playerStore'
import { useUiStore } from '../stores/uiStore'
import { getUnifiedRatingClass } from '../utils/colourSets'
import HiddenAttributesCard from './HiddenAttributesCard.vue'

// Render Quasar pieces as plain slots so the markup is inspectable.
const stubs = {
  QCard: { template: '<div><slot /></div>' },
  QCardSection: { template: '<div><slot /></div>' },
  QList: { template: '<div><slot /></div>' },
  QItem: { template: '<div><slot /></div>' },
  QItemSection: { template: '<div><slot /></div>' },
  QItemLabel: { template: '<div><slot /></div>' },
  QIcon: true,
}

const savePlayer = {
  name: 'Bruno',
  hiddenAttributes: { Consistency: 17, 'Injury Proneness': 3, Controversy: 18, Ambition: 12 },
}

function mountCard({ toggle, datasetPlayers, player }) {
  const ui = useUiStore()
  ui.showHiddenAttributes = toggle
  usePlayerStore().allPlayers = datasetPlayers
  return mount(HiddenAttributesCard, { props: { player }, global: { stubs } })
}

describe('HiddenAttributesCard', () => {
  beforeEach(() => setActivePinia(createPinia()))

  it('shows the card only with the toggle on and hidden attributes in the Dataset', () => {
    expect(
      mountCard({ toggle: false, datasetPlayers: [savePlayer], player: savePlayer }).text()
    ).toBe('')
    // A CSV Dataset: toggle on, no hidden attributes anywhere.
    const csv = { name: 'C' }
    expect(mountCard({ toggle: true, datasetPlayers: [csv], player: csv }).text()).toBe('')
    const shown = mountCard({ toggle: true, datasetPlayers: [savePlayer], player: savePlayer })
    expect(shown.text()).toContain('Hidden')
    expect(shown.text()).toContain('Personality')
    expect(shown.find('[data-attribute="Consistency"]').text()).toBe('17')
    expect(shown.find('[data-attribute="Loyalty"]').text()).toBe('-')
  })

  it('inverts the colour scale for lower-is-better attributes', () => {
    const card = mountCard({ toggle: true, datasetPlayers: [savePlayer], player: savePlayer })
    const classes = (name) => card.find(`[data-attribute="${name}"]`).attributes('class')
    // Injury Proneness 3 is good: coloured like an 18.
    expect(classes('Injury Proneness')).toContain(getUnifiedRatingClass(18, 20))
    expect(classes('Controversy')).toContain(getUnifiedRatingClass(3, 20))
    expect(classes('Consistency')).toContain(getUnifiedRatingClass(17, 20))
  })

  it('says so for a player without values in a save Dataset', () => {
    const card = mountCard({ toggle: true, datasetPlayers: [savePlayer], player: { name: 'X' } })
    expect(card.text()).toContain('Not in this import')
  })
})
