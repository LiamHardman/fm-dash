// Performance utilities for tracking timing bottlenecks

export class PerformanceTracker {
  constructor(name) {
    this.name = name
    this.startTime = performance.now()
    this.checkpoints = []
  }

  checkpoint(label) {
    const now = performance.now()
    const elapsed = now - this.startTime
    const lastCheckpoint =
      this.checkpoints.length > 0
        ? this.checkpoints[this.checkpoints.length - 1]
        : { time: this.startTime }
    const sinceLastCheckpoint = now - lastCheckpoint.time

    this.checkpoints.push({
      label,
      time: now,
      elapsed,
      sinceLastCheckpoint
    })
  }

  finish() {
    const totalTime = performance.now() - this.startTime
    return totalTime
  }
}

// Helper function to defer expensive operations
export function deferToNextFrame(callback) {
  return new Promise(resolve => {
    requestAnimationFrame(() => {
      const result = callback()
      resolve(result)
    })
  })
}

// Helper to batch process large arrays without blocking UI
export async function batchProcess(array, batchSize = 100, processor) {
  const results = []
  for (let i = 0; i < array.length; i += batchSize) {
    const batch = array.slice(i, i + batchSize)
    const batchResults = await deferToNextFrame(() => processor(batch))
    results.push(...batchResults)
  }
  return results
}

// Optimized min/max finder for large arrays
export function findMinMax(array, valueExtractor = x => x) {
  if (array.length === 0) return { min: 0, max: 0 }

  let min = Number.MAX_SAFE_INTEGER
  let max = Number.MIN_SAFE_INTEGER

  for (const item of array) {
    const value = valueExtractor(item)
    if (typeof value === 'number' && !Number.isNaN(value)) {
      if (value < min) min = value
      if (value > max) max = value
    }
  }

  return { min, max }
}
