package main

import (
	"net/http"
)

// GetFormatAwareCacheHandler returns the appropriate player data handler based on configuration
func GetFormatAwareCacheHandler() http.HandlerFunc {
	initFormatAwareCacheConfig()
	
	if formatAwareCacheEnabled {
		LogInfo("Using format-aware player data handler")
		return formatAwarePlayerDataHandler
	}
	
	LogInfo("Using standard player data handler")
	return playerDataHandler
}