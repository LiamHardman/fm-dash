import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { Quasar } from 'quasar'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import colourSetService from '../services/colourSetService'
import { useColourSetStore } from '../stores/colourSetStore'
import { activeColourBands, defaultBands } from '../utils/colourSets'
import ColourSetEditor from './ColourSetEditor.vue'
import StatGauge from './player-table/StatGauge.vue'

vi.mock('../services/colourSetService', () => ({ default: { load: vi.fn(), save: vi.fn() } }))

let wrappers = []
beforeEach(() => {
  vi.resetAllMocks()
  setActivePinia(createPinia())
  activeColourBands.fm = defaultBands('fm')
  activeColourBands.fifa = defaultBands('fifa')
  colourSetService.load.mockResolvedValue({ version: 1, activeSetId: '', sets: [] })
  colourSetService.save.mockImplementation(async (_scale, document) =>
    JSON.parse(JSON.stringify(document))
  )
})
afterEach(() => {
  for (const wrapper of wrappers) wrapper.unmount()
  wrappers = []
})

async function editor(scale) {
  const wrapper = mount(ColourSetEditor, {
    props: { scale },
    global: {
      plugins: [Quasar],
      stubs: { QBtn: false, QInput: false, QIcon: false, QSelect: false },
    },
  })
  wrappers.push(wrapper)
  await flushPromises()
  return wrapper
}
function button(wrapper, label) {
  const found = wrapper.findAll('button').find((node) => node.text().includes(label))
  if (!found) throw new Error(`No button ${label}`)
  return found
}

describe('colour set editing', () => {
  it.each(['fm', 'fifa'])(
    'edits %s down to two and up to twenty bands, then saves and reloads',
    async (scale) => {
      const wrapper = await editor(scale)
      while (wrapper.findAll('.colour-editor__band').length > 2) {
        await wrapper.findAll('button[aria-label^="Remove band"]').at(-1).trigger('click')
      }
      await button(wrapper, 'Save and apply').trigger('click')
      await flushPromises()
      expect(colourSetService.save.mock.lastCall[1].sets[0].bands).toHaveLength(2)
      for (let i = 2; i < 20; i++) await button(wrapper, 'Add band').trigger('click')
      expect(button(wrapper, 'Add band').attributes('disabled')).toBeDefined()
      await wrapper.find('input[type="color"]').setValue('#123456')
      await button(wrapper, 'Save and apply').trigger('click')
      await flushPromises()
      const saved = colourSetService.save.mock.lastCall[1]
      expect(saved.sets).toHaveLength(1)
      expect(saved.sets[0].bands).toHaveLength(20)
      expect(saved.sets[0].bands[0].color).toBe('#123456')
      expect(activeColourBands[scale]).toEqual(saved.sets[0].bands)
      expect(wrapper.text()).toContain('saved and applied')

      wrapper.unmount()
      setActivePinia(createPinia())
      colourSetService.load.mockResolvedValue(saved)
      const restored = await editor(scale)
      expect(restored.findAll('.colour-editor__band')).toHaveLength(20)
      expect(restored.find('input[type="color"]').element.value).toBe('#123456')
    }
  )

  it('keeps failed saves as drafts and supports retry, copies, and deletion', async () => {
    const wrapper = await editor('fm')
    const originalFifa = activeColourBands.fifa
    colourSetService.save.mockRejectedValueOnce(new Error('Storage offline'))
    await wrapper.find('input[type="color"]').setValue('#000000')
    await button(wrapper, 'Save and apply').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('have not been saved')
    expect(activeColourBands.fm[0].color).toBe('#ef5350')
    expect(wrapper.find('input[type="color"]').element.value).toBe('#000000')
    await button(wrapper, 'Save and apply').trigger('click')
    await flushPromises()
    expect(activeColourBands.fm[0].color).toBe('#000000')
    await button(wrapper, 'Save as new set').trigger('click')
    await flushPromises()
    expect(useColourSetStore().documents.fm.sets).toHaveLength(2)
    await button(wrapper, 'Delete set').trigger('click')
    await button(wrapper, 'Confirm delete').trigger('click')
    await flushPromises()
    expect(useColourSetStore().documents.fm.sets).toHaveLength(1)
    expect(useColourSetStore().documents.fm.activeSetId).toBe('')
    expect(activeColourBands.fm).toEqual(defaultBands('fm'))
    expect(activeColourBands.fifa).toBe(originalFifa)
  })

  it('requires loading stored sets before editing and retries a failed load', async () => {
    colourSetService.load.mockRejectedValueOnce(new Error('Cannot connect'))
    const wrapper = await editor('fifa')
    expect(button(wrapper, 'Save and apply').attributes('disabled')).toBeDefined()
    await button(wrapper, 'Retry').trigger('click')
    await flushPromises()
    expect(button(wrapper, 'Save and apply').attributes('disabled')).toBeUndefined()
  })

  it('switches saved sets and restores defaults without changing the other scale', async () => {
    const wrapper = await editor('fm')
    await wrapper.find('input[type="color"]').setValue('#123456')
    await button(wrapper, 'Save and apply').trigger('click')
    await flushPromises()
    const firstId = useColourSetStore().documents.fm.activeSetId
    await wrapper.find('input[type="color"]').setValue('#654321')
    await button(wrapper, 'Save as new set').trigger('click')
    await flushPromises()
    const select = wrapper.findComponent({ name: 'QSelect' })
    select.vm.$emit('update:modelValue', firstId)
    await flushPromises()
    await button(wrapper, 'Use set').trigger('click')
    await flushPromises()
    expect(activeColourBands.fm[0].color).toBe('#123456')
    expect(colourSetService.save.mock.lastCall[1].activeSetId).toBe(firstId)
    select.vm.$emit('update:modelValue', '')
    await flushPromises()
    await button(wrapper, 'Use set').trigger('click')
    await flushPromises()
    expect(activeColourBands.fm).toEqual(defaultBands('fm'))
    expect(activeColourBands.fifa).toEqual(defaultBands('fifa'))
    expect(useColourSetStore().documents.fm.sets).toHaveLength(2)
  })

  it('validates edited thresholds before saving', async () => {
    const wrapper = await editor('fm')
    const start = wrapper.findAll('input[type="number"]')[1]
    await start.setValue('1')
    expect(button(wrapper, 'Save and apply').attributes('disabled')).toBeDefined()
    await start.setValue('9')
    expect(button(wrapper, 'Save and apply').attributes('disabled')).toBeUndefined()
    await button(wrapper, 'Save and apply').trigger('click')
    await flushPromises()
    expect(colourSetService.save.mock.lastCall[1].sets[0].bands[1].min).toBe(9)
  })

  it('updates the visible stat gauge when a set is saved', async () => {
    const wrapper = await editor('fifa')
    const gauge = mount(StatGauge, {
      props: { value: 20 },
      global: {
        plugins: [Quasar],
        stubs: { QBtn: false, QInput: false, QIcon: false, QSelect: false },
      },
    })
    wrappers.push(gauge)
    await wrapper.find('input[type="color"]').setValue('#123456')
    await button(wrapper, 'Save and apply').trigger('click')
    await flushPromises()
    expect(gauge.findAll('circle')[1].attributes('stroke')).toBe('#123456')
  })
})
