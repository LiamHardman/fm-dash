package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"time"
)

const (
	configLoadTimeout = 5 * time.Second
)

func cachedRolesHandler(w http.ResponseWriter, r *http.Request) {
	const cacheKey = "roles_data"

	if cached, found := getFromMemCache(cacheKey); found {
		if roles, ok := cached.([]string); ok {
			log.Printf("Retrieved roles data from memory cache")
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Cache-Source", "memory")
			w.Header().Set("Cache-Control", "public, max-age=86400")
			if err := json.NewEncoder(w).Encode(roles); err != nil {
				log.Printf("Error encoding roles response: %v", err)
			}
			return
		}
	}

	muRoleSpecificOverallWeights.RLock()
	roleNames := make([]string, 0, len(roleSpecificOverallWeights))
	for roleName := range roleSpecificOverallWeights {
		roleNames = append(roleNames, roleName)
	}
	muRoleSpecificOverallWeights.RUnlock()
	sort.Strings(roleNames)

	setInMemCache(cacheKey, roleNames, noExpiration)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	if err := json.NewEncoder(w).Encode(roleNames); err != nil {
		log.Printf("Error encoding role names response: %v", err)
	}
}

func cachedConfigHandler(w http.ResponseWriter, r *http.Request) {
	if err := EnsureConfigInitialized(configLoadTimeout); err != nil {
		log.Printf("Configuration not ready for config request: %v", err)
		http.Error(w, "Configuration not ready, please try again later.", http.StatusServiceUnavailable)
		return
	}
	const cacheKey = "config_data"

	switch r.Method {
	case http.MethodGet:
		if cached, found := getFromMemCache(cacheKey); found {
			if config, ok := cached.(map[string]interface{}); ok {
				log.Printf("Retrieved config data from memory cache")
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Cache-Source", "memory")
				w.Header().Set("Cache-Control", "public, max-age=3600")
				if err := json.NewEncoder(w).Encode(config); err != nil {
					log.Printf("Error encoding config response: %v", err)
				}
				return
			}
		}

		config := map[string]interface{}{
			"maxUploadSizeMB":      getMaxUploadSize() / (1024 * 1024),
			"maxUploadSizeBytes":   getMaxUploadSize(),
			"useScaledRatings":     GetUseScaledRatings(),
			"datasetRetentionDays": int(getRetentionPeriod().Hours() / 24),
		}

		setInMemCache(cacheKey, config, defaultExpiration)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(config); err != nil {
			log.Printf("Error encoding config response: %v", err)
		}

	case http.MethodPost:
		var updateRequest struct {
			UseScaledRatings *bool `json:"useScaledRatings"`
		}

		if err := json.NewDecoder(r.Body).Decode(&updateRequest); err != nil {
			http.Error(w, "Error parsing request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		if updateRequest.UseScaledRatings != nil {
			SetUseScaledRatings(*updateRequest.UseScaledRatings)
			if *updateRequest.UseScaledRatings {
				log.Printf("Rating calculation method updated via API: enabled scaled ratings")
			} else {
				log.Printf("Rating calculation method updated via API: disabled scaled ratings")
			}
			deleteFromMemCache(cacheKey)
		}

		config := map[string]interface{}{
			"maxUploadSizeMB":      getMaxUploadSize() / (1024 * 1024),
			"maxUploadSizeBytes":   getMaxUploadSize(),
			"useScaledRatings":     GetUseScaledRatings(),
			"datasetRetentionDays": int(getRetentionPeriod().Hours() / 24),
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(config); err != nil {
			log.Printf("Error encoding config response: %v", err)
		}

	default:
		http.Error(w, "Only GET and POST methods are allowed", http.StatusMethodNotAllowed)
	}
}
