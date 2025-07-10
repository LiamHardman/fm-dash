package main

import (
	"container/list"
	"sync"
	"time"
)

// CacheItem represents an item in the cache with expiration
type CacheItem struct {
	Value      interface{}
	Expiration time.Time
	HasExpiry  bool
	Size       int64 // Size in bytes (estimated)
	LastAccess time.Time
}

// LRUNode for doubly linked list
type LRUNode struct {
	Key   string
	Value *CacheItem
	Elem  *list.Element
}

// InMemoryCache represents an LRU cache with size limits
type InMemoryCache struct {
	items       map[string]*LRUNode
	lruList     *list.List
	mutex       sync.RWMutex
	cleanup     *time.Ticker
	done        chan bool
	maxSize     int64 // Maximum size in bytes
	currentSize int64 // Current size in bytes
	maxItems    int   // Maximum number of items
}

// Global in-memory cache
var memCache *InMemoryCache

// Default cache expiration times and limits
const (
	defaultExpiration = 90 * time.Second // Further reduced for aggressive cleanup
	noExpiration      = 0
	defaultMaxSize    = 32 * 1024 * 1024 // 32MB max size (reduced from 64MB for better memory usage)
	defaultMaxItems   = 750              // Maximum 750 items (reduced from 1,500 for lower memory footprint)
)

func InitInMemoryCache() {
	memCache = &InMemoryCache{
		items:    make(map[string]*LRUNode),
		lruList:  list.New(),
		done:     make(chan bool),
		maxSize:  defaultMaxSize,
		maxItems: defaultMaxItems,
	}

	// Start cleanup goroutine every 3 minutes (more frequent for better memory management)
	memCache.cleanup = time.NewTicker(3 * time.Minute)
	go memCache.cleanupExpired()

	LogDebug("Enhanced in-memory cache system initialized with aggressive LRU eviction (max size: %d MB, max items: %d)",
		defaultMaxSize/(1024*1024), defaultMaxItems)
}

func (c *InMemoryCache) cleanupExpired() {
	for {
		select {
		case <-c.cleanup.C:
			c.deleteExpired()
			c.enforceSizeLimits()
		case <-c.done:
			return
		}
	}
}

func (c *InMemoryCache) deleteExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	toDelete := make([]string, 0)

	for key, node := range c.items {
		if node.Value.HasExpiry && now.After(node.Value.Expiration) {
			toDelete = append(toDelete, key)
		}
	}

	for _, key := range toDelete {
		c.removeLRU(key)
	}

	if len(toDelete) > 0 {
		LogDebug("Expired %d cache items", len(toDelete))
	}
}

// enforceSizeLimits removes LRU items when cache exceeds limits
func (c *InMemoryCache) enforceSizeLimits() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	removedCount := 0
	removedSize := int64(0)

	// Remove items if over size or count limits
	for (c.currentSize > c.maxSize || len(c.items) > c.maxItems) && c.lruList.Len() > 0 {
		// Remove least recently used item
		oldest := c.lruList.Back()
		if oldest != nil {
			node := oldest.Value.(*LRUNode)
			removedSize += node.Value.Size
			removedCount++
			c.removeLRU(node.Key)
		} else {
			break
		}
	}

	if removedCount > 0 {
		LogDebug("LRU evicted %d items (%.2f MB) to enforce cache limits",
			removedCount, float64(removedSize)/(1024*1024))
	}
}

// removeLRU removes an item from both map and LRU list (must be called with lock held)
func (c *InMemoryCache) removeLRU(key string) {
	if node, exists := c.items[key]; exists {
		c.currentSize -= node.Value.Size
		c.lruList.Remove(node.Elem)
		delete(c.items, key)
	}
}

// moveToFront moves an item to front of LRU list (must be called with lock held)
func (c *InMemoryCache) moveToFront(node *LRUNode) {
	c.lruList.MoveToFront(node.Elem)
	node.Value.LastAccess = time.Now()
}

// estimateSize provides a rough estimate of object size in bytes
func estimateSize(value interface{}) int64 {
	switch v := value.(type) {
	case string:
		return int64(len(v)) + 16 // string overhead
	case []byte:
		return int64(len(v)) + 24 // slice overhead
	case map[string]interface{}:
		size := int64(48) // map overhead
		for key, val := range v {
			size += int64(len(key)) + 16 // key
			size += estimateSize(val)    // recursive estimation
		}
		return size
	case []interface{}:
		size := int64(24) // slice overhead
		for _, item := range v {
			size += estimateSize(item)
		}
		return size
	default:
		return 64 // default estimate for unknown types
	}
}

