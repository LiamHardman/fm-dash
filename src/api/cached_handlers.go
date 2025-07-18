package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"time"
	
	pb "api/proto"
)

const (
	configLoadTimeout = 5 * time.Second
)

func cachedRolesHandler(w http.ResponseWriter, r *http.Request) {
	const baseCacheKey = "roles_data"
	
	// Determine the appropriate format based on the request
	format := GetCacheFormatFromRequest(r)
	
	// Initialize content negotiation
	negotiator := NewContentNegotiator(r)
	serializer := negotiator.SelectSerializer()
	
	// Get request ID for response metadata
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = generateRequestID()
	}
	
	// Try to get from format-specific cache
	if cached, found := GetFormatAwareCacheItem(baseCacheKey, format); found {
		if format == FormatTypeJSON {
			// JSON format cache hit
			if roles, ok := cached.([]string); ok {
				LogDebug("Retrieved roles data from memory cache (JSON format)")
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Cache-Source", "memory")
				w.Header().Set("X-Cache-Format", "json")
				w.Header().Set("Cache-Control", "public, max-age=86400")
				if err := json.NewEncoder(w).Encode(roles); err != nil {
					log.Printf("Error encoding roles response: %v", err)
				}
				return
			}
		} else if format == FormatTypeProtobuf {
			// Protobuf format cache hit
			if protoResponse, ok := cached.(*pb.RolesResponse); ok {
				LogDebug("Retrieved roles data from memory cache (Protobuf format)")
				
				// Serialize the protobuf response
				responseData, err := serializer.Serialize(protoResponse)
				if err != nil {
					// Fallback to JSON on serialization error
					log.Printf("Error serializing cached protobuf roles: %v, falling back to JSON", err)
					// Continue to regenerate the data in JSON format
				} else {
					w.Header().Set("Content-Type", serializer.ContentType())
					w.Header().Set("X-Cache-Source", "memory")
					w.Header().Set("X-Cache-Format", "protobuf")
					w.Header().Set("Cache-Control", "public, max-age=86400")
					w.Write(responseData)
					return
				}
			}
		}
	}

	// Cache miss or serialization error, generate the data
	muRoleSpecificOverallWeights.RLock()
	roleNames := make([]string, 0, len(roleSpecificOverallWeights))
	for roleName := range roleSpecificOverallWeights {
		roleNames = append(roleNames, roleName)
	}
	muRoleSpecificOverallWeights.RUnlock()
	sort.Strings(roleNames)
	
	// Store in format-specific caches
	// For JSON format
	SetFormatAwareCacheItem(baseCacheKey, FormatTypeJSON, roleNames, noExpiration)
	
	// For Protobuf format
	protoResponse := &pb.RolesResponse{
		Roles: roleNames,
		Metadata: CreateResponseMetadata(requestID, int32(len(roleNames)), false),
	}
	// Optimize memory usage for protobuf cached data
	optimizedProtoResponse := OptimizeMemoryForProtobuf(protoResponse)
	SetFormatAwareCacheItem(baseCacheKey, FormatTypeProtobuf, optimizedProtoResponse, noExpiration)
	
	// Respond with the appropriate format
	if format == FormatTypeProtobuf {
		responseData, err := serializer.Serialize(protoResponse)
		if err != nil {
			// Fallback to JSON on serialization error
			log.Printf("Error serializing protobuf roles: %v, falling back to JSON", err)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "public, max-age=86400")
			if err := json.NewEncoder(w).Encode(roleNames); err != nil {
				log.Printf("Error encoding role names response: %v", err)
			}
			return
		}
		
		w.Header().Set("Content-Type", serializer.ContentType())
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Write(responseData)
	} else {
		// Default JSON response
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		if err := json.NewEncoder(w).Encode(roleNames); err != nil {
			log.Printf("Error encoding role names response: %v", err)
		}
	}
}

