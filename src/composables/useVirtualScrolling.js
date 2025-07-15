import { computed, ref, onUnmounted } from 'vue'
import VirtualScrollManager from '../utils/VirtualScrollManager.js'

export function useVirtualScrolling(options = {}) {
  const { itemHeight = 30, bufferSize = 10, containerHeight = 400, visibleRange = null } = options

  // Create advanced virtual scroll manager
  const scrollManager = new VirtualScrollManager({
    itemHeight,
    bufferSize,
    containerHeight,
    enableVariableHeight: options.enableVariableHeight || false,
    enableMomentum: options.enableMomentum !== false,
    overscan: options.overscan || 5,
    recycleThreshold: options.recycleThreshold || 100
  })

  // State
  const scrollTop = ref(0)
  const isScrolling = ref(false)
  const scrollTimeout = ref(null)
  const visibleRangeRef = ref({ startIndex: 0, endIndex: 0, visibleStartIndex: 0, visibleEndIndex: 0 })

  // Setup callbacks
  scrollManager.onScroll = (scrollInfo) => {
    scrollTop.value = scrollInfo.scrollTop
    isScrolling.value = scrollInfo.isScrolling
  }

  scrollManager.onVisibleRangeChange = (range) => {
    visibleRangeRef.value = range
  }

  // Computed properties for virtual scrolling
  const visibleItemCount = computed(() => Math.ceil(containerHeight / itemHeight))

  const startIndex = computed(() => {
    if (visibleRange?.value) {
      return Math.max(0, visibleRange.value.from - bufferSize)
    }
    return visibleRangeRef.value.startIndex
  })

  const endIndex = computed(() => {
    if (visibleRange?.value) {
      return Math.min(visibleRange.value.to + bufferSize)
    }
    return visibleRangeRef.value.endIndex
  })

  // Enhanced rendering function with recycling
  const getVisibleItems = items => {
    if (!items || items.length === 0) return []

    const range = scrollManager.calculateVisibleRange(items.length)
    const recycledItems = scrollManager.recycleItems(items, range)

    return recycledItems.map(recycledItem => ({
      ...recycledItem.data,
      virtualIndex: recycledItem.virtualIndex,
      actualIndex: recycledItem.index,
      offset: recycledItem.offset,
      height: recycledItem.height,
      _recycled: true,
      _id: recycledItem.id
    }))
  }

  // Enhanced scroll handler with momentum
  const handleScroll = scrollInfo => {
    const scrollPosition = scrollInfo.verticalPosition || scrollInfo.scrollTop || scrollInfo
    scrollManager.handleScroll(scrollPosition, performance.now())

    if (scrollTimeout.value) {
      clearTimeout(scrollTimeout.value)
    }

    scrollTimeout.value = setTimeout(() => {
      isScrolling.value = false
    }, 150)
  }

  // Calculate total height for virtual scrolling
  const getTotalHeight = itemCount => scrollManager.getTotalHeight(itemCount)

  // Get offset for virtual positioning
  const getOffset = () => scrollManager.getOffsetForIndex(startIndex.value)

  // Set item height for variable height support
  const setItemHeight = (index, height) => {
    scrollManager.setItemHeight(index, height)
  }

  // Get performance statistics
  const getStats = () => scrollManager.getStats()

  // Optimize memory usage
  const optimizeMemory = () => scrollManager.optimizeMemoryUsage()

  // Cleanup
  onUnmounted(() => {
    scrollManager.destroy()
  })

  return {
    scrollTop,
    isScrolling,
    startIndex,
    endIndex,
    visibleItemCount,
    visibleRange: visibleRangeRef,

    getVisibleItems,
    handleScroll,
    getTotalHeight,
    getOffset,
    setItemHeight,
    getStats,
    optimizeMemory,
    scrollManager
  }
}

