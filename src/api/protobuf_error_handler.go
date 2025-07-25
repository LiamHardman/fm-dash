package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	pb "api/proto"
)

// ProtobufErrorHandler provides centralized error handling for protobuf operations
type ProtobufErrorHandler struct {
	// Configuration
	maxRetries        int
	enableDetailedLogs bool
	
	// Metrics
	fallbackCount      int64
	errorsByType       map[string]int64
	lastErrorTimestamp time.Time
}

// NewProtobufErrorHandler creates a new ProtobufErrorHandler
func NewProtobufErrorHandler() *ProtobufErrorHandler {
	return &ProtobufErrorHandler{
		maxRetries:        3,
		enableDetailedLogs: true,
		errorsByType:       make(map[string]int64),
	}
}

// HandleSerializationError handles protobuf serialization errors with fallback
func (h *ProtobufErrorHandler) HandleSerializationError(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	data interface{},
	err error,
	datasetID string,
) {
	// Record error metrics
	h.recordError("serialization", datasetID)
	
	// Log detailed error information
	logError(ctx, "Protobuf serialization failed",
		"error", err,
		"dataset_id", datasetID,
		"content_type", r.Header.Get("Content-Type"),
		"accept", r.Header.Get("Accept"),
		"user_agent", r.Header.Get("User-Agent"),
	)
	
	// Create fallback event for monitoring
	fallbackEvent := &ProtobufFallbackEvent{
		DatasetID: datasetID,
		Reason:    FallbackReasonMarshalFailed,
		Error:     err,
		Message:   "Falling back to JSON serialization",
	}
	
	// Log fallback event
	logInfo(ctx, fallbackEvent.String())
	
	// Attempt JSON fallback
	jsonSerializer := &JSONSerializer{}
	responseData, jsonErr := jsonSerializer.Serialize(data)
	
	if jsonErr != nil {
		// Both protobuf and JSON serialization failed
		logError(ctx, "Both protobuf and JSON serialization failed",
			"protobuf_error", err,
			"json_error", jsonErr,
			"dataset_id", datasetID,
		)
		
		// Return a simple error response as last resort
		h.writeLastResortErrorResponse(w, "Serialization failed", http.StatusInternalServerError)
		return
	}
	
	// Set appropriate headers for JSON fallback
	w.Header().Set("Content-Type", jsonSerializer.ContentType())
	w.Header().Set("X-Serialization-Fallback", string(FallbackReasonMarshalFailed))
	
	// Write the JSON response
	w.Write(responseData)
}

// HandleProtobufConversionError handles protobuf conversion errors with fallback
func (h *ProtobufErrorHandler) HandleProtobufConversionError(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	data interface{},
	err error,
	datasetID string,
	direction string,
) {
	// Record error metrics
	h.recordError("conversion_"+direction, datasetID)
	
	// Log detailed error information
	logError(ctx, "Protobuf conversion failed",
		"error", err,
		"direction", direction,
		"dataset_id", datasetID,
	)
	
	// Create fallback event for monitoring
	fallbackEvent := &ProtobufFallbackEvent{
		DatasetID: datasetID,
		Reason:    FallbackReasonConversionFailed,
		Error:     err,
		Message:   fmt.Sprintf("Conversion failed (%s)", direction),
	}
	
	// Log fallback event
	logInfo(ctx, fallbackEvent.String())
	
	// For to_protobuf direction, attempt JSON fallback
	if direction == "to_protobuf" {
		jsonSerializer := &JSONSerializer{}
		responseData, jsonErr := jsonSerializer.Serialize(data)
		
		if jsonErr != nil {
			// Both protobuf conversion and JSON serialization failed
			logError(ctx, "Both protobuf conversion and JSON serialization failed",
				"protobuf_error", err,
				"json_error", jsonErr,
				"dataset_id", datasetID,
			)
			
			// Return a simple error response as last resort
			h.writeLastResortErrorResponse(w, "Data conversion failed", http.StatusInternalServerError)
			return
		}
		
		// Set appropriate headers for JSON fallback
		w.Header().Set("Content-Type", jsonSerializer.ContentType())
		w.Header().Set("X-Serialization-Fallback", string(FallbackReasonConversionFailed))
		
		// Write the JSON response
		w.Write(responseData)
		return
	}
	
	// For from_protobuf direction, return appropriate error
	WriteErrorResponse(w, r, "conversion_error", 
		"Failed to convert data from protobuf format", 
		[]string{err.Error()}, 
		http.StatusInternalServerError)
}

