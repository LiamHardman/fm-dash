/**
 * Unit tests for protobufClient
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import protobufClient from './protobufClient'

// Mock the protobufjs library
vi.mock('protobufjs/minimal', () => {
  return {
    default: {
      Root: {
        fromJSON: vi.fn().mockReturnValue({
          lookupType: vi.fn().mockReturnValue({
            decode: vi.fn().mockImplementation(() => ({ test: 'data' })),
            toObject: vi.fn().mockImplementation(() => ({ test: 'data', converted: true }))
          })
        })
      }
    }
  }
})

// Mock the logger
vi.mock('./logger', () => ({
  default: {
    error: vi.fn(),
    warn: vi.fn(),
    info: vi.fn()
  }
}))

describe('protobufClient', () => {
  // Mock fetch
  const originalFetch = global.fetch
  
  beforeEach(() => {
    // Reset the client before each test
    protobufClient.initialized = false
    protobufClient.protobufSupported = false
    protobufClient.protoDefinitions = null
    protobufClient.serverSupportsProtobuf = null
    protobufClient.protobufEnabled = true
    protobufClient.initializationPromise = null
    
    // Mock fetch
    global.fetch = vi.fn()
    
    // Mock performance.now
    global.performance = {
      now: vi.fn().mockReturnValueOnce(0).mockReturnValueOnce(100)
    }
  })
  
  afterEach(() => {
    // Restore fetch
    global.fetch = originalFetch
    
    // Clear all mocks
    vi.clearAllMocks()
  })
  
  describe('initialize', () => {
    it('should detect protobuf support', async () => {
      // Mock the detectProtobufSupport method
      const detectSpy = vi.spyOn(protobufClient, 'detectProtobufSupport')
        .mockResolvedValue(true)
      
      // Mock the loadProtobufDefinitions method
      const loadSpy = vi.spyOn(protobufClient, 'loadProtobufDefinitions')
        .mockResolvedValue(true)
      
      await protobufClient.initialize()
      
      expect(detectSpy).toHaveBeenCalled()
      expect(loadSpy).toHaveBeenCalled()
      expect(protobufClient.initialized).toBe(true)
      expect(protobufClient.protobufSupported).toBe(true)
    })
    
    it('should handle initialization errors', async () => {
      // Mock the detectProtobufSupport method to throw an error
      vi.spyOn(protobufClient, 'detectProtobufSupport')
        .mockRejectedValue(new Error('Test error'))
      
      await protobufClient.initialize()
      
      expect(protobufClient.initialized).toBe(true)
      expect(protobufClient.protobufSupported).toBe(false)
    })
    
    it('should not initialize twice', async () => {
      // Set initialized to true
      protobufClient.initialized = true
      
      // Mock the detectProtobufSupport method
      const detectSpy = vi.spyOn(protobufClient, 'detectProtobufSupport')
      
      await protobufClient.initialize()
      
      expect(detectSpy).not.toHaveBeenCalled()
    })
  })
  
  describe('detectProtobufSupport', () => {
    it('should detect protobuf support correctly', async () => {
      // Mock the necessary browser APIs
      global.ArrayBuffer = function() {}
      global.TextEncoder = function() {}
      global.TextDecoder = function() {}
      
      const result = await protobufClient.detectProtobufSupport()
      
      expect(result).toBe(true)
    })
    
    it('should handle missing browser APIs', async () => {
      // Remove necessary browser APIs
      const originalArrayBuffer = global.ArrayBuffer
      global.ArrayBuffer = undefined
      
      const result = await protobufClient.detectProtobufSupport()
      
      expect(result).toBe(false)
      
      // Restore the API
      global.ArrayBuffer = originalArrayBuffer
    })
  })
  
  describe('checkServerProtobufSupport', () => {
    it('should detect server protobuf support', async () => {
      // Mock fetch to return a response with protobuf content type
      global.fetch.mockResolvedValue({
        headers: {
          get: vi.fn().mockReturnValue('application/x-protobuf')
        }
      })
      
      const result = await protobufClient.checkServerProtobufSupport('/api/test')
      
      expect(result).toBe(true)
      expect(protobufClient.serverSupportsProtobuf).toBe(true)
      expect(global.fetch).toHaveBeenCalledWith('/api/test', {
        method: 'HEAD',
        headers: {
          'Accept': 'application/x-protobuf'
        }
      })
    })
    
    it('should handle server without protobuf support', async () => {
      // Mock fetch to return a response without protobuf content type
      global.fetch.mockResolvedValue({
        headers: {
          get: vi.fn().mockReturnValue('application/json')
        }
      })
      
      const result = await protobufClient.checkServerProtobufSupport('/api/test')
      
      expect(result).toBe(false)
      expect(protobufClient.serverSupportsProtobuf).toBe(false)
    })
    
    it('should handle fetch errors', async () => {
      // Mock fetch to throw an error
      global.fetch.mockRejectedValue(new Error('Network error'))
      
      const result = await protobufClient.checkServerProtobufSupport('/api/test')
      
      expect(result).toBe(false)
      expect(protobufClient.serverSupportsProtobuf).toBe(false)
    })
    
    it('should use cached server support value', async () => {
      // Set serverSupportsProtobuf
      protobufClient.serverSupportsProtobuf = true
      
      const result = await protobufClient.checkServerProtobufSupport('/api/test')
      
      expect(result).toBe(true)
      expect(global.fetch).not.toHaveBeenCalled()
    })
  })
  
  describe('fetchWithProtobuf', () => {
    it('should fetch data with protobuf support', async () => {
      // Mock initialize to set protobufSupported to true
      vi.spyOn(protobufClient, 'initialize').mockResolvedValue(true)
      protobufClient.protobufSupported = true
      
      // Mock checkServerProtobufSupport to return true
      vi.spyOn(protobufClient, 'checkServerProtobufSupport').mockResolvedValue(true)
      
      // Mock fetch to return a protobuf response
      global.fetch.mockResolvedValue({
        ok: true,
        headers: {
          get: vi.fn().mockReturnValue('application/x-protobuf')
        },
        arrayBuffer: vi.fn().mockResolvedValue(new ArrayBuffer(10))
      })
      
      // Mock decodeProtobufResponse
      vi.spyOn(protobufClient, 'decodeProtobufResponse')
        .mockResolvedValue({ data: 'test data' })
      
      const result = await protobufClient.fetchWithProtobuf('/api/test', {}, 'api.TestMessage')
      
      expect(result).toEqual({
        data: 'test data',
        _protobuf: {
          processingTime: 100,
          payloadSize: 10,
          format: 'protobuf'
        }
      })
    })
    
    it('should fall back to JSON when protobuf is not supported', async () => {
      // Mock initialize to set protobufSupported to false
      vi.spyOn(protobufClient, 'initialize').mockResolvedValue(false)
      protobufClient.protobufSupported = false
      
      // Mock fallbackToJSON
      vi.spyOn(protobufClient, 'fallbackToJSON')
        .mockResolvedValue({ data: 'json data' })
      
      const result = await protobufClient.fetchWithProtobuf('/api/test', {}, 'api.TestMessage')
      
      expect(protobufClient.fallbackToJSON).toHaveBeenCalledWith('/api/test', {})
      expect(result).toEqual({ data: 'json data' })
    })
    
    it('should fall back to JSON when server does not support protobuf', async () => {
      // Mock initialize to set protobufSupported to true
      vi.spyOn(protobufClient, 'initialize').mockResolvedValue(true)
      protobufClient.protobufSupported = true
      
      // Mock checkServerProtobufSupport to return false
      vi.spyOn(protobufClient, 'checkServerProtobufSupport').mockResolvedValue(false)
      
      // Mock fallbackToJSON
      vi.spyOn(protobufClient, 'fallbackToJSON')
        .mockResolvedValue({ data: 'json data' })
      
      const result = await protobufClient.fetchWithProtobuf('/api/test', {}, 'api.TestMessage')
      
      expect(protobufClient.fallbackToJSON).toHaveBeenCalledWith('/api/test', {})
      expect(result).toEqual({ data: 'json data' })
    })
    
    it('should handle non-protobuf responses', async () => {
      // Mock initialize to set protobufSupported to true
      vi.spyOn(protobufClient, 'initialize').mockResolvedValue(true)
      protobufClient.protobufSupported = true
      
      // Mock checkServerProtobufSupport to return true
      vi.spyOn(protobufClient, 'checkServerProtobufSupport').mockResolvedValue(true)
      
      // Mock fetch to return a JSON response
      global.fetch.mockResolvedValue({
        ok: true,
        headers: {
          get: vi.fn().mockReturnValue('application/json')
        }
      })
      
      // Mock handleNonProtobufResponse
      vi.spyOn(protobufClient, 'handleNonProtobufResponse')
        .mockResolvedValue({ data: 'json data' })
      
      const result = await protobufClient.fetchWithProtobuf('/api/test', {}, 'api.TestMessage')
      
      expect(protobufClient.handleNonProtobufResponse).toHaveBeenCalled()
      expect(result).toEqual({ data: 'json data' })
    })
    
    it('should handle fetch errors', async () => {
      // Mock initialize to set protobufSupported to true
      vi.spyOn(protobufClient, 'initialize').mockResolvedValue(true)
      protobufClient.protobufSupported = true
      
      // Mock checkServerProtobufSupport to return true
      vi.spyOn(protobufClient, 'checkServerProtobufSupport').mockResolvedValue(true)
      
      // Mock fetch to throw an error
      global.fetch.mockRejectedValue(new Error('Network error'))
      
      // Mock fallbackToJSON
      vi.spyOn(protobufClient, 'fallbackToJSON')
        .mockResolvedValue({ data: 'json data' })
      
      const result = await protobufClient.fetchWithProtobuf('/api/test', {}, 'api.TestMessage')
      
      expect(protobufClient.fallbackToJSON).toHaveBeenCalledWith('/api/test', {})
      expect(result).toEqual({ data: 'json data' })
    })
  })
  
  describe('decodeProtobufResponse', () => {
    it('should decode protobuf response', async () => {
      // Mock initialize
      vi.spyOn(protobufClient, 'initialize').mockResolvedValue(true)
      
      // Set protoDefinitions
      protobufClient.protoDefinitions = {
        lookupType: vi.fn().mockReturnValue({
          decode: vi.fn().mockReturnValue({ field: 'value' }),
          toObject: vi.fn().mockReturnValue({ field: 'value', converted: true })
        })
      }
      
      const buffer = new ArrayBuffer(10)
      const result = await protobufClient.decodeProtobufResponse(buffer, 'api.TestMessage')
      
      expect(protobufClient.protoDefinitions.lookupType).toHaveBeenCalledWith('api.TestMessage')
      expect(result).toEqual({ field: 'value', converted: true })
    })
    
    it('should throw error when protobuf definitions are not loaded', async () => {
      // Mock initialize
      vi.spyOn(protobufClient, 'initialize').mockResolvedValue(true)
      
      // Set protoDefinitions to null
      protobufClient.protoDefinitions = null
      
      const buffer = new ArrayBuffer(10)
      
      await expect(protobufClient.decodeProtobufResponse(buffer, 'api.TestMessage'))
        .rejects.toThrow('Protobuf definitions not loaded')
    })
    
    it('should handle decoding errors', async () => {
      // Mock initialize
      vi.spyOn(protobufClient, 'initialize').mockResolvedValue(true)
      
      // Set protoDefinitions with error
      protobufClient.protoDefinitions = {
        lookupType: vi.fn().mockReturnValue({
          decode: vi.fn().mockImplementation(() => {
            throw new Error('Decoding error')
          })
        })
      }
      
      const buffer = new ArrayBuffer(10)
      
      await expect(protobufClient.decodeProtobufResponse(buffer, 'api.TestMessage'))
        .rejects.toThrow('Failed to decode protobuf response: Decoding error')
    })
  })
  
  describe('handleNonProtobufResponse', () => {
    it('should handle JSON responses', async () => {
      const response = {
        headers: {
          get: vi.fn().mockReturnValue('application/json')
        },
        json: vi.fn().mockResolvedValue({ data: 'json data' })
      }
      
      const result = await protobufClient.handleNonProtobufResponse(response)
      
      expect(result).toEqual({
        data: 'json data',
        _protobuf: {
          format: 'json',
          fallbackReason: 'server_returned_json'
        }
      })
    })
    
    it('should handle text responses', async () => {
      const response = {
        headers: {
          get: vi.fn().mockReturnValue('text/plain')
        },
        text: vi.fn().mockResolvedValue('text data')
      }
      
      const result = await protobufClient.handleNonProtobufResponse(response)
      
      expect(result).toEqual({
        data: 'text data',
        _protobuf: {
          format: 'text',
          fallbackReason: 'unknown_content_type'
        }
      })
    })
    
    it('should handle JSON parsing errors', async () => {
      const response = {
        headers: {
          get: vi.fn().mockReturnValue('application/json')
        },
        json: vi.fn().mockRejectedValue(new Error('JSON parse error'))
      }
      
      await expect(protobufClient.handleNonProtobufResponse(response))
        .rejects.toThrow('Failed to parse JSON response: JSON parse error')
    })
  })
  
  describe('fallbackToJSON', () => {
    it('should make a JSON request', async () => {
      // Mock fetch to return a JSON response
      global.fetch.mockResolvedValue({
        ok: true,
        json: vi.fn().mockResolvedValue({ data: 'json data' })
      })
      
      const result = await protobufClient.fallbackToJSON('/api/test', {})
      
      expect(global.fetch).toHaveBeenCalledWith('/api/test', {
        headers: {
          'Accept': 'application/json'
        }
      })
      
      expect(result).toEqual({
        data: 'json data',
        _protobuf: {
          processingTime: 100,
          payloadSize: expect.any(Number),
          format: 'json',
          fallbackReason: 'unknown'
        }
      })
    })
    
    it('should handle fetch errors', async () => {
      // Mock fetch to throw an error
      global.fetch.mockRejectedValue(new Error('Network error'))
      
      await expect(protobufClient.fallbackToJSON('/api/test', {}))
        .rejects.toThrow('Network error')
    })
    
    it('should handle non-ok responses', async () => {
      // Mock fetch to return a non-ok response
      global.fetch.mockResolvedValue({
        ok: false,
        status: 404,
        statusText: 'Not Found'
      })
      
      await expect(protobufClient.fallbackToJSON('/api/test', {}))
        .rejects.toThrow('HTTP 404: Not Found')
    })
  })
  
  describe('getFallbackReason', () => {
    it('should return client_unsupported when protobuf is not supported', () => {
      protobufClient.protobufSupported = false
      
      expect(protobufClient.getFallbackReason()).toBe('client_unsupported')
    })
    
    it('should return client_disabled when protobuf is disabled', () => {
      protobufClient.protobufSupported = true
      protobufClient.protobufEnabled = false
      
      expect(protobufClient.getFallbackReason()).toBe('client_disabled')
    })
    
    it('should return server_unsupported when server does not support protobuf', () => {
      protobufClient.protobufSupported = true
      protobufClient.protobufEnabled = true
      protobufClient.serverSupportsProtobuf = false
      
      expect(protobufClient.getFallbackReason()).toBe('server_unsupported')
    })
    
    it('should return unknown for other cases', () => {
      protobufClient.protobufSupported = true
      protobufClient.protobufEnabled = true
      protobufClient.serverSupportsProtobuf = null
      
      expect(protobufClient.getFallbackReason()).toBe('unknown')
    })
  })
  
  describe('setProtobufEnabled', () => {
    it('should enable protobuf support', () => {
      protobufClient.protobufEnabled = false
      
      protobufClient.setProtobufEnabled(true)
      
      expect(protobufClient.protobufEnabled).toBe(true)
    })
    
    it('should disable protobuf support', () => {
      protobufClient.protobufEnabled = true
      
      protobufClient.setProtobufEnabled(false)
      
      expect(protobufClient.protobufEnabled).toBe(false)
    })
  })
  
  describe('getStatus', () => {
    it('should return client status', () => {
      protobufClient.protobufSupported = true
      protobufClient.protobufEnabled = true
      protobufClient.initialized = true
      protobufClient.serverSupportsProtobuf = true
      protobufClient.protoDefinitions = {}
      
      const status = protobufClient.getStatus()
      
      expect(status).toEqual({
        protobufSupported: true,
        protobufEnabled: true,
        initialized: true,
        serverSupportsProtobuf: true,
        definitionsLoaded: true
      })
    })
  })
})