/**
 * Dynamic Component Loader Utility
 * Provides lazy loading capabilities for heavy components with loading states and error boundaries
 */

import { computed, defineAsyncComponent, ref } from 'vue'

class DynamicComponentLoader {
  constructor() {
    this.loadingStates = new Map()
    this.errorStates = new Map()
    this.preloadQueue = new Set()
    this.componentCache = new Map()
  }

  /**
   * Create a dynamic component with loading and error states
   * @param {Function} loader - Component loader function
   * @param {Object} options - Configuration options
   * @returns {Object} Vue async component
   */
  createAsyncComponent(loader, options = {}) {
    const {
      loadingComponent = null,
      errorComponent = null,
      delay = 200,
      timeout = 30000,
      suspensible = false,
      onError = null,
      retryAttempts = 3
    } = options

    const componentId = this.generateComponentId(loader)

    return defineAsyncComponent({
      loader: async () => {
        this.setLoadingState(componentId, true)

        try {
          // Check cache first
          if (this.componentCache.has(componentId)) {
            this.setLoadingState(componentId, false)
            return this.componentCache.get(componentId)
          }

          const component = await this.loadWithRetry(loader, retryAttempts)

          // Cache the component
          this.componentCache.set(componentId, component)
          this.setLoadingState(componentId, false)
          this.clearError(componentId)

          return component
        } catch (error) {
          this.setLoadingState(componentId, false)
          this.setError(componentId, error)

          if (onError) {
            onError(error)
          }

          console.error(`Failed to load component ${componentId}:`, error)
          throw error
        }
      },
      loadingComponent,
      errorComponent,
      delay,
      timeout,
      suspensible
    })
  }

  /**
   * Load component with retry mechanism
   * @param {Function} loader - Component loader function
   * @param {number} retryAttempts - Number of retry attempts
   * @returns {Promise} Component promise
   */
  async loadWithRetry(loader, retryAttempts = 3) {
    let lastError

    for (let attempt = 0; attempt <= retryAttempts; attempt++) {
      try {
        const result = await loader()
        return result.default || result
      } catch (error) {
        lastError = error

        if (attempt < retryAttempts) {
          // Exponential backoff
          const delay = 2 ** attempt * 1000
          await new Promise(resolve => setTimeout(resolve, delay))
        }
      }
    }

    throw lastError
  }

  /**
   * Preload a component without rendering it
   * @param {Function} loader - Component loader function
   * @returns {Promise} Preload promise
   */
  async preloadComponent(loader) {
    const componentId = this.generateComponentId(loader)

    if (this.componentCache.has(componentId) || this.preloadQueue.has(componentId)) {
      return
    }

    this.preloadQueue.add(componentId)

    try {
      const component = await this.loadWithRetry(loader)
      this.componentCache.set(componentId, component)
      console.log(`✅ Preloaded component: ${componentId}`)
    } catch (error) {
      console.warn(`⚠️ Failed to preload component ${componentId}:`, error)
    } finally {
      this.preloadQueue.delete(componentId)
    }
  }

  /**
   * Preload multiple components based on priority
   * @param {Array} components - Array of {loader, priority} objects
   */
  async preloadCriticalComponents(components = []) {
    // Sort by priority (higher number = higher priority)
    const sortedComponents = components.sort((a, b) => (b.priority || 0) - (a.priority || 0))

    // Preload high priority components first
    const highPriority = sortedComponents.filter(c => (c.priority || 0) >= 8)
    const mediumPriority = sortedComponents.filter(
      c => (c.priority || 0) >= 5 && (c.priority || 0) < 8
    )
    const lowPriority = sortedComponents.filter(c => (c.priority || 0) < 5)

    // Load high priority components immediately
    await Promise.all(highPriority.map(c => this.preloadComponent(c.loader)))

    // Load medium priority components with slight delay
    setTimeout(() => {
      Promise.all(mediumPriority.map(c => this.preloadComponent(c.loader)))
    }, 1000)

    // Load low priority components when idle
    if (window.requestIdleCallback) {
      window.requestIdleCallback(() => {
        Promise.all(lowPriority.map(c => this.preloadComponent(c.loader)))
      })
    } else {
      setTimeout(() => {
        Promise.all(lowPriority.map(c => this.preloadComponent(c.loader)))
      }, 3000)
    }
  }

  /**
   * Get loading state for a component
   * @param {string} componentId - Component identifier
   * @returns {boolean} Loading state
   */
  isLoading(componentId) {
    return this.loadingStates.get(componentId) || false
  }

  /**
   * Get error state for a component
   * @param {string} componentId - Component identifier
   * @returns {Error|null} Error state
   */
  getError(componentId) {
    return this.errorStates.get(componentId) || null
  }

  /**
   * Clear component cache
   * @param {string} componentId - Optional component ID to clear specific component
   */
  clearCache(componentId = null) {
    if (componentId) {
      this.componentCache.delete(componentId)
      this.loadingStates.delete(componentId)
      this.errorStates.delete(componentId)
    } else {
      this.componentCache.clear()
      this.loadingStates.clear()
      this.errorStates.clear()
    }
  }

  /**
   * Get cache statistics
   * @returns {Object} Cache stats
   */
  getCacheStats() {
    return {
      cachedComponents: this.componentCache.size,
      loadingComponents: Array.from(this.loadingStates.values()).filter(Boolean).length,
      errorComponents: this.errorStates.size,
      preloadQueue: this.preloadQueue.size
    }
  }

  // Private methods
  generateComponentId(loader) {
    return loader.toString().slice(0, 50) + '_' + Date.now()
  }

  setLoadingState(componentId, loading) {
    this.loadingStates.set(componentId, loading)
  }

  setError(componentId, error) {
    this.errorStates.set(componentId, error)
  }

  clearError(componentId) {
    this.errorStates.delete(componentId)
  }
}

// Create singleton instance
const dynamicLoader = new DynamicComponentLoader()

// Composable for using dynamic component loader in components
export function useDynamicLoader() {
  const loadingStates = ref(new Map())
  const errorStates = ref(new Map())

  const createAsyncComponent = (loader, options = {}) => {
    return dynamicLoader.createAsyncComponent(loader, {
      ...options,
      onError: error => {
        const componentId = dynamicLoader.generateComponentId(loader)
        errorStates.value.set(componentId, error)
        if (options.onError) {
          options.onError(error)
        }
      }
    })
  }

  const preloadComponent = loader => {
    return dynamicLoader.preloadComponent(loader)
  }

  const preloadCriticalComponents = components => {
    return dynamicLoader.preloadCriticalComponents(components)
  }

  const isLoading = computed(() => componentId => {
    return dynamicLoader.isLoading(componentId)
  })

  const getError = computed(() => componentId => {
    return dynamicLoader.getError(componentId)
  })

  const cacheStats = computed(() => dynamicLoader.getCacheStats())

  return {
    createAsyncComponent,
    preloadComponent,
    preloadCriticalComponents,
    isLoading,
    getError,
    cacheStats,
    clearCache: dynamicLoader.clearCache.bind(dynamicLoader)
  }
}

export default dynamicLoader
