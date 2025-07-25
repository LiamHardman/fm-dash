package main

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	pb "api/proto"
)

func TestProtobufSerializer_Serialize(t *testing.T) {
	serializer := &ProtobufSerializer{}

	tests := []struct {
		name     string
		data     interface{}
		wantErr  bool
		errMsg   string
	}{
		{
			name: "PlayerDataResponse serialization",
			data: &pb.PlayerDataResponse{
				Players: []*pb.Player{
					{
						Uid:  1,
						Name: "Test Player",
					},
				},
				CurrencySymbol: "£",
				Metadata: &pb.ResponseMetadata{
					Timestamp:  time.Now().Unix(),
					ApiVersion: "1.0",
					FromCache:  false,
					RequestId:  "test-123",
					TotalCount: 1,
				},
			},
			wantErr: false,
		},
		{
			name: "RolesResponse serialization",
			data: &pb.RolesResponse{
				Roles: []string{"Goalkeeper", "Defender"},
				Metadata: &pb.ResponseMetadata{
					Timestamp:  time.Now().Unix(),
					ApiVersion: "1.0",
				},
			},
			wantErr: false,
		},
		{
			name: "ErrorResponse serialization",
			data: &pb.ErrorResponse{
				ErrorCode: "INVALID_REQUEST",
				Message:   "Invalid request parameters",
				Details:   []string{"Missing required field: id"},
				Metadata: &pb.ResponseMetadata{
					Timestamp:  time.Now().Unix(),
					ApiVersion: "1.0",
				},
			},
			wantErr: false,
		},
		{
			name:    "Unsupported type",
			data:    "invalid string data",
			wantErr: true,
			errMsg:  "unsupported data type for protobuf serialization",
		},
		{
			name:    "Nil data",
			data:    nil,
			wantErr: true,
			errMsg:  "unsupported data type for protobuf serialization",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := serializer.Serialize(tt.data)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error message to contain '%s', got: %s", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(result) == 0 {
				t.Error("Expected non-empty serialized data")
			}

			// Verify we can deserialize the data back
			switch original := tt.data.(type) {
			case *pb.PlayerDataResponse:
				var deserialized pb.PlayerDataResponse
				if err := proto.Unmarshal(result, &deserialized); err != nil {
					t.Errorf("Failed to deserialize PlayerDataResponse: %v", err)
				}
				if deserialized.CurrencySymbol != original.CurrencySymbol {
					t.Errorf("Currency symbol mismatch: got %s, want %s", deserialized.CurrencySymbol, original.CurrencySymbol)
				}
			case *pb.RolesResponse:
				var deserialized pb.RolesResponse
				if err := proto.Unmarshal(result, &deserialized); err != nil {
					t.Errorf("Failed to deserialize RolesResponse: %v", err)
				}
				if len(deserialized.Roles) != len(original.Roles) {
					t.Errorf("Roles count mismatch: got %d, want %d", len(deserialized.Roles), len(original.Roles))
				}
			case *pb.ErrorResponse:
				var deserialized pb.ErrorResponse
				if err := proto.Unmarshal(result, &deserialized); err != nil {
					t.Errorf("Failed to deserialize ErrorResponse: %v", err)
				}
				if deserialized.ErrorCode != original.ErrorCode {
					t.Errorf("Error code mismatch: got %s, want %s", deserialized.ErrorCode, original.ErrorCode)
				}
			}
		})
	}
}

func TestProtobufSerializer_Properties(t *testing.T) {
	serializer := &ProtobufSerializer{}

	if serializer.ContentType() != "application/x-protobuf" {
		t.Errorf("Expected content type 'application/x-protobuf', got: %s", serializer.ContentType())
	}

	if !serializer.ShouldCompress() {
		t.Error("Expected protobuf serializer to support compression")
	}
}

