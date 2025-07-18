package main

import (
	"log"
	"os"
	"strings"
	"sync"
)

var (
	// formatAwareCacheEnabled controls whether format-aware caching is enabled
	formatAwareCacheEnabled bool
	
	// formatAwareCacheConfig holds configuration for format-aware caching
	formatAwareCacheConfig struct {
		// EnableProtobufOptimization controls whether protobuf-specific memory optimizations are applied
		EnableProtobufOptimization bool
		
		// SeparateCacheKeys controls whether to use separate cache keys for different formats
		SeparateCacheKeys bool
		
		// DefaultCacheDuration is the default cache duration for format-aware cache items
		DefaultCacheDuration string
	}
	
	// formatAwareCacheConfigOnce ensures the configuration is loaded only once
	formatAwareCacheConfigOnce sync.Once
)

// initFormatAwareCacheConfig initializes the format-aware cache configuration
func initFormatAwareCacheConfig() {
	formatAwareCacheConfigOnce.Do(func() {
		// Check if format-aware caching is enabled
		formatAwareCacheEnabled = strings.ToLower(os.Getenv("FORMAT_AWARE_CACHE_ENABLED")) == "true"
		
		// If not enabled, return early
		if !formatAwareCacheEnabled {
			log.Println("Format-aware caching is disabled")
			return
		}
		
		// Load configuration values
		formatAwareCacheConfig.EnableProtobufOptimization = 
			strings.ToLower(os.Getenv("FORMAT_AWARE_CACHE_PROTOBUF_OPTIMIZATION")) != "false"
		
		formatAwareCacheConfig.SeparateCacheKeys = 
			strings.ToLower(os.Getenv("FORMAT_AWARE_CACHE_SEPARATE_KEYS")) != "false"
		
		cacheDuration := os.Getenv("FORMAT_AWARE_CACHE_DURATION")
		if cacheDuration == "" {
			cacheDuration = "5m" // Default to 5 minutes
		}
		formatAwareCacheConfig.DefaultCacheDuration = cacheDuration
		
		log.Printf("Format-aware caching initialized with config: %+v", formatAwareCacheConfig)
	})
}

// IsFormatAwareCacheEnabled returns whether format-aware caching is enabled
func IsFormatAwareCacheEnabled() bool {
	initFormatAwareCacheConfig()
	return formatAwareCacheEnabled
}

// The GetFormatAwareCacheHandler function is now in format_aware_cache_handler.go