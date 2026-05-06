import { createPinia, setActivePinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

// Mock the logger at the top level to avoid import.meta.env issues
vi.mock('../utils/logger.js', () => ({
  default: {
    log: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
    debug: vi.fn(),
  },
  logger: {
    log: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
    debug: vi.fn(),
  },
  storeLogger: {
    log: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
    debug: vi.fn(),
  },
}))

// Mock the playerService
vi.mock('../services/playerService.js', () => ({
  default: {
    uploadPlayerFile: vi.fn(),
    getPlayersByDatasetId: vi.fn(),
    getAvailableRoles: vi.fn(),
    getClientStatus: vi.fn(),
    setProtobufEnabled: vi.fn(),
  },
}))

// Mock the performance tracker
vi.mock('../utils/performance.js', () => ({
  PerformanceTracker: class {
    checkpoint() {}
    finish() {}
  },
}))

// Mock protobufClient
vi.mock('../utils/protobufClient.js', () => ({
  default: {
    isSupported: vi.fn(() => true),
    setEnabled: vi.fn(),
    isEnabled: vi.fn(() => true),
  },
}))

describe('playerStore', () => {
  let store
  let playerService

  beforeEach(async () => {
    // Create a fresh pinia instance for each test
    setActivePinia(createPinia())

    // Reset mocks
    vi.resetAllMocks()

    // Mock sessionStorage
    vi.spyOn(window.sessionStorage, 'getItem').mockImplementation((key) => {
      if (key === 'currentDatasetId') return '123'
      if (key === 'detectedCurrencySymbol') return '€'
      return null
    })
    vi.spyOn(window.sessionStorage, 'setItem').mockImplementation(() => {})
    vi.spyOn(window.sessionStorage, 'removeItem').mockImplementation(() => {})

    // Dynamically import the store to ensure mocks are in place
    const { usePlayerStore } = await import('./playerStore')
    store = usePlayerStore()

    // Get the mocked playerService
    const playerServiceModule = await import('../services/playerService.js')
    playerService = playerServiceModule.default
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  describe('fetchPlayersByDatasetId', () => {
    it('should handle protobuf responses', async () => {
      // Mock a protobuf response
      const protobufResponse = {
        players: [
          { id: 1, name: 'Player 1', age: '25' },
          { id: 2, name: 'Player 2', age: '30' },
        ],
        currencySymbol: '$',
        _protobuf: {
          format: 'protobuf',
          payloadSize: 1000,
          compressionRatio: 0.6,
        },
      }

      playerService.getPlayersByDatasetId.mockResolvedValue(protobufResponse)

      await store.fetchPlayersByDatasetId('123')

      // Check that players were processed correctly
      expect(store.allPlayers.length).toBe(2)
      expect(store.allPlayers[0].age).toBe(25) // Should be converted to number

      // Check that protobuf metrics were updated
      expect(store.protobufMetrics.enabled).toBe(true)
      expect(store.protobufMetrics.requestCount).toBe(1)
      expect(store.protobufMetrics.averagePayloadSize).toBe(1000)
      expect(store.protobufMetrics.compressionRatio).toBe(0.6)
    })

    it('should handle JSON responses', async () => {
      // Mock a JSON response
      const jsonResponse = {
        players: [
          { id: 1, name: 'Player 1', age: '25' },
          { id: 2, name: 'Player 2', age: '30' },
        ],
        currencySymbol: '$',
        _protobuf: {
          format: 'json',
          payloadSize: 2000,
          fallbackReason: 'client_unsupported',
        },
      }

      playerService.getPlayersByDatasetId.mockResolvedValue(jsonResponse)

      await store.fetchPlayersByDatasetId('123')

      // Check that players were processed correctly
      expect(store.allPlayers.length).toBe(2)

      // Check that protobuf metrics were updated
      expect(store.protobufMetrics.enabled).toBe(false)
      expect(store.protobufMetrics.requestCount).toBe(1)
      expect(store.protobufMetrics.averagePayloadSize).toBe(2000)
    })

    it('should handle legacy JSON responses without protobuf metadata', async () => {
      // Mock a legacy JSON response without protobuf metadata
      const legacyResponse = {
        players: [
          { id: 1, name: 'Player 1', age: '25' },
          { id: 2, name: 'Player 2', age: '30' },
        ],
        currencySymbol: '$',
      }

      playerService.getPlayersByDatasetId.mockResolvedValue(legacyResponse)

      await store.fetchPlayersByDatasetId('123')

      // Check that players were processed correctly
      expect(store.allPlayers.length).toBe(2)

      // Protobuf metrics should not be updated
      expect(store.protobufMetrics.requestCount).toBe(0)
    })
  })

  describe('fetchAllAvailableRoles', () => {
    it('should handle protobuf responses', async () => {
      // Mock a protobuf response
      const protobufResponse = {
        roles: ['Advanced Forward', 'Complete Forward', 'Target Man'],
        _protobuf: {
          format: 'protobuf',
          payloadSize: 500,
        },
      }

      playerService.getAvailableRoles.mockResolvedValue(protobufResponse)

      await store.fetchAllAvailableRoles(true)

      // Check that roles were processed correctly
      expect(store.allAvailableRoles.length).toBe(3)
      expect(store.allAvailableRoles[0]).toBe('Advanced Forward')

      // Check that protobuf metrics were updated
      expect(store.protobufMetrics.requestCount).toBe(1)
      expect(store.protobufMetrics.averagePayloadSize).toBe(500)
    })

    it('should handle legacy JSON responses', async () => {
      // Mock a legacy JSON response (array directly)
      const legacyResponse = ['Advanced Forward', 'Complete Forward', 'Target Man']

      playerService.getAvailableRoles.mockResolvedValue(legacyResponse)

      await store.fetchAllAvailableRoles(true)

      // Check that roles were processed correctly
      expect(store.allAvailableRoles.length).toBe(3)
      expect(store.allAvailableRoles[0]).toBe('Advanced Forward')
    })
  })

  describe('dataset facets', () => {
    it('should compute dropdown facets when players are replaced', () => {
      store.setPlayers([
        {
          id: 1,
          name: 'Player 1',
          club: 'Arsenal',
          nationality: 'England',
          division: 'Premier League',
          media_handling: 'Media-friendly, Unflappable',
          personality: 'Professional',
        },
        {
          id: 2,
          name: 'Player 2',
          club: 'Chelsea',
          nationality: 'France',
          division: 'Premier League',
          media_handling: 'Unflappable',
          personality: 'Ambitious',
        },
        {
          id: 3,
          name: 'Player 3',
          club: 'Arsenal',
          nationality: 'England',
          division: 'Championship',
          media_handling: '',
          personality: 'Professional',
        },
      ])

      expect(store.uniqueClubs).toEqual(['Arsenal', 'Chelsea'])
      expect(store.uniqueNationalities).toEqual(['England', 'France'])
      expect(store.uniqueDivisions).toEqual(['Championship', 'Premier League'])
      expect(store.uniqueMediaHandlings).toEqual(['Media-friendly', 'Unflappable'])
      expect(store.uniquePersonalities).toEqual(['Ambitious', 'Professional'])
    })

    it('should clear dropdown facets on reset', () => {
      store.setPlayers([
        {
          id: 1,
          club: 'Arsenal',
          nationality: 'England',
          division: 'Premier League',
          media_handling: 'Media-friendly',
          personality: 'Professional',
        },
      ])

      store.resetState()

      expect(store.uniqueClubs).toEqual([])
      expect(store.uniqueNationalities).toEqual([])
      expect(store.uniqueDivisions).toEqual([])
      expect(store.uniqueMediaHandlings).toEqual([])
      expect(store.uniquePersonalities).toEqual([])
    })
  })

  describe('protobuf utilities', () => {
    it('should toggle protobuf support', () => {
      store.setProtobufEnabled(false)

      expect(playerService.setProtobufEnabled).toHaveBeenCalledWith(false)
    })

    it('should get protobuf status', () => {
      playerService.getClientStatus.mockReturnValue({
        protobufSupported: true,
        protobufEnabled: true,
        serverSupportsProtobuf: true,
      })

      const status = store.getProtobufStatus()

      expect(status.protobufSupported).toBe(true)
      expect(status.protobufEnabled).toBe(true)
      expect(status.serverSupportsProtobuf).toBe(true)
      expect(status.metrics).toBeDefined()
    })
  })
})
