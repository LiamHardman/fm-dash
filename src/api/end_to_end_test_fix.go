package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"
	pb "api/proto"
)

// This file contains fixes for the end-to-end test to avoid name conflicts

// Rename the test handler to avoid conflict with the real handler
var testPlayerDataHandler = playerDataHandlerForTest

// The original test handler function
func playerDataHandlerForTest(w http.ResponseWriter, r *http.Request) {
	// Create test player data
	player := &pb.Player{
		Uid:  12345,
		Name: "Test Player",
		Age:  "25",
		Club: "Test FC",
		Position: "ST",
		Overall: 85,
	}
	
	// Create response with metadata
	response := &pb.PlayerDataResponse{
		Players: []*pb.Player{player},
		Metadata: &pb.ResponseMetadata{
			Timestamp:  time.Now().Unix(),
			ApiVersion: "1.0",
			RequestId:  "test-request-id",
			TotalCount: 1,
		},
	}
	
	// Write response based on Accept header
	if strings.Contains(r.Header.Get("Accept"), "application/x-protobuf") {
		w.Header().Set("Content-Type", "application/x-protobuf")
		data, err := proto.Marshal(response)
		if err != nil {
			http.Error(w, "Error serializing response", http.StatusInternalServerError)
			return
		}
		w.Write(data)
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"players": []map[string]interface{}{
				{
					"uid":      12345,
					"name":     "Test Player",
					"age":      "25",
					"club":     "Test FC",
					"position": "ST",
					"overall":  85,
				},
			},
		})
	}
}