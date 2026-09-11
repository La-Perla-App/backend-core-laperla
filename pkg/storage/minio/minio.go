package minio

import "time"

// Provider identifies the object-store backend.
const (
	ProviderAuto  = "" // infer from endpoint / provider field
	ProviderMinIO = "minio"
	ProviderS3    = "s3"
)

type PresignedUploadResult struct {
	UploadURL string
	Key       string
	ObjectURL string
	Bucket    string
	ExpiresIn time.Duration
	// RequiredHeaders must be sent on the PUT (e.g. Content-Type when signed).
	RequiredHeaders map[string]string
}

type PresignedGetResult struct {
	URL       string
	Key       string
	Bucket    string
	ExpiresIn time.Duration
}

// PresignPutOptions customizes a PUT presign.
type PresignPutOptions struct {
	Expiry      time.Duration
	ContentType string
	// Headers added to the signed request (e.g. Content-Type already covered).
	Headers map[string]string
}

// PresignGetOptions customizes a GET / HEAD / download presign.
type PresignGetOptions struct {
	Expiry            time.Duration
	ResponseHeaders   map[string]string // e.g. response-content-disposition
	VersionID         string
	RequireObjectLock bool
}

type Config struct {
	// Provider: minio | s3. Empty = auto (amazonaws.com → s3, else minio).
	Provider string
	Endpoint string
	Region   string
	AccessKey string
	SecretKey string
	UseSSL    bool

	// PathStyle forces path-style URLs (bucket in path). Nil = auto
	// (minio → true, s3 → false / virtual-hosted).
	PathStyle *bool

	// PublicHost is the host browsers use for signed URLs.
	// MinIO in-cluster: set to the public ingress. S3: leave empty (same as Endpoint)
	// unless you front with a CDN that still expects S3 signatures on this host.
	PublicHost string
	PublicSSL  bool

	DefaultBucket    string
	DefaultPutExpiry time.Duration
	DefaultGetExpiry time.Duration
}

type Client struct {
	// ops talks to the API endpoint (cluster DNS / S3 regional endpoint).
	ops *minioInternal
	// sign builds URLs clients open (publicHost when set, else same as ops).
	sign *minioInternal

	putExpiry time.Duration
	getExpiry time.Duration
	cfg       Config
}
