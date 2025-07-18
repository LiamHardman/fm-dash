/**
 * Unit tests for useProtobufApi composable
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { useProtobufApi } from './useProtobufApi'

// Mock the protobufClient module
vi.mock('../utils/protobufClient', () => ({
  default: {
    fetchWithProtobuf: vi.fn(),
    getStatus: vi.fn().mockReturnValue({
      protobufSupported: true,
      protobufEnabled: true,
      initialized: true,
      serverSupportsProtobuf: true,
      definitionsLoaded: true
    }),
    setProtobufEnabled: vi.fn(),
    initialize: vi.fn().mockResolvedValue(true)
  }
}))

// Mock the logger
vi.mock('../utils/logger', () => ({
  default: {
    error: vi.fn(),
    warn: vi.fn(),
    info: vi.fn()
  }
}))

// Mock the error handling composable
vi.mock('./useErrorHandling', () => ({
  useErrorHandling: vi.fn().mockReturnValue({
    handleFetchError: vi.fn(),
    withRetry: vi.fn(fn => fn),
    safeAsync: vi.fn(fn => fn)
  })
}))

// Import the mocked module
import protobufClient from '../utils/protobufClient'

describe('useProtobufApi', () => {
  
  // Original window.APP_CONFIG
  const originalAppConfig = window.APP_CONFIG
  
  beforeEach(() => {
    // Mock window.APP_CONFIG
    window.APP_CONFIG = {
      API_ENDPOINT: 'https://api.example.com'
    }
    
    // Mock performance.now
    global.performance = {
      now: vi.fn().mockReturnValueOnce(0).mockReturnValueOnce(100)
    }
    
    // Reset protobufClient mocks
    vi.clearAllMocks()
  })
  
  afterEach(() => {
    // Restore window.APP_CONFIG
    window.APP_CONFIG = originalAppConfig
    
    // Clear all mocks
    vi.clearAllMocks()
  })
  
  describe('initialization', () => {
    it('should initialize with the correct base URL', () => {
      const api = useProtobufApi()
      
      expect(api.isLoading.value).toBe(false)
      expect(api.metrics).toEqual({
        lastRequestTime: 0,
        averageRequestTime: 0,
        totalRequests: 0,
        protobufRequests: 0,
        jsonRequests: 0,
        failedRequests: 0,
        averagePayloadSize: 0,
        compressionRatio: 0
      })
    })
    
    it('should use the provided base URL', () => {
      const api = useProtobufApi('https://custom-api.example.com')
      
      // We can't directly test the baseURL ref, but we can test it indirectly
      // by making a request and checking the URL
      api.get('/test')
      
      expect(protobufClient.fetchWithProtobuf).toHaveBeenCalledWith(
        'https://custom-api.example.com/test',
        expect.objectContaining({
          method: 'GET'
        }),
        undefined
      )
    })
  })
  
  describe('get', () => {
    it('should make a GET request with protobuf support', async () => {
      const api = useProtobufApi()
      
      // Mock the protobufClient response
      protobufClient.fetchWithProtobuf.mockResolvedValue({
        data: 'test data',
        _protobuf: {
          format: 'protobuf',
          processingTime: 50,
          payloadSize: 100
        }
      })
      
      const result = await api.get('/test', { param: 'value' }, 'api.TestMessage')
      
      expect(protobufClient.fetchWithProtobuf).toHaveBeenCalledWith(
        'https://api.example.com/test?param=value',
        expect.objectContaining({
          method: 'GET'
        }),
        'api.TestMessage'
      )
      
      expect(result).toEqual({
        data: 'test data',
        _protobuf: {
          format: 'protobuf',
          processingTime: 50,
          payloadSize: 100
        }
      })
      
      // Check metrics
      expect(api.metrics.totalRequests).toBe(1)
      expect(api.metrics.protobufRequests).toBe(1)
      expect(api.metrics.jsonRequests).toBe(0)
    })
    
    it('should handle JSON fallback responses', async () => {
      const api = useProtobufApi()
      
      // Mock the protobufClient response with JSON fallback
      protobufClient.fetchWithProtobuf.mockResolvedValue({
        data: 'json data',
        _protobuf: {
          format: 'json',
          fallbackReason: 'client_unsupported',
          processingTime: 50,
          payloadSize: 200
        }
      })
      
      const result = await api.get('/test')
      
      expect(result).toEqual({
        data: 'json data',
        _protobuf: {
          format: 'json',
          fallbackReason: 'client_unsupported',
          processingTime: 50,
          payloadSize: 200
        }
      })
      
      // Check metrics
      expect(api.metrics.totalRequests).toBe(1)
      expect(api.metrics.protobufRequests).toBe(0)
      expect(api.metrics.jsonRequests).toBe(1)
    })
    
    it('should handle request errors', async () => {
      const api = useProtobufApi()
      
      // Mock the protobufClient to throw an error
      protobufClient.fetchWithProtobuf.mockRejectedValue(new Error('Network error'))
      
      await expect(api.get('/test')).rejects.toThrow('Network error')
      
      // Check metrics
      expect(api.metrics.failedRequests).toBe(1)
    })
  })
  
  describe('post', () => {
    it('should make a POST request with JSON body', async () => {
      const api = useProtobufApi()
      
      await api.post('/test', { data: 'test' }, {}, 'api.TestMessage')
      
      expect(protobufClient.fetchWithProtobuf).toHaveBeenCalledWith(
        'https://api.example.com/test',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify({ data: 'test' })
        }),
        'api.TestMessage'
      )
    })
    
    it('should handle FormData correctly', async () => {
      const api = useProtobufApi()
      
      const formData = new FormData()
      formData.append('file', 'test-file')
      
      await api.post('/test', formData)
      
      expect(protobufClient.fetchWithProtobuf).toHaveBeenCalledWith(
        'https://api.example.com/test',
        expect.objectContaining({
          method: 'POST',
          body: formData,
          headers: undefined
        }),
        undefined
      )
    })
  })
  
  describe('put', () => {
    it('should make a PUT request', async () => {
      const api = useProtobufApi()
      
      await api.put('/test', { data: 'test' }, 'api.TestMessage')
      
      expect(protobufClient.fetchWithProtobuf).toHaveBeenCalledWith(
        'https://api.example.com/test',
        expect.objectContaining({
          method: 'PUT',
          body: JSON.stringify({ data: 'test' })
        }),
        'api.TestMessage'
      )
    })
  })
  
  describe('delete', () => {
    it('should make a DELETE request', async () => {
      const api = useProtobufApi()
      
      await api.delete('/test', 'api.TestMessage')
      
      expect(protobufClient.fetchWithProtobuf).toHaveBeenCalledWith(
        'https://api.example.com/test',
        expect.objectContaining({
          method: 'DELETE'
        }),
        'api.TestMessage'
      )
    })
  })
  
  describe('getPlayerData', () => {
    it('should fetch player data with the correct message type', async () => {
      const api = useProtobufApi()
      
      // Mock the get method
      const getSpy = vi.spyOn(api, 'get').mockResolvedValue({ players: [] })
      
      await api.getPlayerData('dataset-123', { position: 'GK' })
      
      expect(getSpy).toHaveBeenCalledWith(
        '/api/players/dataset-123',
        { position: 'GK' },
        'api.PlayerDataResponse'
      )
    })
    
    it('should throw error when dataset ID is missing', async () => {
      const api = useProtobufApi()
      
      await expect(api.getPlayerData()).rejects.toThrow('Dataset ID is required.')
    })
    
    it('should handle errors', async () => {
      const api = useProtobufApi()
      
      // Mock the get method to throw an error
      vi.spyOn(api, 'get').mockRejectedValue(new Error('API error'))
      
      await expect(api.getPlayerData('dataset-123')).rejects.toThrow('API error')
    })
  })
  
  describe('getRoles', () => {
    it('should fetch roles with the correct message type', async () => {
      const api = useProtobufApi()
      
      // Mock the get method
      const getSpy = vi.spyOn(api, 'get').mockResolvedValue({ roles: [] })
      
      await api.getRoles()
      
      expect(getSpy).toHaveBeenCalledWith(
        '/api/roles',
        {},
        'api.RolesResponse'
      )
    })
  })
  
  describe('getConfig', () => {
    it('should fetch config with the correct message type', async () => {
      const api = useProtobufApi()
      
      // Mock the get method
      const getSpy = vi.spyOn(api, 'get').mockResolvedValue({ config: {} })
      
      await api.getConfig()
      
      expect(getSpy).toHaveBeenCalledWith(
        '/api/config',
        {},
        'api.GenericResponse'
      )
    })
    
    it('should return default config on error', async () => {
      const api = useProtobufApi()
      
      // Mock the get method to throw an error
      vi.spyOn(api, 'get').mockRejectedValue(new Error('API error'))
      
      const result = await api.getConfig()
      
      expect(result).toEqual({
        maxUploadSizeMB: 15,
        maxUploadSizeBytes: 15 * 1024 * 1024,
        useScaledRatings: true,
        datasetRetentionDays: 30
      })
    })
  })
  
  describe('uploadPlayerFile', () => {
    it('should throw error when file is missing', async () => {
      const api = useProtobufApi()
      
      const formData = new FormData()
      
      await expect(api.uploadPlayerFile(formData)).rejects.toThrow('No file found in form data')
    })
    
    it('should handle file upload with XMLHttpRequest when progress callback is provided', async () => {
      const api = useProtobufApi()
      
      // Mock XMLHttpRequest
      const xhrMock = {
        open: vi.fn(),
        send: vi.fn(),
        setRequestHeader: vi.fn(),
        upload: {
          addEventListener: vi.fn()
        },
        addEventListener: vi.fn()
      }
      
      // Save original XMLHttpRequest
      const originalXHR = global.XMLHttpRequest
      
      // Mock global XMLHttpRequest
      global.XMLHttpRequest = vi.fn(() => xhrMock)
      
      const formData = new FormData()
      formData.append('playerFile', 'test-file')
      
      const onProgress = vi.fn()
      
      // Simulate successful response
      const responsePromise = api.uploadPlayerFile(formData, 15 * 1024 * 1024, onProgress)
      
      // Find the load event listener and call it
      const loadListener = xhrMock.addEventListener.mock.calls.find(call => call[0] === 'load')[1]
      
      // Set up mock response
      xhrMock.status = 200
      xhrMock.responseText = JSON.stringify({ success: true })
      
      // Trigger the load event
      loadListener()
      
      const result = await responsePromise
      
      expect(result).toEqual({ success: true })
      expect(xhrMock.open).toHaveBeenCalledWith('POST', 'https://api.example.com/api/upload')
      expect(xhrMock.send).toHaveBeenCalledWith(formData)
      
      // Restore original XMLHttpRequest
      global.XMLHttpRequest = originalXHR
    })
    
    it('should handle file upload errors with XMLHttpRequest', async () => {
      const api = useProtobufApi()
      
      // Mock XMLHttpRequest
      const xhrMock = {
        open: vi.fn(),
        send: vi.fn(),
        setRequestHeader: vi.fn(),
        upload: {
          addEventListener: vi.fn()
        },
        addEventListener: vi.fn()
      }
      
      // Save original XMLHttpRequest
      const originalXHR = global.XMLHttpRequest
      
      // Mock global XMLHttpRequest
      global.XMLHttpRequest = vi.fn(() => xhrMock)
      
      const formData = new FormData()
      formData.append('playerFile', 'test-file')
      
      const onProgress = vi.fn()
      
      // Simulate error response
      const responsePromise = api.uploadPlayerFile(formData, 15 * 1024 * 1024, onProgress)
      
      // Find the error event listener and call it
      const errorListener = xhrMock.addEventListener.mock.calls.find(call => call[0] === 'error')[1]
      
      // Trigger the error event
      errorListener()
      
      await expect(responsePromise).rejects.toThrow('Upload failed')
      
      // Restore original XMLHttpRequest
      global.XMLHttpRequest = originalXHR
    })
    
    it('should handle file size errors', async () => {
      const api = useProtobufApi()
      
      // Mock fetch to return a 413 error
      global.fetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 413,
        statusText: 'Payload Too Large',
        text: vi.fn().mockResolvedValue('File too large')
      })
      
      const formData = new FormData()
      formData.append('playerFile', 'test-file')
      
      // Mock handleFetchError to throw an error with the 413 status
      const { useErrorHandling } = await import('./useErrorHandling')
      useErrorHandling().handleFetchError.mockImplementation(() => {
        throw new Error('HTTP 413: File too large')
      })
      
      await expect(api.uploadPlayerFile(formData)).rejects.toThrow('File too large. Maximum size allowed is 15MB.')
    })
  })
  
  describe('cancel', () => {
    it('should cancel the current request', () => {
      const api = useProtobufApi()
      
      // Create a mock AbortController
      const abortMock = {
        abort: vi.fn()
      }
      
      // Set the abortController
      api.cancel()
      
      // No abort controller yet, so abort shouldn't be called
      expect(abortMock.abort).not.toHaveBeenCalled()
      
      // Now set an abort controller and try again
      api.abortController = { value: abortMock }
      api.cancel()
      
      expect(abortMock.abort).toHaveBeenCalled()
    })
  })
  
  describe('getClientStatus', () => {
    it('should return client status and metrics', () => {
      const api = useProtobufApi()
      
      const status = api.getClientStatus()
      
      expect(status).toEqual({
        protobufSupported: true,
        protobufEnabled: true,
        initialized: true,
        serverSupportsProtobuf: true,
        definitionsLoaded: true,
        metrics: {
          lastRequestTime: 0,
          averageRequestTime: 0,
          totalRequests: 0,
          protobufRequests: 0,
          jsonRequests: 0,
          failedRequests: 0,
          averagePayloadSize: 0,
          compressionRatio: 0
        }
      })
    })
  })
  
  describe('setProtobufEnabled', () => {
    it('should call protobufClient.setProtobufEnabled', () => {
      const api = useProtobufApi()
      
      api.setProtobufEnabled(false)
      
      expect(protobufClient.setProtobufEnabled).toHaveBeenCalledWith(false)
    })
  })
})