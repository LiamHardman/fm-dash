/**
 * Image optimization utilities for better performance
 */

import { computed, ref } from 'vue'

/**
 * Creates optimized image URLs with fallbacks
 * @param {string} baseUrl - Base image URL
 * @param {object} options - Optimization options
 * @returns {object} - Object with optimized URLs and loading strategies
 */
export function createOptimizedImageUrl(baseUrl, options = {}) {
  const {
    width: _width = null,
    height: _height = null,
    format: _format = 'auto',
    quality: _quality = 80,
    fallbacks = true
  } = options

  // For external services like flagcdn.com, create optimized URLs
  if (baseUrl.includes('flagcdn.com')) {
    const webpUrl = baseUrl.replace(/\.(png|jpg|jpeg)$/, '.webp')
    return {
      webp: webpUrl,
      fallback: baseUrl,
      sources: fallbacks
        ? [
            { srcset: webpUrl, type: 'image/webp' },
            { srcset: baseUrl, type: 'image/png' }
          ]
        : null
    }
  }

  // For other images, return as-is but with WebP preference
  return {
    webp: baseUrl,
    fallback: baseUrl,
    sources: null
  }
}

/**
 * Progressive image loading component helper
 * @param {HTMLElement} imgElement - Image element
 * @param {string} src - Image source
 * @param {function} onLoad - Load callback
 * @param {function} onError - Error callback
 */
export function loadImageProgressively(imgElement, src, onLoad, onError) {
  const img = new Image()

  img.onload = () => {
    imgElement.src = src
    imgElement.classList.add('loaded')
    if (onLoad) onLoad()
  }

  img.onerror = () => {
    imgElement.classList.add('error')
    if (onError) onError()
  }

  img.src = src
}

/**
 * Lazy loading intersection observer
 */
export class LazyImageLoader {
  constructor(options = {}) {
    this.options = {
      rootMargin: '50px',
      threshold: 0.1,
      ...options
    }

    this.observer = new IntersectionObserver(this.handleIntersection.bind(this), this.options)
  }

  observe(element) {
    this.observer.observe(element)
  }

  unobserve(element) {
    this.observer.unobserve(element)
  }

  handleIntersection(entries) {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        const img = entry.target
        const src = img.dataset.src

        if (src) {
          loadImageProgressively(
            img,
            src,
            () => img.removeAttribute('data-src'),
            () => {}
          )
        }

        this.observer.unobserve(img)
      }
    })
  }

  disconnect() {
    this.observer.disconnect()
  }
}

// Global lazy loader instance
export const globalLazyLoader = new LazyImageLoader()

/**
 * Image preloader for critical images
 */
export class ImagePreloader {
  constructor() {
    this.cache = new Set()
  }

  preload(src) {
    if (this.cache.has(src)) {
      return Promise.resolve()
    }

    return new Promise((resolve, reject) => {
      const img = new Image()

      img.onload = () => {
        this.cache.add(src)
        resolve()
      }

      img.onerror = reject
      img.src = src
    })
  }

  preloadMultiple(sources) {
    return Promise.allSettled(sources.map(src => this.preload(src)))
  }
}

// Global preloader instance
export const globalImagePreloader = new ImagePreloader()

/**
 * Vue composable for optimized images
 */
export function useOptimizedImage(src, options = {}) {
  const loading = ref(true)
  const error = ref(false)
  const loaded = ref(false)

  const optimizedUrls = computed(() => createOptimizedImageUrl(src.value || src, options))

  const handleLoad = () => {
    loading.value = false
    loaded.value = true
    error.value = false
  }

  const handleError = () => {
    loading.value = false
    error.value = true
    loaded.value = false
  }

  return {
    optimizedUrls,
    loading,
    error,
    loaded,
    handleLoad,
    handleError
  }
}
