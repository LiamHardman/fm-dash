/**
 * IndexedMap - Efficient data structure for O(1) lookups with multiple indexes
 * Optimized for large datasets with frequent lookups and filtering
 */

export class IndexedMap {
  constructor(options = {}) {
    this.primaryKey = options.primaryKey || 'id'
    this.indexes = new Map() // Map of index name to Map of value to Set of items
    this.data = new Map() // Primary storage: key -> item
    this.reverseIndexes = new Map() // Map of item key to Set of index entries
    this.size = 0
    
    // Performance tracking
    this.stats = {
      lookups: 0,
      indexHits: 0,
      indexMisses: 0,
      rebuilds: 0
    }
    
    // Initialize indexes
    if (options.indexes) {
      for (const indexName of options.indexes) {
        this.createIndex(indexName)
      }
    }
  }

  /**
   * Create a new index on a field
   */
  createIndex(fieldName, options = {}) {
    if (this.indexes.has(fieldName)) {
      return // Index already exists
    }

    const indexMap = new Map()
    this.indexes.set(fieldName, indexMap)

    // Build index for existing data
    if (this.size > 0) {
      this._rebuildIndex(fieldName)
      this.stats.rebuilds++
    }

    return this
  }

  /**
   * Remove an index
   */
  removeIndex(fieldName) {
    if (this.indexes.has(fieldName)) {
      this.indexes.delete(fieldName)
      
      // Clean up reverse indexes
      for (const [itemKey, indexEntries] of this.reverseIndexes) {
        indexEntries.delete(`${fieldName}:*`)
      }
    }
    return this
  }

  /**
   * Add or update an item
   */
  set(item) {
    if (!item || typeof item !== 'object') {
      throw new Error('Item must be an object')
    }

    const key = item[this.primaryKey]
    if (key === undefined || key === null) {
      throw new Error(`Item must have a ${this.primaryKey} property`)
    }

    const existingItem = this.data.get(key)
    const isUpdate = existingItem !== undefined

    // Remove old index entries if updating
    if (isUpdate) {
      this._removeFromIndexes(key, existingItem)
    } else {
      this.size++
    }

    // Store the item
    this.data.set(key, item)

    // Add to indexes
    this._addToIndexes(key, item)

    return this
  }

  /**
   * Get an item by primary key
   */
  get(key) {
    this.stats.lookups++
    return this.data.get(key)
  }

  /**
   * Check if an item exists
   */
  has(key) {
    return this.data.has(key)
  }

  /**
   * Delete an item
   */
  delete(key) {
    const item = this.data.get(key)
    if (!item) return false

    this.data.delete(key)
    this._removeFromIndexes(key, item)
    this.size--
    return true
  }

  /**
   * Clear all data
   */
  clear() {
    this.data.clear()
    this.indexes.clear()
    this.reverseIndexes.clear()
    this.size = 0
    this.stats = { lookups: 0, indexHits: 0, indexMisses: 0, rebuilds: 0 }
  }

  /**
   * Find items by index - O(1) lookup
   */
  findByIndex(fieldName, value) {
    this.stats.lookups++
    
    const index = this.indexes.get(fieldName)
    if (!index) {
      this.stats.indexMisses++
      return []
    }

    const itemKeys = index.get(value)
    if (!itemKeys) {
      this.stats.indexMisses++
      return []
    }

    this.stats.indexHits++
    return Array.from(itemKeys).map(key => this.data.get(key)).filter(Boolean)
  }

  /**
   * Find items matching multiple criteria
   */
  findWhere(criteria) {
    this.stats.lookups++
    
    const criteriaKeys = Object.keys(criteria)
    if (criteriaKeys.length === 0) {
      return this.values()
    }

    // Find the most selective index to start with
    let smallestResultSet = null
    let smallestSize = Infinity
    let selectedField = null

    for (const field of criteriaKeys) {
      if (this.indexes.has(field)) {
        const index = this.indexes.get(field)
        const itemKeys = index.get(criteria[field])
        if (itemKeys && itemKeys.size < smallestSize) {
          smallestSize = itemKeys.size
          smallestResultSet = itemKeys
          selectedField = field
        }
      }
    }

    if (!smallestResultSet) {
      // No indexes available, fall back to linear search
      this.stats.indexMisses++
      return this.values().filter(item => {
        return criteriaKeys.every(key => item[key] === criteria[key])
      })
    }

    this.stats.indexHits++

    // Filter the smallest result set by remaining criteria
    const results = []
    for (const key of smallestResultSet) {
      const item = this.data.get(key)
      if (item && criteriaKeys.every(field => 
        field === selectedField || item[field] === criteria[field]
      )) {
        results.push(item)
      }
    }

    return results
  }

  /**
   * Get all values as array
   */
  values() {
    return Array.from(this.data.values())
  }

