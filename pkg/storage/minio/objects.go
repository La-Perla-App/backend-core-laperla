package minio

import (
	"bytes"
	"context"
	"fmt"
	"io"

	miniogo "github.com/minio/minio-go/v7"
)

// GetObjectBytes downloads an object into memory.
func (c *Client) GetObjectBytes(ctx context.Context, bucket, key string) ([]byte, error) {
	obj, err := c.internal.GetObject(ctx, bucket, key, miniogo.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}
	defer obj.Close()
	return io.ReadAll(obj)
}

// PutObjectBytes uploads an in-memory object.
func (c *Client) PutObjectBytes(ctx context.Context, bucket, key string, data []byte, contentType string) error {
	if err := c.EnsureBucket(bucket); err != nil {
		return fmt.Errorf("failed to ensure bucket: %w", err)
	}
	opts := miniogo.PutObjectOptions{}
	if contentType != "" {
		opts.ContentType = contentType
	}
	_, err := c.internal.PutObject(ctx, bucket, key, bytes.NewReader(data), int64(len(data)), opts)
	if err != nil {
		return fmt.Errorf("put object: %w", err)
	}
	return nil
}
