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

  log(..._args) {
    if (isDev && !isTest) {
    }
  }

  info(..._args) {
    if (isDev && !isTest) {
    }
  }

  warn(..._args) {
    if (isDev && !isTest) {
    }
  }

  error(..._args) {}

  debug(..._args) {
    if (isDev && !isTest) {
    }
  }

  // Performance logging
  time(_label) {
    if (isDev && !isTest) {
    }
  }

  timeEnd(_label) {
    if (isDev && !isTest) {
    }
  }

  // Cache operations
  cache(_operation, _key, ..._args) {
    if (isDev && !isTest) {
    }
  }

  // Performance operations
  perf(_operation, ..._args) {
    if (isDev && !isTest) {
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
  mark: name => {
    if (isDev && window.performance) {
      window.performance.mark(`${name}-start`)
    }
  },

  /**
   * Measure performance between marks
   */
  measure: name => {
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
