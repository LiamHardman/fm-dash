import { ref, computed, watch, shallowRef } from 'vue'

/**
 * Data Pagination and Windowing Composable
 * Provides efficient pagination and windowing for large datasets
 */
export function useDataPagination(options = {}) {
  const {
    pageSize = 100,
    windowSize = 1000,
    enableVirtualization = true,
    preloadPages = 2
  } = options

  // State
  const currentPage = ref(0)
  const totalItems = ref(0)
  const data = shallowRef([])
  const windowedData = shallowRef([])
  const loadedPages = ref(new Set())
  const loading = ref(false)

  // Window state
  const currentWindow = ref(0)
  const windowStart = computed(() => currentWindow.value * windowSize)
  const windowEnd = computed(() => Math.min(windowStart.value + windowSize, totalItems.value))

  // Pagination computed properties
  const totalPages = computed(() => Math.ceil(totalItems.value / pageSize))
  const currentPageStart = computed(() => currentPage.value * pageSize)
  const currentPageEnd = computed(() => Math.min(currentPageStart.value + pageSize, totalItems.value))
  const hasNextPage = computed(() => currentPage.value < totalPages.value - 1)
  const hasPreviousPage = computed(() => currentPage.value > 0)

  // Current page data
  const currentPageData = computed(() => {
    if (!enableVirtualization) {
      return data.value.slice(currentPageStart.value, currentPageEnd.value)
    }
    
    // Use windowed data for virtualization
    const windowRelativeStart = currentPageStart.value - windowStart.value
    const windowRelativeEnd = currentPageEnd.value - windowStart.value
    
    return windowedData.value.slice(
      Math.max(0, windowRelativeStart),
      Math.min(windowedData.value.length, windowRelativeEnd)
    )
  })

  // Page info
  const pageInfo = computed(() => ({
    currentPage: currentPage.value,
    totalPages: totalPages.value,
    pageSize: pageSize,
    totalItems: totalItems.value,
    startItem: currentPageStart.value + 1,
    endItem: Math.min(currentPageEnd.value, totalItems.value),
    hasNext: hasNextPage.value,
    hasPrevious: hasPreviousPage.value
  }))

  /**
   * Set the data source
   */
  const setData = (newData, total = null) => {
    data.value = newData
    totalItems.value = total !== null ? total : newData.length
    
    if (enableVirtualization) {
      updateWindow()
    }
    
    // Reset to first page
    currentPage.value = 0
    loadedPages.value.clear()
    loadedPages.value.add(0)
  }

  /**
   * Update windowed data based on current window
   */
  const updateWindow = () => {
    if (!enableVirtualization) return
    
    const start = windowStart.value
    const end = windowEnd.value
    
    windowedData.value = data.value.slice(start, end)
  }

  /**
   * Navigate to specific page
   */
  const goToPage = (page) => {
    if (page < 0 || page >= totalPages.value) return false
    
    const oldPage = currentPage.value
    currentPage.value = page
    
    // Check if we need to update the window
    if (enableVirtualization) {
      const pageStart = page * pageSize
      const pageEnd = pageStart + pageSize
      
      // Check if page is outside current window
      if (pageStart < windowStart.value || pageEnd > windowEnd.value) {
        // Calculate new window
        const newWindow = Math.floor(pageStart / windowSize)
        if (newWindow !== currentWindow.value) {
          currentWindow.value = newWindow
          updateWindow()
        }
      }
    }
    
    loadedPages.value.add(page)
    return true
  }

  /**
   * Go to next page
   */
  const nextPage = () => {
    return goToPage(currentPage.value + 1)
  }

  /**
   * Go to previous page
   */
  const previousPage = () => {
    return goToPage(currentPage.value - 1)
  }

  /**
   * Go to first page
   */
  const firstPage = () => {
    return goToPage(0)
  }

  /**
   * Go to last page
   */
  const lastPage = () => {
    return goToPage(totalPages.value - 1)
  }

  /**
   * Preload adjacent pages for smoother navigation
   */
  const preloadAdjacentPages = async (loadFn) => {
    if (!loadFn || loading.value) return
    
    const pagesToPreload = []
    
    // Preload previous pages
    for (let i = 1; i <= preloadPages; i++) {
      const prevPage = currentPage.value - i
      if (prevPage >= 0 && !loadedPages.value.has(prevPage)) {
        pagesToPreload.push(prevPage)
      }
    }
    
    // Preload next pages
    for (let i = 1; i <= preloadPages; i++) {
      const nextPage = currentPage.value + i
      if (nextPage < totalPages.value && !loadedPages.value.has(nextPage)) {
        pagesToPreload.push(nextPage)
      }
    }
    
    if (pagesToPreload.length === 0) return
    
    loading.value = true
    
    try {
      await Promise.all(
        pagesToPreload.map(async (page) => {
          const start = page * pageSize
          const end = Math.min(start + pageSize, totalItems.value)
          await loadFn(start, end, page)
          loadedPages.value.add(page)
        })
      )
    } catch (error) {
      console.warn('Failed to preload pages:', error)
    } finally {
      loading.value = false
    }
  }

  /**
   * Get page range for pagination UI
   */
  const getPageRange = (maxPages = 10) => {
    const total = totalPages.value
    const current = currentPage.value
    
    if (total <= maxPages) {
      return Array.from({ length: total }, (_, i) => i)
    }
    
    const half = Math.floor(maxPages / 2)
    let start = Math.max(0, current - half)
    let end = Math.min(total, start + maxPages)
    
    // Adjust start if we're near the end
    if (end - start < maxPages) {
      start = Math.max(0, end - maxPages)
    }
    
    return Array.from({ length: end - start }, (_, i) => start + i)
  }

  /**
   * Search within current data
   */
  const searchInCurrentData = (searchFn) => {
    if (enableVirtualization) {
      return windowedData.value.filter(searchFn)
    }
    return data.value.filter(searchFn)
  }

  /**
   * Get statistics about pagination state
   */
  const getStats = () => {
    return {
      totalItems: totalItems.value,
      totalPages: totalPages.value,
      currentPage: currentPage.value,
      pageSize: pageSize,
      windowSize: windowSize,
      currentWindow: currentWindow.value,
      loadedPages: loadedPages.value.size,
      windowedDataSize: windowedData.value.length,
      enableVirtualization
    }
  }

  /**
   * Reset pagination state
   */
  const reset = () => {
    currentPage.value = 0
    currentWindow.value = 0
    totalItems.value = 0
    data.value = []
    windowedData.value = []
    loadedPages.value.clear()
    loading.value = false
  }

  // Watch for window changes to update windowed data
  watch(currentWindow, updateWindow)

  return {
    // State
    currentPage,
    totalItems,
    data,
    loading,
    
    // Computed
    totalPages,
    currentPageData,
    pageInfo,
    hasNextPage,
    hasPreviousPage,
    
    // Navigation
    goToPage,
    nextPage,
    previousPage,
    firstPage,
    lastPage,
    
    // Data management
    setData,
    preloadAdjacentPages,
    searchInCurrentData,
    
    // Utilities
    getPageRange,
    getStats,
    reset
  }
}

