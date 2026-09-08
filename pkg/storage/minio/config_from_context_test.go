package minio

import (
	"context"
	"testing"

	"github.com/La-Perla-App/backend-core-laperla/pkg/config"
)

func TestConfigFromContext_UsesViewOverlay(t *testing.T) {
	if err := config.MergeYAML([]byte(`
minio:
  endpoint: minio.minio-system:9000
  accessKey: base-key
  secretKey: base-secret
  useSsl: false
  publicHost: s3.example.com
  publicUseSsl: true
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

	cfg := ConfigFromContext(ctx)
	if cfg.Endpoint != "100.64.180.61:30900" {
		t.Fatalf("endpoint: got %q", cfg.Endpoint)
	}
	if cfg.AccessKey != "overlay-key" || cfg.SecretKey != "overlay-secret" {
		t.Fatalf("credentials: %+v", cfg)
	}
	if cfg.PublicHost != "s3.example.com" {
		t.Fatalf("publicHost should fall through from base, got %q", cfg.PublicHost)
	}
	if !cfg.PublicSSL {
		t.Fatal("publicUseSsl should fall through from base")
	}
}

func TestConfigFromContext_FallsBackToProcessConfig(t *testing.T) {
	if err := config.MergeYAML([]byte(`
minio:
  endpoint: minio.minio-system:9000
  accessKey: base-key
  secretKey: base-secret
  useSsl: true
`)); err != nil {
		t.Fatal(err)
	}

	cfg := ConfigFromContext(context.Background())
	if cfg.Endpoint != "minio.minio-system:9000" {
		t.Fatalf("endpoint: got %q", cfg.Endpoint)
	}
	if !cfg.UseSSL {
		t.Fatal("useSsl expected true")
	}
}
