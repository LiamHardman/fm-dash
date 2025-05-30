import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { usePlayerStore } from './playerStore'

// Mock the player service
vi.mock('../services/playerService', () => ({
  default: {
    uploadPlayerFile: vi.fn(),
    getPlayersByDatasetId: vi.fn()
  }
}))

describe('playerStore', () => {
  let store

  beforeEach(() => {
    setActivePinia(createPinia())
    store = usePlayerStore()

    // Mock sessionStorage
    global.sessionStorage = {
      getItem: vi.fn(),
      setItem: vi.fn(),
      removeItem: vi.fn(),
      clear: vi.fn()
    }
  })

  it('initializes with correct default state', () => {
    expect(store.allPlayers).toEqual([])
    expect(store.currentDatasetId).toBeNull()
    expect(store.detectedCurrencySymbol).toBe('$')
    expect(store.loading).toBe(false)
    expect(store.error).toBe('')
    expect(store.allAvailableRoles).toEqual([])
  })

  it('computes unique clubs correctly', () => {
    store.allPlayers = [
      { club: 'Manchester United' },
      { club: 'Liverpool' },
      { club: 'Manchester United' },
      { club: 'Arsenal' },
      { club: null }
    ]

    const uniqueClubs = store.uniqueClubs
    expect(uniqueClubs).toEqual(['Arsenal', 'Liverpool', 'Manchester United'])
  })

  it('handles empty players array for unique clubs', () => {
    store.allPlayers = []
    expect(store.uniqueClubs).toEqual([])
  })

  it('computes unique nationalities correctly', () => {
    store.allPlayers = [
      { nationality: 'England' },
      { nationality: 'Spain' },
      { nationality: 'England' },
      { nationality: 'France' },
      { nationality: null }
    ]

    const uniqueNationalities = store.uniqueNationalities
    expect(uniqueNationalities).toEqual(['England', 'France', 'Spain'])
  })

  it('computes unique media handlings correctly', () => {
    store.allPlayers = [
      { media_handling: 'Confident, Outspoken' },
      { media_handling: 'Reserved' },
      { media_handling: 'Confident, Reserved' },
      { media_handling: null }
    ]

    const uniqueMediaHandlings = store.uniqueMediaHandlings
    expect(uniqueMediaHandlings).toEqual(['Confident', 'Outspoken', 'Reserved'])
  })

  it('computes unique personalities correctly', () => {
    store.allPlayers = [
      { personality: 'Ambitious' },
      { personality: 'Loyal' },
      { personality: 'Ambitious' },
      { personality: 'Professional' },
      { personality: null }
    ]

    const uniquePersonalities = store.uniquePersonalities
    expect(uniquePersonalities).toEqual(['Ambitious', 'Loyal', 'Professional'])
  })

  it('computes unique positions count correctly', () => {
    store.allPlayers = [
      { parsedPositions: ['ST', 'CF'] },
      { parsedPositions: ['CM', 'CAM'] },
      { parsedPositions: ['ST'] },
      { parsedPositions: ['GK'] },
      { parsedPositions: null }
    ]

    expect(store.uniquePositionsCount).toBe(5) // ST, CF, CM, CAM, GK
  })

  it('computes transfer value range correctly', () => {
    store.allPlayers = [
      { transferValueAmount: 1000000 },
      { transferValueAmount: 5000000 },
      { transferValueAmount: 500000 },
      { transferValueAmount: 10000000 }
    ]

    const range = store.currentDatasetTransferValueRange
    expect(range.min).toBe(500000)
    expect(range.max).toBe(10000000)
  })

  it('handles empty transfer value range', () => {
    store.allPlayers = []
    const range = store.currentDatasetTransferValueRange
    expect(range.min).toBe(0)
    expect(range.max).toBe(100000000)
  })

  it('handles single value in transfer value range', () => {
    store.allPlayers = [{ transferValueAmount: 1000000 }]

    const range = store.currentDatasetTransferValueRange
    expect(range.min).toBe(1000000)
    expect(range.max).toBe(1050000) // min + 50000
  })

  it('handles zero transfer values', () => {
    store.allPlayers = [
      { transferValueAmount: 0 },
      { transferValueAmount: 0 },
      { transferValueAmount: 0 }
    ]

    const range = store.currentDatasetTransferValueRange
    expect(range.min).toBe(0)
    expect(range.max).toBe(50000)
  })

  it('computes salary range correctly', () => {
    store.allPlayers = [
      { wageAmount: 50000 },
      { wageAmount: 100000 },
      { wageAmount: 25000 },
      { wageAmount: 200000 }
    ]

    const range = store.salaryRange
    expect(range.min).toBe(25000)
    expect(range.max).toBe(200000)
  })

  it('handles empty salary range', () => {
    store.allPlayers = []
    const range = store.salaryRange
    expect(range.min).toBe(0)
    expect(range.max).toBe(1000000)
  })

  it('filters out non-numeric transfer values', () => {
    store.allPlayers = [
      { transferValueAmount: 1000000 },
      { transferValueAmount: 'invalid' },
      { transferValueAmount: null },
      { transferValueAmount: 5000000 }
    ]

    const range = store.currentDatasetTransferValueRange
    expect(range.min).toBe(1000000)
    expect(range.max).toBe(5000000)
  })

  it('filters out non-numeric wage amounts', () => {
    store.allPlayers = [
      { wageAmount: 50000 },
      { wageAmount: 'invalid' },
      { wageAmount: null },
      { wageAmount: 100000 }
    ]

    const range = store.salaryRange
    expect(range.min).toBe(50000)
    expect(range.max).toBe(100000)
  })

  it('uses sessionStorage for currentDatasetId', () => {
    // Reset the global sessionStorage mock and set up return values
    global.sessionStorage.getItem.mockReturnValue('test-dataset-123')

    // Create a new store instance to test initialization
    setActivePinia(createPinia())
    const _newStore = usePlayerStore()

    expect(global.sessionStorage.getItem).toHaveBeenCalledWith('currentDatasetId')

    // Cleanup
    vi.clearAllMocks()
  })

  it('uses sessionStorage for detectedCurrencySymbol', () => {
    // Reset the global sessionStorage mock and set up return values
    global.sessionStorage.getItem
      .mockReturnValueOnce(null) // currentDatasetId
      .mockReturnValueOnce('€') // detectedCurrencySymbol

    // Create a new store instance to test initialization
    setActivePinia(createPinia())
    const _newStore = usePlayerStore()

    expect(global.sessionStorage.getItem).toHaveBeenCalledWith('detectedCurrencySymbol')

    // Cleanup
    vi.clearAllMocks()
  })

  it('handles null values in computed properties gracefully', () => {
    store.allPlayers = null

    expect(store.uniqueClubs).toEqual([])
    expect(store.uniqueNationalities).toEqual([])
    expect(store.uniqueMediaHandlings).toEqual([])
    expect(store.uniquePersonalities).toEqual([])
    expect(store.uniquePositionsCount).toBe(0)
  })

  it('sorts computed arrays alphabetically', () => {
    store.allPlayers = [{ club: 'Zebra FC' }, { club: 'Alpha FC' }, { club: 'Beta FC' }]

    expect(store.uniqueClubs).toEqual(['Alpha FC', 'Beta FC', 'Zebra FC'])
  })

  it('handles media handling with various separators', () => {
    store.allPlayers = [
      { media_handling: 'Confident,Outspoken,Reserved' },
      { media_handling: ' Professional , Ambitious ' }
    ]

    const uniqueMediaHandlings = store.uniqueMediaHandlings
    expect(uniqueMediaHandlings).toContain('Confident')
    expect(uniqueMediaHandlings).toContain('Outspoken')
    expect(uniqueMediaHandlings).toContain('Reserved')
    expect(uniqueMediaHandlings).toContain('Professional')
    expect(uniqueMediaHandlings).toContain('Ambitious')
  })

  it('trims whitespace from media handling values', () => {
    store.allPlayers = [{ media_handling: ' Confident , Outspoken ' }]

    const uniqueMediaHandlings = store.uniqueMediaHandlings
    expect(uniqueMediaHandlings).toEqual(['Confident', 'Outspoken'])
  })

  it('initializes transfer value range when all values are same', () => {
    store.allPlayers = [
      { transferValueAmount: 1000000 },
      { transferValueAmount: 1000000 },
      { transferValueAmount: 1000000 }
    ]

    const range = store.currentDatasetTransferValueRange
    expect(range.min).toBe(1000000)
    expect(range.max).toBe(1050000) // Ensures max > min for slider
  })

  it('ensures non-negative minimum values', () => {
    store.allPlayers = [{ transferValueAmount: -500000 }, { transferValueAmount: 1000000 }]

    const range = store.currentDatasetTransferValueRange
    expect(range.min).toBe(0) // Negative values set to 0
    expect(range.max).toBe(1000000)
  })
})