func TestJSONSerializer_Serialize(t *testing.T) {
	serializer := &JSONSerializer{}

	tests := []struct {
		name    string
		data    interface{}
		wantErr bool
	}{
		{
			name: "Simple map serialization",
			data: map[string]interface{}{
				"message": "test",
				"count":   42,
			},
			wantErr: false,
		},
		{
			name: "Struct serialization",
			data: struct {
				Name  string `json:"name"`
				Value int    `json:"value"`
			}{
				Name:  "test",
				Value: 123,
			},
			wantErr: false,
		},
		{
			name: "Array serialization",
			data: []string{"item1", "item2", "item3"},
			wantErr: false,
		},
		{
			name:    "Invalid data (channel)",
			data:    make(chan int),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := serializer.Serialize(tt.data)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(result) == 0 {
				t.Error("Expected non-empty serialized data")
			}

			// Verify it's valid JSON
			var parsed interface{}
			if err := json.Unmarshal(result, &parsed); err != nil {
				t.Errorf("Result is not valid JSON: %v", err)
			}
		})
	}
}

func TestJSONSerializer_Properties(t *testing.T) {
	serializer := &JSONSerializer{}

	if serializer.ContentType() != "application/json" {
		t.Errorf("Expected content type 'application/json', got: %s", serializer.ContentType())
	}

	if !serializer.ShouldCompress() {
		t.Error("Expected JSON serializer to support compression")
	}
}

func TestContentNegotiator_SelectSerializer(t *testing.T) {
	tests := []struct {
		name           string
		acceptHeader   string
		expectedType   string
	}{
		{
			name:         "Protobuf preferred",
			acceptHeader: "application/x-protobuf",
			expectedType: "application/x-protobuf",
		},
		{
			name:         "Alternative protobuf content type",
			acceptHeader: "application/protobuf",
			expectedType: "application/x-protobuf", // Both map to same serializer
		},
		{
			name:         "JSON preferred",
			acceptHeader: "application/json",
			expectedType: "application/json",
		},
		{
			name:         "Multiple types with protobuf first",
			acceptHeader: "application/x-protobuf, application/json",
			expectedType: "application/x-protobuf",
		},
		{
			name:         "Multiple types with JSON first",
			acceptHeader: "application/json, application/x-protobuf",
			expectedType: "application/json",
		},
		{
			name:         "Quality values - protobuf higher",
			acceptHeader: "application/json;q=0.8, application/x-protobuf;q=0.9",
			expectedType: "application/x-protobuf",
		},
		{
			name:         "Quality values - JSON higher",
			acceptHeader: "application/x-protobuf;q=0.7, application/json;q=0.9",
			expectedType: "application/json",
		},
		{
			name:         "Wildcard accept",
			acceptHeader: "*/*",
			expectedType: "application/json", // Default fallback
		},
		{
			name:         "Empty accept header",
			acceptHeader: "",
			expectedType: "application/json", // Default fallback
		},
		{
			name:         "Unsupported type",
			acceptHeader: "application/xml",
			expectedType: "application/json", // Default fallback
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.acceptHeader != "" {
				req.Header.Set("Accept", tt.acceptHeader)
			}

			negotiator := NewContentNegotiator(req)
			serializer := negotiator.SelectSerializer()

			if serializer.ContentType() != tt.expectedType {
				t.Errorf("Expected content type %s, got: %s", tt.expectedType, serializer.ContentType())
			}
		})
	}
}

func TestContentNegotiator_SupportsProtobuf(t *testing.T) {
	tests := []struct {
		name         string
		acceptHeader string
		expected     bool
	}{
		{
			name:         "Supports x-protobuf",
			acceptHeader: "application/x-protobuf",
			expected:     true,
		},
		{
			name:         "Supports protobuf",
			acceptHeader: "application/protobuf",
			expected:     true,
		},
		{
			name:         "Mixed with protobuf support",
			acceptHeader: "application/json, application/x-protobuf",
			expected:     true,
		},
		{
			name:         "Only JSON",
			acceptHeader: "application/json",
			expected:     false,
		},
		{
			name:         "Empty header",
			acceptHeader: "",
			expected:     false,
		},
		{
			name:         "Other types",
			acceptHeader: "text/html, application/xml",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.acceptHeader != "" {
				req.Header.Set("Accept", tt.acceptHeader)
			}

			negotiator := NewContentNegotiator(req)
			result := negotiator.SupportsProtobuf()

			if result != tt.expected {
				t.Errorf("Expected %v, got: %v", tt.expected, result)
			}
		})
	}
}

