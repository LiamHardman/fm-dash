package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// Security validation patterns
var (
	// safeFileNamePattern allows alphanumeric characters, dots, dashes, and underscores
	safeFileNamePattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

	// safeIDPattern for UIDs and Team IDs - more restrictive
	safeIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// validateFileName validates that a filename contains only safe characters
// and doesn't contain path traversal sequences
func validateFileName(filename string) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	// Check length limit
	if len(filename) > 255 {
		return fmt.Errorf("filename too long (max 255 characters)")
	}

	// Check for path traversal attempts
	if strings.Contains(filename, "..") {
		return fmt.Errorf("filename contains path traversal sequence")
	}

	// Check for directory separators
	if strings.ContainsAny(filename, "/\\") {
		return fmt.Errorf("filename contains directory separators")
	}

	// Check for null bytes
	if strings.Contains(filename, "\x00") {
		return fmt.Errorf("filename contains null bytes")
	}

	// Check against safe pattern
	if !safeFileNamePattern.MatchString(filename) {
		return fmt.Errorf("filename contains invalid characters")
	}

	return nil
}

// validateID validates user IDs (UIDs, team IDs) for safety
func validateID(id string, maxLength int) error {
	if id == "" {
		return fmt.Errorf("ID cannot be empty")
	}

	if len(id) > maxLength {
		return fmt.Errorf("ID too long (max %d characters)", maxLength)
	}

	// Check for path traversal attempts
	if strings.Contains(id, "..") {
		return fmt.Errorf("ID contains path traversal sequence")
	}

	// Check for directory separators
	if strings.ContainsAny(id, "/\\") {
		return fmt.Errorf("ID contains directory separators")
	}

	// Check for null bytes
	if strings.Contains(id, "\x00") {
		return fmt.Errorf("ID contains null bytes")
	}

	// Check against safe pattern
	if !safeIDPattern.MatchString(id) {
		return fmt.Errorf("ID contains invalid characters")
	}

	return nil
}

// validateAndJoinPath safely joins a base directory with a filename,
// ensuring the result stays within the base directory
func validateAndJoinPath(baseDir, filename string) (string, error) {
	// Validate the filename first
	if err := validateFileName(filename); err != nil {
		return "", fmt.Errorf("invalid filename: %v", err)
	}

	// Clean the base directory path
	cleanBaseDir := filepath.Clean(baseDir)

	// Join the paths
	fullPath := filepath.Join(cleanBaseDir, filename)

	// Clean the full path
	cleanFullPath := filepath.Clean(fullPath)

	// Ensure the result is still within the base directory
	// Convert to absolute paths for comparison
	absBaseDir, err := filepath.Abs(cleanBaseDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for base directory: %v", err)
	}

	absFullPath, err := filepath.Abs(cleanFullPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for full path: %v", err)
	}

	// Check if the full path is within the base directory
	if !strings.HasPrefix(absFullPath, absBaseDir+string(filepath.Separator)) &&
		absFullPath != absBaseDir {
		return "", fmt.Errorf("path escapes base directory")
	}

	return cleanFullPath, nil
}

// sanitizeForLogging sanitizes input for safe logging
func sanitizeForLogging(input string) string {
	// Remove or escape newlines and carriage returns
	sanitized := strings.ReplaceAll(input, "\n", "\\n")
	sanitized = strings.ReplaceAll(sanitized, "\r", "\\r")
	sanitized = strings.ReplaceAll(sanitized, "\t", "\\t")

	// Truncate if too long
	if len(sanitized) > 200 {
		sanitized = sanitized[:200] + "..."
	}

	return sanitized
}
