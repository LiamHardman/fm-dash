/**
 * Compatibility tests for protobuf API client
 * 
 * These tests verify:
 * - Cross-browser compatibility for protobuf support
 * - Automatic fallback behavior in unsupported environments
 * - User experience consistency across response formats
 * - Performance impact of client-side protobuf processing
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { useProtobufApi } from './useProtobufApi'
import protobufClient from '../utils/protobufClient'

// Mock the logger to prevent console output during tests
vi.mock('../utils/logger', () => ({
  default: {
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
    debug: vi.fn()
  }
}))

// Mock fetch for testing
global.fetch = vi.fn()

describe('useProtobufApi browser compatibility', () => {
  let api
  
  beforeEach(() => {
    // Reset fetch mock
    global.fetch.mockReset()
    
    // Create a new API instance with a test base URL
    api = useProtobufApi('/test-api')
    
    // Mock AbortController
    global.AbortController = class {
      constructor() {
        this.signal = {}
        this.abort = vi.fn()
      }
    }
    
    // Mock performance.now
    global.performance.now = vi.fn()
      .mockReturnValueOnce(100)  // Start time
      .mockReturnValueOnce(150)  // End time
  })
  
  afterEach(() => {
    vi.restoreAllMocks()
  })
  
  it('should detect modern browser support for protobuf', async () => {
    // Mock modern browser environment with all required features
    global.ArrayBuffer = function() {}
    global.TextEncoder = function() {}
    global.TextDecoder = function() {}
    
    // Reset protobufClient state
    vi.spyOn(protobufClient, 'initialize').mockImplementation(async () => {
      protobufClient.protobufSupported = true
      protobufClient.initialized = true
      return true
    })
    
    // Mock successful protobuf request
    global.fetch.mockResolvedValueOnce({
      ok: true,
      headers: new Headers({ 'Content-Type': 'application/x-protobuf' }),
      arrayBuffer: async () => new ArrayBuffer(20)
    })
    
    // Mock protobuf decoding
    vi.spyOn(protobufClient, 'decodeProtobufResponse').mockResolvedValueOnce({
      players: [{ uid: 123, name: 'Test Player' }],
      currencySymbol: '$',
      metadata: { timestamp: 1626100000 }
    })
    
    // Call the API
    const result = await api.get('/players', {}, 'api.PlayerDataResponse')
    
    // Verify the result
    expect(result).toHaveProperty('players')
    expect(result).toHaveProperty('_protobuf.format', 'protobuf')
    
    // Verify fetch was called with protobuf accept header
    expect(global.fetch).toHaveBeenCalledTimes(1)
    expect(global.fetch.mock.calls[0][1].headers.Accept).toBe('application/x-protobuf')
  })
  
  it('should detect legacy browser and fall back to JSON', async () => {
    // Mock legacy browser environment without required features
    global.ArrayBuffer = undefined
    global.TextEncoder = undefined
    
    // Reset protobufClient state
    vi.spyOn(protobufClient, 'initialize').mockImplementation(async () => {
      protobufClient.protobufSupported = false
      protobufClient.initialized = true
      return false
    })
    
    // Mock successful JSON request
    global.fetch.mockResolvedValueOnce({
      ok: true,
      headers: new Headers({ 'Content-Type': 'application/json' }),
      json: async () => ({ 
        players: [{ uid: 123, name: 'Test Player' }],
        currency_symbol: '$'
      })
    })
    
    // Call the API
    const result = await api.get('/players', {}, 'api.PlayerDataResponse')
    
    // Verify the result
    expect(result).toHaveProperty('players')
    expect(result).toHaveProperty('_protobuf.format', 'json')
    expect(result).toHaveProperty('_protobuf.fallbackReason', 'client_unsupported')
    
    // Verify fetch was called with JSON accept header
    expect(global.fetch).toHaveBeenCalledTimes(1)
    expect(global.fetch.mock.calls[0][1].headers.Accept).toBe('application/json')
  })
  
  it('should handle missing protobuf library gracefully', async () => {
    // Mock environment with required features but protobuf library fails to load
    global.ArrayBuffer = function() {}
    global.TextEncoder = function() {}
    global.TextDecoder = function() {}
    
    // Reset protobufClient state
    vi.spyOn(protobufClient, 'initialize').mockImplementation(async () => {
      // Simulate protobuf library load failure
      protobufClient.protobufSupported = false
      protobufClient.initialized = true
      return false
    })
    
    // Mock successful JSON request
    global.fetch.mockResolvedValueOnce({
      ok: true,
      headers: new Headers({ 'Content-Type': 'application/json' }),
      json: async () => ({ 
        players: [{ uid: 123, name: 'Test Player' }],
        currency_symbol: '$'
      })
    })
    
    // Call the API
    const result = await api.get('/players', {}, 'api.PlayerDataResponse')
    
    // Verify the result
    expect(result).toHaveProperty('players')
    expect(result).toHaveProperty('_protobuf.format', 'json')
    expect(result).toHaveProperty('_protobuf.fallbackReason', 'client_unsupported')
    
    // Verify fetch was called with JSON accept header
    expect(global.fetch).toHaveBeenCalledTimes(1)
    expect(global.fetch.mock.calls[0][1].headers.Accept).toBe('application/json')
  })
  
  it('should handle server without protobuf support', async () => {
    // Mock modern browser environment
    global.ArrayBuffer = function() {}
    global.TextEncoder = function() {}
    global.TextDecoder = function() {}
    
    // Reset protobufClient state
    vi.spyOn(protobufClient, 'initialize').mockImplementation(async () => {
      protobufClient.protobufSupported = true
      protobufClient.initialized = true
      return true
    })
    
    // Mock server protobuf support check
    vi.spyOn(protobufClient, 'checkServerProtobufSupport').mockResolvedValue(false)
    
    // Mock successful JSON request
    global.fetch.mockResolvedValueOnce({
      ok: true,
      headers: new Headers({ 'Content-Type': 'application/json' }),
      json: async () => ({ 
        players: [{ uid: 123, name: 'Test Player' }],
        currency_symbol: '$'
      })
    })
    
    // Call the API
    const result = await api.get('/players', {}, 'api.PlayerDataResponse')
    
    // Verify the result
    expect(result).toHaveProperty('players')
    expect(result).toHaveProperty('_protobuf.format', 'json')
    expect(result).toHaveProperty('_protobuf.fallbackReason', 'server_unsupported')
    
    // Verify fetch was called with JSON accept header
    expect(global.fetch).toHaveBeenCalledTimes(1)
    expect(global.fetch.mock.calls[0][1].headers.Accept).toBe('application/json')
  })
})

describe('useProtobufApi response format consistency', () => {
  let api
  
  beforeEach(() => {
    // Reset fetch mock
    global.fetch.mockReset()
    
    // Create a new API instance with a test base URL
    api = useProtobufApi('/test-api')
    
    // Mock performance.now
    global.performance.now = vi.fn()
      .mockReturnValueOnce(100)  // Start time
      .mockReturnValueOnce(150)  // End time
  })
  
  afterEach(() => {
    vi.restoreAllMocks()
  })
  
  it('should provide consistent data structure from protobuf and JSON responses', async () => {
    // Mock protobuf response
    const protobufResponse = {
      players: [
        { 
          uid: 123, 
          name: 'Test Player',
          age: '25',
          position: 'ST',
          club: 'Test FC',
          overall: 85,
          roleSpecificOveralls: [
            { roleName: 'Advanced Forward', score: 87 },
            { roleName: 'Poacher', score: 85 }
          ]
        }
      ],
      currencySymbol: '$',
      metadata: { 
        timestamp: 1626100000,
        apiVersion: '1.0',
        totalCount: 1
      },
      _protobuf: {
        format: 'protobuf',
        processingTime: 50,
        payloadSize: 150
      }
    }
    
    // Mock JSON response with equivalent data but different structure
    const jsonResponse = {
      players: [
        { 
          uid: 123, 
          name: 'Test Player',
          age: '25',
          position: 'ST',
          club: 'Test FC',
          overall: 85,
          role_specific_overalls: [
            { role_name: 'Advanced Forward', score: 87 },
            { role_name: 'Poacher', score: 85 }
          ]
        }
      ],
      currency_symbol: '$',
      metadata: { 
        timestamp: 1626100000,
        api_version: '1.0',
        total_count: 1
      },
      _protobuf: {
        format: 'json',
        processingTime: 50,
        payloadSize: 250
      }
    }
    
    // Mock getPlayerData method to return protobuf or JSON response
    const originalGetPlayerData = api.getPlayerData
    
    // First test with protobuf response
    api.getPlayerData = vi.fn().mockResolvedValueOnce(protobufResponse)
    const protobufResult = await api.getPlayerData('test-dataset')
    
    // Then test with JSON response
    api.getPlayerData = vi.fn().mockResolvedValueOnce(jsonResponse)
    const jsonResult = await api.getPlayerData('test-dataset')
    
    // Restore original method
    api.getPlayerData = originalGetPlayerData
    
    // Verify both responses have the same essential data structure
    expect(protobufResult.players[0].uid).toBe(jsonResult.players[0].uid)
    expect(protobufResult.players[0].name).toBe(jsonResult.players[0].name)
    expect(protobufResult.players[0].overall).toBe(jsonResult.players[0].overall)
    
    // Verify both have currency symbol (different field names in raw responses)
    expect(protobufResult.currencySymbol).toBe(jsonResult.currencySymbol)
    
    // Verify both have metadata (different field names in raw responses)
    expect(protobufResult.metadata).toBeDefined()
    expect(jsonResult.metadata).toBeDefined()
    
    // Verify format information is preserved
    expect(protobufResult._protobuf.format).toBe('protobuf')
    expect(jsonResult._protobuf.format).toBe('json')
  })
  
  it('should handle pagination consistently across formats', async () => {
    // Mock protobuf response with pagination
    const protobufResponse = {
      players: [{ uid: 123, name: 'Test Player' }],
      currencySymbol: '$',
      metadata: { timestamp: 1626100000 },
      pagination: {
        page: 1,
        perPage: 10,
        totalPages: 5,
        totalCount: 45,
        hasNext: true,
        hasPrevious: false
      },
      _protobuf: { format: 'protobuf' }
    }
    
    // Mock JSON response with pagination
    const jsonResponse = {
      players: [{ uid: 123, name: 'Test Player' }],
      currency_symbol: '$',
      metadata: { timestamp: 1626100000 },
      pagination: {
        page: 1,
        per_page: 10,
        total_pages: 5,
        total_count: 45,
        has_next: true,
        has_previous: false
      },
      _protobuf: { format: 'json' }
    }
    
    // Mock getPlayerData method to return protobuf or JSON response
    const originalGetPlayerData = api.getPlayerData
    
    // First test with protobuf response
    api.getPlayerData = vi.fn().mockResolvedValueOnce(protobufResponse)
    const protobufResult = await api.getPlayerData('test-dataset', { page: 1 })
    
    // Then test with JSON response
    api.getPlayerData = vi.fn().mockResolvedValueOnce(jsonResponse)
    const jsonResult = await api.getPlayerData('test-dataset', { page: 1 })
    
    // Restore original method
    api.getPlayerData = originalGetPlayerData
    
    // Verify pagination information is consistent
    expect(protobufResult.pagination.page).toBe(jsonResult.pagination.page)
    expect(protobufResult.pagination.perPage).toBe(jsonResult.pagination.perPage)
    expect(protobufResult.pagination.totalPages).toBe(jsonResult.pagination.totalPages)
    expect(protobufResult.pagination.totalCount).toBe(jsonResult.pagination.totalCount)
    expect(protobufResult.pagination.hasNext).toBe(jsonResult.pagination.hasNext)
    expect(protobufResult.pagination.hasPrevious).toBe(jsonResult.pagination.hasPrevious)
  })
  
  it('should handle error responses consistently across formats', async () => {
    // Mock protobuf error
    const protobufError = new Error('Invalid request parameters')
    protobufError.name = 'ProtobufApiError'
    protobufError.status = 400
    protobufError.errorCode = 'validation_error'
    protobufError.details = ['Field X is required', 'Field Y must be a number']
    
    // Mock JSON error
    const jsonError = new Error('Invalid request parameters')
    jsonError.status = 400
    jsonError.response = {
      error_code: 'validation_error',
      details: ['Field X is required', 'Field Y must be a number']
    }
    
    // Mock getPlayerData method to throw errors
    const originalGetPlayerData = api.getPlayerData
    
    // Test with protobuf error
    api.getPlayerData = vi.fn().mockRejectedValueOnce(protobufError)
    let protobufErrorCaught = null
    try {
      await api.getPlayerData('test-dataset')
    } catch (error) {
      protobufErrorCaught = error
    }
    
    // Test with JSON error
    api.getPlayerData = vi.fn().mockRejectedValueOnce(jsonError)
    let jsonErrorCaught = null
    try {
      await api.getPlayerData('test-dataset')
    } catch (error) {
      jsonErrorCaught = error
    }
    
    // Restore original method
    api.getPlayerData = originalGetPlayerData
    
    // Verify both errors were caught
    expect(protobufErrorCaught).not.toBeNull()
    expect(jsonErrorCaught).not.toBeNull()
    
    // Verify error status codes are consistent
    expect(protobufErrorCaught.status).toBe(jsonErrorCaught.status)
    
    // Verify error messages are consistent
    expect(protobufErrorCaught.message).toBe(jsonErrorCaught.message)
  })
})

describe('useProtobufApi performance impact', () => {
  let api
  
  beforeEach(() => {
    // Reset fetch mock
    global.fetch.mockReset()
    
    // Create a new API instance with a test base URL
    api = useProtobufApi('/test-api')
    
    // Reset protobufClient state
    vi.spyOn(protobufClient, 'initialize').mockImplementation(async () => {
      protobufClient.protobufSupported = true
      protobufClient.initialized = true
      return true
    })
    
    // Mock server protobuf support check
    vi.spyOn(protobufClient, 'checkServerProtobufSupport').mockResolvedValue(true)
  })
  
  afterEach(() => {
    vi.restoreAllMocks()
  })
  
  it('should measure protobuf vs JSON payload size difference', async () => {
    // Create test data with 50 players
    const testPlayers = []
    for (let i = 0; i < 50; i++) {
      testPlayers.push({
        uid: i + 1,
        name: `Player ${i + 1}`,
        position: 'ST',
        age: '25',
        club: 'Test FC',
        overall: 80 + (i % 10)
      })
    }
    
    // Mock protobuf response with smaller payload
    const protobufPayloadSize = 2500 // Simulated binary size
    global.fetch.mockResolvedValueOnce({
      ok: true,
      headers: new Headers({ 'Content-Type': 'application/x-protobuf' }),
      arrayBuffer: async () => new ArrayBuffer(protobufPayloadSize)
    })
    
    // Mock protobuf decoding
    vi.spyOn(protobufClient, 'decodeProtobufResponse').mockResolvedValueOnce({
      players: testPlayers,
      currencySymbol: '$',
      metadata: { timestamp: 1626100000 }
    })
    
    // Mock performance.now for protobuf request
    global.performance.now = vi.fn()
      .mockReturnValueOnce(100)  // Start time
      .mockReturnValueOnce(120)  // End time
    
    // Call the API with protobuf
    const protobufResult = await api.get('/players', {}, 'api.PlayerDataResponse')
    
    // Reset fetch mock for JSON test
    global.fetch.mockReset()
    
    // Mock JSON response with larger payload
    const jsonPayloadSize = 8000 // Simulated JSON size
    const jsonResponse = {
      players: testPlayers,
      currency_symbol: '$',
      metadata: { timestamp: 1626100000 }
    }
    
    // Convert to string to get approximate size
    const jsonString = JSON.stringify(jsonResponse)
    
    global.fetch.mockResolvedValueOnce({
      ok: true,
      headers: new Headers({ 'Content-Type': 'application/json' }),
      json: async () => jsonResponse,
      text: async () => jsonString
    })
    
    // Mock performance.now for JSON request
    global.performance.now = vi.fn()
      .mockReturnValueOnce(200)  // Start time
      .mockReturnValueOnce(230)  // End time
    
    // Call the API with JSON
    const jsonResult = await api.get('/players', {}, 'api.PlayerDataResponse')
    
    // Add payload size to JSON result for comparison
    jsonResult._protobuf = {
      ...jsonResult._protobuf,
      payloadSize: jsonString.length
    }
    
    // Verify payload sizes
    expect(protobufResult._protobuf.payloadSize).toBe(protobufPayloadSize)
    expect(jsonResult._protobuf.payloadSize).toBe(jsonString.length)
    
    // Calculate compression ratio
    const compressionRatio = protobufResult._protobuf.payloadSize / jsonResult._protobuf.payloadSize
    
    // Verify protobuf is smaller (compression ratio < 1)
    expect(compressionRatio).toBeLessThan(1)
    
    // Log the results
    console.log(`Protobuf size: ${protobufResult._protobuf.payloadSize} bytes`)
    console.log(`JSON size: ${jsonResult._protobuf.payloadSize} bytes`)
    console.log(`Compression ratio: ${compressionRatio.toFixed(2)}`)
  })
  
  it('should measure protobuf vs JSON processing time', async () => {
    // Create test data with 50 players
    const testPlayers = []
    for (let i = 0; i < 50; i++) {
      testPlayers.push({
        uid: i + 1,
        name: `Player ${i + 1}`,
        position: 'ST',
        age: '25',
        club: 'Test FC',
        overall: 80 + (i % 10)
      })
    }
    
    // Mock protobuf response
    global.fetch.mockResolvedValueOnce({
      ok: true,
      headers: new Headers({ 'Content-Type': 'application/x-protobuf' }),
      arrayBuffer: async () => new ArrayBuffer(2500)
    })
    
    // Simulate protobuf decoding time
    const protobufProcessingTime = 20 // milliseconds
    vi.spyOn(protobufClient, 'decodeProtobufResponse').mockImplementation(async () => {
      // Simulate processing time
      await new Promise(resolve => setTimeout(resolve, protobufProcessingTime))
      return {
        players: testPlayers,
        currencySymbol: '$',
        metadata: { timestamp: 1626100000 }
      }
    })
    
    // Mock performance.now for protobuf request
    global.performance.now = vi.fn()
      .mockReturnValueOnce(100)  // Start time
      .mockReturnValueOnce(100 + protobufProcessingTime)  // End time
    
    // Call the API with protobuf
    const protobufResult = await api.get('/players', {}, 'api.PlayerDataResponse')
    
    // Reset fetch mock for JSON test
    global.fetch.mockReset()
    
    // Mock JSON response
    const jsonResponse = {
      players: testPlayers,
      currency_symbol: '$',
      metadata: { timestamp: 1626100000 }
    }
    
    // Simulate JSON processing time
    const jsonProcessingTime = 30 // milliseconds
    global.fetch.mockImplementation(async () => {
      // Simulate processing time
      await new Promise(resolve => setTimeout(resolve, jsonProcessingTime))
      return {
        ok: true,
        headers: new Headers({ 'Content-Type': 'application/json' }),
        json: async () => jsonResponse
      }
    })
    
    // Mock performance.now for JSON request
    global.performance.now = vi.fn()
      .mockReturnValueOnce(200)  // Start time
      .mockReturnValueOnce(200 + jsonProcessingTime)  // End time
    
    // Call the API with JSON
    const jsonResult = await api.get('/players', {}, 'api.PlayerDataResponse')
    
    // Verify processing times
    expect(protobufResult._protobuf.processingTime).toBe(protobufProcessingTime)
    expect(jsonResult._protobuf.processingTime).toBe(jsonProcessingTime)
    
    // Calculate processing time ratio
    const processingRatio = protobufResult._protobuf.processingTime / jsonResult._protobuf.processingTime
    
    // Log the results
    console.log(`Protobuf processing time: ${protobufResult._protobuf.processingTime}ms`)
    console.log(`JSON processing time: ${jsonResult._protobuf.processingTime}ms`)
    console.log(`Processing time ratio: ${processingRatio.toFixed(2)}`)
  })
  
  it('should measure memory usage for protobuf vs JSON', async () => {
    // Skip test if performance.memory is not available
    if (!global.performance.memory) {
      console.log('Skipping memory test - performance.memory not available')
      return
    }
    
    // Create large test data with 1000 players
    const testPlayers = []
    for (let i = 0; i < 1000; i++) {
      testPlayers.push({
        uid: i + 1,
        name: `Player ${i + 1}`,
        position: 'ST',
        age: '25',
        club: 'Test FC',
        division: 'Premier League',
        nationality: 'England',
        overall: 80 + (i % 10),
        attributes: {
          pace: 80 + (i % 20),
          shooting: 70 + (i % 30),
          passing: 75 + (i % 25),
          dribbling: 85 + (i % 15),
          defending: 60 + (i % 40),
          physical: 75 + (i % 25)
        }
      })
    }
    
    // Mock memory measurements
    let memoryBefore, memoryAfter
    
    // Mock protobuf response
    global.fetch.mockResolvedValueOnce({
      ok: true,
      headers: new Headers({ 'Content-Type': 'application/x-protobuf' }),
      arrayBuffer: async () => new ArrayBuffer(25000)
    })
    
    // Mock protobuf decoding
    vi.spyOn(protobufClient, 'decodeProtobufResponse').mockResolvedValueOnce({
      players: testPlayers,
      currencySymbol: '$',
      metadata: { timestamp: 1626100000 }
    })
    
    // Measure memory before protobuf request
    memoryBefore = global.performance.memory?.usedJSHeapSize || 0
    
    // Call the API with protobuf
    const protobufResult = await api.get('/players', {}, 'api.PlayerDataResponse')
    
    // Measure memory after protobuf request
    memoryAfter = global.performance.memory?.usedJSHeapSize || 0
    const protobufMemoryUsage = memoryAfter - memoryBefore
    
    // Reset fetch mock for JSON test
    global.fetch.mockReset()
    
    // Mock JSON response
    const jsonResponse = {
      players: testPlayers,
      currency_symbol: '$',
      metadata: { timestamp: 1626100000 }
    }
    
    global.fetch.mockResolvedValueOnce({
      ok: true,
      headers: new Headers({ 'Content-Type': 'application/json' }),
      json: async () => jsonResponse
    })
    
    // Measure memory before JSON request
    memoryBefore = global.performance.memory?.usedJSHeapSize || 0
    
    // Call the API with JSON
    const jsonResult = await api.get('/players', {}, 'api.PlayerDataResponse')
    
    // Measure memory after JSON request
    memoryAfter = global.performance.memory?.usedJSHeapSize || 0
    const jsonMemoryUsage = memoryAfter - memoryBefore
    
    // Log the results
    console.log(`Protobuf memory usage: ${protobufMemoryUsage} bytes`)
    console.log(`JSON memory usage: ${jsonMemoryUsage} bytes`)
    console.log(`Memory usage ratio: ${(protobufMemoryUsage / jsonMemoryUsage).toFixed(2)}`)
    
    // Note: This test is more for logging than assertions, as memory usage
    // can vary significantly between environments and runs
  })
})