func TestParseAcceptHeader(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected []AcceptType
	}{
		{
			name:   "Single type",
			header: "application/json",
			expected: []AcceptType{
				{MediaType: "application/json", Quality: 1.0},
			},
		},
		{
			name:   "Multiple types",
			header: "application/json, application/x-protobuf",
			expected: []AcceptType{
				{MediaType: "application/json", Quality: 1.0},
				{MediaType: "application/x-protobuf", Quality: 1.0},
			},
		},
		{
			name:   "With quality values",
			header: "application/json;q=0.8, application/x-protobuf;q=0.9",
			expected: []AcceptType{
				{MediaType: "application/x-protobuf", Quality: 0.9},
				{MediaType: "application/json", Quality: 0.8},
			},
		},
		{
			name:   "Complex header",
			header: "text/html;q=0.9, application/json;q=0.8, */*;q=0.1",
			expected: []AcceptType{
				{MediaType: "text/html", Quality: 0.9},
				{MediaType: "application/json", Quality: 0.8},
				{MediaType: "*/*", Quality: 0.1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAcceptHeader(tt.header)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d types, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i].MediaType != expected.MediaType {
					t.Errorf("Type %d: expected media type %s, got %s", i, expected.MediaType, result[i].MediaType)
				}
				if result[i].Quality != expected.Quality {
					t.Errorf("Type %d: expected quality %f, got %f", i, expected.Quality, result[i].Quality)
				}
			}
		})
	}
}

func TestWriteResponse_FallbackMechanism(t *testing.T) {
	tests := []struct {
		name         string
		data         interface{}
		acceptHeader string
		expectJSON   bool
		expectError  bool
	}{
		{
			name: "Valid protobuf data with protobuf accept",
			data: &pb.PlayerDataResponse{
				Players:        []*pb.Player{{Uid: 1, Name: "Test"}},
				CurrencySymbol: "£",
				Metadata:       &pb.ResponseMetadata{Timestamp: time.Now().Unix()},
			},
			acceptHeader: "application/x-protobuf",
			expectJSON:   false,
			expectError:  false,
		},
		{
			name:         "Invalid protobuf data should fallback to JSON",
			data:         map[string]string{"test": "data"},
			acceptHeader: "application/x-protobuf",
			expectJSON:   true,
			expectError:  false,
		},
		{
			name: "Valid JSON data with JSON accept",
			data: map[string]interface{}{
				"message": "test",
				"count":   42,
			},
			acceptHeader: "application/json",
			expectJSON:   true,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Accept", tt.acceptHeader)
			
			recorder := httptest.NewRecorder()
			
			err := WriteResponse(recorder, req, tt.data)
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
				return
			}
			
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			contentType := recorder.Header().Get("Content-Type")
			
			if tt.expectJSON {
				if contentType != "application/json" {
					t.Errorf("Expected JSON content type, got: %s", contentType)
				}
				
				// Verify response is valid JSON
				var parsed interface{}
				if err := json.Unmarshal(recorder.Body.Bytes(), &parsed); err != nil {
					t.Errorf("Response is not valid JSON: %v", err)
				}
			} else {
				if contentType != "application/x-protobuf" {
					t.Errorf("Expected protobuf content type, got: %s", contentType)
				}
				
				// Verify response is valid protobuf
				if len(recorder.Body.Bytes()) == 0 {
					t.Error("Expected non-empty protobuf response")
				}
			}
		})
	}
}

func TestWriteErrorResponse(t *testing.T) {
	tests := []struct {
		name         string
		errorCode    string
		message      string
		details      []string
		statusCode   int
		acceptHeader string
	}{
		{
			name:         "Protobuf error response",
			errorCode:    "INVALID_REQUEST",
			message:      "Invalid request parameters",
			details:      []string{"Missing field: id"},
			statusCode:   400,
			acceptHeader: "application/x-protobuf",
		},
		{
			name:         "JSON error response",
			errorCode:    "NOT_FOUND",
			message:      "Resource not found",
			details:      []string{"Player with ID 123 not found"},
			statusCode:   404,
			acceptHeader: "application/json",
		},
		{
			name:         "No accept header",
			errorCode:    "INTERNAL_ERROR",
			message:      "Internal server error",
			details:      []string{},
			statusCode:   500,
			acceptHeader: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.acceptHeader != "" {
				req.Header.Set("Accept", tt.acceptHeader)
			}
			
			recorder := httptest.NewRecorder()
			
			WriteErrorResponse(recorder, req, tt.errorCode, tt.message, tt.details, tt.statusCode)
			
			if recorder.Code != tt.statusCode {
				t.Errorf("Expected status code %d, got: %d", tt.statusCode, recorder.Code)
			}
			
			contentType := recorder.Header().Get("Content-Type")
			if contentType == "" {
				t.Error("Expected Content-Type header to be set")
			}
			
			// Verify response body is not empty
			if recorder.Body.Len() == 0 {
				t.Error("Expected non-empty error response body")
			}
			
			// If JSON response, verify it's valid JSON
			if strings.Contains(contentType, "application/json") {
				var parsed interface{}
				if err := json.Unmarshal(recorder.Body.Bytes(), &parsed); err != nil {
					t.Errorf("Error response is not valid JSON: %v", err)
				}
			}
		})
	}
}

