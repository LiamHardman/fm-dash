package main

import (
	"log"
	"sync"
	"time"
)

// CacheItem represents an item in the cache with expiration
type CacheItem struct {
	Value      interface{}
	Expiration time.Time
	HasExpiry  bool
}

// InMemoryCache represents a simple in-memory cache
type InMemoryCache struct {
	items   map[string]CacheItem
	mutex   sync.RWMutex
	cleanup *time.Ticker
	done    chan bool
}

// Global in-memory cache
var memCache *InMemoryCache

// Default cache expiration times
const (
	defaultExpiration = 5 * time.Minute
	noExpiration      = 0
)

func InitInMemoryCache() {
	memCache = &InMemoryCache{
		items: make(map[string]CacheItem),
		done:  make(chan bool),
	}

	// Start cleanup goroutine every 10 minutes
	memCache.cleanup = time.NewTicker(10 * time.Minute)
	go memCache.cleanupExpired()

	log.Println("In-memory cache system initialized")
}

func (c *InMemoryCache) cleanupExpired() {
	for {
		select {
		case <-c.cleanup.C:
			c.deleteExpired()
		case <-c.done:
			return
		}
	}
}

func (c *InMemoryCache) deleteExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if item.HasExpiry && now.After(item.Expiration) {
			delete(c.items, key)
		}
	}
}

func getFromMemCache(key string) (interface{}, bool) {
	if memCache == nil {
		return nil, false
	}

	memCache.mutex.RLock()
	defer memCache.mutex.RUnlock()

	item, found := memCache.items[key]
	if !found {
		return nil, false
	}

	if item.HasExpiry && time.Now().After(item.Expiration) {
		go func() {
			memCache.mutex.Lock()
			delete(memCache.items, key)
			memCache.mutex.Unlock()
		}()
		return nil, false
	}

	return item.Value, true
}

func setInMemCache(key string, value interface{}, expiration time.Duration) {
	if memCache == nil {
		log.Printf("Warning: Memory cache not initialized, cannot set key: %s", key)
		return
	}

	memCache.mutex.Lock()
	defer memCache.mutex.Unlock()

	item := CacheItem{
		Value: value,
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

	memCache.items[key] = item
}

func deleteFromMemCache(key string) {
	if memCache == nil {
		return
	}

	memCache.mutex.Lock()
	defer memCache.mutex.Unlock()
	delete(memCache.items, key)
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

func StopMemCache() {
	if memCache != nil && memCache.cleanup != nil {
		memCache.cleanup.Stop()
		close(memCache.done)
	}
}
