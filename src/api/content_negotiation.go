package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	pb "api/proto"

	"google.golang.org/protobuf/proto"
)

// ResponseSerializer defines the interface for different response serialization formats
type ResponseSerializer interface {
	Serialize(data interface{}) ([]byte, error)
	ContentType() string
	ShouldCompress() bool
}

// ProtobufSerializer implements ResponseSerializer for protobuf format
type ProtobufSerializer struct{}

func (p *ProtobufSerializer) Serialize(data interface{}) ([]byte, error) {
	// Convert data to protobuf message based on type
	switch v := data.(type) {
	case *pb.PlayerDataResponse:
		return proto.Marshal(v)
	case *pb.RolesResponse:
		return proto.Marshal(v)
	case *pb.LeaguesResponse:
		return proto.Marshal(v)
	case *pb.TeamsResponse:
		return proto.Marshal(v)
	case *pb.SearchResponse:
		return proto.Marshal(v)
	case *pb.ErrorResponse:
		return proto.Marshal(v)
	case *pb.GenericResponse:
		return proto.Marshal(v)
	default:
		return nil, fmt.Errorf("unsupported data type for protobuf serialization: %T", data)
	}
}

func (p *ProtobufSerializer) ContentType() string {
	return "application/x-protobuf"
}

func (p *ProtobufSerializer) ShouldCompress() bool {
	return true
}

// JSONSerializer implements ResponseSerializer for JSON format
type JSONSerializer struct{}