func TestCreateResponseMetadata(t *testing.T) {
	requestID := "test-request-123"
	totalCount := int32(42)
	fromCache := true
	
	metadata := CreateResponseMetadata(requestID, totalCount, fromCache)
	
	if metadata == nil {
		t.Fatal("Expected non-nil metadata")
	}
	
	if metadata.RequestId != requestID {
		t.Errorf("Expected request ID %s, got: %s", requestID, metadata.RequestId)
	}
	
	if metadata.TotalCount != totalCount {
		t.Errorf("Expected total count %d, got: %d", totalCount, metadata.TotalCount)
	}
	
	if metadata.FromCache != fromCache {
		t.Errorf("Expected from cache %v, got: %v", fromCache, metadata.FromCache)
	}
	
	if metadata.ApiVersion != "1.0" {
		t.Errorf("Expected API version '1.0', got: %s", metadata.ApiVersion)
	}
	
	if metadata.Timestamp == 0 {
		t.Error("Expected non-zero timestamp")
	}
}

func TestCreatePaginationInfo(t *testing.T) {
	tests := []struct {
		name           string
		page           int32
		perPage        int32
		totalCount     int32
		expectedPages  int32
		expectedNext   bool
		expectedPrev   bool
	}{
		{
			name:          "First page with more pages",
			page:          1,
			perPage:       10,
			totalCount:    25,
			expectedPages: 3,
			expectedNext:  true,
			expectedPrev:  false,
		},
		{
			name:          "Middle page",
			page:          2,
			perPage:       10,
			totalCount:    25,
			expectedPages: 3,
			expectedNext:  true,
			expectedPrev:  true,
		},
		{
			name:          "Last page",
			page:          3,
			perPage:       10,
			totalCount:    25,
			expectedPages: 3,
			expectedNext:  false,
			expectedPrev:  true,
		},
		{
			name:          "Single page",
			page:          1,
			perPage:       10,
			totalCount:    5,
			expectedPages: 1,
			expectedNext:  false,
			expectedPrev:  false,
		},
		{
			name:          "Zero per page",
			page:          1,
			perPage:       0,
			totalCount:    10,
			expectedPages: 0,
			expectedNext:  false,
			expectedPrev:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pagination := CreatePaginationInfo(tt.page, tt.perPage, tt.totalCount)
			
			if pagination == nil {
				t.Fatal("Expected non-nil pagination info")
			}
			
			if pagination.Page != tt.page {
				t.Errorf("Expected page %d, got: %d", tt.page, pagination.Page)
			}
			
			if pagination.PerPage != tt.perPage {
				t.Errorf("Expected per page %d, got: %d", tt.perPage, pagination.PerPage)
			}
			
			if pagination.TotalCount != tt.totalCount {
				t.Errorf("Expected total count %d, got: %d", tt.totalCount, pagination.TotalCount)
			}
			
			if pagination.TotalPages != tt.expectedPages {
				t.Errorf("Expected total pages %d, got: %d", tt.expectedPages, pagination.TotalPages)
			}
			
			if pagination.HasNext != tt.expectedNext {
				t.Errorf("Expected has next %v, got: %v", tt.expectedNext, pagination.HasNext)
			}
			
			if pagination.HasPrevious != tt.expectedPrev {
				t.Errorf("Expected has previous %v, got: %v", tt.expectedPrev, pagination.HasPrevious)
			}
		})
	}
}