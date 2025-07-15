// +build ignore

package main

// This file contains build configuration and dependency management for Parquet/Arrow integration
// It ensures that the required CGO flags and build tags are properly set

/*
#cgo CPPFLAGS: -I/usr/local/include
#cgo LDFLAGS: -L/usr/local/lib
*/
import "C"

// Build tags for Arrow/Parquet support
// These ensure that the necessary features are enabled during compilation

const (
	// Arrow build configuration
	ArrowVersion = "17.0.0"
	
	// Parquet build configuration  
	ParquetVersion = "0.25.1"
	
	// Feature flags
	EnableArrowCompute = true
	EnableParquetCompression = true
	EnableBloomFilters = true
)

// BuildInfo contains information about the build configuration
type BuildInfo struct {
	ArrowVersion    string
	ParquetVersion  string
	CGOEnabled      bool
	ComputeEnabled  bool
	CompressionEnabled bool
}

// GetBuildInfo returns the current build information
func GetBuildInfo() BuildInfo {
	return BuildInfo{
		ArrowVersion:       ArrowVersion,
		ParquetVersion:     ParquetVersion,
		CGOEnabled:         true,
		ComputeEnabled:     EnableArrowCompute,
		CompressionEnabled: EnableParquetCompression,
	}
}