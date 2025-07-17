/**
 * Library Optimizer
 * Handles dynamic loading and optimization of libraries for better performance
 */

// Critical libraries that should be preloaded
const CRITICAL_LIBRARIES = [
  // Add any critical libraries that need to be preloaded here
]

// Development-only libraries
const DEV_LIBRARIES = [
  // Add development tools here if needed
]

/**
 * Preload critical libraries for better performance
 */
export async function preloadCriticalLibraries() {
  const loadPromises = []

  for (const library of CRITICAL_LIBRARIES) {
    // For now, we'll just resolve immediately since no critical libraries are defined
    // In the future, you can add dynamic imports here like:
    // loadPromises.push(import(library).catch(error => {
    //   // Handle library loading errors silently in production
    //   if (process.env.NODE_ENV === 'development') {
    //     console.warn(`Failed to preload critical library: ${library}`, error)
    //   }
    // }))
  }

  try {
    await Promise.all(loadPromises)
    // Libraries preloaded successfully (silent in production)
    if (process.env.NODE_ENV === 'development') {
      // eslint-disable-next-line no-console
      console.log('✅ Critical libraries preloaded successfully')
    }
  } catch (error) {
    // Handle preload errors silently in production
    if (process.env.NODE_ENV === 'development') {
      // eslint-disable-next-line no-console
      console.warn('⚠️ Some critical libraries failed to preload:', error)
    }
  }
}

/**
 * Load development-only libraries conditionally
 */
export async function loadDevLibraries() {
  if (process.env.NODE_ENV !== 'development') {
    return []
  }

  const loadedLibraries = []
  const loadPromises = []

  for (const library of DEV_LIBRARIES) {
    // For now, we'll just resolve immediately since no dev libraries are defined
    // In the future, you can add dynamic imports here like:
    // loadPromises.push(import(library).then(lib => {
    //   loadedLibraries.push(library)
    //   return lib
    // }).catch(error => {
    //   console.warn(`Failed to load dev library: ${library}`, error)
    // }))
  }

  try {
    await Promise.all(loadPromises)
    // eslint-disable-next-line no-console
    console.log('🛠️ Development libraries loaded successfully')
  } catch (error) {
    // eslint-disable-next-line no-console
    console.warn('⚠️ Some development libraries failed to load:', error)
  }

  return loadedLibraries
}

/**
 * Dynamically load a library with error handling
 */
export async function loadLibrary(libraryPath, options = {}) {
  const { timeout = 10000, retries = 3, fallback = null } = options

  let attempt = 0

  while (attempt < retries) {
    try {
      const timeoutPromise = new Promise((_, reject) => {
        setTimeout(() => reject(new Error('Library load timeout')), timeout)
      })

      const loadPromise = import(/* @vite-ignore */ libraryPath)

      const library = await Promise.race([loadPromise, timeoutPromise])
      return library
    } catch (error) {
      attempt++
      if (process.env.NODE_ENV === 'development') {
        // eslint-disable-next-line no-console
        console.warn(`Library load attempt ${attempt} failed for ${libraryPath}:`, error.message)
      }

      if (attempt >= retries) {
        if (fallback) {
          if (process.env.NODE_ENV === 'development') {
            // eslint-disable-next-line no-console
            console.log(`Using fallback for ${libraryPath}`)
          }
          return fallback
        }
        throw error
      }

      // Wait before retry
      await new Promise(resolve => setTimeout(resolve, 1000 * attempt))
    }
  }
}

/**
 * Check if a library is available
 */
export function isLibraryAvailable(libraryName) {
  try {
    // Check if library is available in window object
    if (typeof window !== 'undefined' && window[libraryName]) {
      return true
    }

    // Check if it's a module that can be imported
    // This is a basic check - in practice you might want more sophisticated detection
    return false
  } catch {
    return false
  }
}

/**
 * Lazy load a library only when needed
 */
export function createLazyLoader(libraryPath, options = {}) {
  let loadPromise = null
  let loadedLibrary = null

  return async function lazyLoad() {
    if (loadedLibrary) {
      return loadedLibrary
    }

    if (!loadPromise) {
      loadPromise = loadLibrary(libraryPath, options)
        .then(lib => {
          loadedLibrary = lib
          return lib
        })
        .catch(error => {
          // Reset promise so we can retry
          loadPromise = null
          throw error
        })
    }

    return loadPromise
  }
}

/**
 * Preload libraries based on user interaction hints
 */
export function preloadOnInteraction(libraryPath, options = {}) {
  const {
    events = ['mouseenter', 'touchstart', 'focus'],
    element = document,
    once = true
  } = options

  let loaded = false
  const loader = createLazyLoader(libraryPath, options)

  const handleInteraction = async () => {
    if (loaded) return

    loaded = true

    try {
      await loader()
      if (process.env.NODE_ENV === 'development') {
        // eslint-disable-next-line no-console
        console.log(`📦 Preloaded library on interaction: ${libraryPath}`)
      }
    } catch (error) {
      if (process.env.NODE_ENV === 'development') {
        // eslint-disable-next-line no-console
        console.warn(`Failed to preload library on interaction: ${libraryPath}`, error)
      }
      loaded = false // Allow retry
    }

    if (once) {
      events.forEach(event => {
        element.removeEventListener(event, handleInteraction)
      })
    }
  }

  events.forEach(event => {
    element.addEventListener(event, handleInteraction, { passive: true })
  })

  return () => {
    events.forEach(event => {
      element.removeEventListener(event, handleInteraction)
    })
  }
}

/**
 * Bundle analyzer helper for development
 */
export function analyzeBundleSize() {
  if (process.env.NODE_ENV !== 'development') {
    return
  }

  // Basic bundle size analysis
  const scripts = document.querySelectorAll('script[src]')
  const styles = document.querySelectorAll('link[rel="stylesheet"]')

  // eslint-disable-next-line no-console
  console.group('📊 Bundle Analysis')
  // eslint-disable-next-line no-console
  console.log(`Scripts loaded: ${scripts.length}`)
  // eslint-disable-next-line no-console
  console.log(`Stylesheets loaded: ${styles.length}`)

  if ('performance' in window) {
    const resources = performance.getEntriesByType('resource')
    const jsResources = resources.filter(r => r.name.includes('.js'))
    const cssResources = resources.filter(r => r.name.includes('.css'))

    // eslint-disable-next-line no-console
    console.log(`JS resources: ${jsResources.length}`)
    // eslint-disable-next-line no-console
    console.log(`CSS resources: ${cssResources.length}`)

    const totalSize = resources.reduce((sum, resource) => {
      return sum + (resource.transferSize || 0)
    }, 0)

    // eslint-disable-next-line no-console
    console.log(`Total transfer size: ${(totalSize / 1024).toFixed(2)} KB`)
  }

  // eslint-disable-next-line no-console
  console.groupEnd()
}

// Auto-analyze bundle size in development
if (process.env.NODE_ENV === 'development') {
  window.addEventListener('load', () => {
    setTimeout(analyzeBundleSize, 1000)
  })
}

export default {
  preloadCriticalLibraries,
  loadDevLibraries,
  loadLibrary,
  isLibraryAvailable,
  createLazyLoader,
  preloadOnInteraction,
  analyzeBundleSize
}
