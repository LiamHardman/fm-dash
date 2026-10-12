import { flushPromises } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import { effectScope, nextTick, ref, shallowRef } from 'vue'
import { positionRank, sortPlayerIndices } from '../utils/playerSorting'

it('applies shared sides to Football Manager combined positions', () => {
  expect(positionRank('D/WB (R)')).toBe(positionRank('DR'))
  expect(positionRank('M/AM (LC)')).toBe(positionRank('MC'))
  expect(positionRank('ST (C)')).toBe(positionRank('ST'))
})

import { usePlayerSorting } from './usePlayerSorting'

describe('latest player sorting', () => {
  it('sends the Dataset once and drains the latest sort after an in-flight sort', async () => {
    const messages = []
    const worker = { postMessage: (m) => messages.push(m), terminate() {} }
    const source = shallowRef(
      Array.from({ length: 2000 }, (_, i) => ({ uid: i, age: i, attributes: { Fin: 15 } }))
    )
    const spec = ref({ field: 'age', direction: 'asc', goalkeeper: false })
    const scope = effectScope()
    const state = scope.run(() => usePlayerSorting(source, spec, 1000, () => worker))
    expect(messages.filter((m) => m.type === 'load')).toHaveLength(1)
    expect(messages[0].rows[0].attributes).toBeUndefined()
    spec.value = { ...spec.value, direction: 'desc' }
    await nextTick()
    worker.onmessage({ data: { id: messages[1].id, indices: [0, 1] } })
    await flushPromises()
    const last = messages.at(-1)
    expect(last.spec.direction).toBe('desc')
    worker.onmessage({ data: { id: last.id, indices: [1999, 1998] } })
    await flushPromises()
    expect(state.sortedPlayers.value[0].uid).toBe(1999)
    expect(messages.filter((m) => m.type === 'load')).toHaveLength(1)
    scope.stop()
  })

  it('never maps a new Dataset result onto the previous Dataset objects', async () => {
    const messages = []
    const worker = { postMessage: (m) => messages.push(m), terminate() {} }
    const source = shallowRef(Array.from({ length: 2000 }, (_, i) => ({ uid: i, name: 'Old' })))
    const scope = effectScope()
    const state = scope.run(() =>
      usePlayerSorting(source, { field: 'uid', direction: 'asc' }, 1000, () => worker)
    )
    source.value = source.value.map((p) => ({ ...p, name: 'New' }))
    await nextTick()
    worker.onmessage({ data: { id: messages[1].id, indices: [0] } })
    await flushPromises()
    worker.onmessage({ data: { id: messages.at(-1).id, indices: [1] } })
    await flushPromises()
    expect(state.sortedPlayers.value[0]).toEqual({ uid: 1, name: 'New' })
    scope.stop()
  })

  it('uses the same goalkeeper comparator in the main-thread fallback', () => {
    const source = ref([
      { uid: 1, pac: 90, div: 20 },
      { uid: 2, pac: 20, div: 90 },
    ])
    const spec = { field: 'pac', direction: 'desc', goalkeeper: true }
    const scope = effectScope()
    const state = scope.run(() => usePlayerSorting(source, spec))
    expect(state.sortedPlayers.value.map((p) => p.uid)).toEqual(
      sortPlayerIndices(source.value, spec).map((i) => source.value[i].uid)
    )
    expect(state.sortedPlayers.value[0].uid).toBe(2)
    scope.stop()
  })
})
