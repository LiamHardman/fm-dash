package main

import (
	"path/filepath"
	"regexp"
	"strings"

	apperrors "api/errors"
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
		return apperrors.ErrFilenameEmpty
	}

	// Check length limit
	if len(filename) > 255 {
		return apperrors.ErrFilenameTooLong
	}

	// Check for path traversal attempts
	if strings.Contains(filename, "..") {
		return apperrors.ErrFilenamePathTraversal
	}

	// Check for directory separators
	if strings.ContainsAny(filename, "/\\") {
		return apperrors.ErrFilenameDirectorySeparators
	}

	// Check for null bytes
	if strings.Contains(filename, "\x00") {
		return apperrors.ErrFilenameNullBytes
	}

	// Check against safe pattern
	if !safeFileNamePattern.MatchString(filename) {
		return apperrors.ErrFilenameInvalidChars
	}

	return nil
}

// validateID validates user IDs (UIDs, team IDs) for safety
func validateID(id string, maxLength int) error {
	if id == "" {
		return apperrors.ErrIDEmpty
	}

	if len(id) > maxLength {
		return apperrors.WrapErrIDTooLong(maxLength)
	}

	// Check for path traversal attempts
	if strings.Contains(id, "..") {
		return apperrors.ErrIDPathTraversal
	}

	// Check for directory separators
	if strings.ContainsAny(id, "/\\") {
		return apperrors.ErrIDDirectorySeparators
	}

	// Check for null bytes
	if strings.Contains(id, "\x00") {
		return apperrors.ErrIDNullBytes
	}

	// Check against safe pattern
	if !safeIDPattern.MatchString(id) {
		return apperrors.ErrIDInvalidChars
	}

	return nil
}

// validateAndJoinPath safely joins a base directory with a filename,
// ensuring the result stays within the base directory
func validateAndJoinPath(baseDir, filename string) (string, error) {
	// Validate the filename first
	if err := validateFileName(filename); err != nil {
		return "", apperrors.WrapErrInvalidFilename(err)
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
		return "", apperrors.WrapErrFailedToGetAbsPath(err)
	}

	absFullPath, err := filepath.Abs(cleanFullPath)
	if err != nil {
		return "", apperrors.WrapErrFailedToGetAbsPath(err)
	}

	// Check if the full path is within the base directory
	if !strings.HasPrefix(absFullPath, absBaseDir+string(filepath.Separator)) &&
		absFullPath != absBaseDir {
		return "", apperrors.ErrPathEscapesBase
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
