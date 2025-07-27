# Design Document

## Overview

This design implements protobuf-native API responses to optimize bandwidth usage, reduce serialization overhead, and improve overall system performance. The solution maintains backward compatibility with existing JSON APIs while introducing efficient protobuf endpoints that leverage the existing protobuf schema and conversion infrastructure.

## Architecture

### Content Negotiation Layer
The API will implement HTTP content negotiation to serve both protobuf and JSON responses based on client preferences:

```
Client Request → Content-Type Detection → Response Format Selection → Serialization → Response
```

### Response Flow Comparison

**Current JSON Flow:**
```
Storage (Protobuf) → Conversion to Go Structs → JSON Serialization → HTTP Response
```

**New Protobuf Flow:**
```
Storage (Protobuf) → Direct Protobuf Serialization → HTTP Response
```

### Frontend Integration
The frontend will detect server protobuf support and automatically negotiate the optimal response format, with graceful fallback to JSON for compatibility.

## Components and Interfaces

### 1. Content Negotiation Middleware

**Location:** `src/api/content_negotiation.go`

```go
type ContentNegotiator struct {
    supportedTypes map[string]ResponseSerializer
}

type ResponseSerializer interface {
    Serialize(data interface{}) ([]byte, error)
    ContentType() string
    ShouldCompress() bool
}

type ProtobufSerializer struct{}
type JSONSerializer struct{}
```

**Responsibilities:**
- Parse Accept headers from client requests
- Select appropriate response format (protobuf vs JSON)
- Route to correct serialization handler
- Set appropriate response headers

### 2. Enhanced Handler Functions

**Modified Handlers:**
- `playerDataHandler` - Primary endpoint for player data
- `rolesHandler` - Role definitions
- `leaguesHandler` - League information
- `teamsHandler` - Team data
- `searchHandler` - Search results

**Handler Enhancement Pattern:**
```go
func enhancedPlayerDataHandler(w http.ResponseWriter, r *http.Request) {
    // ... existing logic ...
    
    // Determine response format
    negotiator := NewContentNegotiator(r)
    serializer := negotiator.SelectSerializer()
    
    // Serialize response
    responseData, err := serializer.Serialize(data)
    if err != nil {
        // Fallback to JSON on protobuf errors
        fallbackToJSON(w, data)
        return
    }
    
    // Set headers and send response
    w.Header().Set("Content-Type", serializer.ContentType())
    w.Write(responseData)
}
```

### 3. Protobuf Response Types

**New Protobuf Messages:**
```protobuf
// API Response wrapper for player data
message PlayerDataResponse {
  repeated Player players = 1;
  string currency_symbol = 2;
  ResponseMetadata metadata = 3;
}

// API Response wrapper for roles
message RolesResponse {
  repeated string roles = 1;
  ResponseMetadata metadata = 2;
}

// Common response metadata
message ResponseMetadata {
  int64 timestamp = 1;
  string version = 2;
  int32 total_count = 3;
  bool from_cache = 4;
}
```

### 4. Frontend Protobuf Integration

**New Frontend Components:**

**Location:** `src/utils/protobufClient.js`
```javascript
class ProtobufClient {
  constructor() {
    this.protobufSupported = this.detectProtobufSupport()
    this.protoDefinitions = null
  }
  
  async loadProtobufDefinitions() {
    // Load .proto definitions for client-side decoding
  }
  
  async fetchWithProtobuf(url, options = {}) {
    // Enhanced fetch with protobuf negotiation
  }
  
  decodeProtobufResponse(buffer, messageType) {
    // Decode protobuf response to JavaScript objects
  }
}
```

**Location:** `src/composables/useProtobufApi.js`
```javascript
export function useProtobufApi() {
  const protobufClient = new ProtobufClient()
  
  const fetchPlayerData = async (datasetId, filters = {}) => {
    // Protobuf-aware player data fetching
  }
  
  const fallbackToJSON = async (url, options) => {
    // Graceful fallback mechanism
  }
  
  return { fetchPlayerData, fallbackToJSON }
}
```

### 5. Performance Monitoring Integration