// HandleProtobufCompressionError handles compression/decompression errors
func (h *ProtobufErrorHandler) HandleProtobufCompressionError(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	data interface{},
	err error,
	datasetID string,
	operation string,
) {
	// Record error metrics
	h.recordError("compression_"+operation, datasetID)
	
	// Log detailed error information
	logError(ctx, "Protobuf compression operation failed",
		"error", err,
		"operation", operation,
		"dataset_id", datasetID,
	)
	
	// Create fallback event for monitoring
	var reason ProtobufFallbackReason
	if operation == "compress" {
		reason = FallbackReasonCompressionFailed
	} else {
		reason = FallbackReasonDecompressionFailed
	}
	
	fallbackEvent := &ProtobufFallbackEvent{
		DatasetID: datasetID,
		Reason:    reason,
		Error:     err,
		Message:   fmt.Sprintf("%s operation failed", operation),
	}
	
	// Log fallback event
	logInfo(ctx, fallbackEvent.String())
	
	// For compression failures, try uncompressed protobuf
	if operation == "compress" {
		// Try to serialize without compression
		protobufSerializer := &ProtobufSerializer{}
		responseData, pbErr := protobufSerializer.Serialize(data)
		
		if pbErr != nil {
			// Fall back to JSON as last resort
			h.HandleSerializationError(ctx, w, r, data, pbErr, datasetID)
			return
		}
		
		// Set appropriate headers for uncompressed protobuf
		w.Header().Set("Content-Type", protobufSerializer.ContentType())
		w.Header().Set("X-Compression-Status", "disabled")
		
		// Write the uncompressed protobuf response
		w.Write(responseData)
		return
	}
	
	// For decompression failures, return appropriate error
	WriteErrorResponse(w, r, "decompression_error", 
		"Failed to decompress protobuf data", 
		[]string{err.Error()}, 
		http.StatusInternalServerError)
}

// HandleClientCompatibilityError handles client compatibility issues
func (h *ProtobufErrorHandler) HandleClientCompatibilityError(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	data interface{},
	clientInfo string,
	datasetID string,
) {
	// Record error metrics
	h.recordError("client_compatibility", datasetID)
	
	// Log detailed error information
	logInfo(ctx, "Client compatibility issue detected",
		"client_info", clientInfo,
		"dataset_id", datasetID,
		"user_agent", r.Header.Get("User-Agent"),
	)
	
	// Fall back to JSON for this client
	jsonSerializer := &JSONSerializer{}
	responseData, jsonErr := jsonSerializer.Serialize(data)
	
	if jsonErr != nil {
		// JSON serialization failed
		logError(ctx, "JSON serialization failed for compatibility fallback",
			"error", jsonErr,
			"dataset_id", datasetID,
		)
		
		// Return a simple error response as last resort
		h.writeLastResortErrorResponse(w, "Serialization failed", http.StatusInternalServerError)
		return
	}
	
	// Set appropriate headers for JSON fallback
	w.Header().Set("Content-Type", jsonSerializer.ContentType())
	w.Header().Set("X-Serialization-Fallback", "client_compatibility")
	
	// Write the JSON response
	w.Write(responseData)
}

// CreateErrorResponse creates a protobuf error response
func (h *ProtobufErrorHandler) CreateErrorResponse(
	ctx context.Context,
	errorCode string,
	message string,
	details []string,
	requestID string,
) *pb.ErrorResponse {
	if requestID == "" {
		requestID = fmt.Sprintf("err_%d", time.Now().UnixNano())
	}
	
	metadata := CreateResponseMetadata(requestID, 0, false)
	
	return &pb.ErrorResponse{
		ErrorCode: errorCode,
		Message:   message,
		Details:   details,
		Metadata:  metadata,
	}
}

// writeLastResortErrorResponse writes a simple JSON error when all else fails
func (h *ProtobufErrorHandler) writeLastResortErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	// Use simple map for maximum compatibility
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "serialization_error",
		"message": message,
		"time":    time.Now().Unix(),
	})
}

// recordError records error metrics for monitoring
func (h *ProtobufErrorHandler) recordError(errorType string, datasetID string) {
	h.errorsByType[errorType]++
	h.fallbackCount++
	h.lastErrorTimestamp = time.Now()
	
	// In a real implementation, we would send these metrics to a monitoring system
}

// GetErrorMetrics returns error metrics for monitoring
func (h *ProtobufErrorHandler) GetErrorMetrics() map[string]interface{} {
	return map[string]interface{}{
		"fallback_count":       h.fallbackCount,
		"errors_by_type":       h.errorsByType,
		"last_error_timestamp": h.lastErrorTimestamp.Unix(),
	}
}

// Global instance for convenience
var globalProtobufErrorHandler = NewProtobufErrorHandler()

// GetProtobufErrorHandler returns the global error handler instance
func GetProtobufErrorHandler() *ProtobufErrorHandler {
	return globalProtobufErrorHandler
}