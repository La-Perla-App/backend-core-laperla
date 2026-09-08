package minio

import "time"

type PresignedUploadResult struct {
	UploadURL string
	Key       string
	ObjectURL string
}

type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool

	PublicHost string
	PublicSSL  bool

	DefaultPutExpiry time.Duration
	DefaultGetExpiry time.Duration
}

type Client struct {
	internal  *minioInternal
	presign   *minioInternal
	putExpiry time.Duration
	getExpiry time.Duration
	cfg       Config
}
