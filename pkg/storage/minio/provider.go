package minio

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

var (
	singleton         *Client
	once              sync.Once
	ErrNotInitialized = errors.New("object store provider no inicializado")
)

// Init stores an optional process-wide client. Prefer FromContext for request work.
// If endpoint/credentials are missing, Init is a no-op (services without storage stay up).
func Init(cfg Config) (err error) {
	if !cfg.Configured() {
		slog.Default().Info("object store not configured; skipping Init")
		return nil
	}
	once.Do(func() {
		singleton, err = NewClient(cfg)
		if err != nil {
			err = fmt.Errorf("object store init: %w", err)
			return
		}
	})
	return err
}

// EnsureBucket creates a client from process config and ensures the bucket.
// Deprecated for request paths: use FromContext(ctx).EnsureBucket.
func EnsureBucket(bucket string) error {
	c, err := FromContext(context.Background())
	if err != nil {
		return err
	}
	return c.EnsureBucket(bucket)
}

// PresignedPutURL uses process config.
// Deprecated for request paths: use FromContext(ctx).PresignedPutURL.
func PresignedPutURL(bucket, key string, expiry time.Duration) (*PresignedUploadResult, error) {
	c, err := FromContext(context.Background())
	if err != nil {
		return nil, err
	}
	return c.PresignedPutURL(bucket, key, expiry)
}

// PresignedGetURL uses process config.
// Deprecated for request paths: use FromContext(ctx).PresignedGetURL.
func PresignedGetURL(bucket, key string, expiry time.Duration) (string, error) {
	c, err := FromContext(context.Background())
	if err != nil {
		return "", err
	}
	return c.PresignedGetURL(bucket, key, expiry)
}

// ObjectURL uses process config.
// Deprecated for request paths: use FromContext(ctx).ObjectURL.
func ObjectURL(bucket, key string) string {
	c, err := FromContext(context.Background())
	if err != nil {
		return ""
	}
	return c.ObjectURL(bucket, key)
}

// Get returns the optional boot-time singleton. Prefer FromContext.
func Get() *Client {
	return singleton
}
