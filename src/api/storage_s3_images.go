package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	apperrors "api/errors"

	"github.com/minio/minio-go/v7"
)

// underlyingS3Storage unwraps ProtobufStorage (and any future wrappers) to find
// the S3Storage underneath, returning nil if S3 is not the backing store.
func underlyingS3Storage() *S3Storage {
	if s3, ok := storage.(*S3Storage); ok {
		return s3
	}
	if proto, ok := storage.(*ProtobufStorage); ok {
		if s3, ok := proto.backend.(*S3Storage); ok {
			return s3
		}
	}
	return nil
}

// getFaceImage retrieves a face image from S3 and writes it to the response writer
func (s *S3Storage) getFaceImage(ctx context.Context, filename string, w http.ResponseWriter) error {
	if s.client == nil {
		return apperrors.ErrS3ClientNotAvailable
	}

	// Get the faces bucket name from environment, default to the main bucket + "/faces" prefix
	facesBucketName := os.Getenv("S3_FACES_BUCKET")
	if facesBucketName == "" {
		facesBucketName = s.bucketName // Use same bucket as datasets
	}

	// Construct the object key for faces
	objectKey := "faces/" + filename

	// If we're using a separate faces bucket, don't add the prefix
	if facesBucketName != s.bucketName {
		objectKey = filename
	}

	// Get object from S3
	reader, err := s.client.GetObject(ctx, facesBucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get face image from S3: %w", err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			LogWarn("Failed to close S3 reader: %v", closeErr)
		}
	}()

	// Copy the image data to the response writer
	_, err = io.Copy(w, reader)
	if err != nil {
		return fmt.Errorf("failed to write face image to response: %w", err)
	}

	return nil
}

// getTeamLogo retrieves a team logo image from S3 and writes it to the response writer
func (s *S3Storage) getTeamLogo(ctx context.Context, filename string, w http.ResponseWriter) error {
	if s.client == nil {
		return apperrors.ErrS3ClientNotAvailable
	}

	// Get the logos bucket name from environment, default to the main bucket + "/logos/clubs" prefix
	logosBucketName := os.Getenv("S3_LOGOS_BUCKET")
	if logosBucketName == "" {
		logosBucketName = s.bucketName // Use same bucket as datasets
	}

	// Construct the object key for logos
	objectKey := "logos/Clubs/Normal/Normal/" + filename

	// If we're using a separate logos bucket, don't add the prefix
	if logosBucketName != s.bucketName {
		objectKey = filename
	}

	// Get object from S3
	reader, err := s.client.GetObject(ctx, logosBucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get team logo from S3: %w", err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			LogWarn("Failed to close S3 reader: %v", closeErr)
		}
	}()

	// Copy the image data to the response writer
	_, err = io.Copy(w, reader)
	if err != nil {
		return fmt.Errorf("failed to write team logo to response: %w", err)
	}

	return nil
}
