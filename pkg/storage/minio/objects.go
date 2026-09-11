package minio

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	miniogo "github.com/minio/minio-go/v7"
)

// GetObjectBytes downloads an object into memory.
func (c *Client) GetObjectBytes(ctx context.Context, bucket, key string) ([]byte, error) {
	bucket = c.resolveBucket(bucket)
	key = strings.TrimLeft(strings.TrimSpace(key), "/")
	obj, err := c.ops.GetObject(ctx, bucket, key, miniogo.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}
	defer obj.Close()
	return io.ReadAll(obj)
}

// PutObjectBytes uploads an in-memory object.
func (c *Client) PutObjectBytes(ctx context.Context, bucket, key string, data []byte, contentType string) error {
	bucket = c.resolveBucket(bucket)
	key = strings.TrimLeft(strings.TrimSpace(key), "/")
	if err := c.EnsureBucket(bucket); err != nil {
		return fmt.Errorf("ensure bucket: %w", err)
	}
	opts := miniogo.PutObjectOptions{}
	if contentType != "" {
		opts.ContentType = contentType
	}
	_, err := c.ops.PutObject(ctx, bucket, key, bytes.NewReader(data), int64(len(data)), opts)
	if err != nil {
		return fmt.Errorf("put object: %w", err)
	}
	return nil
}

// DeleteObject removes an object.
func (c *Client) DeleteObject(ctx context.Context, bucket, key string) error {
	bucket = c.resolveBucket(bucket)
	key = strings.TrimLeft(strings.TrimSpace(key), "/")
	return c.ops.RemoveObject(ctx, bucket, key, miniogo.RemoveObjectOptions{})
}