// Optimized sorting composable for large datasets
export function useOptimizedSorting() {
  const sortCache = new Map()
  const sortField = ref('')
  const sortDirection = ref('desc')

  // Memoized sort function
  const createSortFunction = (field, direction, customSortFn = null) => {
    const cacheKey = `${field}-${direction}-${customSortFn ? 'custom' : 'default'}`

    if (sortCache.has(cacheKey)) {
      return sortCache.get(cacheKey)
    }

    const sortFn = (a, b) => {
      if (customSortFn) {
        return customSortFn(a, b, field, direction)
      }

      let valueA = a[field]
      let valueB = b[field]

      // Handle null/undefined values
      if (valueA == null && valueB == null) return 0
      if (valueA == null) return direction === 'asc' ? 1 : -1
      if (valueB == null) return direction === 'asc' ? -1 : 1

      // Numeric comparison
      if (typeof valueA === 'number' && typeof valueB === 'number') {
        return direction === 'asc' ? valueA - valueB : valueB - valueA
      }

      // String comparison
      if (typeof valueA === 'string' && typeof valueB === 'string') {
        valueA = valueA.toLowerCase()
        valueB = valueB.toLowerCase()
        if (valueA < valueB) return direction === 'asc' ? -1 : 1
        if (valueA > valueB) return direction === 'asc' ? 1 : -1
        return 0
      }

      return 0
    }

    sortCache.set(cacheKey, sortFn)
    return sortFn
  }

  // Chunked sorting for large datasets
  const sortLargeArray = async (array, field, direction, customSortFn = null, chunkSize = 2000) => {
    if (array.length <= 500) {
      // Quick sync sort for smaller arrays
      return array.sort(createSortFunction(field, direction, customSortFn))
    }

    // For large arrays, use optimized chunked sorting with larger chunks
    const sortFn = createSortFunction(field, direction, customSortFn)

    // Use larger chunks with less frequent yielding for speed
    return await fastChunkedSort([...array], sortFn, Math.min(chunkSize, 2000))
  }

  // Fast chunked sorting with optimized yielding
  const fastChunkedSort = async (array, sortFn, chunkSize = 2000) => {
    if (array.length <= chunkSize) {
      return array.sort(sortFn)
    }

    // Split into larger chunks for efficiency
    const chunks = []
    for (let i = 0; i < array.length; i += chunkSize) {
      const chunk = array.slice(i, i + chunkSize)
      chunks.push(chunk.sort(sortFn))

      // Only yield every few chunks to maintain speed
      if (chunks.length % 3 === 0) {
        await new Promise(resolve => setTimeout(resolve, 0))
      }
    }

    // Fast merge with minimal yielding
    return await fastMergeChunks(chunks, sortFn)
  }

  // Optimized merging with reduced yielding
  const fastMergeChunks = async (initialChunks, sortFn) => {
    let chunks = initialChunks
    while (chunks.length > 1) {
      const mergedChunks = []

      for (let i = 0; i < chunks.length; i += 2) {
        if (i + 1 < chunks.length) {
          const merged = fastMergeTwoArrays(chunks[i], chunks[i + 1], sortFn)
          mergedChunks.push(merged)
        } else {
          mergedChunks.push(chunks[i])
        }
      }

      chunks = mergedChunks

      // Only yield between major merge iterations
      if (chunks.length > 1) {
        await new Promise(resolve => setTimeout(resolve, 0))
      }
    }

    return chunks[0]
  }

  // Fast synchronous merge without yielding (for speed)
  const fastMergeTwoArrays = (arr1, arr2, sortFn) => {
    const result = []
    let i = 0
    let j = 0

    // Fast merge without yielding - this is the critical path
    while (i < arr1.length && j < arr2.length) {
      if (sortFn(arr1[i], arr2[j]) <= 0) {
        result.push(arr1[i])
        i++
      } else {
        result.push(arr2[j])
        j++
      }
    }

    result.push(...arr1.slice(i), ...arr2.slice(j))
    return result
  }

  // Clear sort cache when needed
  const clearSortCache = () => {
    sortCache.clear()
  }

  return {
    sortField,
    sortDirection,
    sortLargeArray,
    createSortFunction,
    clearSortCache
  }
}

// Intersection Observer for performance monitoring
export function useIntersectionObserver(options = {}) {
  const { threshold = [0, 0.25, 0.5, 0.75, 1], rootMargin = '50px' } = options

  const visibleElements = ref(new Set())
  const observer = ref(null)

  const createObserver = callback => {
    if (!window.IntersectionObserver) return null

    observer.value = new IntersectionObserver(
      entries => {
        for (const entry of entries) {
          if (entry.isIntersecting) {
            visibleElements.value.add(entry.target)
          } else {
            visibleElements.value.delete(entry.target)
          }
        }

        if (callback) callback(entries)
      },
      {
        threshold,
        rootMargin
      }
    )

    return observer.value
  }

  const observe = element => {
    if (observer.value && element) {
      observer.value.observe(element)
    }
  }

  const unobserve = element => {
    if (observer.value && element) {
      observer.value.unobserve(element)
      visibleElements.value.delete(element)
    }
  }

  const disconnect = () => {
    if (observer.value) {
      observer.value.disconnect()
      visibleElements.value.clear()
    }
  }

  return {
    visibleElements,
    createObserver,
    observe,
    unobserve,
    disconnect
  }
}
