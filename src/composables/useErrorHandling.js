/**
 * Error handling composable for API requests
 */

export function useErrorHandling() {
  /**
   * Handle fetch errors with appropriate responses
   * @param {Response} response - Fetch response object
   * @param {Object} requestInfo - Information about the request
   */
  const handleFetchError = async (response, requestInfo = {}) => {
    let errorMessage = `HTTP ${response.status}: ${response.statusText}`
    
    try {
      // Try to parse error response as JSON
      const errorData = await response.json()
      if (errorData.message) {
        errorMessage = errorData.message
      } else if (errorData.error) {
        errorMessage = errorData.error
      }
    } catch (e) {
      // If not JSON, try to get text
      try {
        const errorText = await response.text()
        if (errorText) {
          errorMessage = errorText
        }
      } catch (textError) {
        // Ignore text parsing errors
      }
    }
    
    const error = new Error(errorMessage)
    error.status = response.status
    error.requestInfo = requestInfo
    
    return error
  }
  
  /**
   * Retry a function with exponential backoff
   * @param {Function} fn - Function to retry
   * @param {Object} options - Retry options
   */
  const withRetry = async (fn, options = {}) => {
    const {
      retries = 3,
      initialDelay = 300,
      maxDelay = 5000,
      factor = 2,
      onRetry = null,
      shouldRetry = () => true
    } = options
    
    let lastError
    
    for (let attempt = 0; attempt <= retries; attempt++) {
      try {
        return await fn()
      } catch (error) {
        lastError = error
        
        // Check if we should retry
        if (attempt >= retries || !shouldRetry(error)) {
          throw error
        }
        
        // Calculate delay with exponential backoff
        const delay = Math.min(initialDelay * Math.pow(factor, attempt), maxDelay)
        
        // Call onRetry callback if provided
        if (onRetry) {
          onRetry(attempt, error, delay)
        }
        
        // Wait before retrying
        await new Promise(resolve => setTimeout(resolve, delay))
      }
    }
    
    throw lastError
  }
  
  /**
   * Safely execute an async function with error handling
   * @param {Function} fn - Async function to execute
   * @param {Object} options - Options
   */
  const safeAsync = async (fn, options = {}) => {
    const { fallback = null, onError = null } = options
    
    try {
      return await fn()
    } catch (error) {
      if (onError) {
        onError(error)
      }
      return fallback
    }
  }
  
  return {
    handleFetchError,
    withRetry,
    safeAsync
  }
}