func getFromMemCache(key string) (interface{}, bool) {
	if memCache == nil {
		return nil, false
	}

	memCache.mutex.Lock()
	defer memCache.mutex.Unlock()

	node, found := memCache.items[key]
	if !found {
		return nil, false
	}

	if node.Value.HasExpiry && time.Now().After(node.Value.Expiration) {
		// Expired item - remove it
		memCache.removeLRU(key)
		return nil, false
	}

	// Move to front of LRU list
	memCache.moveToFront(node)
	return node.Value.Value, true
}

func setInMemCache(key string, value interface{}, expiration time.Duration) {
	setInMemCacheWithOptions(key, value, expiration, false)
}

// setInMemCacheForDataset stores dataset items without size restrictions
func setInMemCacheForDataset(key string, value interface{}, expiration time.Duration) {
	setInMemCacheWithOptions(key, value, expiration, true)
}

// setInMemCacheWithOptions allows bypassing size restrictions for datasets
func setInMemCacheWithOptions(key string, value interface{}, expiration time.Duration, bypassSizeCheck bool) {
	if memCache == nil {
		LogWarn("Warning: Memory cache not initialized, cannot set key: %s", sanitizeForLogging(key))
		return
	}

	memCache.mutex.Lock()
	defer memCache.mutex.Unlock()

	// Calculate estimated size
	estimatedSize := estimateSize(value)

	// Check if single item exceeds max cache size (unless bypassing for datasets)
	if !bypassSizeCheck && estimatedSize > memCache.maxSize/4 { // Changed from /2 to /4 for more lenient limits
		LogWarn("Item too large for cache (%.2f MB), skipping: %s",
			float64(estimatedSize)/(1024*1024), sanitizeForLogging(key))
		return
	}

	// For very large datasets, log but allow caching
	if bypassSizeCheck && estimatedSize > memCache.maxSize/2 {
		LogInfo("Caching large dataset (%.2f MB): %s",
			float64(estimatedSize)/(1024*1024), sanitizeForLogging(key))
	}

	item := &CacheItem{
		Value:      value,
		Size:       estimatedSize,
		LastAccess: time.Now(),
	}

	if expiration == noExpiration {
		item.HasExpiry = false
	} else {
		item.HasExpiry = true
		if expiration == 0 {
			expiration = defaultExpiration
		}
		item.Expiration = time.Now().Add(expiration)
	}

	// If key already exists, remove old version first
	if _, exists := memCache.items[key]; exists {
		memCache.removeLRU(key)
	}

	// Create new LRU node
	elem := memCache.lruList.PushFront(&LRUNode{Key: key, Value: item})
	node := &LRUNode{
		Key:   key,
		Value: item,
		Elem:  elem,
	}
	elem.Value = node
	memCache.items[key] = node
	memCache.currentSize += estimatedSize

	// For datasets, be more lenient with size enforcement
	maxSizeThreshold := memCache.maxSize
	if bypassSizeCheck {
		maxSizeThreshold = memCache.maxSize * 2 // Allow cache to grow larger for datasets
	}

	// Enforce size limits after adding
	if memCache.currentSize > maxSizeThreshold || len(memCache.items) > memCache.maxItems {
		go func() {
			time.Sleep(10 * time.Millisecond) // Brief delay to avoid blocking
			memCache.enforceSizeLimits()
		}()
	}
}

func deleteFromMemCache(key string) {
	if memCache == nil {
		return
	}

	memCache.mutex.Lock()
	defer memCache.mutex.Unlock()
	memCache.removeLRU(key)
}

func getMemCacheStats() (itemCount, totalCount int) {
	if memCache == nil {
		return 0, 0
	}

	memCache.mutex.RLock()
	defer memCache.mutex.RUnlock()
	count := len(memCache.items)
	return count, count
}

// GetMemCacheDetailedStats returns detailed cache statistics
func GetMemCacheDetailedStats() map[string]interface{} {
	if memCache == nil {
		return map[string]interface{}{
			"initialized": false,
		}
	}

	memCache.mutex.RLock()
	defer memCache.mutex.RUnlock()

	return map[string]interface{}{
		"initialized":   true,
		"item_count":    len(memCache.items),
		"max_items":     memCache.maxItems,
		"current_size":  memCache.currentSize,
		"max_size":      memCache.maxSize,
		"size_mb":       float64(memCache.currentSize) / (1024 * 1024),
		"max_size_mb":   float64(memCache.maxSize) / (1024 * 1024),
		"usage_percent": float64(memCache.currentSize) / float64(memCache.maxSize) * 100,
	}
}

func StopMemCache() {
	if memCache != nil && memCache.cleanup != nil {
		memCache.cleanup.Stop()
		close(memCache.done)
	}
}
