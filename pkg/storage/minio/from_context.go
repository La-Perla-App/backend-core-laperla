package minio

import (
	"context"
	"fmt"
	"strings"
)

// FromContext builds a MinIO client from the request-scoped config View when
// present, otherwise from process config. Prefer this over Get() in handlers.
func FromContext(ctx context.Context) (*Client, error) {
	cfg := ConfigFromContext(ctx)
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return NewClient(cfg)
}

func (cfg Config) validate() error {
	if strings.TrimSpace(cfg.Endpoint) == "" {
		return fmt.Errorf("minio.endpoint is not configured")
	}
	if strings.TrimSpace(cfg.AccessKey) == "" || strings.TrimSpace(cfg.SecretKey) == "" {
		return fmt.Errorf("minio credentials are not configured")
	}
	return nil
}
