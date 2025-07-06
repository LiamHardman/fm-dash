/**
 * Development logger utility
 * Only logs in development mode, silent in production
 */

const isDev = import.meta.env.DEV
const isTest = import.meta.env.MODE === 'test'

class Logger {
  constructor(namespace = '') {
    this.namespace = namespace
  }

  _formatMessage(level, args) {
    const timestamp = new Date().toISOString().substr(11, 12)
    const prefix = this.namespace ? `[${this.namespace}]` : ''
    return [`${timestamp} ${level} ${prefix}`, ...args]
  }

  log(...args) {
    if (isDev && !isTest) {
      console.log(...this._formatMessage('ℹ️', args))
    }
  }

  info(...args) {
    if (isDev && !isTest) {
      console.info(...this._formatMessage('ℹ️', args))
    }
  }

  warn(...args) {
    if (isDev && !isTest) {
      console.warn(...this._formatMessage('⚠️', args))
    }
  }

  error(...args) {
    // Always log errors, even in production
    console.error(...this._formatMessage('❌', args))
  }

  debug(...args) {
    if (isDev && !isTest) {
      console.debug(...this._formatMessage('🐛', args))
    }
  }

  // Performance logging
  time(label) {
    if (isDev && !isTest) {
      console.time(`⏱️ ${this.namespace} ${label}`)
    }
  }

  timeEnd(label) {
    if (isDev && !isTest) {
      console.timeEnd(`⏱️ ${this.namespace} ${label}`)
    }
  }

  // Cache operations
  cache(operation, key, ...args) {
    if (isDev && !isTest) {
      console.log(...this._formatMessage('💾', [`Cache ${operation}:`, key, ...args]))
    }
  }

  // Performance operations
  perf(operation, ...args) {
    if (isDev && !isTest) {
      console.log(...this._formatMessage('⚡', [operation, ...args]))
    }
  }
}

// Factory function to create loggers with namespaces
export function createLogger(namespace) {
  return new Logger(namespace)
}

// Default logger
export const logger = new Logger()

// Convenience loggers for common areas
export const cacheLogger = new Logger('Cache')
export const perfLogger = new Logger('Performance')
export const apiLogger = new Logger('API')
export const storeLogger = new Logger('Store')

/**
 * Performance monitoring utilities
 */
export const performance = {
  /**
   * Mark the start of a performance measurement
   */
  mark: (name) => {
    if (isDev && window.performance) {
      window.performance.mark(`${name}-start`)
    }
  },

  /**
   * Measure performance between marks
   */
  measure: (name) => {
    if (isDev && window.performance) {
      try {
        window.performance.mark(`${name}-end`)
        window.performance.measure(name, `${name}-start`, `${name}-end`)
        const measures = window.performance.getEntriesByName(name)
        if (measures.length > 0) {
          logger.log(`⚡ ${name}: ${measures[0].duration.toFixed(2)}ms`)
        }
      } catch (e) {
        logger.warn('Performance measurement failed:', e)
      }
    }
  }
}

export default logger 