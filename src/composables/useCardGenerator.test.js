import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, nextTick, ref } from 'vue'
import { useCardGenerator } from './useCardGenerator'

function studio() {
  let state
  const player = ref({
    name: 'Test Player',
    Overall: 70,
    position: 'ST',
    pac: 80,
    sho: 72,
    pas: 61,
    dri: 75,
    def: 21,
    phy: 65,
  })
  const wrapper = mount(
    defineComponent({
      setup() {
        state = useCardGenerator(player)
        return () => null
      },
    })
  )
  return { state, player, wrapper }
}

afterEach(() => vi.restoreAllMocks())

describe('Card Generator integration', () => {
  it('resets edits on tier or player changes while keeping overall independent', async () => {
    const { state, player, wrapper } = studio()
    state.setStat('sho', 99)
    expect(state.editableOverall.value).toBe(70)
    state.selectedTierKey.value = 'icon'
    await nextTick()
    expect(state.editableOverall.value).toBe(85)
    expect(state.editableStats.sho).toBeLessThan(99)
    player.value = {
      Overall: 60,
      shortPositions: ['GK'],
      div: 70,
      han: 65,
      kic: 55,
      ref: 71,
      spd: 30,
      pos: 66,
    }
    await nextTick()
    expect(state.statKeys.value).toContain('div')
    expect(state.editableStats.sho).toBeUndefined()
    expect(state.previewPlayer.value.div).toBeGreaterThanOrEqual(70)
    wrapper.unmount()
  })

  it('clears rare when moving to a special edition', async () => {
    const { state, wrapper } = studio()
    state.rare.value = true
    state.selectedTierKey.value = 'wonderkid'
    await nextTick()
    expect(state.rare.value).toBe(false)
    wrapper.unmount()
  })

  it('releases uploaded image URLs on replacement and unmount', () => {
    const create = vi
      .fn()
      .mockReturnValueOnce('blob:face1')
      .mockReturnValueOnce('blob:face2')
      .mockReturnValueOnce('blob:logo')
    const revoke = vi.fn()
    vi.stubGlobal('URL', { createObjectURL: create, revokeObjectURL: revoke })
    const { state, wrapper } = studio()
    const file = new Blob(['image'])
    state.onFaceFileSelected(file)
    state.onFaceFileSelected(file)
    state.onLogoFileSelected(file)
    expect(revoke).toHaveBeenCalledWith('blob:face1')
    wrapper.unmount()
    expect(revoke).toHaveBeenCalledWith('blob:face2')
    expect(revoke).toHaveBeenCalledWith('blob:logo')
    vi.unstubAllGlobals()
  })

  it('surfaces exporter errors and always clears the loading state', async () => {
    const { state, wrapper } = studio()
    const exporter = vi.fn().mockRejectedValue(new Error('Image server denied export'))
    await state.exportAsPng({ exportPng: exporter }, 'Test Player')
    expect(exporter).toHaveBeenCalledWith(2)
    expect(state.exportError.value).toBe('Image server denied export')
    expect(state.exporting.value).toBe(false)
    wrapper.unmount()
  })
})
