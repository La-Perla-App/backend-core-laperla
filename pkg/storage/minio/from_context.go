package minio

import (
	"context"
	"fmt"
	"strings"
)

// FromContext builds an object-store client from the request-scoped config View
// when present, otherwise from process config. Prefer this over Get() in handlers.
func FromContext(ctx context.Context) (*Client, error) {
	cfg := ConfigFromContext(ctx)
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return NewClient(cfg)
}

func (cfg Config) validate() error {
	n := normalizeConfig(cfg)
	if strings.TrimSpace(n.Endpoint) == "" {
		return fmt.Errorf("object store endpoint is not configured (storage.endpoint / minio.endpoint)")
	}
	if strings.TrimSpace(n.AccessKey) == "" || strings.TrimSpace(n.SecretKey) == "" {
		return fmt.Errorf("object store credentials are not configured")
	}
	return nil
}
