package minio

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	miniogo "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minioInternal = miniogo.Client

func (c *Client) Config() Config {
	return c.cfg
}

func (c *Client) DefaultBucket() string {
	return c.cfg.DefaultBucket
}

func (c *Client) Provider() string {
	return resolveProvider(c.cfg)
}

// NewClient builds ops + optional public signing clients.
//
// Two clients are useful when the API endpoint is not reachable by browsers
// (MinIO Service DNS vs Ingress). For AWS S3, one client is usually enough;
// set PublicHost only if signed URLs must use a different hostname.
func NewClient(cfg Config) (*Client, error) {
	cfg = normalizeConfig(cfg)
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	ops, err := newSDKClient(cfg.Endpoint, cfg.AccessKey, cfg.SecretKey, cfg.UseSSL, cfg.Region, pathStyleFor(cfg))
	if err != nil {
		return nil, fmt.Errorf("object store ops client: %w", err)
	}

	c := &Client{
		ops:       ops,
		sign:      ops,
		cfg:       cfg,
		putExpiry: cfg.DefaultPutExpiry,
		getExpiry: cfg.DefaultGetExpiry,
	}
	if c.putExpiry <= 0 {
		c.putExpiry = 15 * time.Minute
	}
	if c.getExpiry <= 0 {
		c.getExpiry = time.Hour
	}

	if pub := strings.TrimSpace(cfg.PublicHost); pub != "" && !sameHost(cfg.Endpoint, pub) {
		sign, err := newSDKClient(pub, cfg.AccessKey, cfg.SecretKey, cfg.PublicSSL, cfg.Region, pathStyleFor(cfg))
		if err != nil {
			slog.Default().Error("object store presign client failed; using ops endpoint for signed URLs", "error", err)
		} else {
			c.sign = sign
		}
	}

	slog.Default().Info("object store client ready",
		"provider", resolveProvider(cfg),
		"endpoint", cfg.Endpoint,
		"publicHost", strings.TrimSpace(cfg.PublicHost),
		"region", cfg.Region,
		"defaultBucket", cfg.DefaultBucket,
	)
	return c, nil
}

func newSDKClient(endpoint, access, secret string, secure bool, region string, pathStyle bool) (*miniogo.Client, error) {
	endpoint = stripScheme(endpoint)
	lookup := miniogo.BucketLookupAuto
	if pathStyle {
		lookup = miniogo.BucketLookupPath
	} else if resolveProvider(Config{Endpoint: endpoint, Provider: ""}) == ProviderS3 {
		lookup = miniogo.BucketLookupDNS
	}
	return miniogo.New(endpoint, &miniogo.Options{
		Creds:        credentials.NewStaticV4(access, secret, ""),
		Secure:       secure,
		Region:       region,
		BucketLookup: lookup,
	})
}

func normalizeConfig(cfg Config) Config {
	cfg.Provider = strings.ToLower(strings.TrimSpace(cfg.Provider))
	cfg.Endpoint = stripScheme(strings.TrimSpace(cfg.Endpoint))
	cfg.PublicHost = stripScheme(strings.TrimSpace(cfg.PublicHost))
	cfg.Region = strings.TrimSpace(cfg.Region)
	cfg.AccessKey = strings.TrimSpace(cfg.AccessKey)
	cfg.SecretKey = strings.TrimSpace(cfg.SecretKey)
	cfg.DefaultBucket = strings.TrimSpace(cfg.DefaultBucket)

	provider := resolveProvider(cfg)
	if cfg.Region == "" {
		if provider == ProviderS3 {
			cfg.Region = "us-east-1"
		} else {
			cfg.Region = "us-east-1" // MinIO ignores region for most ops but SDK wants one
		}
	}
	if cfg.Endpoint == "" && provider == ProviderS3 {
		if cfg.Region == "us-east-1" {
			cfg.Endpoint = "s3.amazonaws.com"
		} else {
			cfg.Endpoint = "s3." + cfg.Region + ".amazonaws.com"
		}
	}
	if provider == ProviderS3 && !cfg.UseSSL && !strings.Contains(cfg.Endpoint, "localhost") {
		cfg.UseSSL = true
	}
	return cfg
}

func resolveProvider(cfg Config) string {
	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case ProviderS3, "aws", "aws-s3":
		return ProviderS3
	case ProviderMinIO:
		return ProviderMinIO
	}
	ep := strings.ToLower(cfg.Endpoint + " " + cfg.PublicHost)
	if strings.Contains(ep, "amazonaws.com") || strings.Contains(ep, "amazonaws.com.cn") {
		return ProviderS3
	}
	return ProviderMinIO
}

