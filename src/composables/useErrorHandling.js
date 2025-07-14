import { useQuasar } from 'quasar'
import { computed, ref } from 'vue'

export function useErrorHandling() {
  const $q = useQuasar()

  // Error state
  const errors = ref([])
  const isLoading = ref(false)
  const lastError = ref(null)

  // Error types
  const ErrorTypes = {
    NETWORK: 'NETWORK_ERROR',
    VALIDATION: 'VALIDATION_ERROR',
    AUTHENTICATION: 'AUTH_ERROR',
    AUTHORIZATION: 'PERMISSION_ERROR',
    NOT_FOUND: 'NOT_FOUND',
    SERVER: 'SERVER_ERROR',
    FILE_UPLOAD: 'FILE_UPLOAD_ERROR',
    TIMEOUT: 'TIMEOUT_ERROR'
  }

  // Error severity levels
  const ErrorSeverity = {
    LOW: 'low',
    MEDIUM: 'medium',
    HIGH: 'high',
    CRITICAL: 'critical'
  }

  // Computed properties
  const hasErrors = computed(() => errors.value.length > 0)
  const criticalErrors = computed(() =>
    errors.value.filter(err => err.severity === ErrorSeverity.CRITICAL)
  )

  // Add error to the list
  const addError = (error, context = {}) => {
    const errorObj = {
      id: Date.now(),
      timestamp: new Date().toISOString(),
      message: error.message || 'An unknown error occurred',
      type: classifyError(error),
      severity: determineSeverity(error),
      context: context,
      originalError: error,
      handled: false
    }

    errors.value.push(errorObj)
    lastError.value = errorObj

    if (errorObj.severity === ErrorSeverity.HIGH || errorObj.severity === ErrorSeverity.CRITICAL) {
      showErrorNotification(errorObj)
    }

    return errorObj
  }

  // Remove error from list
  const removeError = errorId => {
    const index = errors.value.findIndex(err => err.id === errorId)
    if (index > -1) {
      errors.value.splice(index, 1)
    }
  }

  // Clear all errors
  const clearErrors = () => {
    errors.value = []
    lastError.value = null
  }

  // Handle API errors specifically
  const handleApiError = async (error, context = {}) => {
    let errorMessage = 'An unexpected error occurred'
    let errorType = ErrorTypes.SERVER

    if (error.response) {
      // Server responded with error status
      const status = error.response.status
      const data = error.response.data

      switch (status) {
        case 400:
          errorType = ErrorTypes.VALIDATION
          errorMessage = data?.error?.message || 'Invalid request'
          break
        case 401:
          errorType = ErrorTypes.AUTHENTICATION
          errorMessage = 'Authentication required'
          break
        case 403:
          errorType = ErrorTypes.AUTHORIZATION
          errorMessage = 'Access denied'
          break
        case 404:
          errorType = ErrorTypes.NOT_FOUND
          errorMessage = data?.error?.message || 'Resource not found'
          break
        case 413:
          errorType = ErrorTypes.FILE_UPLOAD
          errorMessage = data?.error?.message || 'File too large'
          break
        case 422:
          errorType = ErrorTypes.VALIDATION
          errorMessage = data?.error?.message || 'Validation failed'
          break
        case 429:
          errorType = ErrorTypes.NETWORK
          errorMessage = 'Too many requests. Please try again later.'
          break
        case 500:
        case 502:
        case 503:
        case 504:
          errorType = ErrorTypes.SERVER
          errorMessage = 'Server error. Please try again later.'
          break
        default:
          errorMessage = data?.error?.message || `Server error (${status})`
      }
    } else if (error.request) {
      // Network error
      errorType = ErrorTypes.NETWORK
      errorMessage = 'Network error. Please check your connection.'
    } else {
      // Other error
      errorMessage = error.message || errorMessage
    }

    const errorObj = {
      ...error,
      message: errorMessage,
      type: errorType
    }

    return addError(errorObj, context)
  }

  // Handle fetch errors specifically
  const handleFetchError = async (response, context = {}) => {
    let errorMessage = 'Request failed'
    let errorType = ErrorTypes.SERVER

    try {
      if (response.status === 413) {
        errorType = ErrorTypes.FILE_UPLOAD
        errorMessage = 'File too large'
      } else if (response.status >= 400 && response.status < 500) {
        errorType = ErrorTypes.VALIDATION
        const data = await response.json()
        errorMessage = data?.error?.message || `Request failed (${response.status})`
      } else if (response.status >= 500) {
        errorType = ErrorTypes.SERVER
        errorMessage = 'Server error. Please try again later.'
      }
    } catch (_parseError) {
      errorMessage = `Request failed (${response.status})`
    }

    const error = new Error(errorMessage)
    error.response = response
    error.type = errorType

    return addError(error, context)
  }

  // Classify error type
  const classifyError = error => {
    if (error.type) return error.type
    if (error.name === 'TypeError' && error.message.includes('fetch')) {
      return ErrorTypes.NETWORK
    }
    if (error.name === 'AbortError') {
      return ErrorTypes.TIMEOUT
    }
    return ErrorTypes.SERVER
  }

  // Determine error severity
  const determineSeverity = error => {
    if (error.severity) return error.severity

    switch (error.type || classifyError(error)) {
      case ErrorTypes.NETWORK:
      case ErrorTypes.SERVER:
        return ErrorSeverity.HIGH
      case ErrorTypes.AUTHENTICATION:
      case ErrorTypes.AUTHORIZATION:
        return ErrorSeverity.CRITICAL
      case ErrorTypes.VALIDATION:
      case ErrorTypes.NOT_FOUND:
        return ErrorSeverity.MEDIUM
      case ErrorTypes.FILE_UPLOAD:
        return ErrorSeverity.MEDIUM
      default:
        return ErrorSeverity.LOW
    }
  }

  // Show error notification
  const showErrorNotification = errorObj => {
    const config = {
      type: 'negative',
      message: errorObj.message,
      position: 'top',
      timeout: 5000,
      actions: [
        {
          label: 'Dismiss',
          color: 'white',
          handler: () => removeError(errorObj.id)
        }
      ]
    }

    if (errorObj.severity === ErrorSeverity.CRITICAL) {
      config.timeout = 0 // Don't auto-dismiss critical errors
      config.color = 'red-10'
    }

    $q.notify(config)
    errorObj.handled = true
  }

  // Retry mechanism
  const withRetry = async (fn, maxRetries = 3, delay = 1000) => {
    let lastError

    for (let attempt = 1; attempt <= maxRetries; attempt++) {
      try {
        return await fn()
      } catch (error) {
        lastError = error

        if (attempt === maxRetries) {
          throw lastError
        }

        // Wait before retry (exponential backoff)
        await new Promise(resolve => setTimeout(resolve, delay * attempt))
      }
    }
  }

  // Loading state management
  const withLoading = async fn => {
    isLoading.value = true
    try {
      return await fn()
    } finally {
      isLoading.value = false
    }
  }

  // Safe async operation wrapper
  const safeAsync = async (fn, context = {}) => {
    try {
      return await withLoading(fn)
    } catch (error) {
      const errorObj = await handleApiError(error, context)
      throw errorObj
    }
  }

  return {
    errors,
    isLoading,
    lastError,
    hasErrors,
    criticalErrors,

    ErrorTypes,
    ErrorSeverity,

    addError,
    removeError,
    clearErrors,
    handleApiError,
    handleFetchError,
    showErrorNotification,
    withRetry,
    withLoading,
    safeAsync
  }
}
