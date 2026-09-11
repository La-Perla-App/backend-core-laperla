package minio

import (
	"context"
	"testing"

	"github.com/La-Perla-App/backend-core-laperla/pkg/config"
)

func TestFromContext_BuildsClientFromView(t *testing.T) {
	if err := config.ReplaceYAML([]byte(`
minio:
  endpoint: minio.minio-system:9000
  accessKey: base-key
  secretKey: base-secret
  useSsl: false
`)); err != nil {
		t.Fatal(err)
	}

	view, err := config.ViewFromYAML([]byte(`
minio:
  endpoint: 100.64.180.61:30900
  accessKey: overlay-key
  secretKey: overlay-secret
  useSsl: false
`))
	if err != nil {
		t.Fatal(err)
	}
	ctx := config.WithView(context.Background(), view)

	c, err := FromContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	cfg := c.Config()
	if cfg.Endpoint != "100.64.180.61:30900" {
		t.Fatalf("endpoint: got %q", cfg.Endpoint)
	}
	if cfg.AccessKey != "overlay-key" {
		t.Fatalf("accessKey: got %q", cfg.AccessKey)
	}
}

func TestFromContext_RequiresEndpoint(t *testing.T) {
	if err := config.ReplaceYAML([]byte(`{}`)); err != nil {
		t.Fatal(err)
	}
	_, err := FromContext(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}
