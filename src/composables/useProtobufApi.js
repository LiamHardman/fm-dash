/**
 * useProtobufApi - Composable for making API requests with protobuf support
 * 
 * This composable provides:
 * - Enhanced fetch methods with protobuf support
 * - Graceful fallback to JSON for unsupported clients
 * - Performance tracking for API operations
 * - Consistent error handling
 */

import { ref, reactive } from 'vue'
import { useErrorHandling } from './useErrorHandling'
import protobufClient from '../utils/protobufClient'
import logger from '../utils/logger'

// Determine the base URL at runtime
const getBaseURL = () => {
  if (window.APP_CONFIG?.API_ENDPOINT) {
    return window.APP_CONFIG.API_ENDPOINT
  }
  // Fallback for development or when config is not injected
  return '/api'
}

export function useProtobufApi(initialBaseURL) {
  const baseURL = ref(initialBaseURL || getBaseURL())
  const { handleFetchError, withRetry, safeAsync } = useErrorHandling()
  const isLoading = ref(false)
  const abortController = ref(null)
  
  // Performance metrics
  const metrics = reactive({
    lastRequestTime: 0,
    averageRequestTime: 0,
    totalRequests: 0,
    protobufRequests: 0,
    jsonRequests: 0,
    failedRequests: 0,
    averagePayloadSize: 0,
    compressionRatio: 0
  })

  // Default request configuration
  const defaultConfig = {
    headers: {
      'Content-Type': 'application/json'
    },
    credentials: 'same-origin'
  }

  // Create request with abort capability
  const createRequest = (url, options = {}) => {
    if (abortController.value) {
      abortController.value.abort()
    }

    abortController.value = new AbortController()

    const config = {
      ...defaultConfig,
      ...options,
      signal: abortController.value.signal
    }

    return { url: `${baseURL.value}${url}`, config }
  }

  /**
   * Make a protobuf-aware API request
   * @param {string} url - API endpoint
   * @param {Object} options - Request options
   * @param {string} messageType - Protobuf message type to decode
   */
  const protobufRequest = async (url, options = {}, messageType) => {
    const { url: fullUrl, config } = createRequest(url, options)
    const requestStartTime = performance.now()
    let errorDetails = {}

    try {
      isLoading.value = true
      
      // Use protobufClient to make the request
      const response = await protobufClient.fetchWithProtobuf(fullUrl, config, messageType)
      
      // Update metrics
      const endTime = performance.now()
      updateMetrics(response, endTime - requestStartTime)
      
      // Log successful requests with format information
      if (response._protobuf) {
        logger.info(`API request succeeded (${response._protobuf.format})`, {
          url,
          messageType,
          format: response._protobuf.format,
          processingTime: response._protobuf.processingTime,
          payloadSize: response._protobuf.payloadSize,
          fallbackReason: response._protobuf.fallbackReason || null,
          retryCount: response._protobuf.retryCount || 0
        })
        
        // Track fallbacks for monitoring
        if (response._protobuf.fallbackReason) {
          metrics.fallbacks = metrics.fallbacks || {}
          metrics.fallbacks[response._protobuf.fallbackReason] = 
            (metrics.fallbacks[response._protobuf.fallbackReason] || 0) + 1
        }
      }
      
      return response
    } catch (error) {
      if (error.name === 'AbortError') {
        logger.info('Request aborted by user', { url })
        return null
      }
      
      // Update failed request count
      metrics.failedRequests++
      
      // Collect detailed error information for logging
      errorDetails = {
        url,
        messageType,
        errorType: error.name,
        errorMessage: error.message,
        requestDuration: performance.now() - requestStartTime
      }
      
      // Add specific error details based on error type
      if (error.name === 'ProtobufDecodingError') {
        errorDetails.bufferSize = error.bufferSize
        errorDetails.originalError = error.originalError?.message
      } else if (error.name === 'ProtobufApiError') {
        errorDetails.status = error.status
        errorDetails.errorCode = error.errorCode
        errorDetails.details = error.details
      }
      
      // Log detailed error information
      logger.error('API request failed', errorDetails)
      
      // Track error types for monitoring
      metrics.errorsByType = metrics.errorsByType || {}
      metrics.errorsByType[error.name] = (metrics.errorsByType[error.name] || 0) + 1
      
      // Enhance error with additional context
      error.requestUrl = url
      error.messageType = messageType
      error.requestTime = new Date().toISOString()
      
      throw error
    } finally {
      isLoading.value = false
      abortController.value = null
    }
  }

  /**
   * Update performance metrics
   * @param {Object} response - API response
   * @param {number} requestTime - Request time in milliseconds
   */
  const updateMetrics = (response, requestTime) => {
    metrics.lastRequestTime = requestTime
    metrics.totalRequests++
    
    // Update average request time
    metrics.averageRequestTime = 
      (metrics.averageRequestTime * (metrics.totalRequests - 1) + requestTime) / 
      metrics.totalRequests
    
    // Update format-specific metrics
    if (response._protobuf) {
      if (response._protobuf.format === 'protobuf') {
        metrics.protobufRequests++
      } else {
        metrics.jsonRequests++
      }
      
      // Update payload size metrics
      if (response._protobuf.payloadSize) {
        metrics.averagePayloadSize = 
          (metrics.averagePayloadSize * (metrics.totalRequests - 1) + response._protobuf.payloadSize) / 
          metrics.totalRequests
      }
    }
  }

  /**
   * Make a GET request with protobuf support
   * @param {string} url - API endpoint
   * @param {Object} params - Query parameters
   * @param {string} messageType - Protobuf message type
   */
  const get = async (url, params = {}, messageType) => {
    const queryString = new URLSearchParams(params).toString()
    const fullUrl = queryString ? `${url}?${queryString}` : url

    return protobufRequest(fullUrl, {
      method: 'GET'
    }, messageType)
  }

  /**
   * Make a POST request with protobuf support
   * @param {string} url - API endpoint
   * @param {Object} data - Request body
   * @param {Object} options - Additional options
   * @param {string} messageType - Protobuf message type
   */
  const post = async (url, data, options = {}, messageType) => {
    const config = {
      method: 'POST',
      ...options
    }

    if (data instanceof FormData) {
      // Remove Content-Type header for FormData (let browser set it)
      config.headers = undefined
    } else if (data) {
      config.body = JSON.stringify(data)
    }

    return protobufRequest(url, config, messageType)
  }

  /**
   * Make a PUT request with protobuf support
   * @param {string} url - API endpoint
   * @param {Object} data - Request body
   * @param {string} messageType - Protobuf message type
   */
  const put = async (url, data, messageType) => {
    return protobufRequest(url, {
      method: 'PUT',
      body: JSON.stringify(data)
    }, messageType)
  }

  /**
   * Make a DELETE request with protobuf support
   * @param {string} url - API endpoint
   * @param {string} messageType - Protobuf message type
   */
  const del = async (url, messageType) => {
    return protobufRequest(url, {
      method: 'DELETE'
    }, messageType)
  }

  /**
   * Get player data with protobuf support
   * @param {string} datasetId - Dataset ID
   * @param {Object} params - Query parameters
   */
  const getPlayerData = async (datasetId, params = {}) => {
    if (!datasetId) {
      throw new Error('Dataset ID is required.')
    }

    try {
      const url = `/players/${datasetId}`
      return await get(url, params, 'api.PlayerDataResponse')
    } catch (error) {
      logger.error('Error fetching player data:', error)
      throw error
    }
  }

  /**
   * Get available roles with protobuf support
   */
  const getRoles = async () => {
    try {
      return await get('/roles', {}, 'api.RolesResponse')
    } catch (error) {
      logger.error('Error fetching roles:', error)
      throw error
    }
  }

  /**
   * Get configuration with protobuf support
   */
  const getConfig = async () => {
    try {
      return await get('/config', {}, 'api.GenericResponse')
    } catch (error) {
      logger.error('Error fetching config:', error)
      // Return default config on error
      return {
        maxUploadSizeMB: 15,
        maxUploadSizeBytes: 15 * 1024 * 1024,
        useScaledRatings: true,
        datasetRetentionDays: 30
      }
    }
  }

  /**
   * Upload player file (always uses JSON as FormData is not protobuf compatible)
   * @param {FormData} formData - Form data with file
   * @param {number} maxSizeBytes - Maximum file size
   * @param {Function} onProgress - Progress callback
   */
  const uploadPlayerFile = async (formData, maxSizeBytes = 15 * 1024 * 1024, onProgress = null) => {
    const file = formData.get('playerFile')
    if (!file) {
      throw new Error('No file found in form data')
    }

    try {
      // File uploads always use JSON
      const { url: fullUrl, config } = createRequest('/upload', {
        method: 'POST',
        body: formData
      })
      
      isLoading.value = true
      
      // Use XMLHttpRequest for progress tracking
      if (onProgress) {
        return new Promise((resolve, reject) => {
          const xhr = new XMLHttpRequest()

          xhr.upload.addEventListener('progress', event => {
            if (event.lengthComputable) {
              const progress = (event.loaded / event.total) * 100
              onProgress(progress)
            }
          })

          xhr.addEventListener('load', () => {
            isLoading.value = false
            if (xhr.status >= 200 && xhr.status < 300) {
              try {
                const response = JSON.parse(xhr.responseText)
                metrics.jsonRequests++
                metrics.totalRequests++
                resolve(response)
              } catch (_error) {
                metrics.jsonRequests++
                metrics.totalRequests++
                resolve(xhr.responseText)
              }
            } else {
              metrics.failedRequests++
              reject(new Error(`HTTP ${xhr.status}: ${xhr.statusText}`))
            }
          })

          xhr.addEventListener('error', () => {
            isLoading.value = false
            metrics.failedRequests++
            reject(new Error('Upload failed'))
          })

          xhr.open('POST', fullUrl)
          xhr.send(formData)
        })
      }
      
      // Use fetch for simple upload
      const response = await fetch(fullUrl, config)
      
      if (!response.ok) {
        await handleFetchError(response, { url: fullUrl, method: 'POST' })
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }
      
      const jsonResponse = await response.json()
      metrics.jsonRequests++
      metrics.totalRequests++
      return jsonResponse
    } catch (error) {
      logger.error('Upload error:', error)

      if (error.message?.includes('413') || error.message?.includes('too large')) {
        const maxSizeMB = Math.round(maxSizeBytes / (1024 * 1024))
        const newError = new Error(`File too large. Maximum size allowed is ${maxSizeMB}MB.`)
        newError.status = 413
        throw newError
      }

      throw error
    } finally {
      isLoading.value = false
      abortController.value = null
    }
  }

  /**
   * Cancel current request
   */
  const cancel = () => {
    if (abortController.value) {
      abortController.value.abort()
      abortController.value = null
    }
  }

  /**
   * Get client status and metrics
   */
  const getClientStatus = () => {
    return {
      ...protobufClient.getStatus(),
      metrics: { ...metrics }
    }
  }

  /**
   * Enable or disable protobuf support
   * @param {boolean} enabled - Whether protobuf should be enabled
   */
  const setProtobufEnabled = (enabled) => {
    protobufClient.setProtobufEnabled(enabled)
  }

  return {
    isLoading,
    metrics,

    // Enhanced API methods
    get,
    post,
    put,
    delete: del,
    getPlayerData,
    getRoles,
    getConfig,
    uploadPlayerFile,
    cancel,
    
    // Protobuf specific methods
    getClientStatus,
    setProtobufEnabled,

    // Error handling utilities
    safeAsync,
    withRetry
  }
}