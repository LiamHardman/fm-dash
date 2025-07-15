import { ref, onUnmounted } from 'vue'

/**
 * Enhanced Web Worker Manager
 * Provides optimized web worker usage with fallback mechanisms and progress reporting
 */
export function useWebWorkers() {
  const workers = ref(new Map())
  const workerHealth = ref(new Map())
  const activeOperations = ref(new Map())
  const fallbackMode = ref(false)

  // Worker configuration
  const WORKER_TIMEOUT = 30000 // 30 seconds
  const MAX_RETRIES = 3
  const HEALTH_CHECK_INTERVAL = 60000 // 1 minute

  /**
   * Create or get existing worker
   */
  const getWorker = (workerName = 'playerCalculation') => {
    if (!workers.value.has(workerName)) {
      try {
        const worker = new Worker('/src/workers/playerCalculationWorker.js', { type: 'module' })
        workers.value.set(workerName, worker)
        
        // Setup worker health monitoring
        setupWorkerHealthCheck(worker, workerName)
        
        // Setup error handling
        worker.onerror = (error) => {
          console.error(`Worker ${workerName} error:`, error)
          workerHealth.value.set(workerName, { status: 'error', lastError: error.message })
        }
        
      } catch (error) {
        console.warn(`Failed to create worker ${workerName}:`, error)
        fallbackMode.value = true
        return null
      }
    }
    
    return workers.value.get(workerName)
  }

  /**
   * Setup worker health monitoring
   */
  const setupWorkerHealthCheck = (worker, workerName) => {
    const healthCheck = () => {
      const healthCheckId = `health_${Date.now()}`
      const timeout = setTimeout(() => {
        workerHealth.value.set(workerName, { status: 'timeout', lastCheck: Date.now() })
      }, 5000)

      worker.postMessage({
        type: 'HEALTH_CHECK',
        id: healthCheckId,
        data: {}
      })

      const healthHandler = (e) => {
        if (e.data.id === healthCheckId) {
          clearTimeout(timeout)
          workerHealth.value.set(workerName, { 
            status: 'healthy', 
            lastCheck: Date.now(),
            capabilities: e.data.result?.capabilities
          })
          worker.removeEventListener('message', healthHandler)
        }
      }

      worker.addEventListener('message', healthHandler)
    }

    // Initial health check
    setTimeout(healthCheck, 1000)
    
    // Periodic health checks
    setInterval(healthCheck, HEALTH_CHECK_INTERVAL)
  }

  /**
   * Execute operation with fallback support
   */
  const executeOperation = async (operation, options = {}) => {
    const {
      workerName = 'playerCalculation',
      timeout = WORKER_TIMEOUT,
      retries = MAX_RETRIES,
      onProgress = null,
      fallbackFn = null
    } = options

    // Check if we should use fallback mode
    if (fallbackMode.value || !window.Worker) {
      if (fallbackFn) {
        return await fallbackFn(operation)
      }
      throw new Error('Web Workers not supported and no fallback provided')
    }

    const worker = getWorker(workerName)
    if (!worker) {
      if (fallbackFn) {
        return await fallbackFn(operation)
      }
      throw new Error('Failed to create worker and no fallback provided')
    }

    const operationId = `op_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
    let attempt = 0

    while (attempt < retries) {
      try {
        const result = await executeWithTimeout(worker, operation, operationId, timeout, onProgress)
        return result
      } catch (error) {
        attempt++
        console.warn(`Worker operation attempt ${attempt} failed:`, error.message)
        
        if (attempt >= retries) {
          // Final attempt failed, try fallback
          if (fallbackFn) {
            console.warn('All worker attempts failed, using fallback')
            return await fallbackFn(operation)
          }
          throw error
        }
        
        // Wait before retry
        await new Promise(resolve => setTimeout(resolve, 1000 * attempt))
      }
    }
  }

  /**
   * Execute operation with timeout
   */
  const executeWithTimeout = (worker, operation, operationId, timeout, onProgress) => {
    return new Promise((resolve, reject) => {
      const timeoutId = setTimeout(() => {
        activeOperations.value.delete(operationId)
        reject(new Error('Worker operation timeout'))
      }, timeout)

      const messageHandler = (e) => {
        const { type, id, result, error, progress } = e.data

        if (id !== operationId) return

        switch (type) {
          case 'SUCCESS':
            clearTimeout(timeoutId)
            activeOperations.value.delete(operationId)
            worker.removeEventListener('message', messageHandler)
            resolve(result)
            break

          case 'ERROR':
            clearTimeout(timeoutId)
            activeOperations.value.delete(operationId)
            worker.removeEventListener('message', messageHandler)
            reject(new Error(error))
            break

          case 'PROGRESS':
            if (onProgress) {
              onProgress(progress)
            }
            break
        }
      }

      worker.addEventListener('message', messageHandler)
      activeOperations.value.set(operationId, { operation, startTime: Date.now() })

      worker.postMessage({
        ...operation,
        id: operationId
      })
    })
  }

  /**
   * Sort players with chunked processing for large datasets
   */
  const sortPlayers = async (players, fieldKey, direction, options = {}) => {
    const operation = {
      type: players.length > 5000 ? 'CHUNKED_SORT' : 'FAST_SORT_PLAYERS',
      data: {
        players,
        fieldKey,
        direction,
        sortField: options.sortField,
        isGoalkeeperView: options.isGoalkeeperView || false,
        chunkSize: options.chunkSize || 2000
      }
    }

    const fallbackFn = async () => {
      // Fallback to main thread sorting
      return players.sort((a, b) => {
        const aVal = a[fieldKey]
        const bVal = b[fieldKey]
        
        if (aVal === null || aVal === undefined) return direction === 'asc' ? 1 : -1
        if (bVal === null || bVal === undefined) return direction === 'asc' ? -1 : 1
        
        if (typeof aVal === 'number' && typeof bVal === 'number') {
          return direction === 'asc' ? aVal - bVal : bVal - aVal
        }
        
        const aStr = String(aVal).toLowerCase()
        const bStr = String(bVal).toLowerCase()
        
        if (aStr < bStr) return direction === 'asc' ? -1 : 1
        if (aStr > bStr) return direction === 'asc' ? 1 : -1
        return 0
      })
    }

    return executeOperation(operation, {
      ...options,
      fallbackFn
    })
  }

  /**
   * Filter players with batch processing
   */
  const filterPlayers = async (players, filters, options = {}) => {
    const operation = {
      type: players.length > 5000 ? 'BATCH_FILTER' : 'FILTER_PLAYERS',
      data: {
        players,
        filters,
        chunkSize: options.chunkSize || 5000
      }
    }

    const fallbackFn = async () => {
      // Fallback to main thread filtering
      return players.filter(player => {
        if (filters.name && !player.name?.toLowerCase().includes(filters.name.toLowerCase())) {
          return false
        }
        if (filters.club && player.club !== filters.club) {
          return false
        }
        if (filters.position && !player.position?.includes(filters.position)) {
          return false
        }
        if (filters.ageMin !== null && player.age < filters.ageMin) {
          return false
        }
        if (filters.ageMax !== null && player.age > filters.ageMax) {
          return false
        }
        return true
      })
    }

    return executeOperation(operation, {
      ...options,
      fallbackFn
    })
  }

  /**
   * Calculate statistics
   */
  const calculateStats = async (players, statKey, options = {}) => {
    const operation = {
      type: 'CALCULATE_STATS',
      data: {
        players,
        statKey
      }
    }

    const fallbackFn = async () => {
      // Fallback to main thread calculation
      const values = players
        .map(p => p[statKey])
        .filter(v => v !== null && v !== undefined && !Number.isNaN(v))
        .sort((a, b) => a - b)

      if (values.length === 0) {
        return { min: 0, max: 0, mean: 0, median: 0, count: 0 }
      }

      const min = values[0]
      const max = values[values.length - 1]
      const sum = values.reduce((acc, val) => acc + val, 0)
      const mean = sum / values.length
      const median = values.length % 2 === 0
        ? (values[values.length / 2 - 1] + values[values.length / 2]) / 2
        : values[Math.floor(values.length / 2)]

      return { min, max, mean: Math.round(mean * 100) / 100, median, count: values.length }
    }

    return executeOperation(operation, {
      ...options,
      fallbackFn
    })
  }

  /**
   * Batch process multiple operations
   */
  const batchProcess = async (players, operations, options = {}) => {
    const operation = {
      type: 'BATCH_PROCESS',
      data: {
        players,
        operations
      }
    }

    const fallbackFn = async () => {
      // Fallback to sequential processing
      const results = {}
      for (const op of operations) {
        try {
          switch (op.type) {
            case 'sort':
              results[op.id] = await sortPlayers(players, op.fieldKey, op.direction, { fallbackFn: null })
              break
            case 'filter':
              results[op.id] = await filterPlayers(players, op.filters, { fallbackFn: null })
              break
            case 'stats':
              results[op.id] = await calculateStats(players, op.statKey, { fallbackFn: null })
              break
          }
        } catch (error) {
          results[op.id] = { error: error.message, fallback: true }
        }
      }
      return results
    }

    return executeOperation(operation, {
      ...options,
      fallbackFn
    })
  }

  /**
   * Get worker statistics
   */
  const getWorkerStats = () => {
    return {
      activeWorkers: workers.value.size,
      activeOperations: activeOperations.value.size,
      workerHealth: Object.fromEntries(workerHealth.value),
      fallbackMode: fallbackMode.value
    }
  }

  /**
   * Terminate all workers
   */
  const terminateWorkers = () => {
    for (const [name, worker] of workers.value) {
      try {
        worker.terminate()
      } catch (error) {
        console.warn(`Error terminating worker ${name}:`, error)
      }
    }
    workers.value.clear()
    workerHealth.value.clear()
    activeOperations.value.clear()
  }

  // Cleanup on unmount
  onUnmounted(() => {
    terminateWorkers()
  })

  return {
    // Core operations
    sortPlayers,
    filterPlayers,
    calculateStats,
    batchProcess,
    
    // Worker management
    getWorkerStats,
    terminateWorkers,
    
    // State
    fallbackMode,
    activeOperations: activeOperations.value,
    workerHealth: workerHealth.value
  }
}

// Alias for backward compatibility
export const usePlayerCalculationWorker = useWebWorkers

export default useWebWorkers