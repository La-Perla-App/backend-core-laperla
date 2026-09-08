package minio

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	miniogo "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minioInternal = miniogo.Client

func (c *Client) Config() Config {
	return c.cfg
}

func NewClient(cfg Config) (*Client, error) {
	internal, err := miniogo.New(cfg.Endpoint, &miniogo.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: "us-east-1",
	})
	if err != nil {
		return nil, err
	}

	c := &Client{
		internal: internal,
		presign:  internal,
		cfg:      cfg,
	}

	if cfg.DefaultPutExpiry > 0 {
		c.putExpiry = cfg.DefaultPutExpiry
	} else {
		c.putExpiry = 15 * time.Minute
	}
	if cfg.DefaultGetExpiry > 0 {
		c.getExpiry = cfg.DefaultGetExpiry
	} else {
		c.getExpiry = 1 * time.Hour
	}

	if cfg.PublicHost != "" {
		presign, err := miniogo.New(cfg.PublicHost, &miniogo.Options{
			Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
			Secure: cfg.PublicSSL,
			Region: "us-east-1",
		})
		if err != nil {
			slog.Default().Error("failed to initialize MinIO presign client, falling back to internal client", "error", err)
		} else {
			c.presign = presign
		}
	}

	return c, nil
}

func (c *Client) EnsureBucket(bucket string) error {
	ctx := context.Background()
	exists, err := c.internal.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if !exists {
		return c.internal.MakeBucket(ctx, bucket, miniogo.MakeBucketOptions{})
	}
	return nil
}

func (c *Client) PresignedPutURL(bucket, key string, expiry time.Duration) (*PresignedUploadResult, error) {
	if err := c.EnsureBucket(bucket); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket: %w", err)
	}
	if expiry <= 0 {
		expiry = c.putExpiry
	}
	u, err := c.presign.PresignedPutObject(context.Background(), bucket, key, expiry)
	if err != nil {
		return nil, err
	}
	return &PresignedUploadResult{
		UploadURL: u.String(),
		Key:       key,
		ObjectURL: c.ObjectURL(bucket, key),
	}, nil
}

func (c *Client) PresignedGetURL(bucket, key string, expiry time.Duration) (string, error) {
	if expiry <= 0 {
		expiry = c.getExpiry
	}
	u, err := c.presign.PresignedGetObject(context.Background(), bucket, key, expiry, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (c *Client) ObjectURL(bucket, key string) string {
	ep := c.presign.EndpointURL()
	return fmt.Sprintf("%s://%s/%s/%s", ep.Scheme, ep.Host, bucket, key)
}
