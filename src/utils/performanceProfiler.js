/**
 * Performance Profiler for Upload Flow
 * Measures each step of the upload → player data flow
 */

class UploadFlowProfiler {
  constructor() {
    this.metrics = {}
    this.startTime = null
    this.currentStep = null
    this.stepStartTime = null
  }

  /**
   * Start profiling the upload flow
   */
  startUploadFlow() {
    this.metrics = {}
    this.startTime = performance.now()
    this.currentStep = null
    this.stepStartTime = null
    
    console.log('🚀 Starting Upload Flow Profiling')
    this.markStep('upload_flow_start')
  }

  /**
   * Mark a step in the upload flow
   */
  markStep(stepName, additionalData = {}) {
    const now = performance.now()
    const stepDuration = this.stepStartTime ? now - this.stepStartTime : 0
    const totalDuration = this.startTime ? now - this.startTime : 0

    if (this.currentStep) {
      this.metrics[this.currentStep] = {
        duration: stepDuration,
        totalDuration,
        timestamp: now,
        ...additionalData
      }
    }

    this.currentStep = stepName
    this.stepStartTime = now

    console.log(`📊 Step: ${stepName}`, {
      stepDuration: `${stepDuration.toFixed(2)}ms`,
      totalDuration: `${totalDuration.toFixed(2)}ms`,
      ...additionalData
    })
  }

  /**
   * End the current step
   */
  endStep(additionalData = {}) {
    if (this.currentStep) {
      this.markStep(this.currentStep, additionalData)
    }
  }

  /**
   * End the upload flow and generate report
   */
  endUploadFlow() {
    const now = performance.now()
    const totalDuration = this.startTime ? now - this.startTime : 0

    if (this.currentStep) {
      this.metrics[this.currentStep] = {
        duration: this.stepStartTime ? now - this.stepStartTime : 0,
        totalDuration,
        timestamp: now
      }
    }

    this.generateReport()
  }

  /**
   * Generate a detailed performance report
   */
  generateReport() {
    console.log('📈 Upload Flow Performance Report')
    console.log('================================')

    const steps = Object.entries(this.metrics)
      .sort((a, b) => a[1].totalDuration - b[1].totalDuration)

    let totalTime = 0
    steps.forEach(([stepName, data]) => {
      totalTime += data.duration
      console.log(`${stepName}:`)
      console.log(`  Duration: ${data.duration.toFixed(2)}ms`)
      console.log(`  Total Time: ${data.totalDuration.toFixed(2)}ms`)
      if (data.playerCount) console.log(`  Player Count: ${data.playerCount}`)
      if (data.cacheSize) console.log(`  Cache Size: ${data.cacheSize}`)
      if (data.cacheType) console.log(`  Cache Type: ${data.cacheType}`)
      console.log('')
    })

    console.log(`Total Flow Time: ${totalTime.toFixed(2)}ms`)
    
    // Identify bottlenecks
    const bottlenecks = steps
      .filter(([_, data]) => data.duration > 100) // Steps taking more than 100ms
      .sort((a, b) => b[1].duration - a[1].duration)

    if (bottlenecks.length > 0) {
      console.log('🐌 Potential Bottlenecks:')
      bottlenecks.forEach(([stepName, data]) => {
        console.log(`  ${stepName}: ${data.duration.toFixed(2)}ms`)
      })
    }

    return {
      totalTime,
      steps: this.metrics,
      bottlenecks
    }
  }

  /**
   * Measure a specific operation
   */
  async measureOperation(operationName, operation) {
    this.markStep(`${operationName}_start`)
    const startTime = performance.now()
    
    try {
      const result = await operation()
      const duration = performance.now() - startTime
      this.markStep(`${operationName}_end`, { duration })
      return result
    } catch (error) {
      const duration = performance.now() - startTime
      this.markStep(`${operationName}_error`, { duration, error: error.message })
      throw error
    }
  }
}

// Global profiler instance
const uploadFlowProfiler = new UploadFlowProfiler()

export default uploadFlowProfiler 