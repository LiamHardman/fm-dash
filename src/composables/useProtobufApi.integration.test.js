/**
 * Integration tests for protobuf API error handling and fallback mechanisms
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

describe('useProtobufApi error handling and fallback', () => {
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
    global.performance.now = vi.fn().mockReturnValue(100)
    
    // Reset protobufClient state
    vi.spyOn(protobufClient, 'initialize').mockImplementation(async () => true)
    vi.spyOn(protobufClient, 'protobufSupported', 'get').mockReturnValue(true)
    vi.spyOn(protobufClient, 'checkServerProtobufSupport').mockResolvedValue(true)
  })
  
  afterEach(() => {
    vi.restoreAllMocks()
  })
  
  it('should handle network errors and fall back to JSON', async () => {
    // Mock a network error
    global.fetch.mockRejectedValueOnce(new TypeError('Failed to fetch'))
    
    // Mock successful JSON fallback
    global.fetch.mockResolvedValueOnce({
      ok: true,
      headers: new Headers({ 'Content-Type': 'application/json' }),
      json: async () => ({ data: 'test', success: true })
    })
    
    // Call the API
    const result = await api.get('/players', {}, 'api.PlayerDataResponse')
    
    // Verify the result
    expect(result).toHaveProperty('data', 'test')
    expect(result).toHaveProperty('success', true)
    expect(result).toHaveProperty('_protobuf.format', 'json')
    expect(result).toHaveProperty('_protobuf.fallbackReason', 'network_error')
    
    // Verify fetch was called twice (once for protobuf, once for JSON fallback)
    expect(global.fetch).toHaveBeenCalledTimes(2)
  })
  
  it('should handle protobuf decoding errors and fall back to JSON', async () => {
    // Mock a successful fetch but with invalid protobuf data
    global.fetch.mockResolvedValueOnce({
      ok: true,
      headers: new Headers({ 'Content-Type': 'application/x-protobuf' }),
      arrayBuffer: async () => new ArrayBuffer(10) // Invalid protobuf data
    })
    
    // Mock protobuf decoding error
    vi.spyOn(protobufClient, 'decodeProtobufResponse').mockRejectedValueOnce(
      Object.assign(new Error('Failed to decode protobuf'), { 
        name: 'ProtobufDecodingError',
        bufferSize: 10
      })
    )
    
    // Mock successful JSON fallback
    global.fetch.mockResolvedValueOnce({
      ok: true,
      headers: new Headers({ 'Content-Type': 'application/json' }),
      json: async () => ({ data: 'test', success: true })
    })
    
    // Call the API
    const result = await api.get('/players', {}, 'api.PlayerDataResponse')
    
    // Verify the result
    expect(result).toHaveProperty('data', 'test')
    expect(result).toHaveProperty('success', true)
    expect(result).toHaveProperty('_protobuf.format', 'json')
    expect(result).toHaveProperty('_protobuf.fallbackReason', 'decoding_failed')
    
    // Verify fetch was called twice (once for protobuf, once for JSON fallback)
    expect(global.fetch).toHaveBeenCalledTimes(2)
  })
  
  it('should handle server-side protobuf errors correctly', async () => {
    // Create a mock protobuf error response
    const mockErrorBuffer = new ArrayBuffer(20) // Simulated protobuf error data
    
    // Mock a failed fetch with protobuf error response
    global.fetch.mockResolvedValueOnce({
      ok: false,
      status: 400,
      statusText: 'Bad Request',
      headers: new Headers({ 'Content-Type': 'application/x-protobuf' }),
      arrayBuffer: async () => mockErrorBuffer
    })
    
    // Mock protobuf decoding for error response
    vi.spyOn(protobufClient, 'decodeProtobufResponse').mockResolvedValueOnce({
      errorCode: 'validation_error',
      message: 'Invalid player data',
      details: ['Field X is required', 'Field Y must be a number'],
      metadata: {
        timestamp: 1626100000,
        apiVersion: '1.0',
        requestId: 'test-request-id'
      }
    })
    
    // Call the API and expect it to throw
    try {
      await api.get('/players', {}, 'api.PlayerDataResponse')
      // Should not reach here
      expect(true).toBe(false)
    } catch (error) {
      // Verify the error
      expect(error).toHaveProperty('name', 'ProtobufApiError')
      expect(error).toHaveProperty('errorCode', 'validation_error')
      expect(error).toHaveProperty('message', 'Invalid player data')
      expect(error).toHaveProperty('details').toEqual([
        'Field X is required', 
        'Field Y must be a number'
      ])
      expect(error).toHaveProperty('status', 400)
      expect(error).toHaveProperty('requestUrl', '/players')
      expect(error).toHaveProperty('messageType', 'api.PlayerDataResponse')
    }
    
    // Verify fetch was called once
    expect(global.fetch).toHaveBeenCalledTimes(1)
  })
  
  it('should handle server-side fallback indicators', async () => {
    // Mock a successful fetch with server-side fallback indicator
    global.fetch.mockResolvedValueOnce({
      ok: true,
      headers: new Headers({ 
        'Content-Type': 'application/json',
        'X-Serialization-Fallback': 'protobuf_marshal_failed'
      }),
      json: async () => ({ data: 'test', success: true })
    })
    
    // Call the API
    const result = await api.get('/players', {}, 'api.PlayerDataResponse')
    
    // Verify the result
    expect(result).toHaveProperty('data', 'test')
    expect(result).toHaveProperty('success', true)
    expect(result).toHaveProperty('_protobuf.format', 'json')
    expect(result).toHaveProperty('_protobuf.fallbackReason', 'server_returned_non_protobuf')
    
    // Verify fetch was called once
    expect(global.fetch).toHaveBeenCalledTimes(1)
  })
  
  it('should retry on specific errors before falling back', async () => {
    // Mock a network error for the first attempt
    global.fetch.mockRejectedValueOnce(new TypeError('Failed to fetch'))
    
    // Mock a successful fetch for the retry
    global.fetch.mockResolvedValueOnce({
      ok: true,
      headers: new Headers({ 'Content-Type': 'application/x-protobuf' }),
      arrayBuffer: async () => new ArrayBuffer(20) // Valid protobuf data
    })
    
    // Mock successful protobuf decoding
    vi.spyOn(protobufClient, 'decodeProtobufResponse').mockResolvedValueOnce({
      players: [{ uid: 123, name: 'Test Player' }],
      currencySymbol: '$',
      metadata: { timestamp: 1626100000 }
    })
    
    // Call the API
    const result = await api.get('/players', {}, 'api.PlayerDataResponse')
    
    // Verify the result
    expect(result).toHaveProperty('players').toEqual([{ uid: 123, name: 'Test Player' }])
    expect(result).toHaveProperty('currencySymbol', '$')
    expect(result).toHaveProperty('_protobuf.format', 'protobuf')
    expect(result).toHaveProperty('_protobuf.retryCount', 1)
    
    // Verify fetch was called twice (initial attempt + retry)
    expect(global.fetch).toHaveBeenCalledTimes(2)
  })
  
  it('should fall back to JSON when protobuf is not supported', async () => {
    // Mock protobuf not supported
    vi.spyOn(protobufClient, 'protobufSupported', 'get').mockReturnValue(false)
    
    // Mock successful JSON fetch
    global.fetch.mockResolvedValueOnce({
      ok: true,
      headers: new Headers({ 'Content-Type': 'application/json' }),
      json: async () => ({ data: 'test', success: true })
    })
    
    // Call the API
    const result = await api.get('/players', {}, 'api.PlayerDataResponse')
    
    // Verify the result
    expect(result).toHaveProperty('data', 'test')
    expect(result).toHaveProperty('success', true)
    expect(result).toHaveProperty('_protobuf.format', 'json')
    expect(result).toHaveProperty('_protobuf.fallbackReason', 'client_unsupported')
    
    // Verify fetch was called once (directly with JSON)
    expect(global.fetch).toHaveBeenCalledTimes(1)
    expect(global.fetch.mock.calls[0][1].headers.Accept).toBe('application/json')
  })
  
  it('should fall back to JSON when server does not support protobuf', async () => {
    // Mock server does not support protobuf
    vi.spyOn(protobufClient, 'checkServerProtobufSupport').mockResolvedValue(false)
    
    // Mock successful JSON fetch
    global.fetch.mockResolvedValueOnce({
      ok: true,
      headers: new Headers({ 'Content-Type': 'application/json' }),
      json: async () => ({ data: 'test', success: true })
    })
    
    // Call the API
    const result = await api.get('/players', {}, 'api.PlayerDataResponse')
    
    // Verify the result
    expect(result).toHaveProperty('data', 'test')
    expect(result).toHaveProperty('success', true)
    expect(result).toHaveProperty('_protobuf.format', 'json')
    expect(result).toHaveProperty('_protobuf.fallbackReason', 'server_unsupported')
    
    // Verify fetch was called once (directly with JSON)
    expect(global.fetch).toHaveBeenCalledTimes(1)
    expect(global.fetch.mock.calls[0][1].headers.Accept).toBe('application/json')
  })
  
  it('should track metrics for failed requests', async () => {
    // Mock a network error that will not be retried
    const networkError = new Error('Network error')
    networkError.name = 'NetworkError'
    global.fetch.mockRejectedValueOnce(networkError)
    
    // Mock successful JSON fallback
    global.fetch.mockResolvedValueOnce({
      ok: true,
      headers: new Headers({ 'Content-Type': 'application/json' }),
      json: async () => ({ data: 'test', success: true })
    })
    
    // Call the API
    await api.get('/players', {}, 'api.PlayerDataResponse')
    
    // Verify metrics were updated
    expect(api.metrics.failedRequests).toBe(1)
    expect(api.metrics.errorsByType).toHaveProperty('NetworkError', 1)
  })
  
  it('should handle request abortion correctly', async () => {
    // Create a mock abort controller
    const mockAbort = vi.fn()
    global.AbortController = class {
      constructor() {
        this.signal = {}
        this.abort = mockAbort
      }
    }
    
    // Start a request
    const promise = api.get('/players', {}, 'api.PlayerDataResponse')
    
    // Cancel the request
    api.cancel()
    
    // Mock fetch to resolve with an AbortError
    const abortError = new Error('The operation was aborted')
    abortError.name = 'AbortError'
    global.fetch.mockRejectedValueOnce(abortError)
    
    // Wait for the promise to resolve
    const result = await promise
    
    // Verify the result is null (aborted)
    expect(result).toBeNull()
    
    // Verify abort was called
    expect(mockAbort).toHaveBeenCalled()
  })
})