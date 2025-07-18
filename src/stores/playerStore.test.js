import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { usePlayerStore } from './playerStore'

// Mock the playerService
vi.mock('../services/playerService.js', () => ({
  default: {
    uploadPlayerFile: vi.fn(),
    getPlayersByDatasetId: vi.fn(),
    getAvailableRoles: vi.fn(),
    getClientStatus: vi.fn(),
    setProtobufEnabled: vi.fn()
  }
}))

// Mock the performance tracker
vi.mock('../utils/performance.js', () => ({
  PerformanceTracker: class {
    constructor() {}
    checkpoint() {}
    finish() {}
  }
}))

describe('playerStore', () => {
  let store
  let playerService

  beforeEach(() => {
    // Create a fresh pinia instance for each test
    setActivePinia(createPinia())
    
    // Get the store instance
    store = usePlayerStore()
    
    // Get the mocked playerService
    playerService = require('../services/playerService.js').default
    
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
          { id: 2, name: 'Player 2', age: '30' }
        ],
        currencySymbol: '$',
        _protobuf: {
          format: 'protobuf',
          payloadSize: 1000,
          compressionRatio: 0.6
        }
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
          { id: 2, name: 'Player 2', age: '30' }
        ],
        currencySymbol: '$',
        _protobuf: {
          format: 'json',
          payloadSize: 2000,
          fallbackReason: 'client_unsupported'
        }
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
          { id: 2, name: 'Player 2', age: '30' }
        ],
        currencySymbol: '$'
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
          payloadSize: 500
        }
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
  
  describe('protobuf utilities', () => {
    it('should toggle protobuf support', () => {
      store.setProtobufEnabled(false)
      
      expect(playerService.setProtobufEnabled).toHaveBeenCalledWith(false)
    })
    
    it('should get protobuf status', () => {
      playerService.getClientStatus.mockReturnValue({
        protobufSupported: true,
        protobufEnabled: true,
        serverSupportsProtobuf: true
      })
      
      const status = store.getProtobufStatus()
      
      expect(status.protobufSupported).toBe(true)
      expect(status.protobufEnabled).toBe(true)
      expect(status.serverSupportsProtobuf).toBe(true)
      expect(status.metrics).toBeDefined()
    })
  })
})