func pathStyleFor(cfg Config) bool {
	if cfg.PathStyle != nil {
		return *cfg.PathStyle
	}
	return resolveProvider(cfg) == ProviderMinIO
}

func stripScheme(host string) string {
	host = strings.TrimSpace(host)
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	return strings.TrimRight(host, "/")
}

func sameHost(a, b string) bool {
	return strings.EqualFold(stripScheme(a), stripScheme(b))
}

func (c *Client) EnsureBucket(bucket string) error {
	bucket = c.resolveBucket(bucket)
	if bucket == "" {
		return fmt.Errorf("bucket is required")
	}
	ctx := context.Background()
	exists, err := c.ops.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if !exists {
		opts := miniogo.MakeBucketOptions{Region: c.cfg.Region}
		return c.ops.MakeBucket(ctx, bucket, opts)
	}
	return nil
}

func (c *Client) resolveBucket(bucket string) string {
	bucket = strings.TrimSpace(bucket)
	if bucket == "" {
		return c.cfg.DefaultBucket
	}
	return bucket
}

func (c *Client) PresignedPutURL(bucket, key string, expiry time.Duration) (*PresignedUploadResult, error) {
	return c.PresignedPut(bucket, key, PresignPutOptions{Expiry: expiry})
}

func (c *Client) PresignedPut(bucket, key string, opts PresignPutOptions) (*PresignedUploadResult, error) {
	bucket = c.resolveBucket(bucket)
	key = strings.TrimLeft(strings.TrimSpace(key), "/")
	if bucket == "" || key == "" {
		return nil, fmt.Errorf("bucket and key are required")
	}
	if err := c.EnsureBucket(bucket); err != nil {
		return nil, fmt.Errorf("ensure bucket: %w", err)
	}
	expiry := opts.Expiry
	if expiry <= 0 {
		expiry = c.putExpiry
	}
	extra := make(http.Header)
	required := map[string]string{}
	if opts.ContentType != "" {
		extra.Set("Content-Type", opts.ContentType)
		required["Content-Type"] = opts.ContentType
	}
	for k, v := range opts.Headers {
		if k == "" || v == "" {
			continue
		}
		extra.Set(k, v)
		required[k] = v
	}

	var (
		u   *url.URL
		err error
	)
	if len(extra) > 0 {
		u, err = c.sign.PresignHeader(context.Background(), http.MethodPut, bucket, key, expiry, nil, extra)
	} else {
		u, err = c.sign.PresignedPutObject(context.Background(), bucket, key, expiry)
	}
	if err != nil {
		return nil, err
	}
	return &PresignedUploadResult{
		UploadURL:       u.String(),
		Key:             key,
		Bucket:          bucket,
		ObjectURL:       c.ObjectURL(bucket, key),
		ExpiresIn:       expiry,
		RequiredHeaders: required,
	}, nil
}

func (c *Client) PresignedGetURL(bucket, key string, expiry time.Duration) (string, error) {
	res, err := c.PresignedGet(bucket, key, PresignGetOptions{Expiry: expiry})
	if err != nil {
		return "", err
	}
	return res.URL, nil
}

func (c *Client) PresignedGet(bucket, key string, opts PresignGetOptions) (*PresignedGetResult, error) {
	bucket = c.resolveBucket(bucket)
	key = strings.TrimLeft(strings.TrimSpace(key), "/")
	if bucket == "" || key == "" {
		return nil, fmt.Errorf("bucket and key are required")
	}
	expiry := opts.Expiry
	if expiry <= 0 {
		expiry = c.getExpiry
	}
	reqParams := make(url.Values)
	for k, v := range opts.ResponseHeaders {
		if k != "" && v != "" {
			reqParams.Set(k, v)
		}
	}
	u, err := c.sign.PresignedGetObject(context.Background(), bucket, key, expiry, reqParams)
	if err != nil {
		return nil, err
	}
	return &PresignedGetResult{
		URL:       u.String(),
		Key:       key,
		Bucket:    bucket,
		ExpiresIn: expiry,
	}, nil
}

// ObjectURL returns a non-signed object locator (useful as stable key reference).
// For private buckets prefer PresignedGet.
func (c *Client) ObjectURL(bucket, key string) string {
	bucket = c.resolveBucket(bucket)
	key = strings.TrimLeft(strings.TrimSpace(key), "/")
	ep := c.sign.EndpointURL()
	if resolveProvider(c.cfg) == ProviderS3 && !pathStyleFor(c.cfg) {
		// Virtual-hosted–style: https://bucket.s3.region.amazonaws.com/key
		host := ep.Host
		return fmt.Sprintf("%s://%s.%s/%s", ep.Scheme, bucket, host, key)
	}
	return fmt.Sprintf("%s://%s/%s/%s", ep.Scheme, ep.Host, bucket, key)
}
