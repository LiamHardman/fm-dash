package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
)

// Write a complete temporary file beside the target before replacing it. Readers
// see either revision; a failed write leaves the previous document intact.
func atomicConfigWrite(path string, data []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".profile-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer func() { _ = os.Remove(name) }()
	if _, err = f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}

func (s *localConfigStorage) CompareAndSwapConfig(name string, data []byte, expected string) error {
	path := filepath.Join(s.dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	// OS locks are released on process exit and coordinate replicas sharing a disk.
	lock, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	defer func() { _ = lock.Close() }()
	if err = lockConfigFile(lock); err != nil {
		return err
	}
	defer unlockConfigFile(lock)
	current, err := os.ReadFile(path)
	if err != nil && !configNotFound(err) {
		return err
	}
	if profileRevision(current) != expected {
		return errProfileConflict
	}
	return atomicConfigWrite(path, data)
}

func (s *s3ConfigStorage) CompareAndSwapConfig(name string, data []byte, expected string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	obj, err := s.client.GetObject(ctx, s.bucketName, name, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = obj.Close() }()
	current, err := io.ReadAll(obj)
	options := minio.PutObjectOptions{ContentType: "application/json", DisableMultipart: true}
	if minio.ToErrorResponse(err).Code == "NoSuchKey" {
		current = nil
		options.SetMatchETagExcept("*")
	} else if err != nil {
		return err
	} else {
		info, statErr := obj.Stat()
		if statErr != nil {
			return statErr
		}
		options.SetMatchETag(info.ETag)
	}
	if profileRevision(current) != expected {
		return errProfileConflict
	}
	_, err = s.client.PutObject(ctx, s.bucketName, name, bytes.NewReader(data), int64(len(data)), options)
	if code := minio.ToErrorResponse(err).Code; code == "PreconditionFailed" || code == "ConditionalRequestConflict" {
		return errProfileConflict
	}
	return err
}
