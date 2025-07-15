/**
 * Advanced Virtual Scrolling Manager
 * Provides optimized viewport calculations, item recycling, and memory pooling
 */

export class VirtualScrollManager {
  constructor(options = {}) {
    // Configuration
    this.itemHeight = options.itemHeight || 30
    this.containerHeight = options.containerHeight || 400
    this.bufferSize = options.bufferSize || 10
    this.overscan = options.overscan || 5
    this.recycleThreshold = options.recycleThreshold || 100
    this.enableVariableHeight = options.enableVariableHeight || false
    this.enableMomentum = options.enableMomentum || true

    // State
    this.scrollTop = 0
    this.isScrolling = false
    this.scrollVelocity = 0
    this.lastScrollTime = 0
    this.lastScrollTop = 0
    this.momentumAnimationId = null

    // Item recycling pools
    this.itemPool = new Map()
    this.recycledItems = []
    this.activeItems = new Map()

    // Variable height support
    this.itemHeights = new Map()
    this.estimatedItemHeight = this.itemHeight
    this.totalHeight = 0

    // Performance tracking
    this.renderCount = 0
    this.recycleCount = 0
    this.poolHits = 0

    // Callbacks
    this.onScroll = null
    this.onVisibleRangeChange = null
    this.onItemRecycle = null
  }

  /**
   * Calculate visible range with optimized viewport calculations
   */
  calculateVisibleRange(totalItems) {
    if (totalItems === 0) {
      return { startIndex: 0, endIndex: 0, visibleStartIndex: 0, visibleEndIndex: 0 }
    }

    let startIndex, endIndex

    if (this.enableVariableHeight) {
      // Variable height calculation
      const result = this._calculateVariableHeightRange(totalItems)
      startIndex = result.startIndex
      endIndex = result.endIndex
    } else {
      // Fixed height calculation (optimized)
      const visibleItemCount = Math.ceil(this.containerHeight / this.itemHeight)
      const scrollIndex = Math.floor(this.scrollTop / this.itemHeight)

      startIndex = Math.max(0, scrollIndex - this.bufferSize)
      endIndex = Math.min(totalItems, scrollIndex + visibleItemCount + this.bufferSize)
    }

    // Add overscan for smoother scrolling
    const overscanStart = Math.max(0, startIndex - this.overscan)
    const overscanEnd = Math.min(totalItems, endIndex + this.overscan)

    const visibleRange = {
      startIndex: overscanStart,
      endIndex: overscanEnd,
      visibleStartIndex: startIndex,
      visibleEndIndex: endIndex,
      totalItems
    }

    // Notify range change
    if (this.onVisibleRangeChange) {
      this.onVisibleRangeChange(visibleRange)
    }

    return visibleRange
  }

  /**
   * Variable height range calculation
   */
  _calculateVariableHeightRange(totalItems) {
    let accumulatedHeight = 0
    let startIndex = 0
    let endIndex = totalItems

    // Find start index
    for (let i = 0; i < totalItems; i++) {
      const itemHeight = this.getItemHeight(i)
      if (
        accumulatedHeight + itemHeight >
        this.scrollTop - this.bufferSize * this.estimatedItemHeight
      ) {
        startIndex = Math.max(0, i - this.bufferSize)
        break
      }
      accumulatedHeight += itemHeight
    }

    // Find end index
    const targetHeight =
      this.scrollTop + this.containerHeight + this.bufferSize * this.estimatedItemHeight
    accumulatedHeight = this.getOffsetForIndex(startIndex)

    for (let i = startIndex; i < totalItems; i++) {
      if (accumulatedHeight > targetHeight) {
        endIndex = Math.min(totalItems, i + this.bufferSize)
        break
      }
      accumulatedHeight += this.getItemHeight(i)
    }

    return { startIndex, endIndex }
  }

  /**
   * Get item height (supports variable heights)
   */
  getItemHeight(index) {
    if (this.enableVariableHeight && this.itemHeights.has(index)) {
      return this.itemHeights.get(index)
    }
    return this.estimatedItemHeight
  }

  /**
   * Set item height for variable height support
   */
  setItemHeight(index, height) {
    if (this.enableVariableHeight) {
      this.itemHeights.set(index, height)
      this._updateTotalHeight()
    }
  }

  /**
   * Get offset for specific index
   */
  getOffsetForIndex(index) {
    if (!this.enableVariableHeight) {
      return index * this.itemHeight
    }

    let offset = 0
    for (let i = 0; i < index; i++) {
      offset += this.getItemHeight(i)
    }
    return offset
  }

  /**
   * Update total height for variable height items
   */
  _updateTotalHeight() {
    if (!this.enableVariableHeight) return

    this.totalHeight = 0
    for (const [index, _height] of this.itemHeights) {
      this.totalHeight = Math.max(this.totalHeight, (index + 1) * this.estimatedItemHeight)
    }
  }

