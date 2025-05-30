import { onUnmounted, ref } from 'vue'

/**
 * Composable for managing web workers
 */
export function useWebWorkers() {
  const workers = ref(new Map())
  const pendingTasks = ref(new Map())
  const taskIdCounter = ref(0)

  /**
   * Create a worker from a URL or file path
   */
  const createWorker = (workerScript, workerName = 'default') => {
    if (workers.value.has(workerName)) {
      return workers.value.get(workerName)
    }

    const worker = new Worker(workerScript)
    workers.value.set(workerName, worker)

    // Handle messages from worker
    worker.onmessage = e => {
      const { type, id, result, error } = e.data
      const task = pendingTasks.value.get(id)

      if (task) {
        if (type === 'SUCCESS') {
          task.resolve(result)
        } else if (type === 'ERROR') {
          task.reject(new Error(error))
        }
        pendingTasks.value.delete(id)
      }
    }

    // Handle worker errors
    worker.onerror = error => {
      console.error(`Worker ${workerName} error:`, error)

      // Reject all pending tasks for this worker
      for (const [id, task] of pendingTasks.value.entries()) {
        if (task.workerName === workerName) {
          task.reject(new Error(`Worker error: ${error.message}`))
          pendingTasks.value.delete(id)
        }
      }
    }

    return worker
  }

  /**
   * Send a task to a worker and return a promise
   */
  const executeTask = (workerName, type, data, options = {}) => {
    const { timeout = 30000 } = options
    const worker = workers.value.get(workerName)

    if (!worker) {
      return Promise.reject(new Error(`Worker ${workerName} not found`))
    }

    const taskId = ++taskIdCounter.value

    return new Promise((resolve, reject) => {
      // Set up timeout
      const timeoutId = setTimeout(() => {
        pendingTasks.value.delete(taskId)
        reject(new Error(`Task timeout after ${timeout}ms`))
      }, timeout)

      // Store pending task
      pendingTasks.value.set(taskId, {
        resolve: result => {
          clearTimeout(timeoutId)
          resolve(result)
        },
        reject: error => {
          clearTimeout(timeoutId)
          reject(error)
        },
        workerName
      })

      // Send task to worker
      worker.postMessage({
        type,
        id: taskId,
        data
      })
    })
  }

  /**
   * Terminate a specific worker
   */
  const terminateWorker = workerName => {
    const worker = workers.value.get(workerName)
    if (worker) {
      worker.terminate()
      workers.value.delete(workerName)

      // Reject pending tasks for this worker
      for (const [id, task] of pendingTasks.value.entries()) {
        if (task.workerName === workerName) {
          task.reject(new Error(`Worker ${workerName} terminated`))
          pendingTasks.value.delete(id)
        }
      }
    }
  }

  /**
   * Terminate all workers
   */
  const terminateAllWorkers = () => {
    for (const [workerName] of workers.value) {
      terminateWorker(workerName)
    }
  }

  /**
   * Check if worker is available
   */
  const hasWorker = workerName => {
    return workers.value.has(workerName)
  }

  /**
   * Get number of pending tasks
   */
  const getPendingTaskCount = (workerName = null) => {
    if (workerName) {
      return Array.from(pendingTasks.value.values()).filter(task => task.workerName === workerName)
        .length
    }
    return pendingTasks.value.size
  }

  // Cleanup on component unmount
  onUnmounted(() => {
    terminateAllWorkers()
  })

  return {
    createWorker,
    executeTask,
    terminateWorker,
    terminateAllWorkers,
    hasWorker,
    getPendingTaskCount,
    workers: workers.value,
    pendingTasks: pendingTasks.value
  }
}

/**
 * Specialized composable for player calculation workers
 */
