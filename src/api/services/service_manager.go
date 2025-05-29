// src/api/services/service_manager.go
package services

import (
	"context"
	"log"
)

// ServiceManager coordinates all business services
type ServiceManager struct {
	PlayerService     *PlayerService
	SearchService     *SearchService
	ProcessingService *ProcessingService
}

// NewServiceManager creates a new service manager with all services
func NewServiceManager(storage StorageInterface) *ServiceManager {
	// Initialize services in dependency order
	playerService := NewPlayerService(storage)
	searchService := NewSearchService(playerService)
	processingService := NewProcessingService(playerService)

	manager := &ServiceManager{
		PlayerService:     playerService,
		SearchService:     searchService,
		ProcessingService: processingService,
	}

	log.Println("Service manager initialized with all services")
	return manager
}

// HealthCheck verifies all services are functioning
func (sm *ServiceManager) HealthCheck(ctx context.Context) map[string]string {
	status := make(map[string]string)

	// Check player service
	datasets := sm.PlayerService.GetAllDatasets(ctx)
	if datasets != nil {
		status["player_service"] = "healthy"
	} else {
		status["player_service"] = "unhealthy"
	}

	// Check search service
	// Simple test search with empty dataset (should not error)
	_, err := sm.SearchService.SearchAll(ctx, "test", "", 1)
	if err == nil || err.Error() == "dataset not found: test" {
		status["search_service"] = "healthy"
	} else {
		status["search_service"] = "unhealthy"
	}

	// Check processing service
	stats := sm.ProcessingService.GetProcessingStats(ctx)
	if stats != nil {
		status["processing_service"] = "healthy"
	} else {
		status["processing_service"] = "unhealthy"
	}

	return status
}

// Shutdown gracefully shuts down all services
func (sm *ServiceManager) Shutdown(ctx context.Context) error {
	log.Println("Shutting down service manager...")

	// Services don't currently need explicit shutdown,
	// but this provides a hook for future cleanup

	log.Println("Service manager shutdown complete")
	return nil
}