**Enhanced Metrics Collection:**
```go
type ResponseMetrics struct {
    SerializationType string        // "protobuf" or "json"
    PayloadSize      int64         // Response size in bytes
    SerializationTime time.Duration // Time spent serializing
    CompressionRatio float64       // Size reduction ratio
}
```

## Data Models

### Protobuf Schema Extensions

**File:** `src/api/proto/api_responses.proto`
```protobuf
syntax = "proto3";
package api;

import "player.proto";

// Wrapper for paginated player responses
message PlayerDataResponse {
  repeated player.Player players = 1;
  string currency_symbol = 2;
  ResponseMetadata metadata = 3;
  PaginationInfo pagination = 4;
}

message PaginationInfo {
  int32 page = 1;
  int32 per_page = 2;
  int32 total_pages = 3;
  int32 total_count = 4;
}

message ResponseMetadata {
  int64 timestamp = 1;
  string api_version = 2;
  bool from_cache = 3;
  string request_id = 4;
}
```

### Frontend Data Models

**TypeScript Definitions:**
```typescript
interface ProtobufResponse<T> {
  data: T
  metadata: ResponseMetadata
  compressionRatio?: number
}

interface ApiClientConfig {
  preferProtobuf: boolean
  fallbackToJSON: boolean
  compressionThreshold: number
}
```

## Error Handling

### Protobuf-Specific Error Handling

**Backend Error Scenarios:**
1. **Protobuf Serialization Failure:** Automatic fallback to JSON
2. **Schema Version Mismatch:** Graceful degradation with warning logs
3. **Client Incompatibility:** Content negotiation fallback

**Frontend Error Scenarios:**
1. **Protobuf Decoding Failure:** Retry with JSON Accept header
2. **Missing Protobuf Definitions:** Fallback to JSON mode
3. **Network Issues:** Standard retry logic with format preservation

**Error Response Format:**
```protobuf
message ErrorResponse {
  string error_code = 1;
  string message = 2;
  repeated string details = 3;
  ResponseMetadata metadata = 4;
}
```

### Fallback Mechanisms

**Three-Tier Fallback Strategy:**
1. **Primary:** Protobuf response with compression
2. **Secondary:** Protobuf response without compression
3. **Tertiary:** JSON response (current behavior)

## Testing Strategy

### Backend Testing

**Unit Tests:**
- Content negotiation logic
- Protobuf serialization accuracy
- Fallback mechanism reliability
- Performance benchmarking

**Integration Tests:**
- End-to-end protobuf flow
- Mixed client compatibility
- Error scenario handling
- Cache behavior with different formats

**Performance Tests:**
- Serialization speed comparison
- Payload size measurements
- Memory usage analysis
- Concurrent request handling

### Frontend Testing

**Unit Tests:**
- Protobuf client functionality
- Response decoding accuracy
- Fallback mechanism triggers
- Error handling scenarios

**Integration Tests:**
- API compatibility testing
- Cross-browser protobuf support
- Performance impact measurement
- User experience consistency

**Browser Compatibility Tests:**
- Modern browser protobuf support
- Legacy browser fallback behavior
- Mobile device performance
- Network condition adaptability

### Benchmarking Strategy

**Metrics to Track:**
- Response payload size reduction (target: 30-50%)
- Serialization time improvement (target: 20-40%)
- Memory usage optimization
- Network transfer time reduction
- Client-side processing efficiency

**Test Scenarios:**
- Small datasets (< 100 players)
- Medium datasets (100-1000 players)
- Large datasets (> 1000 players)
- Complex filtering operations
- High-concurrency scenarios

## Implementation Phases

### Phase 1: Backend Protobuf Support
- Implement content negotiation middleware
- Add protobuf serializers for existing endpoints
- Create comprehensive test suite
- Performance benchmarking setup

### Phase 2: Frontend Integration
- Develop protobuf client utilities
- Implement automatic format negotiation
- Add fallback mechanisms
- Cross-browser compatibility testing

### Phase 3: Optimization and Monitoring
- Fine-tune compression strategies
- Implement advanced caching for protobuf responses
- Add detailed performance monitoring
- Production deployment and monitoring

This design ensures a smooth transition to protobuf responses while maintaining full backward compatibility and providing significant performance improvements for clients that support the enhanced format.