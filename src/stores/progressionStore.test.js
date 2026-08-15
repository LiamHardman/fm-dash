import { createPinia, setActivePinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('../services/progressionService.js', () => ({
  default: {
    analyze: vi.fn(),
  },
}))

describe('progressionStore', () => {
  let store
  let progressionService

  beforeEach(async () => {
    setActivePinia(createPinia())
    const { useProgressionStore } = await import('./progressionStore.js')
    progressionService = (await import('../services/progressionService.js')).default
    store = useProgressionStore()
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  it('starts idle with no slots', () => {
    expect(store.status).toBe('idle')
    expect(store.slots).toEqual([])
  })

  it('tracks slot lifecycle via addSlot/updateSlot/removeSlot', () => {
    store.addSlot({ id: 's1', filename: 'a.csv', status: 'uploading', datasetId: null })
    expect(store.slots).toHaveLength(1)

    store.updateSlot('s1', { status: 'parsed', datasetId: 'ds-1' })
    expect(store.slots[0].status).toBe('parsed')
    expect(store.slots[0].datasetId).toBe('ds-1')

    store.removeSlot('s1')
    expect(store.slots).toHaveLength(0)
  })

  it('does not call analyze with fewer than 2 parsed slots', async () => {
    store.addSlot({ id: 's1', filename: 'a.csv', status: 'parsed', datasetId: 'ds-1' })
    await store.analyze()
    expect(progressionService.analyze).not.toHaveBeenCalled()
    expect(store.status).toBe('idle')
  })

  it('transitions idle -> analyzing -> analyzed on success', async () => {
    store.addSlot({ id: 's1', filename: 'a.csv', status: 'parsed', datasetId: 'ds-1' })
    store.addSlot({ id: 's2', filename: 'b.csv', status: 'parsed', datasetId: 'ds-2' })

    progressionService.analyze.mockResolvedValue({
      order: ['ds-1', 'ds-2'],
      players: [{ uid: 1, name: 'Alice', snapshots: [] }],
      currencySymbol: '£',
    })

    const promise = store.analyze()
    expect(store.status).toBe('analyzing')
    await promise

    expect(store.status).toBe('analyzed')
    expect(store.players).toHaveLength(1)
    expect(store.order).toEqual(['ds-1', 'ds-2'])
    expect(store.currencySymbol).toBe('£')
  })

  it('transitions to ambiguous-order when the backend flags a tie', async () => {
    store.addSlot({ id: 's1', filename: 'a.csv', status: 'parsed', datasetId: 'ds-1' })
    store.addSlot({ id: 's2', filename: 'b.csv', status: 'parsed', datasetId: 'ds-2' })

    progressionService.analyze.mockResolvedValue({
      orderAmbiguous: true,
      ambiguousDatasetIds: ['ds-1', 'ds-2'],
    })

    await store.analyze()

    expect(store.status).toBe('ambiguous-order')
    expect(store.ambiguousDatasetIds).toEqual(['ds-1', 'ds-2'])
  })

  it('confirmOrder re-analyzes with an explicit order and reaches analyzed', async () => {
    store.addSlot({ id: 's1', filename: 'a.csv', status: 'parsed', datasetId: 'ds-1' })
    store.addSlot({ id: 's2', filename: 'b.csv', status: 'parsed', datasetId: 'ds-2' })

    progressionService.analyze.mockResolvedValue({
      order: ['ds-2', 'ds-1'],
      players: [],
      currencySymbol: '£',
    })

    await store.confirmOrder(['ds-2', 'ds-1'])

    expect(progressionService.analyze).toHaveBeenCalledWith(
      ['ds-1', 'ds-2'],
      expect.objectContaining({ order: ['ds-2', 'ds-1'] })
    )
    expect(store.status).toBe('analyzed')
    expect(store.order).toEqual(['ds-2', 'ds-1'])
  })

  it('transitions to error status when the service call fails', async () => {
    store.addSlot({ id: 's1', filename: 'a.csv', status: 'parsed', datasetId: 'ds-1' })
    store.addSlot({ id: 's2', filename: 'b.csv', status: 'parsed', datasetId: 'ds-2' })

    progressionService.analyze.mockRejectedValue(new Error('boom'))

    await store.analyze()

    expect(store.status).toBe('error')
    expect(store.errorMessage).toBe('boom')
  })

  it('reset clears all state back to idle', async () => {
    store.addSlot({ id: 's1', filename: 'a.csv', status: 'parsed', datasetId: 'ds-1' })
    store.addSlot({ id: 's2', filename: 'b.csv', status: 'parsed', datasetId: 'ds-2' })
    progressionService.analyze.mockResolvedValue({ order: [], players: [], currencySymbol: '£' })
    await store.analyze()

    store.reset()

    expect(store.status).toBe('idle')
    expect(store.slots).toEqual([])
    expect(store.players).toEqual([])
    expect(store.order).toEqual([])
  })
})
