/**
 * Third-Party Library Optimization Utilities
 * Provides optimized imports and conditional loading for better performance
 */

// Chart.js optimization - only import what we need
export const optimizedChartImports = {
  // Core Chart.js components used across the app
  core: () =>
    import('chart.js').then(module => ({
      Chart: module.Chart,
      LinearScale: module.LinearScale,
      PointElement: module.PointElement,
      Title: module.Title,
      Tooltip: module.Tooltip,
      Legend: module.Legend
    })),

  // Scatter plot specific components
  scatter: () =>
    import('chart.js').then(module => ({
      Chart: module.Chart,
      LinearScale: module.LinearScale,
      PointElement: module.PointElement,
      Title: module.Title,
      Tooltip: module.Tooltip,
      Legend: module.Legend
    })),

  // Annotation plugin - only load when needed
  annotation: () => import('chartjs-plugin-annotation')
}

// VueUse optimization - only import specific composables
export const optimizedVueUseImports = {
  // Web notification - only used in upload page
  webNotification: () =>
    import('@vueuse/core').then(module => ({
      useWebNotification: module.useWebNotification
    })),

  // Other commonly used VueUse composables
  storage: () =>
    import('@vueuse/core').then(module => ({
      useLocalStorage: module.useLocalStorage,
      useSessionStorage: module.useSessionStorage
    })),

  // DOM utilities
  dom: () =>
    import('@vueuse/core').then(module => ({
      useElementSize: module.useElementSize,
      useWindowSize: module.useWindowSize,
      useIntersectionObserver: module.useIntersectionObserver
    }))
}

// Quasar optimization - conditional component loading
export const optimizedQuasarImports = {
  // Core Quasar utilities that are always needed
  core: () =>
    import('quasar').then(module => ({
      useQuasar: module.useQuasar,
      Notify: module.Notify
    })),

  // Dialog components - only load when needed
  dialogs: () =>
    import('quasar').then(module => ({
      Dialog: module.Dialog
    })),

  // Loading components
  loading: () =>
    import('quasar').then(module => ({
      Loading: module.Loading,
      LoadingBar: module.LoadingBar
    }))
}

// Development-only library loader
export const loadDevLibraries = async () => {
  if (process.env.NODE_ENV === 'development') {
    // Only load development tools in dev mode
    const devTools = await Promise.allSettled([
      // Bundle analyzer
      import('rollup-plugin-visualizer').catch(() => null),
      // Performance monitoring tools
      import('@vue/devtools-api').catch(() => null)
    ])

    return devTools.filter(result => result.status === 'fulfilled').map(result => result.value)
  }
  return []
}

// Conditional library loader based on feature flags or user preferences
export const conditionalLibraryLoader = {
  // Load analytics only if user consents
  analytics: async (userConsent = false) => {
    if (userConsent && typeof window !== 'undefined') {
      return import('../services/analytics.js').catch(() => null)
    }
    return null
  },

  // Load export utilities only when needed
  export: async () => {
    return Promise.all([import('../utils/csvExport.js'), import('../utils/security.js')]).catch(
      () => null
    )
  },

  // Load performance monitoring only in production
  performance: async () => {
    if (process.env.NODE_ENV === 'production') {
      return import('../utils/performance.js').catch(() => null)
    }
    return null
  }
}

// Bundle size analyzer for development
export const analyzeBundleSize = async () => {
  if (process.env.NODE_ENV === 'development') {
    try {
      const { visualizer } = await import('rollup-plugin-visualizer')
      console.log('📊 Bundle analyzer available - run "npm run build:analyze" to generate report')
      return visualizer
    } catch (error) {
      console.warn('Bundle analyzer not available:', error.message)
      return null
    }
  }
  return null
}

// Preload critical third-party libraries
export const preloadCriticalLibraries = async () => {
  const criticalLibraries = [
    // Vue ecosystem - highest priority
    import('vue'),
    import('vue-router'),
    import('pinia'),

    // UI framework - high priority
    import('quasar').then(module => ({
      useQuasar: module.useQuasar,
      Notify: module.Notify
    }))
  ]

  try {
    await Promise.all(criticalLibraries)
    console.log('✅ Critical libraries preloaded')
  } catch (error) {
    console.warn('⚠️ Some critical libraries failed to preload:', error.message)
  }
}

// Library usage tracker for optimization insights
export const libraryUsageTracker = {
  usage: new Map(),

  track(libraryName, feature) {
    const key = `${libraryName}:${feature}`
    const current = this.usage.get(key) || 0
    this.usage.set(key, current + 1)
  },

  getReport() {
    const report = {}
    for (const [key, count] of this.usage.entries()) {
      const [library, feature] = key.split(':')
      if (!report[library]) {
        report[library] = {}
      }
      report[library][feature] = count
    }
    return report
  },

  getUnusedFeatures(threshold = 0) {
    const report = this.getReport()
    const unused = {}

    for (const [library, features] of Object.entries(report)) {
      const unusedFeatures = Object.entries(features)
        .filter(([, count]) => count <= threshold)
        .map(([feature]) => feature)

      if (unusedFeatures.length > 0) {
        unused[library] = unusedFeatures
      }
    }

    return unused
  }
}

// Export optimization utilities
export default {
  optimizedChartImports,
  optimizedVueUseImports,
  optimizedQuasarImports,
  loadDevLibraries,
  conditionalLibraryLoader,
  analyzeBundleSize,
  preloadCriticalLibraries,
  libraryUsageTracker
}