  /**
   * Recycle items with memory pooling
   */
  recycleItems(items, visibleRange) {
    const { startIndex, endIndex } = visibleRange
    const recycledItems = []
    const newActiveItems = new Map()

    // Process visible items
    for (let i = startIndex; i < endIndex && i < items.length; i++) {
      const item = items[i]
      let recycledItem = this.activeItems.get(i)

      if (!recycledItem) {
        // Try to get from pool
        recycledItem = this._getFromPool(item.type || 'default')
        if (recycledItem) {
          this.poolHits++
        } else {
          // Create new item
          recycledItem = this._createNewItem(item, i)
        }
      }

      // Update item data
      recycledItem.data = item
      recycledItem.index = i
      recycledItem.virtualIndex = i - startIndex
      recycledItem.offset = this.getOffsetForIndex(i)
      recycledItem.height = this.getItemHeight(i)

      recycledItems.push(recycledItem)
      newActiveItems.set(i, recycledItem)
    }

    // Return unused items to pool
    for (const [index, item] of this.activeItems) {
      if (!newActiveItems.has(index)) {
        this._returnToPool(item)
        this.recycleCount++
      }
    }

    this.activeItems = newActiveItems
    this.renderCount++

    if (this.onItemRecycle) {
      this.onItemRecycle({
        recycledCount: this.recycleCount,
        poolHits: this.poolHits,
        activeItems: this.activeItems.size
      })
    }

    return recycledItems
  }

  /**
   * Get item from pool
   */
  _getFromPool(type) {
    const pool = this.itemPool.get(type)
    if (pool && pool.length > 0) {
      return pool.pop()
    }
    return null
  }

  /**
   * Return item to pool
   */
  _returnToPool(item) {
    const type = item.type || 'default'
    if (!this.itemPool.has(type)) {
      this.itemPool.set(type, [])
    }

    const pool = this.itemPool.get(type)
    if (pool.length < this.recycleThreshold) {
      // Reset item state
      item.data = null
      item.index = -1
      item.virtualIndex = -1
      pool.push(item)
    }
  }

  /**
   * Create new item
   */
  _createNewItem(data, index) {
    return {
      id: `item-${index}-${Date.now()}`,
      type: data.type || 'default',
      data: null,
      index: -1,
      virtualIndex: -1,
      offset: 0,
      height: this.estimatedItemHeight,
      element: null
    }
  }

  /**
   * Handle scroll with momentum and smooth scrolling
   */
  handleScroll(scrollTop, timestamp = performance.now()) {
    const deltaTime = timestamp - this.lastScrollTime
    const deltaScroll = scrollTop - this.lastScrollTop

    // Calculate velocity
    if (deltaTime > 0) {
      this.scrollVelocity = deltaScroll / deltaTime
    }

    this.scrollTop = scrollTop
    this.isScrolling = true
    this.lastScrollTime = timestamp
    this.lastScrollTop = scrollTop

    // Apply momentum scrolling if enabled
    if (this.enableMomentum && Math.abs(this.scrollVelocity) > 0.1) {
      this._applyMomentum()
    }

    // Notify scroll
    if (this.onScroll) {
      this.onScroll({
        scrollTop,
        velocity: this.scrollVelocity,
        isScrolling: this.isScrolling
      })
    }

    // Clear scrolling state after delay
    clearTimeout(this.scrollTimeout)
    this.scrollTimeout = setTimeout(() => {
      this.isScrolling = false
      this.scrollVelocity = 0
    }, 150)
  }

  /**
   * Apply momentum scrolling with inertia
   */
  _applyMomentum() {
    if (this.momentumAnimationId) {
      cancelAnimationFrame(this.momentumAnimationId)
    }

    const applyInertia = () => {
      if (Math.abs(this.scrollVelocity) < 0.01) {
        this.scrollVelocity = 0
        return
      }

      // Apply friction
      this.scrollVelocity *= 0.95

      // Continue momentum
      this.momentumAnimationId = requestAnimationFrame(applyInertia)
    }

    this.momentumAnimationId = requestAnimationFrame(applyInertia)
  }

  /**
   * Get total height for container
   */
  getTotalHeight(itemCount) {
    if (this.enableVariableHeight) {
      return this.totalHeight || itemCount * this.estimatedItemHeight
    }
    return itemCount * this.itemHeight
  }

  /**
   * Optimize memory usage
   */
  optimizeMemoryUsage() {
    // Clear oversized pools
    for (const [type, pool] of this.itemPool) {
      if (pool.length > this.recycleThreshold) {
        pool.splice(this.recycleThreshold)
      }
    }

    // Clear old height measurements if too many
    if (this.itemHeights.size > 10000) {
      const entries = Array.from(this.itemHeights.entries())
      entries.sort((a, b) => b[0] - a[0]) // Sort by index descending

      // Keep only recent measurements
      this.itemHeights.clear()
      for (let i = 0; i < Math.min(5000, entries.length); i++) {
        this.itemHeights.set(entries[i][0], entries[i][1])
      }
    }
  }

  /**
   * Get performance statistics
   */
  getStats() {
    return {
      renderCount: this.renderCount,
      recycleCount: this.recycleCount,
      poolHits: this.poolHits,
      poolSizes: Object.fromEntries(
        Array.from(this.itemPool.entries()).map(([type, pool]) => [type, pool.length])
      ),
      activeItems: this.activeItems.size,
      measuredHeights: this.itemHeights.size,
      scrollVelocity: this.scrollVelocity,
      isScrolling: this.isScrolling
    }
  }

  /**
   * Reset manager state
   */
  reset() {
    this.scrollTop = 0
    this.isScrolling = false
    this.scrollVelocity = 0
    this.activeItems.clear()
    this.itemHeights.clear()
    this.renderCount = 0
    this.recycleCount = 0
    this.poolHits = 0

    if (this.momentumAnimationId) {
      cancelAnimationFrame(this.momentumAnimationId)
      this.momentumAnimationId = null
    }
  }

  /**
   * Destroy manager and cleanup resources
   */
  destroy() {
    this.reset()
    this.itemPool.clear()
    this.recycledItems.length = 0

    // Clear callbacks
    this.onScroll = null
    this.onVisibleRangeChange = null
    this.onItemRecycle = null
  }
}

export default VirtualScrollManager