  /**
   * Get all keys as array
   */
  keys() {
    return Array.from(this.data.keys())
  }

  /**
   * Iterate over entries
   */
  entries() {
    return this.data.entries()
  }

  /**
   * Get unique values for a field
   */
  getUniqueValues(fieldName) {
    const index = this.indexes.get(fieldName)
    if (index) {
      return Array.from(index.keys())
    }

    // Fall back to scanning all items
    const uniqueValues = new Set()
    for (const item of this.data.values()) {
      const value = item[fieldName]
      if (value !== undefined && value !== null) {
        uniqueValues.add(value)
      }
    }
    return Array.from(uniqueValues)
  }

  /**
   * Get statistics about the IndexedMap
   */
  getStats() {
    return {
      ...this.stats,
      size: this.size,
      indexes: this.indexes.size,
      indexSizes: Object.fromEntries(
        Array.from(this.indexes.entries()).map(([name, index]) => [name, index.size])
      )
    }
  }

  /**
   * Bulk insert items efficiently
   */
  bulkSet(items) {
    if (!Array.isArray(items)) {
      throw new Error('Items must be an array')
    }

    const startTime = performance.now()
    
    for (const item of items) {
      this.set(item)
    }

    const endTime = performance.now()
    return {
      inserted: items.length,
      timeMs: endTime - startTime,
      itemsPerSecond: Math.round(items.length / ((endTime - startTime) / 1000))
    }
  }

  /**
   * Add item to all relevant indexes
   */
  _addToIndexes(key, item) {
    const indexEntries = new Set()

    for (const [fieldName, index] of this.indexes) {
      const value = item[fieldName]
      if (value !== undefined && value !== null) {
        if (!index.has(value)) {
          index.set(value, new Set())
        }
        index.get(value).add(key)
        indexEntries.add(`${fieldName}:${value}`)
      }
    }

    this.reverseIndexes.set(key, indexEntries)
  }

  /**
   * Remove item from all relevant indexes
   */
  _removeFromIndexes(key, item) {
    const indexEntries = this.reverseIndexes.get(key)
    if (!indexEntries) return

    for (const [fieldName, index] of this.indexes) {
      const value = item[fieldName]
      if (value !== undefined && value !== null) {
        const itemSet = index.get(value)
        if (itemSet) {
          itemSet.delete(key)
          if (itemSet.size === 0) {
            index.delete(value)
          }
        }
      }
    }

    this.reverseIndexes.delete(key)
  }

  /**
   * Rebuild a specific index
   */
  _rebuildIndex(fieldName) {
    const index = this.indexes.get(fieldName)
    if (!index) return

    index.clear()

    for (const [key, item] of this.data) {
      const value = item[fieldName]
      if (value !== undefined && value !== null) {
        if (!index.has(value)) {
          index.set(value, new Set())
        }
        index.get(value).add(key)
      }
    }
  }

  /**
   * Optimize indexes by removing empty entries
   */
  optimize() {
    let removedEntries = 0

    for (const [fieldName, index] of this.indexes) {
      const keysToRemove = []
      for (const [value, itemSet] of index) {
        if (itemSet.size === 0) {
          keysToRemove.push(value)
        }
      }
      
      for (const key of keysToRemove) {
        index.delete(key)
        removedEntries++
      }
    }

    return { removedEntries }
  }

  /**
   * Create a filtered view without copying data
   */
  createView(filterFn) {
    return new IndexedMapView(this, filterFn)
  }
}

/**
 * IndexedMapView - A filtered view of an IndexedMap without copying data
 */
export class IndexedMapView {
  constructor(parentMap, filterFn) {
    this.parent = parentMap
    this.filter = filterFn
    this._cachedKeys = null
    this._cacheValid = false
  }

  /**
   * Get filtered keys (cached)
   */
  _getFilteredKeys() {
    if (!this._cacheValid) {
      this._cachedKeys = Array.from(this.parent.keys()).filter(key => {
        const item = this.parent.get(key)
        return item && this.filter(item)
      })
      this._cacheValid = true
    }
    return this._cachedKeys
  }

  /**
   * Invalidate cache
   */
  invalidateCache() {
    this._cacheValid = false
    this._cachedKeys = null
  }

  /**
   * Get filtered values
   */
  values() {
    return this._getFilteredKeys().map(key => this.parent.get(key))
  }

  /**
   * Get filtered size
   */
  get size() {
    return this._getFilteredKeys().length
  }

  /**
   * Check if key exists in filtered view
   */
  has(key) {
    const item = this.parent.get(key)
    return item && this.filter(item)
  }

  /**
   * Get item if it passes filter
   */
  get(key) {
    const item = this.parent.get(key)
    return (item && this.filter(item)) ? item : undefined
  }
}

export default IndexedMap