export function usePlayerCalculationWorker() {
  const { createWorker, executeTask, terminateWorker, getPendingTaskCount, hasWorker } =
    useWebWorkers()

  // Create the player calculation worker
  const initWorker = () => {
    try {
      // Create worker from the workers directory
      const workerUrl = new URL('../workers/playerCalculationWorker.js', import.meta.url)
      return createWorker(workerUrl.href, 'playerCalculation')
    } catch (error) {
      console.warn('Web Worker not supported, falling back to main thread:', error.message || error)
      return null
    }
  }

  // Initialize worker on first use
  let workerInitialized = false
  let workerInitializationFailed = false

  const ensureWorker = () => {
    if (!workerInitialized && !workerInitializationFailed) {
      try {
        initWorker()
        workerInitialized = true
      } catch (error) {
        console.warn('Failed to initialize web worker:', error)
        workerInitializationFailed = true
      }
    }
  }

  /**
   * Sort players using web worker
   */
  const sortPlayers = async (players, fieldKey, direction, sortField, isGoalkeeperView) => {
    ensureWorker()

    // Use Web Worker for larger datasets (2000+) for better performance
    if (workerInitializationFailed || !hasWorker('playerCalculation') || players.length < 2000) {
      return sortPlayersMainThread(players, fieldKey, direction, sortField, isGoalkeeperView)
    }

    // Use the fast sort operation for large datasets
    return executeTask('playerCalculation', 'FAST_SORT_PLAYERS', {
      players,
      fieldKey,
      direction,
      sortField,
      isGoalkeeperView
    })
  }

  /**
   * Filter players using web worker
   */
  const filterPlayers = async (players, filters) => {
    ensureWorker()

    if (workerInitializationFailed || !hasWorker('playerCalculation') || players.length < 50) {
      return filterPlayersMainThread(players, filters)
    }

    return executeTask('playerCalculation', 'FILTER_PLAYERS', {
      players,
      filters
    })
  }

  /**
   * Calculate statistics using web worker
   */
  const calculateStats = async (players, statKey) => {
    ensureWorker()

    if (workerInitializationFailed || !hasWorker('playerCalculation') || players.length < 20) {
      return calculateStatsMainThread(players, statKey)
    }

    return executeTask('playerCalculation', 'CALCULATE_STATS', {
      players,
      statKey
    })
  }

  /**
   * Batch process multiple operations
   */
  const batchProcess = async (players, operations) => {
    ensureWorker()

    if (workerInitializationFailed || !hasWorker('playerCalculation')) {
      return batchProcessMainThread(players, operations)
    }

    return executeTask('playerCalculation', 'BATCH_PROCESS', {
      players,
      operations
    })
  }

  // Main thread fallback functions
  const sortPlayersMainThread = (players, fieldKey, direction, _sortField, _isGoalkeeperView) => {
    return [...players].sort((a, b) => {
      // Simplified sorting logic for fallback
      const aVal = a[fieldKey]
      const bVal = b[fieldKey]

      if (aVal === bVal) return 0
      if (aVal == null) return direction === 'asc' ? 1 : -1
      if (bVal == null) return direction === 'asc' ? -1 : 1

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

  const filterPlayersMainThread = (players, filters) => {
    return players.filter(player => {
      if (filters.name && !player.name.toLowerCase().includes(filters.name.toLowerCase())) {
        return false
      }
      return true
    })
  }

  const calculateStatsMainThread = (players, statKey) => {
    const values = players
      .map(p => p[statKey])
      .filter(v => v != null && !Number.isNaN(v))
      .sort((a, b) => a - b)

    if (values.length === 0) {
      return { min: 0, max: 0, mean: 0, median: 0, count: 0 }
    }

    const min = values[0]
    const max = values[values.length - 1]
    const sum = values.reduce((acc, val) => acc + val, 0)
    const mean = sum / values.length
    const median =
      values.length % 2 === 0
        ? (values[values.length / 2 - 1] + values[values.length / 2]) / 2
        : values[Math.floor(values.length / 2)]

    return { min, max, mean: Math.round(mean * 100) / 100, median, count: values.length }
  }

  const batchProcessMainThread = (players, operations) => {
    const results = {}

    for (const operation of operations) {
      switch (operation.type) {
        case 'sort':
          results[operation.id] = sortPlayersMainThread(
            players,
            operation.fieldKey,
            operation.direction,
            operation.sortField,
            operation.isGoalkeeperView
          )
          break
        case 'filter':
          results[operation.id] = filterPlayersMainThread(players, operation.filters)
          break
        case 'stats':
          results[operation.id] = calculateStatsMainThread(players, operation.statKey)
          break
      }
    }

    return results
  }

  return {
    sortPlayers,
    filterPlayers,
    calculateStats,
    batchProcess,
    getPendingTaskCount: () => getPendingTaskCount('playerCalculation'),
    terminateWorker: () => terminateWorker('playerCalculation')
  }
}
