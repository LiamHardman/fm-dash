import { ref } from 'vue'
import { useErrorHandling } from './useErrorHandling'
import protobufClient from '../utils/protobufClient'

// Determine the base URL at runtime
const getBaseURL = () => {
  if (window.APP_CONFIG?.API_ENDPOINT) {
    return window.APP_CONFIG.API_ENDPOINT
  }
  // Fallback for development or when config is not injected
  return '/api'
}

export function useApi(initialBaseURL) {
  const baseURL = ref(initialBaseURL || getBaseURL())
  const { handleFetchError, withRetry, safeAsync } = useErrorHandling()
  const isLoading = ref(false)
  const abortController = ref(null)
  
  // Track if protobuf has been initialized
  const protobufInitialized = ref(false)
  const protobufSupported = ref(false)

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

  // Initialize protobuf support
  const initializeProtobuf = async () => {
    if (protobufInitialized.value) {
      return protobufSupported.value
    }
    
    try {
      // Check if protobuf is supported
      const supported = await protobufClient.initialize()
      protobufSupported.value = supported
      protobufInitialized.value = true
      return supported
    } catch (error) {
      protobufSupported.value = false
      protobufInitialized.value = true
      return false
    }
  }

  // Generic API request method with protobuf support
  const request = async (url, options = {}, messageType = null) => {
    const { url: fullUrl, config } = createRequest(url, options)

    try {
      isLoading.value = true
      
      // Check if we should use protobuf
      if (messageType && await initializeProtobuf()) {
        try {
          // Try to use protobuf
          return await protobufClient.fetchWithProtobuf(fullUrl, config, messageType)
        } catch (protobufError) {
          // Fall back to JSON on protobuf error
          console.warn('Protobuf request failed, falling back to JSON:', protobufError)
        }
      }
      
      // Standard JSON request
      const response = await fetch(fullUrl, config)

      if (!response.ok) {
        await handleFetchError(response, { url: fullUrl, method: config.method })
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }

      // Handle different content types
      const contentType = response.headers.get('content-type')
      if (contentType?.includes('application/json')) {
        const jsonData = await response.json()
        return {
          ...jsonData,
          _protobuf: {
            format: 'json',
            payloadSize: JSON.stringify(jsonData).length,
            fallbackReason: 'json_request'
          }
        }
      }
      return await response.text()
    } catch (error) {
      if (error.name === 'AbortError') {
        return null
      }
      throw error
    } finally {
      isLoading.value = false
      abortController.value = null
    }
  }

  // GET request with optional protobuf support
  const get = async (url, params = {}, messageType = null) => {
    const queryString = new URLSearchParams(params).toString()
    const fullUrl = queryString ? `${url}?${queryString}` : url

    return request(fullUrl, {
      method: 'GET'
    }, messageType)
  }

  // POST request with optional protobuf support
  const post = async (url, data, options = {}, messageType = null) => {
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

    return request(url, config, messageType)
  }

  // PUT request with optional protobuf support
  const put = async (url, data, messageType = null) => {
    return request(url, {
      method: 'PUT',
      body: JSON.stringify(data)
    }, messageType)
  }

  // DELETE request with optional protobuf support
  const del = async (url, messageType = null) => {
    return request(url, {
      method: 'DELETE'
    }, messageType)
  }

  // File upload with progress
  const uploadFile = async (url, file, onProgress = null) => {
    const formData = new FormData()
    formData.append('playerFile', file)

    const { url: fullUrl, config } = createRequest(url, {
      method: 'POST',
      body: formData
    })

    if (config.headers) {
      delete config.headers['Content-Type']
    }

    try {
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
            if (xhr.status >= 200 && xhr.status < 300) {
              try {
                const response = JSON.parse(xhr.responseText)
                resolve({
                  ...response,
                  _protobuf: {
                    format: 'json',
                    payloadSize: xhr.responseText.length,
                    fallbackReason: 'form_data_upload'
                  }
                })
              } catch (_error) {
                resolve(xhr.responseText)
              }
            } else {
              reject(new Error(`HTTP ${xhr.status}: ${xhr.statusText}`))
            }
          })

          xhr.addEventListener('error', () => {
            reject(new Error('Upload failed'))
          })

          xhr.open('POST', fullUrl)

          // Add any custom headers except Content-Type
          if (config.headers) {
            Object.entries(config.headers).forEach(([key, value]) => {
              if (key.toLowerCase() !== 'content-type' && value !== undefined) {
                xhr.setRequestHeader(key, value)
              }
            })
          }

          xhr.send(formData)
        })
      }
      // Use fetch for simple upload
      const response = await fetch(fullUrl, config)

      if (!response.ok) {
        await handleFetchError(response, { url: fullUrl, method: 'POST' })
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }

      const jsonData = await response.json()
      return {
        ...jsonData,
        _protobuf: {
          format: 'json',
          payloadSize: JSON.stringify(jsonData).length,
          fallbackReason: 'form_data_upload'
        }
      }
    } finally {
      isLoading.value = false
      abortController.value = null
    }
  }

  // Cancel current request
  const cancel = () => {
    if (abortController.value) {
      abortController.value.abort()
      abortController.value = null
    }
  }

  // Batch requests with error handling
  const batch = async (requests) => {
    const results = []
    const errors = []

    for (const req of requests) {
      try {
        const result = await request(
          req.url, 
          req.options, 
          req.messageType // Support for protobuf message type
        )
        results.push({ success: true, data: result })
      } catch (error) {
        errors.push({ success: false, error })
        results.push({ success: false, error })
      }
    }

    return { results, errors: errors.length > 0 ? errors : null }
  }

  // Health check endpoint
  const healthCheck = async () => {
    try {
      const response = await get('/health', {}, 'api.GenericResponse')
      return { healthy: true, data: response }
    } catch (error) {
      return { healthy: false, error }
    }
  }
  
  // Get protobuf client status
  const getProtobufStatus = () => {
    return {
      initialized: protobufInitialized.value,
      supported: protobufSupported.value,
      clientStatus: protobufClient.getStatus()
    }
  }
  
  // Enable or disable protobuf support
  const setProtobufEnabled = (enabled) => {
    protobufClient.setProtobufEnabled(enabled)
  }

  return {
    isLoading,
    protobufSupported,

    request,
    get,
    post,
    put,
    delete: del,
    uploadFile,
    cancel,
    batch,
    healthCheck,
    
    // Protobuf specific methods
    getProtobufStatus,
    setProtobufEnabled,
    initializeProtobuf,

    safeAsync,
    withRetry
  }
}