func (j *JSONSerializer) Serialize(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

func (j *JSONSerializer) ContentType() string {
	return "application/json"
}

func (j *JSONSerializer) ShouldCompress() bool {
	return true
}

// ContentNegotiator handles content negotiation between different response formats
type ContentNegotiator struct {
	request        *http.Request
	supportedTypes map[string]ResponseSerializer
	defaultFormat  string
}

// NewContentNegotiator creates a new content negotiator for the given request
func NewContentNegotiator(r *http.Request) *ContentNegotiator {
	return &ContentNegotiator{
		request: r,
		supportedTypes: map[string]ResponseSerializer{
			"application/x-protobuf": &ProtobufSerializer{},
			"application/protobuf":   &ProtobufSerializer{},
			"application/json":       &JSONSerializer{},
			"*/*":                    &JSONSerializer{}, // Default fallback
		},
		defaultFormat: "application/json",
	}
}

// SelectSerializer determines the best response serializer based on client preferences
func (cn *ContentNegotiator) SelectSerializer() ResponseSerializer {
	acceptHeader := cn.request.Header.Get("Accept")

	// If no Accept header, use default
	if acceptHeader == "" {
		return cn.supportedTypes[cn.defaultFormat]
	}

	// Parse Accept header and find best match
	acceptTypes := parseAcceptHeader(acceptHeader)

	for _, acceptType := range acceptTypes {
		if serializer, exists := cn.supportedTypes[acceptType.MediaType]; exists {
			return serializer
		}
	}

	// Fallback to default
	return cn.supportedTypes[cn.defaultFormat]
}

// SupportsProtobuf checks if the client supports protobuf responses
func (cn *ContentNegotiator) SupportsProtobuf() bool {
	acceptHeader := cn.request.Header.Get("Accept")
	return strings.Contains(acceptHeader, "application/x-protobuf") ||
		strings.Contains(acceptHeader, "application/protobuf")
}

// AcceptType represents a parsed Accept header media type with quality
type AcceptType struct {
	MediaType string
	Quality   float64
}

// parseAcceptHeader parses the Accept header and returns sorted media types by quality
func parseAcceptHeader(acceptHeader string) []AcceptType {
	var acceptTypes []AcceptType

	// Split by comma to get individual media types
	parts := strings.Split(acceptHeader, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split media type and parameters
		segments := strings.Split(part, ";")
		mediaType := strings.TrimSpace(segments[0])

		quality := 1.0 // Default quality

		// Parse quality parameter if present
		for i := 1; i < len(segments); i++ {
			param := strings.TrimSpace(segments[i])
			if strings.HasPrefix(param, "q=") {
				if q, err := parseQuality(param[2:]); err == nil {
					quality = q
				}
			}
		}

		acceptTypes = append(acceptTypes, AcceptType{
			MediaType: mediaType,
			Quality:   quality,
		})
	}

	// Sort by quality (highest first)
	for i := 0; i < len(acceptTypes)-1; i++ {
		for j := i + 1; j < len(acceptTypes); j++ {
			if acceptTypes[i].Quality < acceptTypes[j].Quality {
				acceptTypes[i], acceptTypes[j] = acceptTypes[j], acceptTypes[i]
			}
		}
	}

	return acceptTypes
}

// parseQuality parses quality value from Accept header
func parseQuality(qStr string) (float64, error) {
	qStr = strings.TrimSpace(qStr)
	if qStr == "" {
		return 1.0, nil
	}

	// Simple quality parsing - just handle basic decimal values
	switch qStr {
	case "1", "1.0", "1.00":
		return 1.0, nil
	case "0.9":
		return 0.9, nil
	case "0.8":
		return 0.8, nil
	case "0.7":
		return 0.7, nil
	case "0.6":
		return 0.6, nil
	case "0.5":
		return 0.5, nil
	case "0.4":
		return 0.4, nil
	case "0.3":
		return 0.3, nil
	case "0.2":
		return 0.2, nil
	case "0.1":
		return 0.1, nil
	case "0", "0.0", "0.00":
		return 0.0, nil
	default:
		return 1.0, fmt.Errorf("invalid quality value: %s", qStr)
	}
}

// CreateResponseMetadata creates standard response metadata
func CreateResponseMetadata(requestID string, totalCount int32, fromCache bool) *pb.ResponseMetadata {
	return &pb.ResponseMetadata{
		Timestamp:  time.Now().Unix(),
		ApiVersion: "1.0",
		FromCache:  fromCache,
		RequestId:  requestID,
		TotalCount: totalCount,
	}
}

// CreatePaginationInfo creates pagination information
func CreatePaginationInfo(page, perPage, totalCount int32) *pb.PaginationInfo {
	totalPages := int32(0)
	if perPage > 0 {
		totalPages = (totalCount + perPage - 1) / perPage
	}

	return &pb.PaginationInfo{
		Page:        page,
		PerPage:     perPage,
		TotalPages:  totalPages,
		TotalCount:  totalCount,
		HasNext:     page < totalPages,
		HasPrevious: page > 1,
	}
}

// WriteResponse writes the response using the appropriate serializer
func WriteResponse(w http.ResponseWriter, r *http.Request, data interface{}) error {
	log.Printf("WriteResponse: called with data type = %T", data)

	// Defensive: check if headers have already been written (using custom responseWriter)
	type headerChecker interface {
		Status() int
	}
	if hc, ok := w.(headerChecker); ok {
		status := hc.Status()
		if status != 0 && status != 200 {
			log.Printf("WriteResponse: WARNING - headers already written with status %d, skipping header write", status)
			return nil
		}
	}

	// Don't set CORS headers here as they're already set by CORSMiddleware
	// Only set CORS headers if this is not an OPTIONS request
	if r.Method != "OPTIONS" {
		setCORSHeaders(w, r)
	}

	negotiator := NewContentNegotiator(r)
	serializer := negotiator.SelectSerializer()

	// Debug logging
	log.Printf("WriteResponse: serializer type = %T, content type = %s", serializer, serializer.ContentType())

	// Acquire read lock for concurrent map access protection during serialization
	percentileCalculationMutex.RLock()
	defer percentileCalculationMutex.RUnlock()

	responseData, err := serializer.Serialize(data)
	if err != nil {
		// Fallback to JSON on serialization error
		log.Printf("WriteResponse: serialization failed, falling back to JSON: %v", err)
		jsonSerializer := &JSONSerializer{}
		responseData, jsonErr := jsonSerializer.Serialize(data)
		if jsonErr != nil {
			return fmt.Errorf("both protobuf and JSON serialization failed: %v, %v", err, jsonErr)
		}
		w.Header().Set("Content-Type", jsonSerializer.ContentType())
		log.Printf("WriteResponse: setting JSON content type = %s", jsonSerializer.ContentType())
		w.Write(responseData)
		return nil
	}

	w.Header().Set("Content-Type", serializer.ContentType())
	log.Printf("WriteResponse: setting content type = %s", serializer.ContentType())
	if serializer.ShouldCompress() {
		w.Header().Set("Content-Encoding", "gzip")
	}

	w.Write(responseData)
	return nil
}

// WriteErrorResponse writes an error response using the appropriate format
func WriteErrorResponse(w http.ResponseWriter, r *http.Request, errorCode, message string, details []string, statusCode int) {
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = generateRequestID()
	}

	metadata := CreateResponseMetadata(requestID, 0, false)

	errorResponse := &pb.ErrorResponse{
		ErrorCode: errorCode,
		Message:   message,
		Details:   details,
		Metadata:  metadata,
	}

	w.WriteHeader(statusCode)

	// Try to write protobuf error response, fallback to JSON
	if err := WriteResponse(w, r, errorResponse); err != nil {
		// Final fallback to simple JSON error
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   errorCode,
			"message": message,
			"details": details,
		})
	}
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