/**
 * Infinite scroll pagination for continuous loading
 */
export function useInfiniteScroll(options = {}) {
  const {
    pageSize = 50,
    threshold = 0.8,
    maxItems = 10000
  } = options

  const items = shallowRef([])
  const loading = ref(false)
  const hasMore = ref(true)
  const currentPage = ref(0)
  const error = ref(null)

  /**
   * Load more items
   */
  const loadMore = async (loadFn) => {
    if (loading.value || !hasMore.value) return
    
    loading.value = true
    error.value = null
    
    try {
      const offset = currentPage.value * pageSize
      const newItems = await loadFn(offset, pageSize)
      
      if (!Array.isArray(newItems)) {
        throw new Error('Load function must return an array')
      }
      
      // Append new items
      items.value = [...items.value, ...newItems]
      currentPage.value++
      
      // Check if we have more items
      hasMore.value = newItems.length === pageSize && items.value.length < maxItems
      
    } catch (err) {
      error.value = err.message
      console.error('Failed to load more items:', err)
    } finally {
      loading.value = false
    }
  }

  /**
   * Check if we should load more based on scroll position
   */
  const checkLoadMore = (scrollTop, scrollHeight, clientHeight, loadFn) => {
    const scrollPercentage = (scrollTop + clientHeight) / scrollHeight
    
    if (scrollPercentage >= threshold && !loading.value && hasMore.value) {
      loadMore(loadFn)
    }
  }

  /**
   * Reset infinite scroll state
   */
  const reset = () => {
    items.value = []
    currentPage.value = 0
    hasMore.value = true
    loading.value = false
    error.value = null
  }

  /**
   * Remove items beyond a certain limit to prevent memory issues
   */
  const trimItems = (maxItems = 1000) => {
    if (items.value.length > maxItems) {
      const itemsToRemove = items.value.length - maxItems
      items.value = items.value.slice(itemsToRemove)
      
      // Adjust page count
      currentPage.value = Math.max(0, currentPage.value - Math.ceil(itemsToRemove / pageSize))
    }
  }

  return {
    // State
    items,
    loading,
    hasMore,
    error,
    currentPage,
    
    // Actions
    loadMore,
    checkLoadMore,
    reset,
    trimItems
  }
}

export default useDataPagination