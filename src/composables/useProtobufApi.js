/**
 * useProtobufApi - Composable for making API requests with protobuf support
 */

import { ref, computed } from 'vue'
import protobufClient from '../utils/protobufClient'
import logger from '../utils/logger.js'

// Performance metrics
const metrics = {
  totalRequests: 0,
  protobufRequests: 0,
  jsonRequests: 0,
  failedRequests: 0,
  averageRequestTime: 0,
  averagePayloadSize: 0,
  lastRequestTime: 0,
  fallbacks: {},
  errorsByType: {}
}

// Get base URL from configuration
const getBaseURL = () => {
  try {
    if (window.APP_CONFIG?.API_BASE_URL) {
      return window.APP_CONFIG.API_BASE_URL
    }
    return '/api'
  } catch (error) {
    logger.warn('Error getting API base URL, using default:', error)
    return '/api'
  }
}

export function useProtobufApi(initialBaseURL) {
  const baseURL = ref(initialBaseURL || getBaseURL())
  const isLoading = ref(false)
  const abortController = ref(null)
  const errorDetails = ref(null)

  /**
   * Create a request with proper error handling
   */
  const createRequest = (url, options = {}) => {
    // Cancel any existing request
    if (abortController.value) {
      abortController.value.abort()
    }

    // Create new abort controller
    abortController.value = new AbortController()

    return {
      ...options,
      signal: abortController.value.signal
    }
  }

  /**
   * Make a protobuf request with fallback to JSON
   */
  const protobufRequest = async (url, options = {}, messageType) => {
    const requestStartTime = performance.now()
    const config = createRequest(url, options)
    
    try {
      isLoading.value = true
      
      // Use protobufClient to make the request
      const response = await protobufClient.fetchWithProtobuf(url, config, messageType)
      
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
      const errorDetails = {
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
      const url = `/api/players/${datasetId}`
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
      return await get('/api/roles', {}, 'api.RolesResponse')
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
      return await get('/api/config', {}, 'api.GenericResponse')
    } catch (error) {
      logger.error('Error fetching config:', error)
      throw error
    }
  }

  /**
   * Upload player file with progress tracking
   * @param {FormData} formData - File data
   * @param {number} maxSizeBytes - Maximum file size in bytes
   * @param {Function} onProgress - Progress callback
   */
  const uploadPlayerFile = async (formData, maxSizeBytes = 15 * 1024 * 1024, onProgress = null) => {
    if (!formData) {
      throw new Error('FormData is required for file upload')
    }

    try {
      // Check file size
      const file = formData.get('file')
      if (file && file.size > maxSizeBytes) {
        throw new Error(`File size exceeds maximum allowed size of ${maxSizeBytes / (1024 * 1024)}MB`)
      }

      const config = {
        method: 'POST',
        body: formData
      }

      // Add progress tracking if callback provided
      if (onProgress && typeof XMLHttpRequest !== 'undefined') {
        return new Promise((resolve, reject) => {
          const xhr = new XMLHttpRequest()
          
          xhr.upload.addEventListener('progress', (event) => {
            if (event.lengthComputable) {
              const percentComplete = (event.loaded / event.total) * 100
              onProgress(percentComplete)
            }
          })

          xhr.addEventListener('load', () => {
            if (xhr.status >= 200 && xhr.status < 300) {
              try {
                const response = JSON.parse(xhr.responseText)
                resolve({
                  ...response,
                  _protobuf: {
                    format: 'json',
                    fallbackReason: 'file_upload'
                  }
                })
              } catch (error) {
                reject(new Error('Invalid JSON response from server'))
              }
            } else {
              reject(new Error(`Upload failed with status ${xhr.status}`))
            }
          })

          xhr.addEventListener('error', () => {
            reject(new Error('Upload failed due to network error'))
          })

          xhr.open('POST', '/api/upload')
          xhr.send(formData)
        })
      }

      // Use standard fetch for uploads without progress tracking
      return await protobufRequest('/api/upload', config, 'api.GenericResponse')
    } catch (error) {
      logger.error('Error uploading player file:', error)
      throw error
    }
  }

  /**
   * Cancel the current request
   */
  const cancel = () => {
    if (abortController.value) {
      abortController.value.abort()
      logger.info('Request cancelled by user')
    }
  }

  /**
   * Get client status and capabilities
   */
  const getClientStatus = () => {
    return {
      ...protobufClient.getStatus(),
      baseURL: baseURL.value,
      isLoading: isLoading.value,
      metrics: { ...metrics }
    }
  }

  /**
   * Enable or disable protobuf support
   */
  const setProtobufEnabled = (enabled) => {
    protobufClient.setProtobufEnabled(enabled)
  }

  return {
    // State
    isLoading: computed(() => isLoading.value),
    baseURL: computed(() => baseURL.value),
    
    // Methods
    get,
    post,
    put,
    del,
    getPlayerData,
    getRoles,
    getConfig,
    uploadPlayerFile,
    cancel,
    getClientStatus,
    setProtobufEnabled,
    
    // Metrics
    metrics: computed(() => ({ ...metrics }))
  }
} 