func cachedConfigHandler(w http.ResponseWriter, r *http.Request) {
	if err := EnsureConfigInitialized(configLoadTimeout); err != nil {
		log.Printf("Configuration not ready for config request: %v", err)
		http.Error(w, "Configuration not ready, please try again later.", http.StatusServiceUnavailable)
		return
	}
	const baseCacheKey = "config_data"
	
	// Determine the appropriate format based on the request
	format := GetCacheFormatFromRequest(r)
	
	// Initialize content negotiation
	negotiator := NewContentNegotiator(r)
	serializer := negotiator.SelectSerializer()
	
	// Get request ID for response metadata
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = generateRequestID()
	}

	switch r.Method {
	case http.MethodGet:
		// Try to get from format-specific cache
		if cached, found := GetFormatAwareCacheItem(baseCacheKey, format); found {
			if format == FormatTypeJSON {
				// JSON format cache hit
				if config, ok := cached.(map[string]interface{}); ok {
					LogDebug("Retrieved config data from memory cache (JSON format)")
					w.Header().Set("Content-Type", "application/json")
					w.Header().Set("X-Cache-Source", "memory")
					w.Header().Set("X-Cache-Format", "json")
					w.Header().Set("Cache-Control", "public, max-age=3600")
					if err := json.NewEncoder(w).Encode(config); err != nil {
						log.Printf("Error encoding config response: %v", err)
					}
					return
				}
			} else if format == FormatTypeProtobuf {
				// Protobuf format cache hit
				if protoResponse, ok := cached.(*pb.GenericResponse); ok {
					LogDebug("Retrieved config data from memory cache (Protobuf format)")
					
					// Serialize the protobuf response
					responseData, err := serializer.Serialize(protoResponse)
					if err != nil {
						// Fallback to JSON on serialization error
						log.Printf("Error serializing cached protobuf config: %v, falling back to JSON", err)
						// Continue to regenerate the data in JSON format
					} else {
						w.Header().Set("Content-Type", serializer.ContentType())
						w.Header().Set("X-Cache-Source", "memory")
						w.Header().Set("X-Cache-Format", "protobuf")
						w.Header().Set("Cache-Control", "public, max-age=3600")
						w.Write(responseData)
						return
					}
				}
			}
		}

		// Cache miss or serialization error, generate the data
		config := map[string]interface{}{
			"maxUploadSizeMB":      getMaxUploadSize() / (1024 * 1024),
			"maxUploadSizeBytes":   getMaxUploadSize(),
			"useScaledRatings":     GetUseScaledRatings(),
			"datasetRetentionDays": int(getRetentionPeriod().Hours() / 24),
		}

		// Store in format-specific caches
		// For JSON format
		SetFormatAwareCacheItem(baseCacheKey, FormatTypeJSON, config, defaultExpiration)
		
		// For Protobuf format
		configJSON, err := json.Marshal(config)
		if err != nil {
			log.Printf("Error marshaling config to JSON for protobuf: %v", err)
		} else {
			protoConfig := &pb.GenericResponse{
				Data: string(configJSON),
				Metadata: CreateResponseMetadata(requestID, 1, false),
			}
			// Optimize memory usage for protobuf cached data
			optimizedProtoConfig := OptimizeMemoryForProtobuf(protoConfig)
			SetFormatAwareCacheItem(baseCacheKey, FormatTypeProtobuf, optimizedProtoConfig, defaultExpiration)
		}
		

		
		// Respond with the appropriate format
		if format == FormatTypeProtobuf && configJSON != nil {
			protoConfig := &pb.GenericResponse{
				Data: string(configJSON),
				Metadata: CreateResponseMetadata(requestID, 1, false),
			}
			responseData, err := serializer.Serialize(protoConfig)
			if err != nil {
				// Fallback to JSON on serialization error
				log.Printf("Error serializing protobuf config: %v, falling back to JSON", err)
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Cache-Control", "public, max-age=3600")
				if err := json.NewEncoder(w).Encode(config); err != nil {
					log.Printf("Error encoding config response: %v", err)
				}
				return
			}
			
			w.Header().Set("Content-Type", serializer.ContentType())
			w.Header().Set("Cache-Control", "public, max-age=3600")
			w.Write(responseData)
		} else {
			// Default JSON response
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "public, max-age=3600")
			if err := json.NewEncoder(w).Encode(config); err != nil {
				log.Printf("Error encoding config response: %v", err)
			}
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
			// Delete all format variants from cache
			DeleteAllFormatVariants(baseCacheKey)
		}

		config := map[string]interface{}{
			"maxUploadSizeMB":      getMaxUploadSize() / (1024 * 1024),
			"maxUploadSizeBytes":   getMaxUploadSize(),
			"useScaledRatings":     GetUseScaledRatings(),
			"datasetRetentionDays": int(getRetentionPeriod().Hours() / 24),
		}
		
		// Respond with the appropriate format
		if format == FormatTypeProtobuf {
			configJSON, err := json.Marshal(config)
			if err != nil {
				log.Printf("Error marshaling config to JSON for protobuf: %v", err)
				// Fallback to JSON
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(config); err != nil {
					log.Printf("Error encoding config response: %v", err)
				}
				return
			}
			
			protoConfig := &pb.GenericResponse{
				Data: string(configJSON),
				Metadata: CreateResponseMetadata(requestID, 1, false),
			}
			
			responseData, err := serializer.Serialize(protoConfig)
			if err != nil {
				// Fallback to JSON on serialization error
				log.Printf("Error serializing protobuf config: %v, falling back to JSON", err)
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(config); err != nil {
					log.Printf("Error encoding config response: %v", err)
				}
				return
			}
			
			w.Header().Set("Content-Type", serializer.ContentType())
			w.Write(responseData)
		} else {
			// Default JSON response
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(config); err != nil {
				log.Printf("Error encoding config response: %v", err)
			}
		}

	default:
		http.Error(w, "Only GET and POST methods are allowed", http.StatusMethodNotAllowed)
	}
}
